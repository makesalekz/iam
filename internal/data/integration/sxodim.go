package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	iam_v1 "gitlab.calendaria.team/services/iam/api/iam/v1"
	"gitlab.calendaria.team/services/iam/ent"
	"gitlab.calendaria.team/services/utils/v1/config"
	u_struc "gitlab.calendaria.team/services/utils/v2/struc"

	xoauth2 "golang.org/x/oauth2"
)

// Sxodim constants.
const (
	SxodimAuthUrl   = "https://dev.sxodim.com/oauth/token"
	SxodimGrantType = "authorization_code"
	SxodimClientID  = "3"
)

type SxodimGateway struct {
	config *config.Config
}

func NewSxodimRemote(
	config *config.Config,
) (IProviderGateway, error) {
	return &SxodimGateway{
		config: config,
	}, nil
}

type sxodimAuthRequestBody struct {
	GrantType    string `json:"grant_type"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Code         string `json:"code"`
}

type sxodimAuthResponseBody struct {
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"` // sec from time.Now
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// exchangeSxodimToken exchanges an authorization code for an OAuth2 token via Sxodim's API.
//
// Notes:
//   - Requires "sxodimclientsecret" to be properly configured in the secret store
//   - Requires SxodimAuthUrl is preconfigured with the correct endpoint
//   - Converts ExpiresIn from milliseconds to time.Time expiry
func (r *SxodimGateway) exchangeSxodimToken(ctx context.Context, authCode string) (*xoauth2.Token, error) {
	// get sxodim client secret
	sxodimConfig, err := r.config.ReadSecretsFor(ctx, "sxodimclientsecret")
	if err != nil {
		return nil, iam_v1.ErrorInternal("failed getting sxodim client secret: %s", err)
	}

	// Validate and retrieve the Sxodim client secret from configuration
	sxodimClientSecret, ok := sxodimConfig["secret"].(string)
	if !ok {
		return nil, iam_v1.ErrorInternal("sxodim client secret is not set: %s", err)
	}

	// collect body
	body := sxodimAuthRequestBody{
		GrantType:    SxodimGrantType,
		ClientID:     SxodimClientID,
		ClientSecret: sxodimClientSecret,
		Code:         authCode,
	}

	// marshal body to []byte
	requestBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	// send post request
	res, err := http.Post(SxodimAuthUrl, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, iam_v1.ErrorInternal("unable to create Google Calendar notification channel: %s", err)
	}
	defer res.Body.Close()

	// check status
	if res.StatusCode != http.StatusOK {
		return nil, iam_v1.ErrorInternal(
			"request failed on getting access token: %s",
			res.Status,
		)
	}

	// decode response
	response := &sxodimAuthResponseBody{}
	err = json.NewDecoder(res.Body).Decode(response)
	if err != nil {
		return nil, iam_v1.ErrorInternal("unable to decode response: %s", err)
	}

	// calculate expire date
	expireDate := time.Now().Add(time.Duration(response.ExpiresIn) * time.Second)

	// collect token
	token := &xoauth2.Token{
		AccessToken:  response.AccessToken,
		TokenType:    response.TokenType,
		RefreshToken: response.RefreshToken,
		Expiry:       expireDate,
	}

	return token, nil
}

func (r *SxodimGateway) Authenticate(ctx context.Context, actorID int64, authCode string) (*CredentialDto, error) {
	token, err := r.exchangeSxodimToken(ctx, authCode)
	if err != nil {
		return nil, err
	}

	dto := &CredentialDto{
		UserID:   actorID,
		Provider: u_struc.Sxodim,
		Token:    token,
	}

	return dto, nil
}

func (r *SxodimGateway) RefreshToken(
	ctx context.Context,
	credential *ent.UserCredentials,
) (*CredentialDto, error) {
	// Retrieve dto from credential
	dto := CredentialToDto(credential)

	// Check valid token and refresh token existence
	if dto == nil || dto.Token == nil || dto.Token.RefreshToken == "" {
		return nil, iam_v1.ErrorNeedReauthorization("invalid token or null refresh token, need reauthorization")
	}

	// Check if token valid
	if !dto.Token.Valid() {
		newToken, err := r.exchangeSxodimToken(ctx, dto.Token.RefreshToken)
		if err != nil {
			return nil, err
		}

		dto.Token = newToken
	}

	return dto, nil
}

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	v1 "gitlab.calendaria.team/services/contacts/api/contacts/v1"
	iam_v1 "gitlab.calendaria.team/services/iam/api/iam/v1"
	"gitlab.calendaria.team/services/iam/ent"
	u_struc "gitlab.calendaria.team/services/utils/v2/struc"
	"gitlab.calendaria.team/services/utils/v4/config"

	xoauth2 "golang.org/x/oauth2"
)

// Sxodim constants.
const (
	HTTPTimeout = time.Second * 30

	SxodimAuthUrl     = "/oauth/token"
	SxodimUserDataUrl = "/api/aigenda/user"
	SxodimGrantType   = "authorization_code"
	SxodimClientID    = "3"
)

type SxodimStorage struct {
	SxodimDomain  string
	ClientSecret  string
	AigendaSecret string
}

type SxodimGateway struct {
	config  config.IConfig
	storage SxodimStorage
}

func NewSxodimRemote(
	config config.IConfig,
) (IProviderGateway, error) {
	g := &SxodimGateway{
		config: config,
	}

	// Get Sxodim domain
	sxodimDomain, err := g.config.GetValue("SXODIM_DOMAIN")
	if err != nil {
		return nil, v1.ErrorDatabaseQuery("error on getting SXODIM_DOMAIN value")
	}

	// get sxodim client secret
	sercretClientConfig, err := g.config.ReadSecretsFor(context.Background(), "sxodimclientsecret")
	if err != nil {
		return nil, iam_v1.ErrorInternal("failed getting sxodim client secret: %s", err)
	}

	// Validate and retrieve the Sxodim client secret from configuration
	sxodimClientSecret, ok := sercretClientConfig["secret"].(string)
	if !ok {
		return nil, iam_v1.ErrorInternal("sxodim client secret is not set: %s", err)
	}

	// get sxodim client secret
	secretAigendaConfig, err := g.config.ReadGlobalSecretsFor(context.Background(), "sxodimsecretaigenda")
	if err != nil {
		return nil, v1.ErrorInternal("failed getting sxodim secret aigenda: %s", err)
	}

	// Validate and retrieve the Sxodim secret aigenda from configuration
	sxodimSecretAigenda, ok := secretAigendaConfig["secret"].(string)
	if !ok {
		return nil, v1.ErrorInternal("sxodim secret aigenda is not set: %s", err)
	}

	g.storage = SxodimStorage{
		SxodimDomain:  sxodimDomain,
		ClientSecret:  sxodimClientSecret,
		AigendaSecret: sxodimSecretAigenda,
	}

	return g, nil
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

type sxodimUser struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Phone string `json:"phone"`
}

type sxodimUserDataReply struct {
	User sxodimUser `json:"data"`
}

func (g *SxodimGateway) doSxodimRequest(
	ctx context.Context,
	request *http.Request,
) (*http.Response, func(), error) {
	// Collect header
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "*/*")

	// Do request to Sxodim
	client := &http.Client{
		Timeout: HTTPTimeout,
	}
	response, err := client.Do(request)
	if err != nil {
		return nil, nil, v1.ErrorInternal("unable to do sxodim request: %s", err)
	}

	// Cleanup function to close response.Body
	cleanup := func() {
		response.Body.Close()
	}

	// Check status
	if response.StatusCode != http.StatusOK {
		cleanup()
		return nil, nil, v1.ErrorInternal(
			"failed on request to Sxodim: %s",
			response.Status,
		)
	}

	return response, cleanup, nil
}

// exchangeSxodimToken exchanges an authorization code for an OAuth2 token via Sxodim's API.
//
// Notes:
//   - Requires "sxodimclientsecret" to be properly configured in the secret store
//   - Requires SxodimAuthUrl is preconfigured with the correct endpoint
//   - Converts ExpiresIn from milliseconds to time.Time expiry
func (g *SxodimGateway) exchangeSxodimToken(ctx context.Context, authCode string) (*xoauth2.Token, error) {
	// Collect body
	body := sxodimAuthRequestBody{
		GrantType:    SxodimGrantType,
		ClientID:     SxodimClientID,
		ClientSecret: g.storage.ClientSecret,
		Code:         authCode,
	}

	// Marshal body to []byte
	requestBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	// Initialize http request
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, g.storage.SxodimDomain+SxodimAuthUrl, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, v1.ErrorInternal("failed on creating http request: %s", err)
	}

	// Do request to Sxodim
	response, cleanup, err := g.doSxodimRequest(ctx, request)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	// Decode response
	reply := &sxodimAuthResponseBody{}
	err = json.NewDecoder(response.Body).Decode(reply)
	if err != nil {
		return nil, iam_v1.ErrorInternal("unable to decode response: %s", err)
	}

	// Calculate expire date
	expireDate := time.Now().Add(time.Duration(reply.ExpiresIn) * time.Second)

	// Collect token
	token := &xoauth2.Token{
		AccessToken:  reply.AccessToken,
		TokenType:    reply.TokenType,
		RefreshToken: reply.RefreshToken,
		Expiry:       expireDate,
	}

	return token, nil
}

// getUserData is to get data of the Sxodim user.
func (g *SxodimGateway) getUserData(ctx context.Context, accessToken string) (*sxodimUser, error) {
	// Initialize http request
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, g.storage.SxodimDomain+SxodimUserDataUrl, nil)
	if err != nil {
		return nil, v1.ErrorInternal("failed on creating http request: %s", err)
	}
	request.Header.Set("Authorization", "Bearer "+accessToken)
	request.Header.Set("Sxodim-Secret-Aigenda", g.storage.AigendaSecret)

	// Do request to Sxodim
	response, cleanup, err := g.doSxodimRequest(ctx, request)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	// Decode response
	reply := &sxodimUserDataReply{}
	err = json.NewDecoder(response.Body).Decode(reply)
	if err != nil {
		return nil, iam_v1.ErrorInternal("unable to decode response: %s", err)
	}

	return &reply.User, nil
}

func (g *SxodimGateway) Authenticate(ctx context.Context, actorID int64, authCode string) (*CredentialDto, error) {
	token, err := g.exchangeSxodimToken(ctx, authCode)
	if err != nil {
		return nil, err
	}

	user, err := g.getUserData(ctx, token.AccessToken)
	if err != nil {
		return nil, err
	}

	dto := &CredentialDto{
		UserID:         actorID,
		ExternalUserID: &user.ID,
		Provider:       u_struc.Sxodim,
		Token:          token,
		Email:          user.Email,
		Phone:          user.Phone,
		DisplayName:    user.Name,
	}

	return dto, nil
}

func (g *SxodimGateway) RefreshToken(
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
		newToken, err := g.exchangeSxodimToken(ctx, dto.Token.RefreshToken)
		if err != nil {
			return nil, err
		}

		dto.Token = newToken
	}

	// Get Sxodim user data
	user, err := g.getUserData(ctx, dto.Token.AccessToken)
	if err != nil {
		return nil, err
	}

	// Update Sxodim user data
	dto.ExternalUserID = &user.ID
	dto.Email = user.Email
	dto.Phone = user.Phone
	dto.DisplayName = user.Name

	return dto, nil
}

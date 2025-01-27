package biz

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	iam_v1 "gitlab.calendaria.team/services/iam/api/iam/v1"
	"gitlab.calendaria.team/services/iam/ent"
	"gitlab.calendaria.team/services/iam/internal/data"
	"gitlab.calendaria.team/services/utils/v1/config"
	u_jwt "gitlab.calendaria.team/services/utils/v2/jwt"
	u_nats "gitlab.calendaria.team/services/utils/v2/nats"
	u_struc "gitlab.calendaria.team/services/utils/v2/struc"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/mitchellh/mapstructure"
	xoauth2 "golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"
)

type CredentialsUsecase struct {
	config          *config.Config
	log             *log.Helper
	queue           u_nats.IQueueManager
	jwt             u_jwt.IJwtProcessor
	credentialsRepo data.CredentialsRepo
}

func NewCredentialsUsecase(
	config *config.Config,
	logger log.Logger,
	queue u_nats.IQueueManager,
	jwt u_jwt.IJwtProcessor,
	credentialsRepo data.CredentialsRepo,
) (*CredentialsUsecase, error) {
	return &CredentialsUsecase{
		config:          config,
		log:             log.NewHelper(logger),
		jwt:             jwt,
		queue:           queue,
		credentialsRepo: credentialsRepo,
	}, nil
}

func (uc *CredentialsUsecase) AuthByGoogle(ctx context.Context, actorID int64, authCode string) error {
	// get google credentials
	mapGoogleCredentials, err := uc.config.ReadGlobalSecretsFor(ctx, "gwebcredentials")
	if err != nil {
		return iam_v1.ErrorServiceFailed("Unable to read client secret from vault: %v", err.Error())
	}

	// decode google credentials
	googleCredentials := ""
	err = mapstructure.Decode(mapGoogleCredentials["data"], &googleCredentials)
	if err != nil {
		return iam_v1.ErrorServiceFailed("Unable to decode client secret: %v", err.Error())
	}

	// get config from credentials
	googleConfig, err := google.ConfigFromJSON(
		[]byte(googleCredentials),
		oauth2.UserinfoProfileScope,
		oauth2.UserinfoEmailScope,
	)
	if err != nil {
		return iam_v1.ErrorServiceFailed("Unable to parse client secret file to config: %v", err.Error())
	}

	// exchange auth code to token
	tok, err := googleConfig.Exchange(ctx, authCode)
	if err != nil {
		return iam_v1.ErrorServiceFailed("Unable to retrieve token from web: %v", err.Error())
	}

	// get http client
	client := googleConfig.Client(ctx, tok)

	oauth2Service, err := oauth2.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return iam_v1.ErrorServiceFailed("Unable to retrieve OAuth2 service: %v", err.Error())
	}

	userInfoService := oauth2.NewUserinfoV2MeService(oauth2Service)
	userInfo, err := userInfoService.Get().Do()
	if err != nil {
		return iam_v1.ErrorServiceFailed("Unable to retrieve user info: %v", err.Error())
	}

	// save tokens to database
	dto := data.CredentialDto{
		UserID:      actorID,
		DisplayName: userInfo.Name,
		Email:       userInfo.Email,
		Provider:    u_struc.Google,
		Token:       tok,
	}
	err = uc.credentialsRepo.CreateCredential(ctx, dto)
	if err != nil {
		return iam_v1.ErrorDatabaseQuery("database error: %s", err.Error())
	}

	return nil
}

func (uc *CredentialsUsecase) AuthBySxodim(ctx context.Context, actorID int64, authCode, scope string) error {
	// exchange token with auth code
	token, err := uc.getSxodimToken(ctx, authCode)
	if err != nil {
		return iam_v1.ErrorInternal("error on exchanging Sxodim token: %s", err)
	}

	// save tokens to database
	dto := data.CredentialDto{
		UserID:   actorID,
		Provider: u_struc.Sxodim,
		Token:    token,
	}
	err = uc.credentialsRepo.CreateCredential(ctx, dto)
	if err != nil {
		return iam_v1.ErrorDatabaseQuery("database error: %s", err.Error())
	}

	return nil
}

func (uc *CredentialsUsecase) refreshGoogleCredential(
	ctx context.Context,
	credential *ent.UserCredentials,
) (*xoauth2.Token, error) {
	// get google credentials
	mapGoogleCredentials, err := uc.config.ReadGlobalSecretsFor(ctx, "gwebcredentials")
	if err != nil {
		return nil, fmt.Errorf("unable to read client secret from vault: %w", err)
	}

	// decode google credentials
	googleCredentials := ""
	err = mapstructure.Decode(mapGoogleCredentials["data"], &googleCredentials)
	if err != nil {
		return nil, fmt.Errorf("unable to decode client secret: %w", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON([]byte(googleCredentials), calendar.CalendarScope)
	if err != nil {
		return nil, fmt.Errorf("unable to parse client secret file to config: %w", err)
	}

	// collect token
	token := &xoauth2.Token{
		AccessToken: credential.AccessToken,
	}
	if credential.TokenType != nil {
		token.TokenType = *credential.TokenType
	}
	if credential.RefreshToken != nil {
		token.RefreshToken = *credential.RefreshToken
	}
	if credential.ExpiresAt != nil {
		token.Expiry = credential.ExpiresAt.UTC()
	}

	// refresh token if expired
	if !token.Valid() {
		newToken, err2 := config.TokenSource(ctx, token).Token()
		if err2 != nil {
			return nil, fmt.Errorf("unable to refresh token: %w", err2)
		}

		token = newToken
	}

	return token, nil
}

// getSxodimToken exchanges an authorization code for an Sxodim OAuth2 token, using client credentials and API endpoint.
// for refresh token also uses auth API with requesting refresh token instead of authorization code.
func (uc *CredentialsUsecase) getSxodimToken(
	ctx context.Context,
	authCode string,
) (*xoauth2.Token, error) {
	// get callback url
	sxodimAuthURL, err := uc.config.Value("sxodimauthurl").String()
	if err != nil {
		return nil, err
	}

	// get sxodim client secret
	sxodimConfig, err := uc.config.ReadSecretsFor(ctx, "sxodimclientsecret")
	if err != nil {
		return nil, iam_v1.ErrorInternal("failed getting sxodim client secret: %s", err)
	}

	// Validate and retrieve the Sxodim client secret from configuration
	sxodimClientSecret, ok := sxodimConfig["secret"].(string)
	if !ok {
		return nil, iam_v1.ErrorInternal("sxodim client secret is not set: %s", err)
	}

	// collect body
	body := SxodimAuthRequestBody{
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
	res, err := http.Post(sxodimAuthURL, "application/json", bytes.NewBuffer(requestBody))
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
	response := &SxodimAuthResponseBody{}
	err = json.NewDecoder(res.Body).Decode(response)
	if err != nil {
		return nil, iam_v1.ErrorInternal("unable to decode response: %s", err)
	}

	// calculate expire date
	expireDate := time.Now().Add(time.Duration(response.ExpiresIn) * time.Millisecond)

	// collect token
	token := &xoauth2.Token{
		AccessToken:  response.AccessToken,
		TokenType:    response.TokenType,
		RefreshToken: response.RefreshToken,
		Expiry:       expireDate,
	}

	return token, nil
}

func (uc *CredentialsUsecase) RefreshCredential(
	ctx context.Context,
	actorID, credentialID int64,
) (*ent.UserCredentials, error) {
	credential, err := uc.credentialsRepo.GetCredential(ctx, actorID, credentialID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, iam_v1.ErrorCredentialNotFound("credential not found")
		}
		return nil, iam_v1.ErrorDatabaseQuery("database error: %s", err.Error())
	}

	// validate credential
	if credential == nil || credential.Provider == nil {
		return nil, iam_v1.ErrorInternal("credential don't have provider")
	}

	// initialize token
	token := &xoauth2.Token{}

	// refresh token by provider
	switch *credential.Provider {
	case u_struc.Google:
		token, err = uc.refreshGoogleCredential(ctx, credential)
		if err != nil {
			return nil, err
		}
	case u_struc.Sxodim:
		if credential.RefreshToken == nil {
			return nil, iam_v1.ErrorInternal("credential doesn't have refresh token")
		}

		token, err = uc.getSxodimToken(ctx, *credential.RefreshToken)
		if err != nil {
			return nil, err
		}
	}

	updCredential := data.CredentialDto{
		Token: token,
	}

	return uc.credentialsRepo.UpdateCredential(ctx, credentialID, updCredential)
}

func (uc *CredentialsUsecase) GetCredential(
	ctx context.Context,
	actorID, credentialID int64,
) (*ent.UserCredentials, error) {
	return uc.credentialsRepo.GetCredential(ctx, actorID, credentialID)
}

func (uc *CredentialsUsecase) ListCredentials(
	ctx context.Context,
	actorID int64,
	provider *u_struc.Provider,
) ([]*ent.UserCredentials, error) {
	return uc.credentialsRepo.ListCredentials(ctx, actorID, provider)
}

func (uc *CredentialsUsecase) DeleteCredential(ctx context.Context, actorID, credentialID int64) error {
	deletedCount, err := uc.credentialsRepo.DeleteCredential(ctx, actorID, credentialID)
	if err != nil {
		return iam_v1.ErrorDatabaseQuery("database error: %s", err.Error())
	} else if deletedCount == 0 {
		return iam_v1.ErrorCredentialNotFound("credential not found")
	}

	return nil
}

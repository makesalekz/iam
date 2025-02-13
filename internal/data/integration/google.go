package integration

import (
	"context"
	"fmt"

	iam_v1 "gitlab.calendaria.team/services/iam/api/iam/v1"
	"gitlab.calendaria.team/services/iam/ent"
	u_struc "gitlab.calendaria.team/services/utils/v2/struc"
	"gitlab.calendaria.team/services/utils/v4/config"

	"github.com/mitchellh/mapstructure"
	xoauth2 "golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"
)

type GoogleGateway struct {
	config config.IConfig
}

func NewGoogleRemote(
	config config.IConfig,
) (IProviderGateway, error) {
	return &GoogleGateway{
		config: config,
	}, nil
}

// getGoogleConfig loads and initializes Google OAuth2 configuration from secure storage.
//
// This method handles the full lifecycle of retrieving and preparing Google API credentials:
// 1. Secure retrieval of client credentials from vault
// 2. Credential decoding and validation
// 3. OAuth2 configuration initialization with required scopes
//
// Scopes:
//   - oauth2.UserinfoProfileScope: Required for basic profile information
//   - oauth2.UserinfoEmailScope: Required for email address access
//   - calendar.CalendarScope: Required for Google calendar access
//
// Security:
//   - Requires "gwebcredentials" secret containing valid Google OAuth2 JSON credentials
//   - Never exposes raw client secrets in output or errors
//   - Relies on config reader implementation for secure secret storage
//
// Notes:
//   - Uses mapstructure for credential decoding to handle secret storage format
//   - Requires credentials JSON to follow Google's OAuth2 client ID format
//   - Maintains compatibility with Google's oauth2 package expectations
func (r *GoogleGateway) getGoogleConfig(ctx context.Context) (*xoauth2.Config, error) {
	// get google credentials
	mapGoogleCredentials, err := r.config.ReadGlobalSecretsFor(ctx, "gwebcredentials")
	if err != nil {
		return nil, iam_v1.ErrorServiceFailed("Unable to read client secret from vault: %v", err.Error())
	}

	// decode google credentials
	googleCredentials := ""
	err = mapstructure.Decode(mapGoogleCredentials["data"], &googleCredentials)
	if err != nil {
		return nil, iam_v1.ErrorServiceFailed("Unable to decode client secret: %v", err.Error())
	}

	// get config from credentials
	googleConfig, err := google.ConfigFromJSON(
		[]byte(googleCredentials),
		oauth2.UserinfoProfileScope,
		oauth2.UserinfoEmailScope,
		calendar.CalendarScope,
	)
	if err != nil {
		return nil, iam_v1.ErrorServiceFailed("Unable to parse client secret file to config: %v", err.Error())
	}

	return googleConfig, nil
}

func (r *GoogleGateway) Authenticate(ctx context.Context, actorID int64, authCode string) (*CredentialDto, error) {
	// Get Google config
	googleConfig, err := r.getGoogleConfig(ctx)
	if err != nil {
		return nil, err
	}

	// exchange auth code to token
	token, err := googleConfig.Exchange(ctx, authCode)
	if err != nil {
		return nil, iam_v1.ErrorServiceFailed("Unable to retrieve token from web: %v", err.Error())
	}

	// get http client
	client := googleConfig.Client(ctx, token)

	oauth2Service, err := oauth2.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, iam_v1.ErrorServiceFailed("Unable to retrieve OAuth2 service: %v", err.Error())
	}

	userInfoService := oauth2.NewUserinfoV2MeService(oauth2Service)
	userInfo, err := userInfoService.Get().Do()
	if err != nil {
		return nil, iam_v1.ErrorServiceFailed("Unable to retrieve user info: %v", err.Error())
	}

	// save tokens to database
	dto := &CredentialDto{
		UserID:      actorID,
		DisplayName: userInfo.Name,
		Email:       userInfo.Email,
		Provider:    u_struc.Google,
		Token:       token,
	}

	return dto, nil
}

func (r *GoogleGateway) RefreshToken(
	ctx context.Context,
	credential *ent.UserCredentials,
) (*CredentialDto, error) {
	// Retrieve dto from credential
	dto := CredentialToDto(credential)

	// Check valid token and refresh token existence
	if dto == nil || dto.Token == nil || dto.Token.RefreshToken == "" {
		return nil, iam_v1.ErrorNeedReauthorization("invalid token or null refresh token, need reauthorization")
	}

	// Refresh token if expired
	if !dto.Token.Valid() {
		// Get Google config
		googleConfig, err := r.getGoogleConfig(ctx)
		if err != nil {
			return nil, err
		}

		// Refresh a new token
		newToken, err2 := googleConfig.TokenSource(ctx, dto.Token).Token()
		if err2 != nil {
			return nil, fmt.Errorf("unable to refresh token: %w", err2)
		}

		dto.Token = newToken
	}

	return dto, nil
}

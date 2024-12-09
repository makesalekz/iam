package biz

import (
	"context"

	iam_v1 "gitlab.calendaria.team/services/iam/api/iam/v1"
	"gitlab.calendaria.team/services/iam/ent"
	"gitlab.calendaria.team/services/iam/internal/data"
	"gitlab.calendaria.team/services/utils/v1/config"
	u_jwt "gitlab.calendaria.team/services/utils/v2/jwt"
	u_nats "gitlab.calendaria.team/services/utils/v2/nats"
	u_struc "gitlab.calendaria.team/services/utils/v2/struc"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/mitchellh/mapstructure"
	"golang.org/x/oauth2/google"
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

package biz

import (
	"context"

	iam_v1 "gitlab.calendaria.team/services/iam/api/iam/v1"
	"gitlab.calendaria.team/services/iam/ent"
	"gitlab.calendaria.team/services/iam/internal/data"
	"gitlab.calendaria.team/services/utils/v1/config"
	u_nats "gitlab.calendaria.team/services/utils/v1/nats"
	u_jwt "gitlab.calendaria.team/services/utils/v2/jwt"
	u_struc "gitlab.calendaria.team/services/utils/v2/struc"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/mitchellh/mapstructure"
	"golang.org/x/oauth2/google"
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
		queue:           queue,
		jwt:             jwt,
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
	config, err := google.ConfigFromJSON([]byte(googleCredentials))
	if err != nil {
		return iam_v1.ErrorServiceFailed("Unable to parse client secret file to config: %v", err.Error())
	}

	// exchange auth code to token
	tok, err := config.Exchange(ctx, authCode)
	if err != nil {
		return iam_v1.ErrorServiceFailed("Unable to retrieve token from web: %v", err.Error())
	}

	// save tokens to database
	_, err = uc.credentialsRepo.CreateCredential(ctx, actorID, tok)
	if err != nil {
		return iam_v1.ErrorDatabaseQuery("database error: %s", err.Error())
	}

	return nil
}

func (uc *CredentialsUsecase) GetCredential(
	ctx context.Context, actorID int64, provider u_struc.Provider,
) (*ent.UserCredentials, error) {
	return uc.credentialsRepo.GetCredential(ctx, actorID, provider)
}

func (uc *CredentialsUsecase) ListCredentials(ctx context.Context, actorID int64) ([]*ent.UserCredentials, error) {
	return uc.credentialsRepo.ListCredentials(ctx, actorID)
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

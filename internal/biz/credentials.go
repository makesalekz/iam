package biz

import (
	"context"
	"gitlab.calendaria.team/services/iam/ent"
	"gitlab.calendaria.team/services/iam/ent/enum"
	"os"

	iam_v1 "gitlab.calendaria.team/services/iam/api/iam/v1"
	"gitlab.calendaria.team/services/iam/internal/data"
	u_jwt "gitlab.calendaria.team/services/utils/v1/jwt"
	u_nats "gitlab.calendaria.team/services/utils/v1/nats"

	"github.com/go-kratos/kratos/v2/log"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
)

type CredentialsUsecase struct {
	log             *log.Helper
	queue           *u_nats.QueueManager
	jwt             *u_jwt.JwtProcessor
	credentialsRepo data.CredentialsRepo
}

func NewCredentialsUsecase(
	logger log.Logger,
	queue *u_nats.QueueManager,
	jwt *u_jwt.JwtProcessor,
	credentialsRepo data.CredentialsRepo,
) (*CredentialsUsecase, error) {
	return &CredentialsUsecase{
		log:             log.NewHelper(logger),
		queue:           queue,
		jwt:             jwt,
		credentialsRepo: credentialsRepo,
	}, nil
}

func (uc *CredentialsUsecase) AuthByGoogle(ctx context.Context, actorId int64, authCode string) error {
	// read credentials, change path to your google credentials file
	b, err := os.ReadFile("../configs/credentials_test.json")
	if err != nil {
		return iam_v1.ErrorServiceFailed("Unable to read client secret file: %v", err.Error())
	}

	// get config from credentials
	config, err := google.ConfigFromJSON(b, calendar.CalendarScope)
	if err != nil {
		return iam_v1.ErrorServiceFailed("Unable to parse client secret file to config: %v", err.Error())
	}

	// exchange auth code to token
	tok, err := config.Exchange(ctx, authCode)
	if err != nil {
		return iam_v1.ErrorServiceFailed("Unable to retrieve token from web: %v", err.Error())
	}

	// save tokens to database
	_, err = uc.credentialsRepo.CreateCredential(ctx, actorId, tok)
	if err != nil {
		return iam_v1.ErrorDatabaseQuery("database error: %s", err.Error())
	}

	return nil
}

func (uc *CredentialsUsecase) GetCredential(ctx context.Context, actorId int64, provider enum.Provider) (*ent.UserCredentials, error) {
	return uc.credentialsRepo.GetCredential(ctx, actorId, provider)
}

func (uc *CredentialsUsecase) ListCredentials(ctx context.Context, actorId int64) ([]*ent.UserCredentials, error) {
	return uc.credentialsRepo.ListCredentials(ctx, actorId)
}

func (uc *CredentialsUsecase) DeleteCredential(ctx context.Context, actorId, credentialId int64) error {
	deletedCount, err := uc.credentialsRepo.DeleteCredential(ctx, actorId, credentialId)
	if err != nil {
		return iam_v1.ErrorDatabaseQuery("database error: %s", err.Error())
	} else if deletedCount == 0 {
		return iam_v1.ErrorCredentialNotFound("credential not found")
	}

	return nil
}

package biz

import (
	"context"

	iam_v1 "gitlab.calendaria.team/services/iam/api/iam/v1"
	"gitlab.calendaria.team/services/iam/ent"
	"gitlab.calendaria.team/services/iam/internal/data"
	"gitlab.calendaria.team/services/iam/internal/data/integration"
	"gitlab.calendaria.team/services/utils/v1/config"
	u_jwt "gitlab.calendaria.team/services/utils/v2/jwt"
	u_nats "gitlab.calendaria.team/services/utils/v2/nats"
	u_struc "gitlab.calendaria.team/services/utils/v2/struc"

	"github.com/go-kratos/kratos/v2/log"
)

type CredentialsUsecase struct {
	isTesting       bool
	config          *config.Config
	log             *log.Helper
	queue           u_nats.IQueueManager
	jwt             u_jwt.IJwtProcessor
	provider        integration.IProviderManager
	credentialsRepo data.CredentialsRepo
}

func NewCredentialsUsecase(
	isTesting bool,
	config *config.Config,
	logger log.Logger,
	queue u_nats.IQueueManager,
	jwt u_jwt.IJwtProcessor,
	provide integration.IProviderManager,
	credentialsRepo data.CredentialsRepo,
) (*CredentialsUsecase, error) {
	return &CredentialsUsecase{
		isTesting:       isTesting,
		config:          config,
		log:             log.NewHelper(logger),
		jwt:             jwt,
		queue:           queue,
		provider:        provide,
		credentialsRepo: credentialsRepo,
	}, nil
}

func (uc *CredentialsUsecase) ExternalAuth(
	ctx context.Context,
	actorID int64,
	provider u_struc.Provider,
	authCode string,
) error {
	providerGateway, err := uc.provider.NewProviderGateway(provider)
	if err != nil {
		return err
	}

	if uc.isTesting {
		return nil
	}

	// Exchange token
	credentialDto, err := providerGateway.Authenticate(ctx, actorID, authCode)
	if err != nil {
		return err
	}

	// Check credential existence
	credential, err := uc.credentialsRepo.GetCredentialByMail(ctx, credentialDto.Email)
	if err != nil && !ent.IsNotFound(err) {
		return iam_v1.ErrorDatabaseQuery("Unable to get credential: %v", err.Error())
	}

	if credential != nil && credential.UserID != actorID {
		return iam_v1.ErrorForbidden("This email address is already in use by another user")
	}

	// Save credential to database
	err = uc.credentialsRepo.CreateCredential(ctx, *credentialDto)
	if err != nil {
		return iam_v1.ErrorDatabaseQuery("database error: %s", err.Error())
	}

	return nil
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

	providerGateway, err := uc.provider.NewProviderGateway(*credential.Provider)
	if err != nil {
		return nil, err
	}

	if uc.isTesting {
		return nil, nil
	}

	// Exchange token
	credentialDto, err := providerGateway.RefreshToken(ctx, credential)
	if err != nil {
		return nil, err
	}

	return uc.credentialsRepo.UpdateCredential(ctx, credentialID, *credentialDto)
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

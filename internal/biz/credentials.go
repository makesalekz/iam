package biz

import (
	"context"

	iam_v1 "gitlab.calendaria.team/services/iam/api/iam/v1"
	"gitlab.calendaria.team/services/iam/ent"
	"gitlab.calendaria.team/services/iam/internal/data"
	"gitlab.calendaria.team/services/iam/internal/data/integration"
	u_struc "gitlab.calendaria.team/services/utils/v2/struc"
	"gitlab.calendaria.team/services/utils/v4/config"
	u_jwt "gitlab.calendaria.team/services/utils/v4/jwt"
	u_nats "gitlab.calendaria.team/services/utils/v4/nats"

	"github.com/go-kratos/kratos/v2/log"
)

type CredentialsUsecase struct {
	log             *log.Helper
	config          config.IConfig
	queue           u_nats.IQueueManager
	jwt             u_jwt.IJwtProcessor
	provider        integration.IProviderManager
	credentialsRepo data.CredentialsRepo
}

func NewCredentialsUsecase(
	config config.IConfig,
	logger log.Logger,
	queue u_nats.IQueueManager,
	jwt u_jwt.IJwtProcessor,
	provide integration.IProviderManager,
	credentialsRepo data.CredentialsRepo,
) (*CredentialsUsecase, error) {
	return &CredentialsUsecase{
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
) (*ent.UserCredentials, error) {
	providerGateway, err := uc.provider.NewProviderGateway(provider)
	if err != nil {
		return nil, err
	}

	// Exchange token
	credentialDto, err := providerGateway.Authenticate(ctx, actorID, authCode)
	if err != nil {
		return nil, err
	}

	// Check mail credential existence
	credential := &ent.UserCredentials{}
	if credentialDto.Email != "" {
		// Check credential existence
		credential, err = uc.credentialsRepo.GetCredentialByMail(ctx, credentialDto.Email, provider)
		if err != nil && !ent.IsNotFound(err) {
			if ent.IsNotSingular(err) {
				uc.log.Errorf("multi credential in one e-mail (%s) provider: %v", credentialDto.Email, err)
			}

			return nil, iam_v1.ErrorDatabaseQuery("Unable to get credential: %v", err.Error())
		}

		if credential != nil && credential.UserID != actorID {
			return nil, iam_v1.ErrorCredentialsAlreadyInUse("This email address is already in use by another user")
		}
	}

	// Save new credential to db
	userCredential := &ent.UserCredentials{}
	if ent.IsNotFound(err) {
		// Save credential to database
		userCredential, err = uc.credentialsRepo.CreateCredential(ctx, *credentialDto)
		if err != nil {
			return nil, iam_v1.ErrorDatabaseQuery("Unable to create credential: %v", err.Error())
		}
	} else if credential != nil {
		// Update credential to database
		userCredential, err = uc.credentialsRepo.UpdateCredential(ctx, credential.ID, *credentialDto)
		if err != nil {
			return nil, iam_v1.ErrorDatabaseQuery("Unable to create credential: %v", err.Error())
		}
	}

	return userCredential, nil
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

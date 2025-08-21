package biz

import (
	"context"

	iam_v1 "gitlab.calendaria.team/services/iam/api/iam/v1"
	"gitlab.calendaria.team/services/iam/ent"
	"gitlab.calendaria.team/services/iam/ent/mixins"
	"gitlab.calendaria.team/services/iam/internal/data"
	"gitlab.calendaria.team/services/iam/internal/data/errors"
	"gitlab.calendaria.team/services/iam/internal/data/integration"
	"gitlab.calendaria.team/services/iam/internal/data/remote"
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
	events          remote.IEventsRemote
	credentialsRepo data.CredentialsRepo
}

func NewCredentialsUsecase(
	config config.IConfig,
	logger log.Logger,
	queue u_nats.IQueueManager,
	jwt u_jwt.IJwtProcessor,
	provide integration.IProviderManager,
	events remote.IEventsRemote,
	credentialsRepo data.CredentialsRepo,
) (*CredentialsUsecase, error) {
	return &CredentialsUsecase{
		config: config,
		log: log.NewHelper(
			log.With(logger, "ts", log.DefaultTimestamp, "module", "usecase/credentials"),
		),
		jwt:             jwt,
		queue:           queue,
		provider:        provide,
		events:          events,
		credentialsRepo: credentialsRepo,
	}, nil
}

func (uc *CredentialsUsecase) ExternalAuth(
	ctx context.Context,
	actorID int64,
	provider u_struc.Provider,
	authCode string,
) (*ent.UserCredentials, error) {
	// Input validation
	if authCode == "" {
		return nil, iam_v1.ErrorInvalidRequest("auth code is required")
	}

	// Create provider gateway
	providerGateway, err := uc.provider.NewProviderGateway(provider)
	if err != nil {
		return nil, iam_v1.ErrorInvalidProvider("provider gateway creation failed: %v", err)
	}

	// Step 1: Check if the current user already has a credential for this provider
	existingCredential, err := uc.credentialsRepo.GetCredentialByProvider(ctx, actorID, provider)
	if err != nil && !ent.IsNotFound(err) {
		if ent.IsNotSingular(err) {
			uc.log.Errorf("multiple credentials found for userID (%v) provider %s: %v",
				actorID, provider, err)
			return nil, iam_v1.ErrorDatabaseQuery("multiple credentials found for this email")
		}

		return nil, iam_v1.ErrorDatabaseQuery("error checking existing credentials: %v", err)
	}

	// Step 2: Exchange the token
	credentialDto, err := providerGateway.Authenticate(actorID, authCode)
	if err != nil {
		return nil, errors.MapToExternalError(err)
	}

	// If user has existing credential with same email - update it
	if existingCredential != nil && existingCredential.Mail != nil && *existingCredential.Mail == credentialDto.Email {
		refreshedCredentialDto, err2 := providerGateway.RefreshToken(existingCredential)
		if err2 != nil {
			return nil, errors.MapToExternalError(err2)
		}

		return uc.credentialsRepo.UpdateCredential(ctx, existingCredential.ID, *refreshedCredentialDto)
	}

	// Revoke the token if we can't create it
	var needsRevocation bool
	defer func() {
		if needsRevocation && credentialDto != nil && credentialDto.Token != nil {
			err2 := providerGateway.RevokeToken(&ent.UserCredentials{
				AccessToken:  credentialDto.Token.AccessToken,
				RefreshToken: &credentialDto.Token.RefreshToken,
			})
			if err2 != nil {
				uc.log.Errorf("failed to revoke token for mail (%s)", credentialDto.Email)
			}
		}
	}()

	// Step 3: Check if the email is already in use by ANY user
	emailCredential, err := uc.credentialsRepo.GetCredentialByMail(ctx, credentialDto.Email, provider)
	if err != nil && !ent.IsNotFound(err) {
		needsRevocation = true
		if ent.IsNotSingular(err) {
			uc.log.Errorf("multiple credentials found for email (%s) provider %s: %v",
				credentialDto.Email, provider, err)
			return nil, iam_v1.ErrorDatabaseQuery("multiple credentials found for this email")
		}

		return nil, iam_v1.ErrorDatabaseQuery("error querying credentials: %v", err)
	}

	// If email belongs to another user - reject the auth attempt.
	// Don't revoke another user's token.
	if emailCredential != nil && emailCredential.UserID != actorID {
		return nil, iam_v1.ErrorCredentialsAlreadyInUse("this email address is already in use by another user")
	}

	// User has existing credential with different email - reject the auth attempt
	if existingCredential != nil && existingCredential.Mail != nil && *existingCredential.Mail != credentialDto.Email {
		needsRevocation = true
		return nil, iam_v1.ErrorForbidden("you already have an active credential")
	}

	// If there is a no existing credential for user with this email and provider - create new one
	userCredential, err := uc.credentialsRepo.CreateCredential(ctx, *credentialDto)
	if err != nil {
		needsRevocation = true
		return nil, iam_v1.ErrorDatabaseQuery("failed to create credential: %v", err)
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
	credentialDto, err := providerGateway.RefreshToken(credential)
	if err != nil {
		return nil, errors.MapToExternalError(err)
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
	credential, err := uc.credentialsRepo.GetCredential(ctx, actorID, credentialID)
	if err != nil {
		if ent.IsNotFound(err) {
			return iam_v1.ErrorCredentialNotFound("credential not found")
		}
		return iam_v1.ErrorDatabaseQuery("database error: %s", err.Error())
	}

	// Disconnect calendars from deleted credential
	err = uc.events.DisconnectExternalCalendarsBulk(ctx, credential.ID)
	if err != nil {
		return err
	}

	// Delete credential from db
	_, err = uc.credentialsRepo.DeleteCredential(mixins.SkipSoftDelete(ctx), actorID, credentialID)
	if err != nil {
		return iam_v1.ErrorDatabaseQuery("database error: %s", err.Error())
	}

	// Revoke token after deletion
	if credential.Provider != nil {
		providerGateway, err2 := uc.provider.NewProviderGateway(*credential.Provider)
		if err2 != nil {
			uc.log.Errorf("provider gateway creation failed: %v", err2)
		}

		err = providerGateway.RevokeToken(credential)
		if err != nil {
			uc.log.Warnf("failed to revoke token: %v (user_id: %d, credential_id: %d)",
				err, actorID, credentialID)
		}
	}

	return nil
}

package test_test

import (
	"context"
	"testing"

	v1 "gitlab.calendaria.team/services/iam/api/iam/v1"
	"gitlab.calendaria.team/services/iam/ent"
	"gitlab.calendaria.team/services/iam/internal/data/integration"
	"gitlab.calendaria.team/services/iam/internal/service"
	u_struc "gitlab.calendaria.team/services/utils/v2/struc"

	"github.com/stretchr/testify/require"
)

const (
	provider1   = u_struc.Google
	authCode    = "def502006c1eebbf4d4033d2aa3930a37d944ab044d97b39041e64e3e5df18d3c779095796dcda372db53aaf79e9faec6a578ad3e1c4530d67c8d30bd605376c4a2e1dc22bbf215ca66147df1ee528a1a7d09910123257aabee6e7434d0ab0e013009600762f589e2b1ddd07c6ef1a2d6637d69a213c366114b20fcb572966f91454e256a364baaf9b2c36c83b3e3e1ffed0a40d3b149573e5080d8e51c8097b516d077d4e180d1daed07fb414d1f997ba04a5ba8059d47e8e7541140c38b59e89044e5c61fcc684a52a5cd5332cc79daee54f0e30ac853fcab8b761df144680dfebb73aeecd0547cbf78276c7e3f6c4668097ee316e1bcfb777c5e4c41f00321ec277d54bb46bb83f5a9e6518b6e50ce561caf7d3328d741e70bf0abfb704d12c86893c27ff6953dcad46819ca83b2956c8e216fdb7ecef0b27d4cbc2b5f9c8e6e66a21"
	accessToken = "ya29.a0AXeO80QFDXTORxWPZR5fUUwax7gxo6KtQki8IUWoMhkrMYPgmSZ6oPXP_AvQHBcVXl41YCJcRDwtfDnL6V3fwCUNkyO8WTkJTN552_1hja5McmueztuvX4U8e6WsJ4BxnA6g95yuaubw2UIw8tz-B4j7dCpnTLdptGFt6C2W-QaCgYKAYgSARESFQHGX2Mi1BkJUfenNGB92jLxqlLCIQ0177"
)

func TestCredentialsService_ExternalAuth_Success(t *testing.T) {
	// Set up the test context, mocks, and credentialService.
	ctx, repo, credentialService := createCredentialsService(t)
	// Ensure that actorID is properly set in the context (e.g., via mockServerContext).
	ids := getIDs() // Assume ids.actorID is set appropriately.

	// Success Case 1: Create credential
	{
		// Define a consistent auth code.
		req := &v1.ExternalAuthRequest{
			AuthCode: authCode,
			Provider: provider1.Value(),
		}

		// --- Expectations ---
		// Expect that the credentialService calls the provider mock to create a gateway.
		repo.provider.EXPECT().
			NewProviderGateway(provider1).
			Return(repo.providerGateway, nil)

		// Simulate the scenario where the credential is not found in the repository.
		repo.credentialsRepo.EXPECT().
			GetCredentialByProvider(ctx, ids.actorID, provider1).
			Return(nil, &ent.NotFoundError{})

		// Expect that the Authenticate method is called with the actorID and auth code.
		credentialDto := &integration.CredentialDto{
			UserID: ids.userID,
			Email:  "user@example.com",
		}
		repo.providerGateway.EXPECT().
			Authenticate(ids.actorID, authCode).
			Return(credentialDto, nil)

		// Simulate the scenario where the credential is not found in the repository.
		repo.credentialsRepo.EXPECT().
			GetCredentialByMail(ctx, credentialDto.Email, provider1).
			Return(nil, &ent.NotFoundError{})

		// Expect that a new credential is created successfully.
		provider := provider1
		userCredentials := &ent.UserCredentials{
			UserID:      ids.actorID,
			Provider:    &provider,
			AccessToken: accessToken,
		}
		repo.credentialsRepo.EXPECT().
			CreateCredential(ctx, *credentialDto).
			Return(userCredentials, nil)

		// --- Execute ---
		// Call the ExternalAuth method.
		_, err := credentialService.ExternalAuth(ctx, req)
		require.NoError(t, err)
	}

	// Success Case 2: If credential is already exists (update it)
	{
		// Define a consistent auth code.
		req := &v1.ExternalAuthRequest{
			AuthCode: authCode,
			Provider: provider1.Value(),
		}

		// --- Expectations ---
		// Expect that the credentialService calls the provider mock to create a gateway.
		repo.provider.EXPECT().
			NewProviderGateway(provider1).
			Return(repo.providerGateway, nil)

		provider := provider1
		mail := "user@example.com"
		existingCredential := &ent.UserCredentials{
			UserID:   ids.actorID,
			Provider: &provider,
			Mail:     &mail,
		}
		repo.credentialsRepo.EXPECT().
			GetCredentialByProvider(ctx, ids.actorID, provider1).
			Return(existingCredential, nil)

		// Expect that the Authenticate method is called with the actorID and auth code.
		credentialDto := &integration.CredentialDto{
			UserID: ids.userID,
			Email:  mail,
		}
		repo.providerGateway.EXPECT().
			Authenticate(ids.actorID, authCode).
			Return(credentialDto, nil)

		repo.providerGateway.EXPECT().
			RefreshToken(existingCredential).
			Return(credentialDto, nil)

		// Expect that a new credential is created successfully.
		repo.credentialsRepo.EXPECT().
			UpdateCredential(ctx, existingCredential.ID, *credentialDto).
			Return(existingCredential, nil)

		// --- Execute ---
		// Call the ExternalAuth method.
		_, err := credentialService.ExternalAuth(ctx, req)
		require.NoError(t, err)
	}
}

func TestCredentialsService_ExternalAuth_ErrorCases(t *testing.T) {
	ctx, repo, credentialService := createCredentialsService(t)
	ids := getIDs()

	// Error Case 1: Empty actor id
	{
		req := &v1.ExternalAuthRequest{
			AuthCode: authCode,
			Provider: provider1.Value(),
		}

		// Create a context without an actor id.
		ctxWithoutActor := context.Background()
		result, err := credentialService.ExternalAuth(ctxWithoutActor, req)
		require.Error(t, err)
		require.Nil(t, result)
		require.Equal(t, v1.ErrorEmptyActorId("empty actor id"), err)
	}

	// Error Case 2: Invalid provider
	{
		req := &v1.ExternalAuthRequest{
			AuthCode: authCode,
			Provider: "invalid", // Fails provider.IsValid() check.
		}

		result, err := credentialService.ExternalAuth(ctx, req)
		require.Error(t, err)
		require.Nil(t, result)
		require.Equal(t, v1.ErrorInvalidProvider("invalid provider"), err)
	}

	// Error Case 3: NewProviderGateway Error
	{
		req := &v1.ExternalAuthRequest{
			AuthCode: authCode,
			Provider: provider1.Value(),
		}

		errFunc := v1.ErrorNotFound("unknown provider")
		repo.provider.EXPECT().
			NewProviderGateway(provider1).
			Return(nil, errFunc)

		expectedErr := v1.ErrorInvalidProvider("provider gateway creation failed: %s", errFunc.Error())
		result, err := credentialService.ExternalAuth(ctx, req)
		require.Error(t, err)
		require.Nil(t, result)
		require.Equal(t, expectedErr, err)
	}

	// Error Case 4: Authenticate Error
	{
		req := &v1.ExternalAuthRequest{
			AuthCode: authCode,
			Provider: provider1.Value(),
		}

		repo.provider.EXPECT().
			NewProviderGateway(provider1).
			Return(repo.providerGateway, nil)

		// Simulate the scenario where the credential is not found in the repository.
		repo.credentialsRepo.EXPECT().
			GetCredentialByProvider(ctx, ids.actorID, provider1).
			Return(nil, &ent.NotFoundError{})

		errFunc := v1.ErrorInternal("auth failed")
		repo.providerGateway.EXPECT().
			Authenticate(ids.actorID, authCode).
			Return(nil, errFunc)

		expectedErr := v1.ErrorServiceFailed("service error: %s", errFunc.Error())
		result, err := credentialService.ExternalAuth(ctx, req)
		require.Error(t, err)
		require.Nil(t, result)
		require.Equal(t, expectedErr, err)
	}

	// Error Case 5: GetCredentialByMail Error (non NotFound error)
	{
		req := &v1.ExternalAuthRequest{
			AuthCode: authCode,
			Provider: provider1.Value(),
		}

		repo.provider.EXPECT().
			NewProviderGateway(provider1).
			Return(repo.providerGateway, nil)

		// Simulate the scenario where the credential is not found in the repository.
		repo.credentialsRepo.EXPECT().
			GetCredentialByProvider(ctx, ids.actorID, provider1).
			Return(nil, &ent.NotFoundError{})

		credentialDto := &integration.CredentialDto{
			UserID: ids.userID,
			Email:  "user@example.com",
		}
		repo.providerGateway.EXPECT().
			Authenticate(ids.actorID, authCode).
			Return(credentialDto, nil)

		dbErr := v1.ErrorInternal("db error")
		repo.credentialsRepo.EXPECT().
			GetCredentialByMail(ctx, credentialDto.Email, provider1).
			Return(nil, dbErr)

		expectedErr := v1.ErrorDatabaseQuery("error querying credentials: %s", dbErr.Error())
		result, err := credentialService.ExternalAuth(ctx, req)
		require.Error(t, err)
		require.Nil(t, result)
		require.Equal(t, expectedErr, err)
	}

	// Error Case 6: Credential Conflict (existing credential belongs to a different user)
	{
		req := &v1.ExternalAuthRequest{
			AuthCode: authCode,
			Provider: provider1.Value(),
		}

		repo.provider.EXPECT().
			NewProviderGateway(provider1).
			Return(repo.providerGateway, nil)

		// Simulate the scenario where the credential is not found in the repository.
		repo.credentialsRepo.EXPECT().
			GetCredentialByProvider(ctx, ids.actorID, provider1).
			Return(nil, &ent.NotFoundError{})

		credentialDto := &integration.CredentialDto{
			UserID: ids.userID,
			Email:  "user@example.com",
		}
		repo.providerGateway.EXPECT().
			Authenticate(ids.actorID, authCode).
			Return(credentialDto, nil)

		prov := provider1
		existingCredential := &ent.UserCredentials{
			UserID:      ids.actorID + 1, // Different from ids.actorID.
			Provider:    &prov,
			AccessToken: accessToken,
		}
		repo.credentialsRepo.EXPECT().
			GetCredentialByMail(ctx, credentialDto.Email, provider1).
			Return(existingCredential, nil)

		expectedErr := v1.ErrorCredentialsAlreadyInUse("this email address is already in use by another user")
		result, err := credentialService.ExternalAuth(ctx, req)
		require.Error(t, err)
		require.Nil(t, result)
		require.Equal(t, expectedErr, err)
	}

	// Error Case 7: CreateCredential Error
	{
		req := &v1.ExternalAuthRequest{
			AuthCode: authCode,
			Provider: provider1.Value(),
		}

		repo.provider.EXPECT().
			NewProviderGateway(provider1).
			Return(repo.providerGateway, nil)

		// Simulate the scenario where the credential is not found in the repository.
		repo.credentialsRepo.EXPECT().
			GetCredentialByProvider(ctx, ids.actorID, provider1).
			Return(nil, &ent.NotFoundError{})

		credentialDto := &integration.CredentialDto{
			UserID: ids.actorID,
			Email:  "user@example.com",
		}
		repo.providerGateway.EXPECT().
			Authenticate(ids.actorID, authCode).
			Return(credentialDto, nil)

		repo.credentialsRepo.EXPECT().
			GetCredentialByMail(ctx, credentialDto.Email, provider1).
			Return(nil, &ent.NotFoundError{})

		repo.credentialsRepo.EXPECT().
			GetCredentialByProvider(ctx, ids.actorID, provider1).
			Return(nil, &ent.NotFoundError{})

		dbErr := v1.ErrorInternal("insert failed")
		repo.credentialsRepo.EXPECT().
			CreateCredential(ctx, *credentialDto).
			Return(nil, dbErr)

		expectedErr := v1.ErrorDatabaseQuery("failed to create credential: %v", dbErr.Error())
		result, err := credentialService.ExternalAuth(ctx, req)
		require.Error(t, err)
		require.Nil(t, result)
		require.Equal(t, expectedErr, err)
	}
}

func TestCredentialsService_RefreshCredential_SuccessCases(t *testing.T) {
	ctx, repo, credentialService := createCredentialsService(t)
	ids := getIDs()

	req := &v1.CredentialRequest{
		CredentialId: 111,
	}

	providerVal := provider1
	cred := &ent.UserCredentials{
		UserID:      ids.actorID,
		Provider:    &providerVal,
		AccessToken: accessToken,
	}
	repo.credentialsRepo.EXPECT().
		GetCredential(ctx, ids.actorID, req.GetCredentialId()).
		Return(cred, nil)

	repo.provider.EXPECT().
		NewProviderGateway(provider1).
		Return(repo.providerGateway, nil)

	credentialDto := &integration.CredentialDto{
		UserID: ids.actorID,
		Email:  "user@example.com",
	}
	repo.providerGateway.EXPECT().
		RefreshToken(cred).
		Return(credentialDto, nil)

	updatedCred := &ent.UserCredentials{
		UserID:      ids.actorID,
		Provider:    &providerVal,
		AccessToken: "newAccessToken",
	}
	repo.credentialsRepo.EXPECT().
		UpdateCredential(ctx, req.GetCredentialId(), *credentialDto).
		Return(updatedCred, nil)

	result, err := credentialService.RefreshCredential(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, service.UserCredentialToV1Credential(updatedCred), result.GetCredential())
}

func TestCredentialsService_RefreshCredential_ErrorCases(t *testing.T) {
	ctx, repo, credentialService := createCredentialsService(t)
	ids := getIDs()

	// Error Case 1: Empty actor id
	{
		req := &v1.CredentialRequest{
			CredentialId: 111,
		}

		// Create a context without an actor id.
		ctxWithoutActor := context.Background()

		expectedErr := v1.ErrorEmptyActorId("empty actor id")
		result, err := credentialService.RefreshCredential(ctxWithoutActor, req)
		require.Error(t, err)
		require.Nil(t, result)
		require.Equal(t, expectedErr, err)
	}

	// Error Case 2: GetCredential returns a non-NotFound error
	{
		req := &v1.CredentialRequest{
			CredentialId: 111,
		}

		dbErr := v1.ErrorInternal("db error")
		repo.credentialsRepo.EXPECT().
			GetCredential(ctx, ids.actorID, req.GetCredentialId()).
			Return(nil, dbErr)

		expectedErr := v1.ErrorDatabaseQuery("database error: %s", dbErr.Error())
		result, err := credentialService.RefreshCredential(ctx, req)
		require.Error(t, err)
		require.Nil(t, result)
		require.Equal(t, expectedErr, err)
	}

	// Error Case 3: GetCredential returns a NotFound error
	{
		req := &v1.CredentialRequest{
			CredentialId: 111,
		}

		repo.credentialsRepo.EXPECT().
			GetCredential(ctx, ids.actorID, req.GetCredentialId()).
			Return(nil, &ent.NotFoundError{})

		expectedErr := v1.ErrorCredentialNotFound("credential not found")
		result, err := credentialService.RefreshCredential(ctx, req)
		require.Error(t, err)
		require.Nil(t, result)
		require.Equal(t, expectedErr, err)
	}

	// Error Case 4: Credential is missing provider
	{
		req := &v1.CredentialRequest{
			CredentialId: 111,
		}

		cred := &ent.UserCredentials{
			UserID:      ids.actorID,
			Provider:    nil, // missing provider
			AccessToken: accessToken,
		}
		repo.credentialsRepo.EXPECT().
			GetCredential(ctx, ids.actorID, req.GetCredentialId()).
			Return(cred, nil)

		expectedErr := v1.ErrorInternal("credential don't have provider")
		result, err := credentialService.RefreshCredential(ctx, req)
		require.Error(t, err)
		require.Nil(t, result)
		require.Equal(t, expectedErr, err)
	}

	// Error Case 5: NewProviderGateway returns an error
	{
		req := &v1.CredentialRequest{
			CredentialId: 111,
		}

		providerVal := provider1
		cred := &ent.UserCredentials{
			UserID:      ids.actorID,
			Provider:    &providerVal,
			AccessToken: accessToken,
		}
		repo.credentialsRepo.EXPECT().
			GetCredential(ctx, ids.actorID, req.GetCredentialId()).
			Return(cred, nil)

		expectedErr := v1.ErrorInternal("gateway error")
		repo.provider.EXPECT().
			NewProviderGateway(provider1).
			Return(nil, expectedErr)

		result, err := credentialService.RefreshCredential(ctx, req)
		require.Error(t, err)
		require.Nil(t, result)
		require.Equal(t, expectedErr, err)
	}

	// Error Case 6: RefreshToken returns an error
	{
		req := &v1.CredentialRequest{
			CredentialId: 111,
		}

		providerVal := provider1
		cred := &ent.UserCredentials{
			UserID:      ids.actorID,
			Provider:    &providerVal,
			AccessToken: accessToken,
		}
		repo.credentialsRepo.EXPECT().
			GetCredential(ctx, ids.actorID, req.GetCredentialId()).
			Return(cred, nil)

		repo.provider.EXPECT().
			NewProviderGateway(provider1).
			Return(repo.providerGateway, nil)

		errFunc := v1.ErrorInternal("refresh token failed")
		repo.providerGateway.EXPECT().
			RefreshToken(cred).
			Return(nil, errFunc)

		expectedErr := v1.ErrorServiceFailed("service error: %v", errFunc)
		result, err := credentialService.RefreshCredential(ctx, req)
		require.Error(t, err)
		require.Nil(t, result)
		require.Equal(t, expectedErr, err)
	}

	// Error Case 7: UpdateCredential returns an error
	{
		req := &v1.CredentialRequest{
			CredentialId: 111,
		}

		providerVal := provider1
		cred := &ent.UserCredentials{
			UserID:      ids.actorID,
			Provider:    &providerVal,
			AccessToken: accessToken,
		}
		repo.credentialsRepo.EXPECT().
			GetCredential(ctx, ids.actorID, req.GetCredentialId()).
			Return(cred, nil)

		repo.provider.EXPECT().
			NewProviderGateway(provider1).
			Return(repo.providerGateway, nil)

		credentialDto := &integration.CredentialDto{
			UserID: ids.actorID,
			Email:  "user@example.com",
		}
		repo.providerGateway.EXPECT().
			RefreshToken(cred).
			Return(credentialDto, nil)

		updateErr := v1.ErrorInternal("update failed")
		repo.credentialsRepo.EXPECT().
			UpdateCredential(ctx, req.GetCredentialId(), *credentialDto).
			Return(nil, updateErr)

		expectedErr := updateErr
		result, err := credentialService.RefreshCredential(ctx, req)
		require.Error(t, err)
		require.Nil(t, result)
		require.Equal(t, expectedErr, err)
	}
}

package test_test

import (
	"gitlab.calendaria.team/services/iam/ent"
	"testing"

	v1 "gitlab.calendaria.team/services/iam/api/iam/v1"
	"gitlab.calendaria.team/services/iam/internal/data/integration"
	u_struc "gitlab.calendaria.team/services/utils/v2/struc"
	u_zap "gitlab.calendaria.team/services/utils/v2/zap"

	"github.com/stretchr/testify/require"
)

const (
	provider1 = u_struc.Google
	authCode  = "def502006c1eebbf4d4033d2aa3930a37d944ab044d97b39041e64e3e5df18d3c779095796dcda372db53aaf79e9faec6a578ad3e1c4530d67c8d30bd605376c4a2e1dc22bbf215ca66147df1ee528a1a7d09910123257aabee6e7434d0ab0e013009600762f589e2b1ddd07c6ef1a2d6637d69a213c366114b20fcb572966f91454e256a364baaf9b2c36c83b3e3e1ffed0a40d3b149573e5080d8e51c8097b516d077d4e180d1daed07fb414d1f997ba04a5ba8059d47e8e7541140c38b59e89044e5c61fcc684a52a5cd5332cc79daee54f0e30ac853fcab8b761df144680dfebb73aeecd0547cbf78276c7e3f6c4668097ee316e1bcfb777c5e4c41f00321ec277d54bb46bb83f5a9e6518b6e50ce561caf7d3328d741e70bf0abfb704d12c86893c27ff6953dcad46819ca83b2956c8e216fdb7ecef0b27d4cbc2b5f9c8e6e66a21"
)

func TestCredentialsService_ExternalAuth(t *testing.T) {
	logger := u_zap.NewZapLogger(true)
	ctx, repo, credentialService := createCredentialsService(t)

	req := &v1.ExternalAuthRequest{
		AuthCode: authCode,
		Provider: provider1.Value(),
	}

	providerManager, err := integration.NewProviderManager(nil, logger)
	require.NoError(t, err)

	provider, err := providerManager.NewProviderGateway(provider1)
	require.NoError(t, err)

	googleProvider, _ := provider.(*integration.GoogleGateway)
	repo.provider.EXPECT().NewProviderGateway(provider1).Return(googleProvider, nil)

	_, err = credentialService.ExternalAuth(ctx, req)
	require.NoError(t, err)
}

func TestCredentialsService_ExternalAuth_ErrorCase(t *testing.T) {
	ctx, _, credentialService := createCredentialsService(t)

	req := &v1.ExternalAuthRequest{
		AuthCode: authCode,
		Provider: "unknown",
	}

	_, err := credentialService.ExternalAuth(ctx, req)
	require.Error(t, err, v1.ErrorInvalidProvider("unknown provider"))
}

func TestCredentialsService_RefreshCredential(t *testing.T) {
	logger := u_zap.NewZapLogger(true)
	ctx, repo, credentialService := createCredentialsService(t)
	ids := getIDs()

	req := &v1.CredentialRequest{
		CredentialId: 1,
	}

	providerGoogle := u_struc.Google
	mail := "test@gmail.com"
	credential := &ent.UserCredentials{
		UserID:   ids.actorID,
		Provider: &providerGoogle,
		Mail:     &mail,
	}
	repo.credentialsRepo.EXPECT().GetCredential(ctx, ids.actorID, req.CredentialId).Return(credential, nil)

	providerManager, err := integration.NewProviderManager(nil, logger)
	require.NoError(t, err)

	provider, err := providerManager.NewProviderGateway(provider1)
	require.NoError(t, err)

	googleProvider, _ := provider.(*integration.GoogleGateway)
	repo.provider.EXPECT().NewProviderGateway(provider1).Return(googleProvider, nil)

	_, err = credentialService.RefreshCredential(ctx, req)
	require.NoError(t, err)
}

func TestCredentialsService_RefreshCredential_ErrorCase(t *testing.T) {
	ctx, repo, credentialService := createCredentialsService(t)
	ids := getIDs()

	req := &v1.CredentialRequest{
		CredentialId: 1,
	}

	repo.credentialsRepo.EXPECT().GetCredential(ctx, ids.actorID, req.CredentialId).Return(nil, &ent.NotFoundError{})

	_, err := credentialService.RefreshCredential(ctx, req)
	require.Error(t, err)
	require.Equal(t, err, v1.ErrorCredentialNotFound("credential not found"))
}

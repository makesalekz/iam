package integration

import (
	"context"

	iam_v1 "gitlab.calendaria.team/services/iam/api/iam/v1"
	"gitlab.calendaria.team/services/iam/ent"
	u_struc "gitlab.calendaria.team/services/utils/v2/struc"

	xoauth2 "golang.org/x/oauth2"
)

type IProviderGateway interface {
	Authenticate(ctx context.Context, actorID int64, authCode string) (*CredentialDto, error)
	RefreshToken(ctx context.Context, credential *ent.UserCredentials) (*CredentialDto, error)
}

func (dm *ProviderManager) NewProviderGateway(
	provider u_struc.Provider,
) (IProviderGateway, error) {
	switch provider {
	case u_struc.Google:
		return &GoogleGateway{}, nil
	case u_struc.Sxodim:
		return &SxodimGateway{}, nil
	}

	return nil, iam_v1.ErrorNotFound("unknown provider")
}

func CredentialToDto(credential *ent.UserCredentials) *CredentialDto {
	// Collect token
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

	// Create dto
	dto := &CredentialDto{
		UserID:   credential.UserID,
		Provider: u_struc.Google,
		Token:    token,
	}
	if credential.DisplayName != nil {
		dto.DisplayName = *credential.DisplayName
	}
	if credential.Mail != nil {
		dto.Email = *credential.Mail
	}

	return dto
}

type CredentialDto struct {
	UserID      int64
	Provider    u_struc.Provider
	DisplayName string
	Email       string
	Token       *xoauth2.Token
}

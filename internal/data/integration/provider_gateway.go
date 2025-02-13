package integration

import (
	"context"

	"gitlab.calendaria.team/services/iam/ent"
	u_struc "gitlab.calendaria.team/services/utils/v2/struc"

	xoauth2 "golang.org/x/oauth2"
)

type CredentialDto struct {
	UserID         int64
	ExternalUserID *int64
	Provider       u_struc.Provider
	DisplayName    string
	Email          string
	Phone          string
	Token          *xoauth2.Token
}

type IProviderGateway interface {
	Authenticate(ctx context.Context, actorID int64, authCode string) (*CredentialDto, error)
	RefreshToken(ctx context.Context, credential *ent.UserCredentials) (*CredentialDto, error)
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

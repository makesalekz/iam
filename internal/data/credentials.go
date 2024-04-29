package data

import (
	"context"
	"gitlab.calendaria.team/services/iam/ent/usercredentials"

	"gitlab.calendaria.team/services/iam/ent"
	"gitlab.calendaria.team/services/iam/ent/property"
	"golang.org/x/oauth2"
)

type CredentialsRepo interface {
	CreateCredential(ctx context.Context, actorId int64, token *oauth2.Token) (*ent.UserCredentials, error)
	GetCredential(ctx context.Context, userId int64, provider property.Provider) (*ent.UserCredentials, error)
	DeleteCredential(ctx context.Context, userId, credentialId int64) (int, error)
}

type credentialsRepo struct {
	db *ent.Client
}

func NewCredentialsRepo(d *Data) CredentialsRepo {
	return &credentialsRepo{
		db: d.db,
	}
}

func (r *credentialsRepo) CreateCredential(ctx context.Context, actorId int64, token *oauth2.Token) (*ent.UserCredentials, error) {
	return r.db.UserCredentials.Create().
		SetUserID(actorId).
		SetProvider(property.Google).
		SetAccessToken(token.AccessToken).
		SetTokenType(token.TokenType).
		SetRefreshToken(token.RefreshToken).
		SetExpiresAt(token.Expiry).
		Save(ctx)
}

func (r *credentialsRepo) GetCredential(ctx context.Context, userId int64, provider property.Provider) (*ent.UserCredentials, error) {
	return r.db.UserCredentials.Query().
		Where(
			usercredentials.UserID(userId),
			usercredentials.ProviderEQ(provider),
		).
		Order(ent.Desc(usercredentials.FieldID)).
		First(ctx)
}

func (r *credentialsRepo) DeleteCredential(ctx context.Context, userId, credentialId int64) (int, error) {
	return r.db.UserCredentials.Delete().
		Where(
			usercredentials.ID(credentialId),
			usercredentials.UserID(userId),
		).
		Exec(ctx)
}

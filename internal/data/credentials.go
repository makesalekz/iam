package data

import (
	"context"

	"gitlab.calendaria.team/services/iam/ent"
	"gitlab.calendaria.team/services/iam/ent/usercredentials"
	u_struc "gitlab.calendaria.team/services/utils/v2/struc"

	"golang.org/x/oauth2"
)

type CredentialsRepo interface {
	CreateCredential(ctx context.Context, actorID int64, token *oauth2.Token) (*ent.UserCredentials, error)
	GetCredential(ctx context.Context, userID int64, provider u_struc.Provider) (*ent.UserCredentials, error)
	ListCredentials(ctx context.Context, userID int64) ([]*ent.UserCredentials, error)
	DeleteCredential(ctx context.Context, userID, credentialID int64) (int, error)
}

type credentialsRepo struct {
	db *ent.Client
}

func NewCredentialsRepo(d *Data) CredentialsRepo {
	return &credentialsRepo{
		db: d.db,
	}
}

func (r *credentialsRepo) CreateCredential(
	ctx context.Context, actorID int64, token *oauth2.Token,
) (*ent.UserCredentials, error) {
	return r.db.UserCredentials.Create().
		SetUserID(actorID).
		SetProvider(u_struc.Google).
		SetAccessToken(token.AccessToken).
		SetTokenType(token.TokenType).
		SetRefreshToken(token.RefreshToken).
		SetExpiresAt(token.Expiry).
		Save(ctx)
}

func (r *credentialsRepo) GetCredential(
	ctx context.Context, userID int64, provider u_struc.Provider,
) (*ent.UserCredentials, error) {
	return r.db.UserCredentials.Query().
		Where(
			usercredentials.UserID(userID),
			usercredentials.ProviderEQ(provider),
		).
		Order(ent.Desc(usercredentials.FieldID)).
		First(ctx)
}

func (r *credentialsRepo) ListCredentials(ctx context.Context, userID int64) ([]*ent.UserCredentials, error) {
	return r.db.UserCredentials.Query().
		Where(
			usercredentials.UserID(userID),
		).
		All(ctx)
}

func (r *credentialsRepo) DeleteCredential(ctx context.Context, userID, credentialID int64) (int, error) {
	return r.db.UserCredentials.Delete().
		Where(
			usercredentials.ID(credentialID),
			usercredentials.UserID(userID),
		).
		Exec(ctx)
}

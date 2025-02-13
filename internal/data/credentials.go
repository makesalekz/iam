package data

import (
	"context"

	"gitlab.calendaria.team/services/iam/ent"
	"gitlab.calendaria.team/services/iam/ent/usercredentials"
	"gitlab.calendaria.team/services/iam/internal/data/integration"
	u_struc "gitlab.calendaria.team/services/utils/v2/struc"
)

type CredentialsRepo interface {
	CreateCredential(ctx context.Context, dto integration.CredentialDto) (*ent.UserCredentials, error)
	UpdateCredential(ctx context.Context, credentialID int64, dto integration.CredentialDto) (*ent.UserCredentials, error)
	GetCredential(ctx context.Context, userID, credentialID int64) (*ent.UserCredentials, error)
	GetCredentialByMail(ctx context.Context, mail string, provider u_struc.Provider) (*ent.UserCredentials, error)
	ListCredentials(ctx context.Context, userID int64, provider *u_struc.Provider) ([]*ent.UserCredentials, error)
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
	ctx context.Context, dto integration.CredentialDto,
) (*ent.UserCredentials, error) {
	return r.db.UserCredentials.Create().
		SetUserID(dto.UserID).
		SetNillableExternalUserID(dto.ExternalUserID).
		SetDisplayName(dto.DisplayName).
		SetMail(dto.Email).
		SetPhone(dto.Phone).
		SetProvider(dto.Provider).
		SetAccessToken(dto.Token.AccessToken).
		SetTokenType(dto.Token.TokenType).
		SetRefreshToken(dto.Token.RefreshToken).
		SetExpiresAt(dto.Token.Expiry).
		Save(ctx)
}

func (r *credentialsRepo) UpdateCredential(
	ctx context.Context, credentialID int64, dto integration.CredentialDto,
) (*ent.UserCredentials, error) {
	return r.db.UserCredentials.
		UpdateOneID(credentialID).
		SetAccessToken(dto.Token.AccessToken).
		SetTokenType(dto.Token.TokenType).
		SetRefreshToken(dto.Token.RefreshToken).
		SetExpiresAt(dto.Token.Expiry).
		Save(ctx)
}

func (r *credentialsRepo) GetCredential(
	ctx context.Context, userID, credentialID int64,
) (*ent.UserCredentials, error) {
	return r.db.UserCredentials.Query().
		Where(
			usercredentials.ID(credentialID),
			usercredentials.UserID(userID),
		).
		First(ctx)
}

func (r *credentialsRepo) GetCredentialByMail(
	ctx context.Context, mail string, provider u_struc.Provider,
) (*ent.UserCredentials, error) {
	return r.db.UserCredentials.Query().
		Where(
			usercredentials.MailEQ(mail),
			usercredentials.ProviderEQ(provider),
		).
		Only(ctx)
}

func (r *credentialsRepo) ListCredentials(
	ctx context.Context,
	userID int64,
	provider *u_struc.Provider,
) ([]*ent.UserCredentials, error) {
	query := r.db.UserCredentials.Query().
		Where(
			usercredentials.UserID(userID),
		)

	if provider != nil {
		query = query.Where(usercredentials.ProviderEQ(*provider))
	}

	return query.All(ctx)
}

func (r *credentialsRepo) DeleteCredential(ctx context.Context, userID, credentialID int64) (int, error) {
	return r.db.UserCredentials.Delete().
		Where(
			usercredentials.ID(credentialID),
			usercredentials.UserID(userID),
		).
		Exec(ctx)
}

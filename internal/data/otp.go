package data

import (
	"context"
	"time"

	"gitlab.calendaria.team/services/iam/ent"
	"gitlab.calendaria.team/services/iam/ent/enum"
	"gitlab.calendaria.team/services/iam/ent/onetimepassword"
)

// OtpRepo.
type OtpRepo interface {
	CreateOneTimePassword(
		ctx context.Context,
		userID int64,
		typee enum.OneTimePasswordType,
		code string,
		duration time.Duration,
	) (*ent.OneTimePassword, error)

	CheckOneTimePassword(ctx context.Context, userID int64, code string) (*ent.OneTimePassword, error)
}

type otpRepo struct {
	db *ent.Client
}

// NewOtpRepo .
func NewOtpRepo(d *Data) OtpRepo {
	return &otpRepo{
		db: d.db,
	}
}

func (r *otpRepo) CreateOneTimePassword(
	ctx context.Context,
	userID int64,
	typee enum.OneTimePasswordType,
	code string,
	duration time.Duration,
) (*ent.OneTimePassword, error) {
	expiresAt := time.Now().Add(duration)

	return r.db.OneTimePassword.Create().SetUserID(userID).SetCode(code).SetType(typee).SetExpiresAt(expiresAt).Save(ctx)
}

func (r *otpRepo) CheckOneTimePassword(ctx context.Context, userID int64, code string) (*ent.OneTimePassword, error) {
	otp, err := r.db.OneTimePassword.Query().Where(
		onetimepassword.UserID(userID),
		onetimepassword.Code(code),
		onetimepassword.IsUsed(false),
		onetimepassword.ExpiresAtGT(time.Now()),
	).First(ctx)

	if err != nil {
		return nil, err
	}

	// creating a transaction and rollback method
	tx, err := r.db.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	otp, err = tx.OneTimePassword.UpdateOne(otp).SetIsUsed(true).Save(ctx)
	if err != nil {
		return nil, err
	}

	_, err = tx.User.UpdateOneID(userID).ClearRemoveAt().Save(ctx)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	return otp, nil
}

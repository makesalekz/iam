package data

import (
	"context"
	"time"

	"entgo.io/ent/dialect/sql"
	"gitlab.calendaria.team/services/iam/ent"
	"gitlab.calendaria.team/services/iam/ent/enum"
	"gitlab.calendaria.team/services/iam/ent/onetimepassword"
)

const (
	FailedAttemptsLimit = 5
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
	tx, err := r.db.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	otp, err := tx.OneTimePassword.Query().Where(
		onetimepassword.UserID(userID),
		onetimepassword.Code(code),
		onetimepassword.IsUsed(false),
		onetimepassword.ExpiresAtGT(time.Now()),
		onetimepassword.FailedAttemptsLTE(FailedAttemptsLimit),
	).First(ctx)
	if err != nil && !ent.IsNotFound(err) {
		return nil, err
	}

	if ent.IsNotFound(err) {
		err = tx.OneTimePassword.Update().Where(
			onetimepassword.UserID(userID),
			onetimepassword.IsUsed(false),
			onetimepassword.ExpiresAtGT(time.Now()),
		).Modify(func(s *sql.UpdateBuilder) {
			s.Add(onetimepassword.FieldFailedAttempts, 1)
		}).Exec(ctx)
		if err != nil {
			return nil, err
		}

		err = tx.OneTimePassword.Update().Where(
			onetimepassword.FailedAttemptsGT(FailedAttemptsLimit),
		).SetIsUsed(true).Exec(ctx)
		if err != nil {
			return nil, err
		}

		err = tx.Commit()
		if err != nil {
			return nil, err
		}

		return nil, &ent.NotFoundError{}
	}

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

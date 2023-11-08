package data

import (
	"context"
	"math/rand"
	"os"
	"time"

	"gitlab.calendaria.team/alageum-cloud/iam/ent"
	"gitlab.calendaria.team/alageum-cloud/iam/ent/onetimepassword"
	"gitlab.calendaria.team/alageum-cloud/iam/ent/property"
)

const digits = "0123456789"

// OtpRepo
type OtpRepo interface {
	CreateOneTimePassword(ctx context.Context, userId int64, t property.OneTimePasswordType, duration time.Duration) (*ent.OneTimePassword, error)
	CheckOneTimePassword(ctx context.Context, userId int64, code string) (*ent.OneTimePassword, error)
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

func generateRandomNumber(n int) string {
	result := make([]byte, n)
	for i := range result {
		result[i] = digits[rand.Int63()%int64(len(digits))]
	}
	return string(result)
}

func (r *otpRepo) CreateOneTimePassword(ctx context.Context, userId int64, t property.OneTimePasswordType, duration time.Duration) (*ent.OneTimePassword, error) {
	code := generateRandomNumber(6)
	expiresAt := time.Now().Add(duration)

	debug := os.Getenv("DEBUG")
	if debug != "" { // use fixed code in debug mode
		code = "777333"
	}

	return r.db.OneTimePassword.Create().SetUserID(userId).SetCode(code).SetType(t).SetExpiresAt(expiresAt).Save(ctx)
}

func (r *otpRepo) CheckOneTimePassword(ctx context.Context, userId int64, code string) (*ent.OneTimePassword, error) {
	otp, err := r.db.OneTimePassword.Query().Where(
		onetimepassword.UserID(userId),
		onetimepassword.Code(code),
		onetimepassword.IsUsed(false),
		onetimepassword.ExpiresAtGT(time.Now()),
	).First(ctx)

	if err != nil {
		return nil, err
	}

	otp, err = otp.Update().SetIsUsed(true).Save(ctx)
	if err != nil {
		return nil, err
	}

	return otp, nil
}

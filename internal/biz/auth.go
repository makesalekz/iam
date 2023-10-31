package biz

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"time"

	auth "iam/api/iam/v1"
	"iam/ent"
	"iam/ent/property"
	"iam/internal/conf"
	"iam/internal/data"
	sender "iam/third_party/api/send/v1"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/registry"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/nyaruka/phonenumbers"
)

const DEFAULT_REGION = "KZ"
const AUTH_OTP_DURATION = 5 * time.Minute

// GreeterUsecase is a Greeter usecase.
type AuthUsecase struct {
	conf      *conf.Bootstrap
	log       *log.Helper
	discovery registry.Discovery
	usersRepo data.UsersRepo
	otpRepo   data.OtpRepo
}

// NewAuthUsecase new a Greeter usecase.
func NewAuthUsecase(logger log.Logger, c *data.Config, usersRepo data.UsersRepo, otpRepo data.OtpRepo) (*AuthUsecase, error) {
	return &AuthUsecase{
		conf:      c.Bootstrap,
		log:       log.NewHelper(logger),
		discovery: c.GetRegistry(),
		usersRepo: usersRepo,
		otpRepo:   otpRepo,
	}, nil
}

func (uc *AuthUsecase) dialNotifications(ctx context.Context) (sender.SenderClient, error) {
	conn, err := grpc.DialInsecure(
		ctx,
		grpc.WithEndpoint(uc.conf.Discovery.Notifications),
		grpc.WithDiscovery(uc.discovery),
		grpc.WithTimeout(10*time.Second),
	)
	if err != nil {
		return nil, err
	}
	return sender.NewSenderClient(conn), nil
}

func (uc *AuthUsecase) AuthUserByPhone(ctx context.Context, phone string) (int64, error) {
	phoneNumber, err := phonenumbers.Parse(phone, DEFAULT_REGION)
	if err != nil {
		return 0, auth.ErrorInvalidPhoneNumber("Parse error: %s", err.Error())
	}
	if !phonenumbers.IsValidNumber(phoneNumber) {
		return 0, auth.ErrorInvalidPhoneNumber("Invalid phone number: %s", phone)
	}

	phone = phonenumbers.Format(phoneNumber, phonenumbers.E164)

	uc.log.Infof("phone, %v", phone)

	user, err := uc.usersRepo.GetUserByPhone(ctx, phone)
	if err != nil {
		if ent.IsNotFound(err) {
			user, err = uc.usersRepo.CreateUserWithPhone(ctx, phone)
		}
		if err != nil {
			return 0, auth.ErrorDatabaseQuery("DB Error (UsersRepo): %s", err.Error())
		}
	}

	otp, err := uc.otpRepo.CreateOneTimePassword(ctx, int64(user.ID), property.Phone, AUTH_OTP_DURATION)
	if err != nil {
		return 0, auth.ErrorDatabaseQuery("DB Error (OtpRepo): %s", err.Error())
	}

	debug := os.Getenv("DEBUG")
	if debug == "" { // don't send sms in debug mode
		senderClient, err := uc.dialNotifications(ctx)
		if err != nil {
			return 0, auth.ErrorGrpcConnection("dialNotifications: %s", err.Error())
		}

		reply, err := senderClient.PersonalSmsSender(ctx, &sender.PersonalSmsSenderRequest{
			Phone:   phone,
			Message: fmt.Sprintf("Enter this code to sign in: %s", otp.Code),
		})
		if err != nil {
			return 0, auth.ErrorServiceFailed("sender.PersonalSmsSender: %s", err.Error())
		}
		uc.log.Infof("sender.PersonalSmsSender: %s", reply.Result)
	}

	return int64(user.ID), nil
}

func (uc *AuthUsecase) AuthUserByCode(ctx context.Context, userId int64, code string) error {
	user, err := uc.usersRepo.GetUserById(ctx, userId)
	if err != nil {
		return auth.ErrorDatabaseQuery("DB Error (UsersRepo): %s", err.Error())
	}

	ok, err := uc.otpRepo.CheckOneTimePassword(ctx, user.ID, code)
	if err != nil {
		return auth.ErrorDatabaseQuery("DB Error (OtpRepo): %s", err.Error())
	}
	if !ok {
		return auth.ErrorInvalidCode("Invalid code")
	}
	return nil
}

package biz

import (
	"context"
	"fmt"
	"time"

	"iam/ent"
	"iam/ent/property"
	"iam/internal/data"
	v1 "iam/third_party/api/send/v1"

	consul "github.com/go-kratos/consul/registry"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/registry"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/hashicorp/consul/api"
)

const AUTH_OTP_DURATION = 5 * time.Minute
const DISCOVERY_NOTIFICATIONS = "discovery:///notifications"

// GreeterUsecase is a Greeter usecase.
type UsersUsecase struct {
	conf      config.Config
	log       *log.Helper
	discovery registry.Discovery
	usersRepo data.UsersRepo
	otpRepo   data.OtpRepo
}

// NewGreeterUsecase new a Greeter usecase.
func NewUsersUsecase(c config.Config, logger log.Logger, consulClient *api.Client, usersRepo data.UsersRepo, otpRepo data.OtpRepo) (*UsersUsecase, error) {
	dis := consul.New(consulClient)

	return &UsersUsecase{
		conf:      c,
		log:       log.NewHelper(logger),
		discovery: dis,
		usersRepo: usersRepo,
		otpRepo:   otpRepo,
	}, nil
}

func (uc *UsersUsecase) dialNotifications(ctx context.Context) (v1.SenderClient, error) {
	conn, err := grpc.DialInsecure(ctx, grpc.WithEndpoint(DISCOVERY_NOTIFICATIONS), grpc.WithDiscovery(uc.discovery))
	if err != nil {
		return nil, err
	}
	return v1.NewSenderClient(conn), nil
}

func (uc *UsersUsecase) AuthUserByPhone(ctx context.Context, phone string) (string, error) {
	user, err := uc.usersRepo.GetUserByPhone(ctx, phone)
	if err != nil {
		if ent.IsNotFound(err) {
			user, err = uc.usersRepo.CreateUserWithPhone(ctx, phone)
		}
		if err != nil {
			return "", err
		}
	}

	otp, err := uc.otpRepo.CreateOneTimePassword(ctx, int64(user.ID), property.Phone, AUTH_OTP_DURATION)
	if err != nil {
		return "", err
	}

	senderClient, err := uc.dialNotifications(ctx)
	if err != nil {
		return "", err
	}

	reply, err := senderClient.PersonalSmsSender(ctx, &v1.PersonalSmsSenderRequest{
		Phone:   phone,
		Message: fmt.Sprintf("Enter this code to sign in: %s", otp.Code),
	})
	if err != nil {
		return "", err
	}
	uc.log.Infof("senderClient.PersonalSmsSender: %v", reply)

	return otp.Code, nil
}

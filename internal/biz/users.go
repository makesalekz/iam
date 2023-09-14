package biz

import (
	"context"
	_ "embed"

	"iam/ent"
	"iam/internal/conf"
	"iam/internal/data"

	consul "github.com/go-kratos/consul/registry"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/registry"
	"github.com/hashicorp/consul/api"
)

// UsersUsecase .
type UsersUsecase struct {
	conf      *conf.Bootstrap
	log       *log.Helper
	discovery registry.Discovery
	usersRepo data.UsersRepo
	otpRepo   data.OtpRepo
}

// NewUsersUsecase .
func NewUsersUsecase(c *conf.Bootstrap, logger log.Logger, consulClient *api.Client, usersRepo data.UsersRepo, otpRepo data.OtpRepo) (*UsersUsecase, error) {
	dis := consul.New(consulClient)

	return &UsersUsecase{
		conf:      c,
		log:       log.NewHelper(logger),
		discovery: dis,
		usersRepo: usersRepo,
		otpRepo:   otpRepo,
	}, nil
}

func (uc *UsersUsecase) GetUserProfile(ctx context.Context, userId int64) (*ent.User, error) {
	return uc.usersRepo.GetUserById(ctx, userId)
}

func (uc *UsersUsecase) UpdateUserProfile(ctx context.Context, userId int64, data data.UpdateUserDto) (*ent.User, error) {
	return uc.usersRepo.UpdateUserData(ctx, userId, data)
}

func (uc *UsersUsecase) DeleteUser(ctx context.Context, userId int64) error {
	return uc.usersRepo.DeleteUser(ctx, userId)
}

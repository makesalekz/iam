package biz

import (
	"context"
	_ "embed"

	"iam/ent"
	"iam/internal/data"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/registry"
)

// UsersUsecase .
type UsersUsecase struct {
	log       *log.Helper
	discovery registry.Discovery
	usersRepo data.UsersRepo
	otpRepo   data.OtpRepo
}

// NewUsersUsecase .
func NewUsersUsecase(logger log.Logger, c *data.Config, usersRepo data.UsersRepo, otpRepo data.OtpRepo) (*UsersUsecase, error) {
	return &UsersUsecase{
		log:       log.NewHelper(logger),
		discovery: c.GetRegistry(),
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

func (uc *UsersUsecase) GetUsers(ctx context.Context, usersIds []int64) ([]*ent.User, error) {
	return uc.usersRepo.GetUsersByIds(ctx, usersIds)
}

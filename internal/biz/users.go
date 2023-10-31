package biz

import (
	"context"
	_ "embed"

	iam_v1 "iam/api/iam/v1"
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

func (uc *UsersUsecase) GetUserProfile(ctx context.Context, filter data.GetUserFilterDto) (*ent.User, error) {
	if filter.Phone != "" && filter.Email == "" {
		return uc.usersRepo.GetUserByPhone(ctx, filter.Phone)
	} else if filter.Email != "" && filter.Phone == "" {
		return uc.usersRepo.GetUserByEmail(ctx, filter.Email)
	} else if filter.UserId != 0 && filter.Email == "" && filter.Phone == "" {
		return uc.usersRepo.GetUserById(ctx, filter.UserId)
	}

	return nil, iam_v1.ErrorInvalidRequest("invalid request, please read documentations")
}

func (uc *UsersUsecase) UpdateUserProfile(ctx context.Context, userId int64, data data.UpdateUserDto) (*ent.User, error) {
	return uc.usersRepo.UpdateUserData(ctx, userId, data)
}

func (uc *UsersUsecase) DeleteUser(ctx context.Context, userId int64) error {
	return uc.usersRepo.DeleteUser(ctx, userId)
}

func (uc *UsersUsecase) GetUsers(ctx context.Context, filter data.GetUsersFilterDto) ([]*ent.User, error) {
	return uc.usersRepo.GetUsers(ctx, filter)
}

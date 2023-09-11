package service

import (
	"context"
	"strconv"

	v1 "iam/api/users/v1"
	"iam/internal/biz"
	"iam/internal/data"

	"github.com/go-kratos/kratos/v2/log"
)

type UsersService struct {
	v1.UnimplementedUsersServer

	log *log.Helper
	jwt *biz.JwtProcessor
	uc  *biz.UsersUsecase
}

func NewUsersService(logger log.Logger, jwt *biz.JwtProcessor, uc *biz.UsersUsecase) *UsersService {
	return &UsersService{
		log: log.NewHelper(logger),
		jwt: jwt,
		uc:  uc,
	}
}

func (s *UsersService) GetOwnProfile(ctx context.Context, req *v1.GetOwnProfileRequest) (*v1.GetOwnProfileReply, error) {
	userId, ok := s.jwt.GetUserIdFromContext(ctx)
	if !ok {
		return nil, v1.ErrorUnauthorized("Unauthorized")
	}

	user, err := s.uc.GetUserProfile(ctx, userId)
	if err != nil {
		return nil, v1.ErrorUserNotFound("User not found: %v", err)
	}

	return &v1.GetOwnProfileReply{
		Id:        strconv.FormatInt(user.ID, 10),
		Name:      user.Name,
		Phone:     *user.Phone,
		Email:     *user.Email,
		Bio:       user.Bio,
		Avatar:    *user.Avatar,
		Timezone:  user.Timezone,
		CreatedAt: user.CreatedAt.Local().String(),
		UpdatedAt: user.UpdatedAt.Local().String(),
	}, nil
}

func (s *UsersService) UpdateOwnProfile(ctx context.Context, req *v1.UpdateOwnProfileRequest) (*v1.GetOwnProfileReply, error) {
	userId, ok := s.jwt.GetUserIdFromContext(ctx)
	if !ok {
		return nil, v1.ErrorUnauthorized("Unauthorized")
	}

	user, err := s.uc.UpdateUserProfile(ctx, userId, data.UpdateUserDto{
		Name:     req.Name,
		Bio:      req.Bio,
		Avatar:   req.Avatar,
		Timezone: req.Timezone,
	})
	if err != nil {
		return nil, v1.ErrorUserNotFound("User not found: %v", err)
	}

	return &v1.GetOwnProfileReply{
		Id:        strconv.FormatInt(user.ID, 10),
		Name:      user.Name,
		Phone:     *user.Phone,
		Email:     *user.Email,
		Bio:       user.Bio,
		Avatar:    *user.Avatar,
		Timezone:  user.Timezone,
		CreatedAt: user.CreatedAt.Local().String(),
		UpdatedAt: user.UpdatedAt.Local().String(),
	}, nil
}

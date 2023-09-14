package service

import (
	"context"
	"strconv"

	v1 "iam/api/users/v1"
	"iam/ent"
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

func transformUserForReply(user *ent.User) *v1.User {
	result := &v1.User{
		Id:        strconv.FormatInt(user.ID, 10),
		Name:      user.Name,
		Bio:       user.Bio,
		Timezone:  user.Timezone,
		CreatedAt: user.CreatedAt.Local().String(),
		UpdatedAt: user.UpdatedAt.Local().String(),
	}
	if user.Phone != nil {
		result.Phone = *user.Phone
	}
	if user.Email != nil {
		result.Email = *user.Email
	}
	if user.Avatar != nil {
		result.Avatar = *user.Avatar
	}
	return result
}

func (s *UsersService) GetOwnProfile(ctx context.Context, req *v1.EmptyRequest) (*v1.ProfileReply, error) {
	userId, ok := s.jwt.GetUserIdFromContext(ctx)
	s.log.Infof("userId: %v", userId)
	if !ok {
		return nil, v1.ErrorUnauthorized("Unauthorized")
	}

	user, err := s.uc.GetUserProfile(ctx, userId)
	if err != nil {
		_, notFound := err.(*ent.NotFoundError)
		if notFound {
			return nil, v1.ErrorUserNotFound("User not found: %v", err)
		}
		return nil, v1.ErrorDatabaseQuery("Internal error")
	}

	return &v1.ProfileReply{User: transformUserForReply(user)}, nil
}

func (s *UsersService) UpdateOwnProfile(ctx context.Context, req *v1.UpdateOwnProfileRequest) (*v1.ProfileReply, error) {
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
		_, notFound := err.(*ent.NotFoundError)
		if notFound {
			return nil, v1.ErrorUserNotFound("User not found: %v", err)
		}
		return nil, v1.ErrorDatabaseQuery("Internal error")
	}

	return &v1.ProfileReply{User: transformUserForReply(user)}, nil
}

func (s *UsersService) DeleteOwnProfile(ctx context.Context, req *v1.EmptyRequest) (*v1.EmptyReply, error) {
	userId, ok := s.jwt.GetUserIdFromContext(ctx)
	if !ok {
		return nil, v1.ErrorUnauthorized("Unauthorized")
	}

	err := s.uc.DeleteUser(ctx, userId)
	if err != nil {
		_, notFound := err.(*ent.NotFoundError)
		if notFound {
			return nil, v1.ErrorUserNotFound("User not found: %v", err)
		}
		return nil, v1.ErrorDatabaseQuery("Internal error")
	}

	return &v1.EmptyReply{}, nil
}

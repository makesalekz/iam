package service

import (
	"context"
	"time"

	v1 "iam/api/users/v1"
	"iam/ent"
	"iam/internal/biz"
	"iam/internal/data"

	"github.com/go-kratos/kratos/v2/log"
)

type UsersService struct {
	v1.UnimplementedUsersServer

	log *log.Helper
	jwt *data.JwtProcessor
	uc  *biz.UsersUsecase
}

func NewUsersService(logger log.Logger, jwt *data.JwtProcessor, uc *biz.UsersUsecase) *UsersService {
	return &UsersService{
		log: log.NewHelper(logger),
		jwt: jwt,
		uc:  uc,
	}
}

func replyUser(user *ent.User) *v1.User {
	result := &v1.User{
		Id:           user.ID,
		Name:         user.Name,
		Bio:          user.Bio,
		Timezone:     user.Timezone,
		CreatedAt:    user.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    user.UpdatedAt.Format(time.RFC3339),
		LastLoginAt:  user.LastLoginAt.Format(time.RFC3339),
		IsActive:     user.IsActive,
		BioUpdatedAt: user.BioUpdatedAt.Format(time.RFC3339),
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

func replyUserShort(user *ent.User) *v1.UserShort {
	result := &v1.UserShort{
		Id:          user.ID,
		Name:        user.Name,
		LastLoginAt: user.LastLoginAt.Format(time.RFC3339),
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

func replyUsers(users []*ent.User) []*v1.UserShort {
	var replies []*v1.UserShort
	for _, user := range users {
		replies = append(replies, replyUserShort(user))
	}
	return replies
}

func (s *UsersService) GetOwnProfile(ctx context.Context, req *v1.EmptyRequest) (*v1.UserFullReply, error) {
	userId, ok := s.jwt.GetUserIdFromContext(ctx)
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

	return &v1.UserFullReply{User: replyUser(user)}, nil
}

func (s *UsersService) UpdateOwnProfile(ctx context.Context, req *v1.UpdateOwnProfileRequest) (*v1.UserFullReply, error) {
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

	return &v1.UserFullReply{User: replyUser(user)}, nil
}

func (s *UsersService) DeleteOwnProfile(ctx context.Context, req *v1.EmptyRequest) (*v1.EmptyReply, error) {
	userId, ok := s.jwt.GetUserIdFromContext(ctx)
	if !ok {
		return nil, v1.ErrorUnauthorized("Unauthorized")
	}

	// TODO мягко удалить или "пофиксить" все связанные сущности
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

func (s *UsersService) GetUserFull(ctx context.Context, req *v1.GetUserRequest) (*v1.UserFullReply, error) {
	user, err := s.uc.GetUserProfile(ctx, req.UserId)
	if err != nil {
		_, notFound := err.(*ent.NotFoundError)
		if notFound {
			return nil, v1.ErrorUserNotFound("User not found: %v", err)
		}
		return nil, v1.ErrorDatabaseQuery("Internal error")
	}

	return &v1.UserFullReply{User: replyUser(user)}, nil
}

func (s *UsersService) GetUser(ctx context.Context, req *v1.GetUserRequest) (*v1.UserReply, error) {
	user, err := s.uc.GetUserProfile(ctx, req.UserId)
	if err != nil {
		_, notFound := err.(*ent.NotFoundError)
		if notFound {
			return nil, v1.ErrorUserNotFound("User not found: %v", err)
		}
		return nil, v1.ErrorDatabaseQuery("Internal error")
	}

	return &v1.UserReply{User: replyUserShort(user)}, nil
}

func (s *UsersService) GetUsers(ctx context.Context, req *v1.GetUsersRequest) (*v1.GetUsersReply, error) {
	s.log.Infof("GetUsers: %v", req.Ids)
	users, err := s.uc.GetUsers(ctx, req.Ids)
	if err != nil {
		return nil, v1.ErrorDatabaseQuery("Internal error")
	}

	return &v1.GetUsersReply{Users: replyUsers(users)}, nil
}

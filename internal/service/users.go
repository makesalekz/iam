package service

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	v1 "gitlab.calendaria.team/services/iam/api/iam/v1"
	"gitlab.calendaria.team/services/iam/internal/biz"
	"gitlab.calendaria.team/services/iam/internal/data"
	utils_v1 "gitlab.calendaria.team/services/utils/api/utils/v1"
)

type UsersService struct {
	v1.UnimplementedUsersServer

	log *log.Helper
	sh  *ServiceHelper
	uc  *biz.UsersUsecase
}

func NewUsersService(
	logger log.Logger,
	sh *ServiceHelper,
	uc *biz.UsersUsecase,
) *UsersService {
	return &UsersService{
		log: log.NewHelper(logger),
		sh:  sh,
		uc:  uc,
	}
}

func (s *UsersService) GetOwnProfile(ctx context.Context, req *utils_v1.ActorRequest) (*v1.UserFullReply, error) {
	actorId, err := s.sh.GetActorId(ctx, req.ActorId)
	if err != nil {
		return nil, err
	}

	user, err := s.uc.GetUserProfile(ctx, data.GetUserFilterDto{UserId: actorId})
	if err != nil {
		return nil, err
	}

	result := v1.UserFullReply{User: userItemToV1User(user)}

	tenants, err := s.uc.GetUserTenants(ctx, actorId)
	if err == nil {
		resultTenants := make([]*v1.TenantShort, len(tenants))
		for i, tenant := range tenants {
			resultTenants[i] = &v1.TenantShort{
				Id:   tenant.Id,
				Name: tenant.Name,
			}
		}
		result.Tenants = resultTenants
	} else {
		s.log.Warnf("tenants: ", err)
	}

	return &result, nil
}

func (s *UsersService) UpdateOwnProfile(ctx context.Context, req *v1.UpdateOwnProfileRequest) (*v1.UserFullReply, error) {
	actorId, err := s.sh.GetActorId(ctx, req.ActorId)
	if err != nil {
		return nil, err
	}

	user, err := s.uc.UpdateUserProfile(ctx, actorId, data.UpdateUserDto{
		Phone:    req.Phone,
		Email:    req.Email,
		Name:     req.Name,
		Bio:      req.Bio,
		Avatar:   req.Avatar,
		Timezone: req.Timezone,
	})
	if err != nil {
		return nil, err
	}

	result := v1.UserFullReply{User: userItemToV1User(user)}

	if req.WithTenants {
		tenants, err := s.uc.GetUserTenants(ctx, actorId)
		if err == nil {
			resultTenants := make([]*v1.TenantShort, len(tenants))
			for i, tenant := range tenants {
				resultTenants[i] = &v1.TenantShort{
					Id:   tenant.Id,
					Name: tenant.Name,
				}
			}
			result.Tenants = resultTenants
		} else {
			s.log.Warnf("tenants: ", err)
		}
	}

	return &result, nil
}

func (s *UsersService) DeleteOwnProfile(ctx context.Context, req *utils_v1.ActorRequest) (*utils_v1.EmptyReply, error) {
	actorId, err := s.sh.GetActorId(ctx, req.ActorId)
	if err != nil {
		return nil, err
	}

	// TODO мягко удалить или "пофиксить" все связанные сущности
	err = s.uc.DeleteUser(ctx, actorId)
	if err != nil {
		return nil, err
	}

	return &utils_v1.EmptyReply{}, nil
}

func (s *UsersService) GetUserFull(ctx context.Context, req *v1.GetUserRequest) (*v1.UserFullReply, error) {
	filter := data.GetUserFilterDto{
		UserId: req.GetUserId(),
	}

	user, err := s.uc.GetUserProfile(ctx, filter)
	if err != nil {
		return nil, err
	}

	return &v1.UserFullReply{User: userItemToV1User(user)}, nil
}

func (s *UsersService) GetUser(ctx context.Context, req *v1.GetUserRequest) (*v1.UserReply, error) {
	filter := data.GetUserFilterDto{
		UserId: req.GetUserId(),
	}

	user, err := s.uc.GetUserProfile(ctx, filter)
	if err != nil {
		return nil, err
	}

	return &v1.UserReply{User: userItemToV1ShortUser(user)}, nil
}

func (s *UsersService) GetUsers(ctx context.Context, req *v1.GetUsersRequest) (*v1.GetUsersReply, error) {
	// TODO. Remove this thing later
	actorId, err := s.sh.GetActorId(ctx, req.ActorId)
	if err != nil {
		return nil, err
	}

	filter := data.GetUsersFilterDto{
		UsersIds:      req.GetIds(),
		Phones:        req.GetPhones(),
		Emails:        req.GetEmails(),
		Search:        req.GetSearch(),
		WithRelation:  req.WithRelation,
		WithPrivacies: req.WithPrivacies,
		WithVerified:  req.WithVerified,
	}

	users, err := s.uc.GetUsers(ctx, actorId, filter, req.Sort, req.Paginate)
	if err != nil {
		return nil, err
	}

	return &v1.GetUsersReply{Users: userItemsToV1ShortUser(users)}, nil
}

func (s *UsersService) GetUserByFilter(ctx context.Context, req *v1.GetUserByFilterRequest) (*v1.UserReply, error) {
	filter := data.GetUserFilterDto{
		Phone: req.GetSearch().GetPhone(),
		Email: req.GetSearch().GetEmail(),
	}

	user, err := s.uc.GetUserProfile(ctx, filter)
	if err != nil {
		return nil, err
	}

	return &v1.UserReply{User: userItemToV1ShortUser(user)}, nil
}

func (s *UsersService) GetUserByFilterFull(ctx context.Context, req *v1.GetUserByFilterRequest) (*v1.UserFullReply, error) {
	filter := data.GetUserFilterDto{
		Phone: req.GetSearch().GetPhone(),
		Email: req.GetSearch().GetEmail(),
	}

	user, err := s.uc.GetUserProfile(ctx, filter)
	if err != nil {
		return nil, err
	}

	return &v1.UserFullReply{User: userItemToV1User(user)}, nil
}

func userItemToV1User(user *biz.UserItem) *v1.User {
	if user == nil {
		return &v1.User{}
	}

	replyUser := &v1.User{
		Id:          user.ID,
		Phone:       user.Phone,
		Email:       user.Email,
		Name:        user.Name,
		Bio:         user.Bio,
		Avatar:      user.Avatar,
		Timezone:    user.Timezone,
		CreatedAt:   user.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   user.UpdatedAt.Format(time.RFC3339),
		LastLoginAt: user.LastLoginAt.Format(time.RFC3339),
		IsActive:    user.IsActive,
	}

	if user.WithVerified {
		replyUser.IsPhoneVerified = &user.PhoneVerified
		replyUser.IsEmailVerified = &user.EmailVerified
	}

	if user.BioUpdatedAt != nil {
		bioUpdatedAt := user.BioUpdatedAt.Format(time.RFC3339)
		replyUser.BioUpdatedAt = &bioUpdatedAt
	}

	return replyUser
}

func userItemToV1ShortUser(user *biz.UserItem) *v1.UserShort {
	replyUser := &v1.UserShort{
		Id:          user.ID,
		Name:        user.Name,
		LastLoginAt: user.LastLoginAt.Format(time.RFC3339),
		Relation:    user.Relation,
		Privacies:   user.Privacies,
	}

	if user.WithVerified {
		replyUser.IsPhoneVerified = &user.PhoneVerified
		replyUser.IsEmailVerified = &user.EmailVerified
	}

	if user.Phone != nil {
		replyUser.Phone = *user.Phone
	}
	if user.Email != nil {
		replyUser.Email = *user.Email
	}
	if user.Avatar != nil {
		replyUser.Avatar = *user.Avatar
	}

	return replyUser
}

func userItemsToV1ShortUser(users []*biz.UserItem) []*v1.UserShort {
	replyUsers := make([]*v1.UserShort, len(users))
	for i, user := range users {
		replyUsers[i] = userItemToV1ShortUser(user)
	}

	return replyUsers
}

package service

import (
	"context"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	v1 "gitlab.calendaria.team/services/iam/api/iam/v1"
	"gitlab.calendaria.team/services/iam/internal/biz"
	"gitlab.calendaria.team/services/iam/internal/data"
	utils_v1 "gitlab.calendaria.team/services/utils/api/utils/v1"
	"gitlab.calendaria.team/services/utils/v2/auth"
)

type UsersService struct {
	v1.UnimplementedUsersServer

	log *log.Helper
	uc  *biz.UsersUsecase
}

func NewUsersService(
	logger log.Logger,
	uc *biz.UsersUsecase,
) *UsersService {
	return &UsersService{
		log: log.NewHelper(log.With(logger, "module", "service/users")),
		uc:  uc,
	}
}

func validUsername(actorID int64, username string) bool {
	if username == "" {
		return true
	}

	// username can't be default format (user{number}), only access to own user id
	if len(username) > 4 && strings.ToLower(username[:4]) == "user" {
		userID, err := strconv.ParseInt(username[4:], 10, 64)
		if err == nil && userID != actorID {
			return false
		}
	}

	// username must contain only lowercase letters and numbers
	re := regexp.MustCompile(`^[a-z][a-z0-9]+$`)
	return re.MatchString(username)
}

func (s *UsersService) GetOwnProfile(ctx context.Context, _ *utils_v1.EmptyRequest) (*v1.UserFullReply, error) {
	actorID := auth.GetActorIdFromContext(ctx)
	if actorID == 0 {
		return nil, v1.ErrorEmptyActorId("empty actor id")
	}

	user, err := s.uc.GetUserProfile(ctx, data.GetUserFilterDto{UserID: actorID})
	if err != nil {
		return nil, err
	}

	result := v1.UserFullReply{User: userItemToV1User(user)}

	tenants, err := s.uc.GetUserTenants(ctx)
	if err == nil {
		resultTenants := make([]*v1.TenantShort, len(tenants))
		for i, tenant := range tenants {
			resultTenants[i] = &v1.TenantShort{
				Id:   tenant.GetId(),
				Name: tenant.GetName(),
			}
		}
		result.Tenants = resultTenants
	} else {
		s.log.Error(err.Error())
	}

	return &result, nil
}

func (s *UsersService) UpdateOwnProfile(
	ctx context.Context,
	req *v1.UpdateOwnProfileRequest,
) (*v1.UserFullReply, error) {
	actorID := auth.GetActorIdFromContext(ctx)
	if actorID == 0 {
		return nil, v1.ErrorEmptyActorId("empty actor id")
	}

	if !validUsername(actorID, req.GetUsername()) {
		return nil, v1.ErrorInvalidUsername("forbidden username format")
	}

	user, err := s.uc.UpdateUserProfile(ctx, actorID, data.UpdateUserDto{
		Phone:    req.GetPhone(),
		Email:    req.GetEmail(),
		Name:     req.GetName(),
		Username: req.GetUsername(),
		Bio:      req.GetBio(),
		Avatar:   req.GetAvatar(),
		Timezone: req.GetTimezone(),
	})
	if err != nil {
		return nil, err
	}

	result := v1.UserFullReply{User: userItemToV1User(user)}

	if req.GetWithTenants() {
		tenants, err2 := s.uc.GetUserTenants(ctx)
		if err2 == nil {
			resultTenants := make([]*v1.TenantShort, len(tenants))
			for i, tenant := range tenants {
				resultTenants[i] = &v1.TenantShort{
					Id:   tenant.GetId(),
					Name: tenant.GetName(),
				}
			}
			result.Tenants = resultTenants
		} else {
			s.log.Error(err2.Error())
		}
	}

	return &result, nil
}

func (s *UsersService) DeleteOwnProfile(ctx context.Context, _ *utils_v1.EmptyRequest) (*utils_v1.EmptyReply, error) {
	actorID := auth.GetActorIdFromContext(ctx)
	if actorID == 0 {
		return nil, v1.ErrorEmptyActorId("empty actor id")
	}

	err := s.uc.ScheduleUserDeletion(ctx, actorID)
	if err != nil {
		return nil, err
	}

	return &utils_v1.EmptyReply{}, nil
}

func (s *UsersService) GetUserFull(ctx context.Context, req *v1.GetUserRequest) (*v1.UserFullReply, error) {
	filter := data.GetUserFilterDto{
		UserID: req.GetUserId(),
	}

	user, err := s.uc.GetUserProfile(ctx, filter)
	if err != nil {
		return nil, err
	}

	return &v1.UserFullReply{User: userItemToV1User(user)}, nil
}

func (s *UsersService) GetUser(ctx context.Context, req *v1.GetUserRequest) (*v1.UserReply, error) {
	filter := data.GetUserFilterDto{
		UserID: req.GetUserId(),
	}

	user, err := s.uc.GetUserProfile(ctx, filter)
	if err != nil {
		return nil, err
	}

	return &v1.UserReply{User: userItemToV1ShortUser(user)}, nil
}

func (s *UsersService) ListUsers(ctx context.Context, req *v1.ListUsersRequest) (*v1.UsersReply, error) {
	filter := data.GetUsersFilterDto{
		UsersIDs: req.GetIds(),
		Search:   req.GetSearch(),
	}

	users, err := s.uc.ListUsers(ctx, filter, req.GetSort(), req.GetPaginate())
	if err != nil {
		return nil, err
	}

	return &v1.UsersReply{Users: userItemsToV1ShortUser(users)}, nil
}

func (s *UsersService) GetUsers(ctx context.Context, req *v1.GetUsersRequest) (*v1.UsersReply, error) {
	filter := data.GetUsersFilterDto{
		UsersIDs:      req.GetIds(),
		Phones:        req.GetPhones(),
		Emails:        req.GetEmails(),
		WithPrivacies: req.GetWithPrivacies(),
		WithVerified:  req.GetWithVerified(),
	}

	users, err := s.uc.GetUsers(ctx, filter)
	if err != nil {
		return nil, err
	}

	return &v1.UsersReply{Users: userItemsToV1ShortUser(users)}, nil
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

func (s *UsersService) GetUserByFilterFull(
	ctx context.Context,
	req *v1.GetUserByFilterRequest,
) (*v1.UserFullReply, error) {
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
		Username:    user.Username,
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
	if user.Username != nil {
		replyUser.Username = *user.Username
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

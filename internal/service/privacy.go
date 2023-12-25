package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	v1 "gitlab.calendaria.team/services/iam/api/iam/v1"
	"gitlab.calendaria.team/services/iam/ent"
	"gitlab.calendaria.team/services/iam/internal/biz"
	utils_v1 "gitlab.calendaria.team/services/utils/api/utils/v1"
	"gitlab.calendaria.team/services/utils/v1/jwt"
)

type PrivacyService struct {
	v1.UnimplementedPrivacyServer

	log *log.Helper
	jwt *jwt.JwtProcessor
	uc  *biz.PrivacyUsecase
}

func NewPrivacyService(logger log.Logger, jwt *jwt.JwtProcessor, uc *biz.PrivacyUsecase) *PrivacyService {
	return &PrivacyService{
		log: log.NewHelper(logger),
		jwt: jwt,
		uc:  uc,
	}
}

func (s *PrivacyService) GetPrivacy(ctx context.Context, req *utils_v1.EmptyRequest) (*v1.PrivacyReply, error) {
	userId := s.jwt.GetUserIdFromContext(ctx)
	if userId == 0 {
		return nil, v1.ErrorUnauthorized("invalid token")
	}

	settings, err := s.uc.GetPrivacy(ctx, userId)
	if err != nil {
		return nil, v1.ErrorDatabaseQuery("database error: %s", err.Error())
	}

	return &v1.PrivacyReply{
		Settings: settings,
	}, nil
}

func (s *PrivacyService) UpdatePrivacy(ctx context.Context, req *v1.PrivacyRequest) (*v1.PrivacyReply, error) {
	userId := s.jwt.GetUserIdFromContext(ctx)
	if userId == 0 {
		return nil, v1.ErrorUnauthorized("invalid token")
	}

	settings, err := s.uc.UpdatePrivacy(ctx, userId, req.Settings)
	if err != nil {
		if ent.IsValidationError(err) {
			return nil, v1.ErrorInvalidRequest("invalid request: %s", err.Error())
		}
		return nil, v1.ErrorDatabaseQuery("database error: %s", err.Error())
	}

	return &v1.PrivacyReply{
		Settings: settings,
	}, nil
}

func (s *PrivacyService) GetUsersPrivacies(ctx context.Context, req *v1.UsersPrivaciesRequest) (*v1.UsersPrivaciesReply, error) {
	settings, err := s.uc.GetPrivacies(ctx, req.Ids)
	if err != nil {
		return nil, v1.ErrorDatabaseQuery("database error: %s", err.Error())
	}

	return &v1.UsersPrivaciesReply{
		Users: replyUsersPrivacies(settings),
	}, nil
}

func replyUsersPrivacies(settings []*biz.UserPrivaciesItem) []*v1.UserPrivacies {
	replyPriacies := make([]*v1.UserPrivacies, len(settings))
	for i, setting := range settings {
		replyPriacies[i] = &v1.UserPrivacies{
			Id:        setting.UserId,
			Privacies: setting.Privacies,
		}
	}

	return replyPriacies
}

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

type SettingsService struct {
	v1.UnimplementedSettingsServer

	log *log.Helper
	jwt *jwt.JwtProcessor
	uc  *biz.SettingsUsecase
}

func NewSettingsService(logger log.Logger, jwt *jwt.JwtProcessor, uc *biz.SettingsUsecase) *SettingsService {
	return &SettingsService{
		log: log.NewHelper(logger),
		jwt: jwt,
		uc:  uc,
	}
}

func (s *SettingsService) GetSettings(ctx context.Context, req *utils_v1.EmptyRequest) (*v1.SettingsReply, error) {
	userId := s.jwt.GetUserIdFromContext(ctx)
	if userId == 0 {
		return nil, v1.ErrorUnauthorized("invalid token")
	}

	settings, err := s.uc.GetSettings(ctx, userId)
	if err != nil {
		return nil, v1.ErrorDatabaseQuery("database error: %s", err.Error())
	}

	return &v1.SettingsReply{
		Settings: settings,
	}, nil
}

func (s *SettingsService) UpdateSettings(ctx context.Context, req *v1.SettingsRequest) (*v1.SettingsReply, error) {
	userId := s.jwt.GetUserIdFromContext(ctx)
	if userId == 0 {
		return nil, v1.ErrorUnauthorized("invalid token")
	}

	settings, err := s.uc.UpdateSettings(ctx, userId, req.Settings)
	if err != nil {
		if ent.IsValidationError(err) {
			return nil, v1.ErrorInvalidRequest("invalid request: %s", err.Error())
		}
		return nil, v1.ErrorDatabaseQuery("database error: %s", err.Error())
	}

	return &v1.SettingsReply{
		Settings: settings,
	}, nil
}

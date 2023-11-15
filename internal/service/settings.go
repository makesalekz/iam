package service

import (
	"context"

	v1 "gitlab.calendaria.team/services/iam/api/iam/v1"
	"gitlab.calendaria.team/services/iam/ent"
	"gitlab.calendaria.team/services/iam/internal/biz"
	"gitlab.calendaria.team/services/iam/internal/data"
	utils_v1 "gitlab.calendaria.team/services/utils/api/utils/v1"

	"github.com/go-kratos/kratos/v2/log"
)

type SettingsService struct {
	v1.UnimplementedSettingsServer

	log *log.Helper
	jwt *data.JwtProcessor
	uc  *biz.SettingsUsecase
}

func NewSettingsService(logger log.Logger, jwt *data.JwtProcessor, uc *biz.SettingsUsecase) *SettingsService {
	return &SettingsService{
		log: log.NewHelper(logger),
		jwt: jwt,
		uc:  uc,
	}
}

func (s *SettingsService) GetSettings(ctx context.Context, req *utils_v1.EmptyRequest) (*v1.SettingsReply, error) {
	userId, ok := s.jwt.GetUserIdFromContext(ctx)
	if !ok {
		return nil, v1.ErrorUnauthorized("Unauthorized")
	}

	settings, err := s.uc.GetSettings(ctx, userId)
	if err != nil {
		return nil, v1.ErrorDatabaseQuery("Internal error")
	}

	return &v1.SettingsReply{
		Settings: settings,
	}, nil
}

func (s *SettingsService) UpdateSettings(ctx context.Context, req *v1.SettingsRequest) (*v1.SettingsReply, error) {
	userId, ok := s.jwt.GetUserIdFromContext(ctx)
	if !ok {
		return nil, v1.ErrorUnauthorized("Unauthorized")
	}

	settings, err := s.uc.UpdateSettings(ctx, userId, req.Settings)
	if err != nil {
		if ent.IsValidationError(err) {
			return nil, v1.ErrorInvalidRequest(err.Error())
		}
		s.log.Errorf("UpdatePrivacy error: %v", err)
		return nil, v1.ErrorDatabaseQuery("Internal error")
	}

	return &v1.SettingsReply{
		Settings: settings,
	}, nil
}

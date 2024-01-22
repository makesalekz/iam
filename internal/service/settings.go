package service

import (
	"context"

	v1 "gitlab.calendaria.team/services/iam/api/iam/v1"
	"gitlab.calendaria.team/services/iam/ent"
	"gitlab.calendaria.team/services/iam/internal/biz"
	utils_v1 "gitlab.calendaria.team/services/utils/api/utils/v1"
)

type SettingsService struct {
	v1.UnimplementedSettingsServer

	sh *ServiceHelper
	uc *biz.SettingsUsecase
}

func NewSettingsService(
	sh *ServiceHelper,
	uc *biz.SettingsUsecase,
) *SettingsService {
	return &SettingsService{
		sh: sh,
		uc: uc,
	}
}

func (s *SettingsService) GetSettings(ctx context.Context, req *utils_v1.ActorRequest) (*v1.SettingsReply, error) {
	actorId, err := s.sh.GetActorId(ctx, req.ActorId)
	if err != nil {
		return nil, err
	}

	settings, err := s.uc.GetSettings(ctx, actorId)
	if err != nil {
		return nil, v1.ErrorDatabaseQuery("database error: %s", err.Error())
	}

	return &v1.SettingsReply{
		Settings: settings,
	}, nil
}

func (s *SettingsService) UpdateSettings(ctx context.Context, req *v1.SettingsRequest) (*v1.SettingsReply, error) {
	actorId, err := s.sh.GetActorId(ctx, req.ActorId)
	if err != nil {
		return nil, err
	}

	settings, err := s.uc.UpdateSettings(ctx, actorId, req.Settings)
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

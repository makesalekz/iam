package service

import (
	"context"

	v1 "gitlab.calendaria.team/services/iam/api/iam/v1"
	"gitlab.calendaria.team/services/iam/ent"
	"gitlab.calendaria.team/services/iam/internal/biz"
	utils_v1 "gitlab.calendaria.team/services/utils/api/utils/v1"
	"gitlab.calendaria.team/services/utils/v2/auth"
)

type SettingsService struct {
	v1.UnimplementedSettingsServer

	uc *biz.SettingsUsecase
}

func NewSettingsService(
	uc *biz.SettingsUsecase,
) *SettingsService {
	return &SettingsService{
		uc: uc,
	}
}

func (s *SettingsService) GetSettings(ctx context.Context, req *utils_v1.EmptyRequest) (*v1.SettingsReply, error) {
	actorId := auth.GetActorIdFromContext(ctx)
	if actorId == 0 {
		return nil, v1.ErrorEmptyActorId("empty actor id")
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
	actorId := auth.GetActorIdFromContext(ctx)
	if actorId == 0 {
		return nil, v1.ErrorEmptyActorId("empty actor id")
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

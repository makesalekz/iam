package service

import (
	"context"

	v1 "github.com/makesalekz/iam/api/iam/v1"
	"github.com/makesalekz/iam/ent"
	"github.com/makesalekz/iam/internal/biz"
	utils_v1 "github.com/makesalekz/utils/api/utils/v1"
	"github.com/makesalekz/utils/v2/auth"
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
	actorID := auth.GetActorIdFromContext(ctx)
	if actorID == 0 {
		return nil, v1.ErrorEmptyActorId("empty actor id")
	}

	settings, err := s.uc.GetSettings(ctx, actorID)
	if err != nil {
		return nil, v1.ErrorDatabaseQuery("database error: %s", err.Error())
	}

	return &v1.SettingsReply{
		Settings: settings,
	}, nil
}

func (s *SettingsService) UpdateSettings(ctx context.Context, req *v1.SettingsRequest) (*v1.SettingsReply, error) {
	actorID := auth.GetActorIdFromContext(ctx)
	if actorID == 0 {
		return nil, v1.ErrorEmptyActorId("empty actor id")
	}

	settings, err := s.uc.UpdateSettings(ctx, actorID, req.GetSettings())
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

func (s *SettingsService) GetUsersSettings(
	ctx context.Context, req *v1.GetUsersSettingsRequest,
) (*v1.UsersSettingsReply, error) {
	actorID := auth.GetActorIdFromContext(ctx)
	if actorID == 0 {
		return nil, v1.ErrorEmptyActorId("empty actor id")
	}

	settings, err := s.uc.GetUsersSettings(ctx, req.GetUserIds())
	if err != nil {
		return nil, err
	}

	usersSettings := make(map[int64]*v1.SettingsReply)
	for userID, userSettings := range settings {
		usersSettings[userID] = &v1.SettingsReply{
			Settings: userSettings,
		}
	}

	return &v1.UsersSettingsReply{
		UsersSettings: usersSettings,
	}, nil
}

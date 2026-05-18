package biz

import (
	"context"

	v1 "github.com/makesalekz/iam/api/iam/v1"
	"github.com/makesalekz/iam/internal/data"
)

// SettingsUsecase .
type SettingsUsecase struct {
	settingsRepo data.SettingsRepo
}

// NewSettingsUsecase .
func NewSettingsUsecase(settingsRepo data.SettingsRepo) (*SettingsUsecase, error) {
	return &SettingsUsecase{
		settingsRepo: settingsRepo,
	}, nil
}

func (uc *SettingsUsecase) GetSettings(ctx context.Context, userID int64) (data.SettingsData, error) {
	return uc.settingsRepo.GetSettings(ctx, userID)
}

func (uc *SettingsUsecase) UpdateSettings(ctx context.Context, userID int64, data data.SettingsData) (
	data.SettingsData, error,
) {
	return uc.settingsRepo.UpdateSettings(ctx, userID, data)
}

func (uc *SettingsUsecase) GetUsersSettings(ctx context.Context, userIDs []int64) (map[int64]data.SettingsData, error) {
	settings, err := uc.settingsRepo.GetUsersSettings(ctx, userIDs)
	if err != nil {
		return nil, v1.ErrorDatabaseQuery("database error: %s", err.Error())
	}

	return settings, nil
}

package biz

import (
	"context"
	_ "embed"

	"gitlab.calendaria.team/alageum-cloud/iam/internal/data"
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

func (uc *SettingsUsecase) GetSettings(ctx context.Context, userId int64) (data.SettingsData, error) {
	return uc.settingsRepo.GetSettings(ctx, userId)
}

func (uc *SettingsUsecase) UpdateSettings(ctx context.Context, userId int64, data data.SettingsData) (data.SettingsData, error) {
	return uc.settingsRepo.UpdateSettings(ctx, userId, data)
}

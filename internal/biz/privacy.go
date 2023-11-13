package biz

import (
	"context"
	_ "embed"

	"gitlab.calendaria.team/services/iam/internal/data"
)

// PrivacyUsecase .
type PrivacyUsecase struct {
	privacyRepo data.PrivacyRepo
}

// NewPrivacyUsecase .
func NewPrivacyUsecase(privacyRepo data.PrivacyRepo) (*PrivacyUsecase, error) {
	return &PrivacyUsecase{
		privacyRepo: privacyRepo,
	}, nil
}

func (uc *PrivacyUsecase) GetPrivacy(ctx context.Context, userId int64) (data.PrivacySettingsData, error) {
	return uc.privacyRepo.GetPrivacy(ctx, userId)
}

func (uc *PrivacyUsecase) UpdatePrivacy(ctx context.Context, userId int64, data data.PrivacySettingsData) (data.PrivacySettingsData, error) {
	return uc.privacyRepo.UpdatePrivacy(ctx, userId, data)
}

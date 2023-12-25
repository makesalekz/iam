package biz

import (
	"context"

	iam_v1 "gitlab.calendaria.team/services/iam/api/iam/v1"
	"gitlab.calendaria.team/services/iam/internal/data"
)

type UserPrivaciesItem struct {
	UserId    int64
	Privacies map[string]string
}

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

func (uc *PrivacyUsecase) GetPrivacies(ctx context.Context, userIds []int64) ([]*UserPrivaciesItem, error) {
	usersPrivacies, err := uc.privacyRepo.GetPrivacies(ctx, userIds)
	if err != nil {
		return nil, iam_v1.ErrorServiceFailed("privacy: %s", err.Error())
	}

	privaciesMap := make(map[int64]map[string]string)
	for _, userPrivacies := range usersPrivacies {
		if privaciesMap[userPrivacies.UserID] == nil {
			privaciesMap[userPrivacies.UserID] = make(map[string]string)
		}
		privaciesMap[userPrivacies.UserID][string(userPrivacies.Setting)] = string(userPrivacies.Option)
	}

	userPrivaciesItems := make([]*UserPrivaciesItem, len(privaciesMap))
	i := 0
	for userId, privacies := range privaciesMap {
		userPrivaciesItems[i] = &UserPrivaciesItem{UserId: userId, Privacies: privacies}
		i++
	}

	return userPrivaciesItems, nil
}

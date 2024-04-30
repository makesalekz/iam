package data

import (
	"context"
	"slices"

	"entgo.io/ent/dialect/sql"
	"gitlab.calendaria.team/services/iam/ent"
	"gitlab.calendaria.team/services/iam/ent/enum"
	"gitlab.calendaria.team/services/iam/ent/userprivacy"
)

type PrivacySettingsData map[string]string

// PrivacyRepo
type PrivacyRepo interface {
	GetPrivacy(ctx context.Context, userId int64) (PrivacySettingsData, error)
	GetPrivacies(ctx context.Context, userId []int64) ([]*ent.UserPrivacy, error)
	UpdatePrivacy(ctx context.Context, userId int64, dto PrivacySettingsData) (PrivacySettingsData, error)
}

type privacyRepo struct {
	db *ent.Client
}

// NewPrivacyRepo .
func NewPrivacyRepo(d *Data) PrivacyRepo {
	return &privacyRepo{
		db: d.db,
	}
}

func DefaultPrivacies() map[string]string {
	return map[string]string{
		string(enum.MyLastActions):   string(enum.All),
		string(enum.MyProfileImage):  string(enum.All),
		string(enum.MyEvents):        string(enum.All),
		string(enum.GroupChatInvite): string(enum.All),
		string(enum.EventInvite):     string(enum.All),
		string(enum.MySlots):         string(enum.NoOne),
		string(enum.SlotsDetails):    string(enum.NoOne),
	}
}

func (r *privacyRepo) GetPrivacy(ctx context.Context, userId int64) (PrivacySettingsData, error) {
	settings, err := r.db.UserPrivacy.Query().Where(userprivacy.UserID(userId)).All(ctx)
	if err != nil {
		return nil, err
	}

	result := DefaultPrivacies()
	for _, setting := range settings {
		result[string(setting.Setting)] = string(setting.Option)
	}

	return result, nil
}

func (r *privacyRepo) GetPrivacies(ctx context.Context, userIds []int64) ([]*ent.UserPrivacy, error) {
	usersPrivacies, err := r.db.UserPrivacy.Query().
		Where(
			userprivacy.UserIDIn(userIds...),
		).
		All(ctx)
	if err != nil {
		return nil, err
	}

	return usersPrivacies, nil
}

func (r *privacyRepo) UpdatePrivacy(ctx context.Context, userId int64, dto PrivacySettingsData) (PrivacySettingsData, error) {
	var privacySettings enum.PrivacySettings
	settingsAvailable := privacySettings.Values()

	builders := make([]*ent.UserPrivacyCreate, 0)

	for setting, option := range dto {
		if !slices.Contains(settingsAvailable, setting) {
			return nil, ent.CustomValidationError("SettingsUnavailable", "Unavailable setting: %s", setting)
		}
		builder := r.db.UserPrivacy.Create().
			SetUserID(userId).
			SetSetting(enum.PrivacySettings(setting)).
			SetOption(enum.PrivacyOptions(option))

		builders = append(builders, builder)
	}

	err := r.db.UserPrivacy.CreateBulk(builders...).OnConflict(
		sql.ConflictColumns(userprivacy.FieldUserID, userprivacy.FieldSetting),
	).UpdateNewValues().Exec(ctx)
	if err != nil {
		return nil, err
	}

	return r.GetPrivacy(ctx, userId)
}

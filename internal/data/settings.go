package data

import (
	"context"
	"slices"

	"iam/ent"
	"iam/ent/property"
	"iam/ent/usersettings"

	"entgo.io/ent/dialect/sql"
)

type SettingsData map[string]string

// SettingsRepo
type SettingsRepo interface {
	GetSettings(ctx context.Context, userId int64) (SettingsData, error)
	UpdateSettings(ctx context.Context, userId int64, dto SettingsData) (SettingsData, error)
}

type settingsRepo struct {
	db *ent.Client
}

// NewSettingsRepo .
func NewSettingsRepo(d *Data) SettingsRepo {
	return &settingsRepo{
		db: d.db,
	}
}

func (r *settingsRepo) GetSettings(ctx context.Context, userId int64) (SettingsData, error) {
	settings, err := r.db.UserSettings.Query().Where(usersettings.UserID(userId)).All(ctx)
	if err != nil {
		return nil, err
	}

	result := make(SettingsData)
	for _, setting := range settings {
		result[string(setting.Setting)] = setting.Value
	}

	return result, nil
}

func (r *settingsRepo) UpdateSettings(ctx context.Context, userId int64, dto SettingsData) (SettingsData, error) {
	var settingsSettings property.Settings
	settingsAvailable := settingsSettings.Values()

	builders := make([]*ent.UserSettingsCreate, 0)

	for setting, value := range dto {
		if !slices.Contains(settingsAvailable, setting) {
			return nil, ent.CustomValidationError("SettingsUnavailable", "Unavailable setting: %s", setting)
		}
		builder := r.db.UserSettings.Create().
			SetUserID(userId).
			SetSetting(property.Settings(setting)).
			SetValue(value)

		builders = append(builders, builder)
	}

	err := r.db.UserSettings.CreateBulk(builders...).OnConflict(
		sql.ConflictColumns(usersettings.FieldUserID, usersettings.FieldSetting),
	).UpdateNewValues().Exec(ctx)
	if err != nil {
		return nil, err
	}

	return r.GetSettings(ctx, userId)
}

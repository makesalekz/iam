package data

import (
	"context"
	"slices"

	"entgo.io/ent/dialect/sql"

	"github.com/makesalekz/iam/ent"
	"github.com/makesalekz/iam/ent/enum"
	"github.com/makesalekz/iam/ent/usersettings"
)

type SettingsData map[string]string

// SettingsRepo.
type SettingsRepo interface {
	GetSettings(ctx context.Context, userID int64) (SettingsData, error)
	UpdateSettings(ctx context.Context, userID int64, dto SettingsData) (SettingsData, error)
	GetUsersSettings(ctx context.Context, userIDs []int64) (map[int64]SettingsData, error)
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

func (r *settingsRepo) GetSettings(ctx context.Context, userID int64) (SettingsData, error) {
	settings, err := r.db.UserSettings.Query().Where(usersettings.UserID(userID)).All(ctx)
	if err != nil {
		return nil, err
	}

	result := make(SettingsData)
	for _, setting := range settings {
		result[string(setting.Setting)] = setting.Value
	}

	return result, nil
}

func (r *settingsRepo) UpdateSettings(ctx context.Context, userID int64, dto SettingsData) (SettingsData, error) {
	var settingsSettings enum.Settings
	settingsAvailable := settingsSettings.Values()

	builders := make([]*ent.UserSettingsCreate, 0)

	for setting, value := range dto {
		if !slices.Contains(settingsAvailable, setting) {
			return nil, ent.CustomValidationError("SettingsUnavailable", "Unavailable setting: %s", setting)
		}
		builder := r.db.UserSettings.Create().
			SetUserID(userID).
			SetSetting(enum.Settings(setting)).
			SetValue(value)

		builders = append(builders, builder)
	}

	err := r.db.UserSettings.CreateBulk(builders...).OnConflict(
		sql.ConflictColumns(usersettings.FieldUserID, usersettings.FieldSetting),
	).UpdateNewValues().Exec(ctx)
	if err != nil {
		return nil, err
	}

	return r.GetSettings(ctx, userID)
}

func (r *settingsRepo) GetUsersSettings(ctx context.Context, userIDs []int64) (map[int64]SettingsData, error) {
	if len(userIDs) == 0 {
		return nil, nil
	}

	settings, err := r.db.UserSettings.Query().
		Where(usersettings.UserIDIn(userIDs...)).
		All(ctx)

	if err != nil {
		return nil, err
	}

	result := make(map[int64]SettingsData)
	for _, setting := range settings {
		if _, ok := result[setting.UserID]; !ok {
			result[setting.UserID] = make(SettingsData)
		}
		result[setting.UserID][string(setting.Setting)] = setting.Value
	}

	return result, nil
}

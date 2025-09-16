package enum

type Settings string

const (
	Language                     Settings = "LANGUAGE"
	Theme                        Settings = "THEME"
	NotificationSoundEnabled     Settings = "NOTIFICATION_SOUND_ENABLED"
	NotificationVibrationEnabled Settings = "NOTIFICATION_VIBRATION_ENABLED"
	EventsChatEnabled            Settings = "EVENTS_CHAT_ENABLED"
)

// Values provides list valid values for Enum.
func (Settings) Values() (kinds []string) {
	for _, s := range []Settings{
		Language, Theme, NotificationSoundEnabled, NotificationVibrationEnabled, EventsChatEnabled,
	} {
		kinds = append(kinds, string(s))
	}
	return
}

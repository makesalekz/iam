package enum

type Settings string

const (
	TestSetting Settings = "TEST_SETTING"
	Language    Settings = "LANGUAGE"
	Theme       Settings = "THEME"
)

// Values provides list valid values for Enum.
func (Settings) Values() (kinds []string) {
	for _, s := range []Settings{TestSetting, Language, Theme} {
		kinds = append(kinds, string(s))
	}
	return
}

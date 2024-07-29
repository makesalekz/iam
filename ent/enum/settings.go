package enum

type Settings string

const (
	Language Settings = "LANGUAGE"
	Theme    Settings = "THEME"
)

// Values provides list valid values for Enum.
func (Settings) Values() (kinds []string) {
	for _, s := range []Settings{Language, Theme} {
		kinds = append(kinds, string(s))
	}
	return
}

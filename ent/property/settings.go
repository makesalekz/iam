package property

type Settings string

const (
	TestSetting Settings = "TEST_SETTING"
)

// Values provides list valid values for Enum.
func (Settings) Values() (kinds []string) {
	for _, s := range []Settings{TestSetting} {
		kinds = append(kinds, string(s))
	}
	return
}

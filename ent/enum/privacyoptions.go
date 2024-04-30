package enum

type PrivacyOptions string

const (
	All        PrivacyOptions = "ALL"
	MyContacts PrivacyOptions = "MY_CONTACTS"
	NoOne      PrivacyOptions = "NO_ONE"
)

// Values provides list valid values for Enum.
func (PrivacyOptions) Values() (kinds []string) {
	for _, s := range []PrivacyOptions{All, MyContacts, NoOne} {
		kinds = append(kinds, string(s))
	}
	return
}

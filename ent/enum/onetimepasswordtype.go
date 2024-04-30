package enum

type OneTimePasswordType string

const (
	Email OneTimePasswordType = "EMAIL"
	Phone OneTimePasswordType = "PHONE"
)

// Values provides list valid values for Enum.
func (OneTimePasswordType) Values() (kinds []string) {
	for _, s := range []OneTimePasswordType{Email, Phone} {
		kinds = append(kinds, string(s))
	}
	return
}

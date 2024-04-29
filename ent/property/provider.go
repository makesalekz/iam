package property

type Provider string

const (
	Google  Provider = "GOOGLE"
	Outlook Provider = "OUTLOOK"
	Apple   Provider = "APPLE"
)

func (Provider) Values() (kinds []string) {
	for _, s := range []Provider{Google, Outlook, Apple} {
		kinds = append(kinds, string(s))
	}
	return
}

func (p Provider) Value() string {
	return string(p)
}

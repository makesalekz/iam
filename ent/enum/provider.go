package enum

import "slices"

type Provider string

const (
	Google  Provider = "GOOGLE"
	Outlook Provider = "OUTLOOK"
	Apple   Provider = "APPLE"
)

func getProviders() []Provider {
	return []Provider{Google, Outlook, Apple}
}

func (Provider) Values() (kinds []string) {
	for _, s := range getProviders() {
		kinds = append(kinds, string(s))
	}
	return
}

func (p Provider) Value() string {
	return string(p)
}

func (p Provider) IsValid() bool {
	return slices.Contains(getProviders(), p)
}

package enum

type SxodimGrantType string

const (
	Authorization SxodimGrantType = "authorization_code"
	RefreshToken  SxodimGrantType = "refresh_token"
)

func grandTypeValues() []SxodimGrantType {
	return []SxodimGrantType{Authorization, RefreshToken}
}

func (SxodimGrantType) Values() (kinds []string) {
	for _, value := range grandTypeValues() {
		kinds = append(kinds, string(value))
	}
	return
}

func (m SxodimGrantType) Value() string {
	return string(m)
}

func (m SxodimGrantType) IsValid() bool {
	for _, value := range grandTypeValues() {
		if m == value {
			return true
		}
	}
	return false
}

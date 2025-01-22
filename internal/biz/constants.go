package biz

const (
	QueueContactsPhoneVerified  = "contacts.confirmed_phone"
	QueueContactsEmailVerified  = "contacts.confirmed_emails"
	QueueEventsDefaultCalendars = "events.default_calendars"
)

// --------------------- Sxodim ---------------------.
const (
	SxodimGrantType = "authorization_code"
	SxodimClientID  = "3"
)

type SxodimAuthRequestBody struct {
	GrantType    string `json:"grant_type"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Code         string `json:"code"`
}

type SxodimAuthResponseBody struct {
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"` // ms from time.Now
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

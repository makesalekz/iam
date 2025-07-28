package biz

const (
	QueueContactsPhoneVerified  = "contacts.confirmed_phone"
	QueueContactsEmailVerified  = "contacts.confirmed_emails"
	QueueEventsDefaultCalendars = "events.default_calendars"
	QueueDeleteDeviceTokens     = "notifications.delete_tokens"
)

type NotificationType string

const (
	NotifyAccountDeletion NotificationType = "ACCOUNT_DELETED"
)

// String returns the string representation of the notification type.
func (n NotificationType) String() string {
	return string(n)
}

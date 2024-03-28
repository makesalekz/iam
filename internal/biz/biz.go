package biz

import (
	"github.com/google/wire"
	"gitlab.calendaria.team/services/utils/v1/nats"
)

const (
	QueueContactsPhoneVerified  = "contacts/confirmed_phone"
	QueueContactsEmailVerified  = "contacts/confirmed_emails"
	QueueEventsDefaultCalendars = "events/default_calendars"
)

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(
	nats.NewQueueManager,
	NewAuthUsecase,
	NewUsersUsecase,
	NewPrivacyUsecase,
	NewSettingsUsecase,
)

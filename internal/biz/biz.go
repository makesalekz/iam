package biz

import (
	"github.com/google/wire"
)

const (
	QueueContactsPhoneVerified = "gitlab.calendaria.team/services/contacts/confirmed_phone"
	QueueContactsEmailVerified = "gitlab.calendaria.team/services/contacts/confirmed_emails"
)

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(
	NewQueueManager,
	NewAuthUsecase,
	NewUsersUsecase,
	NewPrivacyUsecase,
	NewSettingsUsecase,
)

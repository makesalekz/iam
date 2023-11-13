package biz

import (
	"github.com/google/wire"
)

const (
	QueueContactsPhoneVerified = "contacts/confirmed_phone"
	QueueContactsEmailVerified = "contacts/confirmed_emails"
)

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(
	NewQueueManager,
	NewAuthUsecase,
	NewUsersUsecase,
	NewPrivacyUsecase,
	NewSettingsUsecase,
)

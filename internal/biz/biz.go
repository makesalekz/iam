package biz

import (
	"github.com/google/wire"
	"gitlab.calendaria.team/services/utils/v2/nats"
)

// ProviderSet is biz providers.
//
//nolint:gochecknoglobals // this global variable is required for wire
var ProviderSet = wire.NewSet(
	nats.NewQueueManager,
	NewAuthUsecase,
	NewUsersUsecase,
	NewPrivacyUsecase,
	NewSettingsUsecase,
	NewCredentialsUsecase,
)

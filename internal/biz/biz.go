package biz

import (
	"gitlab.calendaria.team/services/iam/internal/data/remote"
	"gitlab.calendaria.team/services/utils/v4/nats"

	"github.com/google/wire"
)

// ProviderSet is biz providers.
//
//nolint:gochecknoglobals // this global variable is required for wire
var ProviderSet = wire.NewSet(
	nats.NewQueueManager,
	remote.NewWebsocketsRemote,
	NewAuthUsecase,
	NewUsersUsecase,
	NewPrivacyUsecase,
	NewSettingsUsecase,
	NewCredentialsUsecase,
)

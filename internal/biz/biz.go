package biz

import (
	"github.com/makesalekz/iam/internal/data/remote"
	"github.com/makesalekz/utils/v4/nats"

	"github.com/google/wire"
)

// ProviderSet is biz providers.
//
//nolint:gochecknoglobals // this global variable is required for wire
var ProviderSet = wire.NewSet(
	nats.NewQueueManager,
	remote.NewWebsocketsRemote,
	remote.NewChatsRemote,
	remote.NewContactsRemote,
	remote.NewEventsRemote,
	NewAuthUsecase,
	NewUsersUsecase,
	NewPrivacyUsecase,
	NewSettingsUsecase,
	NewCredentialsUsecase,
)

package service

import "github.com/google/wire"

// ProviderSet is service providers.
//
//nolint:gochecknoglobals // this global variable is required for wire
var ProviderSet = wire.NewSet(
	NewAuthService,
	NewUsersService,
	NewPrivacyService,
	NewSettingsService,
	NewCredentialsService,
)

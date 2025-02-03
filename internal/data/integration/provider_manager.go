package integration

import (
	"gitlab.calendaria.team/services/utils/v1/config"
	u_struc "gitlab.calendaria.team/services/utils/v2/struc"
)

type IProviderManager interface {
	NewProviderGateway(config *config.Config, provider u_struc.Provider) (IProviderGateway, error)
}

// ProviderManager is a service dialer manager
type ProviderManager struct {
	config *config.Config
}

func NewProviderManager(
	config *config.Config,
) (IProviderManager, error) {
	return &ProviderManager{
		config: config,
	}, nil
}

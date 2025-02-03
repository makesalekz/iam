package integration

import (
	"gitlab.calendaria.team/services/utils/v1/config"
	u_struc "gitlab.calendaria.team/services/utils/v2/struc"

	"github.com/go-kratos/kratos/v2/log"
)

type IProviderManager interface {
	NewProviderGateway(provider u_struc.Provider) (IProviderGateway, error)
}

// ProviderManager is a service dialer manager
type ProviderManager struct {
	config *config.Config
	log    *log.Helper
}

func NewProviderManager(
	config *config.Config,
	logger log.Logger,
) (IProviderManager, error) {
	return &ProviderManager{
		config: config,
		log:    log.NewHelper(log.With(logger, "module", "data/provider")),
	}, nil
}

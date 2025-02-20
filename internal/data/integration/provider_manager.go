package integration

import (
	iam_v1 "gitlab.calendaria.team/services/iam/api/iam/v1"
	u_struc "gitlab.calendaria.team/services/utils/v2/struc"
	"gitlab.calendaria.team/services/utils/v4/config"
)

type IProviderManager interface {
	NewProviderGateway(provider u_struc.Provider) (IProviderGateway, error)
}

type ProviderManager struct {
	google IProviderGateway
	sxodim IProviderGateway
}

func NewProviderManager(
	config config.IConfig,
) (IProviderManager, error) {
	googleRemote, err := NewGoogleRemote(config)
	if err != nil {
		return nil, err
	}

	sxodimRemote, err := NewSxodimRemote(config)
	if err != nil {
		return nil, err
	}

	return &ProviderManager{
		google: googleRemote,
		sxodim: sxodimRemote,
	}, nil
}

func (dm *ProviderManager) NewProviderGateway(
	provider u_struc.Provider,
) (IProviderGateway, error) {
	switch provider {
	case u_struc.Google:
		return dm.google, nil
	case u_struc.Sxodim:
		return dm.sxodim, nil
	}

	return nil, iam_v1.ErrorNotFound("unknown provider")
}

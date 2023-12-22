package data

import (
	"context"
	"os"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"gitlab.calendaria.team/services/iam/ent"
	"gitlab.calendaria.team/services/iam/internal/conf"
	"gitlab.calendaria.team/services/utils/v1/config"
	"gitlab.calendaria.team/services/utils/v1/dialer"
	"gitlab.calendaria.team/services/utils/v1/jwt"

	_ "github.com/lib/pq"
	_ "gitlab.calendaria.team/services/iam/ent/runtime"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(
	NewData,
	config.NewConfig,
	jwt.NewJwtProcessor,
	NewNatsClient,
	NewUsersRepo,
	NewOtpRepo,
	NewPrivacyRepo,
	NewSettingsRepo,
	dialer.NewDialer,
	NewContactsRemote,
	NewNotificationsRemote,
	NewTenantsRemote,
)

// Data .
type Data struct {
	db *ent.Client
}

// NewData .
func NewData(c *conf.Bootstrap, logger log.Logger) (*Data, func(), error) {
	l := log.NewHelper(logger)

	client, err := ent.Open("postgres", c.Db.Address)
	if err != nil {
		l.Fatalf("failed opening connection to postgres: %v", err)
		return nil, nil, err
	}

	automigrate := os.Getenv("AUTOMIGRATE")
	if automigrate != "" {
		if err := client.Schema.Create(context.Background()); err != nil {
			l.Errorf("failed creating schema resources: %v", err)
			return nil, nil, err
		}
	}

	l.Info("Connected to postgres")

	cleanup := func() {
		if err := client.Close(); err != nil {
			l.Error(err)
		}
	}

	return &Data{
		db: client,
	}, cleanup, nil
}

package data

import (
	"context"
	"fmt"
	"os"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"github.com/nats-io/nats.go"
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
func NewData(bc *conf.Bootstrap, c *config.Config, logger log.Logger) (*Data, func(), error) {
	l := log.NewHelper(logger)

	dbDsn := bc.Db // read from local config
	if dbDsn == "" {
		// read from vault
		secret, err := c.ReadSecretsFor(context.Background(), "db-dsn")
		if err != nil {
			l.Fatalf("db dsn not found: %v", err)
			return nil, nil, err
		}
		dbDsn = secret["data"].(string)
	}

	// TODO remove <----------------
	nc, err := nats.Connect(bc.Nats)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to nats: %w", err)
	}

	err = nc.Publish("chats/send_message", []byte("fail test"))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to publish to nats: %w", err)
	}

	testConsul := c.Value("TEST_CONSUL")
	l.Debugf("CONSUL TEST: %s", testConsul)
	// TODO remove ---------------->

	l.Debugf("Connecting to postgres: ", dbDsn)

	client, err := ent.Open("postgres", dbDsn)
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
		nc.Close()
	}

	return &Data{
		db: client,
	}, cleanup, nil
}

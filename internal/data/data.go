package data

import (
	"context"
	"os"

	"github.com/makesalekz/iam/ent"
	"github.com/makesalekz/iam/internal/conf"
	"github.com/makesalekz/iam/internal/data/integration"
	"github.com/makesalekz/iam/internal/data/remote"
	u_config "github.com/makesalekz/utils/v4/config"
	u_dialer "github.com/makesalekz/utils/v4/dialer"
	u_jwt "github.com/makesalekz/utils/v4/jwt"
	u_tracing "github.com/makesalekz/utils/v4/tracing"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"

	_ "github.com/lib/pq"
	_ "github.com/makesalekz/iam/ent/runtime"
)

// ProviderSet is data providers.
//
//nolint:gochecknoglobals // this global variable is required for wire
var ProviderSet = wire.NewSet(
	NewData,
	u_config.NewConfig,
	u_jwt.NewJwtProcessor,
	u_dialer.NewServiceDialerManager,
	u_tracing.NewTracer,
	NewNatsClient,
	remote.NewNotificationsRemote,
	remote.NewTenantsRemote,
	remote.NewMediaRemote,
	integration.NewProviderManager,
	NewUsersRepo,
	NewOtpRepo,
	NewPrivacyRepo,
	NewSettingsRepo,
	NewCredentialsRepo,
)

// Data .
type Data struct {
	db *ent.Client
}

const CodeInvalid = 500

// NewData .
func NewData(bc *conf.Bootstrap, c u_config.IConfig, logger log.Logger) (*Data, func(), error) {
	l := log.NewHelper(logger)

	dbDsn := bc.GetDb() // read from local config
	if dbDsn == "" {
		// read from vault
		secret, err := c.ReadSecretsFor(context.Background(), "db-dsn")
		if err != nil {
			l.Fatalf("db dsn not found: %v", err)
			return nil, nil, err
		}

		secretData, ok := secret["data"].(string)
		if !ok {
			return nil, nil, errors.New(CodeInvalid, "internal error", "db dsn data not found")
		}

		dbDsn = secretData
	}

	autoMigrate := os.Getenv("AUTOMIGRATE")
	entLogging := os.Getenv("ENT_LOGGING")
	var options []ent.Option
	if entLogging == "true" {
		options = append(options, ent.Debug(), ent.Log(l.Info))
	}

	client, err := ent.Open("postgres", dbDsn, options...)
	if err != nil {
		l.Fatalf("failed opening connection to postgres: %v", err)
		return nil, nil, err
	}

	if autoMigrate != "" {
		if err2 := client.Schema.Create(context.Background()); err2 != nil {
			l.Errorf("failed creating schema resources: %v", err2)
			return nil, nil, err2
		}
	}

	client.Use(func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			ctx = ent.NewContext(ctx, client)
			return next.Mutate(ctx, m)
		})
	})

	l.Info("Connected to postgres")

	cleanup := func() {
		if err2 := client.Close(); err2 != nil {
			l.Error(err2)
		}
	}

	return &Data{
		db: client,
	}, cleanup, nil
}

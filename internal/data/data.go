package data

import (
	"context"

	"iam/ent"
	"iam/internal/conf"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewConfig, NewUsersRepo)

// Data .
type Data struct {
	log  *log.Helper
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

	if err := client.Schema.Create(context.Background()); err != nil {
		l.Errorf("failed creating schema resources: %v", err)
		return nil, nil, err
    }

	cleanup := func() {
		client.Close()
	}

	return &Data{
		log:  log.NewHelper(logger),
		db: client,
	}, cleanup, nil
}

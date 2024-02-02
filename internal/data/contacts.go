package data

import (
	"context"

	v1 "gitlab.calendaria.team/services/contacts/api/contacts/v1"
	"gitlab.calendaria.team/services/iam/internal/conf"
	"gitlab.calendaria.team/services/utils/v1/config"
	jwtp "gitlab.calendaria.team/services/utils/v1/jwt"
	"gitlab.calendaria.team/services/utils/v2/dialer"
)

type ContactsRemote struct {
	dialer *dialer.Dialer
}

func NewContactsRemote(
	conf *conf.Bootstrap,
	c *config.Config,
	jwt *jwtp.JwtProcessor,
) (*ContactsRemote, error) {
	dialer, err := dialer.NewServiceDialer(c, jwt, "contacts", conf.Discovery.Contacts)
	if err != nil {
		return nil, err
	}

	return &ContactsRemote{
		dialer: dialer,
	}, nil
}

func (r *ContactsRemote) getRelationClient(ctx context.Context) (v1.RelationsClient, error) {
	conn, err := r.dialer.Connect(ctx)
	if err != nil {
		return nil, v1.ErrorGrpcConnection("can't connect to iam: %s", err.Error())
	}

	return v1.NewRelationsClient(conn), nil
}

func (r *ContactsRemote) GetRelations(ctx context.Context, req *v1.GetRelationsRequest) (*v1.UserRelationsReply, error) {
	relationsClient, err := r.getRelationClient(ctx)
	if err != nil {
		return nil, err
	}

	relations, err := relationsClient.GetRelations(ctx, req)
	if err != nil {
		return nil, err
	}

	return relations, nil
}

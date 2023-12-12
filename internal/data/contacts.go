package data

import (
	"context"

	contacts_v1 "gitlab.calendaria.team/services/contacts/api/contacts/v1"
	iam_v1 "gitlab.calendaria.team/services/iam/api/iam/v1"
	"gitlab.calendaria.team/services/iam/internal/conf"
	"gitlab.calendaria.team/services/utils/v1/dialer"
)

type ContactsRemote struct {
	dialer *dialer.Dialer
	conf   *conf.Bootstrap
}

func NewContactsRemote(dialer *dialer.Dialer, conf *conf.Bootstrap) (*ContactsRemote, error) {
	return &ContactsRemote{
		dialer: dialer,
		conf:   conf,
	}, nil
}

func (r *ContactsRemote) GetContactsClient(ctx context.Context) (contacts_v1.ContactsClient, error) {
	return dialer.NewDialerBuilder(r.dialer, contacts_v1.NewContactsClient).
		SetEndpoint(r.conf.Discovery.Contacts).
		SetTimeout(r.conf.Discovery.ContactsTimeout.AsDuration()).
		Conn(ctx, nil)
}

func (r *ContactsRemote) GetRelationClient(ctx context.Context) (contacts_v1.RelationsClient, error) {
	return dialer.NewDialerBuilder(r.dialer, contacts_v1.NewRelationsClient).
		SetEndpoint(r.conf.Discovery.Contacts).
		SetTimeout(r.conf.Discovery.ContactsTimeout.AsDuration()).
		Conn(ctx, nil)
}

func (r *ContactsRemote) GetLabelsByUserId(ctx context.Context, req *contacts_v1.GetLabelsByUserIdRequest) (*contacts_v1.GetLabelsByUserIdResponse, error) {
	contactClient, err := r.GetContactsClient(ctx)
	if err != nil {
		return nil, iam_v1.ErrorGrpcConnection("contacts: %s", err.Error())
	}

	labels, err := contactClient.GetLabelsByUserId(ctx, req)
	if err != nil {
		return nil, iam_v1.ErrorServiceFailed("contacts: %s", err.Error())
	}

	return labels, nil
}

func (r *ContactsRemote) GetRelations(ctx context.Context, req *contacts_v1.GetRelationsRequest) (*contacts_v1.UserRelationsReply, error) {
	relationsClient, err := r.GetRelationClient(ctx)
	if err != nil {
		return nil, iam_v1.ErrorGrpcConnection("contacts: %s", err.Error())
	}

	relations, err := relationsClient.GetRelations(ctx, req)
	if err != nil {
		return nil, iam_v1.ErrorServiceFailed("contacts: %s", err.Error())
	}

	return relations, nil
}

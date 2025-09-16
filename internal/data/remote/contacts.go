// nolint: dupl // different services
package remote

import (
	"context"

	contacts_v1 "gitlab.calendaria.team/services/contacts/api/contacts/v1"
	iam_v1 "gitlab.calendaria.team/services/iam/api/iam/v1"
	"gitlab.calendaria.team/services/iam/internal/conf"
	u_dialer "gitlab.calendaria.team/services/utils/v4/dialer"
)

type IContactsRemote interface {
	GetIncomingRelations(
		ctx context.Context, req *contacts_v1.GetRelationsRequest,
	) (map[int64]*contacts_v1.Relation, error)
	DeleteUsersDataInContacts(ctx context.Context, usersIDs []int64) error
}

type ContactsRemote struct {
	dialer u_dialer.IDialer
}

func NewContactsRemote(
	conf *conf.Bootstrap,
	dm u_dialer.IDialerManager,
) (IContactsRemote, func(), error) {
	dialer, err := dm.NewServiceDialer("contacts", conf.GetDiscovery().GetContacts())
	if err != nil {
		return nil, nil, err
	}

	cleanup := func() {
		dialer.Close()
	}

	return &ContactsRemote{
		dialer: dialer,
	}, cleanup, nil
}

func (r *ContactsRemote) getContactsClient(ctx context.Context) (contacts_v1.ContactsClient, error) {
	conn, err := r.dialer.Connect(ctx)
	if err != nil {
		return nil, iam_v1.ErrorGrpcConnection("can't connect to contacts: %s", err.Error())
	}

	return contacts_v1.NewContactsClient(conn), nil
}

func (r *ContactsRemote) getRelationClient(ctx context.Context) (contacts_v1.RelationsClient, error) {
	conn, err := r.dialer.Connect(ctx)
	if err != nil {
		return nil, iam_v1.ErrorGrpcConnection("can't connect to contacts: %s", err.Error())
	}

	return contacts_v1.NewRelationsClient(conn), nil
}

func (r *ContactsRemote) DeleteUsersDataInContacts(
	ctx context.Context,
	usersIDs []int64,
) error {
	contactsClient, err := r.getContactsClient(ctx)
	if err != nil {
		return err
	}

	_, err = contactsClient.DeleteUsersData(ctx, &contacts_v1.DeleteUsersDataRequest{UserIds: usersIDs})
	if err != nil {
		return err
	}

	return nil
}

func (r *ContactsRemote) GetIncomingRelations(
	ctx context.Context, req *contacts_v1.GetRelationsRequest,
) (map[int64]*contacts_v1.Relation, error) {
	relationsClient, err := r.getRelationClient(ctx)
	if err != nil {
		return nil, err
	}

	relations, err := relationsClient.GetIncomingRelations(ctx, req)
	if err != nil {
		return nil, err
	}

	mapRelations := make(map[int64]*contacts_v1.Relation)
	for _, relation := range relations.GetRelations() {
		mapRelations[relation.GetUserId()] = relation
	}

	return mapRelations, nil
}

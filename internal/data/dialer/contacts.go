//nolint: dupl // different services
package dialer

import (
	"context"

	contacts_v1 "gitlab.calendaria.team/services/contacts/api/contacts/v1"
	iam_v1 "gitlab.calendaria.team/services/iam/api/iam/v1"
	"gitlab.calendaria.team/services/iam/internal/conf"
	u_dialer "gitlab.calendaria.team/services/utils/v2/dialer"
)

type IContactsRemote interface {
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

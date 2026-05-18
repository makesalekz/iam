package remote

import "context"

type IContactsRemote interface {
	// GetIncomingRelations returns map[userID]isExistInContacts
	GetIncomingRelations(ctx context.Context, userIDs []int64) (map[int64]bool, error)
	DeleteUsersDataInContacts(ctx context.Context, usersIDs []int64) error
}

type contactsRemoteStub struct{}

func NewContactsRemote() IContactsRemote {
	return &contactsRemoteStub{}
}

func (r *contactsRemoteStub) GetIncomingRelations(_ context.Context, _ []int64) (map[int64]bool, error) {
	return nil, nil
}

func (r *contactsRemoteStub) DeleteUsersDataInContacts(_ context.Context, _ []int64) error {
	return nil
}

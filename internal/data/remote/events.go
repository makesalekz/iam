package remote

import "context"

type IEventsRemote interface {
	DeleteUsersDataInEvents(ctx context.Context, usersIDs []int64) error
	DisconnectExternalCalendarsBulk(ctx context.Context, credentialID int64) error
}

type eventsRemoteStub struct{}

func NewEventsRemote() IEventsRemote {
	return &eventsRemoteStub{}
}

func (r *eventsRemoteStub) DeleteUsersDataInEvents(_ context.Context, _ []int64) error {
	return nil
}

func (r *eventsRemoteStub) DisconnectExternalCalendarsBulk(_ context.Context, _ int64) error {
	return nil
}

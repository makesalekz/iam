//nolint: dupl // different services
package data

import (
	"context"

	events_v1 "gitlab.calendaria.team/services/events/api/events/v1"
	iam_v1 "gitlab.calendaria.team/services/iam/api/iam/v1"
	"gitlab.calendaria.team/services/iam/internal/conf"
	"gitlab.calendaria.team/services/utils/v2/dialer"
)

type IEventsRemote interface {
	DeleteUsersDataInEvents(ctx context.Context, usersIDs []int64) error
}

type EventsRemote struct {
	dialer dialer.IDialer
}

func NewEventsRemote(
	conf *conf.Bootstrap,
	dm dialer.IDialerManager,
) (IEventsRemote, func(), error) {
	dialer, err := dm.NewServiceDialer("events", conf.GetDiscovery().GetEvents())
	if err != nil {
		return nil, nil, err
	}

	cleanup := func() {
		dialer.Close()
	}

	return &EventsRemote{
		dialer: dialer,
	}, cleanup, nil
}

func (r *EventsRemote) getEventsClient(ctx context.Context) (events_v1.EventsClient, error) {
	conn, err := r.dialer.Connect(ctx)
	if err != nil {
		return nil, iam_v1.ErrorGrpcConnection("can't connect to events: %s", err.Error())
	}

	return events_v1.NewEventsClient(conn), nil
}

func (r *EventsRemote) DeleteUsersDataInEvents(ctx context.Context, usersIDs []int64) error {
	client, err := r.getEventsClient(ctx)
	if err != nil {
		return err
	}

	_, err = client.DeleteUsersData(ctx, &events_v1.DeleteUsersDataRequest{UsersIds: usersIDs})
	if err != nil {
		return err
	}

	return nil
}

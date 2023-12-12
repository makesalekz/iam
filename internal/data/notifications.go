package data

import (
	"context"

	iam_v1 "gitlab.calendaria.team/services/iam/api/iam/v1"
	"gitlab.calendaria.team/services/iam/internal/conf"
	notifications_v1 "gitlab.calendaria.team/services/notifications/api/notifications/v1"
	"gitlab.calendaria.team/services/utils/v1/dialer"
)

type NotificationsRemote struct {
	dialer *dialer.Dialer
	conf   *conf.Bootstrap
}

func NewNotificationsRemote(d *dialer.Dialer, conf *conf.Bootstrap) (*NotificationsRemote, error) {
	return &NotificationsRemote{
		dialer: d,
		conf:   conf,
	}, nil
}

func (r *NotificationsRemote) GetNotificationsClient(ctx context.Context) (notifications_v1.NotificationsClient, error) {
	return dialer.NewDialerBuilder(r.dialer, notifications_v1.NewNotificationsClient).
		SetEndpoint(r.conf.Discovery.Notifications).
		SetTimeout(r.conf.Discovery.NotificationsTimeout.AsDuration()).
		Conn(ctx, nil)
}

func (r *NotificationsRemote) GetSenderClient(ctx context.Context) (notifications_v1.SenderClient, error) {
	return dialer.NewDialerBuilder(r.dialer, notifications_v1.NewSenderClient).
		SetEndpoint(r.conf.Discovery.Notifications).
		SetTimeout(r.conf.Discovery.NotificationsTimeout.AsDuration()).
		Conn(ctx, nil)
}

func (r *NotificationsRemote) PersonalSmsSender(ctx context.Context, phone, message string) error {
	client, err := r.GetSenderClient(ctx)
	if err != nil {
		return iam_v1.ErrorGrpcConnection("notifications: %s", err.Error())
	}

	_, err = client.PersonalSmsSender(ctx, &notifications_v1.PersonalSmsSenderRequest{
		Phone:   phone,
		Message: message,
	})
	if err != nil {
		return iam_v1.ErrorServiceFailed("notifications: %s", err.Error())
	}

	return nil
}

package data

import (
	"context"

	v1 "gitlab.calendaria.team/services/contacts/api/contacts/v1"
	"gitlab.calendaria.team/services/iam/internal/conf"
	notifications_v1 "gitlab.calendaria.team/services/notifications/api/notifications/v1"
	"gitlab.calendaria.team/services/utils/v1/config"
	jwtp "gitlab.calendaria.team/services/utils/v1/jwt"
	"gitlab.calendaria.team/services/utils/v2/dialer"
)

type NotificationsRemote struct {
	dialer *dialer.Dialer
	conf   *conf.Bootstrap
}

func NewNotificationsRemote(
	conf *conf.Bootstrap,
	c *config.Config,
	jwt *jwtp.JwtProcessor,
) (*NotificationsRemote, error) {
	dialer, err := dialer.NewServiceDialer(c, jwt, "notifications", conf.Discovery.Notifications)
	if err != nil {
		return nil, err
	}

	return &NotificationsRemote{
		dialer: dialer,
		conf:   conf,
	}, nil
}

func (r *NotificationsRemote) GetSenderClient(ctx context.Context) (notifications_v1.SenderClient, error) {
	conn, err := r.dialer.Connect(ctx)
	if err != nil {
		return nil, v1.ErrorGrpcConnection("can't connect to iam: %s", err.Error())
	}

	return notifications_v1.NewSenderClient(conn), nil
}

func (r *NotificationsRemote) PersonalSmsSender(ctx context.Context, phone, message string) error {
	client, err := r.GetSenderClient(ctx)
	if err != nil {
		return err
	}

	_, err = client.PersonalSmsSender(ctx, &notifications_v1.PersonalSmsSenderRequest{
		Phone:   phone,
		Message: message,
	})
	if err != nil {
		return err
	}

	return nil
}

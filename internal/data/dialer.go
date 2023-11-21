package data

import (
	"context"

	consul "github.com/go-kratos/consul/registry"
	"github.com/go-kratos/kratos/v2/middleware"
	kjwt "github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	jwtv4 "github.com/golang-jwt/jwt/v4"
	chats_v1 "gitlab.calendaria.team/services/chats/api/chats/v1"
	contacts_v1 "gitlab.calendaria.team/services/contacts/api/contacts/v1"
	v1 "gitlab.calendaria.team/services/iam/api/iam/v1"
	"gitlab.calendaria.team/services/iam/internal/conf"
	notifications_v1 "gitlab.calendaria.team/services/notifications/api/notifications/v1"
	"gitlab.calendaria.team/services/utils/v1/config"
	"gitlab.calendaria.team/services/utils/v1/jwt"
)

type Dialer struct {
	conf      *conf.Bootstrap
	discovery *consul.Registry
	jwt       *jwt.JwtProcessor
}

func NewDialer(c *config.Config, conf *conf.Bootstrap, jwt *jwt.JwtProcessor) (*Dialer, error) {
	return &Dialer{
		conf:      conf,
		discovery: c.GetRegistry(),
		jwt:       jwt,
	}, nil
}

func (d *Dialer) getJwtMiddleware(ctx context.Context) middleware.Middleware {
	return kjwt.Client(func(token *jwtv4.Token) (interface{}, error) {
		return d.jwt.GetSecret(), nil
	}, kjwt.WithSigningMethod(jwtv4.SigningMethodHS256), kjwt.WithClaims(func() jwtv4.Claims {
		claims, _ := d.jwt.GetClaimsFromContext(ctx)
		return claims
	}))
}

func (d *Dialer) Contacts(ctx context.Context) (contacts_v1.ContactsClient, error) {
	conn, err := grpc.DialInsecure(
		ctx,
		grpc.WithEndpoint(d.conf.Discovery.Contacts),
		grpc.WithDiscovery(d.discovery),
		grpc.WithTimeout(d.conf.Discovery.ContactsTimeout.AsDuration()),
		grpc.WithMiddleware(d.getJwtMiddleware(ctx)),
	)

	if err != nil {
		return nil, err
	}

	return contacts_v1.NewContactsClient(conn), nil
}

func (d *Dialer) Relations(ctx context.Context) (contacts_v1.RelationsClient, error) {
	conn, err := grpc.DialInsecure(
		ctx,
		grpc.WithEndpoint(d.conf.Discovery.Contacts),
		grpc.WithDiscovery(d.discovery),
		grpc.WithTimeout(d.conf.Discovery.ContactsTimeout.AsDuration()),
		grpc.WithMiddleware(d.getJwtMiddleware(ctx)),
	)

	if err != nil {
		return nil, err
	}

	return contacts_v1.NewRelationsClient(conn), nil
}

func (d *Dialer) Notifications(ctx context.Context) (notifications_v1.SenderClient, error) {
	conn, err := grpc.DialInsecure(
		ctx,
		grpc.WithEndpoint(d.conf.Discovery.Notifications),
		grpc.WithDiscovery(d.discovery),
		grpc.WithTimeout(d.conf.Discovery.NotificationsTimeout.AsDuration()),
		grpc.WithMiddleware(d.getJwtMiddleware(ctx)),
	)

	if err != nil {
		return nil, err
	}

	return notifications_v1.NewSenderClient(conn), nil
}

func (d *Dialer) Chats(ctx context.Context) (chats_v1.ChatsClient, error) {
	conn, err := grpc.DialInsecure(
		ctx,
		grpc.WithEndpoint(d.conf.Discovery.Chats),
		grpc.WithDiscovery(d.discovery),
		grpc.WithTimeout(d.conf.Discovery.ChatsTimeout.AsDuration()),
		grpc.WithMiddleware(d.getJwtMiddleware(ctx)),
	)
	if err != nil {
		return nil, err
	}
	return chats_v1.NewChatsClient(conn), nil
}

func (d *Dialer) Members(ctx context.Context) (chats_v1.MembersClient, error) {
	conn, err := grpc.DialInsecure(
		ctx,
		grpc.WithEndpoint(d.conf.Discovery.Chats),
		grpc.WithDiscovery(d.discovery),
		grpc.WithTimeout(d.conf.Discovery.ChatsTimeout.AsDuration()),
		grpc.WithMiddleware(d.getJwtMiddleware(ctx)),
	)
	if err != nil {
		return nil, err
	}
	return chats_v1.NewMembersClient(conn), nil
}

func FromChatsToIam(membership *chats_v1.Membership) *v1.CommonChat {
	if membership == nil {
		return nil
	}

	return &v1.CommonChat{
		ChatId:     membership.ChatId,
		Status:     membership.Status,
		Role:       membership.Role,
		IsPinned:   membership.IsPinned,
		IsMuted:    membership.IsMuted,
		MutedTill:  membership.MutedTill,
		ArchivedAt: membership.ArchivedAt,
		AutoSave:   membership.AutoSave,
	}
}

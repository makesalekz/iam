package data

import (
	"context"
	tenants_v1 "tenants/api/tenants/v1"

	"iam/internal/conf"
	notifications_v1 "notifications/api/notifications/v1"

	consul "github.com/go-kratos/consul/registry"
	"github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	jwtv4 "github.com/golang-jwt/jwt/v4"
)

type Dialer struct {
	conf      *conf.Bootstrap
	discovery *consul.Registry
	jwt       *JwtProcessor
}

// NewJwtProcessor .
func NewDialer(c *Config, jwt *JwtProcessor) (*Dialer, error) {
	return &Dialer{
		conf:      c.Bootstrap,
		discovery: c.GetRegistry(),
		jwt:       jwt,
	}, nil
}

func (d *Dialer) Notifications(ctx context.Context) (notifications_v1.SenderClient, error) {
	conn, err := grpc.DialInsecure(
		ctx,
		grpc.WithEndpoint(d.conf.Discovery.Notifications),
		grpc.WithDiscovery(d.discovery),
		grpc.WithTimeout(d.conf.Discovery.NotificationsTimeout.AsDuration()),
		grpc.WithMiddleware(
			jwt.Client(func(token *jwtv4.Token) (interface{}, error) {
				return d.jwt.GetSecret(), nil
			}, jwt.WithSigningMethod(jwtv4.SigningMethodHS256), jwt.WithClaims(func() jwtv4.Claims {
				return d.jwt.GetClaimsFromContext(ctx)
			})),
		),
	)
	if err != nil {
		return nil, err
	}
	return notifications_v1.NewSenderClient(conn), nil
}

func (d *Dialer) TenantsMembers(ctx context.Context) (tenants_v1.MembersClient, error) {
	conn, err := grpc.DialInsecure(
		ctx,
		grpc.WithEndpoint(d.conf.Discovery.Tenants),
		grpc.WithDiscovery(d.discovery),
		grpc.WithTimeout(d.conf.Discovery.TenantsTimeout.AsDuration()),
		grpc.WithMiddleware(
			jwt.Client(func(token *jwtv4.Token) (interface{}, error) {
				return d.jwt.GetSecret(), nil
			}, jwt.WithSigningMethod(jwtv4.SigningMethodHS256), jwt.WithClaims(func() jwtv4.Claims {
				return d.jwt.GetClaimsFromContext(ctx)
			})),
		),
	)
	if err != nil {
		return nil, err
	}
	return tenants_v1.NewMembersClient(conn), nil
}

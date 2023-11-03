package data

import (
	"context"

	"iam/internal/conf"
	contacts_v1 "iam/third_party/contacts/api/contacts/v1"

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

func (d *Dialer) Contacts(ctx context.Context) (contacts_v1.ContactsClient, error) {
	conn, err := grpc.DialInsecure(
		ctx,
		grpc.WithEndpoint(d.conf.Discovery.Contacts),
		grpc.WithDiscovery(d.discovery),
		grpc.WithTimeout(d.conf.Discovery.ContactsTimeout.AsDuration()),
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
	return contacts_v1.NewContactsClient(conn), nil
}

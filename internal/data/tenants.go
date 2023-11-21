package data

import (
	"context"

	consul "github.com/go-kratos/consul/registry"
	kjwt "github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	krpc "github.com/go-kratos/kratos/v2/transport/grpc"
	jwtv4 "github.com/golang-jwt/jwt/v4"
	"gitlab.calendaria.team/services/iam/internal/conf"
	tenants_v1 "gitlab.calendaria.team/services/tenants/api/tenants/v1"
	"gitlab.calendaria.team/services/utils/v1/config"
	"gitlab.calendaria.team/services/utils/v1/jwt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
)

type TenantsRemote struct {
	conf      *conf.Bootstrap
	discovery *consul.Registry
	jwt       *jwt.JwtProcessor

	conn *grpc.ClientConn
}

// NewTenantsRemote .
func NewTenantsRemote(c *config.Config, conf *conf.Bootstrap, jwt *jwt.JwtProcessor) (*TenantsRemote, error) {
	return &TenantsRemote{
		conf:      conf,
		discovery: c.GetRegistry(),
		jwt:       jwt,
	}, nil
}

func (r *TenantsRemote) connect(ctx context.Context, claims *jwt.TenantClaims) error {
	if r.conn != nil && r.conn.GetState() == connectivity.Ready {
		return nil
	}

	conn, err := krpc.DialInsecure(
		ctx,
		krpc.WithDiscovery(r.discovery),
		krpc.WithEndpoint(r.conf.Discovery.Tenants),
		krpc.WithTimeout(r.conf.Discovery.TenantsTimeout.AsDuration()),
		krpc.WithMiddleware(
			kjwt.Client(func(token *jwtv4.Token) (interface{}, error) {
				return r.jwt.GetSecret(), nil
			}, kjwt.WithSigningMethod(jwtv4.SigningMethodHS256), kjwt.WithClaims(func() jwtv4.Claims {
				return claims
			})),
		),
	)
	if err == nil {
		r.conn = conn
	}

	return nil
}

func (r *TenantsRemote) Tenants(ctx context.Context, claims *jwt.TenantClaims) (tenants_v1.TenantsClient, error) {
	err := r.connect(ctx, claims)
	if err != nil {
		return nil, err
	}

	return tenants_v1.NewTenantsClient(r.conn), nil
}

func (r *TenantsRemote) Members(ctx context.Context, claims *jwt.TenantClaims) (tenants_v1.MembersClient, error) {
	err := r.connect(ctx, claims)
	if err != nil {
		return nil, err
	}

	return tenants_v1.NewMembersClient(r.conn), nil
}

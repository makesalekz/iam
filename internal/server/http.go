package server

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	kjwt "github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	"github.com/go-kratos/kratos/v2/middleware/metadata"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/selector"
	khttp "github.com/go-kratos/kratos/v2/transport/http"
	jwtv4 "github.com/golang-jwt/jwt/v4"
	v1 "gitlab.calendaria.team/services/iam/api/iam/v1"
	"gitlab.calendaria.team/services/iam/internal/conf"
	"gitlab.calendaria.team/services/iam/internal/service"
	"gitlab.calendaria.team/services/utils/v1/jwt"
)

func NewWhiteListMatcher() selector.MatchFunc {
	whiteList := make(map[string]struct{})
	whiteList["/iam.v1.Auth/AuthByPhone"] = struct{}{}
	whiteList["/iam.v1.Auth/AuthByCode"] = struct{}{}
	whiteList["/iam.v1.Auth/TempAuthBySuperCode"] = struct{}{}
	return func(ctx context.Context, operation string) bool {
		if _, ok := whiteList[operation]; ok {
			return false
		}
		return true
	}
}

// NewHTTPServer new an HTTP server.
func NewHTTPServer(
	c *conf.Bootstrap,
	logger log.Logger,
	jwtp *jwt.JwtProcessor,
	auth *service.AuthService,
	users *service.UsersService,
	privacy *service.PrivacyService,
	settings *service.SettingsService,
) *khttp.Server {
	var opts = []khttp.ServerOption{
		khttp.Middleware(
			recovery.Recovery(),
			metadata.Server(),
			selector.Server(
				kjwt.Server(func(token *jwtv4.Token) (interface{}, error) {
					return jwtp.GetSecret(), nil
				}, kjwt.WithSigningMethod(jwtv4.SigningMethodHS256), kjwt.WithClaims(func() jwtv4.Claims { return &jwt.TenantClaims{} })),
			).
				Match(NewWhiteListMatcher()).
				Build(),
		),
	}
	if c.Server.Http.Network != "" {
		opts = append(opts, khttp.Network(c.Server.Http.Network))
	}
	if c.Server.Http.Addr != "" {
		opts = append(opts, khttp.Address(c.Server.Http.Addr))
	}
	if c.Server.Http.Timeout != nil {
		opts = append(opts, khttp.Timeout(c.Server.Http.Timeout.AsDuration()))
	}
	srv := khttp.NewServer(opts...)

	v1.RegisterAuthHTTPServer(srv, auth)
	v1.RegisterUsersHTTPServer(srv, users)
	v1.RegisterPrivacyHTTPServer(srv, privacy)
	v1.RegisterSettingsHTTPServer(srv, settings)

	return srv
}

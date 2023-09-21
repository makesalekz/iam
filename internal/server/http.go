package server

import (
	"context"

	auth_v1 "iam/api/auth/v1"
	privacy_v1 "iam/api/privacy/v1"
	users_v1 "iam/api/users/v1"
	"iam/internal/conf"
	"iam/internal/data"
	"iam/internal/service"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	"github.com/go-kratos/kratos/v2/middleware/metadata"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/selector"
	khttp "github.com/go-kratos/kratos/v2/transport/http"
	jwtv4 "github.com/golang-jwt/jwt/v4"
)

func NewWhiteListMatcher() selector.MatchFunc {
	whiteList := make(map[string]struct{})
	whiteList["/api.auth.v1.Auth/AuthByPhone"] = struct{}{}
	whiteList["/api.auth.v1.Auth/AuthByCode"] = struct{}{}
	whiteList["/api.auth.v1.Auth/TempAuthBySuperCode"] = struct{}{}
	return func(ctx context.Context, operation string) bool {
		if _, ok := whiteList[operation]; ok {
			return false
		}
		return true
	}
}

// NewHTTPServer new an HTTP server.
func NewHTTPServer(c *conf.Bootstrap, logger log.Logger, jwtp *data.JwtProcessor, auth *service.AuthService, users *service.UsersService, privacy *service.PrivacyService) *khttp.Server {
	var opts = []khttp.ServerOption{
		khttp.Middleware(
			recovery.Recovery(),
			metadata.Server(),
			selector.Server(
				jwt.Server(func(token *jwtv4.Token) (interface{}, error) {
					return jwtp.GetSecret(), nil
				}, jwt.WithSigningMethod(jwtv4.SigningMethodHS256), jwt.WithClaims(func() jwtv4.Claims { return &jwtv4.RegisteredClaims{} })),
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

	auth_v1.RegisterAuthHTTPServer(srv, auth)
	users_v1.RegisterUsersHTTPServer(srv, users)
	privacy_v1.RegisterPrivacyHTTPServer(srv, privacy)

	return srv
}

package server

import (
	prom "github.com/go-kratos/kratos/contrib/metrics/prometheus/v2"
	"github.com/go-kratos/kratos/v2/middleware/metadata"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	v1 "gitlab.calendaria.team/services/iam/api/iam/v1"
	"gitlab.calendaria.team/services/iam/internal/conf"
	"gitlab.calendaria.team/services/iam/internal/service"
	"gitlab.calendaria.team/services/utils/v1/jwt"
	u_auth "gitlab.calendaria.team/services/utils/v1/middlewares/auth"
	metrics "gitlab.calendaria.team/services/utils/v1/middlewares/metrics"
)

// NewGRPCServer new a gRPC server.
func NewGRPCServer(
	c *conf.Bootstrap,
	jwtp *jwt.JwtProcessor,
	auth *service.AuthService,
	users *service.UsersService,
	privacy *service.PrivacyService,
	settings *service.SettingsService,
) *grpc.Server {
	var opts = []grpc.ServerOption{
		grpc.Middleware(
			recovery.Recovery(),
			metadata.Server(),
			u_auth.Server(jwtp),
			metrics.Server(
				metrics.WithSeconds(prom.NewHistogram(_metricSeconds)),
				metrics.WithRequests(prom.NewCounter(_metricRequests)),
				metrics.WithGauge(prom.NewGauge(_activeRequests)),
			),
		),
	}
	if c.Server.Grpc.Network != "" {
		opts = append(opts, grpc.Network(c.Server.Grpc.Network))
	}
	if c.Server.Grpc.Addr != "" {
		opts = append(opts, grpc.Address(c.Server.Grpc.Addr))
	}
	if c.Server.Grpc.Timeout != nil {
		opts = append(opts, grpc.Timeout(c.Server.Grpc.Timeout.AsDuration()))
	}
	srv := grpc.NewServer(opts...)

	v1.RegisterAuthServer(srv, auth)
	v1.RegisterUsersServer(srv, users)
	v1.RegisterPrivacyServer(srv, privacy)
	v1.RegisterSettingsServer(srv, settings)

	return srv
}

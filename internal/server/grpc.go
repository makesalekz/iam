package server

import (
	v1 "github.com/makesalekz/iam/api/iam/v1"
	"github.com/makesalekz/iam/internal/conf"
	"github.com/makesalekz/iam/internal/service"
	u_metrics "github.com/makesalekz/utils/v1/middlewares/metrics"
	u_jwt "github.com/makesalekz/utils/v4/jwt"
	u_auth "github.com/makesalekz/utils/v4/middlewares/auth"
	u_tracing "github.com/makesalekz/utils/v4/tracing"

	prom "github.com/go-kratos/kratos/contrib/metrics/prometheus/v2"
	"github.com/go-kratos/kratos/v2/middleware/metadata"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport/grpc"
)

// NewGRPCServer new a gRPC server.
func NewGRPCServer(
	c *conf.Bootstrap,
	jwtp u_jwt.IJwtProcessor,
	tracer u_tracing.ITracer,
	auth *service.AuthService,
	users *service.UsersService,
	privacy *service.PrivacyService,
	settings *service.SettingsService,
	credentials *service.CredentialsService,
) *grpc.Server {
	err := tracer.Initialize()
	if err != nil {
		panic(err)
	}

	var opts = []grpc.ServerOption{
		grpc.Middleware(
			recovery.Recovery(),
			metadata.Server(),
			tracing.Server(),
			u_auth.Server(jwtp),
			u_metrics.Server(
				u_metrics.WithSeconds(prom.NewHistogram(_metricSeconds)),
				u_metrics.WithRequests(prom.NewCounter(_metricRequests)),
				u_metrics.WithGauge(prom.NewGauge(_activeRequests)),
			),
		),
	}
	if c.GetServer().GetGrpc().GetNetwork() != "" {
		opts = append(opts, grpc.Network(c.GetServer().GetGrpc().GetNetwork()))
	}
	if c.GetServer().GetGrpc().GetAddr() != "" {
		opts = append(opts, grpc.Address(c.GetServer().GetGrpc().GetAddr()))
	}
	if c.GetServer().GetGrpc().GetTimeout() != nil {
		opts = append(opts, grpc.Timeout(c.GetServer().GetGrpc().GetTimeout().AsDuration()))
	}
	srv := grpc.NewServer(opts...)

	v1.RegisterAuthServer(srv, auth)
	v1.RegisterUsersServer(srv, users)
	v1.RegisterPrivacyServer(srv, privacy)
	v1.RegisterSettingsServer(srv, settings)
	v1.RegisterCredentialsServer(srv, credentials)

	return srv
}

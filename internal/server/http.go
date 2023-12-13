package server

import (
	"context"

	prom "github.com/go-kratos/kratos/contrib/metrics/prometheus/v2"
	kjwt "github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	"github.com/go-kratos/kratos/v2/middleware/metadata"
	"github.com/go-kratos/kratos/v2/middleware/metrics"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/selector"
	khttp "github.com/go-kratos/kratos/v2/transport/http"
	jwtv4 "github.com/golang-jwt/jwt/v4"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gitlab.calendaria.team/services/iam/internal/conf"
	"gitlab.calendaria.team/services/utils/v1/jwt"
)

var _metricSeconds = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Namespace: "server",
	Subsystem: "requests",
	Name:      "duration_sec",
	Help:      "server requests duratio(sec).",
	Buckets:   []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.250, 0.5, 1},
}, []string{"kind", "operation"})

var _metricRequests = prometheus.NewCounterVec(prometheus.CounterOpts{
	Namespace: "client",
	Subsystem: "requests",
	Name:      "code_total",
	Help:      "The total number of processed requests",
}, []string{"kind", "operation", "code", "reason"})

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
	jwtp *jwt.JwtProcessor,
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
			metrics.Server(
				metrics.WithSeconds(prom.NewHistogram(_metricSeconds)),
				metrics.WithRequests(prom.NewCounter(_metricRequests)),
			),
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

	registerTechRoutes(srv)

	return srv
}

func registerTechRoutes(s *khttp.Server) {
	prometheus.MustRegister(_metricSeconds, _metricRequests)

	s.Handle("/metrics", promhttp.Handler())
}

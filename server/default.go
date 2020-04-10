package server

import (
	"context"
	"time"

	"github.com/pthethanh/micro/config/envconfig"
	"github.com/pthethanh/micro/log"
)

type (
	// Config is a common configuration of a default server.
	// Mostly used by lazy guys via NewFromEnv().
	Config struct {
		Name        string `envconfig:"NAME"`
		Address     string `envconfig:"ADDRESS" default:":8000"`
		TLSCertFile string `envconfig:"TLS_CERT_FILE"`
		TLSKeyFile  string `envconfig:"TLS_KEY_FILE"`

		// Paths
		LivenessPath  string `envconfig:"LIVENESS_PATH" default:"/internal/liveness"`
		ReadinessPath string `envconfig:"READINESS_PATH" default:"/internal/readiness"`
		MetricsPath   string `envconfig:"METRICS_PATH" default:"/internal/metrics"`

		// HTTP
		ReadTimeout  time.Duration `envconfig:"READ_TIMEOUT" default:"30s"`
		WriteTimeout time.Duration `envconfig:"WRITE_TIMEOUT" default:"30s"`

		JWTSecret     string `envconfig:"JWT_SECRET"`
		ContextLogger bool   `envconfig:"CONTEXT_LOGGER" default:"true"`
	}
)

// NewFromEnv load configurations from environment and create a new server.
// Additional options can be added to the sever via Server.WithOptions(...).
// See Config for environment names.
func NewFromEnv() *Server {
	conf := Config{}
	envconfig.Read(&conf)
	return newFromConfig(conf)
}

func newFromConfig(conf Config) *Server {
	opts := []Option{
		MetricsPaths(conf.ReadinessPath, conf.LivenessPath, conf.MetricsPath),
		TLS(conf.TLSKeyFile, conf.TLSCertFile),
		Timeout(conf.ReadTimeout, conf.WriteTimeout),
		JWTAuth(conf.JWTSecret),
	}
	if conf.ContextLogger {
		// allow request-id to be passed.
		opts = append(opts, ServeMuxOptions(DefaultHeaderMatcher()))
		opts = append(opts, Logger(log.WithFields(log.Fields{"service": conf.Name})))
	}
	return New(conf.Address, opts...)
}

// ListenAndServe create a new server base on environment configuration
// and serve the services with background context.
// See server.ListenAndServe for detail document.
func ListenAndServe(services ...Service) error {
	return ListenAndServeContext(context.Background(), services...)
}

// ListenAndServeContext create a new server base on environment configuration
// and serve the services with the given context.
// See server.ListenAndServeContext for detail document.
func ListenAndServeContext(ctx context.Context, services ...Service) error {
	return NewFromEnv().ListenAndServeContext(ctx, services...)
}

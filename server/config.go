package server

import (
	"time"

	"github.com/pthethanh/micro/config/env"
	"github.com/pthethanh/micro/health"

	"google.golang.org/grpc"
)

// Config holds the configuration options for the server instance.
type Config struct {
	Address     string `envconfig:"ADDRESS" default:"localhost:8000"`
	TLSCertFile string `envconfig:"TLS_CERT_FILE"`
	TLSKeyFile  string `envconfig:"TLS_KEY_FILE"`

	// Paths
	LivenessPath  string `envconfig:"LIVENESS_PATH" default:"/internal/liveness"`
	ReadinessPath string `envconfig:"READINESS_PATH" default:"/internal/readiness"`
	MetricsPath   string `envconfig:"METRICS_PATH" default:"/internal/metrics"`

	// HTTP
	ReadTimeout  time.Duration `envconfig:"READ_TIMEOUT" default:"30s"`
	WriteTimeout time.Duration `envconfig:"WRITE_TIMEOUT" default:"30s"`

	// Needs to be set manually
	Auth          Authenticator
	HealthChecks  []health.CheckFunc
	ServerOptions []grpc.ServerOption
}

// LoadConfigFromEnv returns a Config object populated
// from environment variables.
func LoadConfigFromEnv() *Config {
	var cfg Config
	env.Load(&cfg)
	return &cfg
}

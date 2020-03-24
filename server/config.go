package server

import (
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/pthethanh/micro/config/envconfig"
	"github.com/pthethanh/micro/health"

	"google.golang.org/grpc"
)

// Config holds the configuration options for the server instance.
type Config struct {
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

	// Logging
	EnableContextLogger bool `envconfig:"ENABLE_CONTEXT_LOGGER" default:"true"`

	// Needs to be set manually
	Auth            Authenticator
	HealthChecks    []health.CheckFunc
	ServerOptions   []grpc.ServerOption
	ServeMuxOptions []runtime.ServeMuxOption
}

// LoadConfigFromEnv returns a Config object populated
// from environment variables.
func LoadConfigFromEnv() *Config {
	var cfg Config
	_ = envconfig.Read(&cfg)
	return &cfg
}

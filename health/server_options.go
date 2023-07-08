package health

import (
	"time"

	"github.com/pthethanh/micro/config"
	"github.com/pthethanh/micro/config/envconfig"
	"github.com/pthethanh/micro/log"
)

// Interval is an option to set interval for health check.
func Interval(d time.Duration) ServerOption {
	return func(srv *MServer) {
		srv.conf.Interval = d
	}
}

// Timeout is an option to set timeout for each service health check.
func Timeout(d time.Duration) ServerOption {
	return func(srv *MServer) {
		srv.conf.Timeout = d
	}
}

// Logger is an option to set logger for the health check server.
func Logger(l log.Logger) ServerOption {
	return func(srv *MServer) {
		srv.log = l
	}
}

// FromEnv is an option to load config from environment variables.
func FromEnv(opts ...config.ReadOption) ServerOption {
	conf := Config{}
	envconfig.Read(&conf, opts...)
	return func(srv *MServer) {
		srv.conf = conf
	}
}

// FromConfig is an option to override server's config.
func FromConfig(conf Config) ServerOption {
	return func(srv *MServer) {
		srv.conf = conf
	}
}

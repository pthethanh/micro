package redis

import (
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/pthethanh/micro/config"
	"github.com/pthethanh/micro/config/envconfig"
)

type (
	// Config holds Redis cache configuration.
	Config struct {
		Addrs []string `envconfig:"REDIS_ADDRS"`

		// Database to be selected after connecting to the server.
		// Only single-node and failover clients.
		DB int `envconfig:"REDIS_DB"`

		Username string `envconfig:"REDIS_USERNAME"`
		Password string `envconfig:"REDIS_PASSWORD"`

		MaxRetries      int           `envconfig:"REDIS_MAX_RETRIES" default:"3"`
		MinRetryBackoff time.Duration `envconfig:"REDIS_MIN_RETRY_BACKOFF" default:"5s"`
		MaxRetryBackoff time.Duration `envconfig:"REDIS_MAX_RETRY_BACKOFF" default:"30s"`

		DialTimeout  time.Duration `envconfig:"REDIS_DIAL_TIMEOUT" default:"30s"`
		ReadTimeout  time.Duration `envconfig:"REDIS_READ_TIMEOUT" default:"30s"`
		WriteTimeout time.Duration `envconfig:"REDIS_WRITE_TIMEOUT" default:"30s"`

		PoolSize           int           `envconfig:"REDIS_POOL_SIZE"`
		MinIdleConns       int           `envconfig:"REDIS_MIN_IDLE_CONNS"`
		MaxConnAge         time.Duration `envconfig:"REDIS_MAX_CONN_AGE"`
		PoolTimeout        time.Duration `envconfig:"REDIS_POOL_TIMEOUT"`
		IdleTimeout        time.Duration `envconfig:"REDIS_IDLE_TIMEOUT"`
		IdleCheckFrequency time.Duration `envconfig:"REDIS_IDLE_CHECK_FREQUENCY"`

		// Only cluster clients.

		MaxRedirects   int  `envconfig:"REDIS_MAX_REDIRECTS"`
		ReadOnly       bool `envconfig:"REDIS_READ_ONLY"`
		RouteByLatency bool `envconfig:"REDIS_ROUTE_BY_LATENCY"`
		RouteRandomly  bool `envconfig:"REDIS_ROUTE_RANDOMLY"`

		// The sentinel master name.
		// Only failover clients.
		MasterName string `envconfig:"REDIS_MASTER_NAME"`
	}

	// Option is Redis configuration option.
	Option func(*Redis)
)

// ReadConfigFromEnv read Redis configuration from environment variables.
func ReadConfigFromEnv(opts ...config.ReadOption) Config {
	conf := Config{}
	envconfig.Read(&conf, opts...)
	return conf
}

// FromConfig is an option to configure the Redis cache from a custom config.
func FromConfig(conf Config) Option {
	return func(r *Redis) {
		r.opts.Addrs = conf.Addrs
		r.opts.DB = conf.DB
		r.opts.DialTimeout = conf.DialTimeout
		r.opts.IdleCheckFrequency = conf.IdleCheckFrequency
		r.opts.IdleTimeout = conf.IdleTimeout
		r.opts.MasterName = conf.MasterName
		r.opts.MaxConnAge = conf.MaxConnAge
		r.opts.MaxRedirects = conf.MaxRedirects
		r.opts.MaxRetries = conf.MaxRetries
		r.opts.MaxRetryBackoff = conf.MaxRetryBackoff
		r.opts.MinIdleConns = conf.MinIdleConns
		r.opts.MinRetryBackoff = conf.MinRetryBackoff
		r.opts.Password = conf.Password
		r.opts.PoolSize = conf.PoolSize
		r.opts.PoolTimeout = conf.PoolTimeout
		r.opts.RouteByLatency = conf.RouteByLatency
		r.opts.RouteRandomly = conf.RouteRandomly
		r.opts.Username = conf.Username
		r.opts.WriteTimeout = conf.WriteTimeout
	}
}

// FromEnv is an option to configure the Redis cache from environment variables.
func FromEnv(opts ...config.ReadOption) Option {
	conf := Config{}
	envconfig.Read(&conf, opts...)
	return FromConfig(conf)
}

// FromUniversalOptions is an option to configure Redis cache from the given options.
func FromUniversalOptions(opts *redis.UniversalOptions) Option {
	return func(r *Redis) {
		r.opts = opts
	}
}

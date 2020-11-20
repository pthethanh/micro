package nats

import (
	"time"

	"github.com/nats-io/nats.go"
	"github.com/pthethanh/micro/broker"
	"github.com/pthethanh/micro/config"
	"github.com/pthethanh/micro/config/envconfig"
	"github.com/pthethanh/micro/log"
)

type (
	// Config hold common NATS configurations.
	Config struct {
		Addrs    string        `envconfig:"NATS_ADDRS" default:"nats:4222"`
		Encoder  string        `envconfig:"NATS_ENCODER" default:"protobuf"`
		Timeout  time.Duration `envconfig:"NATS_TIMEOUT" default:"10s"`
		Username string        `envconfig:"NATS_USERNAME"`
		Password string        `envconfig:"NATS_PASSWORD"`
	}
)

const (
	defaultAddr = "nats:4222"
)

// ReadConfigFromEnv read NATS configuration from environment variables.
func ReadConfigFromEnv(opts ...config.ReadOption) Config {
	conf := Config{}
	envconfig.Read(&conf, opts...)
	return conf
}

// FromEnv is an option to create new broker base on environment variables.
func FromEnv(opts ...config.ReadOption) Option {
	var conf Config
	envconfig.Read(&conf, opts...)
	return FromConfig(conf)
}

// FromConfig is an option to create new broker from an existing config.
func FromConfig(conf Config) Option {
	return func(opts *Nats) {
		opts.addrs = conf.Addrs
		opts.opts = append(opts.opts, nats.Timeout(conf.Timeout))
		if conf.Username != "" {
			opts.opts = append(opts.opts, nats.UserInfo(conf.Username, conf.Password))
		}
		switch conf.Encoder {
		case "json":
			opts.encoder = broker.JSONEncoder{}
		case "protobuf":
			opts.encoder = broker.ProtoEncoder{}
		default:
			opts.getLogger().Warnf("nats: unrecognized encoder type: %s, switching back to default encoder: protobuf", conf.Encoder)
			opts.encoder = broker.ProtoEncoder{}
		}
	}
}

// Encoder is an option to provide a custom encoder.
func Encoder(encoder broker.Encoder) Option {
	return func(opts *Nats) {
		opts.encoder = encoder
	}
}

// Address is an option to set target addresses of NATS server.
// Multiple addresses are separated by comma.
func Address(addrs string) Option {
	return func(opts *Nats) {
		opts.addrs = addrs
	}
}

// Options is an option to provide additional nats.Option.
func Options(opts ...nats.Option) Option {
	return func(n *Nats) {
		n.opts = append(n.opts, opts...)
	}
}

// Logger is an option to provide custom logger.
func Logger(logger log.Logger) Option {
	return func(opts *Nats) {
		opts.log = logger
	}
}

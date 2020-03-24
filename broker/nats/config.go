package nats

import (
	"time"

	"github.com/nats-io/nats.go"
	"github.com/pthethanh/micro/broker"
	"github.com/pthethanh/micro/config/envconfig"
)

type (
	Config struct {
		Addrs    string        `envconfig:"NATS_ADDRS" default:"nats:4222"`
		Encoder  string        `envconfig:"NATS_ENCODER" default:"proto"`
		Timeout  time.Duration `envconfig:"NATS_TIMEOUT" default:"10s"`
		Username string        `envconfig:"NATS_USERNAME"`
		Password string        `envconfig:"NATS_PASSWORD"`
	}
)

func LoadConfigFromEnv() Config {
	var conf Config
	_ = envconfig.Read(&conf)
	return conf
}

func (conf Config) GetEncoder() broker.Encoder {
	switch conf.Encoder {
	case "json":
		return broker.JSONEncoder{}
	default:
		return broker.ProtoEncoder{}
	}
}

func (conf Config) Options() []nats.Option {
	opts := make([]nats.Option, 0)
	opts = append(opts, nats.Timeout(conf.Timeout))
	if conf.Username != "" {
		opts = append(opts, nats.UserInfo(conf.Username, conf.Password))
	}
	return opts
}

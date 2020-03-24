package mongodb

import (
	"context"
	"time"

	"github.com/globalsign/mgo"
	"github.com/pthethanh/micro/config/envconfig"
	"github.com/pthethanh/micro/health"
	"github.com/pthethanh/micro/log"
)

type (
	// Config hold MongoDB configuration information
	Config struct {
		Addrs    []string      `envconfig:"MONGODB_ADDRS" default:"mongodb:27017"`
		Database string        `envconfig:"MONGODB_DATABASE" default:"goway"`
		Username string        `envconfig:"MONGODB_USERNAME"`
		Password string        `envconfig:"MONGODB_PASSWORD"`
		Timeout  time.Duration `envconfig:"MONGODB_TIMEOUT" default:"10s"`
		Mode     mgo.Mode      `envconfig:"MONGODB_MODE" default:"1"`
		Refesh   bool          `envconfig:"MONGODB_REFRESH" default:"true"`
	}
)

// LoadConfigFromEnv returns a Config object populated from environment variables.
func LoadConfigFromEnv() *Config {
	var cfg Config
	_ = envconfig.Read(&cfg)
	return &cfg
}

// Dial dial to target server with Monotonic mode
func Dial(conf *Config) (*mgo.Session, error) {
	log.Infof("dialing to target MongoDB at: %v, database: %v", conf.Addrs, conf.Database)
	ms, err := mgo.DialWithInfo(&mgo.DialInfo{
		Addrs:    conf.Addrs,
		Database: conf.Database,
		Username: conf.Username,
		Password: conf.Password,
		Timeout:  conf.Timeout,
	})
	if err != nil {
		return nil, err
	}

	ms.SetMode(conf.Mode, conf.Refesh)
	log.Infof("successfully dialing to MongoDB at %v", conf.Addrs)
	return ms, nil
}

// DialInfo return dial info from config
func (conf *Config) DialInfo() *mgo.DialInfo {
	return &mgo.DialInfo{
		Addrs:    conf.Addrs,
		Database: conf.Database,
		Username: conf.Username,
		Password: conf.Password,
		Timeout:  conf.Timeout,
	}
}

// HealthCheck return a health check function.
func HealthCheck(s *mgo.Session) health.CheckFunc {
	return func(ctx context.Context) error {
		return s.Ping()
	}
}

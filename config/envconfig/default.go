package envconfig

import (
	"github.com/pthethanh/micro/config"
)

var (
	defaultConf = &Config{}
)

// Read read configuration from environment to the target ptr.
func Read(ptr interface{}, opts ...config.ReadOption) error {
	return defaultConf.Read(ptr, opts...)
}

// Close close the default config reader.
func Close() error {
	return defaultConf.Close()
}

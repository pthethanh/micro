package envconfig

import (
	"github.com/pthethanh/micro/config"
)

var (
	defaultConfig = &Config{}
)

// Read read configuration from environment to the target ptr.
func Read(ptr interface{}, opts ...config.ReadOption) error {
	return defaultConfig.Read(ptr, opts...)
}

// Close close the default config reader.
func Close() error {
	return defaultConfig.Close()
}

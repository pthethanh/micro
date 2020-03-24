package envconfig

import (
	"github.com/pthethanh/micro/config"
)

var (
	defaultConfig = &Config{}
)

func Read(ptr interface{}, opts ...config.ReadOption) error {
	return defaultConfig.Read(ptr, opts...)
}

func Close() error {
	return defaultConfig.Close()
}

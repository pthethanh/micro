package envconfig_test

import (
	"time"

	"github.com/pthethanh/micro/config"
	"github.com/pthethanh/micro/config/envconfig"
)

func Example() {
	var conf struct {
		Name        string            `envconfig:"NAME" default:"micro"`
		Address     string            `envconfig:"ADDRESS" default:"0.0.0.0:8000"`
		Secret      string            `envconfig:"SECRET"`
		Fields      []string          `envconfig:"FIELDS" default:"field1,field2"`
		ReadTimeout time.Duration     `envconfig:"READ_TIMEOUT" default:"30s"`
		Enable      bool              `envconfig:"ENABLE" default:"true"`
		Map         map[string]string `envconfig:"MAP" default:"key:value,key1:value1"`
	}
	envconfig.Read(&conf)
}

func ExampleRead_withOptions() {
	var conf struct {
		Name    string `envconfig:"NAME" default:"micro"`
		Address string `envconfig:"ADDRESS" default:"0.0.0.0:8000"`
		Secret  string `envconfig:"SECRET"`
	}
	envconfig.Read(&conf, config.WithPrefix("HTTP"))
}

package envconfig_test

import (
	"os"
	"testing"

	"github.com/pthethanh/micro/config"
	"github.com/pthethanh/micro/config/envconfig"
)

func TestReadConfig(t *testing.T) {
	type myConfig struct {
		Name      string `envconfig:"NAME" default:"micro"`
		Address   string `envconfig:"ADDRESS" default:"0.0.0.0:8000"`
		Instances int    `envconfig:"INSTANCE"`
		Secret    string `envconfig:"SECRET"`
	}
	defer envconfig.Close()
	cases := []struct {
		name    string
		prepare func()
		do      func() (myConfig, error)
		want    myConfig
	}{
		{
			name: "load from env",
			prepare: func() {
				if err := os.Setenv("SECRET", "123"); err != nil {
					t.Fatal(err)
				}
			},
			do: func() (myConfig, error) {
				conf := myConfig{}
				if err := envconfig.Read(&conf); err != nil {
					return conf, err
				}
				return conf, nil
			},
			want: myConfig{
				Name:      "micro",
				Address:   "0.0.0.0:8000",
				Instances: 0,
				Secret:    "123",
			},
		},
		{
			name:    "load from env and env file",
			prepare: func() {},
			do: func() (myConfig, error) {
				conf := myConfig{}
				if err := envconfig.Read(&conf, config.WithFile("testdata/micro.env")); err != nil {
					return conf, err
				}
				return conf, nil
			},
			want: myConfig{
				Name:      "micro",
				Address:   "1.1.1.1:8080",
				Instances: 0,
				Secret:    "",
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			os.Clearenv() // clear user env
			c.prepare()
			conf, err := c.do()
			if err != nil {
				t.Errorf("read failed: %v", err)
			}
			if conf.Name != c.want.Name {
				t.Errorf("got name=%s, want name=%s", conf.Name, c.want.Name)
			}
			if conf.Address != c.want.Address {
				t.Errorf("got address=%s, want address=%s", conf.Address, c.want.Address)
			}
			if conf.Instances != c.want.Instances {
				t.Errorf("got instance=%d, want instance=%d", conf.Instances, c.want.Instances)
			}
			if conf.Secret != c.want.Secret {
				t.Errorf("got secret=%s, want secret=%s", conf.Secret, c.want.Secret)
			}
		})
	}
}

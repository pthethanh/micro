package envconfig

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/kelseyhightower/envconfig"
	"github.com/pthethanh/micro/config"
)

type (
	// Config implements config.Reader.
	Config struct {
	}
)

// Read implements config.Reader interface.
func (env *Config) Read(ptr interface{}, opts ...config.ReadOption) error {
	ops := &config.ReadOptions{}
	ops.Apply(opts...)
	if ops.File != "" {
		if err := loadEnvFromFile(ops.File); err != nil {
			return err
		}
	}
	return envconfig.Process(ops.Prefix, ptr)
}

// Close implements config.Reader interface.
func (env *Config) Close() error {
	return nil
}

// loadEnvFromFile load environments from file
// and set them to system environment via os.Setenv.
func loadEnvFromFile(f string) error {
	file, err := os.Open(f)
	if err != nil {
		return err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		txt := scanner.Text()
		if strings.HasPrefix(txt, "#") || strings.TrimSpace(txt) == "" {
			continue
		}
		env := strings.SplitN(txt, "=", 2)
		if len(env) != 2 {
			return fmt.Errorf("invalid pair: %v", txt)
		}
		k := env[0]
		v := env[1]
		_ = os.Setenv(k, v)
	}
	return nil
}

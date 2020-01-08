package env

import (
	"bufio"
	"errors"
	"log"
	"os"
	"strings"

	"github.com/kelseyhightower/envconfig"
)

var (
	// ErrInvalidConfigFile report an invalid configuration file
	ErrInvalidConfigFile = errors.New("config: environment must be in key=value pair")

	// Prefix used as a prefix for environment variables,
	// defaults to empty.
	Prefix = ""
)

// Load loads the environment variables into the provided
// struct, using envconfig.
func Load(c interface{}) {
	if err := envconfig.Process(Prefix, c); err != nil {
		log.Fatalf("config: unable to load config for %T: %v", c, err)
	}
}

// LoadWithPrefix loads the environment variables with
// prefix into the provided struct.
func LoadWithPrefix(prefix string, c interface{}) {
	if err := envconfig.Process(prefix, c); err != nil {
		log.Fatalf("config: unable to load config for %T: %v", c, err)
	}
}

// LoadFromFile the environment variables in key=value pairs
// from file into the provided struct.
func LoadFromFile(f string, c interface{}) {
	if err := setEnvFromFile(f); err != nil {
		log.Fatalf("config: unable to load config from env file: %v", err)
	}
	Load(c)
}

// setEnvFromFile load environments from file
// and set them to system environment via os.Setenv.
func setEnvFromFile(f string) error {
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
			return ErrInvalidConfigFile
		}
		k := env[0]
		v := env[1]
		os.Setenv(k, v)
	}
	return nil
}

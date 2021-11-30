package app

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/viper"
)

// Config is our application wide configuration struct.
type Config struct {
	Server struct {
		WriteTimeout    time.Duration `mapstructure:"write_timeout"`
		ReadTimeout     time.Duration `mapstructure:"read_timeout"`
		IdleTimeout     time.Duration `mapstructure:"idle_timeout"`
		ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
		Address         string
	}
	DatabaseURL string `mapstructure:"DATABASE_URL"`
}

func createConfig() (*Config, error) {
	v := viper.New()

	// Allow viper to load environment variables automatically.
	v.AutomaticEnv()

	// Yaml config file for application config that does not change often.
	v.SetConfigFile("config.yaml")

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("unable to read config.yaml: %w", err)
	}

	// .env should be the developers local env config.
	v.SetConfigFile(".env")

	if err := v.MergeInConfig(); err != nil {
		return nil, fmt.Errorf("unable to read .env file: %w", err)
	}

	// If we are in test mode then merge in the test versions of the above to allow overrides for tests.
	currentEnv := os.Getenv("APP_ENV")

	if currentEnv == "test" {
		v.SetConfigFile("config_test.yaml")

		if err := v.MergeInConfig(); err != nil {
			return nil, fmt.Errorf("unable to read config_test.yaml: %w", err)
		}

		v.SetConfigFile("test.env")

		if err := v.MergeInConfig(); err != nil {
			return nil, fmt.Errorf("unable to read test.env file: %w", err)
		}
	}

	var c *Config
	if err := v.Unmarshal(&c); err != nil {
		return nil, fmt.Errorf("unable to unmarshal config: %w", err)
	}

	return c, nil
}

package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
)

const (
	ModeDevelopment = "development"
	ModeProduction  = "production"
)

type Config struct {
	Mode string `env:"SERVER_MODE" envDefault:"development"`

	Host string `env:"SERVER_HOST" envDefault:"127.0.0.1"`
	Port string `env:"SERVER_PORT" envDefault:"8000"`

	// ReadHeaderTimeout is the maximum duration in seconds before timing out
	// reading the headers of the request.
	ReadHeaderTimeout int `env:"SERVER_READ_HEADER_TIMEOUT" envDefault:"1"`

	// DSN is the data source name for the database connection.
	DSN string `env:"SERVER_DSN,required"`
}

func New() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	if err := validate(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func validate(cfg *Config) error {
	switch cfg.Mode {
	case ModeDevelopment:
	case ModeProduction:
	default:
		return fmt.Errorf("invalid mode: %s", cfg.Mode)
	}

	return nil
}

package config

import (
	"fmt"

	"github.com/caarlos0/env/v6"
)

// Config holds all application configuration
type Config struct {
	Server struct {
		Port string `env:"PORT" envDefault:"8081"`
	}
	HubSpot struct {
		APIKey string `env:"HS_API_KEY,required"`
	}
}

// Load loads configuration from .env file and environment variables
func Load() (*Config, error) {
	cfg := &Config{}

	// Parse environment variables into the config struct
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("failed to parse environment variables: %w", err)
	}

	return cfg, nil
}

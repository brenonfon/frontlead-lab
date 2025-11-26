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
	Salesforce struct {
		URL            string `env:"SF_URL,required"`
		ConsumerKey    string `env:"SF_CONSUMER_KEY,required"`
		ConsumerSecret string `env:"SF_CONSUMER_SECRET,required"`
		RedirectURI    string `env:"SF_REDIRECT_URI,required"`
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

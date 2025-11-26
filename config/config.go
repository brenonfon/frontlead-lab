package config

import (
	"fmt"
	"log"

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
	log.Println("[Config] Parsing environment variables...")
	cfg := &Config{}

	// Parse environment variables into the config struct
	if err := env.Parse(cfg); err != nil {
		log.Printf("[Config] Error parsing environment variables: %v", err)
		return nil, fmt.Errorf("failed to parse environment variables: %w", err)
	}

	// Validate configuration
	if cfg.HubSpot.APIKey == "" {
		log.Println("[Config] Error: HubSpot API key is missing")
		return nil, fmt.Errorf("HubSpot API key is required (HS_API_KEY)")
	}

	log.Println("[Config] Environment variables parsed successfully")
	return cfg, nil
}

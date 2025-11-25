package main

import (
	"log"

	"frontlead-lab/config"
	"frontlead-lab/hubspot"
	"frontlead-lab/server"

	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: Error loading .env file:", err)
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Println("✅ Successfully loaded configuration")

	// Initialize HubSpot client
	hsClient := hubspot.NewClient(cfg.HubSpot.APIKey)
	log.Println("✅ HubSpot client initialized")

	// Create and start server
	srv := server.New(hsClient, cfg.Server.Port)
	if err := srv.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

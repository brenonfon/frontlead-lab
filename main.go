package main

import (
	"log"

	"frontlead-lab/config"
	"frontlead-lab/salesforce"
	"frontlead-lab/server"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Println("✅ Successfully loaded configuration", cfg)
	log.Println("Authenticating with Salesforce...")

	// Authenticate with Salesforce
	sfClient, err := salesforce.Authenticate(
		cfg.Salesforce.URL,
		cfg.Salesforce.Username,
		cfg.Salesforce.Password,
		cfg.Salesforce.ConsumerKey,
		cfg.Salesforce.ConsumerSecret,
		cfg.Salesforce.SecurityToken,
	)
	if err != nil {
		log.Fatalf("Failed to authenticate with Salesforce: %v", err)
	}

	log.Println("✅ Successfully authenticated with Salesforce")

	// Create and start server
	srv := server.New(sfClient, cfg.Server.Port)
	if err := srv.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

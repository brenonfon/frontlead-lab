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

	log.Println("✅ Successfully loaded configuration")

	// Setup OAuth config
	oauthConfig := &salesforce.OAuthConfig{
		BaseURL:        cfg.Salesforce.URL,
		ConsumerKey:    cfg.Salesforce.ConsumerKey,
		ConsumerSecret: cfg.Salesforce.ConsumerSecret,
		RedirectURI:    cfg.Salesforce.RedirectURI,
	}

	// Start OAuth authentication flow
	log.Println("🔐 Starting OAuth authentication flow...")
	if err := server.StartAuthServer(oauthConfig, cfg.Server.Port); err != nil {
		log.Fatalf("Failed to authenticate with Salesforce: %v", err)
	}

	// Get the authenticated client
	sfClient := server.GetAuthenticatedClient()
	if sfClient == nil {
		log.Fatalf("Failed to get authenticated Salesforce client")
	}

	log.Println("✅ Successfully authenticated with Salesforce")

	// Create and start main server
	srv := server.New(sfClient, cfg.Server.Port)
	if err := srv.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

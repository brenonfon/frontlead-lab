package main

import (
	"log"

	"frontlead-lab/callapi"
	"frontlead-lab/config"
	"frontlead-lab/hubspot"
	"frontlead-lab/server"

	"github.com/joho/godotenv"
)

func main() {
	log.Println("==========================================")
	log.Println("   Starting FrontLead Lab Application    ")
	log.Println("==========================================")

	// Load .env file
	log.Println("[Main] Loading environment variables...")
	if err := godotenv.Load(); err != nil {
		log.Println("[Main] Warning: Error loading .env file:", err)
		log.Println("[Main] Continuing with environment variables from system...")
	} else {
		log.Println("[Main] ✅ Environment file loaded successfully")
	}

	// Load configuration
	log.Println("[Main] Loading application configuration...")
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("[Main] ❌ Failed to load configuration: %v", err)
	}

	log.Printf("[Main] ✅ Configuration loaded successfully")
	log.Printf("[Main]    - Server Port: %s", cfg.Server.Port)
	log.Printf("[Main]    - HubSpot API Key: %s", maskAPIKey(cfg.HubSpot.APIKey))

	// Initialize HubSpot client
	log.Println("[Main] Initializing HubSpot client...")
	hsClient := hubspot.NewClient(cfg.HubSpot.APIKey)
	log.Println("[Main] ✅ HubSpot client initialized")

	// Initialize Call API client
	log.Println("[Main] Initializing Call API client...")
	callAPIClient := callapi.NewClient(cfg.CallAPI.APIKey, cfg.CallAPI.Bot, cfg.CallAPI.Agents)
	log.Printf("[Main] ✅ Call API client initialized with API key: %s", maskAPIKey(cfg.CallAPI.APIKey))

	// Create and start server
	log.Println("[Main] Creating HTTP server...")
	srv := server.New(hsClient, callAPIClient, cfg.Server.Port)

	log.Println("[Main] Starting server...")
	if err := srv.Start(); err != nil {
		log.Fatalf("[Main] ❌ Failed to start server: %v", err)
	}
}

// maskAPIKey masks the API key for logging purposes
func maskAPIKey(key string) string {
	if len(key) <= 8 {
		return "****"
	}
	return key[:4] + "..." + key[len(key)-4:]
}

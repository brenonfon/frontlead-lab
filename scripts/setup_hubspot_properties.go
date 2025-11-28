package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
)

const hubspotPropertiesURL = "https://api.hubapi.com/crm/v3/properties/contacts"

type PropertyRequest struct {
	Name        string `json:"name"`
	Label       string `json:"label"`
	Type        string `json:"type"`
	FieldType   string `json:"fieldType"`
	GroupName   string `json:"groupName"`
	Description string `json:"description,omitempty"`
}

type PropertyResponse struct {
	Name      string `json:"name"`
	Label     string `json:"label"`
	Type      string `json:"type"`
	FieldType string `json:"fieldType"`
}

var properties = []PropertyRequest{
	{
		Name:        "business_needs",
		Label:       "Business Needs",
		Type:        "string",
		FieldType:   "textarea",
		GroupName:   "contactinformation",
		Description: "Customer's business needs and requirements",
	},
	{
		Name:        "integration_timeline",
		Label:       "Integration Timeline",
		Type:        "string",
		FieldType:   "text",
		GroupName:   "contactinformation",
		Description: "Expected timeline for integration",
	},
	{
		Name:        "seats_or_extensions",
		Label:       "Seats or Extensions",
		Type:        "string",
		FieldType:   "text",
		GroupName:   "contactinformation",
		Description: "Number of seats or extensions needed",
	},
	{
		Name:        "last_offer_summary",
		Label:       "Last Offer Summary",
		Type:        "string",
		FieldType:   "textarea",
		GroupName:   "contactinformation",
		Description: "Summary of the last offer provided to the contact",
	},
	{
		Name:        "interest_topic",
		Label:       "Interest Topic",
		Type:        "string",
		FieldType:   "text",
		GroupName:   "contactinformation",
		Description: "Topic of interest for the contact",
	},
	{
		Name:        "interest_source",
		Label:       "Interest Source",
		Type:        "string",
		FieldType:   "text",
		GroupName:   "contactinformation",
		Description: "Source of the contact's interest",
	},
	{
		Name:        "campaign_offer",
		Label:       "Campaign Offer",
		Type:        "string",
		FieldType:   "text",
		GroupName:   "contactinformation",
		Description: "Campaign offer associated with the contact",
	},
}

func createProperty(apiKey string, prop PropertyRequest) error {
	payloadBytes, err := json.Marshal(prop)
	if err != nil {
		return fmt.Errorf("failed to marshal property: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, hubspotPropertiesURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusConflict {
		log.Printf("⚠️  Property '%s' already exists, skipping...", prop.Name)
		return nil
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errorBody bytes.Buffer
		errorBody.ReadFrom(resp.Body)
		return fmt.Errorf("API returned status %d: %s", resp.StatusCode, errorBody.String())
	}

	var response PropertyResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	log.Printf("✅ Created property: %s (%s)", response.Label, response.Name)
	return nil
}

func main() {
	log.Println("🚀 HubSpot Property Setup Script")
	log.Println("================================")

	// Load environment variables
	if err := godotenv.Load("config/.env"); err != nil {
		log.Printf("Warning: Error loading config/.env file: %v", err)
		log.Println("Attempting to use system environment variables...")
	}

	apiKey := os.Getenv("HS_API_KEY")
	if apiKey == "" {
		log.Fatal("❌ Error: HS_API_KEY environment variable is not set")
	}

	log.Printf("📝 Found API Key: %s...%s", apiKey[:4], apiKey[len(apiKey)-4:])
	log.Printf("📋 Creating %d custom properties in HubSpot...\n", len(properties))

	successCount := 0
	failCount := 0

	for i, prop := range properties {
		log.Printf("[%d/%d] Creating property: %s", i+1, len(properties), prop.Name)

		if err := createProperty(apiKey, prop); err != nil {
			log.Printf("❌ Failed to create property '%s': %v", prop.Name, err)
			failCount++
		} else {
			successCount++
		}

		// Small delay to avoid rate limiting
		if i < len(properties)-1 {
			time.Sleep(500 * time.Millisecond)
		}
	}

	log.Println("\n================================")
	log.Printf("✅ Successfully created: %d properties", successCount)
	if failCount > 0 {
		log.Printf("❌ Failed to create: %d properties", failCount)
	}
	log.Println("🎉 Setup complete!")
}

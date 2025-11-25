package hubspot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

// Client represents a HubSpot API client
type Client struct {
	APIKey     string
	HTTPClient *http.Client
}

// Contact represents a HubSpot Contact object
type Contact struct {
	ID         string                 `json:"id,omitempty"`
	Properties map[string]interface{} `json:"properties"`
}

// ContactSearchResponse represents the response from searching contacts
type ContactSearchResponse struct {
	Results []struct {
		ID         string    `json:"id"`
		Properties map[string]string `json:"properties"`
		CreatedAt  string    `json:"createdAt"`
		UpdatedAt  string    `json:"updatedAt"`
		Archived   bool      `json:"archived"`
	} `json:"results"`
	Total int `json:"total"`
	Paging *struct {
		Next *struct {
			After string `json:"after"`
			Link  string `json:"link"`
		} `json:"next,omitempty"`
	} `json:"paging,omitempty"`
}

// NewClient creates a new HubSpot API client
func NewClient(apiKey string) *Client {
	return &Client{
		APIKey: apiKey,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// IsExistingContact checks if a contact exists by email
func (c *Client) IsExistingContact(email string) (bool, error) {
	searchPayload := map[string]interface{}{
		"filterGroups": []map[string]interface{}{
			{
				"filters": []map[string]interface{}{
					{
						"propertyName": "email",
						"operator":     "EQ",
						"value":        email,
					},
				},
			},
		},
		"properties": []string{"email", "phone", "lifecyclestage"},
		"limit":      1,
	}

	payloadBytes, err := json.Marshal(searchPayload)
	if err != nil {
		return false, err
	}

	req, err := http.NewRequest("POST", "https://api.hubapi.com/crm/v3/objects/contacts/search", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return false, err
	}

	req.Header.Add("Authorization", "Bearer "+c.APIKey)
	req.Header.Add("Content-Type", "application/json")

	// Make the request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("HubSpot API returned status code: %d", resp.StatusCode)
	}

	// Parse HubSpot response
	var hubspotResponse ContactSearchResponse

	if err := json.NewDecoder(resp.Body).Decode(&hubspotResponse); err != nil {
		return false, err
	}

	return hubspotResponse.Total > 0, nil
}

// IsExistingContactByPhone checks if a contact exists by phone number
func (c *Client) IsExistingContactByPhone(phoneNumber string) (bool, error) {
	searchPayload := map[string]interface{}{
		"filterGroups": []map[string]interface{}{
			{
				"filters": []map[string]interface{}{
					{
						"propertyName": "phone",
						"operator":     "EQ",
						"value":        phoneNumber,
					},
				},
			},
		},
		"properties": []string{"phone", "lifecyclestage"},
		"limit":      1,
	}

	payloadBytes, err := json.Marshal(searchPayload)
	if err != nil {
		return false, err
	}

	req, err := http.NewRequest("POST", "https://api.hubapi.com/crm/v3/objects/contacts/search", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return false, err
	}

	req.Header.Add("Authorization", "Bearer "+c.APIKey)
	req.Header.Add("Content-Type", "application/json")

	// Make the request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("HubSpot API returned status code: %d", resp.StatusCode)
	}

	// Parse HubSpot response
	var hubspotResponse ContactSearchResponse

	if err := json.NewDecoder(resp.Body).Decode(&hubspotResponse); err != nil {
		return false, err
	}

	return hubspotResponse.Total > 0, nil
}

// Helper function to maintain backward compatibility
func isExistingLeadInHubspot(email string) (bool, error) {
	apiKey := os.Getenv("HS_API_KEY")
	client := NewClient(apiKey)
	return client.IsExistingContact(email)
}

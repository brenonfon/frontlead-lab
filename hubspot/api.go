package hubspot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
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

// ContactInfo represents simplified contact information
type ContactInfo struct {
	Exists         bool   `json:"exists"`
	Email          string `json:"email,omitempty"`
	Phone          string `json:"phone,omitempty"`
	FirstName      string `json:"firstname,omitempty"`
	LastName       string `json:"lastname,omitempty"`
	Company        string `json:"company,omitempty"`
	LifecycleStage string `json:"lifecyclestage,omitempty"`
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
func (c *Client) IsExistingContact(email string) (*ContactInfo, error) {
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
		"properties": []string{"email", "phone", "firstname", "lastname", "company", "lifecyclestage"},
		"limit":      1,
	}

	payloadBytes, err := json.Marshal(searchPayload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", "https://api.hubapi.com/crm/v3/objects/contacts/search", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+c.APIKey)
	req.Header.Add("Content-Type", "application/json")

	// Make the request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HubSpot API returned status code: %d", resp.StatusCode)
	}

	// Parse HubSpot response
	var hubspotResponse ContactSearchResponse

	if err := json.NewDecoder(resp.Body).Decode(&hubspotResponse); err != nil {
		return nil, err
	}

	if hubspotResponse.Total == 0 {
		return &ContactInfo{Exists: false}, nil
	}

	// Extract contact information
	contact := hubspotResponse.Results[0]
	return &ContactInfo{
		Exists:         true,
		Email:          contact.Properties["email"],
		Phone:          contact.Properties["phone"],
		FirstName:      contact.Properties["firstname"],
		LastName:       contact.Properties["lastname"],
		Company:        contact.Properties["company"],
		LifecycleStage: contact.Properties["lifecyclestage"],
	}, nil
}

// IsExistingContactByPhone checks if a contact exists by phone number
func (c *Client) IsExistingContactByPhone(phoneNumber string) (*ContactInfo, error) {
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
		"properties": []string{"phone", "firstname", "lastname", "company", "lifecyclestage"},
		"limit":      1,
	}

	payloadBytes, err := json.Marshal(searchPayload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", "https://api.hubapi.com/crm/v3/objects/contacts/search", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+c.APIKey)
	req.Header.Add("Content-Type", "application/json")

	// Make the request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HubSpot API returned status code: %d", resp.StatusCode)
	}

	// Parse HubSpot response
	var hubspotResponse ContactSearchResponse

	if err := json.NewDecoder(resp.Body).Decode(&hubspotResponse); err != nil {
		return nil, err
	}

	if hubspotResponse.Total == 0 {
		return &ContactInfo{Exists: false}, nil
	}

	// Extract contact information
	contact := hubspotResponse.Results[0]
	return &ContactInfo{
		Exists:         true,
		Phone:          contact.Properties["phone"],
		FirstName:      contact.Properties["firstname"],
		LastName:       contact.Properties["lastname"],
		Company:        contact.Properties["company"],
		LifecycleStage: contact.Properties["lifecyclestage"],
	}, nil
}
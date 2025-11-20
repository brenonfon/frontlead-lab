package salesforce

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// Client represents a Salesforce API client
type Client struct {
	InstanceURL string
	AccessToken string
	APIVersion  string
	HTTPClient  *http.Client
}

// Lead represents a Salesforce Lead object
type Lead struct {
	ID          string `json:"Id,omitempty"`
	FirstName   string `json:"FirstName,omitempty"`
	LastName    string `json:"LastName"`
	Company     string `json:"Company"`
	Email       string `json:"Email,omitempty"`
	Phone       string `json:"Phone,omitempty"`
	Status      string `json:"Status,omitempty"`
	LeadSource  string `json:"LeadSource,omitempty"`
	Title       string `json:"Title,omitempty"`
	Website     string `json:"Website,omitempty"`
	Description string `json:"Description,omitempty"`
}

// CreateLeadResponse represents the response from creating a lead
type CreateLeadResponse struct {
	ID      string `json:"id"`
	Success bool   `json:"success"`
	Errors  []struct {
		Message    string   `json:"message"`
		StatusCode string   `json:"statusCode"`
		Fields     []string `json:"fields"`
	} `json:"errors"`
}

// AuthResponse represents the OAuth token response
type AuthResponse struct {
	AccessToken string `json:"access_token"`
}

// NewClient creates a new Salesforce API client
func NewClient(instanceURL, accessToken string) *Client {
	return &Client{
		InstanceURL: instanceURL,
		AccessToken: accessToken,
		APIVersion:  "v59.0",
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Authenticate authenticates with Salesforce using OAuth 2.0 Username-Password flow
func Authenticate(url, username, password, consumerKey, consumerSecret, securityToken string) (*Client, error) {
	authURL := url + "/services/oauth2/token"

	data := fmt.Sprintf(
		"grant_type=password&client_id=%s&client_secret=%s&username=%s&password=%s%s",
		consumerKey,
		consumerSecret,
		username,
		password,
		securityToken,
	)
	log.Println("Authenticating with Salesforce using the following data:")
	log.Println(data)

	req, err := http.NewRequest("POST", authURL, bytes.NewBufferString(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create auth request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("authentication request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("authentication failed with status %d: %s", resp.StatusCode, string(body))
	}

	var authResp AuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return nil, fmt.Errorf("failed to decode auth response: %w", err)
	}

	return NewClient(url, authResp.AccessToken), nil
}

// CreateLead creates a new lead in Salesforce
func (c *Client) CreateLead(lead *Lead) (*CreateLeadResponse, error) {
	url := fmt.Sprintf("%s/services/data/%s/sobjects/Lead/", c.InstanceURL, c.APIVersion)

	jsonData, err := json.Marshal(lead)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal lead: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("create lead failed with status %d: %s", resp.StatusCode, string(body))
	}

	var createResp CreateLeadResponse
	if err := json.Unmarshal(body, &createResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &createResp, nil
}

// GetLead retrieves a lead by ID from Salesforce
func (c *Client) GetLead(leadID string) (*Lead, error) {
	url := fmt.Sprintf("%s/services/data/%s/sobjects/Lead/%s", c.InstanceURL, c.APIVersion, leadID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.AccessToken)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get lead failed with status %d: %s", resp.StatusCode, string(body))
	}

	var lead Lead
	if err := json.Unmarshal(body, &lead); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &lead, nil
}

// UpdateLead updates an existing lead in Salesforce
func (c *Client) UpdateLead(leadID string, lead *Lead) error {
	url := fmt.Sprintf("%s/services/data/%s/sobjects/Lead/%s", c.InstanceURL, c.APIVersion, leadID)

	jsonData, err := json.Marshal(lead)
	if err != nil {
		return fmt.Errorf("failed to marshal lead: %w", err)
	}

	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("update lead failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

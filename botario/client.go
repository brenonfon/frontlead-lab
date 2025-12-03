package botario

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

const (
	botarioBaseURL = "https://gpt.nfon.botario.com/api/bots"
)

// Client represents a Botario API client
type Client struct {
	APIKey     string
	HTTPClient *http.Client
}

// NewClient creates a new Botario API client
func NewClient(apiKey string) *Client {
	log.Println("[Botario] Initializing Botario client...")
	return &Client{
		APIKey: apiKey,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// StartCallRequest represents the request to start a call
type StartCallRequest struct {
	Phone               string            `json:"phone"`
	Name                string            `json:"name,omitempty"`
	LeadID              string            `json:"leadId,omitempty"`
	Email               string            `json:"email,omitempty"`
	Company             string            `json:"company,omitempty"`
	BusinessNeeds       string            `json:"businessNeeds,omitempty"`
	IntegrationTimeline string            `json:"integrationTimeline,omitempty"`
	CustomData          map[string]string `json:"customData,omitempty"`
}

// StartCallResponse represents the response from Botario
type StartCallResponse struct {
	Success bool   `json:"success"`
	CallID  string `json:"callId,omitempty"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

// APIError represents an error from the Botario API
type APIError struct {
	StatusCode int
	Message    string
}

// Error implements the error interface
func (e *APIError) Error() string {
	return e.Message
}

// doRequest makes an HTTP request to the Botario API
func (c *Client) doRequest(method, endpoint string, payload interface{}, response interface{}) error {
	url := botarioBaseURL + endpoint

	var body []byte
	var err error
	if payload != nil {
		body, err = json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to marshal payload: %w", err)
		}
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	if c.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
	}

	log.Printf("[Botario] %s request to %s", method, url)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check for errors
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errorBody bytes.Buffer
		errorBody.ReadFrom(resp.Body)
		errorMsg := fmt.Sprintf("Botario API error (status %d): %s", resp.StatusCode, errorBody.String())
		log.Printf("[Botario] %s", errorMsg)
		return &APIError{
			StatusCode: resp.StatusCode,
			Message:    errorMsg,
		}
	}

	// Decode response
	if response != nil {
		if err := json.NewDecoder(resp.Body).Decode(response); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

// StartCall initiates a call via Botario
func (c *Client) StartCall(req StartCallRequest) (*StartCallResponse, error) {
	log.Printf("[Botario] Starting call to %s (Lead: %s)", req.Phone, req.LeadID)

	var response StartCallResponse
	err := c.doRequest(http.MethodPost, "/68ff27a71e00c0ee44698e5b/chats/send-message", req, &response)
	if err != nil {
		return nil, err
	}

	log.Printf("[Botario] Call started successfully - CallID: %s", response.CallID)
	return &response, nil
}

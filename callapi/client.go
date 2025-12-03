package callapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

const (
	callAPIBaseURL = "https://providersupportdata-test.cloud-cfg.com/v1"
)

// Client represents a Call API client
type Client struct {
	APIKey     string
	HTTPClient *http.Client
}

// NewClient creates a new Call API client
func NewClient(apiKey string) *Client {
	log.Println("[CallAPI] Initializing Call API client...")
	return &Client{
		APIKey: apiKey,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// MakeCallRequest represents the request to initiate a call
type MakeCallRequest struct {
	Extension     string `json:"extension,omitempty"`
	Caller        string `json:"caller"`
	CallerContext string `json:"caller_context"`
	Callee        string `json:"callee"`
	CalleeContext string `json:"callee_context"`
}

// MakeCallResponse represents the response from Call API
type MakeCallResponse struct {
	Success bool   `json:"success"`
	CallID  string `json:"callId,omitempty"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

// APIError represents an error from the Call API
type APIError struct {
	StatusCode int
	Message    string
}

// Error implements the error interface
func (e *APIError) Error() string {
	return e.Message
}

// doRequest makes an HTTP request to the Call API
func (c *Client) doRequest(method, endpoint string, payload interface{}, response interface{}) error {
	url := callAPIBaseURL + endpoint

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

	log.Printf("[CallAPI] %s request to %s", method, url)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check for errors
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errorBody bytes.Buffer
		errorBody.ReadFrom(resp.Body)
		errorMsg := fmt.Sprintf("Call API error (status %d): %s", resp.StatusCode, errorBody.String())
		log.Printf("[CallAPI] %s", errorMsg)
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

// MakeCall initiates a call via Call API
func (c *Client) MakeCall(req MakeCallRequest) (*MakeCallResponse, error) {
	log.Printf("[CallAPI] Initiating call from %s to %s (extension: %s)", req.Caller, req.Callee, req.Extension)

	var response MakeCallResponse
	err := c.doRequest(http.MethodPost, "/extensions/phone/calls", req, &response)
	if err != nil {
		return nil, err
	}

	log.Printf("[CallAPI] Call initiated successfully - CallID: %s", response.CallID)
	return &response, nil
}

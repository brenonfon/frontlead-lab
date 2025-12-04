package callapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/r3labs/sse/v2"
	"gopkg.in/cenkalti/backoff.v1"
)

const (
	callAPIBaseURL = "https://providersupportdata-test.cloud-cfg.com/v1"
	sseURL         = "https://providersupportdata.cloud-cfg.com/v1/extensions/phone/calls"
)

// Client represents a Call API client
type Client struct {
	Agents     []*SalesAgent
	APIKey     string
	BOT        string
	HTTPClient *http.Client
	sseClient  *sse.Client
	ctx        context.Context
	cancel     context.CancelFunc
}

// callEvent represents a call event from the SSE stream
type callEvent struct {
	ID      string    `json:"uuid"`
	Updated time.Time `json:"updated"`
	LegA    string    `json:"caller"`
	LegB    string    `json:"callee"`
	State   string    `json:"state"`
}

// NewClient creates a new Call API client
func NewClient(apiKey, bot string, phones []string) *Client {
	log.Println("[CallAPI] Initializing Call API client...")
	agents := make([]*SalesAgent, 0, len(phones))
	for _, phone := range phones {
		agent := &SalesAgent{
			Phone:       phone,
			IsAvailable: true,
		}
		agents = append(agents, agent)
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Create SSE client
	sseClient := sse.NewClient(sseURL)
	sseClient.ReconnectStrategy = &backoff.StopBackOff{}
	sseClient.Connection = &http.Client{}
	sseClient.Headers = map[string]string{
		"Authorization": "Bearer " + apiKey,
		"Accept":        "text/event-stream",
	}

	client := &Client{
		APIKey:     apiKey,
		BOT:        bot,
		Agents:     agents,
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
		sseClient:  sseClient,
		ctx:        ctx,
		cancel:     cancel,
	}

	// Start listening to SSE stream in a goroutine
	go client.listenToCallEvents()

	return client
}

// listenToCallEvents listens to the SSE stream for call events
func (c *Client) listenToCallEvents() {
	log.Println("[CallAPI] Starting SSE stream listener...")

	events := make(chan *sse.Event)

	// Subscribe to all events (empty string means all events)
	err := c.sseClient.SubscribeChanRawWithContext(c.ctx, events)
	if err != nil {
		log.Printf("[CallAPI] Error subscribing to SSE stream: %v", err)
		return
	}

	for {
		select {
		case <-c.ctx.Done():
			log.Println("[CallAPI] SSE stream listener stopped")
			return
		case sseEvent := <-events:
			if len(sseEvent.Data) == 0 {
				continue
			}

			event := callEvent{}
			err := json.Unmarshal(sseEvent.Data, &event)
			if err != nil {
				log.Printf("[CallAPI] Error unmarshaling call event: %v", err)
				continue
			}

			log.Printf("[CallAPI] Call Event - ID: %s, State: %s, LegA: %s, LegB: %s, Updated: %s",
				event.ID, event.State, event.LegA, event.LegB, event.Updated)

			// Check if the call has ended and update agent availability
			if event.State == "hangup" {
				// LegB is the customer's phone number (callee)
				customerPhone := event.LegB

				// Check if there's an active agent for this customer
				if SetActiveAgentReady(customerPhone, true) {
					log.Printf("[CallAPI] Agent marked as ready after call hangup for customer: %s", customerPhone)
				} else {
					log.Printf("[CallAPI] No active agent found for customer: %s", customerPhone)
				}
			}
		}
	}
}

// Close closes the SSE client and cancels the context
func (c *Client) Close() {
	log.Println("[CallAPI] Closing Call API client...")
	c.cancel()
}

func (c *Client) GetAvailableAgent() *SalesAgent {
	for _, agent := range c.Agents {
		if agent.IsAvailable {
			return agent
		}
	}
	return nil
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

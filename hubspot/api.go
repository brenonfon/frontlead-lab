package hubspot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

const (
	hubspotURL        = "https://api.hubapi.com/crm/v3/objects/contacts/search"
	hubspotContactURL = "https://api.hubapi.com/crm/v3/objects/contacts"
)

var (
	contactProperties = []string{"email", "phone", "firstname", "lastname", "company", "lifecyclestage"}
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
	ID             string `json:"id,omitempty"`
	Email          string `json:"email,omitempty"`
	Phone          string `json:"phone,omitempty"`
	FirstName      string `json:"firstname,omitempty"`
	LastName       string `json:"lastname,omitempty"`
	Company        string `json:"company,omitempty"`
	LifecycleStage string `json:"lifecyclestage,omitempty"`
}

// Filter represents a single filter criterion
type Filter struct {
	PropertyName string `json:"propertyName"`
	Operator     string `json:"operator"`
	Value        string `json:"value"`
}

// FilterGroup represents a group of filters
type FilterGroup struct {
	Filters []Filter `json:"filters"`
}

// SearchRequest represents a search request payload
type SearchRequest struct {
	FilterGroups []FilterGroup `json:"filterGroups"`
	Properties   []string      `json:"properties"`
	Limit        int           `json:"limit"`
}

// ContactResult represents a single contact result
type ContactResult struct {
	ID         string            `json:"id"`
	Properties map[string]string `json:"properties"`
	CreatedAt  string            `json:"createdAt"`
	UpdatedAt  string            `json:"updatedAt"`
	Archived   bool              `json:"archived"`
}

// PagingNext represents pagination information
type PagingNext struct {
	After string `json:"after"`
	Link  string `json:"link"`
}

// Paging represents pagination details
type Paging struct {
	Next *PagingNext `json:"next,omitempty"`
	Prev *PagingNext `json:"prev,omitempty"`
}

// ContactSearchResponse represents the response from searching contacts
type ContactSearchResponse struct {
	Results []ContactResult `json:"results"`
	Total   int             `json:"total"`
	Paging  *Paging         `json:"paging,omitempty"`
}

// GetAllContactsResponse represents the response from getting all contacts
type GetAllContactsResponse struct {
	Results []ContactResult `json:"results"`
	Paging  *Paging         `json:"paging,omitempty"`
}

// AssociationType represents an association type
type AssociationType struct {
	AssociationCategory string `json:"associationCategory"`
	AssociationTypeID   int    `json:"associationTypeId"`
}

// AssociationTo represents the target of an association
type AssociationTo struct {
	ID string `json:"id"`
}

// Association represents an association with another CRM object
type Association struct {
	To    AssociationTo     `json:"to"`
	Types []AssociationType `json:"types"`
}

// CreateContactRequest represents a request to create a new contact
type CreateContactRequest struct {
	Properties   map[string]string `json:"properties"`
	Associations []Association     `json:"associations,omitempty"`
}

// ContactEntity represents the created contact entity in the response
type ContactEntity struct {
	ID         string            `json:"id"`
	Properties map[string]string `json:"properties"`
	CreatedAt  string            `json:"createdAt"`
	UpdatedAt  string            `json:"updatedAt"`
	Archived   bool              `json:"archived"`
}

// CreateContactResponse represents the response from creating a contact
type CreateContactResponse struct {
	CreatedResourceID string        `json:"createdResourceId"`
	Entity            ContactEntity `json:"entity"`
	Location          string        `json:"location"`
}

// UpdateContactRequest represents a request to update a contact
type UpdateContactRequest struct {
	Properties map[string]string `json:"properties"`
}

// UpdateContactResponse represents the response from updating a contact
type UpdateContactResponse struct {
	ID         string            `json:"id"`
	Properties map[string]string `json:"properties"`
	CreatedAt  string            `json:"createdAt"`
	UpdatedAt  string            `json:"updatedAt"`
	Archived   bool              `json:"archived"`
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

// GetContactByEmail checks if a contact exists by email
func (c *Client) GetContactByEmail(email string) (*ContactInfo, error) {
	log.Printf("[HubSpot] Searching for contact by email: %s", email)

	searchReq := &SearchRequest{
		FilterGroups: []FilterGroup{
			{
				Filters: []Filter{
					{
						PropertyName: "email",
						Operator:     "EQ",
						Value:        email,
					},
				},
			},
		},
		Properties: contactProperties,
		Limit:      1,
	}

	response, err := c.searchContacts(searchReq)
	if err != nil {
		log.Printf("[HubSpot] Error searching contact by email %s: %v", email, err)
		return nil, fmt.Errorf("failed to search contact by email: %w", err)
	}

	if response.Total == 0 {
		log.Printf("[HubSpot] No contact found with email: %s", email)
		return &ContactInfo{Exists: false}, nil
	}

	log.Printf("[HubSpot] Contact found with email %s: ID=%s", email, response.Results[0].ID)
	return buildContactInfo(response.Results[0]), nil
}

// GetContactByPhone checks if a contact exists by phone number
func (c *Client) GetContactByPhone(phoneNumber string) (*ContactInfo, error) {
	log.Printf("[HubSpot] Searching for contact by phone: %s", phoneNumber)

	searchReq := &SearchRequest{
		FilterGroups: []FilterGroup{
			{
				Filters: []Filter{
					{
						PropertyName: "phone",
						Operator:     "EQ",
						Value:        phoneNumber,
					},
				},
			},
		},
		Properties: contactProperties,
		Limit:      1,
	}

	response, err := c.searchContacts(searchReq)
	if err != nil {
		log.Printf("[HubSpot] Error searching contact by phone %s: %v", phoneNumber, err)
		return nil, fmt.Errorf("failed to search contact by phone: %w", err)
	}

	if response.Total == 0 {
		log.Printf("[HubSpot] No contact found with phone: %s", phoneNumber)
		return &ContactInfo{Exists: false}, nil
	}

	log.Printf("[HubSpot] Contact found with phone %s: ID=%s", phoneNumber, response.Results[0].ID)
	return buildContactInfo(response.Results[0]), nil
}

// GetAllContacts retrieves all contacts with optional filtering and pagination
func (c *Client) GetAllContacts(limit int, after string, properties []string, archived bool) (*GetAllContactsResponse, error) {
	log.Printf("[HubSpot] Fetching all contacts (limit=%d, archived=%v, properties=%d)", limit, archived, len(properties))

	url := hubspotContactURL + "?"

	// Build query parameters
	params := make(map[string]string)

	if limit > 0 {
		params["limit"] = fmt.Sprintf("%d", limit)
	}

	if after != "" {
		params["after"] = after
	}

	if archived {
		params["archived"] = "true"
	}

	if len(properties) > 0 {
		for _, prop := range properties {
			url += fmt.Sprintf("properties=%s&", prop)
		}
	}

	// Add other parameters
	first := len(properties) == 0
	for key, value := range params {
		if !first {
			url += "&"
		}
		url += fmt.Sprintf("%s=%s", key, value)
		first = false
	}

	var response GetAllContactsResponse
	err := c.doRequest(http.MethodGet, url, nil, &response)
	if err != nil {
		log.Printf("[HubSpot] Error fetching all contacts: %v", err)
		return nil, fmt.Errorf("failed to get all contacts: %w", err)
	}

	log.Printf("[HubSpot] Successfully fetched %d contacts", len(response.Results))
	return &response, nil
}

// CreateContact creates a new contact in HubSpot
func (c *Client) CreateContact(properties map[string]string, associations []Association) (*CreateContactResponse, error) {
	log.Printf("[HubSpot] Creating new contact with properties: %v", properties)

	createReq := &CreateContactRequest{
		Properties:   properties,
		Associations: associations,
	}

	var response CreateContactResponse
	err := c.doRequest(http.MethodPost, hubspotContactURL, createReq, &response)
	if err != nil {
		log.Printf("[HubSpot] Error creating contact: %v", err)
		return nil, fmt.Errorf("failed to create contact: %w", err)
	}

	log.Printf("[HubSpot] Successfully created contact with ID: %s", response.Entity.ID)
	return &response, nil
}

// UpdateContactByPhone updates a contact identified by phone number
func (c *Client) UpdateContactByPhone(phoneNumber string, properties map[string]string) (*UpdateContactResponse, error) {
	log.Printf("[HubSpot] Updating contact by phone: %s with properties: %v", phoneNumber, properties)

	// Get the full contact ID from the search
	searchReq := &SearchRequest{
		FilterGroups: []FilterGroup{
			{
				Filters: []Filter{
					{
						PropertyName: "phone",
						Operator:     "EQ",
						Value:        phoneNumber,
					},
				},
			},
		},
		Properties: []string{"email", "phone"},
		Limit:      1,
	}

	searchResponse, err := c.searchContacts(searchReq)
	if err != nil || searchResponse.Total == 0 {
		return nil, fmt.Errorf("failed to get contact ID")
	}

	contactID := searchResponse.Results[0].ID
	log.Printf("[HubSpot] Found contact ID: %s for phone: %s", contactID, phoneNumber)

	// Now update the contact
	updateReq := &UpdateContactRequest{
		Properties: properties,
	}

	url := fmt.Sprintf("%s/%s", hubspotContactURL, contactID)
	var response UpdateContactResponse
	err = c.doRequest(http.MethodPatch, url, updateReq, &response)
	if err != nil {
		log.Printf("[HubSpot] Error updating contact: %v", err)
		return nil, fmt.Errorf("failed to update contact: %w", err)
	}

	log.Printf("[HubSpot] Successfully updated contact ID: %s", response.ID)
	return &response, nil
}

// searchContacts performs a contact search with the given search request
func (c *Client) searchContacts(searchReq *SearchRequest) (*ContactSearchResponse, error) {
	var response ContactSearchResponse
	err := c.doRequest(http.MethodPost, hubspotURL, searchReq, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// doRequest performs an HTTP request with proper headers and error handling
func (c *Client) doRequest(method, url string, payload, response any) error {
	var body *bytes.Buffer
	if payload != nil {
		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			log.Printf("[HubSpot] Failed to marshal payload: %v", err)
			return fmt.Errorf("failed to marshal payload: %w", err)
		}
		body = bytes.NewBuffer(payloadBytes)
	} else {
		body = &bytes.Buffer{}
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		log.Printf("[HubSpot] Failed to create request: %v", err)
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")

	log.Printf("[HubSpot] Making %s request to: %s", method, url)
	log.Printf("[HubSpot] API Key (first 20 chars): %s...", c.APIKey[:min(20, len(c.APIKey))])

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		log.Printf("[HubSpot] HTTP request failed: %v", err)
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Read the error response body for more details
		var errorBody bytes.Buffer
		errorBody.ReadFrom(resp.Body)

		log.Printf("[HubSpot] API returned non-OK status: %d", resp.StatusCode)
		log.Printf("[HubSpot] Response body: %s", errorBody.String())

		if resp.StatusCode == http.StatusUnauthorized {
			return fmt.Errorf("authentication failed (401): check your HubSpot API key and permissions. Error: %s", errorBody.String())
		}

		return fmt.Errorf("HubSpot API returned status code: %d, body: %s", resp.StatusCode, errorBody.String())
	}

	if response != nil {
		if err := json.NewDecoder(resp.Body).Decode(response); err != nil {
			log.Printf("[HubSpot] Failed to decode response: %v", err)
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	log.Printf("[HubSpot] Request completed successfully with status: %d", resp.StatusCode)
	return nil
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// buildContactInfo builds ContactInfo from a ContactResult
func buildContactInfo(contact ContactResult) *ContactInfo {
	return &ContactInfo{
		ID:             contact.ID,
		Exists:         true,
		Email:          contact.Properties["email"],
		Phone:          contact.Properties["phone"],
		FirstName:      contact.Properties["firstname"],
		LastName:       contact.Properties["lastname"],
		Company:        contact.Properties["company"],
		LifecycleStage: contact.Properties["lifecyclestage"],
	}
}

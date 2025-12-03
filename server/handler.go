package server

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"frontlead-lab/callapi"
	"frontlead-lab/hubspot"

	"github.com/gin-gonic/gin"
)

// Handler holds the dependencies for HTTP handlers
type Handler struct {
	hsClient      *hubspot.Client
	callAPIClient *callapi.Client
}

// NewHandler creates a new Handler instance
func NewHandler(hsClient *hubspot.Client, callAPIClient *callapi.Client) *Handler {
	return &Handler{
		hsClient:      hsClient,
		callAPIClient: callAPIClient,
	}
}

// handleError extracts the status code from HubSpot API errors and returns appropriate HTTP response
func (h *Handler) handleError(c *gin.Context, err error, context string) {
	log.Printf("[Handler] %s: %v", context, err)

	var apiErr *hubspot.APIError
	if errors.As(err, &apiErr) {
		// Return the same status code from HubSpot
		c.JSON(apiErr.StatusCode, gin.H{"error": apiErr.Message})
		return
	}

	// For non-API errors, return 500
	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
}

// GetContacts handles GET /contacts - unified endpoint for all contact operations
func (h *Handler) GetContacts(c *gin.Context) {
	email := c.Query("email")
	phoneNumber := c.Query("phone")
	contactID := c.Query("id")

	// Determine which operation to perform
	switch {
	case email != "":
		h.getContactByEmail(c, email)
	case phoneNumber != "":
		h.getContactByPhone(c, phoneNumber)
	case contactID != "":
		h.getContactByID(c, contactID)
	default:
		h.getAllContacts(c)
	}
}

// getContactByEmail retrieves a contact by email address
func (h *Handler) getContactByEmail(c *gin.Context, email string) {
	log.Printf("[Handler] GetContacts - Check mode: email=%s", email)

	contactInfo, err := h.hsClient.GetContactByEmail(email)
	if err != nil {
		h.handleError(c, err, "Error checking contact by email")
		return
	}

	log.Printf("[Handler] Contact check completed: exists=%v", contactInfo.Exists)
	c.JSON(http.StatusOK, contactInfo)
}

// getContactByPhone retrieves a contact by phone number
func (h *Handler) getContactByPhone(c *gin.Context, phoneNumber string) {
	log.Printf("[Handler] GetContacts - Check mode: phone=%s", phoneNumber)

	// Handle URL-decoded phone number (space becomes +)
	if len(phoneNumber) > 0 && phoneNumber[0] == ' ' {
		phoneNumber = "+" + phoneNumber[1:]
	}

	contactInfo, err := h.hsClient.GetContactByPhone(phoneNumber)
	if err != nil {
		h.handleError(c, err, "Error checking contact by phone")
		return
	}

	log.Printf("[Handler] Contact check completed: exists=%v", contactInfo.Exists)
	c.JSON(http.StatusOK, contactInfo)
}

// getContactByID retrieves a contact by HubSpot ID
func (h *Handler) getContactByID(c *gin.Context, contactID string) {
	log.Printf("[Handler] GetContacts - Check mode: id=%s", contactID)

	contactInfo, err := h.hsClient.GetContactByID(contactID)
	if err != nil {
		h.handleError(c, err, "Error checking contact by ID")
		return
	}

	log.Printf("[Handler] Contact check completed: exists=%v", contactInfo.Exists)
	c.JSON(http.StatusOK, contactInfo)
}

// getAllContacts retrieves all contacts with optional filtering and pagination
func (h *Handler) getAllContacts(c *gin.Context) {
	log.Printf("[Handler] GetContacts - List mode")

	// Parse query parameters
	limit := h.parseLimitParam(c)
	after := c.Query("after")
	archived := c.Query("archived") == "true"
	properties := h.parsePropertiesParam(c)

	log.Printf("[Handler] GetAllContacts request: limit=%d, archived=%v, properties=%d", limit, archived, len(properties))

	response, err := h.hsClient.GetAllContacts(limit, after, properties, archived)
	if err != nil {
		h.handleError(c, err, "Error fetching contacts")
		return
	}

	log.Printf("[Handler] Successfully retrieved %d contacts", len(response.Results))
	c.JSON(http.StatusOK, response)
}

// parseLimitParam extracts and validates the limit query parameter
func (h *Handler) parseLimitParam(c *gin.Context) int {
	limit := 10 // default
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	return limit
}

// parsePropertiesParam extracts properties from query parameters
func (h *Handler) parsePropertiesParam(c *gin.Context) []string {
	var properties []string
	if props := c.Query("properties"); props != "" {
		properties = []string{props}
	}
	if propsArray := c.QueryArray("properties"); len(propsArray) > 0 {
		properties = propsArray
	}
	return properties
}

// CreateContact handles POST /contacts - creates a new contact in HubSpot
func (h *Handler) CreateContact(c *gin.Context) {
	var req hubspot.CreateContactRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[Handler] Invalid request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body: " + err.Error()})
		return
	}

	log.Printf("[Handler] CreateContact request: properties=%v", req.Properties)

	// Validate that at least email or phone is provided
	if req.Properties["email"] == "" && req.Properties["phone"] == "" {
		log.Printf("[Handler] Bad request: neither email nor phone in properties")
		c.JSON(http.StatusBadRequest, gin.H{"error": "either email or phone property is required"})
		return
	}

	response, err := h.hsClient.CreateContact(req.Properties, req.Associations)
	if err != nil {
		h.handleError(c, err, "Error creating contact")
		return
	}

	log.Printf("[Handler] Contact created successfully: ID=%s", response.ID)
	c.JSON(http.StatusCreated, response)
}

// UpdateContactByPhone handles PATCH /contacts/phone/:phone - updates a contact by phone number
func (h *Handler) UpdateContactByPhone(c *gin.Context) {
	phoneNumber := c.Param("phone")

	var req hubspot.UpdateContactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[Handler] Invalid request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body: " + err.Error()})
		return
	}

	log.Printf("[Handler] UpdateContactByPhone request: phone=%s, properties=%v", phoneNumber, req.Properties)

	// Handle URL-decoded phone number (space becomes +)
	if len(phoneNumber) > 0 && phoneNumber[0] == ' ' {
		phoneNumber = "+" + phoneNumber[1:]
	}

	response, err := h.hsClient.UpdateContactByPhone(phoneNumber, req.Properties)
	if err != nil {
		h.handleError(c, err, "Error updating contact by phone")
		return
	}

	log.Printf("[Handler] Contact updated successfully: ID=%s", response.ID)
	c.JSON(http.StatusOK, response)
}

// UpdateContactByID handles PATCH /contacts/id/:id - updates a contact by HubSpot ID
func (h *Handler) UpdateContactByID(c *gin.Context) {
	contactID := c.Param("id")

	var req hubspot.UpdateContactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[Handler] Invalid request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body: " + err.Error()})
		return
	}

	log.Printf("[Handler] UpdateContactByID request: id=%s, properties=%v", contactID, req.Properties)

	response, err := h.hsClient.UpdateContactByID(contactID, req.Properties)
	if err != nil {
		h.handleError(c, err, "Error updating contact by ID")
		return
	}

	log.Printf("[Handler] Contact updated successfully: ID=%s", response.ID)
	c.JSON(http.StatusOK, response)
}

// HubSpotWebhookRequest represents the incoming request from HubSpot's custom code action
type HubSpotWebhookRequest struct {
	Phone string `json:"phone" binding:"required"`
}

// TriggerVoiceCall handles POST /trigger-call - bridge endpoint between HubSpot and Call API
// This endpoint receives calls from HubSpot's custom code action and forwards them to Call API
func (h *Handler) TriggerVoiceCall(c *gin.Context) {
	var req HubSpotWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[Handler] Invalid trigger-call request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid request body: " + err.Error(),
		})
		return
	}

	log.Printf("[Handler] TriggerVoiceCall request: phone=%s", req.Phone)

	// Build Call API request
	callReq := callapi.MakeCallRequest{
		Caller:        req.Phone,
		CallerContext: "global",
		Callee:        "498920171680",
		CalleeContext: "global",
	}

	// Call the Call API
	log.Printf("[Handler] Forwarding call request to Call API for phone: %s", req.Phone)
	response, err := h.callAPIClient.MakeCall(callReq)
	if err != nil {
		log.Printf("[Handler] Error calling Call API: %v", err)

		var apiErr *callapi.APIError
		statusCode := http.StatusInternalServerError
		if errors.As(err, &apiErr) {
			statusCode = apiErr.StatusCode
		}

		c.Status(statusCode)
		return
	}

	log.Printf("[Handler] Call initiated successfully - CallID: %s", response.CallID)

	// Return success response to HubSpot
	c.Status(http.StatusOK)
}

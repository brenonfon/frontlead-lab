package server

import (
	"log"
	"net/http"
	"strconv"

	"frontlead-lab/hubspot"

	"github.com/gin-gonic/gin"
)

// Handler holds the dependencies for HTTP handlers
type Handler struct {
	hsClient *hubspot.Client
}

// NewHandler creates a new Handler instance
func NewHandler(hsClient *hubspot.Client) *Handler {
	return &Handler{
		hsClient: hsClient,
	}
}

// GetContacts handles GET /contacts - unified endpoint for all contact operations
func (h *Handler) GetContacts(c *gin.Context) {
	email := c.Query("email")
	phoneNumber := c.Query("phone")

	// If email or phone is provided, check for specific contact
	if email != "" || phoneNumber != "" {
		log.Printf("[Handler] GetContacts - Check mode: email=%s, phone=%s", email, phoneNumber)

		var contactInfo *hubspot.ContactInfo
		var err error

		if email != "" {
			contactInfo, err = h.hsClient.GetContactByEmail(email)
			if err != nil {
				log.Printf("[Handler] Error checking contact by email: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		} else {
			// Handle URL-decoded phone number (space becomes +)
			if len(phoneNumber) > 0 && phoneNumber[0] == ' ' {
				phoneNumber = "+" + phoneNumber[1:]
			}

			contactInfo, err = h.hsClient.GetContactByPhone(phoneNumber)
			if err != nil {
				log.Printf("[Handler] Error checking contact by phone: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}

		log.Printf("[Handler] Contact check completed: exists=%v", contactInfo.Exists)
		c.JSON(http.StatusOK, contactInfo)
		return
	}

	// Otherwise, get all contacts
	log.Printf("[Handler] GetContacts - List mode")

	// Parse query parameters
	limit := 10 // default
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	after := c.Query("after")
	archived := c.Query("archived") == "true"

	// Parse properties (comma-separated or multiple query params)
	var properties []string
	if props := c.Query("properties"); props != "" {
		properties = []string{props}
	}
	if propsArray := c.QueryArray("properties"); len(propsArray) > 0 {
		properties = propsArray
	}

	log.Printf("[Handler] GetAllContacts request: limit=%d, archived=%v, properties=%d", limit, archived, len(properties))

	response, err := h.hsClient.GetAllContacts(limit, after, properties, archived)
	if err != nil {
		log.Printf("[Handler] Error fetching contacts: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Printf("[Handler] Successfully retrieved %d contacts", len(response.Results))
	c.JSON(http.StatusOK, response)
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
		log.Printf("[Handler] Error creating contact: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Printf("[Handler] Contact created successfully: ID=%s", response.Entity.ID)
	c.JSON(http.StatusOK, response)
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
		log.Printf("[Handler] Error updating contact: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Printf("[Handler] Contact updated successfully: ID=%s", response.ID)
	c.JSON(http.StatusOK, response)
}

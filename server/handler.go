package server

import (
	"net/http"

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

// CheckContact handles GET /contacts/check - checks if a contact exists in HubSpot by email or phone
func (h *Handler) CheckContact(c *gin.Context) {
	email := c.Query("email")
	phoneNumber := c.Query("phone")

	if email == "" && phoneNumber == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "either email or phone parameter is required"})
		return
	}

	var contactInfo *hubspot.ContactInfo
	var err error

	if email != "" {
		contactInfo, err = h.hsClient.IsExistingContact(email)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	} else {
		// Handle URL-decoded phone number (space becomes +)
		if len(phoneNumber) > 0 && phoneNumber[0] == ' ' {
			phoneNumber = "+" + phoneNumber[1:]
		}
		
		contactInfo, err = h.hsClient.IsExistingContactByPhone(phoneNumber)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, contactInfo)
}


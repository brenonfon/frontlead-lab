package server

import (
	"net/http"

	"frontlead-lab/salesforce"

	"github.com/gin-gonic/gin"
)

// Handler holds the dependencies for HTTP handlers
type Handler struct {
	sfClient *salesforce.Client
}

// NewHandler creates a new Handler instance
func NewHandler(sfClient *salesforce.Client) *Handler {
	return &Handler{
		sfClient: sfClient,
	}
}

// CreateLead handles POST /leads - creates a new lead in Salesforce
func (h *Handler) CreateLead(c *gin.Context) {
	var lead salesforce.Lead
	if err := c.ShouldBindJSON(&lead); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.sfClient.CreateLead(&lead)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// GetLead handles GET /leads/:id - retrieves a lead by ID from Salesforce
func (h *Handler) GetLead(c *gin.Context) {
	leadID := c.Param("id")

	lead, err := h.sfClient.GetLead(leadID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, lead)
}

// UpdateLead handles PATCH /leads/:id - updates an existing lead in Salesforce
func (h *Handler) UpdateLead(c *gin.Context) {
	leadID := c.Param("id")

	var lead salesforce.Lead
	if err := c.ShouldBindJSON(&lead); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.sfClient.UpdateLead(leadID, &lead); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Lead updated successfully"})
}

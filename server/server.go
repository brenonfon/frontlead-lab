package server

import (
	"log"

	"frontlead-lab/callapi"
	"frontlead-lab/hubspot"

	"github.com/gin-gonic/gin"
)

// Server represents the HTTP server
type Server struct {
	router  *gin.Engine
	handler *Handler
	port    string
}

// New creates a new HTTP server instance
func New(hsClient *hubspot.Client, callAPIClient *callapi.Client, port string) *Server {
	// Disable Gin debug logs
	gin.SetMode(gin.ReleaseMode)

	router := gin.Default()
	handler := NewHandler(hsClient, callAPIClient)

	s := &Server{
		router:  router,
		handler: handler,
		port:    port,
	}

	s.setupRoutes()

	return s
}

// setupRoutes configures all the routes for the server
func (s *Server) setupRoutes() {
	// Apply middleware to all routes
	s.router.Use(LoggingMiddleware())

	// Contact routes
	s.router.GET("/contacts", s.handler.GetContacts)
	s.router.POST("/contacts", s.handler.CreateContact)
	s.router.PATCH("/contacts/phone/:phone", s.handler.UpdateContactByPhone)
	s.router.PATCH("/contacts/id/:id", s.handler.UpdateContactByID)

	// Bridge endpoint between HubSpot and Botario
	s.router.POST("/trigger-call", s.handler.TriggerVoiceCall)
}

// Start starts the HTTP server
func (s *Server) Start() error {
	addr := ":" + s.port
	log.Printf("🚀 Server starting on http://localhost%s", addr)
	log.Printf("Routes:")
	log.Printf("  GET    /contacts                    - Get all contacts or check specific contact (query: ?email=, ?phone=, or ?id=)")
	log.Printf("  POST   /contacts                    - Create a new contact in HubSpot")
	log.Printf("  PATCH  /contacts/phone/:phone       - Update a contact by phone number")
	log.Printf("  PATCH  /contacts/id/:id             - Update a contact by HubSpot ID")
	log.Printf("  POST   /trigger-call                - Bridge endpoint: HubSpot → Call API voice call")

	return s.router.Run(addr)
}

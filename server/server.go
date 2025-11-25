package server

import (
	"log"

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
func New(hsClient *hubspot.Client, port string) *Server {
	router := gin.Default()
	handler := NewHandler(hsClient)

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
	s.router.GET("/contacts/check", s.handler.CheckContact)
}

// Start starts the HTTP server
func (s *Server) Start() error {
	addr := ":" + s.port
	log.Printf("🚀 Server starting on http://localhost%s", addr)
	log.Printf("Routes:")
	log.Printf("  GET    /contacts/check?email=<email>   - Check if contact exists in HubSpot")

	return s.router.Run(addr)
}

package server

import (
	"log"

	"frontlead-lab/salesforce"

	"github.com/gin-gonic/gin"
)

// Server represents the HTTP server
type Server struct {
	router  *gin.Engine
	handler *Handler
	port    string
}

// New creates a new HTTP server instance
func New(sfClient *salesforce.Client, port string) *Server {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	handler := NewHandler(sfClient)

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

	// Lead routes
	s.router.POST("/leads", s.handler.CreateLead)
	s.router.GET("/leads/:id", s.handler.GetLead)
	s.router.PATCH("/leads/:id", s.handler.UpdateLead)
}

// Start starts the HTTP server
func (s *Server) Start() error {
	addr := ":" + s.port
	log.Printf("🚀 Server starting on http://localhost%s", addr)
	log.Printf("Routes:")
	log.Printf("  POST   /leads       - Create a new lead")
	log.Printf("  GET    /leads/:id   - Get a lead by ID")
	log.Printf("  PATCH  /leads/:id   - Update a lead")

	return s.router.Run(addr)
}

package server

import (
	"fmt"
	"log"
	"net/http"

	"frontlead-lab/salesforce"

	"github.com/gin-gonic/gin"
)

var (
	// Global variable to store the authenticated client
	sfClientInstance *salesforce.Client
	authComplete     = make(chan bool)
)

// SetupOAuthRoutes adds OAuth callback routes to the router
func SetupOAuthRoutes(router *gin.Engine, config *salesforce.OAuthConfig) {
	router.GET("/auth/salesforce", func(c *gin.Context) {
		authURL := salesforce.GetAuthorizationURL(config)
		c.Redirect(http.StatusTemporaryRedirect, authURL)
	})

	router.GET("/auth/salesforce/callback", func(c *gin.Context) {
		code := c.Query("code")
		if code == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No authorization code received"})
			return
		}

		client, err := salesforce.ExchangeCodeForToken(config, code)
		if err != nil {
			log.Printf("Failed to exchange code for token: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		sfClientInstance = client
		authComplete <- true

		c.JSON(http.StatusOK, gin.H{
			"message": "Successfully authenticated with Salesforce! You can close this window.",
		})
	})
}

// GetAuthenticatedClient returns the authenticated Salesforce client
func GetAuthenticatedClient() *salesforce.Client {
	return sfClientInstance
}

// WaitForAuth waits for OAuth authentication to complete
func WaitForAuth() {
	log.Print("⏳ Waiting for OAuth authentication...")
	log.Println("📱 Please open your browser to: http://localhost:8081/auth/salesforce")
	<-authComplete
	log.Println("✅ Authentication complete!")
}

// StartAuthServer starts a temporary server for OAuth callback
func StartAuthServer(config *salesforce.OAuthConfig, port string) error {
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	SetupOAuthRoutes(router, config)

	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message":  "OAuth server is running",
			"auth_url": fmt.Sprintf("http://localhost:%s/auth/salesforce", port),
		})
	})

	addr := ":" + port
	log.Printf("🔐 OAuth server starting on http://localhost%s", addr)
	go func() {
		if err := router.Run(addr); err != nil {
			log.Fatalf("Failed to start OAuth server: %v", err)
		}
	}()

	WaitForAuth()
	return nil
}

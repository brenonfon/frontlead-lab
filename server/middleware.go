package server

import (
	"github.com/gin-gonic/gin"
)

// LoggingMiddleware is a middleware that passes requests through
// In the future, this can be extended to add logging, authentication, etc.
func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// For now, just pass the request to the next handler
		c.Next()
	}
}

package middleware

import (
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// CORS returns a gin.HandlerFunc for handling CORS
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Get allowed origins from environment or use default
		allowedOriginsEnv := os.Getenv("CORS_ALLOWED_ORIGINS")
		if allowedOriginsEnv == "" {
			// Default to common development origins
			allowedOriginsEnv = "http://localhost:3000,http://localhost:5173,http://localhost:8080"
		}

		// Split allowed origins
		allowedOrigins := strings.Split(allowedOriginsEnv, ",")

		// Check if the request origin is allowed
		originAllowed := false
		for _, allowedOrigin := range allowedOrigins {
			trimmedOrigin := strings.TrimSpace(allowedOrigin)
			if trimmedOrigin == "*" || trimmedOrigin == origin {
				originAllowed = true
				break
			}
		}

		// Set CORS headers
		if originAllowed {
			if origin != "" {
				// Use the actual origin instead of wildcard when credentials are used
				c.Header("Access-Control-Allow-Origin", origin)
			} else {
				// Fallback to first allowed origin if no origin header
				c.Header("Access-Control-Allow-Origin", strings.TrimSpace(allowedOrigins[0]))
			}
			c.Header("Access-Control-Allow-Credentials", "true")
		} else {
			// If origin not allowed, don't set credentials header
			c.Header("Access-Control-Allow-Origin", strings.TrimSpace(allowedOrigins[0]))
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Request-ID, X-Auth-Skip")
		c.Header("Access-Control-Expose-Headers", "Content-Length")
		c.Header("Access-Control-Max-Age", "86400") // Cache preflight for 24 hours

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

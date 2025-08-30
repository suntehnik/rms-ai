package routes

import (
	"net/http"
	"product-requirements-management/internal/config"

	"github.com/gin-gonic/gin"
)

// Setup configures all routes for the application
func Setup(router *gin.Engine, cfg *config.Config) {
	// Health check endpoints
	router.GET("/health", healthCheck)
	router.GET("/health/deep", deepHealthCheck)
	router.GET("/ready", readinessCheck)
	router.GET("/live", livenessCheck)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Placeholder routes - will be implemented in subsequent tasks
		v1.GET("/epics", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "Epics endpoint - to be implemented"})
		})

		v1.GET("/user-stories", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "User stories endpoint - to be implemented"})
		})

		v1.GET("/requirements", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "Requirements endpoint - to be implemented"})
		})
	}
}

// healthCheck provides basic health status
func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "product-requirements-management",
	})
}

// deepHealthCheck provides detailed health status including dependencies
func deepHealthCheck(c *gin.Context) {
	// TODO: Add database and Redis connectivity checks in future tasks
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "product-requirements-management",
		"checks": gin.H{
			"database": "not_configured",
			"redis":    "not_configured",
		},
	})
}

// readinessCheck indicates if the service is ready to accept traffic
func readinessCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ready",
	})
}

// livenessCheck indicates if the service is alive
func livenessCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "alive",
	})
}

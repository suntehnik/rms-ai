package routes

import (
	"context"
	"net/http"
	"product-requirements-management/internal/config"
	"product-requirements-management/internal/database"
	"time"

	"github.com/gin-gonic/gin"
)

// Setup configures all routes for the application
func Setup(router *gin.Engine, cfg *config.Config, db *database.DB) {
	// Health check endpoints
	router.GET("/health", healthCheck)
	router.GET("/health/deep", deepHealthCheck(db))
	router.GET("/ready", readinessCheck(db))
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
func deepHealthCheck(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		healthCheck := db.CheckHealth(ctx)
		
		status := http.StatusOK
		if healthCheck.Overall.Status != "healthy" {
			status = http.StatusServiceUnavailable
		}

		c.JSON(status, gin.H{
			"status":  healthCheck.Overall.Status,
			"service": "product-requirements-management",
			"message": healthCheck.Overall.Message,
			"checks": gin.H{
				"postgresql": gin.H{
					"status":  healthCheck.PostgreSQL.Status,
					"message": healthCheck.PostgreSQL.Message,
				},
				"redis": gin.H{
					"status":  healthCheck.Redis.Status,
					"message": healthCheck.Redis.Message,
				},
			},
		})
	}
}

// readinessCheck indicates if the service is ready to accept traffic
func readinessCheck(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
		defer cancel()

		if db.IsHealthy(ctx) {
			c.JSON(http.StatusOK, gin.H{
				"status": "ready",
			})
		} else {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "not_ready",
				"reason": "database_unhealthy",
			})
		}
	}
}

// livenessCheck indicates if the service is alive
func livenessCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "alive",
	})
}

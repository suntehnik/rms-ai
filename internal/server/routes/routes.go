package routes

import (
	"context"
	"net/http"
	"product-requirements-management/internal/config"
	"product-requirements-management/internal/database"
	"product-requirements-management/internal/handlers"
	"product-requirements-management/internal/repository"
	"product-requirements-management/internal/service"
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

	// Initialize repositories
	epicRepo := repository.NewEpicRepository(db.Postgres)
	userRepo := repository.NewUserRepository(db.Postgres)
	userStoryRepo := repository.NewUserStoryRepository(db.Postgres)
	acceptanceCriteriaRepo := repository.NewAcceptanceCriteriaRepository(db.Postgres)

	// Initialize services
	epicService := service.NewEpicService(epicRepo, userRepo)
	userStoryService := service.NewUserStoryService(userStoryRepo, epicRepo, userRepo)
	acceptanceCriteriaService := service.NewAcceptanceCriteriaService(acceptanceCriteriaRepo, userStoryRepo, userRepo)

	// Initialize handlers
	epicHandler := handlers.NewEpicHandler(epicService)
	userStoryHandler := handlers.NewUserStoryHandler(userStoryService)
	acceptanceCriteriaHandler := handlers.NewAcceptanceCriteriaHandler(acceptanceCriteriaService)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Epic routes
		epics := v1.Group("/epics")
		{
			epics.POST("", epicHandler.CreateEpic)
			epics.GET("", epicHandler.ListEpics)
			epics.GET("/:id", epicHandler.GetEpic)
			epics.PUT("/:id", epicHandler.UpdateEpic)
			epics.DELETE("/:id", epicHandler.DeleteEpic)
			epics.GET("/:id/user-stories", epicHandler.GetEpicWithUserStories)
			epics.POST("/:id/user-stories", userStoryHandler.CreateUserStoryInEpic)
			epics.PATCH("/:id/status", epicHandler.ChangeEpicStatus)
			epics.PATCH("/:id/assign", epicHandler.AssignEpic)
		}

		// User Story routes
		userStories := v1.Group("/user-stories")
		{
			userStories.POST("", userStoryHandler.CreateUserStory)
			userStories.GET("", userStoryHandler.ListUserStories)
			userStories.GET("/:id", userStoryHandler.GetUserStory)
			userStories.PUT("/:id", userStoryHandler.UpdateUserStory)
			userStories.DELETE("/:id", userStoryHandler.DeleteUserStory)
			userStories.GET("/:id/acceptance-criteria", acceptanceCriteriaHandler.GetAcceptanceCriteriaByUserStory)
			userStories.POST("/:id/acceptance-criteria", acceptanceCriteriaHandler.CreateAcceptanceCriteria)
			userStories.GET("/:id/requirements", userStoryHandler.GetUserStoryWithRequirements)
			userStories.PATCH("/:id/status", userStoryHandler.ChangeUserStoryStatus)
			userStories.PATCH("/:id/assign", userStoryHandler.AssignUserStory)
		}

		// Acceptance Criteria routes
		acceptanceCriteria := v1.Group("/acceptance-criteria")
		{
			acceptanceCriteria.GET("", acceptanceCriteriaHandler.ListAcceptanceCriteria)
			acceptanceCriteria.GET("/:id", acceptanceCriteriaHandler.GetAcceptanceCriteria)
			acceptanceCriteria.PUT("/:id", acceptanceCriteriaHandler.UpdateAcceptanceCriteria)
			acceptanceCriteria.DELETE("/:id", acceptanceCriteriaHandler.DeleteAcceptanceCriteria)
		}

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

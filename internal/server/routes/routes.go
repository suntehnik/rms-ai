package routes

import (
	"context"
	"net/http"
	"product-requirements-management/internal/config"
	"product-requirements-management/internal/database"
	"product-requirements-management/internal/handlers"
	"product-requirements-management/internal/logger"
	"product-requirements-management/internal/repository"
	"product-requirements-management/internal/server/middleware"
	"product-requirements-management/internal/service"
	"time"

	"github.com/gin-gonic/gin"
)

// Setup configures all routes for the application
func Setup(router *gin.Engine, cfg *config.Config, db *database.DB) {
	// Setup Swagger documentation routes
	middleware.SetupSwaggerRoutes(router, cfg)

	// Note: Health check endpoints are handled by the health checker in server.go
	// Only keeping non-conflicting health endpoints here
	router.GET("/ready", readinessCheck(db))
	router.GET("/live", livenessCheck)

	// Initialize repositories
	repos := repository.NewRepositories(db.Postgres)

	// Initialize Redis client (optional)
	var redisClient *database.RedisClient
	if cfg.Redis.Host != "" {
		var err error
		redisClient, err = database.NewRedisClient(&cfg.Redis, logger.Logger)
		if err != nil {
			logger.Logger.WithError(err).Warn("Failed to connect to Redis, search caching will be disabled")
			redisClient = nil
		}
	}

	// Initialize services
	epicService := service.NewEpicService(repos.Epic, repos.User)
	userStoryService := service.NewUserStoryService(repos.UserStory, repos.Epic, repos.User)
	acceptanceCriteriaService := service.NewAcceptanceCriteriaService(repos.AcceptanceCriteria, repos.UserStory, repos.User)
	requirementService := service.NewRequirementService(
		repos.Requirement,
		repos.RequirementType,
		repos.RelationshipType,
		repos.RequirementRelationship,
		repos.UserStory,
		repos.AcceptanceCriteria,
		repos.User,
	)
	configService := service.NewConfigService(
		repos.RequirementType,
		repos.RelationshipType,
		repos.Requirement,
		repos.RequirementRelationship,
		repos.StatusModel,
		repos.Status,
		repos.StatusTransition,
	)
	deletionService := service.NewDeletionService(
		repos.Epic,
		repos.UserStory,
		repos.AcceptanceCriteria,
		repos.Requirement,
		repos.RequirementRelationship,
		repos.Comment,
		repos.User,
		logger.Logger,
	)
	commentService := service.NewCommentService(repos)

	// Initialize search service
	var searchService *service.SearchService
	if redisClient != nil {
		searchService = service.NewSearchService(
			db.Postgres,
			redisClient.Client,
			repos.Epic,
			repos.UserStory,
			repos.AcceptanceCriteria,
			repos.Requirement,
		)
	} else {
		searchService = service.NewSearchService(
			db.Postgres,
			nil,
			repos.Epic,
			repos.UserStory,
			repos.AcceptanceCriteria,
			repos.Requirement,
		)
	}

	// Initialize navigation service
	navigationService := service.NewNavigationService(
		repos.Epic,
		repos.UserStory,
		repos.AcceptanceCriteria,
		repos.Requirement,
		repos.RequirementRelationship,
		repos.User,
	)

	// Initialize handlers
	epicHandler := handlers.NewEpicHandler(epicService)
	userStoryHandler := handlers.NewUserStoryHandler(userStoryService)
	acceptanceCriteriaHandler := handlers.NewAcceptanceCriteriaHandler(acceptanceCriteriaService)
	requirementHandler := handlers.NewRequirementHandler(requirementService)
	configHandler := handlers.NewConfigHandler(configService)
	deletionHandler := handlers.NewDeletionHandler(deletionService, logger.Logger)
	commentHandler := handlers.NewCommentHandler(commentService)
	searchHandler := handlers.NewSearchHandler(searchService, logger.Logger)
	navigationHandler := handlers.NewNavigationHandler(navigationService)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Search routes
		v1.GET("/search", searchHandler.Search)
		v1.GET("/search/suggestions", searchHandler.SearchSuggestions)

		// Hierarchy and navigation routes
		hierarchy := v1.Group("/hierarchy")
		{
			hierarchy.GET("", navigationHandler.GetHierarchy)
			hierarchy.GET("/epics/:id", navigationHandler.GetEpicHierarchy)
			hierarchy.GET("/user-stories/:id", navigationHandler.GetUserStoryHierarchy)
			hierarchy.GET("/path/:entity_type/:id", navigationHandler.GetEntityPath)
		}
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
			// Comprehensive deletion routes
			epics.GET("/:id/validate-deletion", deletionHandler.ValidateEpicDeletion)
			epics.DELETE("/:id/delete", deletionHandler.DeleteEpic)
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
			userStories.POST("/:id/requirements", requirementHandler.CreateRequirementInUserStory)
			userStories.PATCH("/:id/status", userStoryHandler.ChangeUserStoryStatus)
			userStories.PATCH("/:id/assign", userStoryHandler.AssignUserStory)
			// Comprehensive deletion routes
			userStories.GET("/:id/validate-deletion", deletionHandler.ValidateUserStoryDeletion)
			userStories.DELETE("/:id/delete", deletionHandler.DeleteUserStory)
		}

		// Acceptance Criteria routes
		acceptanceCriteria := v1.Group("/acceptance-criteria")
		{
			acceptanceCriteria.GET("", acceptanceCriteriaHandler.ListAcceptanceCriteria)
			acceptanceCriteria.GET("/:id", acceptanceCriteriaHandler.GetAcceptanceCriteria)
			acceptanceCriteria.PUT("/:id", acceptanceCriteriaHandler.UpdateAcceptanceCriteria)
			acceptanceCriteria.DELETE("/:id", acceptanceCriteriaHandler.DeleteAcceptanceCriteria)
			// Comprehensive deletion routes
			acceptanceCriteria.GET("/:id/validate-deletion", deletionHandler.ValidateAcceptanceCriteriaDeletion)
			acceptanceCriteria.DELETE("/:id/delete", deletionHandler.DeleteAcceptanceCriteria)
		}

		// Requirement routes
		requirements := v1.Group("/requirements")
		{
			requirements.POST("", requirementHandler.CreateRequirement)
			requirements.GET("", requirementHandler.ListRequirements)
			requirements.GET("/search", requirementHandler.SearchRequirements)
			requirements.GET("/:id", requirementHandler.GetRequirement)
			requirements.PUT("/:id", requirementHandler.UpdateRequirement)
			requirements.DELETE("/:id", requirementHandler.DeleteRequirement)
			requirements.GET("/:id/relationships", requirementHandler.GetRequirementWithRelationships)
			requirements.PATCH("/:id/status", requirementHandler.ChangeRequirementStatus)
			requirements.PATCH("/:id/assign", requirementHandler.AssignRequirement)
			requirements.POST("/relationships", requirementHandler.CreateRelationship)
			// Comprehensive deletion routes
			requirements.GET("/:id/validate-deletion", deletionHandler.ValidateRequirementDeletion)
			requirements.DELETE("/:id/delete", deletionHandler.DeleteRequirement)
		}

		// Requirement Relationship routes
		v1.DELETE("/requirement-relationships/:id", requirementHandler.DeleteRelationship)

		// Configuration routes (admin only)
		config := v1.Group("/config")
		{
			// Requirement Type routes
			requirementTypes := config.Group("/requirement-types")
			{
				requirementTypes.POST("", configHandler.CreateRequirementType)
				requirementTypes.GET("", configHandler.ListRequirementTypes)
				requirementTypes.GET("/:id", configHandler.GetRequirementType)
				requirementTypes.PUT("/:id", configHandler.UpdateRequirementType)
				requirementTypes.DELETE("/:id", configHandler.DeleteRequirementType)
			}

			// Relationship Type routes
			relationshipTypes := config.Group("/relationship-types")
			{
				relationshipTypes.POST("", configHandler.CreateRelationshipType)
				relationshipTypes.GET("", configHandler.ListRelationshipTypes)
				relationshipTypes.GET("/:id", configHandler.GetRelationshipType)
				relationshipTypes.PUT("/:id", configHandler.UpdateRelationshipType)
				relationshipTypes.DELETE("/:id", configHandler.DeleteRelationshipType)
			}

			// Status Model routes
			statusModels := config.Group("/status-models")
			{
				statusModels.POST("", configHandler.CreateStatusModel)
				statusModels.GET("", configHandler.ListStatusModels)
				statusModels.GET("/:id", configHandler.GetStatusModel)
				statusModels.PUT("/:id", configHandler.UpdateStatusModel)
				statusModels.DELETE("/:id", configHandler.DeleteStatusModel)
				statusModels.GET("/default/:entity_type", configHandler.GetDefaultStatusModel)
				statusModels.GET("/:id/statuses", configHandler.ListStatusesByModel)
				statusModels.GET("/:id/transitions", configHandler.ListStatusTransitionsByModel)
			}

			// Status routes
			statuses := config.Group("/statuses")
			{
				statuses.POST("", configHandler.CreateStatus)
				statuses.GET("/:id", configHandler.GetStatus)
				statuses.PUT("/:id", configHandler.UpdateStatus)
				statuses.DELETE("/:id", configHandler.DeleteStatus)
			}

			// Status Transition routes
			statusTransitions := config.Group("/status-transitions")
			{
				statusTransitions.POST("", configHandler.CreateStatusTransition)
				statusTransitions.GET("/:id", configHandler.GetStatusTransition)
				statusTransitions.PUT("/:id", configHandler.UpdateStatusTransition)
				statusTransitions.DELETE("/:id", configHandler.DeleteStatusTransition)
			}
		}

		// General deletion confirmation route
		v1.GET("/deletion/confirm", deletionHandler.GetDeletionConfirmation)

		// Comment routes
		comments := v1.Group("/comments")
		{
			comments.GET("/:id", commentHandler.GetComment)
			comments.PUT("/:id", commentHandler.UpdateComment)
			comments.DELETE("/:id", commentHandler.DeleteComment)
			comments.POST("/:id/resolve", commentHandler.ResolveComment)
			comments.POST("/:id/unresolve", commentHandler.UnresolveComment)
			comments.GET("/status/:status", commentHandler.GetCommentsByStatus)
			comments.GET("/:id/replies", commentHandler.GetCommentReplies)
			comments.POST("/:id/replies", commentHandler.CreateCommentReply)
		}

		// Entity comment routes - these need to be added to each entity group
		// Epic comments
		epics.GET("/:id/comments", commentHandler.GetEpicComments)
		epics.POST("/:id/comments", commentHandler.CreateEpicComment)
		epics.POST("/:id/comments/inline", commentHandler.CreateEpicInlineComment)
		epics.GET("/:id/comments/inline/visible", commentHandler.GetEpicVisibleInlineComments)
		epics.POST("/:id/comments/inline/validate", commentHandler.ValidateEpicInlineComments)

		// User Story comments
		userStories.GET("/:id/comments", commentHandler.GetUserStoryComments)
		userStories.POST("/:id/comments", commentHandler.CreateUserStoryComment)
		userStories.POST("/:id/comments/inline", commentHandler.CreateUserStoryInlineComment)
		userStories.GET("/:id/comments/inline/visible", commentHandler.GetUserStoryVisibleInlineComments)
		userStories.POST("/:id/comments/inline/validate", commentHandler.ValidateUserStoryInlineComments)

		// Acceptance Criteria comments
		acceptanceCriteria.GET("/:id/comments", commentHandler.GetAcceptanceCriteriaComments)
		acceptanceCriteria.POST("/:id/comments", commentHandler.CreateAcceptanceCriteriaComment)
		acceptanceCriteria.POST("/:id/comments/inline", commentHandler.CreateAcceptanceCriteriaInlineComment)
		acceptanceCriteria.GET("/:id/comments/inline/visible", commentHandler.GetAcceptanceCriteriaVisibleInlineComments)
		acceptanceCriteria.POST("/:id/comments/inline/validate", commentHandler.ValidateAcceptanceCriteriaInlineComments)

		// Requirement comments
		requirements.GET("/:id/comments", commentHandler.GetRequirementComments)
		requirements.POST("/:id/comments", commentHandler.CreateRequirementComment)
		requirements.POST("/:id/comments/inline", commentHandler.CreateRequirementInlineComment)
		requirements.GET("/:id/comments/inline/visible", commentHandler.GetRequirementVisibleInlineComments)
		requirements.POST("/:id/comments/inline/validate", commentHandler.ValidateRequirementInlineComments)
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

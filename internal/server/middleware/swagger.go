package middleware

import (
	"os"
	"product-requirements-management/internal/config"
	"product-requirements-management/internal/docs"
	"product-requirements-management/internal/logger"

	"github.com/gin-gonic/gin"
)

// SetupSwaggerRoutes configures Swagger UI routes based on comprehensive configuration
func SetupSwaggerRoutes(router *gin.Engine, cfg *config.Config) {
	// Apply environment-specific configuration
	environment := os.Getenv("ENVIRONMENT")
	if environment == "" {
		environment = "development"
	}

	// Apply environment configuration overrides
	docs.ApplyEnvironmentConfig(environment)

	// Get comprehensive Swagger configuration
	swaggerCfg := docs.DefaultSwaggerConfig()

	// Validate environment configuration and log warnings
	warnings := docs.ValidateEnvironmentConfig(environment)
	if len(warnings) > 0 && logger.Logger != nil {
		for _, warning := range warnings {
			logger.Logger.Warn("Swagger configuration warning: " + warning)
		}
	}

	// Log configuration status
	if logger.Logger != nil {
		if swaggerCfg.Enabled {
			logger.Logger.Infof("Setting up Swagger documentation for environment: %s", environment)
			logger.Logger.Infof("Swagger UI will be available at: %s/index.html", swaggerCfg.BasePath)

			// Log security settings
			if swaggerCfg.SecurityConfig.RequireAuth {
				logger.Logger.Info("Swagger authentication is enabled")
			}
			if swaggerCfg.SecurityConfig.HideInProduction && environment == "production" {
				logger.Logger.Info("Swagger is configured to be hidden in production")
			}
		} else {
			logger.Logger.Infof("Swagger documentation is disabled for environment: %s", environment)
		}
	}

	// Register Swagger routes with comprehensive configuration
	docs.RegisterSwaggerRoutes(router, swaggerCfg)

	// Add environment status endpoint
	router.GET("/api/v1/environment", func(c *gin.Context) {
		c.JSON(200, docs.GetEnvironmentStatus())
	})

	// Add deployment status endpoint
	router.GET("/api/v1/deployment", func(c *gin.Context) {
		c.JSON(200, docs.GetDeploymentStatus())
	})
}

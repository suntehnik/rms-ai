package middleware

import (
	"os"
	"product-requirements-management/internal/config"
	"product-requirements-management/internal/logger"
	"strconv"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SwaggerConfig holds Swagger-specific configuration
type SwaggerConfig struct {
	Enabled     bool
	BasePath    string
	Title       string
	Version     string
	Description string
}

// NewSwaggerConfig creates a new SwaggerConfig from environment variables
func NewSwaggerConfig() *SwaggerConfig {
	return &SwaggerConfig{
		Enabled:     getEnvAsBool("SWAGGER_ENABLED", true),
		BasePath:    getEnv("SWAGGER_BASE_PATH", "/swagger"),
		Title:       getEnv("SWAGGER_TITLE", "Product Requirements Management API"),
		Version:     getEnv("SWAGGER_VERSION", "1.0.0"),
		Description: getEnv("SWAGGER_DESCRIPTION", "API for managing product requirements through hierarchical structure"),
	}
}

// SetupSwaggerRoutes configures Swagger UI routes based on configuration
func SetupSwaggerRoutes(router *gin.Engine, cfg *config.Config) {
	swaggerCfg := NewSwaggerConfig()

	// Only serve Swagger documentation if enabled
	if !swaggerCfg.Enabled {
		// Safe logger call - check if logger is initialized
		if logger.Logger != nil {
			logger.Info("Swagger documentation is disabled")
		}
		return
	}

	// Safe logger call - check if logger is initialized
	if logger.Logger != nil {
		logger.Infof("Setting up Swagger documentation at %s/*any", swaggerCfg.BasePath)
	}

	// Configure Swagger UI with custom configuration
	url := ginSwagger.URL("/swagger/doc.json") // The url pointing to API definition
	router.GET(swaggerCfg.BasePath+"/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, url))

	// Add a redirect from base path to Swagger UI
	router.GET(swaggerCfg.BasePath, func(c *gin.Context) {
		c.Redirect(302, swaggerCfg.BasePath+"/index.html")
	})

	// Safe logger call - check if logger is initialized
	if logger.Logger != nil {
		logger.Infof("Swagger UI available at: %s/index.html", swaggerCfg.BasePath)
	}
}

// Helper functions for environment variable parsing
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvAsBool(key string, fallback bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return fallback
}

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

	// Configure Swagger UI with enhanced interactive testing capabilities
	url := ginSwagger.URL("/swagger/doc.json") // The url pointing to API definition

	// Enhanced Swagger UI configuration for better interactive testing
	swaggerHandler := ginSwagger.WrapHandler(swaggerFiles.Handler,
		url,
		ginSwagger.DefaultModelsExpandDepth(1),
		ginSwagger.DocExpansion("list"),
		ginSwagger.DeepLinking(true),
		ginSwagger.PersistAuthorization(true),
	)

	router.GET(swaggerCfg.BasePath+"/*any", swaggerHandler)

	// Add a redirect from base path to Swagger UI
	router.GET(swaggerCfg.BasePath, func(c *gin.Context) {
		c.Redirect(302, swaggerCfg.BasePath+"/index.html")
	})

	// Interactive testing helper endpoints are handled by the Swagger UI itself
	// No additional routes needed as they would conflict with the wildcard /*any route

	// Safe logger call - check if logger is initialized
	if logger.Logger != nil {
		logger.Infof("Swagger UI available at: %s/index.html", swaggerCfg.BasePath)
		logger.Infof("Interactive testing features enabled with authentication support")
	}
}

// Note: Interactive testing helper endpoints would conflict with the Swagger UI wildcard route
// The enhanced Swagger UI configuration provides the necessary interactive testing capabilities
// through the persistent authorization and improved UI features.

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

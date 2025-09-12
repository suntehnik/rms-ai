package docs

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SwaggerConfig holds configuration for Swagger documentation
type SwaggerConfig struct {
	Enabled     bool     `json:"enabled"`
	BasePath    string   `json:"base_path"`
	Title       string   `json:"title"`
	Version     string   `json:"version"`
	Description string   `json:"description"`
	Host        string   `json:"host"`
	Schemes     []string `json:"schemes"`
}

// DefaultSwaggerConfig returns default Swagger configuration
func DefaultSwaggerConfig() *SwaggerConfig {
	return &SwaggerConfig{
		Enabled:     getEnvBool("SWAGGER_ENABLED", true),
		BasePath:    getEnvString("SWAGGER_BASE_PATH", "/swagger"),
		Title:       getEnvString("SWAGGER_TITLE", "Product Requirements Management API"),
		Version:     getEnvString("SWAGGER_VERSION", "1.0.0"),
		Description: getEnvString("SWAGGER_DESCRIPTION", "API for managing product requirements through hierarchical structure"),
		Host:        getEnvString("SWAGGER_HOST", "localhost:8080"),
		Schemes:     []string{"http", "https"},
	}
}

// RegisterSwaggerRoutes registers Swagger UI routes with the Gin router
func RegisterSwaggerRoutes(router *gin.Engine, config *SwaggerConfig) {
	if !config.Enabled {
		return
	}

	// Create Swagger route group
	swaggerGroup := router.Group(config.BasePath)
	{
		// Swagger UI endpoint
		swaggerGroup.GET("/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

		// Health check for documentation
		swaggerGroup.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"status":  "ok",
				"service": "swagger-documentation",
				"version": config.Version,
			})
		})
	}
}

// SetupSwaggerInfo configures the Swagger documentation metadata
func SetupSwaggerInfo(config *SwaggerConfig) {
	// This will be populated by swag init command
	// The actual swagger info will be generated in docs/docs.go
}

// Helper functions for environment variable handling
func getEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return value == "true" || value == "1"
	}
	return defaultValue
}

// SwaggerInfo holds the basic information for Swagger documentation
// This will be used by the generated docs package
var SwaggerInfo = struct {
	Version     string
	Host        string
	BasePath    string
	Schemes     []string
	Title       string
	Description string
}{
	Version:     "1.0.0",
	Host:        "localhost:8080",
	BasePath:    "/api/v1",
	Schemes:     []string{"http", "https"},
	Title:       "Product Requirements Management API",
	Description: "API for managing product requirements through hierarchical structure of Epics, User Stories, and Requirements",
}

// GetSwaggerURL returns the full URL for accessing Swagger documentation
func GetSwaggerURL(config *SwaggerConfig) string {
	scheme := "http"
	if len(config.Schemes) > 0 {
		scheme = config.Schemes[0]
	}
	return fmt.Sprintf("%s://%s%s/index.html", scheme, config.Host, config.BasePath)
}

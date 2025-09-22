package swagger

import (
	"fmt"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SwaggerConfig holds configuration for Swagger documentation
type SwaggerConfig struct {
	Enabled        bool           `json:"enabled"`
	BasePath       string         `json:"base_path"`
	Title          string         `json:"title"`
	Version        string         `json:"version"`
	Description    string         `json:"description"`
	Host           string         `json:"host"`
	SecurityConfig SecurityConfig `json:"security"`
}

// SecurityConfig holds security-related Swagger configuration
type SecurityConfig struct {
	RequireAuth      bool `json:"require_auth"`
	HideInProduction bool `json:"hide_in_production"`
}

// EnvironmentStatus represents the current environment status
type EnvironmentStatus struct {
	Environment    string `json:"environment"`
	SwaggerEnabled bool   `json:"swagger_enabled"`
	RequireAuth    bool   `json:"require_auth"`
}

// DeploymentStatus represents the current deployment status
type DeploymentStatus struct {
	Environment    string `json:"environment"`
	SwaggerEnabled bool   `json:"swagger_enabled"`
	BasePath       string `json:"base_path"`
	Host           string `json:"host"`
}

// DefaultSwaggerConfig returns the default Swagger configuration
func DefaultSwaggerConfig() *SwaggerConfig {
	return &SwaggerConfig{
		Enabled:     getEnvBool("SWAGGER_ENABLED", true),
		BasePath:    getEnvString("SWAGGER_BASE_PATH", "/swagger"),
		Title:       getEnvString("SWAGGER_TITLE", "Product Requirements Management API"),
		Version:     getEnvString("SWAGGER_VERSION", "1.0.0"),
		Description: getEnvString("SWAGGER_DESCRIPTION", "API for managing product requirements"),
		Host:        getEnvString("SWAGGER_HOST", "localhost:8080"),
		SecurityConfig: SecurityConfig{
			RequireAuth:      getEnvBool("SWAGGER_REQUIRE_AUTH", false),
			HideInProduction: getEnvBool("SWAGGER_HIDE_IN_PRODUCTION", true),
		},
	}
}

// ApplyEnvironmentConfig applies environment-specific configuration
func ApplyEnvironmentConfig(environment string) {
	switch environment {
	case "development":
		os.Setenv("SWAGGER_ENABLED", "true")
		os.Setenv("SWAGGER_REQUIRE_AUTH", "false")
		os.Setenv("LOG_LEVEL", "debug")
		os.Setenv("CORS_ENABLED", "true")
		os.Setenv("RATE_LIMIT_ENABLED", "false")
		os.Setenv("COMPRESSION_ENABLED", "false")
		os.Setenv("CACHE_ENABLED", "false")
		fmt.Printf("Applying configuration for environment: %s\n", environment)
		fmt.Println("Set SWAGGER_REQUIRE_AUTH=false")
		fmt.Println("Set LOG_LEVEL=debug")
		fmt.Println("Set CORS_ENABLED=true")
		fmt.Println("Set RATE_LIMIT_ENABLED=false")
		fmt.Println("Set COMPRESSION_ENABLED=false")
		fmt.Println("Set CACHE_ENABLED=false")
		fmt.Println("Set SWAGGER_ENABLED=true")
	case "staging":
		os.Setenv("SWAGGER_ENABLED", "true")
		os.Setenv("SWAGGER_REQUIRE_AUTH", "true")
		os.Setenv("LOG_LEVEL", "info")
		os.Setenv("CORS_ENABLED", "true")
		os.Setenv("RATE_LIMIT_ENABLED", "true")
		os.Setenv("COMPRESSION_ENABLED", "true")
		os.Setenv("CACHE_ENABLED", "true")
	case "production":
		os.Setenv("SWAGGER_ENABLED", "false")
		os.Setenv("SWAGGER_REQUIRE_AUTH", "true")
		os.Setenv("LOG_LEVEL", "warn")
		os.Setenv("CORS_ENABLED", "false")
		os.Setenv("RATE_LIMIT_ENABLED", "true")
		os.Setenv("COMPRESSION_ENABLED", "true")
		os.Setenv("CACHE_ENABLED", "true")
		os.Setenv("CSP_ENABLED", "true")
		os.Setenv("SECURITY_HEADERS", "true")
	}
}

// ValidateEnvironmentConfig validates the environment configuration
func ValidateEnvironmentConfig(environment string) []string {
	var warnings []string

	if environment == "production" {
		if getEnvBool("SWAGGER_ENABLED", false) {
			warnings = append(warnings, "Swagger is enabled in production environment")
		}
		if !getEnvBool("SWAGGER_REQUIRE_AUTH", true) {
			warnings = append(warnings, "Swagger authentication is disabled in production")
		}
	}

	return warnings
}

// RegisterSwaggerRoutes registers Swagger routes with the given configuration
func RegisterSwaggerRoutes(router *gin.Engine, cfg *SwaggerConfig) {
	if !cfg.Enabled {
		return
	}

	// Configure Swagger UI
	url := ginSwagger.URL(fmt.Sprintf("%s/doc.json", cfg.BasePath))
	router.GET(fmt.Sprintf("%s/*any", cfg.BasePath), ginSwagger.WrapHandler(swaggerFiles.Handler, url))
}

// GetEnvironmentStatus returns the current environment status
func GetEnvironmentStatus() EnvironmentStatus {
	environment := getEnvString("ENVIRONMENT", "development")
	return EnvironmentStatus{
		Environment:    environment,
		SwaggerEnabled: getEnvBool("SWAGGER_ENABLED", true),
		RequireAuth:    getEnvBool("SWAGGER_REQUIRE_AUTH", false),
	}
}

// GetDeploymentStatus returns the current deployment status
func GetDeploymentStatus() DeploymentStatus {
	environment := getEnvString("ENVIRONMENT", "development")
	return DeploymentStatus{
		Environment:    environment,
		SwaggerEnabled: getEnvBool("SWAGGER_ENABLED", true),
		BasePath:       getEnvString("SWAGGER_BASE_PATH", "/swagger"),
		Host:           getEnvString("SWAGGER_HOST", "localhost:8080"),
	}
}

// Helper functions
func getEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return strings.ToLower(value) == "true"
	}
	return defaultValue
}

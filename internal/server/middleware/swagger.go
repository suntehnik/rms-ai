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

	// Add interactive testing helper endpoints
	setupInteractiveTestingEndpoints(router, swaggerCfg.BasePath)

	// Safe logger call - check if logger is initialized
	if logger.Logger != nil {
		logger.Infof("Swagger UI available at: %s/index.html", swaggerCfg.BasePath)
		logger.Infof("Interactive testing features enabled with authentication support")
	}
}

// setupInteractiveTestingEndpoints adds helper endpoints for interactive testing
func setupInteractiveTestingEndpoints(router *gin.Engine, basePath string) {
	// Add endpoint to validate authentication tokens
	router.POST(basePath+"/validate-token", func(c *gin.Context) {
		var request struct {
			Token string `json:"token" binding:"required"`
		}

		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(400, gin.H{"error": "Invalid request", "details": err.Error()})
			return
		}

		// Basic token format validation
		if len(request.Token) < 10 {
			c.JSON(400, gin.H{"error": "Token too short", "valid": false})
			return
		}

		c.JSON(200, gin.H{
			"message": "Token format appears valid",
			"valid":   true,
			"note":    "This is a format check only. Actual validation occurs during API calls.",
		})
	})

	// Add endpoint to get example tokens for testing (development only)
	if getEnvAsBool("SWAGGER_ENABLE_TEST_TOKENS", false) {
		router.GET(basePath+"/example-tokens", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"warning": "These are example tokens for development testing only",
				"tokens": gin.H{
					"administrator": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.example.admin.token",
					"user":          "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.example.user.token",
					"commenter":     "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.example.commenter.token",
				},
				"note": "Replace with actual tokens from your authentication system",
			})
		})
	}

	// Add endpoint to get API testing guide
	router.GET(basePath+"/testing-guide", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"title": "Interactive API Testing Guide",
			"authentication": gin.H{
				"method": "JWT Bearer Token",
				"header": "Authorization: Bearer <your-jwt-token>",
				"roles":  []string{"administrator", "user", "commenter"},
			},
			"steps": []string{
				"1. Obtain a JWT token from your authentication system",
				"2. Click the 'Authorize' button in Swagger UI",
				"3. Enter 'Bearer <your-token>' in the authorization field",
				"4. Click 'Authorize' to save the token",
				"5. Try out any endpoint using the 'Try it out' button",
				"6. Fill in required parameters and request body",
				"7. Click 'Execute' to send the request",
				"8. Review the response in the Swagger UI",
			},
			"example_requests": gin.H{
				"search":      "/api/v1/search?query=authentication&limit=10",
				"list_epics":  "/api/v1/epics?limit=20&offset=0",
				"create_epic": "POST /api/v1/epics with JSON body",
			},
			"troubleshooting": gin.H{
				"401_unauthorized": "Check that your token is valid and properly formatted",
				"403_forbidden":    "Verify your user role has permission for this endpoint",
				"400_bad_request":  "Review the request parameters and body format",
			},
		})
	})
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

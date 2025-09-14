package docs

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
	Enabled     bool     `json:"enabled"`
	BasePath    string   `json:"base_path"`
	Title       string   `json:"title"`
	Version     string   `json:"version"`
	Description string   `json:"description"`
	Host        string   `json:"host"`
	Schemes     []string `json:"schemes"`

	// Production deployment settings
	Environment     string          `json:"environment"`
	BuildInfo       BuildInfo       `json:"build_info"`
	CustomBranding  CustomBranding  `json:"custom_branding"`
	SecurityConfig  SecurityConfig  `json:"security_config"`
	UICustomization UICustomization `json:"ui_customization"`
}

// BuildInfo contains build and version information
type BuildInfo struct {
	Version     string `json:"version"`
	BuildTime   string `json:"build_time"`
	GitCommit   string `json:"git_commit"`
	GitBranch   string `json:"git_branch"`
	GoVersion   string `json:"go_version"`
	Environment string `json:"environment"`
}

// CustomBranding contains branding customization options
type CustomBranding struct {
	CompanyName    string `json:"company_name"`
	CompanyURL     string `json:"company_url"`
	LogoURL        string `json:"logo_url"`
	FaviconURL     string `json:"favicon_url"`
	PrimaryColor   string `json:"primary_color"`
	SecondaryColor string `json:"secondary_color"`
	FooterText     string `json:"footer_text"`
}

// SecurityConfig contains security-related documentation settings
type SecurityConfig struct {
	ShowSecuritySchemes bool     `json:"show_security_schemes"`
	HideInProduction    bool     `json:"hide_in_production"`
	AllowedOrigins      []string `json:"allowed_origins"`
	RequireAuth         bool     `json:"require_auth"`
}

// UICustomization contains Swagger UI customization options
type UICustomization struct {
	Theme                 string            `json:"theme"`                   // light, dark, auto
	DefaultModelRendering string            `json:"default_model_rendering"` // example, model
	DocExpansion          string            `json:"doc_expansion"`           // none, list, full
	ShowExtensions        bool              `json:"show_extensions"`
	ShowCommonExtensions  bool              `json:"show_common_extensions"`
	CustomCSS             string            `json:"custom_css"`
	CustomJS              string            `json:"custom_js"`
	AdditionalHeaders     map[string]string `json:"additional_headers"`
}

// DefaultSwaggerConfig returns default Swagger configuration with environment-based settings
func DefaultSwaggerConfig() *SwaggerConfig {
	environment := getEnvString("ENVIRONMENT", "development")

	return &SwaggerConfig{
		Enabled:     getSwaggerEnabled(environment),
		BasePath:    getEnvString("SWAGGER_BASE_PATH", "/swagger"),
		Title:       getEnvString("SWAGGER_TITLE", "Product Requirements Management API"),
		Version:     getEnvString("SWAGGER_VERSION", "1.0.0"),
		Description: getEnvString("SWAGGER_DESCRIPTION", "API for managing product requirements through hierarchical structure"),
		Host:        getEnvString("SWAGGER_HOST", "localhost:8080"),
		Schemes:     getSwaggerSchemes(environment),
		Environment: environment,

		BuildInfo: BuildInfo{
			Version:     getEnvString("BUILD_VERSION", "1.0.0"),
			BuildTime:   getEnvString("BUILD_TIME", "unknown"),
			GitCommit:   getEnvString("GIT_COMMIT", "unknown"),
			GitBranch:   getEnvString("GIT_BRANCH", "unknown"),
			GoVersion:   getEnvString("GO_VERSION", "unknown"),
			Environment: environment,
		},

		CustomBranding: CustomBranding{
			CompanyName:    getEnvString("SWAGGER_COMPANY_NAME", "Product Requirements Management"),
			CompanyURL:     getEnvString("SWAGGER_COMPANY_URL", ""),
			LogoURL:        getEnvString("SWAGGER_LOGO_URL", ""),
			FaviconURL:     getEnvString("SWAGGER_FAVICON_URL", ""),
			PrimaryColor:   getEnvString("SWAGGER_PRIMARY_COLOR", "#1976d2"),
			SecondaryColor: getEnvString("SWAGGER_SECONDARY_COLOR", "#424242"),
			FooterText:     getEnvString("SWAGGER_FOOTER_TEXT", "Product Requirements Management API"),
		},

		SecurityConfig: SecurityConfig{
			ShowSecuritySchemes: getEnvBool("SWAGGER_SHOW_SECURITY", true),
			HideInProduction:    getEnvBool("SWAGGER_HIDE_IN_PRODUCTION", false),
			AllowedOrigins:      getEnvStringSlice("SWAGGER_ALLOWED_ORIGINS", []string{"*"}),
			RequireAuth:         getEnvBool("SWAGGER_REQUIRE_AUTH", false),
		},

		UICustomization: UICustomization{
			Theme:                 getEnvString("SWAGGER_THEME", "light"),
			DefaultModelRendering: getEnvString("SWAGGER_MODEL_RENDERING", "example"),
			DocExpansion:          getEnvString("SWAGGER_DOC_EXPANSION", "list"),
			ShowExtensions:        getEnvBool("SWAGGER_SHOW_EXTENSIONS", true),
			ShowCommonExtensions:  getEnvBool("SWAGGER_SHOW_COMMON_EXTENSIONS", true),
			CustomCSS:             getEnvString("SWAGGER_CUSTOM_CSS", ""),
			CustomJS:              getEnvString("SWAGGER_CUSTOM_JS", ""),
			AdditionalHeaders:     getEnvStringMap("SWAGGER_ADDITIONAL_HEADERS"),
		},
	}
}

// getSwaggerEnabled determines if Swagger should be enabled based on environment
func getSwaggerEnabled(environment string) bool {
	// Check explicit environment variable first
	if enabled := os.Getenv("SWAGGER_ENABLED"); enabled != "" {
		return enabled == "true" || enabled == "1"
	}

	// Default behavior based on environment
	switch environment {
	case "production":
		return false // Disabled by default in production
	case "staging", "development", "testing":
		return true // Enabled in non-production environments
	default:
		return true // Default to enabled for unknown environments
	}
}

// getSwaggerSchemes returns appropriate schemes based on environment
func getSwaggerSchemes(environment string) []string {
	schemes := getEnvStringSlice("SWAGGER_SCHEMES", nil)
	if len(schemes) > 0 {
		return schemes
	}

	// Default schemes based on environment
	switch environment {
	case "production", "staging":
		return []string{"https"} // HTTPS only in production/staging
	case "development", "testing":
		return []string{"http", "https"} // Both HTTP and HTTPS in development
	default:
		return []string{"http", "https"}
	}
}

// RegisterSwaggerRoutes registers Swagger UI routes with the Gin router
func RegisterSwaggerRoutes(router *gin.Engine, config *SwaggerConfig) {
	if !config.Enabled {
		// Add a disabled endpoint that returns information about why Swagger is disabled
		router.GET(config.BasePath+"/*any", func(c *gin.Context) {
			c.JSON(404, gin.H{
				"error":       "Swagger documentation is disabled",
				"message":     "Documentation is not available in this environment",
				"environment": config.Environment,
			})
		})
		return
	}

	// Apply CORS settings if configured
	if len(config.SecurityConfig.AllowedOrigins) > 0 {
		router.Use(func(c *gin.Context) {
			origin := c.Request.Header.Get("Origin")
			if isAllowedOrigin(origin, config.SecurityConfig.AllowedOrigins) {
				c.Header("Access-Control-Allow-Origin", origin)
				c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
				c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
			}

			if c.Request.Method == "OPTIONS" {
				c.AbortWithStatus(204)
				return
			}

			c.Next()
		})
	}

	// Create Swagger route group with wildcard route
	swaggerGroup := router.Group(config.BasePath)
	{
		// Apply authentication middleware if required
		if config.SecurityConfig.RequireAuth {
			swaggerGroup.Use(requireAuthMiddleware())
		}

		// Swagger UI endpoint with customization
		swaggerGroup.GET("/*any", ginSwagger.WrapHandler(
			swaggerFiles.Handler,
			ginSwagger.URL(config.BasePath+"/doc.json"),
			ginSwagger.DeepLinking(true),
			ginSwagger.DocExpansion(config.UICustomization.DocExpansion),
			ginSwagger.DefaultModelsExpandDepth(-1),
		))
	}

	// Add API information endpoints on a separate path to avoid conflicts
	// These are registered outside the swagger group to avoid wildcard conflicts
	apiGroup := router.Group(config.BasePath + "-api")
	{
		apiGroup.GET("/info", func(c *gin.Context) {
			c.JSON(200, GetAPIInfo())
		})

		apiGroup.GET("/build", func(c *gin.Context) {
			c.JSON(200, config.BuildInfo)
		})

		apiGroup.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"status":      "ok",
				"service":     "swagger-documentation",
				"version":     config.Version,
				"environment": config.Environment,
				"build_info":  config.BuildInfo,
			})
		})

		apiGroup.GET("/config", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"branding":    config.CustomBranding,
				"ui_config":   config.UICustomization,
				"environment": config.Environment,
				"version":     config.Version,
			})
		})

		apiGroup.GET("/patterns", func(c *gin.Context) {
			c.JSON(200, GetAPIPatternDocumentation())
		})

		apiGroup.GET("/organization", func(c *gin.Context) {
			c.JSON(200, GetAPIOrganization())
		})

		apiGroup.GET("/guides", func(c *gin.Context) {
			c.JSON(200, GetUsageGuides())
		})

		apiGroup.GET("/quickstart", func(c *gin.Context) {
			c.JSON(200, GetQuickStartGuide())
		})

		// Custom CSS endpoint
		if config.UICustomization.CustomCSS != "" {
			apiGroup.GET("/custom.css", func(c *gin.Context) {
				c.Header("Content-Type", "text/css")
				c.String(200, config.UICustomization.CustomCSS)
			})
		}

		// Custom JavaScript endpoint
		if config.UICustomization.CustomJS != "" {
			apiGroup.GET("/custom.js", func(c *gin.Context) {
				c.Header("Content-Type", "application/javascript")
				c.String(200, config.UICustomization.CustomJS)
			})
		}
	}
}

// isAllowedOrigin checks if the origin is in the allowed origins list
func isAllowedOrigin(origin string, allowedOrigins []string) bool {
	for _, allowed := range allowedOrigins {
		if allowed == "*" || allowed == origin {
			return true
		}
	}
	return false
}

// requireAuthMiddleware returns a middleware that requires authentication for Swagger access
func requireAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(401, gin.H{
				"error":   "Authentication required",
				"message": "Swagger documentation requires authentication in this environment",
			})
			c.Abort()
			return
		}

		// Basic token validation (implement proper JWT validation as needed)
		if !isValidToken(authHeader) {
			c.JSON(401, gin.H{
				"error":   "Invalid token",
				"message": "The provided authentication token is invalid",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// isValidToken performs basic token validation (implement proper JWT validation)
func isValidToken(authHeader string) bool {
	// This is a placeholder - implement proper JWT validation
	// For now, just check if it starts with "Bearer "
	return len(authHeader) > 7 && authHeader[:7] == "Bearer "
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

func getEnvStringSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}

func getEnvStringMap(key string) map[string]string {
	result := make(map[string]string)
	if value := os.Getenv(key); value != "" {
		pairs := strings.Split(value, ",")
		for _, pair := range pairs {
			if kv := strings.SplitN(pair, ":", 2); len(kv) == 2 {
				result[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
			}
		}
	}
	return result
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
	Description: "Comprehensive API for managing product requirements through hierarchical structure of Epics → User Stories → Requirements with full-text search, comment system, and relationship mapping.",
}

// APIInfo provides comprehensive API information for documentation
type APIInfo struct {
	Title       string                 `json:"title"`
	Version     string                 `json:"version"`
	Description string                 `json:"description"`
	Host        string                 `json:"host"`
	BasePath    string                 `json:"base_path"`
	Schemes     []string               `json:"schemes"`
	Tags        []TagInfo              `json:"tags"`
	Patterns    map[string]interface{} `json:"patterns"`
	Statistics  map[string]interface{} `json:"statistics"`
}

// TagInfo provides detailed information about API tags
type TagInfo struct {
	Name         string        `json:"name"`
	Description  string        `json:"description"`
	ExternalDocs *ExternalDocs `json:"external_docs,omitempty"`
}

// ExternalDocs provides links to external documentation
type ExternalDocs struct {
	Description string `json:"description"`
	URL         string `json:"url"`
}

// GetAPIInfo returns comprehensive API information including patterns and statistics
func GetAPIInfo() *APIInfo {
	return &APIInfo{
		Title:       SwaggerInfo.Title,
		Version:     SwaggerInfo.Version,
		Description: SwaggerInfo.Description,
		Host:        SwaggerInfo.Host,
		BasePath:    SwaggerInfo.BasePath,
		Schemes:     SwaggerInfo.Schemes,
		Tags: []TagInfo{
			{
				Name:        "epics",
				Description: "Epic management endpoints for high-level features and initiatives. Epics serve as containers for user stories and provide project-level organization with reference IDs like EP-001, EP-002.",
			},
			{
				Name:        "user-stories",
				Description: "User story management within epics. User stories represent feature requirements from the user perspective and contain acceptance criteria and detailed requirements with reference IDs like US-001, US-002.",
			},
			{
				Name:        "acceptance-criteria",
				Description: "Acceptance criteria management for user stories. Define testable conditions that must be met for user story completion using EARS format (Easy Approach to Requirements Syntax) with reference IDs like AC-001, AC-002.",
			},
			{
				Name:        "requirements",
				Description: "Detailed requirement management with relationship mapping. Requirements provide technical specifications and can be linked with various relationship types (depends_on, blocks, relates_to, conflicts_with, derives_from) with reference IDs like REQ-001, REQ-002.",
			},
			{
				Name:        "comments",
				Description: "Comment system for collaboration and feedback. Supports both general comments and inline comments with threading, resolution tracking, and entity associations for comprehensive discussion management.",
			},
			{
				Name:        "search",
				Description: "Full-text search capabilities across all entities. Provides advanced filtering, sorting, and suggestion features for efficient content discovery with PostgreSQL-powered search indexing.",
			},
			{
				Name:        "navigation",
				Description: "Hierarchical navigation and entity relationship endpoints. Retrieve entity hierarchies, paths, and relationship structures for building navigation interfaces and breadcrumb systems.",
			},
			{
				Name:        "configuration",
				Description: "System configuration management for requirement types, relationship types, and status models. Administrative endpoints for customizing system behavior and workflows (requires administrator role).",
			},
			{
				Name:        "deletion",
				Description: "Comprehensive deletion management with dependency validation. Provides safe deletion with cascade options and dependency impact analysis to prevent data integrity issues.",
			},
			{
				Name:        "health",
				Description: "System health and monitoring endpoints for service status, database connectivity, and operational metrics. Used for health checks and monitoring integrations.",
			},
		},
		Patterns:   GetAPIPatternDocumentation(),
		Statistics: GetAPIStatistics(),
	}
}

// GetSwaggerURL returns the full URL for accessing Swagger documentation
func GetSwaggerURL(config *SwaggerConfig) string {
	scheme := "http"
	if len(config.Schemes) > 0 {
		scheme = config.Schemes[0]
	}
	return fmt.Sprintf("%s://%s%s/index.html", scheme, config.Host, config.BasePath)
}

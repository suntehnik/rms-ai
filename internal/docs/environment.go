package docs

import (
	"fmt"
	"os"
	"strconv"
)

// EnvironmentConfig contains environment-specific Swagger configuration
type EnvironmentConfig struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Settings    map[string]interface{} `json:"settings"`
	Overrides   map[string]string      `json:"overrides"`
}

// GetEnvironmentConfigs returns all available environment configurations
func GetEnvironmentConfigs() map[string]EnvironmentConfig {
	return map[string]EnvironmentConfig{
		"production": {
			Name:        "Production",
			Description: "Production environment with security-focused settings",
			Settings: map[string]interface{}{
				"swagger_enabled":       false,
				"swagger_require_auth":  true,
				"swagger_hide_schemas":  true,
				"swagger_hide_examples": true,
				"debug_mode":            false,
				"log_level":             "warn",
				"cors_enabled":          false,
				"rate_limiting":         true,
				"compression":           true,
				"caching":               true,
				"security_headers":      true,
				"csp_policy":            "strict",
				"allowed_origins":       []string{},
				"max_request_size":      "1MB",
				"timeout":               "30s",
			},
			Overrides: map[string]string{
				"SWAGGER_ENABLED":      "false",
				"SWAGGER_REQUIRE_AUTH": "true",
				"LOG_LEVEL":            "warn",
				"CORS_ENABLED":         "false",
				"RATE_LIMIT_ENABLED":   "true",
				"COMPRESSION_ENABLED":  "true",
				"CACHE_ENABLED":        "true",
			},
		},
		"staging": {
			Name:        "Staging",
			Description: "Staging environment for testing with production-like settings",
			Settings: map[string]interface{}{
				"swagger_enabled":       true,
				"swagger_require_auth":  true,
				"swagger_hide_schemas":  false,
				"swagger_hide_examples": false,
				"debug_mode":            false,
				"log_level":             "info",
				"cors_enabled":          true,
				"rate_limiting":         true,
				"compression":           true,
				"caching":               true,
				"security_headers":      true,
				"csp_policy":            "moderate",
				"allowed_origins":       []string{"https://staging.example.com"},
				"max_request_size":      "5MB",
				"timeout":               "60s",
			},
			Overrides: map[string]string{
				"SWAGGER_ENABLED":      "true",
				"SWAGGER_REQUIRE_AUTH": "true",
				"LOG_LEVEL":            "info",
				"CORS_ENABLED":         "true",
				"RATE_LIMIT_ENABLED":   "true",
				"COMPRESSION_ENABLED":  "true",
				"CACHE_ENABLED":        "true",
			},
		},
		"development": {
			Name:        "Development",
			Description: "Development environment with full access and debugging features",
			Settings: map[string]interface{}{
				"swagger_enabled":       true,
				"swagger_require_auth":  false,
				"swagger_hide_schemas":  false,
				"swagger_hide_examples": false,
				"debug_mode":            true,
				"log_level":             "debug",
				"cors_enabled":          true,
				"rate_limiting":         false,
				"compression":           false,
				"caching":               false,
				"security_headers":      false,
				"csp_policy":            "permissive",
				"allowed_origins":       []string{"*"},
				"max_request_size":      "10MB",
				"timeout":               "300s",
			},
			Overrides: map[string]string{
				"SWAGGER_ENABLED":      "true",
				"SWAGGER_REQUIRE_AUTH": "false",
				"LOG_LEVEL":            "debug",
				"CORS_ENABLED":         "true",
				"RATE_LIMIT_ENABLED":   "false",
				"COMPRESSION_ENABLED":  "false",
				"CACHE_ENABLED":        "false",
			},
		},
		"testing": {
			Name:        "Testing",
			Description: "Testing environment optimized for automated testing",
			Settings: map[string]interface{}{
				"swagger_enabled":       true,
				"swagger_require_auth":  false,
				"swagger_hide_schemas":  false,
				"swagger_hide_examples": false,
				"debug_mode":            true,
				"log_level":             "debug",
				"cors_enabled":          true,
				"rate_limiting":         false,
				"compression":           false,
				"caching":               false,
				"security_headers":      false,
				"csp_policy":            "permissive",
				"allowed_origins":       []string{"*"},
				"max_request_size":      "10MB",
				"timeout":               "60s",
			},
			Overrides: map[string]string{
				"SWAGGER_ENABLED":      "true",
				"SWAGGER_REQUIRE_AUTH": "false",
				"LOG_LEVEL":            "debug",
				"CORS_ENABLED":         "true",
				"RATE_LIMIT_ENABLED":   "false",
				"COMPRESSION_ENABLED":  "false",
				"CACHE_ENABLED":        "false",
			},
		},
	}
}

// ApplyEnvironmentConfig applies environment-specific configuration
func ApplyEnvironmentConfig(environment string) {
	configs := GetEnvironmentConfigs()
	config, exists := configs[environment]

	if !exists {
		fmt.Printf("Warning: Unknown environment '%s', using default settings\n", environment)
		return
	}

	fmt.Printf("Applying configuration for environment: %s\n", config.Name)

	// Apply environment variable overrides
	for key, value := range config.Overrides {
		if os.Getenv(key) == "" { // Only set if not already set
			os.Setenv(key, value)
			fmt.Printf("Set %s=%s\n", key, value)
		}
	}
}

// ValidateEnvironmentConfig validates the current environment configuration
func ValidateEnvironmentConfig(environment string) []string {
	var warnings []string

	configs := GetEnvironmentConfigs()
	_, exists := configs[environment]

	if !exists {
		warnings = append(warnings, fmt.Sprintf("Unknown environment: %s", environment))
		return warnings
	}

	// Validate production-specific settings
	if environment == "production" {
		if getEnvBool("SWAGGER_ENABLED", false) {
			warnings = append(warnings, "Swagger is enabled in production - consider disabling for security")
		}

		if !getEnvBool("SWAGGER_REQUIRE_AUTH", true) {
			warnings = append(warnings, "Swagger authentication is disabled in production")
		}

		if getEnvString("LOG_LEVEL", "warn") == "debug" {
			warnings = append(warnings, "Debug logging is enabled in production")
		}

		if getEnvBool("CORS_ENABLED", false) {
			allowedOrigins := getEnvStringSlice("SWAGGER_ALLOWED_ORIGINS", []string{})
			if len(allowedOrigins) == 0 || (len(allowedOrigins) == 1 && allowedOrigins[0] == "*") {
				warnings = append(warnings, "CORS is enabled with wildcard origins in production")
			}
		}
	}

	// Validate staging-specific settings
	if environment == "staging" {
		if !getEnvBool("SWAGGER_REQUIRE_AUTH", true) {
			warnings = append(warnings, "Consider enabling Swagger authentication in staging")
		}
	}

	// Validate required environment variables
	requiredVars := []string{"JWT_SECRET", "DATABASE_URL"}
	for _, varName := range requiredVars {
		if os.Getenv(varName) == "" {
			warnings = append(warnings, fmt.Sprintf("Required environment variable %s is not set", varName))
		}
	}

	// Validate configuration consistency
	if getEnvBool("SWAGGER_ENABLED", true) && getEnvBool("SWAGGER_REQUIRE_AUTH", false) && environment == "production" {
		warnings = append(warnings, "Swagger is enabled without authentication in production")
	}

	return warnings
}

// GetEnvironmentStatus returns the current environment status and configuration
func GetEnvironmentStatus() map[string]interface{} {
	environment := getEnvString("ENVIRONMENT", "development")
	configs := GetEnvironmentConfigs()
	config, exists := configs[environment]

	status := map[string]interface{}{
		"environment": environment,
		"valid":       exists,
		"warnings":    ValidateEnvironmentConfig(environment),
	}

	if exists {
		status["name"] = config.Name
		status["description"] = config.Description
		status["settings"] = config.Settings
	}

	// Add current environment variable values
	currentSettings := map[string]interface{}{
		"swagger_enabled":      getEnvBool("SWAGGER_ENABLED", true),
		"swagger_require_auth": getEnvBool("SWAGGER_REQUIRE_AUTH", false),
		"log_level":            getEnvString("LOG_LEVEL", "info"),
		"cors_enabled":         getEnvBool("CORS_ENABLED", true),
		"rate_limit_enabled":   getEnvBool("RATE_LIMIT_ENABLED", false),
		"compression_enabled":  getEnvBool("COMPRESSION_ENABLED", true),
		"cache_enabled":        getEnvBool("CACHE_ENABLED", true),
		"debug_mode":           getEnvBool("DEBUG_MODE", false),
	}

	status["current_settings"] = currentSettings

	return status
}

// GetRecommendedSettings returns recommended settings for each environment
func GetRecommendedSettings() map[string]interface{} {
	return map[string]interface{}{
		"production": map[string]interface{}{
			"description": "Security-focused settings for production deployment",
			"settings": map[string]string{
				"SWAGGER_ENABLED":      "false",
				"SWAGGER_REQUIRE_AUTH": "true",
				"LOG_LEVEL":            "warn",
				"CORS_ENABLED":         "false",
				"RATE_LIMIT_ENABLED":   "true",
				"COMPRESSION_ENABLED":  "true",
				"CACHE_ENABLED":        "true",
				"DEBUG_MODE":           "false",
				"CSP_ENABLED":          "true",
				"SECURITY_HEADERS":     "true",
			},
			"notes": []string{
				"Swagger should be disabled or require authentication",
				"Use HTTPS only (no HTTP)",
				"Enable rate limiting and compression",
				"Set restrictive CORS policy",
				"Use warn or error log level",
				"Enable security headers and CSP",
			},
		},
		"staging": map[string]interface{}{
			"description": "Production-like settings with documentation access for testing",
			"settings": map[string]string{
				"SWAGGER_ENABLED":      "true",
				"SWAGGER_REQUIRE_AUTH": "true",
				"LOG_LEVEL":            "info",
				"CORS_ENABLED":         "true",
				"RATE_LIMIT_ENABLED":   "true",
				"COMPRESSION_ENABLED":  "true",
				"CACHE_ENABLED":        "true",
				"DEBUG_MODE":           "false",
				"CSP_ENABLED":          "true",
				"SECURITY_HEADERS":     "true",
			},
			"notes": []string{
				"Enable Swagger with authentication for testing",
				"Use production-like security settings",
				"Allow specific origins for CORS",
				"Enable monitoring and metrics",
			},
		},
		"development": map[string]interface{}{
			"description": "Developer-friendly settings with full access and debugging",
			"settings": map[string]string{
				"SWAGGER_ENABLED":      "true",
				"SWAGGER_REQUIRE_AUTH": "false",
				"LOG_LEVEL":            "debug",
				"CORS_ENABLED":         "true",
				"RATE_LIMIT_ENABLED":   "false",
				"COMPRESSION_ENABLED":  "false",
				"CACHE_ENABLED":        "false",
				"DEBUG_MODE":           "true",
				"CSP_ENABLED":          "false",
				"SECURITY_HEADERS":     "false",
			},
			"notes": []string{
				"Full Swagger access without authentication",
				"Debug logging for development",
				"Permissive CORS for local development",
				"Disable caching and compression for easier debugging",
				"Relaxed security for development convenience",
			},
		},
	}
}

// GenerateEnvironmentFile generates a .env file for the specified environment
func GenerateEnvironmentFile(environment string) string {
	configs := GetEnvironmentConfigs()
	config, exists := configs[environment]

	if !exists {
		return fmt.Sprintf("# Error: Unknown environment '%s'\n", environment)
	}

	content := fmt.Sprintf("# Environment configuration for %s\n", config.Name)
	content += fmt.Sprintf("# %s\n\n", config.Description)

	content += "# Core Settings\n"
	content += fmt.Sprintf("ENVIRONMENT=%s\n", environment)

	content += "\n# Swagger Documentation Settings\n"
	for key, value := range config.Overrides {
		content += fmt.Sprintf("%s=%s\n", key, value)
	}

	content += "\n# Database Configuration\n"
	content += "DATABASE_URL=postgresql://user:password@localhost:5432/product_requirements\n"
	content += "REDIS_URL=redis://localhost:6379/0\n"

	content += "\n# Security Configuration\n"
	content += "JWT_SECRET=your-secret-key-here\n"

	if environment == "production" {
		content += "\n# Production-specific settings\n"
		content += "HTTPS_ONLY=true\n"
		content += "SECURE_COOKIES=true\n"
		content += "CSRF_PROTECTION=true\n"
	}

	content += "\n# Optional: Custom Branding\n"
	content += "# SWAGGER_COMPANY_NAME=Your Company\n"
	content += "# SWAGGER_COMPANY_URL=https://yourcompany.com\n"
	content += "# SWAGGER_LOGO_URL=https://yourcompany.com/logo.png\n"
	content += "# SWAGGER_PRIMARY_COLOR=#1976d2\n"

	return content
}

// Helper function to safely convert string to int
func safeAtoi(s string, defaultValue int) int {
	if i, err := strconv.Atoi(s); err == nil {
		return i
	}
	return defaultValue
}

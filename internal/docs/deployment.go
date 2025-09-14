package docs

import (
	"fmt"
	"os"
	"runtime"
	"time"
)

// DeploymentConfig contains deployment-specific configuration for Swagger documentation
type DeploymentConfig struct {
	Environment    string              `json:"environment"`
	DeploymentMode string              `json:"deployment_mode"` // standalone, kubernetes, docker
	HealthChecks   HealthCheckConfig   `json:"health_checks"`
	Monitoring     MonitoringConfig    `json:"monitoring"`
	Performance    PerformanceConfig   `json:"performance"`
	Accessibility  AccessibilityConfig `json:"accessibility"`
}

// HealthCheckConfig contains health check configuration
type HealthCheckConfig struct {
	Enabled          bool          `json:"enabled"`
	Interval         time.Duration `json:"interval"`
	Timeout          time.Duration `json:"timeout"`
	FailureThreshold int           `json:"failure_threshold"`
	SuccessThreshold int           `json:"success_threshold"`
}

// MonitoringConfig contains monitoring and metrics configuration
type MonitoringConfig struct {
	MetricsEnabled    bool     `json:"metrics_enabled"`
	TracingEnabled    bool     `json:"tracing_enabled"`
	LogLevel          string   `json:"log_level"`
	MetricsEndpoint   string   `json:"metrics_endpoint"`
	TracingEndpoint   string   `json:"tracing_endpoint"`
	AlertingEndpoints []string `json:"alerting_endpoints"`
}

// PerformanceConfig contains performance optimization settings
type PerformanceConfig struct {
	CacheEnabled       bool          `json:"cache_enabled"`
	CacheTTL           time.Duration `json:"cache_ttl"`
	CompressionEnabled bool          `json:"compression_enabled"`
	RateLimitEnabled   bool          `json:"rate_limit_enabled"`
	RateLimitRPS       int           `json:"rate_limit_rps"`
	MaxRequestSize     int64         `json:"max_request_size"`
}

// AccessibilityConfig contains accessibility and security settings
type AccessibilityConfig struct {
	PublicAccess    bool     `json:"public_access"`
	IPWhitelist     []string `json:"ip_whitelist"`
	RequiredHeaders []string `json:"required_headers"`
	CORSEnabled     bool     `json:"cors_enabled"`
	CSPEnabled      bool     `json:"csp_enabled"`
	CSPPolicy       string   `json:"csp_policy"`
}

// GetDeploymentConfig returns deployment configuration based on environment
func GetDeploymentConfig() *DeploymentConfig {
	environment := getEnvString("ENVIRONMENT", "development")

	config := &DeploymentConfig{
		Environment:    environment,
		DeploymentMode: getEnvString("DEPLOYMENT_MODE", "standalone"),

		HealthChecks: HealthCheckConfig{
			Enabled:          getEnvBool("HEALTH_CHECKS_ENABLED", true),
			Interval:         getEnvDuration("HEALTH_CHECK_INTERVAL", 30*time.Second),
			Timeout:          getEnvDuration("HEALTH_CHECK_TIMEOUT", 5*time.Second),
			FailureThreshold: getEnvInt("HEALTH_CHECK_FAILURE_THRESHOLD", 3),
			SuccessThreshold: getEnvInt("HEALTH_CHECK_SUCCESS_THRESHOLD", 1),
		},

		Monitoring: MonitoringConfig{
			MetricsEnabled:    getEnvBool("METRICS_ENABLED", true),
			TracingEnabled:    getEnvBool("TRACING_ENABLED", false),
			LogLevel:          getEnvString("LOG_LEVEL", "info"),
			MetricsEndpoint:   getEnvString("METRICS_ENDPOINT", "/metrics"),
			TracingEndpoint:   getEnvString("TRACING_ENDPOINT", ""),
			AlertingEndpoints: getEnvStringSlice("ALERTING_ENDPOINTS", []string{}),
		},

		Performance: PerformanceConfig{
			CacheEnabled:       getEnvBool("CACHE_ENABLED", true),
			CacheTTL:           getEnvDuration("CACHE_TTL", 1*time.Hour),
			CompressionEnabled: getEnvBool("COMPRESSION_ENABLED", true),
			RateLimitEnabled:   getEnvBool("RATE_LIMIT_ENABLED", false),
			RateLimitRPS:       getEnvInt("RATE_LIMIT_RPS", 100),
			MaxRequestSize:     getEnvInt64("MAX_REQUEST_SIZE", 10*1024*1024), // 10MB
		},

		Accessibility: AccessibilityConfig{
			PublicAccess:    getPublicAccessDefault(environment),
			IPWhitelist:     getEnvStringSlice("IP_WHITELIST", []string{}),
			RequiredHeaders: getEnvStringSlice("REQUIRED_HEADERS", []string{}),
			CORSEnabled:     getEnvBool("CORS_ENABLED", true),
			CSPEnabled:      getEnvBool("CSP_ENABLED", true),
			CSPPolicy:       getCSPPolicy(environment),
		},
	}

	return config
}

// getPublicAccessDefault returns default public access setting based on environment
func getPublicAccessDefault(environment string) bool {
	if value := os.Getenv("PUBLIC_ACCESS"); value != "" {
		return value == "true" || value == "1"
	}

	switch environment {
	case "production":
		return false // Private by default in production
	case "staging":
		return false // Private by default in staging
	case "development", "testing":
		return true // Public by default in development
	default:
		return false // Conservative default
	}
}

// getCSPPolicy returns Content Security Policy based on environment
func getCSPPolicy(environment string) string {
	if policy := os.Getenv("CSP_POLICY"); policy != "" {
		return policy
	}

	switch environment {
	case "production":
		return "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self'"
	case "staging":
		return "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self'"
	default:
		return "default-src 'self' 'unsafe-inline' 'unsafe-eval'; img-src 'self' data:; font-src 'self'"
	}
}

// Additional helper functions for environment variables
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := fmt.Sscanf(value, "%d", &defaultValue); err == nil && intValue == 1 {
			return defaultValue
		}
	}
	return defaultValue
}

func getEnvInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := fmt.Sscanf(value, "%d", &defaultValue); err == nil && intValue == 1 {
			return defaultValue
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

// GetBuildInformation returns comprehensive build information for deployment
func GetBuildInformation() map[string]interface{} {
	return map[string]interface{}{
		"version":         getEnvString("BUILD_VERSION", "unknown"),
		"build_time":      getEnvString("BUILD_TIME", "unknown"),
		"git_commit":      getEnvString("GIT_COMMIT", "unknown"),
		"git_branch":      getEnvString("GIT_BRANCH", "unknown"),
		"go_version":      runtime.Version(),
		"go_os":           runtime.GOOS,
		"go_arch":         runtime.GOARCH,
		"environment":     getEnvString("ENVIRONMENT", "development"),
		"deployment_mode": getEnvString("DEPLOYMENT_MODE", "standalone"),
		"build_user":      getEnvString("BUILD_USER", "unknown"),
		"build_host":      getEnvString("BUILD_HOST", "unknown"),
	}
}

// GetDeploymentStatus returns current deployment status and health
func GetDeploymentStatus() map[string]interface{} {
	config := GetDeploymentConfig()

	return map[string]interface{}{
		"status":          "running",
		"environment":     config.Environment,
		"deployment_mode": config.DeploymentMode,
		"uptime":          time.Since(startTime).String(),
		"health_checks": map[string]interface{}{
			"enabled":  config.HealthChecks.Enabled,
			"interval": config.HealthChecks.Interval.String(),
			"status":   "healthy",
		},
		"performance": map[string]interface{}{
			"cache_enabled":       config.Performance.CacheEnabled,
			"compression_enabled": config.Performance.CompressionEnabled,
			"rate_limit_enabled":  config.Performance.RateLimitEnabled,
		},
		"accessibility": map[string]interface{}{
			"public_access": config.Accessibility.PublicAccess,
			"cors_enabled":  config.Accessibility.CORSEnabled,
			"csp_enabled":   config.Accessibility.CSPEnabled,
		},
		"build_info": GetBuildInformation(),
	}
}

// GetEnvironmentConfiguration returns environment-specific configuration
func GetEnvironmentConfiguration(environment string) map[string]interface{} {
	switch environment {
	case "production":
		return map[string]interface{}{
			"swagger_enabled":     false,
			"debug_mode":          false,
			"log_level":           "warn",
			"metrics_enabled":     true,
			"tracing_enabled":     true,
			"public_access":       false,
			"rate_limit_enabled":  true,
			"cache_enabled":       true,
			"compression_enabled": true,
			"security_headers":    true,
			"csp_enabled":         true,
		}
	case "staging":
		return map[string]interface{}{
			"swagger_enabled":     true,
			"debug_mode":          false,
			"log_level":           "info",
			"metrics_enabled":     true,
			"tracing_enabled":     true,
			"public_access":       false,
			"rate_limit_enabled":  true,
			"cache_enabled":       true,
			"compression_enabled": true,
			"security_headers":    true,
			"csp_enabled":         true,
		}
	case "development":
		return map[string]interface{}{
			"swagger_enabled":     true,
			"debug_mode":          true,
			"log_level":           "debug",
			"metrics_enabled":     true,
			"tracing_enabled":     false,
			"public_access":       true,
			"rate_limit_enabled":  false,
			"cache_enabled":       false,
			"compression_enabled": false,
			"security_headers":    false,
			"csp_enabled":         false,
		}
	case "testing":
		return map[string]interface{}{
			"swagger_enabled":     true,
			"debug_mode":          true,
			"log_level":           "debug",
			"metrics_enabled":     false,
			"tracing_enabled":     false,
			"public_access":       true,
			"rate_limit_enabled":  false,
			"cache_enabled":       false,
			"compression_enabled": false,
			"security_headers":    false,
			"csp_enabled":         false,
		}
	default:
		return map[string]interface{}{
			"swagger_enabled":     true,
			"debug_mode":          false,
			"log_level":           "info",
			"metrics_enabled":     true,
			"tracing_enabled":     false,
			"public_access":       false,
			"rate_limit_enabled":  true,
			"cache_enabled":       true,
			"compression_enabled": true,
			"security_headers":    true,
			"csp_enabled":         true,
		}
	}
}

// Kubernetes-specific configuration
func GetKubernetesConfig() map[string]interface{} {
	return map[string]interface{}{
		"readiness_probe": map[string]interface{}{
			"path":                  "/ready",
			"port":                  8080,
			"initial_delay_seconds": 10,
			"period_seconds":        10,
			"timeout_seconds":       5,
			"failure_threshold":     3,
		},
		"liveness_probe": map[string]interface{}{
			"path":                  "/live",
			"port":                  8080,
			"initial_delay_seconds": 30,
			"period_seconds":        30,
			"timeout_seconds":       5,
			"failure_threshold":     3,
		},
		"resources": map[string]interface{}{
			"requests": map[string]string{
				"memory": "128Mi",
				"cpu":    "100m",
			},
			"limits": map[string]string{
				"memory": "512Mi",
				"cpu":    "500m",
			},
		},
		"service": map[string]interface{}{
			"type": "ClusterIP",
			"ports": []map[string]interface{}{
				{
					"name":        "http",
					"port":        80,
					"target_port": 8080,
					"protocol":    "TCP",
				},
			},
		},
	}
}

// Docker-specific configuration
func GetDockerConfig() map[string]interface{} {
	return map[string]interface{}{
		"image": map[string]interface{}{
			"repository":  getEnvString("DOCKER_REPOSITORY", "product-requirements-management"),
			"tag":         getEnvString("DOCKER_TAG", "latest"),
			"pull_policy": getEnvString("DOCKER_PULL_POLICY", "IfNotPresent"),
		},
		"ports": []map[string]interface{}{
			{
				"container_port": 8080,
				"protocol":       "TCP",
			},
		},
		"environment": []map[string]string{
			{"name": "ENVIRONMENT", "value": getEnvString("ENVIRONMENT", "production")},
			{"name": "LOG_LEVEL", "value": getEnvString("LOG_LEVEL", "info")},
			{"name": "SWAGGER_ENABLED", "value": getEnvString("SWAGGER_ENABLED", "false")},
		},
		"health_check": map[string]interface{}{
			"test":     []string{"CMD", "curl", "-f", "http://localhost:8080/health"},
			"interval": "30s",
			"timeout":  "5s",
			"retries":  3,
		},
	}
}

// startTime is used to calculate uptime
var startTime = time.Now()

package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds all configuration for the application
type Config struct {
	Server        ServerConfig
	Database      DatabaseConfig
	Redis         RedisConfig
	JWT           JWTConfig
	Log           LogConfig
	Observability ObservabilityConfig
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Port string
	Host string
}

// DatabaseConfig holds database connection configuration
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// RedisConfig holds Redis connection configuration
type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

// JWTConfig holds JWT-related configuration
type JWTConfig struct {
	Secret string
}

// LogConfig holds logging configuration
type LogConfig struct {
	Level  string
	Format string // json or text
}

// ObservabilityConfig holds observability configuration
type ObservabilityConfig struct {
	ServiceName     string
	ServiceVersion  string
	Environment     string
	MetricsEnabled  bool
	TracingEnabled  bool
	TracingEndpoint string
}

// Load loads configuration from environment variables with defaults
func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", ""),
			DBName:   getEnv("DB_NAME", "requirements_db"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
		JWT: JWTConfig{
			Secret: getEnv("JWT_SECRET", "your-secret-key"),
		},
		Log: LogConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
		},
		Observability: ObservabilityConfig{
			ServiceName:     getEnv("SERVICE_NAME", "product-requirements-management"),
			ServiceVersion:  getEnv("SERVICE_VERSION", "1.0.0"),
			Environment:     getEnv("ENVIRONMENT", "development"),
			MetricsEnabled:  getEnvAsBool("METRICS_ENABLED", true),
			TracingEnabled:  getEnvAsBool("TRACING_ENABLED", true),
			TracingEndpoint: getEnv("TRACING_ENDPOINT", "http://localhost:4318/v1/traces"),
		},
	}

	// Validate required configuration
	if cfg.JWT.Secret == "" || cfg.JWT.Secret == "your-secret-key" {
		return nil, fmt.Errorf("JWT_SECRET must be set in production")
	}

	return cfg, nil
}

// getEnv gets an environment variable with a fallback value
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// getEnvAsInt gets an environment variable as integer with a fallback value
func getEnvAsInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return fallback
}

// getEnvAsBool gets an environment variable as boolean with a fallback value
func getEnvAsBool(key string, fallback bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return fallback
}

package mcp

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

// Config holds the configuration for the MCP Server console application.
// Configuration is loaded from ~/.requirements-mcp/config.json
type Config struct {
	// BackendAPIURL is the base URL of the backend API server
	BackendAPIURL string `json:"backend_api_url"`

	// PATToken is the Personal Access Token for authentication with the backend API
	PATToken string `json:"pat_token"`

	// RequestTimeout is the timeout for HTTP requests to the backend API
	RequestTimeout string `json:"request_timeout"`

	// LogLevel controls the logging verbosity (debug, info, warn, error)
	LogLevel string `json:"log_level"`
}

// LoadConfig loads the MCP server configuration from ~/.requirements-mcp/config.json
// Returns an error if the configuration file cannot be read or parsed.
func LoadConfig() (*Config, error) {
	// Get user home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}

	// Construct config file path
	configPath := filepath.Join(homeDir, ".requirements-mcp", "config.json")

	return LoadConfigFromPath(configPath)
}

// LoadConfigFromPath loads the MCP server configuration from the specified file path.
// Returns an error if the configuration file cannot be read or parsed.
func LoadConfigFromPath(configPath string) (*Config, error) {
	// Read configuration file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}

	// Parse JSON configuration
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", configPath, err)
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

// Validate checks that all required configuration fields are present and valid.
func (c *Config) Validate() error {
	// Check required fields
	if c.BackendAPIURL == "" {
		return fmt.Errorf("backend_api_url is required")
	}

	if c.PATToken == "" {
		return fmt.Errorf("pat_token is required")
	}

	// Validate URL format
	if _, err := url.Parse(c.BackendAPIURL); err != nil {
		return fmt.Errorf("invalid backend_api_url: %w", err)
	}

	// Validate timeout format if provided
	if c.RequestTimeout != "" {
		if _, err := time.ParseDuration(c.RequestTimeout); err != nil {
			return fmt.Errorf("invalid request_timeout format: %w", err)
		}
	}

	// Set defaults for optional fields
	if c.RequestTimeout == "" {
		c.RequestTimeout = "30s"
	}

	if c.LogLevel == "" {
		c.LogLevel = "info"
	}

	return nil
}

// GetRequestTimeout returns the request timeout as a time.Duration.
// Returns 30 seconds if not configured or invalid.
func (c *Config) GetRequestTimeout() time.Duration {
	if c.RequestTimeout == "" {
		return 30 * time.Second
	}

	duration, err := time.ParseDuration(c.RequestTimeout)
	if err != nil {
		return 30 * time.Second
	}

	return duration
}

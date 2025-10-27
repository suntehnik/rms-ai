package init

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"product-requirements-management/internal/mcp"
)

// ConfigGenerator handles the generation of MCP server configuration files
// during the initialization process.
type ConfigGenerator struct{}

// GeneratedConfig represents the configuration structure that will be
// written to the config file. It matches the existing mcp.Config structure
// to ensure compatibility.
type GeneratedConfig struct {
	BackendAPIURL  string `json:"backend_api_url"`
	PATToken       string `json:"pat_token"`
	RequestTimeout string `json:"request_timeout"`
	LogLevel       string `json:"log_level"`
}

// NewConfigGenerator creates a new ConfigGenerator instance.
func NewConfigGenerator() *ConfigGenerator {
	return &ConfigGenerator{}
}

// GenerateConfig creates a new configuration structure with the provided
// API URL and PAT token, along with sensible defaults for other fields.
// The generated configuration is compatible with the existing mcp.Config structure.
func (g *ConfigGenerator) GenerateConfig(apiURL, patToken string) *GeneratedConfig {
	return &GeneratedConfig{
		BackendAPIURL:  apiURL,
		PATToken:       patToken,
		RequestTimeout: "30s",  // Default timeout matching existing validation
		LogLevel:       "info", // Default log level matching existing validation
	}
}

// ValidateConfig ensures the generated configuration is valid and compatible
// with the existing configuration system. It performs the same validation
// as the existing mcp.Config.Validate() method.
func (g *ConfigGenerator) ValidateConfig(config *GeneratedConfig) error {
	// Check required fields
	if config.BackendAPIURL == "" {
		return fmt.Errorf("backend_api_url is required")
	}

	if config.PATToken == "" {
		return fmt.Errorf("pat_token is required")
	}

	// Validate URL format
	if _, err := url.Parse(config.BackendAPIURL); err != nil {
		return fmt.Errorf("invalid backend_api_url: %w", err)
	}

	// Validate timeout format if provided
	if config.RequestTimeout != "" {
		if _, err := time.ParseDuration(config.RequestTimeout); err != nil {
			return fmt.Errorf("invalid request_timeout format: %w", err)
		}
	}

	// Validate log level
	validLogLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if config.LogLevel != "" && !validLogLevels[config.LogLevel] {
		return fmt.Errorf("invalid log_level: must be one of debug, info, warn, error")
	}

	return nil
}

// ToJSON converts the generated configuration to a JSON byte array
// with proper indentation for human readability.
func (g *ConfigGenerator) ToJSON(config *GeneratedConfig) ([]byte, error) {
	return json.MarshalIndent(config, "", "  ")
}

// TestCompatibility verifies that the generated configuration is compatible
// with the existing mcp.Config system by attempting to convert and validate it.
func (g *ConfigGenerator) TestCompatibility(config *GeneratedConfig) error {
	// Convert to JSON and back to test serialization
	jsonData, err := g.ToJSON(config)
	if err != nil {
		return fmt.Errorf("failed to serialize config to JSON: %w", err)
	}

	// Test compatibility with existing Config struct
	var mcpConfig mcp.Config
	if err := json.Unmarshal(jsonData, &mcpConfig); err != nil {
		return fmt.Errorf("generated config is not compatible with mcp.Config: %w", err)
	}

	// Test existing validation logic
	if err := mcpConfig.Validate(); err != nil {
		return fmt.Errorf("generated config fails existing validation: %w", err)
	}

	return nil
}

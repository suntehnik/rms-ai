package mcp

import (
	"context"
)

// CapabilitiesManager manages server capabilities for MCP initialization
type CapabilitiesManager struct {
	toolProvider   ToolProvider
	promptProvider PromptProvider
}

// NewCapabilitiesManager creates a new capabilities manager
func NewCapabilitiesManager(toolProvider ToolProvider, promptProvider PromptProvider) *CapabilitiesManager {
	return &CapabilitiesManager{
		toolProvider:   toolProvider,
		promptProvider: promptProvider,
	}
}

// ServerCapabilities represents the capabilities provided by the server
type ServerCapabilities struct {
	// Logging   LoggingCapability   `json:"logging"`
	Prompts   PromptsCapability   `json:"prompts"`
	Resources ResourcesCapability `json:"resources"`
	Tools     ToolsCapability     `json:"tools"`
}

// LoggingCapability represents server's logging capability
type LoggingCapability struct{}

// PromptsCapability represents server's prompts capability
type PromptsCapability struct {
	ListChanged bool `json:"listChanged"`
}

// ResourcesCapability represents server's resources capability
type ResourcesCapability struct {
	ListChanged bool `json:"listChanged"`
	Subscribe   bool `json:"subscribe,omitempty"`
}

// ToolsCapability represents server's tools capability
type ToolsCapability struct {
	ListChanged bool `json:"listChanged"`
}

// ToolProvider interface for tool capability detection
type ToolProvider interface {
	HasTools(ctx context.Context) bool
	SupportsListChanged(ctx context.Context) bool
}

// PromptProvider interface for prompt capability detection
type PromptProvider interface {
	HasPrompts(ctx context.Context) bool
	SupportsListChanged(ctx context.Context) bool
}

// GenerateCapabilities generates server capabilities based on current system state
func (cm *CapabilitiesManager) GenerateCapabilities(ctx context.Context) (*ServerCapabilities, error) {
	capabilities := &ServerCapabilities{
		// Logging capability is temporary not available because should implement logging/setLevel
		// Logging: LoggingCapability{},
	}

	// Configure prompts capability
	if cm.promptProvider != nil && cm.promptProvider.HasPrompts(ctx) {
		capabilities.Prompts = PromptsCapability{
			ListChanged: cm.promptProvider.SupportsListChanged(ctx),
		}
	} else {
		// Default prompts capability even if no provider
		capabilities.Prompts = PromptsCapability{
			ListChanged: true,
		}
	}

	// Configure resources capability
	capabilities.Resources = ResourcesCapability{
		ListChanged: true,
		Subscribe:   true, // Optional subscribe property for resources
	}

	// Configure tools capability
	if cm.toolProvider != nil && cm.toolProvider.HasTools(ctx) {
		capabilities.Tools = ToolsCapability{
			ListChanged: cm.toolProvider.SupportsListChanged(ctx),
		}
	} else {
		// Default tools capability even if no provider
		capabilities.Tools = ToolsCapability{
			ListChanged: true,
		}
	}

	return capabilities, nil
}

// UpdateCapabilities updates capabilities based on server state changes
func (cm *CapabilitiesManager) UpdateCapabilities(ctx context.Context) (*ServerCapabilities, error) {
	// For now, capabilities are static, but this method allows for dynamic updates
	return cm.GenerateCapabilities(ctx)
}

// HasCapability checks if a specific capability is supported
func (cm *CapabilitiesManager) HasCapability(ctx context.Context, capability string) bool {
	switch capability {
	case "logging":
		return false // Logging capability is not supported
	case "prompts":
		return cm.promptProvider == nil || cm.promptProvider.HasPrompts(ctx)
	case "resources":
		return true
	case "tools":
		return cm.toolProvider == nil || cm.toolProvider.HasTools(ctx)
	default:
		return false
	}
}

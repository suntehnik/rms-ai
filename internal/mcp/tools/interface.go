package tools

import "context"

// ToolHandler defines the interface for domain-specific MCP tool handlers
type ToolHandler interface {
	// GetSupportedTools returns the list of tools this handler supports
	GetSupportedTools() []string

	// HandleTool processes a specific tool call for this domain
	HandleTool(ctx context.Context, toolName string, args map[string]interface{}) (interface{}, error)
}

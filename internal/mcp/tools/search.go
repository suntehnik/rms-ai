package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"product-requirements-management/internal/jsonrpc"
	"product-requirements-management/internal/mcp/types"
	"product-requirements-management/internal/service"
)

// SearchHandler handles MCP search tool operations
type SearchHandler struct {
	searchService      service.SearchServiceInterface
	requirementService service.RequirementService
}

// NewSearchHandler creates a new search handler with required dependencies
func NewSearchHandler(
	searchService service.SearchServiceInterface,
	requirementService service.RequirementService,
) *SearchHandler {
	return &SearchHandler{
		searchService:      searchService,
		requirementService: requirementService,
	}
}

// GetSupportedTools returns the list of tools this handler supports
func (h *SearchHandler) GetSupportedTools() []string {
	return []string{
		"search_global",
		"search_requirements",
	}
}

// HandleTool processes a specific tool call for this domain
func (h *SearchHandler) HandleTool(ctx context.Context, toolName string, args map[string]interface{}) (interface{}, error) {
	switch toolName {
	case "search_global":
		return h.Global(ctx, args)
	case "search_requirements":
		return h.Requirements(ctx, args)
	default:
		return nil, jsonrpc.NewMethodNotFoundError(fmt.Sprintf("Unknown search tool: %s", toolName))
	}
}

// Global handles the search_global tool
func (h *SearchHandler) Global(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	// Validate required arguments
	query, ok := args["query"].(string)
	if !ok || query == "" {
		return nil, jsonrpc.NewInvalidParamsError("Missing or invalid 'query' argument")
	}

	// Optional arguments
	var entityTypes []string
	if entityTypesInterface, ok := args["entity_types"].([]interface{}); ok {
		for _, et := range entityTypesInterface {
			if etStr, ok := et.(string); ok {
				entityTypes = append(entityTypes, etStr)
			}
		}
	}

	limit := 50 // Default limit
	if limitFloat, ok := args["limit"].(float64); ok {
		limit = int(limitFloat)
	}

	offset := 0 // Default offset
	if offsetFloat, ok := args["offset"].(float64); ok {
		offset = int(offsetFloat)
	}

	// Create search options with entity types
	searchOptions := service.SearchOptions{
		Query:       query,
		EntityTypes: entityTypes,
		Limit:       limit,
		Offset:      offset,
	}

	// Perform the search using the search service
	response, err := h.searchService.Search(ctx, searchOptions)
	if err != nil {
		return nil, jsonrpc.NewInternalError(fmt.Sprintf("Search failed: %v", err))
	}

	// Convert search results to JSON string for MCP compatibility
	searchData := map[string]interface{}{
		"results":      response.Results,
		"total_count":  response.Total,
		"query":        query,
		"entity_types": entityTypes,
		"limit":        limit,
		"offset":       offset,
	}

	jsonData, err := json.MarshalIndent(searchData, "", "  ")
	if err != nil {
		return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to marshal search results: %v", err))
	}

	return &types.ToolResponse{
		Content: []types.ContentItem{
			{
				Type: "text",
				Text: fmt.Sprintf("Found %d results for query '%s'", response.Total, query),
			},
			{
				Type: "text",
				Text: string(jsonData),
			},
		},
	}, nil
}

// Requirements handles the search_requirements tool
func (h *SearchHandler) Requirements(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	// Validate required arguments
	query, ok := args["query"].(string)
	if !ok || query == "" {
		return nil, jsonrpc.NewInvalidParamsError("Missing or invalid 'query' argument")
	}

	// Perform the search using the requirement service
	requirements, err := h.requirementService.SearchRequirements(query)
	if err != nil {
		return nil, jsonrpc.NewInternalError(fmt.Sprintf("Requirements search failed: %v", err))
	}

	// Convert requirements search results to JSON string for MCP compatibility
	searchData := map[string]interface{}{
		"requirements": requirements,
		"query":        query,
		"count":        len(requirements),
	}

	jsonData, err := json.MarshalIndent(searchData, "", "  ")
	if err != nil {
		return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to marshal requirements search results: %v", err))
	}

	return &types.ToolResponse{
		Content: []types.ContentItem{
			{
				Type: "text",
				Text: fmt.Sprintf("Found %d requirements matching query '%s'", len(requirements), query),
			},
			{
				Type: "text",
				Text: string(jsonData),
			},
		},
	}, nil
}

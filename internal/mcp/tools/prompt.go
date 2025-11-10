package tools

import (
	"context"
	"fmt"

	"product-requirements-management/internal/jsonrpc"
	"product-requirements-management/internal/mcp/types"
	"product-requirements-management/internal/models"
	"product-requirements-management/internal/service"

	"github.com/google/uuid"
)

// PromptServiceInterface defines the interface for prompt service operations
type PromptServiceInterface interface {
	Create(ctx context.Context, req *service.CreatePromptRequest, creatorID uuid.UUID) (*models.Prompt, error)
	Update(ctx context.Context, id uuid.UUID, req *service.UpdatePromptRequest) (*models.Prompt, error)
	Delete(ctx context.Context, id uuid.UUID) error
	Activate(ctx context.Context, id uuid.UUID) error
	GetByReferenceID(ctx context.Context, referenceID string) (*models.Prompt, error)
	GetActive(ctx context.Context) (*models.Prompt, error)
	List(ctx context.Context, limit, offset int, creatorID *uuid.UUID) ([]*models.Prompt, int64, error)
}

// PromptHandler handles MCP tools for Prompt domain operations
type PromptHandler struct {
	promptService PromptServiceInterface
}

// NewPromptHandler creates a new Prompt handler instance
func NewPromptHandler(promptService PromptServiceInterface) *PromptHandler {
	return &PromptHandler{
		promptService: promptService,
	}
}

// GetSupportedTools returns the list of tools this handler supports
func (h *PromptHandler) GetSupportedTools() []string {
	return []string{
		ToolCreatePrompt,
		ToolUpdatePrompt,
		ToolDeletePrompt,
		ToolActivatePrompt,
		ToolListPrompts,
		ToolGetActivePrompt,
	}
}

// HandleTool processes a specific tool call for Prompt domain operations
func (h *PromptHandler) HandleTool(ctx context.Context, toolName string, args map[string]interface{}) (interface{}, error) {
	switch toolName {
	case ToolCreatePrompt:
		return h.Create(ctx, args)
	case ToolUpdatePrompt:
		return h.Update(ctx, args)
	case ToolDeletePrompt:
		return h.Delete(ctx, args)
	case ToolActivatePrompt:
		return h.Activate(ctx, args)
	case ToolListPrompts:
		return h.List(ctx, args)
	case ToolGetActivePrompt:
		return h.GetActive(ctx, args)
	default:
		return nil, jsonrpc.NewMethodNotFoundError(fmt.Sprintf("Unknown Prompt tool: %s", toolName))
	}
}

// Create handles the create_prompt tool
func (h *PromptHandler) Create(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	// Get current user from context
	user, err := getUserFromContext(ctx)
	if err != nil {
		return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to get user from context: %v", err))
	}

	// Check if user has Administrator role
	if user.Role != models.RoleAdministrator {
		return nil, jsonrpc.NewJSONRPCError(-32002, "Insufficient permissions: Administrator role required", nil)
	}

	// Validate required arguments
	name, ok := getStringArg(args, "name")
	if !ok || name == "" {
		return nil, jsonrpc.NewInvalidParamsError("Missing or invalid 'name' argument")
	}

	title, ok := getStringArg(args, "title")
	if !ok || title == "" {
		return nil, jsonrpc.NewInvalidParamsError("Missing or invalid 'title' argument")
	}

	content, ok := getStringArg(args, "content")
	if !ok || content == "" {
		return nil, jsonrpc.NewInvalidParamsError("Missing or invalid 'content' argument")
	}

	// Optional arguments
	var description *string
	if desc, ok := getStringArg(args, "description"); ok && desc != "" {
		description = &desc
	}

	var role *models.MCPRole
	if roleStr, ok := getStringArg(args, "role"); ok && roleStr != "" {
		if err := models.ValidateMCPRole(roleStr); err != nil {
			return nil, jsonrpc.NewInvalidParamsError(fmt.Sprintf("Invalid role: %v", err))
		}
		mcpRole := models.MCPRole(roleStr)
		role = &mcpRole
	}

	// Create the prompt
	req := service.CreatePromptRequest{
		Name:        name,
		Title:       title,
		Description: description,
		Content:     content,
		Role:        role,
	}

	prompt, err := h.promptService.Create(ctx, &req, user.ID)
	if err != nil {
		if err == service.ErrDuplicateEntry {
			return nil, jsonrpc.NewJSONRPCError(-32003, "Prompt with this name already exists", nil)
		}
		return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to create prompt: %v", err))
	}

	message := fmt.Sprintf("Successfully created prompt %s: %s", prompt.ReferenceID, prompt.Title)
	return types.CreateDataResponse(message, map[string]interface{}{
		"id":           prompt.ID,
		"reference_id": prompt.ReferenceID,
		"name":         prompt.Name,
		"title":        prompt.Title,
	}), nil
}

// Update handles the update_prompt tool
func (h *PromptHandler) Update(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	// Get current user from context
	user, err := getUserFromContext(ctx)
	if err != nil {
		return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to get user from context: %v", err))
	}

	// Check if user has Administrator role
	if user.Role != models.RoleAdministrator {
		return nil, jsonrpc.NewJSONRPCError(-32002, "Insufficient permissions: Administrator role required", nil)
	}

	// Validate required arguments
	promptIDStr, ok := getStringArg(args, "prompt_id")
	if !ok || promptIDStr == "" {
		return nil, jsonrpc.NewInvalidParamsError("Missing or invalid 'prompt_id' argument")
	}

	var promptID uuid.UUID
	if id, parseErr := uuid.Parse(promptIDStr); parseErr == nil {
		promptID = id
	} else {
		// Try as reference ID
		prompt, getErr := h.promptService.GetByReferenceID(ctx, promptIDStr)
		if getErr != nil {
			if getErr == service.ErrNotFound {
				return nil, jsonrpc.NewJSONRPCError(-32002, "Prompt not found", nil)
			}
			return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to get prompt: %v", getErr))
		}
		promptID = prompt.ID
	}

	// Build update request from optional arguments
	req := service.UpdatePromptRequest{}
	if title, ok := getStringArg(args, "title"); ok && title != "" {
		req.Title = &title
	}
	if description, ok := getStringArg(args, "description"); ok {
		req.Description = &description
	}
	if content, ok := getStringArg(args, "content"); ok && content != "" {
		req.Content = &content
	}
	if roleStr, ok := getStringArg(args, "role"); ok && roleStr != "" {
		if err := models.ValidateMCPRole(roleStr); err != nil {
			return nil, jsonrpc.NewInvalidParamsError(fmt.Sprintf("Invalid role: %v", err))
		}
		mcpRole := models.MCPRole(roleStr)
		req.Role = &mcpRole
	}

	prompt, err := h.promptService.Update(ctx, promptID, &req)
	if err != nil {
		if err == service.ErrNotFound {
			return nil, jsonrpc.NewJSONRPCError(-32002, "Prompt not found", nil)
		}
		return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to update prompt: %v", err))
	}

	message := fmt.Sprintf("Successfully updated prompt %s: %s", prompt.ReferenceID, prompt.Title)
	return types.CreateSuccessResponse(message), nil
}

// Delete handles the delete_prompt tool
func (h *PromptHandler) Delete(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	// Get current user from context
	user, err := getUserFromContext(ctx)
	if err != nil {
		return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to get user from context: %v", err))
	}

	// Check if user has Administrator role
	if user.Role != models.RoleAdministrator {
		return nil, jsonrpc.NewJSONRPCError(-32002, "Insufficient permissions: Administrator role required", nil)
	}

	// Validate required arguments
	promptIDStr, ok := getStringArg(args, "prompt_id")
	if !ok || promptIDStr == "" {
		return nil, jsonrpc.NewInvalidParamsError("Missing or invalid 'prompt_id' argument")
	}

	var promptID uuid.UUID
	if id, parseErr := uuid.Parse(promptIDStr); parseErr == nil {
		promptID = id
	} else {
		// Try as reference ID
		prompt, getErr := h.promptService.GetByReferenceID(ctx, promptIDStr)
		if getErr != nil {
			if getErr == service.ErrNotFound {
				return nil, jsonrpc.NewJSONRPCError(-32002, "Prompt not found", nil)
			}
			return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to get prompt: %v", getErr))
		}
		promptID = prompt.ID
	}

	err = h.promptService.Delete(ctx, promptID)
	if err != nil {
		if err == service.ErrNotFound {
			return nil, jsonrpc.NewJSONRPCError(-32002, "Prompt not found", nil)
		}
		return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to delete prompt: %v", err))
	}

	message := fmt.Sprintf("Successfully deleted prompt %s", promptIDStr)
	return types.CreateSuccessResponse(message), nil
}

// Activate handles the activate_prompt tool
func (h *PromptHandler) Activate(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	// Get current user from context
	user, err := getUserFromContext(ctx)
	if err != nil {
		return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to get user from context: %v", err))
	}

	// Check if user has Administrator role
	if user.Role != models.RoleAdministrator {
		return nil, jsonrpc.NewJSONRPCError(-32002, "Insufficient permissions: Administrator role required", nil)
	}

	// Validate required arguments
	promptIDStr, ok := getStringArg(args, "prompt_id")
	if !ok || promptIDStr == "" {
		return nil, jsonrpc.NewInvalidParamsError("Missing or invalid 'prompt_id' argument")
	}

	var promptID uuid.UUID
	if id, parseErr := uuid.Parse(promptIDStr); parseErr == nil {
		promptID = id
	} else {
		// Try as reference ID
		prompt, getErr := h.promptService.GetByReferenceID(ctx, promptIDStr)
		if getErr != nil {
			if getErr == service.ErrNotFound {
				return nil, jsonrpc.NewJSONRPCError(-32002, "Prompt not found", nil)
			}
			return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to get prompt: %v", getErr))
		}
		promptID = prompt.ID
	}

	err = h.promptService.Activate(ctx, promptID)
	if err != nil {
		if err == service.ErrNotFound {
			return nil, jsonrpc.NewJSONRPCError(-32002, "Prompt not found", nil)
		}
		return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to activate prompt: %v", err))
	}

	message := fmt.Sprintf("Successfully activated prompt %s", promptIDStr)
	return types.CreateSuccessResponse(message), nil
}

// List handles the list_prompts tool
func (h *PromptHandler) List(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	// Parse optional pagination parameters
	limit := 50
	if l, ok := getIntArg(args, "limit"); ok && l > 0 && l <= 100 {
		limit = l
	}

	offset := 0
	if o, ok := getIntArg(args, "offset"); ok && o >= 0 {
		offset = o
	}

	// Parse optional creator filter
	var creatorID *uuid.UUID
	if creatorIDStr, ok := getStringArg(args, "creator_id"); ok && creatorIDStr != "" {
		if parsed, parseErr := uuid.Parse(creatorIDStr); parseErr == nil {
			creatorID = &parsed
		}
	}

	prompts, total, err := h.promptService.List(ctx, limit, offset, creatorID)
	if err != nil {
		return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to list prompts: %v", err))
	}

	promptsData := map[string]interface{}{
		"prompts":     prompts,
		"total_count": total,
		"limit":       limit,
		"offset":      offset,
	}

	message := fmt.Sprintf("Found %d prompts (total: %d)", len(prompts), total)
	return types.CreateDataResponse(message, promptsData), nil
}

// GetActive handles the get_active_prompt tool
func (h *PromptHandler) GetActive(ctx context.Context, _ map[string]interface{}) (interface{}, error) {
	prompt, err := h.promptService.GetActive(ctx)
	if err != nil {
		if err == service.ErrNotFound {
			return nil, jsonrpc.NewJSONRPCError(-32002, "No active prompt found", nil)
		}
		return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to get active prompt: %v", err))
	}

	message := fmt.Sprintf("Active prompt: %s (%s)", prompt.Title, prompt.ReferenceID)
	return types.CreateDataResponse(message, prompt), nil
}

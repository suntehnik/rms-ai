package tools

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"product-requirements-management/internal/jsonrpc"
	"product-requirements-management/internal/mcp/types"
	"product-requirements-management/internal/models"
	"product-requirements-management/internal/service"
)

// SteeringDocumentHandler handles MCP tools for steering document operations
type SteeringDocumentHandler struct {
	steeringDocumentService service.SteeringDocumentService
	epicService             service.EpicService
}

// NewSteeringDocumentHandler creates a new steering document handler instance
func NewSteeringDocumentHandler(
	steeringDocumentService service.SteeringDocumentService,
	epicService service.EpicService,
) *SteeringDocumentHandler {
	return &SteeringDocumentHandler{
		steeringDocumentService: steeringDocumentService,
		epicService:             epicService,
	}
}

// GetSupportedTools returns the list of tools this handler supports
func (h *SteeringDocumentHandler) GetSupportedTools() []string {
	return []string{
		"list_steering_documents",
		"create_steering_document",
		"get_steering_document",
		"update_steering_document",
		"link_steering_to_epic",
		"unlink_steering_from_epic",
		"get_epic_steering_documents",
	}
}

// HandleTool processes a specific tool call for steering document operations
func (h *SteeringDocumentHandler) HandleTool(ctx context.Context, toolName string, args map[string]interface{}) (interface{}, error) {
	switch toolName {
	case "list_steering_documents":
		return h.ListSteeringDocuments(ctx, args)
	case "create_steering_document":
		return h.CreateSteeringDocument(ctx, args)
	case "get_steering_document":
		return h.GetSteeringDocument(ctx, args)
	case "update_steering_document":
		return h.UpdateSteeringDocument(ctx, args)
	case "link_steering_to_epic":
		return h.LinkSteeringToEpic(ctx, args)
	case "unlink_steering_from_epic":
		return h.UnlinkSteeringFromEpic(ctx, args)
	case "get_epic_steering_documents":
		return h.GetEpicSteeringDocuments(ctx, args)
	default:
		return nil, jsonrpc.NewMethodNotFoundError(fmt.Sprintf("Unknown steering document tool: %s", toolName))
	}
}

// ListSteeringDocuments handles the list_steering_documents tool
func (h *SteeringDocumentHandler) ListSteeringDocuments(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	// Get current user from context
	user, err := getUserFromContext(ctx)
	if err != nil {
		return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to get user from context: %v", err))
	}

	// Build filters from optional arguments
	filters := service.SteeringDocumentFilters{}

	if creatorIDStr, ok := getStringArg(args, "creator_id"); ok && creatorIDStr != "" {
		creatorID, err := uuid.Parse(creatorIDStr)
		if err != nil {
			return nil, jsonrpc.NewInvalidParamsError("Invalid 'creator_id' format")
		}
		filters.CreatorID = &creatorID
	}

	if search, ok := getStringArg(args, "search"); ok {
		filters.Search = search
	}

	if orderBy, ok := getStringArg(args, "order_by"); ok {
		filters.OrderBy = orderBy
	}

	if limit, ok := getIntArg(args, "limit"); ok {
		filters.Limit = limit
	}

	if offset, ok := getIntArg(args, "offset"); ok {
		filters.Offset = offset
	}

	// List steering documents
	docs, total, err := h.steeringDocumentService.ListSteeringDocuments(filters, user)
	if err != nil {
		return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to list steering documents: %v", err))
	}

	// Convert results to JSON string for MCP compatibility
	listData := map[string]interface{}{
		"steering_documents": docs,
		"total_count":        total,
		"limit":              filters.Limit,
		"offset":             filters.Offset,
	}

	return types.CreateDataResponse(
		fmt.Sprintf("Found %d steering documents (total: %d)", len(docs), total),
		listData,
	), nil
}

// CreateSteeringDocument handles the create_steering_document tool
func (h *SteeringDocumentHandler) CreateSteeringDocument(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	// Get current user from context
	user, err := getUserFromContext(ctx)
	if err != nil {
		return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to get user from context: %v", err))
	}

	// Validate required arguments
	title, ok := getStringArg(args, "title")
	if !ok || title == "" {
		return nil, jsonrpc.NewInvalidParamsError("Missing or invalid 'title' argument")
	}

	// Optional arguments
	var description *string
	if desc, ok := getStringArg(args, "description"); ok && desc != "" {
		description = &desc
	}

	var epicID *string
	if epicIDStr, ok := getStringArg(args, "epic_id"); ok && epicIDStr != "" {
		epicID = &epicIDStr
	}

	// Create the steering document
	req := service.CreateSteeringDocumentRequest{
		Title:       title,
		Description: description,
		EpicID:      epicID,
	}

	doc, err := h.steeringDocumentService.CreateSteeringDocument(req, user)
	if err != nil {
		return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to create steering document: %v", err))
	}

	// Prepare success message
	successMsg := fmt.Sprintf("Successfully created steering document %s: %s", doc.ReferenceID, doc.Title)
	if epicID != nil {
		successMsg += " and linked to epic"
	}

	return types.CreateDataResponse(successMsg, doc), nil
}

// GetSteeringDocument handles the get_steering_document tool
func (h *SteeringDocumentHandler) GetSteeringDocument(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	// Get current user from context
	user, err := getUserFromContext(ctx)
	if err != nil {
		return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to get user from context: %v", err))
	}

	// Validate required arguments
	steeringDocIDStr, ok := getStringArg(args, "steering_document_id")
	if !ok || steeringDocIDStr == "" {
		return nil, jsonrpc.NewInvalidParamsError("Missing or invalid 'steering_document_id' argument")
	}

	var doc *models.SteeringDocument

	// Try to parse as UUID first, then as reference ID
	if steeringDocID, err := uuid.Parse(steeringDocIDStr); err == nil {
		doc, err = h.steeringDocumentService.GetSteeringDocumentByID(steeringDocID, user)
		if err != nil {
			return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to get steering document by ID: %v", err))
		}
	} else {
		// Try to get by reference ID
		doc, err = h.steeringDocumentService.GetSteeringDocumentByReferenceID(steeringDocIDStr, user)
		if err != nil {
			return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to get steering document by reference ID: %v", err))
		}
	}

	return types.CreateDataResponse(
		fmt.Sprintf("Retrieved steering document %s: %s", doc.ReferenceID, doc.Title),
		doc,
	), nil
}

// UpdateSteeringDocument handles the update_steering_document tool
func (h *SteeringDocumentHandler) UpdateSteeringDocument(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	// Get current user from context
	user, err := getUserFromContext(ctx)
	if err != nil {
		return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to get user from context: %v", err))
	}

	// Validate required arguments
	steeringDocIDStr, ok := getStringArg(args, "steering_document_id")
	if !ok || steeringDocIDStr == "" {
		return nil, jsonrpc.NewInvalidParamsError("Missing or invalid 'steering_document_id' argument")
	}

	var steeringDocID uuid.UUID

	// Try to parse as UUID first, then as reference ID
	if parsedID, err := uuid.Parse(steeringDocIDStr); err == nil {
		steeringDocID = parsedID
	} else {
		// Try to get by reference ID
		doc, err := h.steeringDocumentService.GetSteeringDocumentByReferenceID(steeringDocIDStr, user)
		if err != nil {
			return nil, jsonrpc.NewInvalidParamsError("Invalid 'steering_document_id': not a valid UUID or reference ID")
		}
		steeringDocID = doc.ID
	}

	// Build update request from optional arguments
	req := service.UpdateSteeringDocumentRequest{}

	if title, ok := getStringArg(args, "title"); ok && title != "" {
		req.Title = &title
	}

	if description, ok := getStringArg(args, "description"); ok {
		req.Description = &description
	}

	// Update the steering document
	doc, err := h.steeringDocumentService.UpdateSteeringDocument(steeringDocID, req, user)
	if err != nil {
		return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to update steering document: %v", err))
	}

	return types.CreateDataResponse(
		fmt.Sprintf("Successfully updated steering document %s: %s", doc.ReferenceID, doc.Title),
		doc,
	), nil
}

// LinkSteeringToEpic handles the link_steering_to_epic tool
func (h *SteeringDocumentHandler) LinkSteeringToEpic(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	// Get current user from context
	user, err := getUserFromContext(ctx)
	if err != nil {
		return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to get user from context: %v", err))
	}

	// Validate required arguments
	steeringDocIDStr, ok := getStringArg(args, "steering_document_id")
	if !ok || steeringDocIDStr == "" {
		return nil, jsonrpc.NewInvalidParamsError("Missing or invalid 'steering_document_id' argument")
	}

	epicIDStr, ok := getStringArg(args, "epic_id")
	if !ok || epicIDStr == "" {
		return nil, jsonrpc.NewInvalidParamsError("Missing or invalid 'epic_id' argument")
	}

	// Parse steering document ID
	var steeringDocID uuid.UUID
	if parsedID, err := uuid.Parse(steeringDocIDStr); err == nil {
		steeringDocID = parsedID
	} else {
		// Try to get by reference ID
		doc, err := h.steeringDocumentService.GetSteeringDocumentByReferenceID(steeringDocIDStr, user)
		if err != nil {
			return nil, jsonrpc.NewInvalidParamsError("Invalid 'steering_document_id': not a valid UUID or reference ID")
		}
		steeringDocID = doc.ID
	}

	// Parse epic ID
	var epicID uuid.UUID
	if parsedID, err := uuid.Parse(epicIDStr); err == nil {
		epicID = parsedID
	} else {
		// Try to get by reference ID
		epic, err := h.epicService.GetEpicByReferenceID(epicIDStr)
		if err != nil {
			return nil, jsonrpc.NewInvalidParamsError("Invalid 'epic_id': not a valid UUID or reference ID")
		}
		epicID = epic.ID
	}

	// Create the link
	err = h.steeringDocumentService.LinkSteeringDocumentToEpic(steeringDocID, epicID, user)
	if err != nil {
		return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to link steering document to epic: %v", err))
	}

	return types.CreateSuccessResponse(
		fmt.Sprintf("Successfully linked steering document %s to epic %s", steeringDocIDStr, epicIDStr),
	), nil
}

// UnlinkSteeringFromEpic handles the unlink_steering_from_epic tool
func (h *SteeringDocumentHandler) UnlinkSteeringFromEpic(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	// Get current user from context
	user, err := getUserFromContext(ctx)
	if err != nil {
		return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to get user from context: %v", err))
	}

	// Validate required arguments
	steeringDocIDStr, ok := getStringArg(args, "steering_document_id")
	if !ok || steeringDocIDStr == "" {
		return nil, jsonrpc.NewInvalidParamsError("Missing or invalid 'steering_document_id' argument")
	}

	epicIDStr, ok := getStringArg(args, "epic_id")
	if !ok || epicIDStr == "" {
		return nil, jsonrpc.NewInvalidParamsError("Missing or invalid 'epic_id' argument")
	}

	// Parse steering document ID
	var steeringDocID uuid.UUID
	if parsedID, err := uuid.Parse(steeringDocIDStr); err == nil {
		steeringDocID = parsedID
	} else {
		// Try to get by reference ID
		doc, err := h.steeringDocumentService.GetSteeringDocumentByReferenceID(steeringDocIDStr, user)
		if err != nil {
			return nil, jsonrpc.NewInvalidParamsError("Invalid 'steering_document_id': not a valid UUID or reference ID")
		}
		steeringDocID = doc.ID
	}

	// Parse epic ID
	var epicID uuid.UUID
	if parsedID, err := uuid.Parse(epicIDStr); err == nil {
		epicID = parsedID
	} else {
		// Try to get by reference ID
		epic, err := h.epicService.GetEpicByReferenceID(epicIDStr)
		if err != nil {
			return nil, jsonrpc.NewInvalidParamsError("Invalid 'epic_id': not a valid UUID or reference ID")
		}
		epicID = epic.ID
	}

	// Remove the link
	err = h.steeringDocumentService.UnlinkSteeringDocumentFromEpic(steeringDocID, epicID, user)
	if err != nil {
		return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to unlink steering document from epic: %v", err))
	}

	return types.CreateSuccessResponse(
		fmt.Sprintf("Successfully unlinked steering document %s from epic %s", steeringDocIDStr, epicIDStr),
	), nil
}

// GetEpicSteeringDocuments handles the get_epic_steering_documents tool
func (h *SteeringDocumentHandler) GetEpicSteeringDocuments(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	// Get current user from context
	user, err := getUserFromContext(ctx)
	if err != nil {
		return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to get user from context: %v", err))
	}

	// Validate required arguments
	epicIDStr, ok := getStringArg(args, "epic_id")
	if !ok || epicIDStr == "" {
		return nil, jsonrpc.NewInvalidParamsError("Missing or invalid 'epic_id' argument")
	}

	var epicID uuid.UUID

	// Try to parse as UUID first, then as reference ID
	if parsedID, err := uuid.Parse(epicIDStr); err == nil {
		epicID = parsedID
	} else {
		// Try to get by reference ID
		epic, err := h.epicService.GetEpicByReferenceID(epicIDStr)
		if err != nil {
			return nil, jsonrpc.NewInvalidParamsError("Invalid 'epic_id': not a valid UUID or reference ID")
		}
		epicID = epic.ID
	}

	// Get steering documents for the epic
	docs, err := h.steeringDocumentService.GetSteeringDocumentsByEpicID(epicID, user)
	if err != nil {
		return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to get steering documents for epic: %v", err))
	}

	// Convert results to JSON string for MCP compatibility
	epicData := map[string]interface{}{
		"epic_id":            epicIDStr,
		"steering_documents": docs,
		"count":              len(docs),
	}

	return types.CreateDataResponse(
		fmt.Sprintf("Found %d steering documents linked to epic %s", len(docs), epicIDStr),
		epicData,
	), nil
}

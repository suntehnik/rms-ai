package tools

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"product-requirements-management/internal/auth"
	"product-requirements-management/internal/jsonrpc"
	"product-requirements-management/internal/models"
)

// getUserFromContext extracts user information from the context
func getUserFromContext(ctx context.Context) (*models.User, error) {
	// Extract Gin context from the context
	ginCtx, ok := ctx.Value("gin_context").(*gin.Context)
	if !ok {
		return nil, fmt.Errorf("gin context not found")
	}

	// Get user from Gin context using auth helper
	user, ok := auth.GetUserFromContext(ginCtx)
	if !ok {
		return nil, fmt.Errorf("user not found in context")
	}

	return user, nil
}

// parseUUIDOrReferenceID attempts to parse an ID string as UUID first, then uses a reference ID lookup function
func parseUUIDOrReferenceID(idStr string, getByRefFunc func(string) (interface{}, error)) (uuid.UUID, error) {
	// Try to parse as UUID first
	if parsedUUID, err := uuid.Parse(idStr); err == nil {
		return parsedUUID, nil
	}

	// Try to get by reference ID
	entity, err := getByRefFunc(idStr)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid ID: not a valid UUID or reference ID")
	}

	// Extract UUID from the entity (assuming it has an ID field)
	switch e := entity.(type) {
	case *models.Epic:
		return e.ID, nil
	case *models.UserStory:
		return e.ID, nil
	case *models.Requirement:
		return e.ID, nil
	case *models.SteeringDocument:
		return e.ID, nil
	default:
		return uuid.Nil, fmt.Errorf("unsupported entity type for ID extraction")
	}
}

// validateRequiredArgs checks that all required arguments are present in the args map
func validateRequiredArgs(args map[string]interface{}, required []string) error {
	for _, arg := range required {
		if _, exists := args[arg]; !exists {
			return jsonrpc.NewInvalidParamsError(fmt.Sprintf("Missing required argument: %s", arg))
		}
	}
	return nil
}

// getStringArg safely extracts a string argument from the args map
func getStringArg(args map[string]interface{}, key string) (string, bool) {
	if val, exists := args[key]; exists {
		if str, ok := val.(string); ok {
			return str, true
		}
	}
	return "", false
}

// getIntArg safely extracts an integer argument from the args map
func getIntArg(args map[string]interface{}, key string) (int, bool) {
	if val, exists := args[key]; exists {
		switch v := val.(type) {
		case int:
			return v, true
		case float64:
			return int(v), true
		}
	}
	return 0, false
}

// getUUIDArg safely extracts and parses a UUID argument from the args map
func getUUIDArg(args map[string]interface{}, key string) (uuid.UUID, bool) {
	if str, exists := getStringArg(args, key); exists {
		if parsedUUID, err := uuid.Parse(str); err == nil {
			return parsedUUID, true
		}
	}
	return uuid.Nil, false
}

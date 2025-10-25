package handlers

import (
	"context"
	"errors"
	"product-requirements-management/internal/models"

	"github.com/sirupsen/logrus"

	"product-requirements-management/internal/service"
)

// PromptsHandler handles MCP prompts protocol methods
type PromptsHandler struct {
	promptService             *service.PromptService
	epicService               service.EpicService
	userStoryService          service.UserStoryService
	requirementService        service.RequirementService
	acceptanceCriteriaService service.AcceptanceCriteriaService
	logger                    *logrus.Logger
}

// NewPromptsHandler creates a new PromptsHandler instance
func NewPromptsHandler(
	promptService *service.PromptService,
	epicService service.EpicService,
	userStoryService service.UserStoryService,
	requirementService service.RequirementService,
	acceptanceCriteriaService service.AcceptanceCriteriaService,
	logger *logrus.Logger,
) *PromptsHandler {
	return &PromptsHandler{
		promptService:             promptService,
		epicService:               epicService,
		userStoryService:          userStoryService,
		requirementService:        requirementService,
		acceptanceCriteriaService: acceptanceCriteriaService,
		logger:                    logger,
	}
}

// MCP protocol request/response structures
type PromptListRequest struct {
	// No parameters needed for listing prompts
}

type PromptGetRequest struct {
	Name string `json:"name" validate:"required"`
}

type PromptListResponse struct {
	Prompts []*models.MCPPromptDescriptor `json:"prompts"`
}

type PromptGetResponse struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Messages    []models.PromptMessage `json:"messages"`
}

// HandlePromptsList handles the prompts/list method
func (ph *PromptsHandler) HandlePromptsList(ctx context.Context, params interface{}) (interface{}, error) {
	ph.logger.WithField("method", "prompts/list").Info("Processing prompts list request")

	descriptors, err := ph.promptService.GetMCPPromptDescriptors(ctx)
	if err != nil {
		ph.logger.WithError(err).Error("Failed to get prompt descriptors")
		return nil, err
	}

	response := &PromptListResponse{
		Prompts: descriptors,
	}

	ph.logger.WithField("prompt_count", len(descriptors)).Info("Successfully retrieved prompt descriptors")
	return response, nil
}

// HandlePromptsGet handles the prompts/get method
func (ph *PromptsHandler) HandlePromptsGet(ctx context.Context, params interface{}) (interface{}, error) {
	ph.logger.WithField("method", "prompts/get").Info("Processing prompts get request")

	// Parse parameters
	var req PromptGetRequest
	if paramsMap, ok := params.(map[string]interface{}); ok {
		if name, ok := paramsMap["name"].(string); ok {
			req.Name = name
		} else {
			return nil, errors.New("missing or invalid 'name' parameter")
		}
	} else {
		return nil, errors.New("invalid parameters format")
	}

	// Validate required fields
	if req.Name == "" {
		return nil, errors.New("name parameter is required")
	}

	definition, err := ph.promptService.GetMCPPromptDefinition(ctx, req.Name)
	if err != nil {
		if err == service.ErrNotFound {
			return nil, errors.New("prompt not found")
		}
		ph.logger.WithError(err).WithField("name", req.Name).Error("Failed to get prompt definition")
		return nil, err
	}

	ph.logger.WithField("name", req.Name).Info("Successfully retrieved prompt definition")
	return definition, nil
}

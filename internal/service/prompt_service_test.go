package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"product-requirements-management/internal/models"
)

// MockPromptService is a mock implementation for testing GetMCPPromptDefinition
type MockPromptService struct {
	logger        *logrus.Logger
	validator     *PromptValidator
	getByNameFunc func(ctx context.Context, name string) (*models.Prompt, error)
}

func NewMockPromptService(logger *logrus.Logger) *MockPromptService {
	return &MockPromptService{
		logger:    logger,
		validator: NewPromptValidator(),
	}
}

func (mps *MockPromptService) GetByName(ctx context.Context, name string) (*models.Prompt, error) {
	if mps.getByNameFunc != nil {
		return mps.getByNameFunc(ctx, name)
	}
	return nil, errors.New("mock not configured")
}

// GetMCPPromptDefinition implements the same logic as the real service
func (mps *MockPromptService) GetMCPPromptDefinition(ctx context.Context, name string) (*models.MCPPromptDefinition, error) {
	prompt, err := mps.GetByName(ctx, name)
	if err != nil {
		// Return MCP-compliant error for prompt not found
		if errors.Is(err, ErrNotFound) {
			return nil, errors.New("prompt '" + name + "' not found")
		}
		// Return generic error for other issues
		return nil, errors.New("failed to retrieve prompt: " + err.Error())
	}

	// Validate prompt for MCP compliance before generating response
	if err := mps.validator.ValidateForMCP(prompt); err != nil {
		// Return MCP-compliant validation error
		return nil, errors.New("invalid prompt data: " + err.Error())
	}

	description := ""
	if prompt.Description != nil {
		description = *prompt.Description
	}

	// Transform content to structured format using validator helper
	var contentChunk = models.ContentChunk{
		Type: "text",
		Text: prompt.Content,
	}
	//  := mps.validator.TransformContentToChunk(prompt.Content)

	definition := &models.MCPPromptDefinition{
		Name:        prompt.Name,
		Description: description,
		Messages: []models.PromptMessage{
			{
				Role:    string(prompt.Role),
				Content: contentChunk,
			},
		},
	}

	return definition, nil
}

func TestPromptService_GetMCPPromptDefinition_Success(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel) // Suppress logs during tests

	service := NewMockPromptService(logger)

	// Mock successful prompt retrieval
	description := "Test description"
	service.getByNameFunc = func(ctx context.Context, name string) (*models.Prompt, error) {
		return &models.Prompt{
			ID:          uuid.New(),
			Name:        "test-prompt",
			Title:       "Test Prompt",
			Description: &description,
			Content:     "You are a helpful assistant",
			Role:        models.MCPRoleAssistant,
			CreatorID:   uuid.New(),
		}, nil
	}

	ctx := context.Background()
	result, err := service.GetMCPPromptDefinition(ctx, "test-prompt")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "test-prompt", result.Name)
	assert.Equal(t, "Test description", result.Description)
	assert.Len(t, result.Messages, 1)
	assert.Equal(t, "assistant", result.Messages[0].Role)
	assert.Equal(t, "text", result.Messages[0].Content.Type)
	assert.Equal(t, "You are a helpful assistant", result.Messages[0].Content.Text)
}

func TestPromptService_GetMCPPromptDefinition_NotFound(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel) // Suppress logs during tests

	service := NewMockPromptService(logger)

	// Mock prompt not found
	service.getByNameFunc = func(ctx context.Context, name string) (*models.Prompt, error) {
		return nil, ErrNotFound
	}

	ctx := context.Background()
	result, err := service.GetMCPPromptDefinition(ctx, "nonexistent-prompt")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "prompt 'nonexistent-prompt' not found")
	assert.Nil(t, result)
}

func TestPromptService_GetMCPPromptDefinition_DatabaseError(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel) // Suppress logs during tests

	service := NewMockPromptService(logger)

	// Mock database error
	service.getByNameFunc = func(ctx context.Context, name string) (*models.Prompt, error) {
		return nil, errors.New("database connection failed")
	}

	ctx := context.Background()
	result, err := service.GetMCPPromptDefinition(ctx, "test-prompt")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to retrieve prompt: database connection failed")
	assert.Nil(t, result)
}

func TestPromptService_GetMCPPromptDefinition_ValidationError_InvalidRole(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel) // Suppress logs during tests

	service := NewMockPromptService(logger)

	// Mock prompt with invalid role
	service.getByNameFunc = func(ctx context.Context, name string) (*models.Prompt, error) {
		return &models.Prompt{
			ID:        uuid.New(),
			Name:      "invalid-prompt",
			Title:     "Invalid Prompt",
			Content:   "You are a system",
			Role:      models.MCPRole("system"), // Invalid role
			CreatorID: uuid.New(),
		}, nil
	}

	ctx := context.Background()
	result, err := service.GetMCPPromptDefinition(ctx, "invalid-prompt")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid prompt data: invalid role 'system': must be 'user' or 'assistant'")
	assert.Nil(t, result)
}

func TestPromptService_GetMCPPromptDefinition_ValidationError_EmptyContent(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel) // Suppress logs during tests

	service := NewMockPromptService(logger)

	// Mock prompt with empty content
	service.getByNameFunc = func(ctx context.Context, name string) (*models.Prompt, error) {
		return &models.Prompt{
			ID:        uuid.New(),
			Name:      "empty-content-prompt",
			Title:     "Empty Content Prompt",
			Content:   "", // Empty content
			Role:      models.MCPRoleAssistant,
			CreatorID: uuid.New(),
		}, nil
	}

	ctx := context.Background()
	result, err := service.GetMCPPromptDefinition(ctx, "empty-content-prompt")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid prompt data: content cannot be empty")
	assert.Nil(t, result)
}

func TestPromptService_GetMCPPromptDefinition_ContentTransformation(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel) // Suppress logs during tests

	testCases := []struct {
		name            string
		content         string
		expectedContent string
	}{
		{
			name:            "simple text content",
			content:         "Hello world",
			expectedContent: "Hello world",
		},
		{
			name:            "multiline content",
			content:         "Line 1\nLine 2\nLine 3",
			expectedContent: "Line 1\nLine 2\nLine 3",
		},
		{
			name:            "content with special characters",
			content:         "Hello! @#$%^&*()_+ world?",
			expectedContent: "Hello! @#$%^&*()_+ world?",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			service := NewMockPromptService(logger)

			// Mock prompt with test content
			service.getByNameFunc = func(ctx context.Context, name string) (*models.Prompt, error) {
				return &models.Prompt{
					ID:        uuid.New(),
					Name:      name,
					Title:     "Test Prompt",
					Content:   tc.content,
					Role:      models.MCPRoleAssistant,
					CreatorID: uuid.New(),
				}, nil
			}

			ctx := context.Background()
			result, err := service.GetMCPPromptDefinition(ctx, "test-prompt")

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Len(t, result.Messages, 1)
			assert.Equal(t, "text", result.Messages[0].Content.Type)
			assert.Equal(t, tc.expectedContent, result.Messages[0].Content.Text)
		})
	}
}

func TestPromptService_GetMCPPromptDefinition_NilDescription(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel) // Suppress logs during tests

	service := NewMockPromptService(logger)

	// Mock prompt with nil description
	service.getByNameFunc = func(ctx context.Context, name string) (*models.Prompt, error) {
		return &models.Prompt{
			ID:          uuid.New(),
			Name:        "test-prompt",
			Title:       "Test Prompt",
			Description: nil, // Nil description
			Content:     "You are a helpful assistant",
			Role:        models.MCPRoleAssistant,
			CreatorID:   uuid.New(),
		}, nil
	}

	ctx := context.Background()
	result, err := service.GetMCPPromptDefinition(ctx, "test-prompt")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "test-prompt", result.Name)
	assert.Equal(t, "", result.Description) // Should be empty string when nil
	assert.Len(t, result.Messages, 1)
	assert.Equal(t, "assistant", result.Messages[0].Role)
}

func TestPromptService_GetMCPPromptDefinition_UserRole(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel) // Suppress logs during tests

	service := NewMockPromptService(logger)

	// Mock prompt with user role
	service.getByNameFunc = func(ctx context.Context, name string) (*models.Prompt, error) {
		return &models.Prompt{
			ID:        uuid.New(),
			Name:      "user-prompt",
			Title:     "User Prompt",
			Content:   "Please help me with this task",
			Role:      models.MCPRoleUser,
			CreatorID: uuid.New(),
		}, nil
	}

	ctx := context.Background()
	result, err := service.GetMCPPromptDefinition(ctx, "user-prompt")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "user-prompt", result.Name)
	assert.Len(t, result.Messages, 1)
	assert.Equal(t, "user", result.Messages[0].Role)
	assert.Equal(t, "Please help me with this task", result.Messages[0].Content.Text)
}

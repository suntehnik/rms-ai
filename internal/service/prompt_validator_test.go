package service

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"product-requirements-management/internal/models"
)

func TestPromptValidator_ValidateForMCP(t *testing.T) {
	validator := NewPromptValidator()

	tests := []struct {
		name      string
		prompt    *models.Prompt
		expectErr bool
		errMsg    string
	}{
		{
			name: "valid prompt with assistant role",
			prompt: &models.Prompt{
				ID:        uuid.New(),
				Name:      "test-prompt",
				Title:     "Test Prompt",
				Content:   "You are a helpful assistant",
				Role:      models.MCPRoleAssistant,
				CreatorID: uuid.New(),
			},
			expectErr: false,
		},
		{
			name: "valid prompt with user role",
			prompt: &models.Prompt{
				ID:        uuid.New(),
				Name:      "test-prompt",
				Title:     "Test Prompt",
				Content:   "Please help me with this task",
				Role:      models.MCPRoleUser,
				CreatorID: uuid.New(),
			},
			expectErr: false,
		},
		{
			name:      "nil prompt",
			prompt:    nil,
			expectErr: true,
			errMsg:    "prompt cannot be nil",
		},
		{
			name: "invalid role",
			prompt: &models.Prompt{
				ID:        uuid.New(),
				Name:      "test-prompt",
				Title:     "Test Prompt",
				Content:   "You are a system",
				Role:      models.MCPRole("system"),
				CreatorID: uuid.New(),
			},
			expectErr: true,
			errMsg:    "invalid role 'system': must be 'user' or 'assistant'",
		},
		{
			name: "empty content",
			prompt: &models.Prompt{
				ID:        uuid.New(),
				Name:      "test-prompt",
				Title:     "Test Prompt",
				Content:   "",
				Role:      models.MCPRoleAssistant,
				CreatorID: uuid.New(),
			},
			expectErr: true,
			errMsg:    "content cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateForMCP(tt.prompt)
			if tt.expectErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPromptValidator_ValidateCreateRequest(t *testing.T) {
	validator := NewPromptValidator()

	tests := []struct {
		name      string
		request   *CreatePromptRequest
		expectErr bool
		errMsg    string
	}{
		{
			name: "valid request with assistant role",
			request: &CreatePromptRequest{
				Name:    "test-prompt",
				Title:   "Test Prompt",
				Content: "You are a helpful assistant",
				Role:    &[]models.MCPRole{models.MCPRoleAssistant}[0],
			},
			expectErr: false,
		},
		{
			name: "valid request with user role",
			request: &CreatePromptRequest{
				Name:    "test-prompt",
				Title:   "Test Prompt",
				Content: "Please help me with this task",
				Role:    &[]models.MCPRole{models.MCPRoleUser}[0],
			},
			expectErr: false,
		},
		{
			name: "valid request without role",
			request: &CreatePromptRequest{
				Name:    "test-prompt",
				Title:   "Test Prompt",
				Content: "You are a helpful assistant",
				Role:    nil,
			},
			expectErr: false,
		},
		{
			name:      "nil request",
			request:   nil,
			expectErr: true,
			errMsg:    "create request cannot be nil",
		},
		{
			name: "invalid role",
			request: &CreatePromptRequest{
				Name:    "test-prompt",
				Title:   "Test Prompt",
				Content: "You are a system",
				Role:    &[]models.MCPRole{models.MCPRole("system")}[0],
			},
			expectErr: true,
			errMsg:    "invalid role 'system': must be 'user' or 'assistant'",
		},
		{
			name: "empty content",
			request: &CreatePromptRequest{
				Name:    "test-prompt",
				Title:   "Test Prompt",
				Content: "",
				Role:    &[]models.MCPRole{models.MCPRoleAssistant}[0],
			},
			expectErr: true,
			errMsg:    "content cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateCreateRequest(tt.request)
			if tt.expectErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPromptValidator_ValidateUpdateRequest(t *testing.T) {
	validator := NewPromptValidator()

	tests := []struct {
		name      string
		request   *UpdatePromptRequest
		expectErr bool
		errMsg    string
	}{
		{
			name: "valid request with content",
			request: &UpdatePromptRequest{
				Title:   &[]string{"Updated Title"}[0],
				Content: &[]string{"Updated content"}[0],
			},
			expectErr: false,
		},
		{
			name: "valid request without content",
			request: &UpdatePromptRequest{
				Title: &[]string{"Updated Title"}[0],
			},
			expectErr: false,
		},
		{
			name: "valid request with assistant role",
			request: &UpdatePromptRequest{
				Title: &[]string{"Updated Title"}[0],
				Role:  &[]models.MCPRole{models.MCPRoleAssistant}[0],
			},
			expectErr: false,
		},
		{
			name: "valid request with user role",
			request: &UpdatePromptRequest{
				Title: &[]string{"Updated Title"}[0],
				Role:  &[]models.MCPRole{models.MCPRoleUser}[0],
			},
			expectErr: false,
		},
		{
			name: "valid request without role",
			request: &UpdatePromptRequest{
				Title: &[]string{"Updated Title"}[0],
				Role:  nil,
			},
			expectErr: false,
		},
		{
			name:      "nil request",
			request:   nil,
			expectErr: true,
			errMsg:    "update request cannot be nil",
		},
		{
			name: "invalid role",
			request: &UpdatePromptRequest{
				Title: &[]string{"Updated Title"}[0],
				Role:  &[]models.MCPRole{models.MCPRole("system")}[0],
			},
			expectErr: true,
			errMsg:    "invalid role 'system': must be 'user' or 'assistant'",
		},
		{
			name: "empty content",
			request: &UpdatePromptRequest{
				Title:   &[]string{"Updated Title"}[0],
				Content: &[]string{""}[0],
			},
			expectErr: true,
			errMsg:    "content cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateUpdateRequest(tt.request)
			if tt.expectErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNewPromptValidator(t *testing.T) {
	validator := NewPromptValidator()
	assert.NotNil(t, validator)
	assert.IsType(t, &PromptValidator{}, validator)
}

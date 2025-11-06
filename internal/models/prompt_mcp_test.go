package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMCPRole_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		role     MCPRole
		expected bool
	}{
		{
			name:     "valid user role",
			role:     MCPRoleUser,
			expected: true,
		},
		{
			name:     "valid assistant role",
			role:     MCPRoleAssistant,
			expected: true,
		},
		{
			name:     "invalid system role",
			role:     MCPRole("system"),
			expected: false,
		},
		{
			name:     "invalid empty role",
			role:     MCPRole(""),
			expected: false,
		},
		{
			name:     "invalid custom role",
			role:     MCPRole("custom"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.role.IsValid()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMCPRole_String(t *testing.T) {
	tests := []struct {
		name     string
		role     MCPRole
		expected string
	}{
		{
			name:     "user role string",
			role:     MCPRoleUser,
			expected: "user",
		},
		{
			name:     "assistant role string",
			role:     MCPRoleAssistant,
			expected: "assistant",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.role.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateMCPRole(t *testing.T) {
	tests := []struct {
		name      string
		role      string
		expectErr bool
	}{
		{
			name:      "valid user role",
			role:      "user",
			expectErr: false,
		},
		{
			name:      "valid assistant role",
			role:      "assistant",
			expectErr: false,
		},
		{
			name:      "invalid system role",
			role:      "system",
			expectErr: true,
		},
		{
			name:      "invalid empty role",
			role:      "",
			expectErr: true,
		},
		{
			name:      "invalid custom role",
			role:      "custom",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMCPRole(tt.role)
			if tt.expectErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "invalid MCP role")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestContentChunk_ValidateForMCP(t *testing.T) {
	tests := []struct {
		name      string
		chunk     ContentChunk
		expectErr bool
		errMsg    string
	}{
		{
			name: "valid text chunk",
			chunk: ContentChunk{
				Type: "text",
				Text: "Hello world",
			},
			expectErr: false,
		},
		{
			name: "empty type",
			chunk: ContentChunk{
				Type: "",
				Text: "Hello world",
			},
			expectErr: true,
			errMsg:    "content chunk type cannot be empty",
		},
		{
			name: "unsupported type",
			chunk: ContentChunk{
				Type: "image",
				Text: "Hello world",
			},
			expectErr: true,
			errMsg:    "unsupported content chunk type 'image'",
		},
		{
			name: "text type with empty text",
			chunk: ContentChunk{
				Type: "text",
				Text: "",
			},
			expectErr: true,
			errMsg:    "text content cannot be empty for type 'text'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.chunk.ValidateForMCP()
			if tt.expectErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPromptMessage_ValidateForMCP(t *testing.T) {
	tests := []struct {
		name      string
		message   PromptMessage
		expectErr bool
		errMsg    string
	}{
		{
			name: "valid message",
			message: PromptMessage{
				Role: "user",
				Content: ContentChunk{
					Type: "text",
					Text: "Hello world",
				},
			},
			expectErr: false,
		},
		{
			name: "invalid role",
			message: PromptMessage{
				Role: "system",
				Content: ContentChunk{
					Type: "text",
					Text: "Hello world",
				},
			},
			expectErr: true,
			errMsg:    "invalid MCP role 'system'",
		},
		{
			name: "empty content",
			message: PromptMessage{
				Role:    "user",
				Content: ContentChunk{},
			},
			expectErr: true,
			errMsg:    "message content cannot be empty",
		},
		{
			name: "invalid content chunk",
			message: PromptMessage{
				Role: "user",
				Content: ContentChunk{
					Type: "text",
					Text: "",
				},
			},
			expectErr: true,
			errMsg:    "content chunk",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.message.ValidateForMCP()
			if tt.expectErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTransformContentToChunks(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected ContentChunk
	}{
		{
			name:    "non-empty content",
			content: "Hello world",
			expected: ContentChunk{
				Type: "text",
				Text: "Hello world",
			},
		},
		{
			name:     "empty content",
			content:  "",
			expected: ContentChunk{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TransformContentToChunk(tt.content)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPrompt_BeforeCreate_DefaultRole(t *testing.T) {
	prompt := &Prompt{
		Name:    "test-prompt",
		Title:   "Test Prompt",
		Content: "Test content",
	}

	// Test that default role is set when role is empty
	assert.Equal(t, MCPRole(""), prompt.Role)

	// Set a reference ID to avoid the generator call
	prompt.ReferenceID = "PROMPT-001"

	// Simulate BeforeCreate hook
	err := prompt.BeforeCreate(nil)
	assert.NoError(t, err)
	assert.Equal(t, MCPRoleAssistant, prompt.Role)
}

func TestPrompt_BeforeCreate_PreserveExistingRole(t *testing.T) {
	prompt := &Prompt{
		Name:    "test-prompt",
		Title:   "Test Prompt",
		Content: "Test content",
		Role:    MCPRoleUser,
	}

	// Set a reference ID to avoid the generator call
	prompt.ReferenceID = "PROMPT-001"

	// Simulate BeforeCreate hook
	err := prompt.BeforeCreate(nil)
	assert.NoError(t, err)
	assert.Equal(t, MCPRoleUser, prompt.Role)
}

package service

import (
	"fmt"

	"product-requirements-management/internal/models"
)

// PromptValidator provides validation functionality for prompts and MCP compliance
type PromptValidator struct{}

// NewPromptValidator creates a new PromptValidator instance
func NewPromptValidator() *PromptValidator {
	return &PromptValidator{}
}

// ValidateForMCP validates that a prompt is compliant with MCP specification
// This method checks:
// - Role enum values are valid ("user" or "assistant")
// - Content is not empty
// - Content can be transformed to valid ContentChunk array
func (v *PromptValidator) ValidateForMCP(prompt *models.Prompt) error {
	if prompt == nil {
		return fmt.Errorf("prompt cannot be nil")
	}

	// Validate role enum values
	if !prompt.Role.IsValid() {
		return fmt.Errorf("invalid role '%s': must be 'user' or 'assistant'", prompt.Role)
	}

	// Validate content requirements
	if prompt.Content == "" {
		return fmt.Errorf("content cannot be empty")
	}

	// Validate that content can be transformed to valid ContentChunk array
	contentChunk := models.TransformContentToChunk(prompt.Content)

	// Create a test PromptMessage to validate the complete structure
	testMessage := &models.PromptMessage{
		Role:    string(prompt.Role),
		Content: contentChunk,
	}

	if err := testMessage.ValidateForMCP(); err != nil {
		return fmt.Errorf("prompt message validation failed: %w", err)
	}

	return nil
}

// ValidateCreateRequest validates a CreatePromptRequest for MCP compliance
func (v *PromptValidator) ValidateCreateRequest(req *CreatePromptRequest) error {
	if req == nil {
		return fmt.Errorf("create request cannot be nil")
	}

	// Validate role if provided
	if req.Role != nil {
		if !req.Role.IsValid() {
			return fmt.Errorf("invalid role '%s': must be 'user' or 'assistant'", *req.Role)
		}
	}

	// Validate content requirements
	if req.Content == "" {
		return fmt.Errorf("content cannot be empty")
	}

	return nil
}

// ValidateUpdateRequest validates an UpdatePromptRequest for MCP compliance
func (v *PromptValidator) ValidateUpdateRequest(req *UpdatePromptRequest) error {
	if req == nil {
		return fmt.Errorf("update request cannot be nil")
	}

	// Validate role if provided
	if req.Role != nil {
		if !req.Role.IsValid() {
			return fmt.Errorf("invalid role '%s': must be 'user' or 'assistant'", *req.Role)
		}
	}

	// Validate content if provided
	if req.Content != nil {
		if *req.Content == "" {
			return fmt.Errorf("content cannot be empty")
		}
	}

	return nil
}

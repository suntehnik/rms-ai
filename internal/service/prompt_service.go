package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"product-requirements-management/internal/models"
)

// PromptService handles business logic for system prompts
type PromptService struct {
	db     *gorm.DB
	logger *logrus.Logger
}

// NewPromptService creates a new PromptService instance
func NewPromptService(db *gorm.DB, logger *logrus.Logger) *PromptService {
	return &PromptService{
		db:     db,
		logger: logger,
	}
}

// CreatePromptRequest represents the request to create a new prompt
type CreatePromptRequest struct {
	Name        string  `json:"name" validate:"required,max=255"`
	Title       string  `json:"title" validate:"required,max=500"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=50000"`
	Content     string  `json:"content" validate:"required"`
}

// UpdatePromptRequest represents the request to update an existing prompt
type UpdatePromptRequest struct {
	Title       *string `json:"title,omitempty" validate:"omitempty,max=500"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=50000"`
	Content     *string `json:"content,omitempty"`
}

// Create creates a new prompt
func (ps *PromptService) Create(ctx context.Context, req *CreatePromptRequest, creatorID uuid.UUID) (*models.Prompt, error) {
	prompt := &models.Prompt{
		Name:        req.Name,
		Title:       req.Title,
		Description: req.Description,
		Content:     req.Content,
		CreatorID:   creatorID,
		IsActive:    false, // New prompts are not active by default
	}

	if err := ps.db.WithContext(ctx).Create(prompt).Error; err != nil {
		ps.logger.WithError(err).Error("Failed to create prompt")
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, ErrDuplicateEntry
		}
		return nil, fmt.Errorf("failed to create prompt: %w", err)
	}

	ps.logger.WithFields(logrus.Fields{
		"prompt_id":    prompt.ID,
		"reference_id": prompt.ReferenceID,
		"name":         prompt.Name,
		"creator_id":   creatorID,
	}).Info("Prompt created successfully")

	return prompt, nil
}

// GetByID retrieves a prompt by its UUID
func (ps *PromptService) GetByID(ctx context.Context, id uuid.UUID) (*models.Prompt, error) {
	var prompt models.Prompt
	if err := ps.db.WithContext(ctx).Where("id = ?", id).First(&prompt).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		ps.logger.WithError(err).WithField("prompt_id", id).Error("Failed to get prompt by ID")
		return nil, fmt.Errorf("failed to get prompt: %w", err)
	}
	return &prompt, nil
}

// GetByReferenceID retrieves a prompt by its reference ID
func (ps *PromptService) GetByReferenceID(ctx context.Context, referenceID string) (*models.Prompt, error) {
	var prompt models.Prompt
	if err := ps.db.WithContext(ctx).Where("reference_id = ?", referenceID).First(&prompt).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		ps.logger.WithError(err).WithField("reference_id", referenceID).Error("Failed to get prompt by reference ID")
		return nil, fmt.Errorf("failed to get prompt: %w", err)
	}
	return &prompt, nil
}

// GetByName retrieves a prompt by its name
func (ps *PromptService) GetByName(ctx context.Context, name string) (*models.Prompt, error) {
	var prompt models.Prompt
	if err := ps.db.WithContext(ctx).Where("name = ?", name).First(&prompt).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		ps.logger.WithError(err).WithField("name", name).Error("Failed to get prompt by name")
		return nil, fmt.Errorf("failed to get prompt: %w", err)
	}
	return &prompt, nil
}

// GetActive retrieves the currently active prompt
func (ps *PromptService) GetActive(ctx context.Context) (*models.Prompt, error) {
	var prompt models.Prompt
	if err := ps.db.WithContext(ctx).Where("is_active = ?", true).First(&prompt).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		ps.logger.WithError(err).Error("Failed to get active prompt")
		return nil, fmt.Errorf("failed to get active prompt: %w", err)
	}
	return &prompt, nil
}

// List retrieves prompts with pagination and optional filtering
func (ps *PromptService) List(ctx context.Context, limit, offset int, creatorID *uuid.UUID) ([]*models.Prompt, int64, error) {
	var prompts []*models.Prompt
	var total int64

	query := ps.db.WithContext(ctx).Model(&models.Prompt{})

	// Apply creator filter if provided
	if creatorID != nil {
		query = query.Where("creator_id = ?", *creatorID)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		ps.logger.WithError(err).Error("Failed to count prompts")
		return nil, 0, fmt.Errorf("failed to count prompts: %w", err)
	}

	// Get prompts with pagination
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&prompts).Error; err != nil {
		ps.logger.WithError(err).Error("Failed to list prompts")
		return nil, 0, fmt.Errorf("failed to list prompts: %w", err)
	}

	return prompts, total, nil
}

// Update updates an existing prompt
func (ps *PromptService) Update(ctx context.Context, id uuid.UUID, req *UpdatePromptRequest) (*models.Prompt, error) {
	var prompt models.Prompt
	if err := ps.db.WithContext(ctx).Where("id = ?", id).First(&prompt).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		ps.logger.WithError(err).WithField("prompt_id", id).Error("Failed to find prompt for update")
		return nil, fmt.Errorf("failed to find prompt: %w", err)
	}

	// Update fields if provided
	if req.Title != nil {
		prompt.Title = *req.Title
	}
	if req.Description != nil {
		prompt.Description = req.Description
	}
	if req.Content != nil {
		prompt.Content = *req.Content
	}

	if err := ps.db.WithContext(ctx).Save(&prompt).Error; err != nil {
		ps.logger.WithError(err).WithField("prompt_id", id).Error("Failed to update prompt")
		return nil, fmt.Errorf("failed to update prompt: %w", err)
	}

	ps.logger.WithFields(logrus.Fields{
		"prompt_id":    prompt.ID,
		"reference_id": prompt.ReferenceID,
		"name":         prompt.Name,
	}).Info("Prompt updated successfully")

	return &prompt, nil
}

// Activate activates a prompt and deactivates all others
func (ps *PromptService) Activate(ctx context.Context, id uuid.UUID) error {
	return ps.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// First check if the prompt exists
		var prompt models.Prompt
		if err := tx.Where("id = ?", id).First(&prompt).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrNotFound
			}
			return fmt.Errorf("failed to find prompt: %w", err)
		}

		// Deactivate all prompts (add WHERE clause to satisfy GORM safety requirements)
		if err := tx.Model(&models.Prompt{}).Where("is_active = ?", true).Update("is_active", false).Error; err != nil {
			return fmt.Errorf("failed to deactivate prompts: %w", err)
		}

		// Activate selected prompt
		if err := tx.Model(&models.Prompt{}).Where("id = ?", id).Update("is_active", true).Error; err != nil {
			return fmt.Errorf("failed to activate prompt: %w", err)
		}

		ps.logger.WithFields(logrus.Fields{
			"prompt_id":    id,
			"reference_id": prompt.ReferenceID,
			"name":         prompt.Name,
		}).Info("Prompt activated successfully")

		return nil
	})
}

// Delete deletes a prompt
func (ps *PromptService) Delete(ctx context.Context, id uuid.UUID) error {
	// First check if the prompt exists and get its info for logging
	var prompt models.Prompt
	if err := ps.db.WithContext(ctx).Where("id = ?", id).First(&prompt).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		ps.logger.WithError(err).WithField("prompt_id", id).Error("Failed to find prompt for deletion")
		return fmt.Errorf("failed to find prompt: %w", err)
	}

	if err := ps.db.WithContext(ctx).Where("id = ?", id).Delete(&models.Prompt{}).Error; err != nil {
		ps.logger.WithError(err).WithField("prompt_id", id).Error("Failed to delete prompt")
		return fmt.Errorf("failed to delete prompt: %w", err)
	}

	ps.logger.WithFields(logrus.Fields{
		"prompt_id":    id,
		"reference_id": prompt.ReferenceID,
		"name":         prompt.Name,
	}).Info("Prompt deleted successfully")

	return nil
}

// GetMCPPromptDescriptors returns prompt descriptors for MCP protocol
func (ps *PromptService) GetMCPPromptDescriptors(ctx context.Context) ([]*models.MCPPromptDescriptor, error) {
	var prompts []*models.Prompt
	if err := ps.db.WithContext(ctx).Find(&prompts).Error; err != nil {
		ps.logger.WithError(err).Error("Failed to get prompts for MCP descriptors")
		return nil, fmt.Errorf("failed to get prompts: %w", err)
	}

	descriptors := make([]*models.MCPPromptDescriptor, len(prompts))
	for i, prompt := range prompts {
		description := ""
		if prompt.Description != nil {
			description = *prompt.Description
		}
		descriptors[i] = &models.MCPPromptDescriptor{
			Name:        prompt.Name,
			Description: description,
		}
	}

	return descriptors, nil
}

// GetMCPPromptDefinition returns a full prompt definition for MCP protocol
func (ps *PromptService) GetMCPPromptDefinition(ctx context.Context, name string) (*models.MCPPromptDefinition, error) {
	prompt, err := ps.GetByName(ctx, name)
	if err != nil {
		return nil, err
	}

	description := ""
	if prompt.Description != nil {
		description = *prompt.Description
	}

	definition := &models.MCPPromptDefinition{
		Name:        prompt.Name,
		Description: description,
		Messages: []models.PromptMessage{
			{
				Role:    "system",
				Content: prompt.Content,
			},
		},
	}

	return definition, nil
}

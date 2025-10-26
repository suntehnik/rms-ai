package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Package-level generator instance for Prompt reference IDs.
//
// This uses the production PostgreSQLReferenceIDGenerator which provides:
// - Thread-safe reference ID generation for production environments
// - PostgreSQL advisory locks for atomic generation (lock key: 2147483642)
// - UUID fallback when locks are unavailable
// - Automatic PostgreSQL vs SQLite detection
//
// For unit tests, use TestReferenceIDGenerator from reference_id_test.go instead.
// The static selection approach ensures the right generator is used in the right context.
var promptGenerator = NewPostgreSQLReferenceIDGenerator(2147483642, "PROMPT")

// Prompt represents a system prompt in the system
// @Description System prompt contains instructions for AI assistant behavior and can be activated for use
type Prompt struct {
	// ID is the unique identifier for the prompt
	// @Description Unique UUID identifier for the prompt
	// @Example "123e4567-e89b-12d3-a456-426614174000"
	ID uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`

	// ReferenceID is the human-readable identifier for the prompt
	// @Description Human-readable reference identifier (auto-generated, format: PROMPT-XXX)
	// @Example "PROMPT-001"
	ReferenceID string `gorm:"uniqueIndex;not null" json:"reference_id"`

	// Name is the unique name identifier for the prompt
	// @Description Unique name identifier for the prompt (required, max 255 characters)
	// @MaxLength 255
	// @Example "requirements-analyst"
	Name string `gorm:"uniqueIndex;not null" json:"name" validate:"required,max=255"`

	// Title is the display name of the prompt
	// @Description Display title of the prompt (required, max 500 characters)
	// @MaxLength 500
	// @Example "Requirements Analyst Assistant"
	Title string `gorm:"not null" json:"title" validate:"required,max=500"`

	// Description provides information about the prompt's purpose
	// @Description Detailed description of the prompt's purpose (optional, max 50000 characters)
	// @MaxLength 50000
	// @Example "AI assistant specialized in requirements analysis and management"
	Description *string `json:"description,omitempty" validate:"omitempty,max=50000"`

	// Content is the actual prompt text
	// @Description The actual prompt content/instructions (required)
	// @Example "You are an expert requirements analyst..."
	Content string `gorm:"type:text;not null" json:"content" validate:"required"`

	// IsActive indicates if this prompt is currently active
	// @Description Whether this prompt is currently active (only one can be active at a time)
	// @Example true
	IsActive bool `gorm:"default:false" json:"is_active"`

	// CreatorID is the UUID of the user who created the prompt
	// @Description UUID of the user who created this prompt
	// @Example "123e4567-e89b-12d3-a456-426614174001"
	CreatorID uuid.UUID `gorm:"not null" json:"creator_id"`

	// CreatedAt is the timestamp when the prompt was created
	// @Description Timestamp when the prompt was created (RFC3339 format)
	// @Example "2023-01-15T10:30:00Z"
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is the timestamp when the prompt was last updated
	// @Description Timestamp when the prompt was last modified (RFC3339 format)
	// @Example "2023-01-16T14:45:30Z"
	UpdatedAt time.Time `json:"updated_at"`

	// Relationships - These fields are populated when explicitly requested and contain related entities

	// Creator contains the user information of who created the prompt
	// @Description User who created this prompt (included when explicitly preloaded)
	Creator User `gorm:"foreignKey:CreatorID;constraint:OnDelete:RESTRICT" json:"-"`
}

// BeforeCreate sets the ID if not already set and generates reference ID
func (p *Prompt) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}

	// Generate reference ID if not set
	if p.ReferenceID == "" {
		referenceID, err := promptGenerator.Generate(tx, &Prompt{})
		if err != nil {
			return err
		}
		p.ReferenceID = referenceID
	}

	return nil
}

// BeforeUpdate updates the UpdatedAt timestamp and handles active prompt constraints
func (p *Prompt) BeforeUpdate(tx *gorm.DB) error {
	p.UpdatedAt = time.Now().UTC()

	// If setting this prompt as active, deactivate all others
	if p.IsActive {
		if err := tx.Model(&Prompt{}).Where("id != ?", p.ID).Update("is_active", false).Error; err != nil {
			return err
		}
	}

	return nil
}

// TableName returns the table name for the Prompt model
func (Prompt) TableName() string {
	return "prompts"
}

// MarshalJSON implements custom JSON marshaling for Prompt
// This ensures that creator object is only included when it is actually populated
func (p *Prompt) MarshalJSON() ([]byte, error) {
	// Create a map to build the JSON response
	result := map[string]interface{}{
		"id":           p.ID,
		"reference_id": p.ReferenceID,
		"name":         p.Name,
		"title":        p.Title,
		"content":      p.Content,
		"is_active":    p.IsActive,
		"creator_id":   p.CreatorID,
		"created_at":   p.CreatedAt,
		"updated_at":   p.UpdatedAt,
	}

	// Only include description if it's not nil
	if p.Description != nil {
		result["description"] = *p.Description
	}

	// Only include creator if it has been populated (has a username, indicating it was preloaded)
	if p.Creator.Username != "" {
		result["creator"] = p.Creator
	}

	return json.Marshal(result)
}

// MCPPromptDescriptor represents a prompt descriptor for MCP protocol
// @Description Prompt descriptor for MCP protocol listing
type MCPPromptDescriptor struct {
	// Name is the unique identifier for the prompt
	// @Description Unique name identifier for the prompt
	// @Example "requirements-analyst"
	Name string `json:"name"`

	// Description provides information about the prompt's purpose
	// @Description Description of the prompt's purpose
	// @Example "AI assistant specialized in requirements analysis and management"
	Description string `json:"description"`
}

// MCPPromptDefinition represents a full prompt definition for MCP protocol
// @Description Full prompt definition for MCP protocol
type MCPPromptDefinition struct {
	// Name is the unique identifier for the prompt
	// @Description Unique name identifier for the prompt
	// @Example "requirements-analyst"
	Name string `json:"name"`

	// Description provides information about the prompt's purpose
	// @Description Description of the prompt's purpose
	// @Example "AI assistant specialized in requirements analysis and management"
	Description string `json:"description"`

	// Messages contains the prompt messages
	// @Description Array of prompt messages
	Messages []PromptMessage `json:"messages"`
}

// PromptMessage represents a message in the MCP prompt format
// @Description Message structure for MCP prompt protocol
type PromptMessage struct {
	// Role indicates the message role (system, user, assistant)
	// @Description Role of the message sender
	// @Example "system"
	Role string `json:"role"`

	// Content is the message content
	// @Description Content of the message
	// @Example "You are an expert requirements analyst..."
	Content string `json:"content"`
}

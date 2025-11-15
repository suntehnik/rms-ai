package models

import (
	"encoding/json"
	"fmt"
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
var promptGenerator ReferenceIDGenerator = NewPostgreSQLReferenceIDGenerator(2147483642, "PROMPT")

// GetPromptGenerator returns the current generator (for testing)
func GetPromptGenerator() ReferenceIDGenerator {
	return promptGenerator
}

// SetPromptGenerator sets a custom generator (for testing)
func SetPromptGenerator(gen ReferenceIDGenerator) {
	promptGenerator = gen
}

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

	// Role specifies the message role for MCP compliance
	// @Description Role for MCP prompt messages (must be "user" or "assistant", defaults to "assistant")
	// @Example "assistant"
	Role MCPRole `gorm:"type:varchar(20);not null;default:'assistant'" json:"role" validate:"required"`

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

	// Set default role if not specified
	if p.Role == "" {
		p.Role = MCPRoleAssistant
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
		"role":         p.Role,
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

// MCPRole represents valid MCP message roles
// @Description Valid roles for MCP prompt messages according to 2025-06-18 specification
type MCPRole string

const (
	// MCPRoleUser represents a user message role
	// @Description User role for MCP messages
	MCPRoleUser MCPRole = "user"

	// MCPRoleAssistant represents an assistant message role
	// @Description Assistant role for MCP messages
	MCPRoleAssistant MCPRole = "assistant"
)

// IsValid checks if the MCPRole is valid according to MCP specification
func (r MCPRole) IsValid() bool {
	return r == MCPRoleUser || r == MCPRoleAssistant
}

// String returns the string representation of MCPRole
func (r MCPRole) String() string {
	return string(r)
}

// ValidateMCPRole validates that a role string is a valid MCP role
func ValidateMCPRole(role string) error {
	mcpRole := MCPRole(role)
	if !mcpRole.IsValid() {
		return fmt.Errorf("invalid MCP role '%s': must be 'user' or 'assistant'", role)
	}
	return nil
}

// ValidateForMCP validates that a ContentChunk is MCP-compliant
func (c *ContentChunk) ValidateForMCP() error {
	if c.Type == "" {
		return fmt.Errorf("content chunk type cannot be empty")
	}

	// For now, we only support "text" type
	if c.Type != "text" {
		return fmt.Errorf("unsupported content chunk type '%s': only 'text' is supported", c.Type)
	}

	if c.Type == "text" && c.Text == "" {
		return fmt.Errorf("text content cannot be empty for type 'text'")
	}

	return nil
}

// ValidateForMCP validates that a PromptMessage is MCP-compliant
func (pm *PromptMessage) ValidateForMCP() error {
	// Validate role
	if err := ValidateMCPRole(pm.Role); err != nil {
		return err
	}

	// Validate content - check if it's empty (zero value)
	if pm.Content.Type == "" && pm.Content.Text == "" {
		return fmt.Errorf("message content cannot be empty")
	}

	// Validate the content chunk
	if err := pm.Content.ValidateForMCP(); err != nil {
		return fmt.Errorf("content chunk: %w", err)
	}

	return nil
}

// TransformContentToChunk transforms plain string content into structured ContentChunk
func TransformContentToChunk(content string) ContentChunk {
	if content == "" {
		return ContentChunk{}
	}

	return ContentChunk{
		Type: "text",
		Text: content,
	}
}

// ContentChunk represents a typed content object for MCP messages
// @Description Structured content chunk for MCP message content array
type ContentChunk struct {
	// Type specifies the content type (e.g., "text")
	// @Description Content type discriminator
	// @Example "text"
	Type string `json:"type" validate:"required"`

	// Text contains the text content for type="text"
	// @Description Text content (required when type="text")
	// @Example "You are an expert requirements analyst..."
	Text string `json:"text,omitempty"`

	// Future: additional fields for other content types can be added here
}

// PromptMessage represents a message in the MCP prompt format
// @Description Message structure for MCP prompt protocol with structured content
type PromptMessage struct {
	// Role indicates the message role (user, assistant)
	// @Description Role of the message sender (must be "user" or "assistant")
	// @Example "assistant"
	Role string `json:"role" validate:"required"`

	// Content is the structured message content array
	// @Description Array of typed content objects
	Content ContentChunk `json:"content" validate:"required,min=1"`
}

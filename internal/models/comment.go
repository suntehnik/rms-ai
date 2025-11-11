package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// EntityType represents the type of entity that can be commented on
// @Description Type of entity that can receive comments in the system
// @Example "epic"
type EntityType string

const (
	EntityTypeEpic               EntityType = "epic"                // Epic - top-level feature container
	EntityTypeUserStory          EntityType = "user_story"          // User Story - feature requirement within an epic
	EntityTypeAcceptanceCriteria EntityType = "acceptance_criteria" // Acceptance Criteria - testable conditions for user stories
	EntityTypeRequirement        EntityType = "requirement"         // Requirement - detailed technical requirement
)

// Comment represents a comment on any entity in the system
// @Description A comment that can be attached to any entity, supporting both general and inline comments with threading
type Comment struct {
	ID              uuid.UUID  `gorm:"type:uuid;primary_key" json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`                                         // Unique identifier for the comment
	EntityType      EntityType `gorm:"not null" json:"entity_type" validate:"required" example:"epic"`                                                         // Type of entity this comment is attached to
	EntityID        uuid.UUID  `gorm:"not null" json:"entity_id" example:"123e4567-e89b-12d3-a456-426614174001"`                                               // ID of the entity this comment is attached to
	ParentCommentID *uuid.UUID `json:"parent_comment_id" example:"123e4567-e89b-12d3-a456-426614174002"`                                                       // Optional ID of parent comment for threaded discussions
	AuthorID        uuid.UUID  `gorm:"not null" json:"author_id" example:"123e4567-e89b-12d3-a456-426614174003"`                                               // ID of the user who authored this comment
	CreatedAt       time.Time  `json:"created_at" example:"2023-01-01T00:00:00Z"`                                                                              // Timestamp when the comment was created
	UpdatedAt       time.Time  `json:"updated_at" example:"2023-01-02T12:30:00Z"`                                                                              // Timestamp when the comment was last updated
	Content         string     `gorm:"not null" json:"content" validate:"required" example:"This requirement needs clarification on the authentication flow."` // Text content of the comment
	IsResolved      bool       `json:"is_resolved" example:"false"`                                                                                            // Whether this comment has been resolved

	// For inline comments
	LinkedText        *string `json:"linked_text" example:"OAuth 2.0 authentication flow"` // Text that this inline comment is linked to
	TextPositionStart *int    `json:"text_position_start" example:"45"`                    // Start position of linked text for inline comments
	TextPositionEnd   *int    `json:"text_position_end" example:"73"`                      // End position of linked text for inline comments

	// Relationships
	ParentComment *Comment  `gorm:"foreignKey:ParentCommentID;constraint:OnDelete:CASCADE" json:"parent_comment,omitempty"` // Parent comment for threaded discussions
	Author        User      `gorm:"foreignKey:AuthorID;constraint:OnDelete:RESTRICT" json:"author,omitempty"`               // User who authored this comment
	Replies       []Comment `gorm:"foreignKey:ParentCommentID;constraint:OnDelete:CASCADE" json:"replies,omitempty"`        // Replies to this comment
}

// BeforeCreate sets the ID if not already set
func (c *Comment) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

// TableName returns the table name for the Comment model
func (Comment) TableName() string {
	return "comments"
}

// IsValidEntityType checks if the provided entity type is valid
func (c *Comment) IsValidEntityType(entityType EntityType) bool {
	validTypes := []EntityType{
		EntityTypeEpic,
		EntityTypeUserStory,
		EntityTypeAcceptanceCriteria,
		EntityTypeRequirement,
	}

	for _, validType := range validTypes {
		if entityType == validType {
			return true
		}
	}
	return false
}

// IsInlineComment checks if this is an inline comment (has text position data)
func (c *Comment) IsInlineComment() bool {
	return c.LinkedText != nil && c.TextPositionStart != nil && c.TextPositionEnd != nil
}

// IsGeneralComment checks if this is a general comment (not inline)
func (c *Comment) IsGeneralComment() bool {
	return !c.IsInlineComment()
}

// IsReply checks if this comment is a reply to another comment
func (c *Comment) IsReply() bool {
	return c.ParentCommentID != nil
}

// IsTopLevel checks if this comment is a top-level comment (not a reply)
func (c *Comment) IsTopLevel() bool {
	return c.ParentCommentID == nil
}

// HasReplies checks if this comment has any replies
func (c *Comment) HasReplies() bool {
	return len(c.Replies) > 0
}

// MarkAsResolved marks the comment as resolved
func (c *Comment) MarkAsResolved() {
	c.IsResolved = true
}

// MarkAsUnresolved marks the comment as unresolved
func (c *Comment) MarkAsUnresolved() {
	c.IsResolved = false
}

// GetDepth calculates the depth of the comment in the thread
// Top-level comments have depth 0, replies have depth 1, etc.
func (c *Comment) GetDepth() int {
	if c.ParentCommentID == nil {
		return 0
	}
	// In a real implementation, you would need to traverse up the parent chain
	// For now, we'll return 1 for any reply (assuming single-level threading for simplicity)
	return 1
}

// IsValidTextPosition validates that the text position data is consistent
func (c *Comment) IsValidTextPosition() bool {
	if !c.IsInlineComment() {
		return true // General comments don't need text position validation
	}

	return c.TextPositionStart != nil &&
		c.TextPositionEnd != nil &&
		*c.TextPositionStart >= 0 &&
		*c.TextPositionEnd >= *c.TextPositionStart
}

// MarshalJSON implements custom JSON marshaling for Comment
// This ensures that related objects are only included when they are actually populated
func (c *Comment) MarshalJSON() ([]byte, error) {
	// Create a map to build the JSON response
	result := map[string]interface{}{
		"id":          c.ID,
		"entity_type": c.EntityType,
		"entity_id":   c.EntityID,
		"author_id":   c.AuthorID,
		"created_at":  c.CreatedAt,
		"updated_at":  c.UpdatedAt,
		"content":     c.Content,
		"is_resolved": c.IsResolved,
	}

	// Only include parent_comment_id if it's not nil
	if c.ParentCommentID != nil {
		result["parent_comment_id"] = *c.ParentCommentID
	}

	// Only include inline comment fields if they are set
	if c.LinkedText != nil {
		result["linked_text"] = *c.LinkedText
	}
	if c.TextPositionStart != nil {
		result["text_position_start"] = *c.TextPositionStart
	}
	if c.TextPositionEnd != nil {
		result["text_position_end"] = *c.TextPositionEnd
	}

	// Only include parent_comment if it has been populated (has content, indicating it was preloaded)
	if c.ParentComment != nil && c.ParentComment.Content != "" {
		result["parent_comment"] = c.ParentComment
	}

	// Only include author if it has been populated (has a username, indicating it was preloaded)
	if c.Author.Username != "" {
		result["author"] = c.Author
	}

	// Only include replies if they have been populated
	if len(c.Replies) > 0 {
		result["replies"] = c.Replies
	}

	return json.Marshal(result)
}

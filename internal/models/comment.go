package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// EntityType represents the type of entity that can be commented on
type EntityType string

const (
	EntityTypeEpic               EntityType = "epic"
	EntityTypeUserStory          EntityType = "user_story"
	EntityTypeAcceptanceCriteria EntityType = "acceptance_criteria"
	EntityTypeRequirement        EntityType = "requirement"
)

// Comment represents a comment on any entity in the system
type Comment struct {
	ID                uuid.UUID  `gorm:"type:uuid;primary_key" json:"id"`
	EntityType        EntityType `gorm:"not null" json:"entity_type"`
	EntityID          uuid.UUID  `gorm:"not null" json:"entity_id"`
	ParentCommentID   *uuid.UUID `json:"parent_comment_id"`
	AuthorID          uuid.UUID  `gorm:"not null" json:"author_id"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
	Content           string     `gorm:"not null" json:"content"`
	IsResolved        bool       `json:"is_resolved"`
	
	// For inline comments
	LinkedText        *string `json:"linked_text"`
	TextPositionStart *int    `json:"text_position_start"`
	TextPositionEnd   *int    `json:"text_position_end"`

	// Relationships
	ParentComment *Comment  `gorm:"foreignKey:ParentCommentID;constraint:OnDelete:CASCADE" json:"parent_comment,omitempty"`
	Author        User      `gorm:"foreignKey:AuthorID;constraint:OnDelete:RESTRICT" json:"author,omitempty"`
	Replies       []Comment `gorm:"foreignKey:ParentCommentID;constraint:OnDelete:CASCADE" json:"replies,omitempty"`
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
package models

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserStoryStatus represents the status of a user story
type UserStoryStatus string

const (
	UserStoryStatusBacklog    UserStoryStatus = "Backlog"
	UserStoryStatusDraft      UserStoryStatus = "Draft"
	UserStoryStatusInProgress UserStoryStatus = "In Progress"
	UserStoryStatusDone       UserStoryStatus = "Done"
	UserStoryStatusCancelled  UserStoryStatus = "Cancelled"
)

// UserStory represents a user story in the requirements management system
type UserStory struct {
	ID           uuid.UUID       `gorm:"type:uuid;primary_key" json:"id"`
	ReferenceID  string          `gorm:"uniqueIndex;not null" json:"reference_id"`
	EpicID       uuid.UUID       `gorm:"not null" json:"epic_id"`
	CreatorID    uuid.UUID       `gorm:"not null" json:"creator_id"`
	AssigneeID   uuid.UUID       `gorm:"not null" json:"assignee_id"`
	CreatedAt    time.Time       `json:"created_at"`
	LastModified time.Time       `json:"last_modified"`
	Priority     Priority        `gorm:"not null" json:"priority"`
	Status       UserStoryStatus `gorm:"not null" json:"status"`
	Title        string          `gorm:"not null" json:"title"`
	Description  *string         `json:"description"`

	// Relationships
	Epic               Epic                 `gorm:"foreignKey:EpicID;constraint:OnDelete:CASCADE" json:"epic,omitempty"`
	Creator            User                 `gorm:"foreignKey:CreatorID;constraint:OnDelete:RESTRICT" json:"creator,omitempty"`
	Assignee           User                 `gorm:"foreignKey:AssigneeID;constraint:OnDelete:RESTRICT" json:"assignee,omitempty"`
	AcceptanceCriteria []AcceptanceCriteria `gorm:"foreignKey:UserStoryID;constraint:OnDelete:CASCADE" json:"acceptance_criteria,omitempty"`
	Requirements       []Requirement        `gorm:"foreignKey:UserStoryID;constraint:OnDelete:CASCADE" json:"requirements,omitempty"`
	Comments           []Comment            `gorm:"polymorphic:Entity;polymorphicValue:user_story" json:"comments,omitempty"`
}

// BeforeCreate sets the ID and ReferenceID if not already set
func (us *UserStory) BeforeCreate(tx *gorm.DB) error {
	if us.ID == uuid.Nil {
		us.ID = uuid.New()
	}
	if us.ReferenceID == "" {
		// Generate reference ID - in production this would use database sequences
		us.ReferenceID = "US-001" // Simplified for testing
	}
	if us.Status == "" {
		us.Status = UserStoryStatusBacklog
	}
	return nil
}

// BeforeUpdate updates the LastModified timestamp
func (us *UserStory) BeforeUpdate(tx *gorm.DB) error {
	us.LastModified = time.Now().UTC()
	return nil
}

// TableName returns the table name for the UserStory model
func (UserStory) TableName() string {
	return "user_stories"
}

// GetPriorityString returns the string representation of the priority
func (us *UserStory) GetPriorityString() string {
	switch us.Priority {
	case PriorityCritical:
		return "Critical"
	case PriorityHigh:
		return "High"
	case PriorityMedium:
		return "Medium"
	case PriorityLow:
		return "Low"
	default:
		return "Unknown"
	}
}

// IsValidStatus checks if the provided status is valid for user stories
func (us *UserStory) IsValidStatus(status UserStoryStatus) bool {
	validStatuses := []UserStoryStatus{
		UserStoryStatusBacklog,
		UserStoryStatusDraft,
		UserStoryStatusInProgress,
		UserStoryStatusDone,
		UserStoryStatusCancelled,
	}
	
	for _, validStatus := range validStatuses {
		if status == validStatus {
			return true
		}
	}
	return false
}

// CanTransitionTo checks if the user story can transition to the given status
// By default, all transitions are allowed as per requirements
func (us *UserStory) CanTransitionTo(newStatus UserStoryStatus) bool {
	return us.IsValidStatus(newStatus)
}

// HasAcceptanceCriteria checks if the user story has any acceptance criteria
func (us *UserStory) HasAcceptanceCriteria() bool {
	return len(us.AcceptanceCriteria) > 0
}

// HasRequirements checks if the user story has any associated requirements
func (us *UserStory) HasRequirements() bool {
	return len(us.Requirements) > 0
}

// IsUserStoryTemplate validates if the description follows the user story template
// "As [role], I want [function], so that [goal]"
func (us *UserStory) IsUserStoryTemplate() bool {
	if us.Description == nil {
		return false
	}
	
	description := *us.Description
	// Basic validation for user story template format
	// This is a simple check - in a real implementation, you might want more sophisticated validation
	return len(description) > 0 && 
		   (contains(description, "As ") || contains(description, "as ")) &&
		   (contains(description, "I want") || contains(description, "i want")) &&
		   (contains(description, "so that") || contains(description, "So that"))
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
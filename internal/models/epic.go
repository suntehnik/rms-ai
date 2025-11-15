package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Package-level generator instance for Epic reference IDs.
//
// This uses the production PostgreSQLReferenceIDGenerator which provides:
// - Thread-safe reference ID generation for production environments
// - PostgreSQL advisory locks for atomic generation (lock key: 2147483647)
// - UUID fallback when locks are unavailable
// - Automatic PostgreSQL vs SQLite detection
//
// For unit tests, use TestReferenceIDGenerator from reference_id_test.go instead.
// The static selection approach ensures the right generator is used in the right context.
var epicGenerator ReferenceIDGenerator = NewPostgreSQLReferenceIDGenerator(2147483647, "EP")

// GetEpicGenerator returns the current generator (for testing)
func GetEpicGenerator() ReferenceIDGenerator {
	return epicGenerator
}

// SetEpicGenerator sets a custom generator (for testing)
func SetEpicGenerator(gen ReferenceIDGenerator) {
	epicGenerator = gen
}

// Priority represents the priority level of an entity
// @Description Priority level for entities (1=Critical, 2=High, 3=Medium, 4=Low)
// @Example 1
type Priority int

const (
	PriorityCritical Priority = 1 // Critical priority - highest urgency, immediate attention required
	PriorityHigh     Priority = 2 // High priority - important, should be addressed soon
	PriorityMedium   Priority = 3 // Medium priority - normal importance, standard timeline
	PriorityLow      Priority = 4 // Low priority - nice to have, can be deferred
)

// EpicStatus represents the status of an epic in the workflow
// @Description Status of an epic in the workflow lifecycle
// @Example "Backlog"
type EpicStatus string

const (
	EpicStatusBacklog    EpicStatus = "Backlog"     // Epic is in the backlog - not yet started, awaiting prioritization
	EpicStatusDraft      EpicStatus = "Draft"       // Epic is in draft state - being defined and refined
	EpicStatusInProgress EpicStatus = "In Progress" // Epic is being actively worked on
	EpicStatusDone       EpicStatus = "Done"        // Epic is completed - all user stories finished
	EpicStatusCancelled  EpicStatus = "Cancelled"   // Epic has been cancelled - will not be implemented
)

// Epic represents a high-level feature or initiative in the requirements management system
// @Description Epic is a large body of work that can be broken down into smaller user stories. It represents a significant feature or initiative that delivers business value.
type Epic struct {
	// ID is the unique identifier for the epic
	// @Description Unique UUID identifier for the epic
	// @Example "123e4567-e89b-12d3-a456-426614174000"
	ID uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`

	// ReferenceID is the human-readable identifier for the epic
	// @Description Human-readable reference identifier (auto-generated, format: EP-XXX)
	// @Example "EP-001"
	ReferenceID string `gorm:"uniqueIndex;not null" json:"reference_id"`

	// CreatorID is the UUID of the user who created the epic
	// @Description UUID of the user who created this epic
	// @Example "123e4567-e89b-12d3-a456-426614174001"
	CreatorID uuid.UUID `gorm:"not null" json:"creator_id"`

	// AssigneeID is the UUID of the user assigned to the epic
	// @Description UUID of the user currently assigned to work on this epic
	// @Example "123e4567-e89b-12d3-a456-426614174002"
	AssigneeID uuid.UUID `gorm:"not null" json:"assignee_id"`

	// CreatedAt is the timestamp when the epic was created
	// @Description Timestamp when the epic was created (RFC3339 format)
	// @Example "2023-01-15T10:30:00Z"
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is the timestamp when the epic was last updated
	// @Description Timestamp when the epic was last modified (RFC3339 format)
	// @Example "2023-01-16T14:45:30Z"
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`

	// Priority indicates the importance level of the epic
	// @Description Priority level of the epic (1=Critical, 2=High, 3=Medium, 4=Low)
	// @Minimum 1
	// @Maximum 4
	// @Example 1
	Priority Priority `gorm:"not null" json:"priority" validate:"required,min=1,max=4"`

	// Status represents the current workflow state of the epic
	// @Description Current status of the epic in the workflow
	// @Enum Backlog,Draft,In Progress,Done,Cancelled
	// @Example "Backlog"
	Status EpicStatus `gorm:"not null" json:"status" validate:"required"`

	// Title is the name/summary of the epic
	// @Description Title or name of the epic (required, max 500 characters)
	// @MaxLength 500
	// @Example "User Authentication System"
	Title string `gorm:"not null" json:"title" validate:"required,max=500"`

	// Description provides detailed information about the epic
	// @Description Detailed description of the epic's purpose and scope (optional, max 50000 characters)
	// @MaxLength 50000
	// @Example "Implement a comprehensive user authentication and authorization system with JWT tokens, role-based access control, and secure password management."
	Description *string `json:"description,omitempty" validate:"omitempty,max=50000"`

	// Relationships - These fields are populated when explicitly requested and contain related entities

	// Creator contains the user information of who created the epic
	// @Description User who created this epic (included when explicitly preloaded)
	Creator User `gorm:"foreignKey:CreatorID;constraint:OnDelete:RESTRICT" json:"-"`

	// Assignee contains the user information of who is assigned to the epic
	// @Description User currently assigned to this epic (included when explicitly preloaded)
	Assignee User `gorm:"foreignKey:AssigneeID;constraint:OnDelete:RESTRICT" json:"-"`

	// UserStories contains all user stories that belong to this epic
	// @Description List of user stories that belong to this epic (populated when requested with ?include=user_stories)
	UserStories []UserStory `gorm:"foreignKey:EpicID;constraint:OnDelete:CASCADE" json:"user_stories,omitempty"`

	// Comments contains all comments associated with this epic
	// @Description List of comments on this epic (populated when requested with ?include=comments)
	Comments []Comment `gorm:"polymorphic:Entity;polymorphicValue:epic" json:"comments,omitempty"`

	// SteeringDocuments contains all steering documents linked to this epic
	// @Description List of steering documents linked to this epic (populated when requested with ?include=steering_documents)
	SteeringDocuments []SteeringDocument `gorm:"many2many:epic_steering_documents;" json:"steering_documents,omitempty"`
}

// BeforeCreate sets the ID if not already set and ensures default status
func (e *Epic) BeforeCreate(tx *gorm.DB) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	if e.Status == "" {
		e.Status = EpicStatusBacklog
	}

	// Generate reference ID if not set
	if e.ReferenceID == "" {
		referenceID, err := epicGenerator.Generate(tx, &Epic{})
		if err != nil {
			return err
		}
		e.ReferenceID = referenceID
	}

	return nil
}

// BeforeUpdate updates the UpdatedAt timestamp
func (e *Epic) BeforeUpdate(tx *gorm.DB) error {
	e.UpdatedAt = time.Now().UTC()
	return nil
}

// TableName returns the table name for the Epic model
func (Epic) TableName() string {
	return "epics"
}

// GetPriorityString returns the string representation of the priority
func (e *Epic) GetPriorityString() string {
	switch e.Priority {
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

// IsValidStatus checks if the provided status is valid for epics
func (e *Epic) IsValidStatus(status EpicStatus) bool {
	validStatuses := []EpicStatus{
		EpicStatusBacklog,
		EpicStatusDraft,
		EpicStatusInProgress,
		EpicStatusDone,
		EpicStatusCancelled,
	}

	for _, validStatus := range validStatuses {
		if status == validStatus {
			return true
		}
	}
	return false
}

// CanTransitionTo checks if the epic can transition to the given status
// By default, all transitions are allowed as per requirements
func (e *Epic) CanTransitionTo(newStatus EpicStatus) bool {
	return e.IsValidStatus(newStatus)
}

// HasUserStories checks if the epic has any associated user stories
func (e *Epic) HasUserStories() bool {
	return len(e.UserStories) > 0
}

// MarshalJSON implements custom JSON marshaling for Epic
// This ensures that creator and assignee objects are only included when they are actually populated
func (e *Epic) MarshalJSON() ([]byte, error) {
	type Alias Epic

	// Create a map to build the JSON response
	result := map[string]interface{}{
		"id":           e.ID,
		"reference_id": e.ReferenceID,
		"creator_id":   e.CreatorID,
		"assignee_id":  e.AssigneeID,
		"created_at":   e.CreatedAt,
		"updated_at":   e.UpdatedAt,
		"priority":     e.Priority,
		"status":       e.Status,
		"title":        e.Title,
	}

	// Only include description if it's not nil
	if e.Description != nil {
		result["description"] = *e.Description
	}

	// Only include creator if it has been populated (has a username, indicating it was preloaded)
	if e.Creator.Username != "" {
		result["creator"] = e.Creator
	}

	// Only include assignee if it has been populated (has a username, indicating it was preloaded)
	if e.Assignee.Username != "" {
		result["assignee"] = e.Assignee
	}

	// Only include user_stories if they have been populated
	if len(e.UserStories) > 0 {
		result["user_stories"] = e.UserStories
	}

	// Only include comments if they have been populated
	if len(e.Comments) > 0 {
		result["comments"] = e.Comments
	}

	// Only include steering_documents if they have been populated
	if len(e.SteeringDocuments) > 0 {
		result["steering_documents"] = e.SteeringDocuments
	}

	return json.Marshal(result)
}

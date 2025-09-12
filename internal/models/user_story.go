package models

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Package-level generator instance for UserStory reference IDs.
//
// This uses the production PostgreSQLReferenceIDGenerator which provides:
// - Thread-safe reference ID generation for production environments
// - PostgreSQL advisory locks for atomic generation (lock key: 2147483646)
// - UUID fallback when locks are unavailable
// - Automatic PostgreSQL vs SQLite detection
//
// For unit tests, use TestReferenceIDGenerator from reference_id_test.go instead.
// The static selection approach ensures the right generator is used in the right context.
var userStoryGenerator = NewPostgreSQLReferenceIDGenerator(2147483646, "US")

// UserStoryStatus represents the status of a user story in the workflow
// @Description Status of a user story in the workflow lifecycle
// @Example "Backlog"
type UserStoryStatus string

const (
	UserStoryStatusBacklog    UserStoryStatus = "Backlog"     // User story is in the backlog - not yet started, awaiting prioritization
	UserStoryStatusDraft      UserStoryStatus = "Draft"       // User story is being drafted - requirements are being defined
	UserStoryStatusInProgress UserStoryStatus = "In Progress" // User story is actively being worked on
	UserStoryStatusDone       UserStoryStatus = "Done"        // User story has been completed and meets acceptance criteria
	UserStoryStatusCancelled  UserStoryStatus = "Cancelled"   // User story has been cancelled and will not be implemented
)

// UserStory represents a user story in the requirements management system
// @Description User story is a short, simple description of a feature told from the perspective of the person who desires the new capability. It belongs to an epic and can have multiple acceptance criteria and requirements.
type UserStory struct {
	// ID is the unique identifier for the user story
	// @Description Unique UUID identifier for the user story
	// @Example "123e4567-e89b-12d3-a456-426614174000"
	ID uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`

	// ReferenceID is the human-readable identifier for the user story
	// @Description Human-readable reference identifier (auto-generated, format: US-XXX)
	// @Example "US-001"
	ReferenceID string `gorm:"uniqueIndex;not null" json:"reference_id"`

	// EpicID is the UUID of the epic this user story belongs to
	// @Description UUID of the epic that contains this user story
	// @Example "123e4567-e89b-12d3-a456-426614174001"
	EpicID uuid.UUID `gorm:"not null" json:"epic_id"`

	// CreatorID is the UUID of the user who created the user story
	// @Description UUID of the user who created this user story
	// @Example "123e4567-e89b-12d3-a456-426614174002"
	CreatorID uuid.UUID `gorm:"not null" json:"creator_id"`

	// AssigneeID is the UUID of the user assigned to the user story
	// @Description UUID of the user currently assigned to work on this user story
	// @Example "123e4567-e89b-12d3-a456-426614174003"
	AssigneeID uuid.UUID `gorm:"not null" json:"assignee_id"`

	// CreatedAt is the timestamp when the user story was created
	// @Description Timestamp when the user story was created (RFC3339 format)
	// @Example "2023-01-15T10:30:00Z"
	CreatedAt time.Time `json:"created_at"`

	// LastModified is the timestamp when the user story was last updated
	// @Description Timestamp when the user story was last modified (RFC3339 format)
	// @Example "2023-01-16T14:45:30Z"
	LastModified time.Time `json:"last_modified"`

	// Priority indicates the importance level of the user story
	// @Description Priority level of the user story (1=Critical, 2=High, 3=Medium, 4=Low)
	// @Minimum 1
	// @Maximum 4
	// @Example 2
	Priority Priority `gorm:"not null" json:"priority" validate:"required,min=1,max=4"`

	// Status represents the current workflow state of the user story
	// @Description Current status of the user story in the workflow
	// @Enum Backlog,Draft,In Progress,Done,Cancelled
	// @Example "Backlog"
	Status UserStoryStatus `gorm:"not null" json:"status" validate:"required"`

	// Title is the name/summary of the user story
	// @Description Title or name of the user story (required, max 500 characters)
	// @MaxLength 500
	// @Example "User Login with Email and Password"
	Title string `gorm:"not null" json:"title" validate:"required,max=500"`

	// Description provides detailed information about the user story
	// @Description Detailed description of the user story, preferably in the format 'As [role], I want [function], so that [goal]' (optional, max 2000 characters)
	// @MaxLength 2000
	// @Example "As a registered user, I want to log in with my email and password, so that I can access my personalized dashboard and account features."
	Description *string `json:"description,omitempty" validate:"omitempty,max=2000"`

	// Relationships
	// Epic contains the epic information this user story belongs to
	// @Description Epic that contains this user story (populated when requested with ?include=epic)
	Epic Epic `gorm:"foreignKey:EpicID;constraint:OnDelete:CASCADE" json:"epic,omitempty"`

	// Creator contains the user information of who created the user story
	// @Description User who created this user story (populated when requested with ?include=creator)
	Creator User `gorm:"foreignKey:CreatorID;constraint:OnDelete:RESTRICT" json:"creator,omitempty"`

	// Assignee contains the user information of who is assigned to the user story
	// @Description User currently assigned to this user story (populated when requested with ?include=assignee)
	Assignee User `gorm:"foreignKey:AssigneeID;constraint:OnDelete:RESTRICT" json:"assignee,omitempty"`

	// AcceptanceCriteria contains all acceptance criteria that belong to this user story
	// @Description List of acceptance criteria that define when this user story is considered complete (populated when requested with ?include=acceptance_criteria)
	AcceptanceCriteria []AcceptanceCriteria `gorm:"foreignKey:UserStoryID;constraint:OnDelete:CASCADE" json:"acceptance_criteria,omitempty"`

	// Requirements contains all requirements that belong to this user story
	// @Description List of detailed requirements that belong to this user story (populated when requested with ?include=requirements)
	Requirements []Requirement `gorm:"foreignKey:UserStoryID;constraint:OnDelete:CASCADE" json:"requirements,omitempty"`

	// Comments contains all comments associated with this user story
	// @Description List of comments on this user story (populated when requested with ?include=comments)
	Comments []Comment `gorm:"polymorphic:Entity;polymorphicValue:user_story" json:"comments,omitempty"`
}

// BeforeCreate sets the ID and ReferenceID if not already set
func (us *UserStory) BeforeCreate(tx *gorm.DB) error {
	if us.ID == uuid.Nil {
		us.ID = uuid.New()
	}
	if us.Status == "" {
		us.Status = UserStoryStatusBacklog
	}

	// Generate reference ID if not set
	if us.ReferenceID == "" {
		referenceID, err := userStoryGenerator.Generate(tx, &UserStory{})
		if err != nil {
			return err
		}
		us.ReferenceID = referenceID
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

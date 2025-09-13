package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Package-level generator instance for AcceptanceCriteria.
//
// This uses the production PostgreSQLReferenceIDGenerator which provides:
// - Thread-safe reference ID generation for production environments
// - PostgreSQL advisory locks for atomic generation (lock key: 2147483644)
// - UUID fallback when locks are unavailable
// - Automatic PostgreSQL vs SQLite detection
//
// For unit tests, use TestReferenceIDGenerator from reference_id_test.go instead.
// The static selection approach ensures the right generator is used in the right context.
var acceptanceCriteriaGenerator = NewPostgreSQLReferenceIDGenerator(2147483644, "AC")

// AcceptanceCriteria represents acceptance criteria for a user story
// @Description Testable conditions that define when a user story is considered complete and acceptable
type AcceptanceCriteria struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key" json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`                                                                                           // Unique identifier for the acceptance criteria
	ReferenceID  string    `gorm:"uniqueIndex;not null" json:"reference_id" example:"AC-001"`                                                                                                                // Human-readable reference identifier
	UserStoryID  uuid.UUID `gorm:"not null" json:"user_story_id" example:"123e4567-e89b-12d3-a456-426614174001"`                                                                                             // ID of the parent user story
	AuthorID     uuid.UUID `gorm:"not null" json:"author_id" example:"123e4567-e89b-12d3-a456-426614174002"`                                                                                                 // ID of the user who authored this acceptance criteria
	CreatedAt    time.Time `json:"created_at" example:"2023-01-01T00:00:00Z"`                                                                                                                                // Timestamp when the acceptance criteria was created
	LastModified time.Time `json:"last_modified" example:"2023-01-02T12:30:00Z"`                                                                                                                             // Timestamp when the acceptance criteria was last modified
	Description  string    `gorm:"not null" json:"description" validate:"required" example:"WHEN a user enters valid credentials THEN the system SHALL authenticate the user and redirect to the dashboard"` // EARS format description of the acceptance criteria

	// Relationships
	UserStory    UserStory     `gorm:"foreignKey:UserStoryID;constraint:OnDelete:CASCADE" json:"user_story,omitempty"`             // Parent user story that this acceptance criteria belongs to
	Author       User          `gorm:"foreignKey:AuthorID;constraint:OnDelete:RESTRICT" json:"author,omitempty"`                   // User who authored this acceptance criteria
	Requirements []Requirement `gorm:"foreignKey:AcceptanceCriteriaID;constraint:OnDelete:SET NULL" json:"requirements,omitempty"` // Requirements linked to this acceptance criteria
	Comments     []Comment     `gorm:"polymorphic:Entity;polymorphicValue:acceptance_criteria" json:"comments,omitempty"`          // Comments associated with this acceptance criteria
}

// BeforeCreate sets the ID and ReferenceID if not already set
func (ac *AcceptanceCriteria) BeforeCreate(tx *gorm.DB) error {
	if ac.ID == uuid.Nil {
		ac.ID = uuid.New()
	}
	if ac.ReferenceID == "" {
		referenceID, err := acceptanceCriteriaGenerator.Generate(tx, &AcceptanceCriteria{})
		if err != nil {
			return err
		}
		ac.ReferenceID = referenceID
	}
	now := time.Now().UTC()
	ac.CreatedAt = now
	ac.LastModified = now
	return nil
}

// BeforeUpdate updates the LastModified timestamp
func (ac *AcceptanceCriteria) BeforeUpdate(tx *gorm.DB) error {
	ac.LastModified = time.Now().UTC()
	return nil
}

// TableName returns the table name for the AcceptanceCriteria model
func (AcceptanceCriteria) TableName() string {
	return "acceptance_criteria"
}

// HasRequirements checks if the acceptance criteria has any associated requirements
func (ac *AcceptanceCriteria) HasRequirements() bool {
	return len(ac.Requirements) > 0
}

// IsEARSFormat validates if the description follows EARS format
// EARS format examples: "WHEN [event] THEN [system] SHALL [response]"
// "IF [precondition] THEN [system] SHALL [response]"
func (ac *AcceptanceCriteria) IsEARSFormat() bool {
	if ac.Description == "" {
		return false
	}

	description := ac.Description
	// Basic validation for EARS format
	// Check for common EARS keywords
	hasWhen := contains(description, "WHEN ") || contains(description, "when ")
	hasIf := contains(description, "IF ") || contains(description, "if ")
	hasThen := contains(description, "THEN ") || contains(description, "then ")
	hasShall := contains(description, "SHALL ") || contains(description, "shall ")

	// Must have either WHEN or IF, and must have THEN and SHALL
	return (hasWhen || hasIf) && hasThen && hasShall
}

// GetFormattedDescription returns the description with proper formatting
func (ac *AcceptanceCriteria) GetFormattedDescription() string {
	return ac.Description
}

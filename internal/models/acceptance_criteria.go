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
type AcceptanceCriteria struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	ReferenceID  string    `gorm:"uniqueIndex;not null" json:"reference_id"`
	UserStoryID  uuid.UUID `gorm:"not null" json:"user_story_id"`
	AuthorID     uuid.UUID `gorm:"not null" json:"author_id"`
	CreatedAt    time.Time `json:"created_at"`
	LastModified time.Time `json:"last_modified"`
	Description  string    `gorm:"not null" json:"description"`

	// Relationships
	UserStory    UserStory     `gorm:"foreignKey:UserStoryID;constraint:OnDelete:CASCADE" json:"user_story,omitempty"`
	Author       User          `gorm:"foreignKey:AuthorID;constraint:OnDelete:RESTRICT" json:"author,omitempty"`
	Requirements []Requirement `gorm:"foreignKey:AcceptanceCriteriaID;constraint:OnDelete:SET NULL" json:"requirements,omitempty"`
	Comments     []Comment     `gorm:"polymorphic:Entity;polymorphicValue:acceptance_criteria" json:"comments,omitempty"`
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

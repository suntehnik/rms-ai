package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Package-level generator instance for requirements.
//
// This uses the production PostgreSQLReferenceIDGenerator which provides:
// - Thread-safe reference ID generation for production environments
// - PostgreSQL advisory locks for atomic generation (lock key: 2147483645)
// - UUID fallback when locks are unavailable
// - Automatic PostgreSQL vs SQLite detection
//
// For unit tests, use TestReferenceIDGenerator from reference_id_test.go instead.
// The static selection approach ensures the right generator is used in the right context.
var requirementGenerator ReferenceIDGenerator = NewPostgreSQLReferenceIDGenerator(2147483645, "REQ")

// GetRequirementGenerator returns the current generator (for testing)
func GetRequirementGenerator() ReferenceIDGenerator {
	return requirementGenerator
}

// SetRequirementGenerator sets a custom generator (for testing)
func SetRequirementGenerator(gen ReferenceIDGenerator) {
	requirementGenerator = gen
}

// RequirementStatus represents the status of a requirement
// @Description Status of a requirement in the workflow lifecycle
// @Example "Draft"
type RequirementStatus string

const (
	RequirementStatusDraft    RequirementStatus = "Draft"    // Draft - requirement is being written and refined
	RequirementStatusActive   RequirementStatus = "Active"   // Active - requirement is approved and being implemented
	RequirementStatusObsolete RequirementStatus = "Obsolete" // Obsolete - requirement is no longer needed or has been superseded
)

// Requirement represents a detailed requirement in the system
// @Description A detailed requirement that specifies what needs to be implemented within a user story
type Requirement struct {
	ID                   uuid.UUID         `gorm:"type:uuid;primary_key" json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`                                                                                                                                                                            // Unique identifier for the requirement
	ReferenceID          string            `gorm:"uniqueIndex;not null" json:"reference_id" example:"REQ-001"`                                                                                                                                                                                                // Human-readable reference identifier
	UserStoryID          uuid.UUID         `gorm:"not null" json:"user_story_id" example:"123e4567-e89b-12d3-a456-426614174001"`                                                                                                                                                                              // ID of the parent user story
	AcceptanceCriteriaID *uuid.UUID        `json:"acceptance_criteria_id" example:"123e4567-e89b-12d3-a456-426614174002"`                                                                                                                                                                                     // Optional ID of linked acceptance criteria
	CreatorID            uuid.UUID         `gorm:"not null" json:"creator_id" example:"123e4567-e89b-12d3-a456-426614174003"`                                                                                                                                                                                 // ID of the user who created the requirement
	AssigneeID           uuid.UUID         `gorm:"not null" json:"assignee_id" example:"123e4567-e89b-12d3-a456-426614174004"`                                                                                                                                                                                // ID of the user assigned to implement the requirement
	CreatedAt            time.Time         `json:"created_at" example:"2023-01-01T00:00:00Z"`                                                                                                                                                                                                                 // Timestamp when the requirement was created
	UpdatedAt            time.Time         `json:"updated_at" db:"updated_at" example:"2023-01-02T12:30:00Z"`                                                                                                                                                                                                 // Timestamp when the requirement was last updated
	Priority             Priority          `gorm:"not null" json:"priority" validate:"required,min=1,max=4" example:"2"`                                                                                                                                                                                      // Priority level (1=Critical, 2=High, 3=Medium, 4=Low)
	Status               RequirementStatus `gorm:"not null" json:"status" validate:"required" example:"Draft"`                                                                                                                                                                                                // Current status of the requirement
	TypeID               uuid.UUID         `gorm:"not null" json:"type_id" example:"123e4567-e89b-12d3-a456-426614174005"`                                                                                                                                                                                    // ID of the requirement type (Functional, Non-Functional, etc.)
	Title                string            `gorm:"not null" json:"title" validate:"required,max=500" example:"User authentication must support OAuth 2.0"`                                                                                                                                                    // Brief title describing the requirement
	Description          *string           `json:"description" validate:"omitempty,max=50000" example:"The system shall support OAuth 2.0 authentication flow with support for Google, GitHub, and Microsoft providers. The implementation must handle token refresh and provide secure session management."` // Detailed description of the requirement

	// Relationships - These fields are populated when explicitly preloaded and included in JSON via custom MarshalJSON
	// @Description Parent user story containing this requirement (included only when preloaded via repository methods)
	UserStory UserStory `gorm:"foreignKey:UserStoryID;constraint:OnDelete:CASCADE" json:"-"`
	// @Description Optional linked acceptance criteria (included only when preloaded via repository methods)
	AcceptanceCriteria *AcceptanceCriteria `gorm:"foreignKey:AcceptanceCriteriaID;constraint:OnDelete:SET NULL" json:"-"`
	// @Description User who created this requirement (included only when preloaded via repository methods)
	Creator User `gorm:"foreignKey:CreatorID;constraint:OnDelete:RESTRICT" json:"-"`
	// @Description User assigned to implement this requirement (included only when preloaded via repository methods)
	Assignee User `gorm:"foreignKey:AssigneeID;constraint:OnDelete:RESTRICT" json:"-"`
	// @Description Type classification of this requirement (included only when preloaded via repository methods)
	Type RequirementType `gorm:"foreignKey:TypeID;constraint:OnDelete:RESTRICT" json:"-"`
	// @Description Relationships where this requirement is the source (included only when preloaded)
	SourceRelationships []RequirementRelationship `gorm:"foreignKey:SourceRequirementID;constraint:OnDelete:CASCADE" json:"source_relationships,omitempty"`
	// @Description Relationships where this requirement is the target (included only when preloaded)
	TargetRelationships []RequirementRelationship `gorm:"foreignKey:TargetRequirementID;constraint:OnDelete:CASCADE" json:"target_relationships,omitempty"`
	// @Description Comments associated with this requirement (included only when preloaded)
	Comments []Comment `gorm:"polymorphic:Entity;polymorphicValue:requirement" json:"comments,omitempty"`
}

// BeforeCreate sets the ID if not already set and ensures default status
func (r *Requirement) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	if r.Status == "" {
		r.Status = RequirementStatusDraft
	}

	// Generate reference ID if not set
	if r.ReferenceID == "" {
		referenceID, err := requirementGenerator.Generate(tx, &Requirement{})
		if err != nil {
			return err
		}
		r.ReferenceID = referenceID
	}

	return nil
}

// BeforeUpdate updates the UpdatedAt timestamp
func (r *Requirement) BeforeUpdate(tx *gorm.DB) error {
	r.UpdatedAt = time.Now().UTC()
	return nil
}

// TableName returns the table name for the Requirement model
func (Requirement) TableName() string {
	return "requirements"
}

// GetPriorityString returns the string representation of the priority
func (r *Requirement) GetPriorityString() string {
	switch r.Priority {
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

// IsValidStatus checks if the provided status is valid for requirements
func (r *Requirement) IsValidStatus(status RequirementStatus) bool {
	validStatuses := []RequirementStatus{
		RequirementStatusDraft,
		RequirementStatusActive,
		RequirementStatusObsolete,
	}

	for _, validStatus := range validStatuses {
		if status == validStatus {
			return true
		}
	}
	return false
}

// CanTransitionTo checks if the requirement can transition to the given status
// By default, all transitions are allowed as per requirements
func (r *Requirement) CanTransitionTo(newStatus RequirementStatus) bool {
	return r.IsValidStatus(newStatus)
}

// HasRelationships checks if the requirement has any relationships
func (r *Requirement) HasRelationships() bool {
	return len(r.SourceRelationships) > 0 || len(r.TargetRelationships) > 0
}

// GetAllRelationships returns all relationships (both source and target)
func (r *Requirement) GetAllRelationships() []RequirementRelationship {
	var allRelationships []RequirementRelationship
	allRelationships = append(allRelationships, r.SourceRelationships...)
	allRelationships = append(allRelationships, r.TargetRelationships...)
	return allRelationships
}

// IsLinkedToAcceptanceCriteria checks if the requirement is linked to acceptance criteria
func (r *Requirement) IsLinkedToAcceptanceCriteria() bool {
	return r.AcceptanceCriteriaID != nil
}

// MarshalJSON implements custom JSON marshaling for Requirement
// This ensures that related objects are only included when they are actually populated
func (r *Requirement) MarshalJSON() ([]byte, error) {
	// Create a map to build the JSON response
	result := map[string]interface{}{
		"id":            r.ID,
		"reference_id":  r.ReferenceID,
		"user_story_id": r.UserStoryID,
		"creator_id":    r.CreatorID,
		"assignee_id":   r.AssigneeID,
		"created_at":    r.CreatedAt,
		"updated_at":    r.UpdatedAt,
		"priority":      r.Priority,
		"status":        r.Status,
		"type_id":       r.TypeID,
		"title":         r.Title,
	}

	// Only include acceptance_criteria_id if it's not nil
	if r.AcceptanceCriteriaID != nil {
		result["acceptance_criteria_id"] = *r.AcceptanceCriteriaID
	}

	// Only include description if it's not nil
	if r.Description != nil {
		result["description"] = *r.Description
	}

	// Only include user_story if it has been populated (has a title, indicating it was preloaded)
	if r.UserStory.Title != "" {
		result["user_story"] = r.UserStory
	}

	// Only include acceptance_criteria if it has been populated and is not nil
	if r.AcceptanceCriteria != nil && r.AcceptanceCriteria.Description != "" {
		result["acceptance_criteria"] = r.AcceptanceCriteria
	}

	// Only include creator if it has been populated (has a username, indicating it was preloaded)
	if r.Creator.Username != "" {
		result["creator"] = r.Creator
	}

	// Only include assignee if it has been populated (has a username, indicating it was preloaded)
	if r.Assignee.Username != "" {
		result["assignee"] = r.Assignee
	}

	// Only include type if it has been populated (has a name, indicating it was preloaded)
	if r.Type.Name != "" {
		result["type"] = r.Type
	}

	// Only include source_relationships if they have been populated
	if len(r.SourceRelationships) > 0 {
		result["source_relationships"] = r.SourceRelationships
	}

	// Only include target_relationships if they have been populated
	if len(r.TargetRelationships) > 0 {
		result["target_relationships"] = r.TargetRelationships
	}

	// Only include comments if they have been populated
	if len(r.Comments) > 0 {
		result["comments"] = r.Comments
	}

	return json.Marshal(result)
}

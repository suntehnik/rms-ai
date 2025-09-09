package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RequirementStatus represents the status of a requirement
type RequirementStatus string

const (
	RequirementStatusDraft    RequirementStatus = "Draft"
	RequirementStatusActive   RequirementStatus = "Active"
	RequirementStatusObsolete RequirementStatus = "Obsolete"
)

// Requirement represents a detailed requirement in the system
type Requirement struct {
	ID                   uuid.UUID          `gorm:"type:uuid;primary_key" json:"id"`
	ReferenceID          string             `gorm:"uniqueIndex;not null" json:"reference_id"`
	UserStoryID          uuid.UUID          `gorm:"not null" json:"user_story_id"`
	AcceptanceCriteriaID *uuid.UUID         `json:"acceptance_criteria_id"`
	CreatorID            uuid.UUID          `gorm:"not null" json:"creator_id"`
	AssigneeID           uuid.UUID          `gorm:"not null" json:"assignee_id"`
	CreatedAt            time.Time          `json:"created_at"`
	LastModified         time.Time          `json:"last_modified"`
	Priority             Priority           `gorm:"not null" json:"priority"`
	Status               RequirementStatus  `gorm:"not null" json:"status"`
	TypeID               uuid.UUID          `gorm:"not null" json:"type_id"`
	Title                string             `gorm:"not null" json:"title"`
	Description          *string            `json:"description"`

	// Relationships
	UserStory            UserStory                   `gorm:"foreignKey:UserStoryID;constraint:OnDelete:CASCADE" json:"user_story,omitempty"`
	AcceptanceCriteria   *AcceptanceCriteria         `gorm:"foreignKey:AcceptanceCriteriaID;constraint:OnDelete:SET NULL" json:"acceptance_criteria,omitempty"`
	Creator              User                        `gorm:"foreignKey:CreatorID;constraint:OnDelete:RESTRICT" json:"creator,omitempty"`
	Assignee             User                        `gorm:"foreignKey:AssigneeID;constraint:OnDelete:RESTRICT" json:"assignee,omitempty"`
	Type                 RequirementType             `gorm:"foreignKey:TypeID;constraint:OnDelete:RESTRICT" json:"type,omitempty"`
	SourceRelationships  []RequirementRelationship   `gorm:"foreignKey:SourceRequirementID;constraint:OnDelete:CASCADE" json:"source_relationships,omitempty"`
	TargetRelationships  []RequirementRelationship   `gorm:"foreignKey:TargetRequirementID;constraint:OnDelete:CASCADE" json:"target_relationships,omitempty"`
	Comments             []Comment                   `gorm:"polymorphic:Entity;polymorphicValue:requirement" json:"comments,omitempty"`
}

// BeforeCreate sets the ID if not already set and ensures default status
func (r *Requirement) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	if r.Status == "" {
		r.Status = RequirementStatusDraft
	}
	
	// Reference ID generation is now handled at the repository level
	// to properly handle concurrency and retries
	
	return nil
}

// BeforeUpdate updates the LastModified timestamp
func (r *Requirement) BeforeUpdate(tx *gorm.DB) error {
	r.LastModified = time.Now().UTC()
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
package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RequirementRelationship represents a relationship between two requirements
type RequirementRelationship struct {
	ID                  uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	SourceRequirementID uuid.UUID `gorm:"not null" json:"source_requirement_id"`
	TargetRequirementID uuid.UUID `gorm:"not null" json:"target_requirement_id"`
	RelationshipTypeID  uuid.UUID `gorm:"not null" json:"relationship_type_id"`
	CreatedAt           time.Time `json:"created_at"`
	CreatedBy           uuid.UUID `gorm:"not null" json:"created_by"`

	// Relationships
	SourceRequirement Requirement      `gorm:"foreignKey:SourceRequirementID;constraint:OnDelete:CASCADE" json:"source_requirement,omitempty"`
	TargetRequirement Requirement      `gorm:"foreignKey:TargetRequirementID;constraint:OnDelete:CASCADE" json:"target_requirement,omitempty"`
	RelationshipType  RelationshipType `gorm:"foreignKey:RelationshipTypeID;constraint:OnDelete:RESTRICT" json:"relationship_type,omitempty"`
	Creator           User             `gorm:"foreignKey:CreatedBy;constraint:OnDelete:RESTRICT" json:"creator,omitempty"`
}

// BeforeCreate sets the ID if not already set and validates the relationship
func (rr *RequirementRelationship) BeforeCreate(tx *gorm.DB) error {
	if rr.ID == uuid.Nil {
		rr.ID = uuid.New()
	}

	// Validate that source and target requirements are different
	if rr.SourceRequirementID == rr.TargetRequirementID {
		return gorm.ErrInvalidData
	}

	return nil
}

// TableName returns the table name for the RequirementRelationship model
func (RequirementRelationship) TableName() string {
	return "requirement_relationships"
}

// IsValid checks if the relationship is valid (source != target)
func (rr *RequirementRelationship) IsValid() bool {
	return rr.SourceRequirementID != rr.TargetRequirementID
}

// GetRelationshipDescription returns a human-readable description of the relationship
func (rr *RequirementRelationship) GetRelationshipDescription() string {
	if rr.RelationshipType.Name == "" {
		return "Unknown relationship"
	}

	switch rr.RelationshipType.Name {
	case "depends_on":
		return "depends on"
	case "blocks":
		return "blocks"
	case "relates_to":
		return "relates to"
	case "conflicts_with":
		return "conflicts with"
	case "derives_from":
		return "derives from"
	default:
		return rr.RelationshipType.Name
	}
}

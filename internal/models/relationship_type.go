package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RelationshipType represents a configurable type of relationship between requirements
type RelationshipType struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	Name        string    `gorm:"uniqueIndex;not null" json:"name"`
	Description *string   `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Relationships
	RequirementRelationships []RequirementRelationship `gorm:"foreignKey:RelationshipTypeID;constraint:OnDelete:RESTRICT" json:"requirement_relationships,omitempty"`
}

// BeforeCreate sets the ID if not already set
func (rt *RelationshipType) BeforeCreate(tx *gorm.DB) error {
	if rt.ID == uuid.Nil {
		rt.ID = uuid.New()
	}
	return nil
}

// TableName returns the table name for the RelationshipType model
func (RelationshipType) TableName() string {
	return "relationship_types"
}

// HasRelationships checks if the relationship type has any associated relationships
func (rt *RelationshipType) HasRelationships() bool {
	return len(rt.RequirementRelationships) > 0
}

// GetDefaultRelationshipTypes returns the default relationship types that should be created
func GetDefaultRelationshipTypes() []RelationshipType {
	return []RelationshipType{
		{
			Name:        "depends_on",
			Description: stringPtr("This requirement depends on another requirement"),
		},
		{
			Name:        "blocks",
			Description: stringPtr("This requirement blocks another requirement"),
		},
		{
			Name:        "relates_to",
			Description: stringPtr("This requirement is related to another requirement"),
		},
		{
			Name:        "conflicts_with",
			Description: stringPtr("This requirement conflicts with another requirement"),
		},
		{
			Name:        "derives_from",
			Description: stringPtr("This requirement is derived from another requirement"),
		},
	}
}

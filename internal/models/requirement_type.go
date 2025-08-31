package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RequirementType represents a configurable type of requirement
type RequirementType struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	Name        string    `gorm:"uniqueIndex;not null" json:"name"`
	Description *string   `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Relationships
	Requirements []Requirement `gorm:"foreignKey:TypeID;constraint:OnDelete:RESTRICT" json:"requirements,omitempty"`
}

// BeforeCreate sets the ID if not already set
func (rt *RequirementType) BeforeCreate(tx *gorm.DB) error {
	if rt.ID == uuid.Nil {
		rt.ID = uuid.New()
	}
	return nil
}

// TableName returns the table name for the RequirementType model
func (RequirementType) TableName() string {
	return "requirement_types"
}

// HasRequirements checks if the requirement type has any associated requirements
func (rt *RequirementType) HasRequirements() bool {
	return len(rt.Requirements) > 0
}

// GetDefaultRequirementTypes returns the default requirement types that should be created
func GetDefaultRequirementTypes() []RequirementType {
	return []RequirementType{
		{
			Name:        "Functional",
			Description: stringPtr("Functional requirements that describe what the system should do"),
		},
		{
			Name:        "Non-Functional",
			Description: stringPtr("Non-functional requirements that describe how the system should behave"),
		},
		{
			Name:        "Business Rule",
			Description: stringPtr("Business rules and constraints"),
		},
		{
			Name:        "Interface",
			Description: stringPtr("Interface and integration requirements"),
		},
		{
			Name:        "Data",
			Description: stringPtr("Data and information requirements"),
		},
	}
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}
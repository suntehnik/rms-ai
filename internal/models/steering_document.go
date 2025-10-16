package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Package-level generator instance for SteeringDocument reference IDs.
//
// This uses the production PostgreSQLReferenceIDGenerator which provides:
// - Thread-safe reference ID generation for production environments
// - PostgreSQL advisory locks for atomic generation (lock key: 2147483643)
// - UUID fallback when locks are unavailable
// - Automatic PostgreSQL vs SQLite detection
//
// For unit tests, use TestReferenceIDGenerator from reference_id_test.go instead.
// The static selection approach ensures the right generator is used in the right context.
var steeringDocumentGenerator = NewPostgreSQLReferenceIDGenerator(2147483643, "STD")

// SteeringDocument represents a steering document in the system
// @Description Steering document contains instructions, standards and team norms that can be linked to epics for additional context
type SteeringDocument struct {
	// ID is the unique identifier for the steering document
	// @Description Unique UUID identifier for the steering document
	// @Example "123e4567-e89b-12d3-a456-426614174000"
	ID uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`

	// ReferenceID is the human-readable identifier for the steering document
	// @Description Human-readable reference identifier (auto-generated, format: STD-XXX)
	// @Example "STD-001"
	ReferenceID string `gorm:"uniqueIndex;not null" json:"reference_id"`

	// Title is the name/summary of the steering document
	// @Description Title or name of the steering document (required, max 500 characters)
	// @MaxLength 500
	// @Example "Code Review Standards"
	Title string `gorm:"not null" json:"title" validate:"required,max=500"`

	// Description provides detailed information about the steering document
	// @Description Detailed description of the steering document content (optional, max 50000 characters)
	// @MaxLength 50000
	// @Example "This document outlines the code review standards and practices for the development team..."
	Description *string `json:"description,omitempty" validate:"omitempty,max=50000"`

	// CreatorID is the UUID of the user who created the steering document
	// @Description UUID of the user who created this steering document
	// @Example "123e4567-e89b-12d3-a456-426614174001"
	CreatorID uuid.UUID `gorm:"not null" json:"creator_id"`

	// CreatedAt is the timestamp when the steering document was created
	// @Description Timestamp when the steering document was created (RFC3339 format)
	// @Example "2023-01-15T10:30:00Z"
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is the timestamp when the steering document was last updated
	// @Description Timestamp when the steering document was last modified (RFC3339 format)
	// @Example "2023-01-16T14:45:30Z"
	UpdatedAt time.Time `json:"updated_at"`

	// Relationships - These fields are populated when explicitly requested and contain related entities

	// Creator contains the user information of who created the steering document
	// @Description User who created this steering document (included when explicitly preloaded)
	Creator User `gorm:"foreignKey:CreatorID;constraint:OnDelete:RESTRICT" json:"-"`

	// Epics contains all epics that are linked to this steering document
	// @Description List of epics linked to this steering document (populated when requested with ?include=epics)
	Epics []Epic `gorm:"many2many:epic_steering_documents;" json:"epics,omitempty"`
}

// BeforeCreate sets the ID if not already set and generates reference ID
func (sd *SteeringDocument) BeforeCreate(tx *gorm.DB) error {
	if sd.ID == uuid.Nil {
		sd.ID = uuid.New()
	}

	// Generate reference ID if not set
	if sd.ReferenceID == "" {
		referenceID, err := steeringDocumentGenerator.Generate(tx, &SteeringDocument{})
		if err != nil {
			return err
		}
		sd.ReferenceID = referenceID
	}

	return nil
}

// BeforeUpdate updates the UpdatedAt timestamp
func (sd *SteeringDocument) BeforeUpdate(tx *gorm.DB) error {
	sd.UpdatedAt = time.Now().UTC()
	return nil
}

// TableName returns the table name for the SteeringDocument model
func (SteeringDocument) TableName() string {
	return "steering_documents"
}

// MarshalJSON implements custom JSON marshaling for SteeringDocument
// This ensures that creator and epics objects are only included when they are actually populated
func (sd *SteeringDocument) MarshalJSON() ([]byte, error) {
	// Create a map to build the JSON response
	result := map[string]interface{}{
		"id":           sd.ID,
		"reference_id": sd.ReferenceID,
		"title":        sd.Title,
		"creator_id":   sd.CreatorID,
		"created_at":   sd.CreatedAt,
		"updated_at":   sd.UpdatedAt,
	}

	// Only include description if it's not nil
	if sd.Description != nil {
		result["description"] = *sd.Description
	}

	// Only include creator if it has been populated (has a username, indicating it was preloaded)
	if sd.Creator.Username != "" {
		result["creator"] = sd.Creator
	}

	// Only include epics if they have been populated
	if len(sd.Epics) > 0 {
		result["epics"] = sd.Epics
	}

	return json.Marshal(result)
}

// EpicSteeringDocument represents the many-to-many relationship between epics and steering documents
// @Description Junction table linking epics to steering documents
type EpicSteeringDocument struct {
	// ID is the unique identifier for the relationship
	// @Description Unique UUID identifier for the epic-steering document relationship
	// @Example "123e4567-e89b-12d3-a456-426614174002"
	ID uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`

	// EpicID is the UUID of the linked epic
	// @Description UUID of the epic in this relationship
	// @Example "123e4567-e89b-12d3-a456-426614174003"
	EpicID uuid.UUID `gorm:"not null" json:"epic_id"`

	// SteeringDocumentID is the UUID of the linked steering document
	// @Description UUID of the steering document in this relationship
	// @Example "123e4567-e89b-12d3-a456-426614174004"
	SteeringDocumentID uuid.UUID `gorm:"not null" json:"steering_document_id"`

	// CreatedAt is the timestamp when the relationship was created
	// @Description Timestamp when the relationship was created (RFC3339 format)
	// @Example "2023-01-15T10:30:00Z"
	CreatedAt time.Time `json:"created_at"`

	// Relationships

	// Epic contains the epic information for this relationship
	// @Description Epic linked in this relationship (populated when preloaded)
	Epic Epic `gorm:"foreignKey:EpicID;constraint:OnDelete:CASCADE" json:"epic,omitempty"`

	// SteeringDocument contains the steering document information for this relationship
	// @Description Steering document linked in this relationship (populated when preloaded)
	SteeringDocument SteeringDocument `gorm:"foreignKey:SteeringDocumentID;constraint:OnDelete:CASCADE" json:"steering_document,omitempty"`
}

// BeforeCreate sets the ID if not already set
func (esd *EpicSteeringDocument) BeforeCreate(tx *gorm.DB) error {
	if esd.ID == uuid.Nil {
		esd.ID = uuid.New()
	}
	return nil
}

// TableName returns the table name for the EpicSteeringDocument model
func (EpicSteeringDocument) TableName() string {
	return "epic_steering_documents"
}

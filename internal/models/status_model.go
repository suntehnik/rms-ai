package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)



// StatusModel represents a configurable status model for different entity types
type StatusModel struct {
	ID          uuid.UUID  `gorm:"type:uuid;primary_key" json:"id"`
	EntityType  EntityType `gorm:"not null;uniqueIndex:idx_status_model_entity_name" json:"entity_type"`
	Name        string     `gorm:"not null;uniqueIndex:idx_status_model_entity_name" json:"name"`
	Description *string    `json:"description"`
	IsDefault   bool       `gorm:"not null;default:false" json:"is_default"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`

	// Relationships
	Statuses    []Status           `gorm:"foreignKey:StatusModelID;constraint:OnDelete:CASCADE" json:"statuses,omitempty"`
	Transitions []StatusTransition `gorm:"foreignKey:StatusModelID;constraint:OnDelete:CASCADE" json:"transitions,omitempty"`
}

// Status represents an individual status within a status model
type Status struct {
	ID            uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	StatusModelID uuid.UUID `gorm:"not null" json:"status_model_id"`
	Name          string    `gorm:"not null" json:"name"`
	Description   *string   `json:"description"`
	Color         *string   `json:"color"` // Hex color code for UI display
	IsInitial     bool      `gorm:"not null;default:false" json:"is_initial"`
	IsFinal       bool      `gorm:"not null;default:false" json:"is_final"`
	Order         int       `gorm:"not null;default:0" json:"order"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	// Relationships
	StatusModel       StatusModel        `gorm:"foreignKey:StatusModelID;constraint:OnDelete:CASCADE" json:"status_model,omitempty"`
	FromTransitions   []StatusTransition `gorm:"foreignKey:FromStatusID;constraint:OnDelete:CASCADE" json:"from_transitions,omitempty"`
	ToTransitions     []StatusTransition `gorm:"foreignKey:ToStatusID;constraint:OnDelete:CASCADE" json:"to_transitions,omitempty"`
}

// StatusTransition represents allowed transitions between statuses
type StatusTransition struct {
	ID            uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	StatusModelID uuid.UUID `gorm:"not null" json:"status_model_id"`
	FromStatusID  uuid.UUID `gorm:"not null" json:"from_status_id"`
	ToStatusID    uuid.UUID `gorm:"not null" json:"to_status_id"`
	Name          *string   `json:"name"`          // Optional name for the transition
	Description   *string   `json:"description"`   // Optional description
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	// Relationships
	StatusModel StatusModel `gorm:"foreignKey:StatusModelID;constraint:OnDelete:CASCADE" json:"status_model,omitempty"`
	FromStatus  Status      `gorm:"foreignKey:FromStatusID;constraint:OnDelete:CASCADE" json:"from_status,omitempty"`
	ToStatus    Status      `gorm:"foreignKey:ToStatusID;constraint:OnDelete:CASCADE" json:"to_status,omitempty"`
}

// BeforeCreate sets the ID if not already set
func (sm *StatusModel) BeforeCreate(tx *gorm.DB) error {
	if sm.ID == uuid.Nil {
		sm.ID = uuid.New()
	}
	return nil
}

// BeforeCreate sets the ID if not already set
func (s *Status) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

// BeforeCreate sets the ID if not already set
func (st *StatusTransition) BeforeCreate(tx *gorm.DB) error {
	if st.ID == uuid.Nil {
		st.ID = uuid.New()
	}
	return nil
}

// TableName returns the table name for the StatusModel model
func (StatusModel) TableName() string {
	return "status_models"
}

// TableName returns the table name for the Status model
func (Status) TableName() string {
	return "statuses"
}

// TableName returns the table name for the StatusTransition model
func (StatusTransition) TableName() string {
	return "status_transitions"
}

// GetDefaultStatusModels returns the default status models for all entity types
func GetDefaultStatusModels() []StatusModel {
	return []StatusModel{
		{
			EntityType:  EntityTypeEpic,
			Name:        "Default Epic Workflow",
			Description: stringPtr("Default status workflow for epics"),
			IsDefault:   true,
		},
		{
			EntityType:  EntityTypeUserStory,
			Name:        "Default User Story Workflow",
			Description: stringPtr("Default status workflow for user stories"),
			IsDefault:   true,
		},
		{
			EntityType:  EntityTypeRequirement,
			Name:        "Default Requirement Workflow",
			Description: stringPtr("Default status workflow for requirements"),
			IsDefault:   true,
		},
	}
}

// GetDefaultStatusesForEpic returns the default statuses for epic entities
func GetDefaultStatusesForEpic() []Status {
	return []Status{
		{Name: "Backlog", Description: stringPtr("Epic is in the backlog"), Color: stringPtr("#6c757d"), IsInitial: true, Order: 1},
		{Name: "Draft", Description: stringPtr("Epic is being drafted"), Color: stringPtr("#ffc107"), Order: 2},
		{Name: "In Progress", Description: stringPtr("Epic is in progress"), Color: stringPtr("#007bff"), Order: 3},
		{Name: "Done", Description: stringPtr("Epic is completed"), Color: stringPtr("#28a745"), IsFinal: true, Order: 4},
		{Name: "Cancelled", Description: stringPtr("Epic has been cancelled"), Color: stringPtr("#dc3545"), IsFinal: true, Order: 5},
	}
}

// GetDefaultStatusesForUserStory returns the default statuses for user story entities
func GetDefaultStatusesForUserStory() []Status {
	return []Status{
		{Name: "Backlog", Description: stringPtr("User story is in the backlog"), Color: stringPtr("#6c757d"), IsInitial: true, Order: 1},
		{Name: "Draft", Description: stringPtr("User story is being drafted"), Color: stringPtr("#ffc107"), Order: 2},
		{Name: "In Progress", Description: stringPtr("User story is in progress"), Color: stringPtr("#007bff"), Order: 3},
		{Name: "Done", Description: stringPtr("User story is completed"), Color: stringPtr("#28a745"), IsFinal: true, Order: 4},
		{Name: "Cancelled", Description: stringPtr("User story has been cancelled"), Color: stringPtr("#dc3545"), IsFinal: true, Order: 5},
	}
}

// GetDefaultStatusesForRequirement returns the default statuses for requirement entities
func GetDefaultStatusesForRequirement() []Status {
	return []Status{
		{Name: "Draft", Description: stringPtr("Requirement is being drafted"), Color: stringPtr("#ffc107"), IsInitial: true, Order: 1},
		{Name: "Active", Description: stringPtr("Requirement is active"), Color: stringPtr("#28a745"), Order: 2},
		{Name: "Obsolete", Description: stringPtr("Requirement is obsolete"), Color: stringPtr("#6c757d"), IsFinal: true, Order: 3},
	}
}



// IsValidEntityType checks if the entity type is valid
func IsValidEntityType(entityType EntityType) bool {
	validTypes := []EntityType{
		EntityTypeEpic,
		EntityTypeUserStory,
		EntityTypeAcceptanceCriteria,
		EntityTypeRequirement,
	}
	
	for _, validType := range validTypes {
		if entityType == validType {
			return true
		}
	}
	return false
}

// GetInitialStatus returns the initial status for the status model
func (sm *StatusModel) GetInitialStatus() *Status {
	for _, status := range sm.Statuses {
		if status.IsInitial {
			return &status
		}
	}
	return nil
}

// GetFinalStatuses returns all final statuses for the status model
func (sm *StatusModel) GetFinalStatuses() []Status {
	var finalStatuses []Status
	for _, status := range sm.Statuses {
		if status.IsFinal {
			finalStatuses = append(finalStatuses, status)
		}
	}
	return finalStatuses
}

// CanTransitionTo checks if a transition is allowed from one status to another
func (sm *StatusModel) CanTransitionTo(fromStatusID, toStatusID uuid.UUID) bool {
	// If no transitions are defined, allow all transitions (default behavior)
	if len(sm.Transitions) == 0 {
		return true
	}
	
	// Check if there's an explicit transition defined
	for _, transition := range sm.Transitions {
		if transition.FromStatusID == fromStatusID && transition.ToStatusID == toStatusID {
			return true
		}
	}
	
	return false
}

// GetAvailableTransitions returns all available transitions from a given status
func (sm *StatusModel) GetAvailableTransitions(fromStatusID uuid.UUID) []StatusTransition {
	var availableTransitions []StatusTransition
	
	// If no transitions are defined, return empty slice (all transitions allowed)
	if len(sm.Transitions) == 0 {
		return availableTransitions
	}
	
	for _, transition := range sm.Transitions {
		if transition.FromStatusID == fromStatusID {
			availableTransitions = append(availableTransitions, transition)
		}
	}
	
	return availableTransitions
}
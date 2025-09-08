package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Priority represents the priority level of an entity
type Priority int

const (
	PriorityCritical Priority = 1
	PriorityHigh     Priority = 2
	PriorityMedium   Priority = 3
	PriorityLow      Priority = 4
)

// EpicStatus represents the status of an epic
type EpicStatus string

const (
	EpicStatusBacklog    EpicStatus = "Backlog"
	EpicStatusDraft      EpicStatus = "Draft"
	EpicStatusInProgress EpicStatus = "In Progress"
	EpicStatusDone       EpicStatus = "Done"
	EpicStatusCancelled  EpicStatus = "Cancelled"
)

// Epic represents an epic in the requirements management system
type Epic struct {
	ID           uuid.UUID  `gorm:"type:uuid;primary_key" json:"id"`
	ReferenceID  string     `gorm:"uniqueIndex;not null" json:"reference_id"`
	CreatorID    uuid.UUID  `gorm:"not null" json:"creator_id"`
	AssigneeID   uuid.UUID  `gorm:"not null" json:"assignee_id"`
	CreatedAt    time.Time  `json:"created_at"`
	LastModified time.Time  `json:"last_modified"`
	Priority     Priority   `gorm:"not null" json:"priority"`
	Status       EpicStatus `gorm:"not null" json:"status"`
	Title        string     `gorm:"not null" json:"title"`
	Description  *string    `json:"description"`

	// Relationships
	Creator     User        `gorm:"foreignKey:CreatorID;constraint:OnDelete:RESTRICT" json:"creator,omitempty"`
	Assignee    User        `gorm:"foreignKey:AssigneeID;constraint:OnDelete:RESTRICT" json:"assignee,omitempty"`
	UserStories []UserStory `gorm:"foreignKey:EpicID;constraint:OnDelete:CASCADE" json:"user_stories,omitempty"`
	Comments    []Comment   `gorm:"polymorphic:Entity;polymorphicValue:epic" json:"comments,omitempty"`
}

// BeforeCreate sets the ID if not already set and ensures default status
func (e *Epic) BeforeCreate(tx *gorm.DB) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	if e.Status == "" {
		e.Status = EpicStatusBacklog
	}
	
	// Generate reference ID if not set
	if e.ReferenceID == "" {
		// Check if we're using PostgreSQL for advisory locks
		if tx.Dialector.Name() == "postgres" {
			// Use PostgreSQL advisory lock for atomic reference ID generation
			// Lock key: fixed value for epic reference ID generation
			lockKey := int64(2147483647) // Fixed key for epic reference ID generation
			
			// Acquire advisory lock (automatically released at transaction end)
			var lockAcquired bool
			if err := tx.Raw("SELECT pg_try_advisory_xact_lock(?)", lockKey).Scan(&lockAcquired).Error; err != nil {
				return fmt.Errorf("failed to acquire advisory lock: %w", err)
			}
			
			if !lockAcquired {
				// If lock not acquired, fall back to UUID-based ID
				e.ReferenceID = fmt.Sprintf("EP-%s", uuid.New().String()[:8])
			} else {
				// Lock acquired, safely generate sequential reference ID
				var count int64
				if err := tx.Model(&Epic{}).Count(&count).Error; err != nil {
					return fmt.Errorf("failed to count epics: %w", err)
				}
				e.ReferenceID = fmt.Sprintf("EP-%03d", count+1)
			}
		} else {
			// For non-PostgreSQL databases (like SQLite in tests), use simple count method
			var count int64
			if err := tx.Model(&Epic{}).Count(&count).Error; err != nil {
				return fmt.Errorf("failed to count epics: %w", err)
			}
			e.ReferenceID = fmt.Sprintf("EP-%03d", count+1)
		}
	}
	
	return nil
}

// BeforeUpdate updates the LastModified timestamp
func (e *Epic) BeforeUpdate(tx *gorm.DB) error {
	e.LastModified = time.Now().UTC()
	return nil
}

// TableName returns the table name for the Epic model
func (Epic) TableName() string {
	return "epics"
}

// GetPriorityString returns the string representation of the priority
func (e *Epic) GetPriorityString() string {
	switch e.Priority {
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

// IsValidStatus checks if the provided status is valid for epics
func (e *Epic) IsValidStatus(status EpicStatus) bool {
	validStatuses := []EpicStatus{
		EpicStatusBacklog,
		EpicStatusDraft,
		EpicStatusInProgress,
		EpicStatusDone,
		EpicStatusCancelled,
	}
	
	for _, validStatus := range validStatuses {
		if status == validStatus {
			return true
		}
	}
	return false
}

// CanTransitionTo checks if the epic can transition to the given status
// By default, all transitions are allowed as per requirements
func (e *Epic) CanTransitionTo(newStatus EpicStatus) bool {
	return e.IsValidStatus(newStatus)
}

// HasUserStories checks if the epic has any associated user stories
func (e *Epic) HasUserStories() bool {
	return len(e.UserStories) > 0
}
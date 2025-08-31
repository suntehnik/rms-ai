package models

import (
	"gorm.io/gorm"
)

// AllModels returns a slice of all model structs for migration purposes
func AllModels() []interface{} {
	return []interface{}{
		&User{},
		&Epic{},
		&UserStory{},
		&AcceptanceCriteria{},
		&RequirementType{},
		&RelationshipType{},
		&Requirement{},
		&RequirementRelationship{},
		&Comment{},
	}
}

// AutoMigrate runs auto-migration for all models
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(AllModels()...)
}

// SeedDefaultData seeds the database with default requirement types and relationship types
func SeedDefaultData(db *gorm.DB) error {
	// Seed default requirement types
	for _, reqType := range GetDefaultRequirementTypes() {
		var existingType RequirementType
		result := db.Where("name = ?", reqType.Name).First(&existingType)
		if result.Error == gorm.ErrRecordNotFound {
			if err := db.Create(&reqType).Error; err != nil {
				return err
			}
		}
	}

	// Seed default relationship types
	for _, relType := range GetDefaultRelationshipTypes() {
		var existingType RelationshipType
		result := db.Where("name = ?", relType.Name).First(&existingType)
		if result.Error == gorm.ErrRecordNotFound {
			if err := db.Create(&relType).Error; err != nil {
				return err
			}
		}
	}

	return nil
}

// ValidatePriority checks if the priority value is valid (1-4)
func ValidatePriority(priority Priority) bool {
	return priority >= PriorityCritical && priority <= PriorityLow
}

// GetPriorityString returns the string representation of a priority
func GetPriorityString(priority Priority) string {
	switch priority {
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

// GetAllValidEpicStatuses returns all valid epic statuses
func GetAllValidEpicStatuses() []EpicStatus {
	return []EpicStatus{
		EpicStatusBacklog,
		EpicStatusDraft,
		EpicStatusInProgress,
		EpicStatusDone,
		EpicStatusCancelled,
	}
}

// GetAllValidUserStoryStatuses returns all valid user story statuses
func GetAllValidUserStoryStatuses() []UserStoryStatus {
	return []UserStoryStatus{
		UserStoryStatusBacklog,
		UserStoryStatusDraft,
		UserStoryStatusInProgress,
		UserStoryStatusDone,
		UserStoryStatusCancelled,
	}
}

// GetAllValidRequirementStatuses returns all valid requirement statuses
func GetAllValidRequirementStatuses() []RequirementStatus {
	return []RequirementStatus{
		RequirementStatusDraft,
		RequirementStatusActive,
		RequirementStatusObsolete,
	}
}

// GetAllValidEntityTypes returns all valid entity types for comments
func GetAllValidEntityTypes() []EntityType {
	return []EntityType{
		EntityTypeEpic,
		EntityTypeUserStory,
		EntityTypeAcceptanceCriteria,
		EntityTypeRequirement,
	}
}
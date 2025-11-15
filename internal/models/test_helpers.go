package models

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB creates an in-memory SQLite database for testing
// Note: This is only for tests that don't involve reference ID generation
// For reference ID tests, use integration tests with PostgreSQL
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Auto-migrate all models
	err = db.AutoMigrate(
		&User{},
		&Epic{},
		&UserStory{},
		&Requirement{},
		&AcceptanceCriteria{},
		&RequirementType{},
		&RelationshipType{},
		&RequirementRelationship{},
		&Comment{},
		&PersonalAccessToken{},
		&SteeringDocument{},
		&EpicSteeringDocument{},
		&Prompt{},
		&RefreshToken{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

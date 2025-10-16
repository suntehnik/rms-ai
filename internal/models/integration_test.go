//go:build integration
// +build integration

package models

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"product-requirements-management/internal/config"
)

// TestPostgreSQLIntegration tests the models with actual PostgreSQL database
// Run with: go test -tags=integration ./internal/models
func TestPostgreSQLIntegration(t *testing.T) {
	// Skip if not running integration tests
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup test database connection
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host:     "localhost",
			Port:     "5432",
			User:     "postgres",
			Password: "postgres",
			DBName:   "product_requirements_test",
			SSLMode:  "disable",
		},
	}

	dsn := "host=localhost user=postgres password=postgres dbname=product_requirements_test port=5432 sslmode=disable TimeZone=UTC"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skipf("Could not connect to test database: %v", err)
	}

	// Clean up tables before test
	cleanupTables(t, db)

	// Auto-migrate models
	err = AutoMigrate(db)
	require.NoError(t, err)

	// Seed default data
	err = SeedDefaultData(db)
	require.NoError(t, err)

	// Test creating entities with dual ID system
	t.Run("TestDualIDSystem", func(t *testing.T) {
		testDualIDSystem(t, db)
	})

	t.Run("TestRelationships", func(t *testing.T) {
		testRelationships(t, db)
	})

	t.Run("TestDefaultData", func(t *testing.T) {
		testDefaultData(t, db)
	})

	// Clean up after test
	cleanupTables(t, db)
}

func testDualIDSystem(t *testing.T, db *gorm.DB) {
	// Create a user
	user := User{
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
		Role:         RoleUser,
	}
	err := db.Create(&user).Error
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, user.ID)

	// Create an epic
	epic := Epic{
		CreatorID:   user.ID,
		AssigneeID:  user.ID,
		Priority:    PriorityHigh,
		Status:      EpicStatusBacklog,
		Title:       "Test Epic",
		Description: stringPtr("Test epic description"),
	}
	err = db.Create(&epic).Error
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, epic.ID)
	assert.NotEmpty(t, epic.ReferenceID)

	// Test finding by UUID
	var foundEpicByUUID Epic
	err = db.Where("id = ?", epic.ID).First(&foundEpicByUUID).Error
	require.NoError(t, err)
	assert.Equal(t, epic.Title, foundEpicByUUID.Title)

	// Test finding by reference ID
	var foundEpicByRef Epic
	err = db.Where("reference_id = ?", epic.ReferenceID).First(&foundEpicByRef).Error
	require.NoError(t, err)
	assert.Equal(t, epic.ID, foundEpicByRef.ID)
}

func testRelationships(t *testing.T, db *gorm.DB) {
	// Create a user
	user := User{
		Username:     "testuser2",
		Email:        "test2@example.com",
		PasswordHash: "hashedpassword",
		Role:         RoleUser,
	}
	err := db.Create(&user).Error
	require.NoError(t, err)

	// Create an epic
	epic := Epic{
		CreatorID:  user.ID,
		AssigneeID: user.ID,
		Priority:   PriorityHigh,
		Status:     EpicStatusBacklog,
		Title:      "Test Epic for Relationships",
	}
	err = db.Create(&epic).Error
	require.NoError(t, err)

	// Create a user story
	userStory := UserStory{
		EpicID:     epic.ID,
		CreatorID:  user.ID,
		AssigneeID: user.ID,
		Priority:   PriorityMedium,
		Status:     UserStoryStatusBacklog,
		Title:      "Test User Story",
	}
	err = db.Create(&userStory).Error
	require.NoError(t, err)

	// Create acceptance criteria
	ac := AcceptanceCriteria{
		UserStoryID: userStory.ID,
		AuthorID:    user.ID,
		Description: "WHEN user performs action THEN system SHALL respond",
	}
	err = db.Create(&ac).Error
	require.NoError(t, err)

	// Get requirement type
	var reqType RequirementType
	err = db.Where("name = ?", "Functional").First(&reqType).Error
	require.NoError(t, err)

	// Create a requirement
	requirement := Requirement{
		UserStoryID:          userStory.ID,
		AcceptanceCriteriaID: &ac.ID,
		CreatorID:            user.ID,
		AssigneeID:           user.ID,
		Priority:             PriorityHigh,
		Status:               RequirementStatusActive,
		TypeID:               reqType.ID,
		Title:                "Test Requirement",
	}
	err = db.Create(&requirement).Error
	require.NoError(t, err)

	// Test loading relationships
	var loadedEpic Epic
	err = db.Preload("UserStories.AcceptanceCriteria").Preload("UserStories.Requirements").Where("id = ?", epic.ID).First(&loadedEpic).Error
	require.NoError(t, err)
	assert.Len(t, loadedEpic.UserStories, 1)
	assert.Len(t, loadedEpic.UserStories[0].AcceptanceCriteria, 1)
	assert.Len(t, loadedEpic.UserStories[0].Requirements, 1)
}

func testDefaultData(t *testing.T, db *gorm.DB) {
	// Test that default requirement types were created
	var reqTypes []RequirementType
	err := db.Find(&reqTypes).Error
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(reqTypes), 5)

	// Check for specific default types
	var functionalType RequirementType
	err = db.Where("name = ?", "Functional").First(&functionalType).Error
	require.NoError(t, err)
	assert.Equal(t, "Functional", functionalType.Name)

	// Test that default relationship types were created
	var relTypes []RelationshipType
	err = db.Find(&relTypes).Error
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(relTypes), 5)

	// Check for specific default types
	var dependsOnType RelationshipType
	err = db.Where("name = ?", "depends_on").First(&dependsOnType).Error
	require.NoError(t, err)
	assert.Equal(t, "depends_on", dependsOnType.Name)
}

func cleanupTables(t *testing.T, db *gorm.DB) {
	// Drop tables in reverse dependency order
	tables := []string{
		"comments",
		"requirement_relationships",
		"requirements",
		"acceptance_criteria",
		"user_stories",
		"epics",
		"users",
		"relationship_types",
		"requirement_types",
	}

	for _, table := range tables {
		err := db.Exec("DROP TABLE IF EXISTS " + table + " CASCADE").Error
		if err != nil {
			t.Logf("Warning: Could not drop table %s: %v", table, err)
		}
	}
}

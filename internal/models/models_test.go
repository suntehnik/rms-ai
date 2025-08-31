package models

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Auto-migrate all models
	err = AutoMigrate(db)
	if err != nil {
		t.Fatalf("Failed to migrate models: %v", err)
	}

	return db
}

func TestUserModel(t *testing.T) {
	db := setupTestDB(t)

	user := User{
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
		Role:         RoleUser,
	}

	// Test creation
	err := db.Create(&user).Error
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, user.ID)

	// Test role methods
	assert.True(t, user.IsUser())
	assert.False(t, user.IsAdministrator())
	assert.False(t, user.IsCommenter())
	assert.True(t, user.CanEdit())
	assert.True(t, user.CanDelete())
	assert.False(t, user.CanManageUsers())
}

func TestEpicModel(t *testing.T) {
	db := setupTestDB(t)

	// Create a user first
	user := User{
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
		Role:         RoleUser,
	}
	err := db.Create(&user).Error
	assert.NoError(t, err)

	epic := Epic{
		CreatorID:   user.ID,
		AssigneeID:  user.ID,
		Priority:    PriorityHigh,
		Status:      EpicStatusBacklog,
		Title:       "Test Epic",
		Description: stringPtr("Test epic description"),
		ReferenceID: "EP-001", // Set manually for SQLite test
	}

	// Test creation
	err = db.Create(&epic).Error
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, epic.ID)
	assert.NotEmpty(t, epic.ReferenceID)

	// Test priority methods
	assert.Equal(t, "High", epic.GetPriorityString())

	// Test status validation
	assert.True(t, epic.IsValidStatus(EpicStatusDone))
	assert.False(t, epic.IsValidStatus("InvalidStatus"))

	// Test status transitions
	assert.True(t, epic.CanTransitionTo(EpicStatusInProgress))
}

func TestUserStoryModel(t *testing.T) {
	db := setupTestDB(t)

	// Create user and epic first
	user := User{
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
		Role:         RoleUser,
	}
	err := db.Create(&user).Error
	assert.NoError(t, err)

	epic := Epic{
		CreatorID:   user.ID,
		AssigneeID:  user.ID,
		Priority:    PriorityHigh,
		Status:      EpicStatusBacklog,
		Title:       "Test Epic",
	}
	err = db.Create(&epic).Error
	assert.NoError(t, err)

	userStory := UserStory{
		EpicID:      epic.ID,
		CreatorID:   user.ID,
		AssigneeID:  user.ID,
		Priority:    PriorityMedium,
		Status:      UserStoryStatusBacklog,
		Title:       "Test User Story",
		Description: stringPtr("As a user, I want to test, so that I can verify functionality"),
	}

	// Test creation
	err = db.Create(&userStory).Error
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, userStory.ID)
	assert.NotEmpty(t, userStory.ReferenceID)

	// Test user story template validation
	assert.True(t, userStory.IsUserStoryTemplate())

	// Test invalid template
	userStory.Description = stringPtr("Invalid description")
	assert.False(t, userStory.IsUserStoryTemplate())
}

func TestAcceptanceCriteriaModel(t *testing.T) {
	db := setupTestDB(t)

	// Create user, epic, and user story first
	user := User{
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
		Role:         RoleUser,
	}
	err := db.Create(&user).Error
	assert.NoError(t, err)

	epic := Epic{
		CreatorID:  user.ID,
		AssigneeID: user.ID,
		Priority:   PriorityHigh,
		Status:     EpicStatusBacklog,
		Title:      "Test Epic",
	}
	err = db.Create(&epic).Error
	assert.NoError(t, err)

	userStory := UserStory{
		EpicID:     epic.ID,
		CreatorID:  user.ID,
		AssigneeID: user.ID,
		Priority:   PriorityMedium,
		Status:     UserStoryStatusBacklog,
		Title:      "Test User Story",
	}
	err = db.Create(&userStory).Error
	assert.NoError(t, err)

	ac := AcceptanceCriteria{
		UserStoryID: userStory.ID,
		AuthorID:    user.ID,
		Description: "WHEN user clicks button THEN system SHALL display message",
	}

	// Test creation
	err = db.Create(&ac).Error
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, ac.ID)
	assert.NotEmpty(t, ac.ReferenceID)

	// Test EARS format validation
	assert.True(t, ac.IsEARSFormat())

	// Test invalid EARS format
	ac.Description = "Invalid format"
	assert.False(t, ac.IsEARSFormat())
}

func TestRequirementTypeModel(t *testing.T) {
	db := setupTestDB(t)

	reqType := RequirementType{
		Name:        "Test Type",
		Description: stringPtr("Test requirement type"),
	}

	// Test creation
	err := db.Create(&reqType).Error
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, reqType.ID)

	// Test default types
	defaultTypes := GetDefaultRequirementTypes()
	assert.Len(t, defaultTypes, 5)
	assert.Equal(t, "Functional", defaultTypes[0].Name)
}

func TestRelationshipTypeModel(t *testing.T) {
	db := setupTestDB(t)

	relType := RelationshipType{
		Name:        "test_relationship",
		Description: stringPtr("Test relationship type"),
	}

	// Test creation
	err := db.Create(&relType).Error
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, relType.ID)

	// Test default types
	defaultTypes := GetDefaultRelationshipTypes()
	assert.Len(t, defaultTypes, 5)
	assert.Equal(t, "depends_on", defaultTypes[0].Name)
}

func TestCommentModel(t *testing.T) {
	db := setupTestDB(t)

	// Create user and epic first
	user := User{
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
		Role:         RoleUser,
	}
	err := db.Create(&user).Error
	assert.NoError(t, err)

	epic := Epic{
		CreatorID:  user.ID,
		AssigneeID: user.ID,
		Priority:   PriorityHigh,
		Status:     EpicStatusBacklog,
		Title:      "Test Epic",
	}
	err = db.Create(&epic).Error
	assert.NoError(t, err)

	comment := Comment{
		EntityType: EntityTypeEpic,
		EntityID:   epic.ID,
		AuthorID:   user.ID,
		Content:    "Test comment",
	}

	// Test creation
	err = db.Create(&comment).Error
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, comment.ID)

	// Test comment type methods
	assert.True(t, comment.IsGeneralComment())
	assert.False(t, comment.IsInlineComment())
	assert.True(t, comment.IsTopLevel())
	assert.False(t, comment.IsReply())

	// Test inline comment
	inlineComment := Comment{
		EntityType:        EntityTypeEpic,
		EntityID:          epic.ID,
		AuthorID:          user.ID,
		Content:           "Inline comment",
		LinkedText:        stringPtr("selected text"),
		TextPositionStart: intPtr(0),
		TextPositionEnd:   intPtr(13),
	}

	err = db.Create(&inlineComment).Error
	assert.NoError(t, err)
	assert.True(t, inlineComment.IsInlineComment())
	assert.True(t, inlineComment.IsValidTextPosition())
}

func TestPriorityValidation(t *testing.T) {
	assert.True(t, ValidatePriority(PriorityCritical))
	assert.True(t, ValidatePriority(PriorityHigh))
	assert.True(t, ValidatePriority(PriorityMedium))
	assert.True(t, ValidatePriority(PriorityLow))
	assert.False(t, ValidatePriority(Priority(0)))
	assert.False(t, ValidatePriority(Priority(5)))
}

func TestGetPriorityString(t *testing.T) {
	assert.Equal(t, "Critical", GetPriorityString(PriorityCritical))
	assert.Equal(t, "High", GetPriorityString(PriorityHigh))
	assert.Equal(t, "Medium", GetPriorityString(PriorityMedium))
	assert.Equal(t, "Low", GetPriorityString(PriorityLow))
	assert.Equal(t, "Unknown", GetPriorityString(Priority(0)))
}

// Helper functions for tests
func intPtr(i int) *int {
	return &i
}
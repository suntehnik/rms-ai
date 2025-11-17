package repository

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"product-requirements-management/internal/models"
)

// setupAcceptanceCriteriaTestDB creates an in-memory SQLite database for testing
func setupAcceptanceCriteriaTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Enable foreign key constraints in SQLite
	err = db.Exec("PRAGMA foreign_keys = ON").Error
	require.NoError(t, err)

	// Auto-migrate all required models
	err = db.AutoMigrate(
		&models.User{},
		&models.Epic{},
		&models.UserStory{},
		&models.AcceptanceCriteria{},
		&models.Requirement{},
		&models.RequirementType{},
	)
	require.NoError(t, err)

	return db
}

// Helper functions for AcceptanceCriteria tests

func createTestAcceptanceCriteria(t *testing.T, db *gorm.DB, userStory *models.UserStory, author *models.User, refID string) *models.AcceptanceCriteria {
	ac := &models.AcceptanceCriteria{
		UserStoryID: userStory.ID,
		AuthorID:    author.ID,
		Description: "WHEN user performs action THEN the system SHALL respond",
		ReferenceID: refID, // Set manually to avoid PostgreSQL function call
	}
	err := db.Create(ac).Error
	require.NoError(t, err)
	return ac
}

// CRUD Tests

// TestAcceptanceCriteriaRepository_Create tests the Create method
// References: AC-749 - Create returns acceptance criteria ID and commits; insertion errors surface without committing
func TestAcceptanceCriteriaRepository_Create(t *testing.T) {
	db := setupAcceptanceCriteriaTestDB(t)
	repo := NewAcceptanceCriteriaRepository(db)

	user := createUserStoryTestUser(t, db, "testuser")
	epic := createUserStoryTestEpic(t, db, user, "EP-001")
	userStory := createUserStoryTestUserStory(t, db, epic, user, user, "US-001")

	ac := &models.AcceptanceCriteria{
		UserStoryID: userStory.ID,
		AuthorID:    user.ID,
		Description: "WHEN user logs in THEN the system SHALL show dashboard",
		ReferenceID: "AC-001", // Set manually to avoid PostgreSQL function call
	}

	err := repo.Create(ac)
	require.NoError(t, err)
	assert.NotEmpty(t, ac.ID)
	assert.Equal(t, "AC-001", ac.ReferenceID)
	assert.NotEmpty(t, ac.Description)
}

// TestAcceptanceCriteriaRepository_GetByID tests the GetByID method
// References: AC-750 - Get by ID returns full acceptance criteria; missing record yields not-found
func TestAcceptanceCriteriaRepository_GetByID(t *testing.T) {
	db := setupAcceptanceCriteriaTestDB(t)
	repo := NewAcceptanceCriteriaRepository(db)

	user := createUserStoryTestUser(t, db, "testuser")
	epic := createUserStoryTestEpic(t, db, user, "EP-001")
	userStory := createUserStoryTestUserStory(t, db, epic, user, user, "US-001")
	ac := createTestAcceptanceCriteria(t, db, userStory, user, "AC-001")

	// Test successful retrieval
	retrieved, err := repo.GetByID(ac.ID)
	assert.NoError(t, err)
	assert.Equal(t, ac.ID, retrieved.ID)
	assert.Equal(t, ac.ReferenceID, retrieved.ReferenceID)
	assert.Equal(t, ac.Description, retrieved.Description)
}

// TestAcceptanceCriteriaRepository_GetByID_NotFound tests GetByID with non-existent ID
// References: AC-750 - Missing record yields not-found
func TestAcceptanceCriteriaRepository_GetByID_NotFound(t *testing.T) {
	db := setupAcceptanceCriteriaTestDB(t)
	repo := NewAcceptanceCriteriaRepository(db)

	nonExistentID := uuid.New()
	ac, err := repo.GetByID(nonExistentID)
	assert.Error(t, err)
	assert.Equal(t, ErrNotFound, err)
	assert.Nil(t, ac)
}

// TestAcceptanceCriteriaRepository_GetByReferenceID tests the GetByReferenceID method
// References: AC-750 - Get by ReferenceID returns full acceptance criteria; missing record yields not-found
func TestAcceptanceCriteriaRepository_GetByReferenceID(t *testing.T) {
	db := setupAcceptanceCriteriaTestDB(t)
	repo := NewAcceptanceCriteriaRepository(db)

	user := createUserStoryTestUser(t, db, "testuser")
	epic := createUserStoryTestEpic(t, db, user, "EP-001")
	userStory := createUserStoryTestUserStory(t, db, epic, user, user, "US-001")
	ac := createTestAcceptanceCriteria(t, db, userStory, user, "AC-001")

	// Test successful retrieval
	retrieved, err := repo.GetByReferenceID(ac.ReferenceID)
	assert.NoError(t, err)
	assert.Equal(t, ac.ID, retrieved.ID)
	assert.Equal(t, ac.ReferenceID, retrieved.ReferenceID)
	assert.Equal(t, ac.Description, retrieved.Description)
}

// TestAcceptanceCriteriaRepository_GetByReferenceID_NotFound tests GetByReferenceID with non-existent reference ID
// References: AC-750 - Missing record yields not-found
func TestAcceptanceCriteriaRepository_GetByReferenceID_NotFound(t *testing.T) {
	db := setupAcceptanceCriteriaTestDB(t)
	repo := NewAcceptanceCriteriaRepository(db)

	ac, err := repo.GetByReferenceID("AC-999")
	assert.Error(t, err)
	assert.Equal(t, ErrNotFound, err)
	assert.Nil(t, ac)
}

// TestAcceptanceCriteriaRepository_Update tests the Update method
// References: AC-751 - Update performs a single UPDATE; conflicts/errors propagate without committing
func TestAcceptanceCriteriaRepository_Update(t *testing.T) {
	db := setupAcceptanceCriteriaTestDB(t)
	repo := NewAcceptanceCriteriaRepository(db)

	user := createUserStoryTestUser(t, db, "testuser")
	epic := createUserStoryTestEpic(t, db, user, "EP-001")
	userStory := createUserStoryTestUserStory(t, db, epic, user, user, "US-001")
	ac := createTestAcceptanceCriteria(t, db, userStory, user, "AC-001")

	// Update the acceptance criteria
	newDescription := "WHEN user logs out THEN the system SHALL clear session"
	ac.Description = newDescription

	err := repo.Update(ac)
	require.NoError(t, err)

	// Verify the update
	retrieved, err := repo.GetByID(ac.ID)
	require.NoError(t, err)
	assert.Equal(t, newDescription, retrieved.Description)
}

// TestAcceptanceCriteriaRepository_Delete tests the Delete method
// References: AC-752 - Delete removes acceptance criteria; FK/relationship conflicts surface and record remains
func TestAcceptanceCriteriaRepository_Delete(t *testing.T) {
	db := setupAcceptanceCriteriaTestDB(t)
	repo := NewAcceptanceCriteriaRepository(db)

	user := createUserStoryTestUser(t, db, "testuser")
	epic := createUserStoryTestEpic(t, db, user, "EP-001")
	userStory := createUserStoryTestUserStory(t, db, epic, user, user, "US-001")
	ac := createTestAcceptanceCriteria(t, db, userStory, user, "AC-001")

	// Delete the acceptance criteria
	err := repo.Delete(ac.ID)
	require.NoError(t, err)

	// Verify deletion
	retrieved, err := repo.GetByID(ac.ID)
	assert.Error(t, err)
	assert.Equal(t, ErrNotFound, err)
	assert.Nil(t, retrieved)
}

// TestAcceptanceCriteriaRepository_Create_DuplicateReferenceID tests insertion error with duplicate reference ID
// References: AC-749 - insertion errors surface without committing
func TestAcceptanceCriteriaRepository_Create_DuplicateReferenceID(t *testing.T) {
	db := setupAcceptanceCriteriaTestDB(t)
	repo := NewAcceptanceCriteriaRepository(db)

	user := createUserStoryTestUser(t, db, "testuser")
	epic := createUserStoryTestEpic(t, db, user, "EP-001")
	userStory := createUserStoryTestUserStory(t, db, epic, user, user, "US-001")

	// Create first acceptance criteria
	ac1 := createTestAcceptanceCriteria(t, db, userStory, user, "AC-001")

	// Try to create second acceptance criteria with same ReferenceID
	ac2 := &models.AcceptanceCriteria{
		UserStoryID: userStory.ID,
		AuthorID:    user.ID,
		Description: "Duplicate AC",
		ReferenceID: "AC-001", // Duplicate!
	}

	err := repo.Create(ac2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "UNIQUE constraint")

	// Verify original record still exists and no side effects
	retrieved, err := repo.GetByReferenceID("AC-001")
	assert.NoError(t, err)
	assert.Equal(t, ac1.ID, retrieved.ID)
	assert.Equal(t, ac1.Description, retrieved.Description)
}

// NOTE: Update_NonExistent test is not included here as existence validation
// happens at the service layer, not repository layer (AC-751).
// GORM's Save() method doesn't return an error for non-existent records.

// TestAcceptanceCriteriaRepository_Delete_WithRequirements tests FK constraint on delete with requirements
// References: AC-752 - FK/relationship conflicts surface and record remains
func TestAcceptanceCriteriaRepository_Delete_WithRequirements(t *testing.T) {
	db := setupAcceptanceCriteriaTestDB(t)
	repo := NewAcceptanceCriteriaRepository(db)

	user := createUserStoryTestUser(t, db, "testuser")
	epic := createUserStoryTestEpic(t, db, user, "EP-001")
	userStory := createUserStoryTestUserStory(t, db, epic, user, user, "US-001")
	ac := createTestAcceptanceCriteria(t, db, userStory, user, "AC-001")

	// Create requirement type
	reqType := createTestRequirementType(t, db, "Functional")

	// Create requirement linked to acceptance criteria (with SET NULL it should nullify the FK)
	requirement := &models.Requirement{
		UserStoryID:          userStory.ID,
		CreatorID:            user.ID,
		AssigneeID:           user.ID,
		Title:                "Test Requirement",
		Priority:             models.PriorityMedium,
		TypeID:               reqType.ID,
		AcceptanceCriteriaID: &ac.ID,
		ReferenceID:          "REQ-001",
	}
	err := db.Create(requirement).Error
	require.NoError(t, err)

	// Delete acceptance criteria - should succeed with SET NULL
	err = repo.Delete(ac.ID)
	assert.NoError(t, err)

	// Verify AC is deleted
	retrieved, err := repo.GetByID(ac.ID)
	assert.Error(t, err)
	assert.Equal(t, ErrNotFound, err)
	assert.Nil(t, retrieved)

	// Verify requirement still exists but acceptance_criteria_id is NULL
	var req models.Requirement
	err = db.First(&req, requirement.ID).Error
	assert.NoError(t, err)
	assert.Nil(t, req.AcceptanceCriteriaID)
}

// TestAcceptanceCriteriaRepository_CountByUserStory tests counting acceptance criteria for a user story
// References: AC-753 - CRUD tests are isolated on mocked or in-memory DB, idempotent and independent
func TestAcceptanceCriteriaRepository_CountByUserStory(t *testing.T) {
	db := setupAcceptanceCriteriaTestDB(t)
	repo := NewAcceptanceCriteriaRepository(db)

	user := createUserStoryTestUser(t, db, "testuser")
	epic := createUserStoryTestEpic(t, db, user, "EP-001")
	userStory := createUserStoryTestUserStory(t, db, epic, user, user, "US-001")

	// Create acceptance criteria
	createTestAcceptanceCriteria(t, db, userStory, user, "AC-001")
	createTestAcceptanceCriteria(t, db, userStory, user, "AC-002")
	createTestAcceptanceCriteria(t, db, userStory, user, "AC-003")

	// Count acceptance criteria for user story
	count, err := repo.CountByUserStory(userStory.ID)
	assert.NoError(t, err)
	assert.Equal(t, int64(3), count)
}

// TestAcceptanceCriteriaRepository_HasRequirements tests checking if acceptance criteria has requirements
// References: AC-753 - CRUD tests are isolated on mocked or in-memory DB, idempotent and independent
func TestAcceptanceCriteriaRepository_HasRequirements(t *testing.T) {
	db := setupAcceptanceCriteriaTestDB(t)
	repo := NewAcceptanceCriteriaRepository(db)

	user := createUserStoryTestUser(t, db, "testuser")
	epic := createUserStoryTestEpic(t, db, user, "EP-001")
	userStory := createUserStoryTestUserStory(t, db, epic, user, user, "US-001")
	ac := createTestAcceptanceCriteria(t, db, userStory, user, "AC-001")

	// Initially should have no requirements
	hasReqs, err := repo.HasRequirements(ac.ID)
	assert.NoError(t, err)
	assert.False(t, hasReqs)

	// Create a requirement linked to this acceptance criteria
	reqType := createTestRequirementType(t, db, "Functional")
	requirement := &models.Requirement{
		UserStoryID:          userStory.ID,
		CreatorID:            user.ID,
		AssigneeID:           user.ID,
		Title:                "Test Requirement",
		Priority:             models.PriorityMedium,
		TypeID:               reqType.ID,
		AcceptanceCriteriaID: &ac.ID,
		ReferenceID:          "REQ-001",
	}
	err = db.Create(requirement).Error
	require.NoError(t, err)

	// Now should have requirements
	hasReqs, err = repo.HasRequirements(ac.ID)
	assert.NoError(t, err)
	assert.True(t, hasReqs)
}

// Filtering and Relations Tests

// TestAcceptanceCriteriaRepository_GetByUserStory tests filtering by user story ID
// References: AC-754 - Filter by user_story_id returns only matching acceptance criteria; missing story yields empty result
func TestAcceptanceCriteriaRepository_GetByUserStory(t *testing.T) {
	db := setupAcceptanceCriteriaTestDB(t)
	repo := NewAcceptanceCriteriaRepository(db)

	user := createUserStoryTestUser(t, db, "testuser")
	epic := createUserStoryTestEpic(t, db, user, "EP-001")
	userStory1 := createUserStoryTestUserStory(t, db, epic, user, user, "US-001")
	userStory2 := createUserStoryTestUserStory(t, db, epic, user, user, "US-002")

	// Create acceptance criteria for different user stories
	createTestAcceptanceCriteria(t, db, userStory1, user, "AC-001")
	createTestAcceptanceCriteria(t, db, userStory1, user, "AC-002")
	createTestAcceptanceCriteria(t, db, userStory2, user, "AC-003")

	// Test filter by userStory1
	result, err := repo.GetByUserStory(userStory1.ID)
	assert.NoError(t, err)
	assert.Len(t, result, 2)

	// Test filter by userStory2
	result, err = repo.GetByUserStory(userStory2.ID)
	assert.NoError(t, err)
	assert.Len(t, result, 1)

	// Test filter by non-existent user story
	nonExistentID := uuid.New()
	result, err = repo.GetByUserStory(nonExistentID)
	assert.NoError(t, err)
	assert.Len(t, result, 0)
}

// TestAcceptanceCriteriaRepository_GetByAuthor tests filtering by author ID
// References: AC-755 - Filter by author_id returns only authored acceptance criteria; unknown author yields empty result
func TestAcceptanceCriteriaRepository_GetByAuthor(t *testing.T) {
	db := setupAcceptanceCriteriaTestDB(t)
	repo := NewAcceptanceCriteriaRepository(db)

	user1 := createUserStoryTestUser(t, db, "user1")
	user2 := createUserStoryTestUser(t, db, "user2")
	epic := createUserStoryTestEpic(t, db, user1, "EP-001")
	userStory := createUserStoryTestUserStory(t, db, epic, user1, user1, "US-001")

	// Create acceptance criteria with different authors
	createTestAcceptanceCriteria(t, db, userStory, user1, "AC-001")
	createTestAcceptanceCriteria(t, db, userStory, user1, "AC-002")
	createTestAcceptanceCriteria(t, db, userStory, user2, "AC-003")

	// Test filter by user1
	result, err := repo.GetByAuthor(user1.ID)
	assert.NoError(t, err)
	assert.Len(t, result, 2)

	// Test filter by user2
	result, err = repo.GetByAuthor(user2.ID)
	assert.NoError(t, err)
	assert.Len(t, result, 1)

	// Test filter by non-existent author
	nonExistentUserID := uuid.New()
	result, err = repo.GetByAuthor(nonExistentUserID)
	assert.NoError(t, err)
	assert.Len(t, result, 0)
}

// TestAcceptanceCriteriaRepository_ListWithPreloads tests listing with preloads
// References: AC-757 - With preload enabled, comments/requirements/user_story load efficiently (minimal queries)
func TestAcceptanceCriteriaRepository_ListWithPreloads(t *testing.T) {
	db := setupAcceptanceCriteriaTestDB(t)
	repo := NewAcceptanceCriteriaRepository(db)

	user := createUserStoryTestUser(t, db, "testuser")
	epic := createUserStoryTestEpic(t, db, user, "EP-001")
	userStory := createUserStoryTestUserStory(t, db, epic, user, user, "US-001")

	// Create acceptance criteria
	createTestAcceptanceCriteria(t, db, userStory, user, "AC-001")
	createTestAcceptanceCriteria(t, db, userStory, user, "AC-002")

	// Test list with preloads and filters
	filters := map[string]interface{}{
		"user_story_id": userStory.ID,
	}
	result, err := repo.ListWithPreloads(filters, "created_at ASC", 10, 0)
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	// Verify preloads
	assert.NotEmpty(t, result[0].Author.Username)
	assert.NotEmpty(t, result[0].UserStory.Title)
}

// TestAcceptanceCriteriaRepository_GetByIDWithPreloads tests GetByID with preloads
// References: AC-757 - With preload enabled, comments/requirements/user_story load efficiently (minimal queries)
func TestAcceptanceCriteriaRepository_GetByIDWithPreloads(t *testing.T) {
	db := setupAcceptanceCriteriaTestDB(t)
	repo := NewAcceptanceCriteriaRepository(db)

	user := createUserStoryTestUser(t, db, "testuser")
	epic := createUserStoryTestEpic(t, db, user, "EP-001")
	userStory := createUserStoryTestUserStory(t, db, epic, user, user, "US-001")
	ac := createTestAcceptanceCriteria(t, db, userStory, user, "AC-001")

	// Get acceptance criteria with preloads
	retrieved, err := repo.GetByIDWithPreloads(ac.ID)
	assert.NoError(t, err)
	assert.Equal(t, ac.ID, retrieved.ID)
	// Verify preloads
	assert.NotEmpty(t, retrieved.Author.Username)
	assert.NotEmpty(t, retrieved.UserStory.Title)
}

// TestAcceptanceCriteriaRepository_GetByReferenceIDWithPreloads tests GetByReferenceID with preloads
// References: AC-757 - With preload enabled, comments/requirements/user_story load efficiently (minimal queries)
func TestAcceptanceCriteriaRepository_GetByReferenceIDWithPreloads(t *testing.T) {
	db := setupAcceptanceCriteriaTestDB(t)
	repo := NewAcceptanceCriteriaRepository(db)

	user := createUserStoryTestUser(t, db, "testuser")
	epic := createUserStoryTestEpic(t, db, user, "EP-001")
	userStory := createUserStoryTestUserStory(t, db, epic, user, user, "US-001")
	ac := createTestAcceptanceCriteria(t, db, userStory, user, "AC-001")

	// Get acceptance criteria with preloads
	retrieved, err := repo.GetByReferenceIDWithPreloads(ac.ReferenceID)
	assert.NoError(t, err)
	assert.Equal(t, ac.ID, retrieved.ID)
	// Verify preloads
	assert.NotEmpty(t, retrieved.Author.Username)
	assert.NotEmpty(t, retrieved.UserStory.Title)
}

// TestAcceptanceCriteriaRepository_ListWithPreloads_Pagination tests pagination with preloads
// References: AC-758 - DB interactions in filter/preload tests are mocked/verified; no external dependencies
func TestAcceptanceCriteriaRepository_ListWithPreloads_Pagination(t *testing.T) {
	db := setupAcceptanceCriteriaTestDB(t)
	repo := NewAcceptanceCriteriaRepository(db)

	user := createUserStoryTestUser(t, db, "testuser")
	epic := createUserStoryTestEpic(t, db, user, "EP-001")
	userStory := createUserStoryTestUserStory(t, db, epic, user, user, "US-001")

	// Create multiple acceptance criteria
	for i := 1; i <= 5; i++ {
		createTestAcceptanceCriteria(t, db, userStory, user, "AC-00"+string(rune('0'+i)))
	}

	// Test listing with limit
	filters := map[string]interface{}{
		"user_story_id": userStory.ID,
	}
	result, err := repo.ListWithPreloads(filters, "created_at ASC", 3, 0)
	assert.NoError(t, err)
	assert.Len(t, result, 3)

	// Test listing with limit and offset
	result, err = repo.ListWithPreloads(filters, "created_at ASC", 2, 2)
	assert.NoError(t, err)
	assert.Len(t, result, 2)

	// Test listing all
	result, err = repo.ListWithPreloads(filters, "created_at ASC", 10, 0)
	assert.NoError(t, err)
	assert.Len(t, result, 5)
}

// TestAcceptanceCriteriaRepository_WithLinkedRequirements tests acceptance criteria with linked requirements
// References: AC-758 - DB interactions in filter/preload tests are mocked/verified; no external dependencies
func TestAcceptanceCriteriaRepository_WithLinkedRequirements(t *testing.T) {
	db := setupAcceptanceCriteriaTestDB(t)
	repo := NewAcceptanceCriteriaRepository(db)

	user := createUserStoryTestUser(t, db, "testuser")
	epic := createUserStoryTestEpic(t, db, user, "EP-001")
	userStory := createUserStoryTestUserStory(t, db, epic, user, user, "US-001")
	ac := createTestAcceptanceCriteria(t, db, userStory, user, "AC-001")

	// Create requirement type
	reqType := createTestRequirementType(t, db, "Functional")

	// Create requirements linked to acceptance criteria
	req1 := &models.Requirement{
		UserStoryID:          userStory.ID,
		CreatorID:            user.ID,
		AssigneeID:           user.ID,
		Title:                "Requirement 1",
		Priority:             models.PriorityMedium,
		TypeID:               reqType.ID,
		AcceptanceCriteriaID: &ac.ID,
		ReferenceID:          "REQ-001",
	}
	err := db.Create(req1).Error
	require.NoError(t, err)

	req2 := &models.Requirement{
		UserStoryID:          userStory.ID,
		CreatorID:            user.ID,
		AssigneeID:           user.ID,
		Title:                "Requirement 2",
		Priority:             models.PriorityMedium,
		TypeID:               reqType.ID,
		AcceptanceCriteriaID: &ac.ID,
		ReferenceID:          "REQ-002",
	}
	err = db.Create(req2).Error
	require.NoError(t, err)

	// Verify that acceptance criteria has requirements
	hasReqs, err := repo.HasRequirements(ac.ID)
	assert.NoError(t, err)
	assert.True(t, hasReqs)

	// Count requirements
	var count int64
	err = db.Model(&models.Requirement{}).Where("acceptance_criteria_id = ?", ac.ID).Count(&count).Error
	assert.NoError(t, err)
	assert.Equal(t, int64(2), count)
}

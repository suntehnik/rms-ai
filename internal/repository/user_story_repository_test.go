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

// setupUserStoryTestDB creates an in-memory SQLite database for testing
func setupUserStoryTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate all required models
	err = db.AutoMigrate(
		&models.User{},
		&models.Epic{},
		&models.UserStory{},
		&models.Requirement{},
		&models.AcceptanceCriteria{},
	)
	require.NoError(t, err)

	return db
}

// Helper functions for user story tests

func createUserStoryTestUser(t *testing.T, db *gorm.DB, username string) *models.User {
	user := &models.User{
		ID:       uuid.New(),
		Username: username,
		Email:    username + "@example.com",
		Role:     models.RoleUser,
	}
	err := db.Create(user).Error
	require.NoError(t, err)
	return user
}

func createUserStoryTestEpic(t *testing.T, db *gorm.DB, user *models.User, refID string) *models.Epic {
	epic := &models.Epic{
		ID:          uuid.New(),
		Title:       "Test Epic",
		Priority:    models.PriorityHigh,
		CreatorID:   user.ID,
		Status:      models.EpicStatusBacklog,
		ReferenceID: refID, // Set manually to avoid PostgreSQL function call
	}
	err := db.Create(epic).Error
	require.NoError(t, err)
	return epic
}

func createUserStoryTestUserStory(t *testing.T, db *gorm.DB, epic *models.Epic, creator *models.User, assignee *models.User, refID string) *models.UserStory {
	userStory := &models.UserStory{
		Title:       "Test User Story",
		Priority:    models.PriorityMedium,
		EpicID:      epic.ID,
		CreatorID:   creator.ID,
		AssigneeID:  assignee.ID,
		Status:      models.UserStoryStatusBacklog,
		ReferenceID: refID, // Set manually to avoid PostgreSQL function call
	}
	err := db.Create(userStory).Error
	require.NoError(t, err)
	return userStory
}

// CRUD Tests

// TestUserStoryRepository_Create tests the Create method
// References: AC-759 - Create user story returns ID and commits; insertion error surfaces without committing
func TestUserStoryRepository_Create(t *testing.T) {
	db := setupUserStoryTestDB(t)
	repo := NewUserStoryRepository(db, nil)

	user := createUserStoryTestUser(t, db, "testuser")
	epic := createUserStoryTestEpic(t, db, user, "EP-001")

	userStory := &models.UserStory{
		Title:       "Test User Story",
		Priority:    models.PriorityMedium,
		EpicID:      epic.ID,
		CreatorID:   user.ID,
		AssigneeID:  user.ID,
		ReferenceID: "US-001", // Set manually to avoid PostgreSQL function call
	}

	err := repo.Create(userStory)
	require.NoError(t, err)
	assert.NotEmpty(t, userStory.ID)
	assert.Equal(t, "US-001", userStory.ReferenceID)
	assert.Equal(t, models.UserStoryStatusBacklog, userStory.Status) // Default status
}

// TestUserStoryRepository_GetByID tests the GetByID method
// References: AC-760 - Get by ID returns full user story; missing record yields not-found error
func TestUserStoryRepository_GetByID(t *testing.T) {
	db := setupUserStoryTestDB(t)
	repo := NewUserStoryRepository(db, nil)

	user := createUserStoryTestUser(t, db, "testuser")
	epic := createUserStoryTestEpic(t, db, user, "EP-001")
	userStory := createUserStoryTestUserStory(t, db, epic, user, user, "US-001")

	// Test successful retrieval
	retrieved, err := repo.GetByID(userStory.ID)
	assert.NoError(t, err)
	assert.Equal(t, userStory.ID, retrieved.ID)
	assert.Equal(t, userStory.ReferenceID, retrieved.ReferenceID)
	assert.Equal(t, userStory.Title, retrieved.Title)
}

// TestUserStoryRepository_GetByID_NotFound tests GetByID with non-existent ID
// References: AC-760 - Missing record yields not-found error
func TestUserStoryRepository_GetByID_NotFound(t *testing.T) {
	db := setupUserStoryTestDB(t)
	repo := NewUserStoryRepository(db, nil)

	nonExistentID := uuid.New()
	us, err := repo.GetByID(nonExistentID)
	assert.Error(t, err)
	assert.Equal(t, ErrNotFound, err)
	assert.Nil(t, us)
}

// TestUserStoryRepository_GetByReferenceID tests the GetByReferenceID method
// References: AC-760 - Get by ReferenceID returns full user story; missing record yields not-found error
func TestUserStoryRepository_GetByReferenceID(t *testing.T) {
	db := setupUserStoryTestDB(t)
	repo := NewUserStoryRepository(db, nil)

	user := createUserStoryTestUser(t, db, "testuser")
	epic := createUserStoryTestEpic(t, db, user, "EP-001")
	userStory := createUserStoryTestUserStory(t, db, epic, user, user, "US-001")

	// Test successful retrieval
	retrieved, err := repo.GetByReferenceID(userStory.ReferenceID)
	assert.NoError(t, err)
	assert.Equal(t, userStory.ID, retrieved.ID)
	assert.Equal(t, userStory.ReferenceID, retrieved.ReferenceID)
	assert.Equal(t, userStory.Title, retrieved.Title)
}

// TestUserStoryRepository_GetByReferenceID_NotFound tests GetByReferenceID with non-existent reference ID
// References: AC-760 - Missing record yields not-found error
func TestUserStoryRepository_GetByReferenceID_NotFound(t *testing.T) {
	db := setupUserStoryTestDB(t)
	repo := NewUserStoryRepository(db, nil)

	us, err := repo.GetByReferenceID("US-999")
	assert.Error(t, err)
	assert.Equal(t, ErrNotFound, err)
	assert.Nil(t, us)
}

// TestUserStoryRepository_Update tests the Update method
// References: AC-761 - Update performs a single UPDATE; conflicts/errors propagate without committing
func TestUserStoryRepository_Update(t *testing.T) {
	db := setupUserStoryTestDB(t)
	repo := NewUserStoryRepository(db, nil)

	user := createUserStoryTestUser(t, db, "testuser")
	epic := createUserStoryTestEpic(t, db, user, "EP-001")
	userStory := createUserStoryTestUserStory(t, db, epic, user, user, "US-001")

	// Update the user story
	newTitle := "Updated User Story"
	userStory.Title = newTitle
	userStory.Status = models.UserStoryStatusInProgress

	err := repo.Update(userStory)
	require.NoError(t, err)

	// Verify the update
	retrieved, err := repo.GetByID(userStory.ID)
	require.NoError(t, err)
	assert.Equal(t, newTitle, retrieved.Title)
	assert.Equal(t, models.UserStoryStatusInProgress, retrieved.Status)
}

// TestUserStoryRepository_Delete tests the Delete method
// References: AC-762 - Delete removes user story; FK/relationship conflicts surface and record remains
func TestUserStoryRepository_Delete(t *testing.T) {
	db := setupUserStoryTestDB(t)
	repo := NewUserStoryRepository(db, nil)

	user := createUserStoryTestUser(t, db, "testuser")
	epic := createUserStoryTestEpic(t, db, user, "EP-001")
	userStory := createUserStoryTestUserStory(t, db, epic, user, user, "US-001")

	// Delete the user story
	err := repo.Delete(userStory.ID)
	require.NoError(t, err)

	// Verify deletion
	retrieved, err := repo.GetByID(userStory.ID)
	assert.Error(t, err)
	assert.Equal(t, ErrNotFound, err)
	assert.Nil(t, retrieved)
}

// TestUserStoryRepository_GetWithAcceptanceCriteria tests GetWithAcceptanceCriteria method
// References: AC-763 - CRUD tests are isolated on mocked or in-memory DB, idempotent and independent
func TestUserStoryRepository_GetWithAcceptanceCriteria(t *testing.T) {
	db := setupUserStoryTestDB(t)
	repo := NewUserStoryRepository(db, nil)

	user := createUserStoryTestUser(t, db, "testuser")
	epic := createUserStoryTestEpic(t, db, user, "EP-001")
	userStory := createUserStoryTestUserStory(t, db, epic, user, user, "US-001")

	// Create acceptance criteria
	ac := &models.AcceptanceCriteria{
		UserStoryID: userStory.ID,
		AuthorID:    user.ID,
		Description: "WHEN user logs in THEN the system SHALL show dashboard",
		ReferenceID: "AC-001", // Set manually
	}
	err := db.Create(ac).Error
	require.NoError(t, err)

	// Get user story with acceptance criteria
	retrieved, err := repo.GetWithAcceptanceCriteria(userStory.ID)
	assert.NoError(t, err)
	assert.Equal(t, userStory.ID, retrieved.ID)
	assert.Len(t, retrieved.AcceptanceCriteria, 1)
	assert.Equal(t, ac.Description, retrieved.AcceptanceCriteria[0].Description)
}

// TestUserStoryRepository_GetWithRequirements tests GetWithRequirements method
// References: AC-763 - CRUD tests are isolated on mocked or in-memory DB, idempotent and independent
func TestUserStoryRepository_GetWithRequirements(t *testing.T) {
	db := setupUserStoryTestDB(t)
	repo := NewUserStoryRepository(db, nil)

	user := createUserStoryTestUser(t, db, "testuser")
	epic := createUserStoryTestEpic(t, db, user, "EP-001")
	userStory := createUserStoryTestUserStory(t, db, epic, user, user, "US-001")

	// Create requirement type first
	desc := "Functional requirement"
	reqType := &models.RequirementType{
		ID:          uuid.New(),
		Name:        "Functional",
		Description: &desc,
	}
	err := db.Create(reqType).Error
	require.NoError(t, err)

	// Create requirement
	req := &models.Requirement{
		UserStoryID: userStory.ID,
		CreatorID:   user.ID,
		AssigneeID:  user.ID,
		Title:       "Test Requirement",
		Priority:    models.PriorityMedium,
		TypeID:      reqType.ID,
		ReferenceID: "REQ-001", // Set manually
	}
	err = db.Create(req).Error
	require.NoError(t, err)

	// Get user story with requirements
	retrieved, err := repo.GetWithRequirements(userStory.ID)
	assert.NoError(t, err)
	assert.Equal(t, userStory.ID, retrieved.ID)
	assert.Len(t, retrieved.Requirements, 1)
	assert.Equal(t, req.Title, retrieved.Requirements[0].Title)
}

// Filtering and Pagination Tests

// TestUserStoryRepository_ListWithIncludes_Pagination tests pagination with limit and offset
// References: AC-764 - List with limit/offset returns correct slice and ordering; invalid limit/offset handled gracefully
func TestUserStoryRepository_ListWithIncludes_Pagination(t *testing.T) {
	db := setupUserStoryTestDB(t)
	repo := NewUserStoryRepository(db, nil)

	user := createUserStoryTestUser(t, db, "testuser")
	epic := createUserStoryTestEpic(t, db, user, "EP-001")

	// Create multiple user stories
	for i := 1; i <= 5; i++ {
		createUserStoryTestUserStory(t, db, epic, user, user, "US-00"+string(rune('0'+i)))
	}

	// Test listing with limit
	result, err := repo.ListWithIncludes(map[string]interface{}{}, []string{}, "created_at ASC", 3, 0)
	assert.NoError(t, err)
	assert.Len(t, result, 3)

	// Test listing with limit and offset
	result, err = repo.ListWithIncludes(map[string]interface{}{}, []string{}, "created_at ASC", 2, 2)
	assert.NoError(t, err)
	assert.Len(t, result, 2)

	// Test listing all (limit 0 means no limit)
	result, err = repo.ListWithIncludes(map[string]interface{}{}, []string{}, "created_at ASC", 0, 0)
	assert.NoError(t, err)
	assert.Len(t, result, 5)
}

// TestUserStoryRepository_GetByStatus tests filtering by status
// References: AC-765 - Filter by status returns only matching stories; unknown status yields empty result
func TestUserStoryRepository_GetByStatus(t *testing.T) {
	db := setupUserStoryTestDB(t)
	repo := NewUserStoryRepository(db, nil)

	user := createUserStoryTestUser(t, db, "testuser")
	epic := createUserStoryTestEpic(t, db, user, "EP-001")

	// Create user stories with different statuses
	us1 := createUserStoryTestUserStory(t, db, epic, user, user, "US-001")
	us1.Status = models.UserStoryStatusBacklog
	db.Save(us1)

	us2 := createUserStoryTestUserStory(t, db, epic, user, user, "US-002")
	us2.Status = models.UserStoryStatusInProgress
	db.Save(us2)

	us3 := createUserStoryTestUserStory(t, db, epic, user, user, "US-003")
	us3.Status = models.UserStoryStatusBacklog
	db.Save(us3)

	// Test filter by Backlog status
	result, err := repo.GetByStatus(models.UserStoryStatusBacklog)
	assert.NoError(t, err)
	assert.Len(t, result, 2)

	// Test filter by InProgress status
	result, err = repo.GetByStatus(models.UserStoryStatusInProgress)
	assert.NoError(t, err)
	assert.Len(t, result, 1)

	// Test filter by Done status (should be empty)
	result, err = repo.GetByStatus(models.UserStoryStatusDone)
	assert.NoError(t, err)
	assert.Len(t, result, 0)
}

// TestUserStoryRepository_GetByEpic tests filtering by epic ID
// References: AC-766 - Filter by epic_id returns only stories of the specified epic; missing epic yields empty result
func TestUserStoryRepository_GetByEpic(t *testing.T) {
	db := setupUserStoryTestDB(t)
	repo := NewUserStoryRepository(db, nil)

	user := createUserStoryTestUser(t, db, "testuser")
	epic1 := createUserStoryTestEpic(t, db, user, "EP-001")
	epic2 := createUserStoryTestEpic(t, db, user, "EP-002")

	// Create user stories for different epics
	createUserStoryTestUserStory(t, db, epic1, user, user, "US-001")
	createUserStoryTestUserStory(t, db, epic1, user, user, "US-002")
	createUserStoryTestUserStory(t, db, epic2, user, user, "US-003")

	// Test filter by epic1
	result, err := repo.GetByEpic(epic1.ID)
	assert.NoError(t, err)
	assert.Len(t, result, 2)

	// Test filter by epic2
	result, err = repo.GetByEpic(epic2.ID)
	assert.NoError(t, err)
	assert.Len(t, result, 1)

	// Test filter by non-existent epic
	nonExistentEpicID := uuid.New()
	result, err = repo.GetByEpic(nonExistentEpicID)
	assert.NoError(t, err)
	assert.Len(t, result, 0)
}

// TestUserStoryRepository_GetByAssignee tests filtering by assignee ID
// References: AC-767 - Filter by assignee_id returns only stories assigned to that user; missing assignee yields empty result
func TestUserStoryRepository_GetByAssignee(t *testing.T) {
	db := setupUserStoryTestDB(t)
	repo := NewUserStoryRepository(db, nil)

	user1 := createUserStoryTestUser(t, db, "user1")
	user2 := createUserStoryTestUser(t, db, "user2")
	epic := createUserStoryTestEpic(t, db, user1, "EP-001")

	// Create user stories assigned to different users
	createUserStoryTestUserStory(t, db, epic, user1, user1, "US-001")
	createUserStoryTestUserStory(t, db, epic, user1, user1, "US-002")
	createUserStoryTestUserStory(t, db, epic, user1, user2, "US-003")

	// Test filter by user1
	result, err := repo.GetByAssignee(user1.ID)
	assert.NoError(t, err)
	assert.Len(t, result, 2)

	// Test filter by user2
	result, err = repo.GetByAssignee(user2.ID)
	assert.NoError(t, err)
	assert.Len(t, result, 1)

	// Test filter by non-existent assignee
	nonExistentUserID := uuid.New()
	result, err = repo.GetByAssignee(nonExistentUserID)
	assert.NoError(t, err)
	assert.Len(t, result, 0)
}

// TestUserStoryRepository_GetByCreator tests filtering by creator ID
// References: AC-768 - DB interactions in filter/pagination tests are mocked/verified; no external dependencies
func TestUserStoryRepository_GetByCreator(t *testing.T) {
	db := setupUserStoryTestDB(t)
	repo := NewUserStoryRepository(db, nil)

	user1 := createUserStoryTestUser(t, db, "user1")
	user2 := createUserStoryTestUser(t, db, "user2")
	epic := createUserStoryTestEpic(t, db, user1, "EP-001")

	// Create user stories created by different users
	createUserStoryTestUserStory(t, db, epic, user1, user1, "US-001")
	createUserStoryTestUserStory(t, db, epic, user1, user1, "US-002")
	createUserStoryTestUserStory(t, db, epic, user2, user2, "US-003")

	// Test filter by user1
	result, err := repo.GetByCreator(user1.ID)
	assert.NoError(t, err)
	assert.Len(t, result, 2)

	// Test filter by user2
	result, err = repo.GetByCreator(user2.ID)
	assert.NoError(t, err)
	assert.Len(t, result, 1)

	// Test filter by non-existent creator
	nonExistentUserID := uuid.New()
	result, err = repo.GetByCreator(nonExistentUserID)
	assert.NoError(t, err)
	assert.Len(t, result, 0)
}

// TestUserStoryRepository_ListWithIncludes_WithFiltersAndPreloads tests complex filtering with includes
// References: AC-768 - DB interactions in filter/pagination tests are mocked/verified; no external dependencies
func TestUserStoryRepository_ListWithIncludes_WithFiltersAndPreloads(t *testing.T) {
	db := setupUserStoryTestDB(t)
	repo := NewUserStoryRepository(db, nil)

	user := createUserStoryTestUser(t, db, "testuser")
	epic := createUserStoryTestEpic(t, db, user, "EP-001")

	// Create user stories
	createUserStoryTestUserStory(t, db, epic, user, user, "US-001")
	createUserStoryTestUserStory(t, db, epic, user, user, "US-002")

	// Test with epic preload
	filters := map[string]interface{}{
		"epic_id": epic.ID,
	}
	result, err := repo.ListWithIncludes(filters, []string{"epic", "creator", "assignee"}, "created_at ASC", 10, 0)
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	// Verify preloads worked (epic should have a title)
	assert.NotEmpty(t, result[0].Epic.Title)
	assert.NotEmpty(t, result[0].Creator.Username)
	assert.NotEmpty(t, result[0].Assignee.Username)
}

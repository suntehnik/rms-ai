package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"product-requirements-management/internal/models"
)

func setupEpicTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate models
	err = db.AutoMigrate(&models.User{}, &models.Epic{}, &models.UserStory{})
	require.NoError(t, err)

	return db
}

func createTestEpic(t *testing.T, repo EpicRepository, userRepo UserRepository, title string, status models.EpicStatus, priority models.Priority) (*models.Epic, *models.User) {
	// Create a test user first with unique username/email
	username := "testuser_" + title
	email := "test_" + title + "@example.com"
	user := &models.User{
		Username:     username,
		Email:        email,
		PasswordHash: "hashed_password",
		Role:         models.RoleUser,
	}
	err := userRepo.Create(user)
	require.NoError(t, err)

	epic := &models.Epic{
		ReferenceID: "EP-" + title,
		CreatorID:   user.ID,
		AssigneeID:  user.ID,
		Priority:    priority,
		Status:      status,
		Title:       title,
	}

	err = repo.Create(epic)
	require.NoError(t, err)

	return epic, user
}

func TestEpicRepository_Create(t *testing.T) {
	db := setupEpicTestDB(t)
	epicRepo := NewEpicRepository(db)
	userRepo := NewUserRepository(db)

	epic, _ := createTestEpic(t, epicRepo, userRepo, "Test Epic", models.EpicStatusBacklog, models.PriorityHigh)

	assert.NotNil(t, epic.ID)
	assert.Equal(t, "Test Epic", epic.Title)
	assert.Equal(t, models.EpicStatusBacklog, epic.Status)
}

func TestEpicRepository_GetWithUserStories(t *testing.T) {
	db := setupEpicTestDB(t)
	epicRepo := NewEpicRepository(db)
	userRepo := NewUserRepository(db)
	userStoryRepo := NewUserStoryRepository(db)

	epic, user := createTestEpic(t, epicRepo, userRepo, "Test Epic", models.EpicStatusBacklog, models.PriorityHigh)

	// Create a user story for the epic
	userStory := &models.UserStory{
		ReferenceID: "US-001",
		EpicID:      epic.ID,
		CreatorID:   user.ID,
		AssigneeID:  user.ID,
		Priority:    models.PriorityMedium,
		Status:      models.UserStoryStatusBacklog,
		Title:       "Test User Story",
	}
	err := userStoryRepo.Create(userStory)
	require.NoError(t, err)

	// Get epic with user stories
	retrieved, err := epicRepo.GetWithUserStories(epic.ID)
	assert.NoError(t, err)
	assert.Equal(t, epic.ID, retrieved.ID)
	assert.Len(t, retrieved.UserStories, 1)
	assert.Equal(t, "Test User Story", retrieved.UserStories[0].Title)
}

func TestEpicRepository_GetByCreator(t *testing.T) {
	db := setupEpicTestDB(t)
	epicRepo := NewEpicRepository(db)
	userRepo := NewUserRepository(db)

	epic, user := createTestEpic(t, epicRepo, userRepo, "Test Epic", models.EpicStatusBacklog, models.PriorityHigh)

	// Get epics by creator
	epics, err := epicRepo.GetByCreator(user.ID)
	assert.NoError(t, err)
	assert.Len(t, epics, 1)
	assert.Equal(t, epic.ID, epics[0].ID)
}

func TestEpicRepository_GetByAssignee(t *testing.T) {
	db := setupEpicTestDB(t)
	epicRepo := NewEpicRepository(db)
	userRepo := NewUserRepository(db)

	epic, user := createTestEpic(t, epicRepo, userRepo, "Test Epic", models.EpicStatusBacklog, models.PriorityHigh)

	// Get epics by assignee
	epics, err := epicRepo.GetByAssignee(user.ID)
	assert.NoError(t, err)
	assert.Len(t, epics, 1)
	assert.Equal(t, epic.ID, epics[0].ID)
}

func TestEpicRepository_GetByStatus(t *testing.T) {
	db := setupEpicTestDB(t)
	epicRepo := NewEpicRepository(db)
	userRepo := NewUserRepository(db)

	// Create epics with different statuses
	createTestEpic(t, epicRepo, userRepo, "Epic 1", models.EpicStatusBacklog, models.PriorityHigh)
	createTestEpic(t, epicRepo, userRepo, "Epic 2", models.EpicStatusInProgress, models.PriorityMedium)

	// Get epics by status
	epics, err := epicRepo.GetByStatus(models.EpicStatusBacklog)
	assert.NoError(t, err)
	assert.Len(t, epics, 1)
	assert.Equal(t, models.EpicStatusBacklog, epics[0].Status)
}

func TestEpicRepository_GetByPriority(t *testing.T) {
	db := setupEpicTestDB(t)
	epicRepo := NewEpicRepository(db)
	userRepo := NewUserRepository(db)

	// Create epics with different priorities
	createTestEpic(t, epicRepo, userRepo, "Epic 1", models.EpicStatusBacklog, models.PriorityHigh)
	createTestEpic(t, epicRepo, userRepo, "Epic 2", models.EpicStatusBacklog, models.PriorityLow)

	// Get epics by priority
	epics, err := epicRepo.GetByPriority(models.PriorityHigh)
	assert.NoError(t, err)
	assert.Len(t, epics, 1)
	assert.Equal(t, models.PriorityHigh, epics[0].Priority)
}

func TestEpicRepository_HasUserStories(t *testing.T) {
	db := setupEpicTestDB(t)
	epicRepo := NewEpicRepository(db)
	userRepo := NewUserRepository(db)
	userStoryRepo := NewUserStoryRepository(db)

	epic, user := createTestEpic(t, epicRepo, userRepo, "Test Epic", models.EpicStatusBacklog, models.PriorityHigh)

	// Initially should have no user stories
	hasUserStories, err := epicRepo.HasUserStories(epic.ID)
	assert.NoError(t, err)
	assert.False(t, hasUserStories)

	// Create a user story for the epic
	userStory := &models.UserStory{
		ReferenceID: "US-001",
		EpicID:      epic.ID,
		CreatorID:   user.ID,
		AssigneeID:  user.ID,
		Priority:    models.PriorityMedium,
		Status:      models.UserStoryStatusBacklog,
		Title:       "Test User Story",
	}
	err = userStoryRepo.Create(userStory)
	require.NoError(t, err)

	// Now should have user stories
	hasUserStories, err = epicRepo.HasUserStories(epic.ID)
	assert.NoError(t, err)
	assert.True(t, hasUserStories)
}

func TestEpicRepository_GetByReferenceID(t *testing.T) {
	db := setupEpicTestDB(t)
	epicRepo := NewEpicRepository(db)
	userRepo := NewUserRepository(db)

	epic, _ := createTestEpic(t, epicRepo, userRepo, "Test Epic", models.EpicStatusBacklog, models.PriorityHigh)

	// Get by reference ID
	retrieved, err := epicRepo.GetByReferenceID(epic.ReferenceID)
	assert.NoError(t, err)
	assert.Equal(t, epic.ID, retrieved.ID)
	assert.Equal(t, epic.Title, retrieved.Title)
}

func TestEpicRepository_Update(t *testing.T) {
	db := setupEpicTestDB(t)
	epicRepo := NewEpicRepository(db)
	userRepo := NewUserRepository(db)

	epic, _ := createTestEpic(t, epicRepo, userRepo, "Test Epic", models.EpicStatusBacklog, models.PriorityHigh)

	// Update epic
	epic.Status = models.EpicStatusInProgress
	epic.Title = "Updated Epic"
	err := epicRepo.Update(epic)
	assert.NoError(t, err)

	// Verify update
	retrieved, err := epicRepo.GetByID(epic.ID)
	assert.NoError(t, err)
	assert.Equal(t, models.EpicStatusInProgress, retrieved.Status)
	assert.Equal(t, "Updated Epic", retrieved.Title)
}

func TestEpicRepository_Delete(t *testing.T) {
	db := setupEpicTestDB(t)
	epicRepo := NewEpicRepository(db)
	userRepo := NewUserRepository(db)

	epic, _ := createTestEpic(t, epicRepo, userRepo, "Test Epic", models.EpicStatusBacklog, models.PriorityHigh)

	// Delete epic
	err := epicRepo.Delete(epic.ID)
	assert.NoError(t, err)

	// Verify deletion
	retrieved, err := epicRepo.GetByID(epic.ID)
	assert.Error(t, err)
	assert.Equal(t, ErrNotFound, err)
	assert.Nil(t, retrieved)
}

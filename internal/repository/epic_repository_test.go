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

func setupEpicTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate models
	err = db.AutoMigrate(
		&models.User{},
		&models.Epic{},
		&models.SteeringDocument{},
		&models.EpicSteeringDocument{},
		&models.UserStory{},
		&models.AcceptanceCriteria{},
		&models.Requirement{},
		&models.RequirementType{},
	)
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
	userStoryRepo := NewUserStoryRepository(db, nil)

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
	userStoryRepo := NewUserStoryRepository(db, nil)

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

func TestEpicRepository_GetCompleteHierarchy(t *testing.T) {
	db := setupEpicTestDB(t)
	epicRepo := NewEpicRepository(db)
	userRepo := NewUserRepository(db)
	userStoryRepo := NewUserStoryRepository(db, nil)

	// Create test user
	user := &models.User{
		Username:     "testuser_hierarchy",
		Email:        "hierarchy@example.com",
		PasswordHash: "hashed_password",
		Role:         models.RoleUser,
	}
	err := userRepo.Create(user)
	require.NoError(t, err)

	// Create requirement type
	reqType := &models.RequirementType{
		Name:        "Functional",
		Description: strPtr("Functional requirement"),
	}
	err = db.Create(reqType).Error
	require.NoError(t, err)

	// Create epic
	epic := &models.Epic{
		ReferenceID: "EP-HIERARCHY",
		CreatorID:   user.ID,
		AssigneeID:  user.ID,
		Priority:    models.PriorityHigh,
		Status:      models.EpicStatusBacklog,
		Title:       "Test Epic for Hierarchy",
	}
	err = epicRepo.Create(epic)
	require.NoError(t, err)

	// Create user story 1
	userStory1 := &models.UserStory{
		ReferenceID: "US-HIERARCHY-1",
		EpicID:      epic.ID,
		CreatorID:   user.ID,
		AssigneeID:  user.ID,
		Priority:    models.PriorityMedium,
		Status:      models.UserStoryStatusBacklog,
		Title:       "User Story 1",
	}
	err = userStoryRepo.Create(userStory1)
	require.NoError(t, err)

	// Create user story 2
	userStory2 := &models.UserStory{
		ReferenceID: "US-HIERARCHY-2",
		EpicID:      epic.ID,
		CreatorID:   user.ID,
		AssigneeID:  user.ID,
		Priority:    models.PriorityLow,
		Status:      models.UserStoryStatusBacklog,
		Title:       "User Story 2",
	}
	err = userStoryRepo.Create(userStory2)
	require.NoError(t, err)

	// Create requirements for user story 1
	req1 := &models.Requirement{
		ReferenceID: "REQ-HIERARCHY-1",
		UserStoryID: userStory1.ID,
		CreatorID:   user.ID,
		AssigneeID:  user.ID,
		Priority:    models.PriorityHigh,
		Status:      models.RequirementStatusDraft,
		TypeID:      reqType.ID,
		Title:       "Requirement 1",
	}
	err = db.Create(req1).Error
	require.NoError(t, err)

	req2 := &models.Requirement{
		ReferenceID: "REQ-HIERARCHY-2",
		UserStoryID: userStory1.ID,
		CreatorID:   user.ID,
		AssigneeID:  user.ID,
		Priority:    models.PriorityMedium,
		Status:      models.RequirementStatusDraft,
		TypeID:      reqType.ID,
		Title:       "Requirement 2",
	}
	err = db.Create(req2).Error
	require.NoError(t, err)

	// Create acceptance criteria for user story 1
	ac1 := &models.AcceptanceCriteria{
		ReferenceID: "AC-HIERARCHY-1",
		UserStoryID: userStory1.ID,
		AuthorID:    user.ID,
		Description: "Acceptance Criteria 1",
	}
	err = db.Create(ac1).Error
	require.NoError(t, err)

	ac2 := &models.AcceptanceCriteria{
		ReferenceID: "AC-HIERARCHY-2",
		UserStoryID: userStory1.ID,
		AuthorID:    user.ID,
		Description: "Acceptance Criteria 2",
	}
	err = db.Create(ac2).Error
	require.NoError(t, err)

	// Create requirement for user story 2
	req3 := &models.Requirement{
		ReferenceID: "REQ-HIERARCHY-3",
		UserStoryID: userStory2.ID,
		CreatorID:   user.ID,
		AssigneeID:  user.ID,
		Priority:    models.PriorityLow,
		Status:      models.RequirementStatusDraft,
		TypeID:      reqType.ID,
		Title:       "Requirement 3",
	}
	err = db.Create(req3).Error
	require.NoError(t, err)

	// Test GetCompleteHierarchy
	retrieved, err := epicRepo.GetCompleteHierarchy(epic.ID)
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)

	// Verify epic
	assert.Equal(t, epic.ID, retrieved.ID)
	assert.Equal(t, "Test Epic for Hierarchy", retrieved.Title)

	// Verify user stories are loaded and ordered
	assert.Len(t, retrieved.UserStories, 2)
	assert.Equal(t, "User Story 1", retrieved.UserStories[0].Title)
	assert.Equal(t, "User Story 2", retrieved.UserStories[1].Title)

	// Verify requirements for user story 1
	assert.Len(t, retrieved.UserStories[0].Requirements, 2)
	assert.Equal(t, "Requirement 1", retrieved.UserStories[0].Requirements[0].Title)
	assert.Equal(t, "Requirement 2", retrieved.UserStories[0].Requirements[1].Title)

	// Verify requirement types are loaded
	assert.NotNil(t, retrieved.UserStories[0].Requirements[0].Type)
	assert.Equal(t, "Functional", retrieved.UserStories[0].Requirements[0].Type.Name)

	// Verify acceptance criteria for user story 1
	assert.Len(t, retrieved.UserStories[0].AcceptanceCriteria, 2)
	assert.Equal(t, "Acceptance Criteria 1", retrieved.UserStories[0].AcceptanceCriteria[0].Description)
	assert.Equal(t, "Acceptance Criteria 2", retrieved.UserStories[0].AcceptanceCriteria[1].Description)

	// Verify requirements for user story 2
	assert.Len(t, retrieved.UserStories[1].Requirements, 1)
	assert.Equal(t, "Requirement 3", retrieved.UserStories[1].Requirements[0].Title)

	// Verify user story 2 has no acceptance criteria
	assert.Len(t, retrieved.UserStories[1].AcceptanceCriteria, 0)
}

func TestEpicRepository_GetCompleteHierarchy_EmptyUserStories(t *testing.T) {
	db := setupEpicTestDB(t)
	epicRepo := NewEpicRepository(db)
	userRepo := NewUserRepository(db)

	// Create test user
	user := &models.User{
		Username:     "testuser_empty",
		Email:        "empty@example.com",
		PasswordHash: "hashed_password",
		Role:         models.RoleUser,
	}
	err := userRepo.Create(user)
	require.NoError(t, err)

	// Create epic without user stories
	epic := &models.Epic{
		ReferenceID: "EP-EMPTY",
		CreatorID:   user.ID,
		AssigneeID:  user.ID,
		Priority:    models.PriorityHigh,
		Status:      models.EpicStatusBacklog,
		Title:       "Empty Epic",
	}
	err = epicRepo.Create(epic)
	require.NoError(t, err)

	// Test GetCompleteHierarchy
	retrieved, err := epicRepo.GetCompleteHierarchy(epic.ID)
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, epic.ID, retrieved.ID)
	assert.Len(t, retrieved.UserStories, 0)
}

func TestEpicRepository_GetCompleteHierarchy_NotFound(t *testing.T) {
	db := setupEpicTestDB(t)
	epicRepo := NewEpicRepository(db)

	// Test with non-existent epic ID
	retrieved, err := epicRepo.GetCompleteHierarchy(uuid.New())
	assert.Error(t, err)
	assert.Equal(t, ErrNotFound, err)
	assert.Nil(t, retrieved)
}

// Helper function to create string pointer
func strPtr(s string) *string {
	return &s
}

func TestEpicRepository_GetCompleteHierarchy_WithSteeringDocuments(t *testing.T) {
	db := setupEpicTestDB(t)
	epicRepo := NewEpicRepository(db)
	userRepo := NewUserRepository(db)
	userStoryRepo := NewUserStoryRepository(db, nil)

	// Create test user
	user := &models.User{
		Username:     "testuser_steering",
		Email:        "steering@example.com",
		PasswordHash: "hashed_password",
		Role:         models.RoleUser,
	}
	err := userRepo.Create(user)
	require.NoError(t, err)

	// Create requirement type
	reqType := &models.RequirementType{
		Name:        "Functional",
		Description: strPtr("Functional requirement"),
	}
	err = db.Create(reqType).Error
	require.NoError(t, err)

	// Create epic
	epic := &models.Epic{
		ReferenceID: "EP-STEERING",
		CreatorID:   user.ID,
		AssigneeID:  user.ID,
		Priority:    models.PriorityHigh,
		Status:      models.EpicStatusBacklog,
		Title:       "Test Epic with Steering Documents",
	}
	err = epicRepo.Create(epic)
	require.NoError(t, err)

	// Create steering documents
	std1 := &models.SteeringDocument{
		ReferenceID: "STD-001",
		Title:       "Technical Architecture Guidelines",
		Description: strPtr("Guidelines for system architecture"),
		CreatorID:   user.ID,
	}
	err = db.Create(std1).Error
	require.NoError(t, err)

	std2 := &models.SteeringDocument{
		ReferenceID: "STD-002",
		Title:       "API Design Standards",
		Description: strPtr("Standards for API design"),
		CreatorID:   user.ID,
	}
	err = db.Create(std2).Error
	require.NoError(t, err)

	// Link steering documents to epic via many-to-many relationship
	err = db.Model(&epic).Association("SteeringDocuments").Append(std1, std2)
	require.NoError(t, err)

	// Create user story
	userStory := &models.UserStory{
		ReferenceID: "US-STEERING-1",
		EpicID:      epic.ID,
		CreatorID:   user.ID,
		AssigneeID:  user.ID,
		Priority:    models.PriorityMedium,
		Status:      models.UserStoryStatusBacklog,
		Title:       "User Story with Requirements",
	}
	err = userStoryRepo.Create(userStory)
	require.NoError(t, err)

	// Create requirement
	req := &models.Requirement{
		ReferenceID: "REQ-STEERING-1",
		UserStoryID: userStory.ID,
		CreatorID:   user.ID,
		AssigneeID:  user.ID,
		Priority:    models.PriorityHigh,
		Status:      models.RequirementStatusDraft,
		TypeID:      reqType.ID,
		Title:       "Requirement 1",
	}
	err = db.Create(req).Error
	require.NoError(t, err)

	// Create acceptance criteria
	ac := &models.AcceptanceCriteria{
		ReferenceID: "AC-STEERING-1",
		UserStoryID: userStory.ID,
		AuthorID:    user.ID,
		Description: "Acceptance Criteria 1",
	}
	err = db.Create(ac).Error
	require.NoError(t, err)

	// Test GetCompleteHierarchy
	retrieved, err := epicRepo.GetCompleteHierarchy(epic.ID)
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)

	// Verify epic
	assert.Equal(t, epic.ID, retrieved.ID)
	assert.Equal(t, "Test Epic with Steering Documents", retrieved.Title)

	// Verify steering documents are loaded and ordered
	assert.Len(t, retrieved.SteeringDocuments, 2)
	assert.Equal(t, "Technical Architecture Guidelines", retrieved.SteeringDocuments[0].Title)
	assert.Equal(t, "API Design Standards", retrieved.SteeringDocuments[1].Title)
	assert.NotNil(t, retrieved.SteeringDocuments[0].Description)
	assert.Equal(t, "Guidelines for system architecture", *retrieved.SteeringDocuments[0].Description)

	// Verify user stories are loaded
	assert.Len(t, retrieved.UserStories, 1)
	assert.Equal(t, "User Story with Requirements", retrieved.UserStories[0].Title)

	// Verify requirements are loaded
	assert.Len(t, retrieved.UserStories[0].Requirements, 1)
	assert.Equal(t, "Requirement 1", retrieved.UserStories[0].Requirements[0].Title)

	// Verify requirement types are loaded
	assert.NotNil(t, retrieved.UserStories[0].Requirements[0].Type)
	assert.Equal(t, "Functional", retrieved.UserStories[0].Requirements[0].Type.Name)

	// Verify acceptance criteria are loaded
	assert.Len(t, retrieved.UserStories[0].AcceptanceCriteria, 1)
	assert.Equal(t, "Acceptance Criteria 1", retrieved.UserStories[0].AcceptanceCriteria[0].Description)
}

func TestEpicRepository_GetCompleteHierarchy_OnlySteeringDocuments(t *testing.T) {
	db := setupEpicTestDB(t)
	epicRepo := NewEpicRepository(db)
	userRepo := NewUserRepository(db)

	// Create test user
	user := &models.User{
		Username:     "testuser_only_steering",
		Email:        "only_steering@example.com",
		PasswordHash: "hashed_password",
		Role:         models.RoleUser,
	}
	err := userRepo.Create(user)
	require.NoError(t, err)

	// Create epic
	epic := &models.Epic{
		ReferenceID: "EP-ONLY-STEERING",
		CreatorID:   user.ID,
		AssigneeID:  user.ID,
		Priority:    models.PriorityHigh,
		Status:      models.EpicStatusBacklog,
		Title:       "Epic with Only Steering Documents",
	}
	err = epicRepo.Create(epic)
	require.NoError(t, err)

	// Create steering document
	std := &models.SteeringDocument{
		ReferenceID: "STD-003",
		Title:       "Security Standards",
		Description: strPtr("Security guidelines and best practices"),
		CreatorID:   user.ID,
	}
	err = db.Create(std).Error
	require.NoError(t, err)

	// Link steering document to epic
	err = db.Model(&epic).Association("SteeringDocuments").Append(std)
	require.NoError(t, err)

	// Test GetCompleteHierarchy
	retrieved, err := epicRepo.GetCompleteHierarchy(epic.ID)
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)

	// Verify epic
	assert.Equal(t, epic.ID, retrieved.ID)
	assert.Equal(t, "Epic with Only Steering Documents", retrieved.Title)

	// Verify steering documents are loaded
	assert.Len(t, retrieved.SteeringDocuments, 1)
	assert.Equal(t, "Security Standards", retrieved.SteeringDocuments[0].Title)

	// Verify no user stories
	assert.Len(t, retrieved.UserStories, 0)
}

func TestEpicRepository_GetCompleteHierarchy_NoSteeringDocuments(t *testing.T) {
	db := setupEpicTestDB(t)
	epicRepo := NewEpicRepository(db)
	userRepo := NewUserRepository(db)
	userStoryRepo := NewUserStoryRepository(db, nil)

	// Create test user
	user := &models.User{
		Username:     "testuser_no_steering",
		Email:        "no_steering@example.com",
		PasswordHash: "hashed_password",
		Role:         models.RoleUser,
	}
	err := userRepo.Create(user)
	require.NoError(t, err)

	// Create epic
	epic := &models.Epic{
		ReferenceID: "EP-NO-STEERING",
		CreatorID:   user.ID,
		AssigneeID:  user.ID,
		Priority:    models.PriorityHigh,
		Status:      models.EpicStatusBacklog,
		Title:       "Epic without Steering Documents",
	}
	err = epicRepo.Create(epic)
	require.NoError(t, err)

	// Create user story
	userStory := &models.UserStory{
		ReferenceID: "US-NO-STEERING-1",
		EpicID:      epic.ID,
		CreatorID:   user.ID,
		AssigneeID:  user.ID,
		Priority:    models.PriorityMedium,
		Status:      models.UserStoryStatusBacklog,
		Title:       "User Story without Steering Docs",
	}
	err = userStoryRepo.Create(userStory)
	require.NoError(t, err)

	// Test GetCompleteHierarchy
	retrieved, err := epicRepo.GetCompleteHierarchy(epic.ID)
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)

	// Verify epic
	assert.Equal(t, epic.ID, retrieved.ID)
	assert.Equal(t, "Epic without Steering Documents", retrieved.Title)

	// Verify no steering documents
	assert.Len(t, retrieved.SteeringDocuments, 0)

	// Verify user stories are loaded
	assert.Len(t, retrieved.UserStories, 1)
	assert.Equal(t, "User Story without Steering Docs", retrieved.UserStories[0].Title)
}

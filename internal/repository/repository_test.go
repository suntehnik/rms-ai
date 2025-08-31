package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	
	"product-requirements-management/internal/models"
)

func setupRepositoryTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	
	// Auto-migrate all models
	err = models.AutoMigrate(db)
	require.NoError(t, err)
	
	return db
}

func TestNewRepositories(t *testing.T) {
	db := setupRepositoryTestDB(t)
	repos := NewRepositories(db)
	
	// Verify all repositories are created
	assert.NotNil(t, repos.User)
	assert.NotNil(t, repos.Epic)
	assert.NotNil(t, repos.UserStory)
	assert.NotNil(t, repos.AcceptanceCriteria)
	assert.NotNil(t, repos.Requirement)
	assert.NotNil(t, repos.RequirementType)
	assert.NotNil(t, repos.RelationshipType)
	assert.NotNil(t, repos.RequirementRelationship)
	assert.NotNil(t, repos.Comment)
}

func TestRepositories_WithTransaction(t *testing.T) {
	db := setupRepositoryTestDB(t)
	repos := NewRepositories(db)
	
	// Test successful transaction
	var createdUser *models.User
	err := repos.WithTransaction(func(txRepos *Repositories) error {
		user := &models.User{
			Username:     "testuser",
			Email:        "test@example.com",
			PasswordHash: "hashed_password",
			Role:         models.RoleUser,
		}
		
		if err := txRepos.User.Create(user); err != nil {
			return err
		}
		
		createdUser = user
		return nil
	})
	
	assert.NoError(t, err)
	assert.NotNil(t, createdUser)
	
	// Verify user was created
	retrieved, err := repos.User.GetByID(createdUser.ID)
	assert.NoError(t, err)
	assert.Equal(t, createdUser.Username, retrieved.Username)
}

func TestRepositories_WithTransaction_Rollback(t *testing.T) {
	db := setupRepositoryTestDB(t)
	repos := NewRepositories(db)
	
	// Test transaction rollback
	err := repos.WithTransaction(func(txRepos *Repositories) error {
		user := &models.User{
			Username:     "testuser",
			Email:        "test@example.com",
			PasswordHash: "hashed_password",
			Role:         models.RoleUser,
		}
		
		if err := txRepos.User.Create(user); err != nil {
			return err
		}
		
		// Force an error to trigger rollback
		return assert.AnError
	})
	
	assert.Error(t, err)
	
	// Verify user was not created due to rollback
	count, err := repos.User.Count(nil)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

func TestRepositories_IntegrationWorkflow(t *testing.T) {
	db := setupRepositoryTestDB(t)
	repos := NewRepositories(db)
	
	// Seed default data
	err := models.SeedDefaultData(db)
	require.NoError(t, err)
	
	// Create a complete workflow: User -> Epic -> UserStory -> AcceptanceCriteria -> Requirement
	err = repos.WithTransaction(func(txRepos *Repositories) error {
		// Create user
		user := &models.User{
			Username:     "testuser",
			Email:        "test@example.com",
			PasswordHash: "hashed_password",
			Role:         models.RoleUser,
		}
		if err := txRepos.User.Create(user); err != nil {
			return err
		}
		
		// Create epic
		epic := &models.Epic{
			ReferenceID: "EP-001",
			CreatorID:   user.ID,
			AssigneeID:  user.ID,
			Priority:    models.PriorityHigh,
			Status:      models.EpicStatusBacklog,
			Title:       "Test Epic",
		}
		if err := txRepos.Epic.Create(epic); err != nil {
			return err
		}
		
		// Create user story
		userStory := &models.UserStory{
			ReferenceID: "US-001",
			EpicID:      epic.ID,
			CreatorID:   user.ID,
			AssigneeID:  user.ID,
			Priority:    models.PriorityMedium,
			Status:      models.UserStoryStatusBacklog,
			Title:       "Test User Story",
		}
		if err := txRepos.UserStory.Create(userStory); err != nil {
			return err
		}
		
		// Create acceptance criteria
		acceptanceCriteria := &models.AcceptanceCriteria{
			ReferenceID: "AC-001",
			UserStoryID: userStory.ID,
			AuthorID:    user.ID,
			Description: "WHEN user performs action THEN system SHALL respond",
		}
		if err := txRepos.AcceptanceCriteria.Create(acceptanceCriteria); err != nil {
			return err
		}
		
		// Get a requirement type
		reqTypes, err := txRepos.RequirementType.List(nil, "", 1, 0)
		if err != nil {
			return err
		}
		if len(reqTypes) == 0 {
			return assert.AnError
		}
		
		// Create requirement
		requirement := &models.Requirement{
			ReferenceID:          "REQ-001",
			UserStoryID:          userStory.ID,
			AcceptanceCriteriaID: &acceptanceCriteria.ID,
			CreatorID:            user.ID,
			AssigneeID:           user.ID,
			Priority:             models.PriorityLow,
			Status:               models.RequirementStatusDraft,
			TypeID:               reqTypes[0].ID,
			Title:                "Test Requirement",
		}
		if err := txRepos.Requirement.Create(requirement); err != nil {
			return err
		}
		
		return nil
	})
	
	assert.NoError(t, err)
	
	// Verify the complete hierarchy was created
	epics, err := repos.Epic.List(nil, "", 0, 0)
	assert.NoError(t, err)
	assert.Len(t, epics, 1)
	
	userStories, err := repos.UserStory.GetByEpic(epics[0].ID)
	assert.NoError(t, err)
	assert.Len(t, userStories, 1)
	
	acceptanceCriteria, err := repos.AcceptanceCriteria.GetByUserStory(userStories[0].ID)
	assert.NoError(t, err)
	assert.Len(t, acceptanceCriteria, 1)
	
	requirements, err := repos.Requirement.GetByUserStory(userStories[0].ID)
	assert.NoError(t, err)
	assert.Len(t, requirements, 1)
}
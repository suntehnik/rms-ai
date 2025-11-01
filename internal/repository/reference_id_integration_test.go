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

// TestReferenceIDLookupIntegration tests that all entity repositories properly support
// direct reference ID lookups as required by task 3.1
func TestReferenceIDLookupIntegration(t *testing.T) {
	// Setup in-memory SQLite database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate all models
	err = db.AutoMigrate(
		&models.User{},
		&models.Epic{},
		&models.UserStory{},
		&models.AcceptanceCriteria{},
		&models.Requirement{},
		&models.RequirementType{},
		&models.SteeringDocument{},
	)
	require.NoError(t, err)

	// Create test data
	user := &models.User{
		ID:       uuid.New(),
		Username: "testuser",
		Email:    "test@example.com",
		Role:     models.RoleUser,
	}
	err = db.Create(user).Error
	require.NoError(t, err)

	reqType := &models.RequirementType{
		ID:   uuid.New(),
		Name: "Functional",
	}
	err = db.Create(reqType).Error
	require.NoError(t, err)

	// Test Epic Repository
	t.Run("Epic Repository", func(t *testing.T) {
		epicRepo := NewEpicRepository(db)

		epic := &models.Epic{
			ID:          uuid.New(),
			ReferenceID: "EP-001",
			Title:       "Test Epic",
			Priority:    models.PriorityHigh,
			CreatorID:   user.ID,
		}
		err := epicRepo.Create(epic)
		require.NoError(t, err)

		// Test case-sensitive lookup
		retrieved, err := epicRepo.GetByReferenceID("EP-001")
		assert.NoError(t, err)
		assert.Equal(t, epic.ID, retrieved.ID)
		assert.Equal(t, "EP-001", retrieved.ReferenceID)

		// Test case-insensitive lookup
		retrieved, err = epicRepo.GetByReferenceIDCaseInsensitive("ep-001")
		assert.NoError(t, err)
		assert.Equal(t, epic.ID, retrieved.ID)

		// Test not found
		_, err = epicRepo.GetByReferenceID("EP-999")
		assert.Error(t, err)
		assert.Equal(t, ErrNotFound, err)
	})

	// Test User Story Repository
	t.Run("UserStory Repository", func(t *testing.T) {
		userStoryRepo := NewUserStoryRepository(db, nil)
		epicRepo := NewEpicRepository(db)

		// Create epic first
		epic := &models.Epic{
			ID:          uuid.New(),
			ReferenceID: "EP-002",
			Title:       "Test Epic for US",
			Priority:    models.PriorityHigh,
			CreatorID:   user.ID,
		}
		err := epicRepo.Create(epic)
		require.NoError(t, err)

		userStory := &models.UserStory{
			ID:          uuid.New(),
			ReferenceID: "US-001",
			Title:       "Test User Story",
			Priority:    models.PriorityMedium,
			EpicID:      epic.ID,
			CreatorID:   user.ID,
		}
		err = userStoryRepo.Create(userStory)
		require.NoError(t, err)

		// Test case-sensitive lookup
		retrieved, err := userStoryRepo.GetByReferenceID("US-001")
		assert.NoError(t, err)
		assert.Equal(t, userStory.ID, retrieved.ID)
		assert.Equal(t, "US-001", retrieved.ReferenceID)

		// Test case-insensitive lookup
		retrieved, err = userStoryRepo.GetByReferenceIDCaseInsensitive("us-001")
		assert.NoError(t, err)
		assert.Equal(t, userStory.ID, retrieved.ID)

		// Test not found
		_, err = userStoryRepo.GetByReferenceID("US-999")
		assert.Error(t, err)
		assert.Equal(t, ErrNotFound, err)
	})

	// Test Acceptance Criteria Repository
	t.Run("AcceptanceCriteria Repository", func(t *testing.T) {
		acRepo := NewAcceptanceCriteriaRepository(db)
		userStoryRepo := NewUserStoryRepository(db, nil)
		epicRepo := NewEpicRepository(db)

		// Create epic and user story first
		epic := &models.Epic{
			ID:          uuid.New(),
			ReferenceID: "EP-003",
			Title:       "Test Epic for AC",
			Priority:    models.PriorityHigh,
			CreatorID:   user.ID,
		}
		err := epicRepo.Create(epic)
		require.NoError(t, err)

		userStory := &models.UserStory{
			ID:          uuid.New(),
			ReferenceID: "US-002",
			Title:       "Test User Story for AC",
			Priority:    models.PriorityMedium,
			EpicID:      epic.ID,
			CreatorID:   user.ID,
		}
		err = userStoryRepo.Create(userStory)
		require.NoError(t, err)

		ac := &models.AcceptanceCriteria{
			ID:          uuid.New(),
			ReferenceID: "AC-001",
			Description: "Test Acceptance Criteria",
			UserStoryID: userStory.ID,
			AuthorID:    user.ID,
		}
		err = acRepo.Create(ac)
		require.NoError(t, err)

		// Test case-sensitive lookup
		retrieved, err := acRepo.GetByReferenceID("AC-001")
		assert.NoError(t, err)
		assert.Equal(t, ac.ID, retrieved.ID)
		assert.Equal(t, "AC-001", retrieved.ReferenceID)

		// Test case-insensitive lookup
		retrieved, err = acRepo.GetByReferenceIDCaseInsensitive("ac-001")
		assert.NoError(t, err)
		assert.Equal(t, ac.ID, retrieved.ID)

		// Test not found
		_, err = acRepo.GetByReferenceID("AC-999")
		assert.Error(t, err)
		assert.Equal(t, ErrNotFound, err)
	})

	// Test Requirement Repository
	t.Run("Requirement Repository", func(t *testing.T) {
		reqRepo := NewRequirementRepository(db)
		userStoryRepo := NewUserStoryRepository(db, nil)
		epicRepo := NewEpicRepository(db)

		// Create epic and user story first
		epic := &models.Epic{
			ID:          uuid.New(),
			ReferenceID: "EP-004",
			Title:       "Test Epic for REQ",
			Priority:    models.PriorityHigh,
			CreatorID:   user.ID,
		}
		err := epicRepo.Create(epic)
		require.NoError(t, err)

		userStory := &models.UserStory{
			ID:          uuid.New(),
			ReferenceID: "US-003",
			Title:       "Test User Story for REQ",
			Priority:    models.PriorityMedium,
			EpicID:      epic.ID,
			CreatorID:   user.ID,
		}
		err = userStoryRepo.Create(userStory)
		require.NoError(t, err)

		req := &models.Requirement{
			ID:          uuid.New(),
			ReferenceID: "REQ-001",
			Title:       "Test Requirement",
			Priority:    models.PriorityLow,
			UserStoryID: userStory.ID,
			TypeID:      reqType.ID,
			CreatorID:   user.ID,
		}
		err = reqRepo.Create(req)
		require.NoError(t, err)

		// Test case-sensitive lookup
		retrieved, err := reqRepo.GetByReferenceID("REQ-001")
		assert.NoError(t, err)
		assert.Equal(t, req.ID, retrieved.ID)
		assert.Equal(t, "REQ-001", retrieved.ReferenceID)

		// Test case-insensitive lookup
		retrieved, err = reqRepo.GetByReferenceIDCaseInsensitive("req-001")
		assert.NoError(t, err)
		assert.Equal(t, req.ID, retrieved.ID)

		// Test not found
		_, err = reqRepo.GetByReferenceID("REQ-999")
		assert.Error(t, err)
		assert.Equal(t, ErrNotFound, err)
	})

	// Test Steering Document Repository
	t.Run("SteeringDocument Repository", func(t *testing.T) {
		steeringRepo := NewSteeringDocumentRepository(db)

		steering := &models.SteeringDocument{
			ID:          uuid.New(),
			ReferenceID: "STD-001",
			Title:       "Test Steering Document",
			CreatorID:   user.ID,
		}
		err = steeringRepo.Create(steering)
		require.NoError(t, err)

		// Test case-sensitive lookup
		retrieved, err := steeringRepo.GetByReferenceID("STD-001")
		assert.NoError(t, err)
		assert.Equal(t, steering.ID, retrieved.ID)
		assert.Equal(t, "STD-001", retrieved.ReferenceID)

		// Test case-insensitive lookup
		retrieved, err = steeringRepo.GetByReferenceIDCaseInsensitive("std-001")
		assert.NoError(t, err)
		assert.Equal(t, steering.ID, retrieved.ID)

		// Test not found
		_, err = steeringRepo.GetByReferenceID("STD-999")
		assert.Error(t, err)
		assert.Equal(t, ErrNotFound, err)
	})
}

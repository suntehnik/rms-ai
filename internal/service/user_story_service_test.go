package service

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"

	"product-requirements-management/internal/models"
	"product-requirements-management/internal/repository"
)

// MockUserStoryRepository is a mock implementation of UserStoryRepository
type MockUserStoryRepository struct {
	mock.Mock
}

func (m *MockUserStoryRepository) Create(entity *models.UserStory) error {
	args := m.Called(entity)
	return args.Error(0)
}

func (m *MockUserStoryRepository) GetByID(id uuid.UUID) (*models.UserStory, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserStory), args.Error(1)
}

func (m *MockUserStoryRepository) GetByReferenceID(referenceID string) (*models.UserStory, error) {
	args := m.Called(referenceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserStory), args.Error(1)
}

func (m *MockUserStoryRepository) GetByReferenceIDCaseInsensitive(referenceID string) (*models.UserStory, error) {
	args := m.Called(referenceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserStory), args.Error(1)
}

func (m *MockUserStoryRepository) Update(entity *models.UserStory) error {
	args := m.Called(entity)
	return args.Error(0)
}

func (m *MockUserStoryRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockUserStoryRepository) List(filters map[string]interface{}, orderBy string, limit, offset int) ([]models.UserStory, error) {
	args := m.Called(filters, orderBy, limit, offset)
	return args.Get(0).([]models.UserStory), args.Error(1)
}

func (m *MockUserStoryRepository) Count(filters map[string]interface{}) (int64, error) {
	args := m.Called(filters)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockUserStoryRepository) Exists(id uuid.UUID) (bool, error) {
	args := m.Called(id)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserStoryRepository) ExistsByReferenceID(referenceID string) (bool, error) {
	args := m.Called(referenceID)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserStoryRepository) WithTransaction(fn func(*gorm.DB) error) error {
	args := m.Called(fn)
	return args.Error(0)
}

func (m *MockUserStoryRepository) GetDB() *gorm.DB {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*gorm.DB)
}

func (m *MockUserStoryRepository) GetWithAcceptanceCriteria(id uuid.UUID) (*models.UserStory, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserStory), args.Error(1)
}

func (m *MockUserStoryRepository) GetWithRequirements(id uuid.UUID) (*models.UserStory, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserStory), args.Error(1)
}

func (m *MockUserStoryRepository) GetByEpic(epicID uuid.UUID) ([]models.UserStory, error) {
	args := m.Called(epicID)
	return args.Get(0).([]models.UserStory), args.Error(1)
}

func (m *MockUserStoryRepository) GetByCreator(creatorID uuid.UUID) ([]models.UserStory, error) {
	args := m.Called(creatorID)
	return args.Get(0).([]models.UserStory), args.Error(1)
}

func (m *MockUserStoryRepository) GetByAssignee(assigneeID uuid.UUID) ([]models.UserStory, error) {
	args := m.Called(assigneeID)
	return args.Get(0).([]models.UserStory), args.Error(1)
}

func (m *MockUserStoryRepository) GetByStatus(status models.UserStoryStatus) ([]models.UserStory, error) {
	args := m.Called(status)
	return args.Get(0).([]models.UserStory), args.Error(1)
}

func (m *MockUserStoryRepository) GetByPriority(priority models.Priority) ([]models.UserStory, error) {
	args := m.Called(priority)
	return args.Get(0).([]models.UserStory), args.Error(1)
}

func (m *MockUserStoryRepository) HasAcceptanceCriteria(id uuid.UUID) (bool, error) {
	args := m.Called(id)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserStoryRepository) HasRequirements(id uuid.UUID) (bool, error) {
	args := m.Called(id)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserStoryRepository) GetByIDWithUsers(id uuid.UUID) (*models.UserStory, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserStory), args.Error(1)
}

func (m *MockUserStoryRepository) GetByReferenceIDWithUsers(referenceID string) (*models.UserStory, error) {
	args := m.Called(referenceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserStory), args.Error(1)
}

func (m *MockUserStoryRepository) GetByReferenceIDWithUsersCaseInsensitive(referenceID string) (*models.UserStory, error) {
	args := m.Called(referenceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserStory), args.Error(1)
}

func (m *MockUserStoryRepository) ListWithIncludes(filters map[string]interface{}, includes []string, orderBy string, limit, offset int) ([]models.UserStory, error) {
	args := m.Called(filters, includes, orderBy, limit, offset)
	return args.Get(0).([]models.UserStory), args.Error(1)
}

func TestUserStoryService_CreateUserStory(t *testing.T) {
	mockUserStoryRepo := new(MockUserStoryRepository)
	mockEpicRepo := new(MockEpicRepository)
	mockUserRepo := new(MockUserRepository)

	service := NewUserStoryService(mockUserStoryRepo, mockEpicRepo, mockUserRepo)

	t.Run("successful creation", func(t *testing.T) {
		epicID := uuid.New()
		creatorID := uuid.New()
		description := "As a user, I want to login, so that I can access my account"

		req := CreateUserStoryRequest{
			EpicID:      epicID,
			CreatorID:   creatorID,
			Priority:    models.PriorityHigh,
			Title:       "User Login",
			Description: &description,
		}

		mockEpicRepo.On("Exists", epicID).Return(true, nil)
		mockUserRepo.On("Exists", creatorID).Return(true, nil)
		mockUserStoryRepo.On("Create", mock.AnythingOfType("*models.UserStory")).Return(nil)

		result, err := service.CreateUserStory(req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, epicID, result.EpicID)
		assert.Equal(t, creatorID, result.CreatorID)
		assert.Equal(t, creatorID, result.AssigneeID) // Should default to creator
		assert.Equal(t, models.PriorityHigh, result.Priority)
		assert.Equal(t, models.UserStoryStatusBacklog, result.Status)
		assert.Equal(t, "User Login", result.Title)
		assert.Equal(t, &description, result.Description)

		mockEpicRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
		mockUserStoryRepo.AssertExpectations(t)
	})

	t.Run("invalid priority", func(t *testing.T) {
		req := CreateUserStoryRequest{
			EpicID:    uuid.New(),
			CreatorID: uuid.New(),
			Priority:  models.Priority(5), // Invalid priority
			Title:     "Test User Story",
		}

		result, err := service.CreateUserStory(req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, ErrInvalidPriority, err)
	})

	t.Run("epic not found", func(t *testing.T) {
		epicID := uuid.New()
		creatorID := uuid.New()

		req := CreateUserStoryRequest{
			EpicID:    epicID,
			CreatorID: creatorID,
			Priority:  models.PriorityMedium,
			Title:     "Test User Story",
		}

		mockEpicRepo.On("Exists", epicID).Return(false, nil)

		result, err := service.CreateUserStory(req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, ErrEpicNotFound, err)

		mockEpicRepo.AssertExpectations(t)
	})

	t.Run("creator not found", func(t *testing.T) {
		epicID := uuid.New()
		creatorID := uuid.New()

		req := CreateUserStoryRequest{
			EpicID:    epicID,
			CreatorID: creatorID,
			Priority:  models.PriorityMedium,
			Title:     "Test User Story",
		}

		mockEpicRepo.On("Exists", epicID).Return(true, nil)
		mockUserRepo.On("Exists", creatorID).Return(false, nil)

		result, err := service.CreateUserStory(req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, ErrUserNotFound, err)

		mockEpicRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("invalid user story template", func(t *testing.T) {
		epicID := uuid.New()
		creatorID := uuid.New()
		invalidDescription := "This is not a proper user story template"

		req := CreateUserStoryRequest{
			EpicID:      epicID,
			CreatorID:   creatorID,
			Priority:    models.PriorityMedium,
			Title:       "Test User Story",
			Description: &invalidDescription,
		}

		mockEpicRepo.On("Exists", epicID).Return(true, nil)
		mockUserRepo.On("Exists", creatorID).Return(true, nil)

		result, err := service.CreateUserStory(req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, ErrInvalidUserStoryTemplate, err)

		mockEpicRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
	})
}

func TestUserStoryService_GetUserStoryByID(t *testing.T) {
	mockUserStoryRepo := new(MockUserStoryRepository)
	mockEpicRepo := new(MockEpicRepository)
	mockUserRepo := new(MockUserRepository)

	service := NewUserStoryService(mockUserStoryRepo, mockEpicRepo, mockUserRepo)

	t.Run("successful retrieval", func(t *testing.T) {
		userStoryID := uuid.New()
		expectedUserStory := &models.UserStory{
			ID:    userStoryID,
			Title: "Test User Story",
		}

		mockUserStoryRepo.On("GetByIDWithUsers", userStoryID).Return(expectedUserStory, nil)

		result, err := service.GetUserStoryByID(userStoryID)

		assert.NoError(t, err)
		assert.Equal(t, expectedUserStory, result)

		mockUserStoryRepo.AssertExpectations(t)
	})

	t.Run("user story not found", func(t *testing.T) {
		userStoryID := uuid.New()

		mockUserStoryRepo.On("GetByIDWithUsers", userStoryID).Return(nil, repository.ErrNotFound)

		result, err := service.GetUserStoryByID(userStoryID)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, ErrUserStoryNotFound, err)

		mockUserStoryRepo.AssertExpectations(t)
	})
}

func TestUserStoryService_UpdateUserStory(t *testing.T) {
	mockUserStoryRepo := new(MockUserStoryRepository)
	mockEpicRepo := new(MockEpicRepository)
	mockUserRepo := new(MockUserRepository)

	service := NewUserStoryService(mockUserStoryRepo, mockEpicRepo, mockUserRepo)

	t.Run("successful update", func(t *testing.T) {
		userStoryID := uuid.New()
		assigneeID := uuid.New()
		newTitle := "Updated User Story"
		newPriority := models.PriorityLow
		newStatus := models.UserStoryStatusInProgress
		newDescription := "As an admin, I want to manage users, so that I can control access"

		existingUserStory := &models.UserStory{
			ID:       userStoryID,
			Title:    "Original Title",
			Priority: models.PriorityHigh,
			Status:   models.UserStoryStatusBacklog,
		}

		req := UpdateUserStoryRequest{
			AssigneeID:  &assigneeID,
			Priority:    &newPriority,
			Status:      &newStatus,
			Title:       &newTitle,
			Description: &newDescription,
		}

		mockUserStoryRepo.On("GetByID", userStoryID).Return(existingUserStory, nil)
		mockUserRepo.On("Exists", assigneeID).Return(true, nil)
		mockUserStoryRepo.On("Update", mock.AnythingOfType("*models.UserStory")).Return(nil)

		result, err := service.UpdateUserStory(userStoryID, req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, assigneeID, result.AssigneeID)
		assert.Equal(t, newPriority, result.Priority)
		assert.Equal(t, newStatus, result.Status)
		assert.Equal(t, newTitle, result.Title)
		assert.Equal(t, &newDescription, result.Description)

		mockUserStoryRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("user story not found", func(t *testing.T) {
		userStoryID := uuid.New()
		req := UpdateUserStoryRequest{}

		mockUserStoryRepo.On("GetByID", userStoryID).Return(nil, repository.ErrNotFound)

		result, err := service.UpdateUserStory(userStoryID, req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, ErrUserStoryNotFound, err)

		mockUserStoryRepo.AssertExpectations(t)
	})

	t.Run("invalid user story template in update", func(t *testing.T) {
		userStoryID := uuid.New()
		invalidDescription := "This is not a proper user story template"

		existingUserStory := &models.UserStory{
			ID:    userStoryID,
			Title: "Original Title",
		}

		req := UpdateUserStoryRequest{
			Description: &invalidDescription,
		}

		mockUserStoryRepo.On("GetByID", userStoryID).Return(existingUserStory, nil)

		result, err := service.UpdateUserStory(userStoryID, req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, ErrInvalidUserStoryTemplate, err)

		mockUserStoryRepo.AssertExpectations(t)
	})
}

func TestUserStoryService_DeleteUserStory(t *testing.T) {
	mockUserStoryRepo := new(MockUserStoryRepository)
	mockEpicRepo := new(MockEpicRepository)
	mockUserRepo := new(MockUserRepository)

	service := NewUserStoryService(mockUserStoryRepo, mockEpicRepo, mockUserRepo)

	t.Run("successful deletion without requirements", func(t *testing.T) {
		userStoryID := uuid.New()
		existingUserStory := &models.UserStory{
			ID:    userStoryID,
			Title: "Test User Story",
		}

		mockUserStoryRepo.On("GetByID", userStoryID).Return(existingUserStory, nil)
		mockUserStoryRepo.On("HasRequirements", userStoryID).Return(false, nil)
		mockUserStoryRepo.On("Delete", userStoryID).Return(nil)

		err := service.DeleteUserStory(userStoryID, false)

		assert.NoError(t, err)

		mockUserStoryRepo.AssertExpectations(t)
	})

	t.Run("deletion blocked by requirements", func(t *testing.T) {
		userStoryID := uuid.New()
		existingUserStory := &models.UserStory{
			ID:    userStoryID,
			Title: "Test User Story",
		}

		mockUserStoryRepo.On("GetByID", userStoryID).Return(existingUserStory, nil)
		mockUserStoryRepo.On("HasRequirements", userStoryID).Return(true, nil)

		err := service.DeleteUserStory(userStoryID, false)

		assert.Error(t, err)
		assert.Equal(t, ErrUserStoryHasRequirements, err)

		mockUserStoryRepo.AssertExpectations(t)
	})

	t.Run("force deletion with requirements", func(t *testing.T) {
		userStoryID := uuid.New()
		existingUserStory := &models.UserStory{
			ID:    userStoryID,
			Title: "Test User Story",
		}

		mockUserStoryRepo.On("GetByID", userStoryID).Return(existingUserStory, nil)
		mockUserStoryRepo.On("Delete", userStoryID).Return(nil)

		err := service.DeleteUserStory(userStoryID, true)

		assert.NoError(t, err)

		mockUserStoryRepo.AssertExpectations(t)
	})

	t.Run("user story not found", func(t *testing.T) {
		userStoryID := uuid.New()

		mockUserStoryRepo.On("GetByID", userStoryID).Return(nil, repository.ErrNotFound)

		err := service.DeleteUserStory(userStoryID, false)

		assert.Error(t, err)
		assert.Equal(t, ErrUserStoryNotFound, err)

		mockUserStoryRepo.AssertExpectations(t)
	})
}

func TestUserStoryService_ListUserStories(t *testing.T) {
	mockUserStoryRepo := new(MockUserStoryRepository)
	mockEpicRepo := new(MockEpicRepository)
	mockUserRepo := new(MockUserRepository)

	service := NewUserStoryService(mockUserStoryRepo, mockEpicRepo, mockUserRepo)

	t.Run("successful listing with filters", func(t *testing.T) {
		epicID := uuid.New()
		status := models.UserStoryStatusInProgress
		priority := models.PriorityHigh

		filters := UserStoryFilters{
			EpicID:   &epicID,
			Status:   &status,
			Priority: &priority,
			OrderBy:  "priority ASC",
			Limit:    10,
			Offset:   0,
		}

		expectedUserStories := []models.UserStory{
			{ID: uuid.New(), Title: "User Story 1"},
			{ID: uuid.New(), Title: "User Story 2"},
		}

		expectedFilters := map[string]interface{}{
			"epic_id":  epicID,
			"status":   status,
			"priority": priority,
		}

		mockUserStoryRepo.On("Count", expectedFilters).Return(int64(2), nil)
		mockUserStoryRepo.On("ListWithIncludes", expectedFilters, []string{"Epic", "Creator", "Assignee"}, "priority ASC", 10, 0).Return(expectedUserStories, nil)

		result, count, err := service.ListUserStories(filters)

		assert.NoError(t, err)
		assert.Equal(t, expectedUserStories, result)
		assert.Equal(t, int64(2), count)

		mockUserStoryRepo.AssertExpectations(t)
	})

	t.Run("successful listing with default values", func(t *testing.T) {
		filters := UserStoryFilters{}

		expectedUserStories := []models.UserStory{
			{ID: uuid.New(), Title: "User Story 1"},
		}

		expectedFilters := map[string]interface{}{}

		mockUserStoryRepo.On("Count", expectedFilters).Return(int64(1), nil)
		mockUserStoryRepo.On("ListWithIncludes", expectedFilters, []string{"Epic", "Creator", "Assignee"}, "created_at DESC", 50, 0).Return(expectedUserStories, nil)

		result, count, err := service.ListUserStories(filters)

		assert.NoError(t, err)
		assert.Equal(t, expectedUserStories, result)
		assert.Equal(t, int64(1), count)

		mockUserStoryRepo.AssertExpectations(t)
	})
}

func TestUserStoryService_ValidateUserStoryTemplate(t *testing.T) {
	mockUserStoryRepo := new(MockUserStoryRepository)
	mockEpicRepo := new(MockEpicRepository)
	mockUserRepo := new(MockUserRepository)

	service := &userStoryService{
		userStoryRepo: mockUserStoryRepo,
		epicRepo:      mockEpicRepo,
		userRepo:      mockUserRepo,
	}

	t.Run("valid user story template", func(t *testing.T) {
		validDescription := "As a user, I want to login, so that I can access my account"
		err := service.validateUserStoryTemplate(&validDescription)
		assert.NoError(t, err)
	})

	t.Run("valid user story template with different case", func(t *testing.T) {
		validDescription := "AS A USER, I WANT TO LOGIN, SO THAT I CAN ACCESS MY ACCOUNT"
		err := service.validateUserStoryTemplate(&validDescription)
		assert.NoError(t, err)
	})

	t.Run("missing 'as' component", func(t *testing.T) {
		invalidDescription := "I want to login, so that I can access my account"
		err := service.validateUserStoryTemplate(&invalidDescription)
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidUserStoryTemplate, err)
	})

	t.Run("missing 'i want' component", func(t *testing.T) {
		invalidDescription := "As a user, so that I can access my account"
		err := service.validateUserStoryTemplate(&invalidDescription)
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidUserStoryTemplate, err)
	})

	t.Run("missing 'so that' component", func(t *testing.T) {
		invalidDescription := "As a user, I want to login"
		err := service.validateUserStoryTemplate(&invalidDescription)
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidUserStoryTemplate, err)
	})

	t.Run("nil description", func(t *testing.T) {
		err := service.validateUserStoryTemplate(nil)
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidUserStoryTemplate, err)
	})

	t.Run("empty description", func(t *testing.T) {
		emptyDescription := ""
		err := service.validateUserStoryTemplate(&emptyDescription)
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidUserStoryTemplate, err)
	})
}

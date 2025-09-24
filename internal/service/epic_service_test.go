package service

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"

	"product-requirements-management/internal/models"
	"product-requirements-management/internal/repository"
)

// MockEpicRepository is a mock implementation of EpicRepository
type MockEpicRepository struct {
	mock.Mock
}

func (m *MockEpicRepository) Create(epic *models.Epic) error {
	args := m.Called(epic)
	return args.Error(0)
}

func (m *MockEpicRepository) GetByID(id uuid.UUID) (*models.Epic, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Epic), args.Error(1)
}

func (m *MockEpicRepository) GetByReferenceID(referenceID string) (*models.Epic, error) {
	args := m.Called(referenceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Epic), args.Error(1)
}

func (m *MockEpicRepository) Update(epic *models.Epic) error {
	args := m.Called(epic)
	return args.Error(0)
}

func (m *MockEpicRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockEpicRepository) List(filters map[string]interface{}, orderBy string, limit, offset int) ([]models.Epic, error) {
	args := m.Called(filters, orderBy, limit, offset)
	return args.Get(0).([]models.Epic), args.Error(1)
}

func (m *MockEpicRepository) Count(filters map[string]interface{}) (int64, error) {
	args := m.Called(filters)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockEpicRepository) Exists(id uuid.UUID) (bool, error) {
	args := m.Called(id)
	return args.Bool(0), args.Error(1)
}

func (m *MockEpicRepository) ExistsByReferenceID(referenceID string) (bool, error) {
	args := m.Called(referenceID)
	return args.Bool(0), args.Error(1)
}

func (m *MockEpicRepository) WithTransaction(fn func(*gorm.DB) error) error {
	args := m.Called(fn)
	return args.Error(0)
}

func (m *MockEpicRepository) GetDB() *gorm.DB {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*gorm.DB)
}

func (m *MockEpicRepository) GetWithUserStories(id uuid.UUID) (*models.Epic, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Epic), args.Error(1)
}

func (m *MockEpicRepository) GetByCreator(creatorID uuid.UUID) ([]models.Epic, error) {
	args := m.Called(creatorID)
	return args.Get(0).([]models.Epic), args.Error(1)
}

func (m *MockEpicRepository) GetByAssignee(assigneeID uuid.UUID) ([]models.Epic, error) {
	args := m.Called(assigneeID)
	return args.Get(0).([]models.Epic), args.Error(1)
}

func (m *MockEpicRepository) GetByStatus(status models.EpicStatus) ([]models.Epic, error) {
	args := m.Called(status)
	return args.Get(0).([]models.Epic), args.Error(1)
}

func (m *MockEpicRepository) GetByPriority(priority models.Priority) ([]models.Epic, error) {
	args := m.Called(priority)
	return args.Get(0).([]models.Epic), args.Error(1)
}

func (m *MockEpicRepository) HasUserStories(id uuid.UUID) (bool, error) {
	args := m.Called(id)
	return args.Bool(0), args.Error(1)
}

func (m *MockEpicRepository) GetByIDWithUsers(id uuid.UUID) (*models.Epic, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Epic), args.Error(1)
}

func (m *MockEpicRepository) GetByReferenceIDWithUsers(referenceID string) (*models.Epic, error) {
	args := m.Called(referenceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Epic), args.Error(1)
}

// MockUserRepository is a mock implementation of UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(id uuid.UUID) (*models.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByReferenceID(referenceID string) (*models.User, error) {
	args := m.Called(referenceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) Update(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockUserRepository) List(filters map[string]interface{}, orderBy string, limit, offset int) ([]models.User, error) {
	args := m.Called(filters, orderBy, limit, offset)
	return args.Get(0).([]models.User), args.Error(1)
}

func (m *MockUserRepository) Count(filters map[string]interface{}) (int64, error) {
	args := m.Called(filters)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockUserRepository) Exists(id uuid.UUID) (bool, error) {
	args := m.Called(id)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepository) ExistsByReferenceID(referenceID string) (bool, error) {
	args := m.Called(referenceID)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepository) WithTransaction(fn func(*gorm.DB) error) error {
	args := m.Called(fn)
	return args.Error(0)
}

func (m *MockUserRepository) GetDB() *gorm.DB {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*gorm.DB)
}

func (m *MockUserRepository) GetByUsername(username string) (*models.User, error) {
	args := m.Called(username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(email string) (*models.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) ExistsByUsername(username string) (bool, error) {
	args := m.Called(username)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepository) ExistsByEmail(email string) (bool, error) {
	args := m.Called(email)
	return args.Bool(0), args.Error(1)
}

func TestEpicService_CreateEpic(t *testing.T) {
	tests := []struct {
		name          string
		request       CreateEpicRequest
		setupMocks    func(*MockEpicRepository, *MockUserRepository)
		expectedError error
	}{
		{
			name: "successful epic creation",
			request: CreateEpicRequest{
				CreatorID:   uuid.New(),
				Priority:    models.PriorityHigh,
				Title:       "Test Epic",
				Description: stringPtr("Test Description"),
			},
			setupMocks: func(epicRepo *MockEpicRepository, userRepo *MockUserRepository) {
				userRepo.On("Exists", mock.AnythingOfType("uuid.UUID")).Return(true, nil)
				epicRepo.On("Create", mock.AnythingOfType("*models.Epic")).Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "creator not found",
			request: CreateEpicRequest{
				CreatorID: uuid.New(),
				Priority:  models.PriorityHigh,
				Title:     "Test Epic",
			},
			setupMocks: func(epicRepo *MockEpicRepository, userRepo *MockUserRepository) {
				userRepo.On("Exists", mock.AnythingOfType("uuid.UUID")).Return(false, nil)
			},
			expectedError: ErrUserNotFound,
		},
		{
			name: "invalid priority",
			request: CreateEpicRequest{
				CreatorID: uuid.New(),
				Priority:  models.Priority(5), // Invalid priority
				Title:     "Test Epic",
			},
			setupMocks: func(epicRepo *MockEpicRepository, userRepo *MockUserRepository) {
				// No mocks needed as validation happens before repository calls
			},
			expectedError: ErrInvalidPriority,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			epicRepo := new(MockEpicRepository)
			userRepo := new(MockUserRepository)

			tt.setupMocks(epicRepo, userRepo)

			service := NewEpicService(epicRepo, userRepo)

			epic, err := service.CreateEpic(tt.request)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tt.expectedError))
				assert.Nil(t, epic)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, epic)
				assert.Equal(t, tt.request.Title, epic.Title)
				assert.Equal(t, tt.request.Priority, epic.Priority)
				assert.Equal(t, models.EpicStatusBacklog, epic.Status)
			}

			epicRepo.AssertExpectations(t)
			userRepo.AssertExpectations(t)
		})
	}
}

func TestEpicService_GetEpicByID(t *testing.T) {
	tests := []struct {
		name          string
		epicID        uuid.UUID
		setupMocks    func(*MockEpicRepository, *MockUserRepository)
		expectedError error
	}{
		{
			name:   "successful get epic",
			epicID: uuid.New(),
			setupMocks: func(epicRepo *MockEpicRepository, userRepo *MockUserRepository) {
				epic := &models.Epic{
					ID:    uuid.New(),
					Title: "Test Epic",
				}
				epicRepo.On("GetByIDWithUsers", mock.AnythingOfType("uuid.UUID")).Return(epic, nil)
			},
			expectedError: nil,
		},
		{
			name:   "epic not found",
			epicID: uuid.New(),
			setupMocks: func(epicRepo *MockEpicRepository, userRepo *MockUserRepository) {
				epicRepo.On("GetByIDWithUsers", mock.AnythingOfType("uuid.UUID")).Return(nil, repository.ErrNotFound)
			},
			expectedError: ErrEpicNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			epicRepo := new(MockEpicRepository)
			userRepo := new(MockUserRepository)

			tt.setupMocks(epicRepo, userRepo)

			service := NewEpicService(epicRepo, userRepo)

			epic, err := service.GetEpicByID(tt.epicID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tt.expectedError))
				assert.Nil(t, epic)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, epic)
			}

			epicRepo.AssertExpectations(t)
			userRepo.AssertExpectations(t)
		})
	}
}

func TestEpicService_DeleteEpic(t *testing.T) {
	tests := []struct {
		name          string
		epicID        uuid.UUID
		force         bool
		setupMocks    func(*MockEpicRepository, *MockUserRepository)
		expectedError error
	}{
		{
			name:   "successful delete without user stories",
			epicID: uuid.New(),
			force:  false,
			setupMocks: func(epicRepo *MockEpicRepository, userRepo *MockUserRepository) {
				epic := &models.Epic{ID: uuid.New()}
				epicRepo.On("GetByID", mock.AnythingOfType("uuid.UUID")).Return(epic, nil)
				epicRepo.On("HasUserStories", mock.AnythingOfType("uuid.UUID")).Return(false, nil)
				epicRepo.On("Delete", mock.AnythingOfType("uuid.UUID")).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:   "delete with user stories but force=true",
			epicID: uuid.New(),
			force:  true,
			setupMocks: func(epicRepo *MockEpicRepository, userRepo *MockUserRepository) {
				epic := &models.Epic{ID: uuid.New()}
				epicRepo.On("GetByID", mock.AnythingOfType("uuid.UUID")).Return(epic, nil)
				epicRepo.On("Delete", mock.AnythingOfType("uuid.UUID")).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:   "delete with user stories and force=false",
			epicID: uuid.New(),
			force:  false,
			setupMocks: func(epicRepo *MockEpicRepository, userRepo *MockUserRepository) {
				epic := &models.Epic{ID: uuid.New()}
				epicRepo.On("GetByID", mock.AnythingOfType("uuid.UUID")).Return(epic, nil)
				epicRepo.On("HasUserStories", mock.AnythingOfType("uuid.UUID")).Return(true, nil)
			},
			expectedError: ErrEpicHasUserStories,
		},
		{
			name:   "epic not found",
			epicID: uuid.New(),
			force:  false,
			setupMocks: func(epicRepo *MockEpicRepository, userRepo *MockUserRepository) {
				epicRepo.On("GetByID", mock.AnythingOfType("uuid.UUID")).Return(nil, repository.ErrNotFound)
			},
			expectedError: ErrEpicNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			epicRepo := new(MockEpicRepository)
			userRepo := new(MockUserRepository)

			tt.setupMocks(epicRepo, userRepo)

			service := NewEpicService(epicRepo, userRepo)

			err := service.DeleteEpic(tt.epicID, tt.force)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tt.expectedError))
			} else {
				assert.NoError(t, err)
			}

			epicRepo.AssertExpectations(t)
			userRepo.AssertExpectations(t)
		})
	}
}

func TestEpicService_ChangeEpicStatus(t *testing.T) {
	tests := []struct {
		name          string
		epicID        uuid.UUID
		newStatus     models.EpicStatus
		setupMocks    func(*MockEpicRepository, *MockUserRepository)
		expectedError error
	}{
		{
			name:      "successful status change",
			epicID:    uuid.New(),
			newStatus: models.EpicStatusInProgress,
			setupMocks: func(epicRepo *MockEpicRepository, userRepo *MockUserRepository) {
				epic := &models.Epic{
					ID:     uuid.New(),
					Status: models.EpicStatusBacklog,
				}
				epicRepo.On("GetByID", mock.AnythingOfType("uuid.UUID")).Return(epic, nil)
				epicRepo.On("Update", mock.AnythingOfType("*models.Epic")).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:      "invalid status",
			epicID:    uuid.New(),
			newStatus: models.EpicStatus("InvalidStatus"),
			setupMocks: func(epicRepo *MockEpicRepository, userRepo *MockUserRepository) {
				epic := &models.Epic{
					ID:     uuid.New(),
					Status: models.EpicStatusBacklog,
				}
				epicRepo.On("GetByID", mock.AnythingOfType("uuid.UUID")).Return(epic, nil)
			},
			expectedError: ErrInvalidEpicStatus,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			epicRepo := new(MockEpicRepository)
			userRepo := new(MockUserRepository)

			tt.setupMocks(epicRepo, userRepo)

			service := NewEpicService(epicRepo, userRepo)

			epic, err := service.ChangeEpicStatus(tt.epicID, tt.newStatus)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tt.expectedError))
				assert.Nil(t, epic)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, epic)
				assert.Equal(t, tt.newStatus, epic.Status)
			}

			epicRepo.AssertExpectations(t)
			userRepo.AssertExpectations(t)
		})
	}
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}

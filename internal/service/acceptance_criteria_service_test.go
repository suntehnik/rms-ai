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

// MockAcceptanceCriteriaRepository is a mock implementation of AcceptanceCriteriaRepository
type MockAcceptanceCriteriaRepository struct {
	mock.Mock
}

func (m *MockAcceptanceCriteriaRepository) Create(entity *models.AcceptanceCriteria) error {
	args := m.Called(entity)
	return args.Error(0)
}

func (m *MockAcceptanceCriteriaRepository) GetByID(id uuid.UUID) (*models.AcceptanceCriteria, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AcceptanceCriteria), args.Error(1)
}

func (m *MockAcceptanceCriteriaRepository) GetByReferenceID(referenceID string) (*models.AcceptanceCriteria, error) {
	args := m.Called(referenceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AcceptanceCriteria), args.Error(1)
}

func (m *MockAcceptanceCriteriaRepository) Update(entity *models.AcceptanceCriteria) error {
	args := m.Called(entity)
	return args.Error(0)
}

func (m *MockAcceptanceCriteriaRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockAcceptanceCriteriaRepository) List(filters map[string]interface{}, orderBy string, limit, offset int) ([]models.AcceptanceCriteria, error) {
	args := m.Called(filters, orderBy, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.AcceptanceCriteria), args.Error(1)
}

func (m *MockAcceptanceCriteriaRepository) Count(filters map[string]interface{}) (int64, error) {
	args := m.Called(filters)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockAcceptanceCriteriaRepository) Exists(id uuid.UUID) (bool, error) {
	args := m.Called(id)
	return args.Get(0).(bool), args.Error(1)
}

func (m *MockAcceptanceCriteriaRepository) ExistsByReferenceID(referenceID string) (bool, error) {
	args := m.Called(referenceID)
	return args.Get(0).(bool), args.Error(1)
}

func (m *MockAcceptanceCriteriaRepository) WithTransaction(fn func(*gorm.DB) error) error {
	args := m.Called(fn)
	return args.Error(0)
}

func (m *MockAcceptanceCriteriaRepository) GetDB() *gorm.DB {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*gorm.DB)
}

func (m *MockAcceptanceCriteriaRepository) GetByUserStory(userStoryID uuid.UUID) ([]models.AcceptanceCriteria, error) {
	args := m.Called(userStoryID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.AcceptanceCriteria), args.Error(1)
}

func (m *MockAcceptanceCriteriaRepository) GetByAuthor(authorID uuid.UUID) ([]models.AcceptanceCriteria, error) {
	args := m.Called(authorID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.AcceptanceCriteria), args.Error(1)
}

func (m *MockAcceptanceCriteriaRepository) HasRequirements(id uuid.UUID) (bool, error) {
	args := m.Called(id)
	return args.Get(0).(bool), args.Error(1)
}

func (m *MockAcceptanceCriteriaRepository) CountByUserStory(userStoryID uuid.UUID) (int64, error) {
	args := m.Called(userStoryID)
	return args.Get(0).(int64), args.Error(1)
}

func TestAcceptanceCriteriaService_CreateAcceptanceCriteria(t *testing.T) {
	mockAcceptanceCriteriaRepo := new(MockAcceptanceCriteriaRepository)
	mockUserStoryRepo := new(MockUserStoryRepository)
	mockUserRepo := new(MockUserRepository)

	service := NewAcceptanceCriteriaService(mockAcceptanceCriteriaRepo, mockUserStoryRepo, mockUserRepo)

	userStoryID := uuid.New()
	authorID := uuid.New()

	tests := []struct {
		name          string
		request       CreateAcceptanceCriteriaRequest
		setupMocks    func()
		expectedError error
	}{
		{
			name: "successful creation",
			request: CreateAcceptanceCriteriaRequest{
				UserStoryID: userStoryID,
				AuthorID:    authorID,
				Description: "WHEN user clicks submit THEN system SHALL validate the form",
			},
			setupMocks: func() {
				mockUserStoryRepo.On("Exists", userStoryID).Return(true, nil)
				mockUserRepo.On("Exists", authorID).Return(true, nil)
				mockAcceptanceCriteriaRepo.On("Create", mock.AnythingOfType("*models.AcceptanceCriteria")).Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "user story not found",
			request: CreateAcceptanceCriteriaRequest{
				UserStoryID: userStoryID,
				AuthorID:    authorID,
				Description: "WHEN user clicks submit THEN system SHALL validate the form",
			},
			setupMocks: func() {
				mockUserStoryRepo.On("Exists", userStoryID).Return(false, nil)
			},
			expectedError: ErrUserStoryNotFound,
		},
		{
			name: "author not found",
			request: CreateAcceptanceCriteriaRequest{
				UserStoryID: userStoryID,
				AuthorID:    authorID,
				Description: "WHEN user clicks submit THEN system SHALL validate the form",
			},
			setupMocks: func() {
				mockUserStoryRepo.On("Exists", userStoryID).Return(true, nil)
				mockUserRepo.On("Exists", authorID).Return(false, nil)
			},
			expectedError: ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mocks
			mockAcceptanceCriteriaRepo.ExpectedCalls = nil
			mockUserStoryRepo.ExpectedCalls = nil
			mockUserRepo.ExpectedCalls = nil

			tt.setupMocks()

			result, err := service.CreateAcceptanceCriteria(tt.request)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tt.expectedError) || err.Error() == tt.expectedError.Error())
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.request.UserStoryID, result.UserStoryID)
				assert.Equal(t, tt.request.AuthorID, result.AuthorID)
				assert.Equal(t, tt.request.Description, result.Description)
			}

			mockAcceptanceCriteriaRepo.AssertExpectations(t)
			mockUserStoryRepo.AssertExpectations(t)
			mockUserRepo.AssertExpectations(t)
		})
	}
}

func TestAcceptanceCriteriaService_GetAcceptanceCriteriaByID(t *testing.T) {
	mockAcceptanceCriteriaRepo := new(MockAcceptanceCriteriaRepository)
	mockUserStoryRepo := new(MockUserStoryRepository)
	mockUserRepo := new(MockUserRepository)

	service := NewAcceptanceCriteriaService(mockAcceptanceCriteriaRepo, mockUserStoryRepo, mockUserRepo)

	acceptanceCriteriaID := uuid.New()
	expectedAcceptanceCriteria := &models.AcceptanceCriteria{
		ID:          acceptanceCriteriaID,
		ReferenceID: "AC-001",
		Description: "WHEN user clicks submit THEN system SHALL validate the form",
	}

	tests := []struct {
		name          string
		id            uuid.UUID
		setupMocks    func()
		expectedError error
	}{
		{
			name: "successful retrieval",
			id:   acceptanceCriteriaID,
			setupMocks: func() {
				mockAcceptanceCriteriaRepo.On("GetByID", acceptanceCriteriaID).Return(expectedAcceptanceCriteria, nil)
			},
			expectedError: nil,
		},
		{
			name: "acceptance criteria not found",
			id:   acceptanceCriteriaID,
			setupMocks: func() {
				mockAcceptanceCriteriaRepo.On("GetByID", acceptanceCriteriaID).Return(nil, repository.ErrNotFound)
			},
			expectedError: ErrAcceptanceCriteriaNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mocks
			mockAcceptanceCriteriaRepo.ExpectedCalls = nil

			tt.setupMocks()

			result, err := service.GetAcceptanceCriteriaByID(tt.id)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tt.expectedError))
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, expectedAcceptanceCriteria.ID, result.ID)
			}

			mockAcceptanceCriteriaRepo.AssertExpectations(t)
		})
	}
}

func TestAcceptanceCriteriaService_DeleteAcceptanceCriteria(t *testing.T) {
	mockAcceptanceCriteriaRepo := new(MockAcceptanceCriteriaRepository)
	mockUserStoryRepo := new(MockUserStoryRepository)
	mockUserRepo := new(MockUserRepository)

	service := NewAcceptanceCriteriaService(mockAcceptanceCriteriaRepo, mockUserStoryRepo, mockUserRepo)

	acceptanceCriteriaID := uuid.New()
	userStoryID := uuid.New()
	acceptanceCriteria := &models.AcceptanceCriteria{
		ID:          acceptanceCriteriaID,
		UserStoryID: userStoryID,
		ReferenceID: "AC-001",
		Description: "WHEN user clicks submit THEN system SHALL validate the form",
	}

	tests := []struct {
		name          string
		id            uuid.UUID
		force         bool
		setupMocks    func()
		expectedError error
	}{
		{
			name:  "successful deletion",
			id:    acceptanceCriteriaID,
			force: false,
			setupMocks: func() {
				mockAcceptanceCriteriaRepo.On("GetByID", acceptanceCriteriaID).Return(acceptanceCriteria, nil)
				mockAcceptanceCriteriaRepo.On("HasRequirements", acceptanceCriteriaID).Return(false, nil)
				mockAcceptanceCriteriaRepo.On("CountByUserStory", userStoryID).Return(int64(2), nil)
				mockAcceptanceCriteriaRepo.On("Delete", acceptanceCriteriaID).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:  "acceptance criteria not found",
			id:    acceptanceCriteriaID,
			force: false,
			setupMocks: func() {
				mockAcceptanceCriteriaRepo.On("GetByID", acceptanceCriteriaID).Return(nil, repository.ErrNotFound)
			},
			expectedError: ErrAcceptanceCriteriaNotFound,
		},
		{
			name:  "has requirements without force",
			id:    acceptanceCriteriaID,
			force: false,
			setupMocks: func() {
				mockAcceptanceCriteriaRepo.On("GetByID", acceptanceCriteriaID).Return(acceptanceCriteria, nil)
				mockAcceptanceCriteriaRepo.On("HasRequirements", acceptanceCriteriaID).Return(true, nil)
			},
			expectedError: ErrAcceptanceCriteriaHasRequirements,
		},
		{
			name:  "last acceptance criteria without force",
			id:    acceptanceCriteriaID,
			force: false,
			setupMocks: func() {
				mockAcceptanceCriteriaRepo.On("GetByID", acceptanceCriteriaID).Return(acceptanceCriteria, nil)
				mockAcceptanceCriteriaRepo.On("HasRequirements", acceptanceCriteriaID).Return(false, nil)
				mockAcceptanceCriteriaRepo.On("CountByUserStory", userStoryID).Return(int64(1), nil)
			},
			expectedError: ErrUserStoryMustHaveAcceptanceCriteria,
		},
		{
			name:  "force deletion with requirements",
			id:    acceptanceCriteriaID,
			force: true,
			setupMocks: func() {
				mockAcceptanceCriteriaRepo.On("GetByID", acceptanceCriteriaID).Return(acceptanceCriteria, nil)
				mockAcceptanceCriteriaRepo.On("Delete", acceptanceCriteriaID).Return(nil)
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mocks
			mockAcceptanceCriteriaRepo.ExpectedCalls = nil

			tt.setupMocks()

			err := service.DeleteAcceptanceCriteria(tt.id, tt.force)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tt.expectedError))
			} else {
				assert.NoError(t, err)
			}

			mockAcceptanceCriteriaRepo.AssertExpectations(t)
		})
	}
}

func TestAcceptanceCriteriaService_ValidateUserStoryHasAcceptanceCriteria(t *testing.T) {
	mockAcceptanceCriteriaRepo := new(MockAcceptanceCriteriaRepository)
	mockUserStoryRepo := new(MockUserStoryRepository)
	mockUserRepo := new(MockUserRepository)

	service := NewAcceptanceCriteriaService(mockAcceptanceCriteriaRepo, mockUserStoryRepo, mockUserRepo)

	userStoryID := uuid.New()

	tests := []struct {
		name          string
		userStoryID   uuid.UUID
		setupMocks    func()
		expectedError error
	}{
		{
			name:        "user story has acceptance criteria",
			userStoryID: userStoryID,
			setupMocks: func() {
				mockAcceptanceCriteriaRepo.On("CountByUserStory", userStoryID).Return(int64(2), nil)
			},
			expectedError: nil,
		},
		{
			name:        "user story has no acceptance criteria",
			userStoryID: userStoryID,
			setupMocks: func() {
				mockAcceptanceCriteriaRepo.On("CountByUserStory", userStoryID).Return(int64(0), nil)
			},
			expectedError: ErrUserStoryMustHaveAcceptanceCriteria,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mocks
			mockAcceptanceCriteriaRepo.ExpectedCalls = nil

			tt.setupMocks()

			err := service.ValidateUserStoryHasAcceptanceCriteria(tt.userStoryID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tt.expectedError))
			} else {
				assert.NoError(t, err)
			}

			mockAcceptanceCriteriaRepo.AssertExpectations(t)
		})
	}
}

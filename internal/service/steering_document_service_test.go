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

// MockSteeringDocumentRepository is a mock implementation of SteeringDocumentRepository
type MockSteeringDocumentRepository struct {
	mock.Mock
}

func (m *MockSteeringDocumentRepository) Create(entity *models.SteeringDocument) error {
	args := m.Called(entity)
	return args.Error(0)
}

func (m *MockSteeringDocumentRepository) GetByID(id uuid.UUID) (*models.SteeringDocument, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.SteeringDocument), args.Error(1)
}

func (m *MockSteeringDocumentRepository) GetByReferenceID(referenceID string) (*models.SteeringDocument, error) {
	args := m.Called(referenceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.SteeringDocument), args.Error(1)
}

func (m *MockSteeringDocumentRepository) Update(entity *models.SteeringDocument) error {
	args := m.Called(entity)
	return args.Error(0)
}

func (m *MockSteeringDocumentRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockSteeringDocumentRepository) ListWithFilters(filters repository.SteeringDocumentFilters) ([]models.SteeringDocument, int64, error) {
	args := m.Called(filters)
	return args.Get(0).([]models.SteeringDocument), args.Get(1).(int64), args.Error(2)
}

func (m *MockSteeringDocumentRepository) Search(query string) ([]models.SteeringDocument, error) {
	args := m.Called(query)
	return args.Get(0).([]models.SteeringDocument), args.Error(1)
}

func (m *MockSteeringDocumentRepository) GetByEpicID(epicID uuid.UUID) ([]models.SteeringDocument, error) {
	args := m.Called(epicID)
	return args.Get(0).([]models.SteeringDocument), args.Error(1)
}

func (m *MockSteeringDocumentRepository) LinkToEpic(steeringDocumentID, epicID uuid.UUID) error {
	args := m.Called(steeringDocumentID, epicID)
	return args.Error(0)
}

func (m *MockSteeringDocumentRepository) UnlinkFromEpic(steeringDocumentID, epicID uuid.UUID) error {
	args := m.Called(steeringDocumentID, epicID)
	return args.Error(0)
}

func (m *MockSteeringDocumentRepository) Count(filters map[string]interface{}) (int64, error) {
	args := m.Called(filters)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockSteeringDocumentRepository) List(filters map[string]interface{}, orderBy string, limit, offset int) ([]models.SteeringDocument, error) {
	args := m.Called(filters, orderBy, limit, offset)
	return args.Get(0).([]models.SteeringDocument), args.Error(1)
}

func (m *MockSteeringDocumentRepository) ExistsByReferenceID(referenceID string) (bool, error) {
	args := m.Called(referenceID)
	return args.Bool(0), args.Error(1)
}

func (m *MockSteeringDocumentRepository) Exists(id uuid.UUID) (bool, error) {
	args := m.Called(id)
	return args.Bool(0), args.Error(1)
}

func (m *MockSteeringDocumentRepository) WithTransaction(fn func(*gorm.DB) error) error {
	args := m.Called(fn)
	return args.Error(0)
}

func (m *MockSteeringDocumentRepository) GetDB() *gorm.DB {
	args := m.Called()
	return args.Get(0).(*gorm.DB)
}

// MockSteeringUserRepository is a mock implementation of UserRepository for steering document tests
type MockSteeringUserRepository struct {
	mock.Mock
}

func (m *MockSteeringUserRepository) GetByID(id uuid.UUID) (*models.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockSteeringUserRepository) Exists(id uuid.UUID) (bool, error) {
	args := m.Called(id)
	return args.Bool(0), args.Error(1)
}

// Add other required methods as stubs
func (m *MockSteeringUserRepository) Create(entity *models.User) error { return nil }
func (m *MockSteeringUserRepository) GetByReferenceID(referenceID string) (*models.User, error) {
	return nil, nil
}
func (m *MockSteeringUserRepository) Update(entity *models.User) error { return nil }
func (m *MockSteeringUserRepository) Delete(id uuid.UUID) error        { return nil }
func (m *MockSteeringUserRepository) List(filters map[string]interface{}, orderBy string, limit, offset int) ([]models.User, error) {
	return nil, nil
}
func (m *MockSteeringUserRepository) Count(filters map[string]interface{}) (int64, error) {
	return 0, nil
}
func (m *MockSteeringUserRepository) ExistsByReferenceID(referenceID string) (bool, error) {
	return false, nil
}
func (m *MockSteeringUserRepository) WithTransaction(fn func(*gorm.DB) error) error { return nil }
func (m *MockSteeringUserRepository) GetDB() *gorm.DB                               { return nil }
func (m *MockSteeringUserRepository) GetByUsername(username string) (*models.User, error) {
	return nil, nil
}
func (m *MockSteeringUserRepository) GetByEmail(email string) (*models.User, error) { return nil, nil }
func (m *MockSteeringUserRepository) ExistsByUsername(username string) (bool, error) {
	return false, nil
}
func (m *MockSteeringUserRepository) ExistsByEmail(email string) (bool, error) { return false, nil }

// MockSteeringEpicRepository is a mock implementation of EpicRepository for steering document tests
type MockSteeringEpicRepository struct {
	mock.Mock
}

func (m *MockSteeringEpicRepository) Exists(id uuid.UUID) (bool, error) {
	args := m.Called(id)
	return args.Bool(0), args.Error(1)
}

// Add other required methods as stubs
func (m *MockSteeringEpicRepository) Create(entity *models.Epic) error { return nil }
func (m *MockSteeringEpicRepository) GetByID(id uuid.UUID) (*models.Epic, error) {
	return nil, nil
}
func (m *MockSteeringEpicRepository) GetByReferenceID(referenceID string) (*models.Epic, error) {
	return nil, nil
}
func (m *MockSteeringEpicRepository) Update(entity *models.Epic) error { return nil }
func (m *MockSteeringEpicRepository) Delete(id uuid.UUID) error        { return nil }
func (m *MockSteeringEpicRepository) List(filters map[string]interface{}, orderBy string, limit, offset int) ([]models.Epic, error) {
	return nil, nil
}
func (m *MockSteeringEpicRepository) Count(filters map[string]interface{}) (int64, error) {
	return 0, nil
}
func (m *MockSteeringEpicRepository) ExistsByReferenceID(referenceID string) (bool, error) {
	return false, nil
}
func (m *MockSteeringEpicRepository) WithTransaction(fn func(*gorm.DB) error) error { return nil }
func (m *MockSteeringEpicRepository) GetDB() *gorm.DB                               { return nil }
func (m *MockSteeringEpicRepository) GetWithUserStories(id uuid.UUID) (*models.Epic, error) {
	return nil, nil
}
func (m *MockSteeringEpicRepository) GetUserStoriesByEpicID(epicID uuid.UUID) ([]models.UserStory, error) {
	return nil, nil
}
func (m *MockSteeringEpicRepository) GetByCreator(creatorID uuid.UUID) ([]models.Epic, error) {
	return nil, nil
}
func (m *MockSteeringEpicRepository) GetByAssignee(assigneeID uuid.UUID) ([]models.Epic, error) {
	return nil, nil
}
func (m *MockSteeringEpicRepository) GetByStatus(status models.EpicStatus) ([]models.Epic, error) {
	return nil, nil
}
func (m *MockSteeringEpicRepository) GetByPriority(priority models.Priority) ([]models.Epic, error) {
	return nil, nil
}
func (m *MockSteeringEpicRepository) HasUserStories(id uuid.UUID) (bool, error) {
	return false, nil
}
func (m *MockSteeringEpicRepository) GetByIDWithUsers(id uuid.UUID) (*models.Epic, error) {
	return nil, nil
}
func (m *MockSteeringEpicRepository) GetByReferenceIDWithUsers(referenceID string) (*models.Epic, error) {
	return nil, nil
}
func (m *MockSteeringEpicRepository) ListWithIncludes(filters map[string]interface{}, includes []string, orderBy string, limit, offset int) ([]models.Epic, error) {
	return nil, nil
}

func TestNewSteeringDocumentService(t *testing.T) {
	mockRepo := &MockSteeringDocumentRepository{}
	mockUserRepo := &MockSteeringUserRepository{}
	mockEpicRepo := &MockSteeringEpicRepository{}

	service := NewSteeringDocumentService(mockRepo, mockEpicRepo, mockUserRepo)

	assert.NotNil(t, service)
}

func TestSteeringDocumentService_CreateSteeringDocument_Success(t *testing.T) {
	mockRepo := &MockSteeringDocumentRepository{}
	mockUserRepo := &MockSteeringUserRepository{}
	mockEpicRepo := &MockSteeringEpicRepository{}
	service := NewSteeringDocumentService(mockRepo, mockEpicRepo, mockUserRepo)

	// Create test user
	user := &models.User{
		ID:   uuid.New(),
		Role: models.RoleUser,
	}

	// Create request
	req := CreateSteeringDocumentRequest{
		Title:       "Test Document",
		Description: steeringStringPtr("Test description"),
	}

	// Mock expectations
	mockUserRepo.On("Exists", user.ID).Return(true, nil)
	mockRepo.On("Create", mock.AnythingOfType("*models.SteeringDocument")).Return(nil).Run(func(args mock.Arguments) {
		doc := args.Get(0).(*models.SteeringDocument)
		doc.ID = uuid.New()
		doc.ReferenceID = "STD-001"
	})
	mockRepo.On("GetByID", mock.AnythingOfType("uuid.UUID")).Return(&models.SteeringDocument{
		ID:          uuid.New(),
		ReferenceID: "STD-001",
		Title:       req.Title,
		Description: req.Description,
		CreatorID:   user.ID,
	}, nil)

	// Execute
	result, err := service.CreateSteeringDocument(req, user)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, req.Title, result.Title)
	assert.Equal(t, req.Description, result.Description)
	assert.Equal(t, user.ID, result.CreatorID)
	mockRepo.AssertExpectations(t)
}

func TestSteeringDocumentService_CreateSteeringDocument_ValidationError(t *testing.T) {
	mockRepo := &MockSteeringDocumentRepository{}
	mockUserRepo := &MockSteeringUserRepository{}
	mockEpicRepo := &MockSteeringEpicRepository{}
	service := NewSteeringDocumentService(mockRepo, mockEpicRepo, mockUserRepo)

	user := &models.User{ID: uuid.New(), Role: models.RoleUser}

	// Setup mock for user existence check
	mockUserRepo.On("Exists", user.ID).Return(true, nil)

	tests := []struct {
		name    string
		request CreateSteeringDocumentRequest
		wantErr string
	}{
		{
			name:    "empty title",
			request: CreateSteeringDocumentRequest{Title: ""},
			wantErr: "title is required",
		},
		{
			name:    "title too long",
			request: CreateSteeringDocumentRequest{Title: string(make([]byte, 501))},
			wantErr: "title must be at most 500 characters",
		},
		{
			name:    "description too long",
			request: CreateSteeringDocumentRequest{Title: "Valid", Description: steeringStringPtr(string(make([]byte, 50001)))},
			wantErr: "description must be at most 50000 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.CreateSteeringDocument(tt.request, user)
			assert.Error(t, err)
			assert.Nil(t, result)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestSteeringDocumentService_GetSteeringDocumentByID_Success(t *testing.T) {
	mockRepo := &MockSteeringDocumentRepository{}
	mockUserRepo := &MockSteeringUserRepository{}
	mockEpicRepo := &MockSteeringEpicRepository{}
	service := NewSteeringDocumentService(mockRepo, mockEpicRepo, mockUserRepo)

	// Create test data
	docID := uuid.New()
	userID := uuid.New()
	user := &models.User{ID: userID, Role: models.RoleUser}
	doc := &models.SteeringDocument{
		ID:        docID,
		Title:     "Test Document",
		CreatorID: userID,
	}

	// Mock expectations
	mockRepo.On("GetByID", docID).Return(doc, nil)

	// Execute
	result, err := service.GetSteeringDocumentByID(docID, user)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, doc, result)
	mockRepo.AssertExpectations(t)
}

func TestSteeringDocumentService_GetSteeringDocumentByID_NotFound(t *testing.T) {
	mockRepo := &MockSteeringDocumentRepository{}
	mockUserRepo := &MockSteeringUserRepository{}
	mockEpicRepo := &MockSteeringEpicRepository{}
	service := NewSteeringDocumentService(mockRepo, mockEpicRepo, mockUserRepo)

	docID := uuid.New()
	user := &models.User{ID: uuid.New(), Role: models.RoleUser}

	// Mock expectations
	mockRepo.On("GetByID", docID).Return(nil, repository.ErrNotFound)

	// Execute
	result, err := service.GetSteeringDocumentByID(docID, user)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, ErrSteeringDocumentNotFound, err)
	mockRepo.AssertExpectations(t)
}

func TestSteeringDocumentService_GetSteeringDocumentByID_AccessDenied(t *testing.T) {
	mockRepo := &MockSteeringDocumentRepository{}
	mockUserRepo := &MockSteeringUserRepository{}
	mockEpicRepo := &MockSteeringEpicRepository{}
	service := NewSteeringDocumentService(mockRepo, mockEpicRepo, mockUserRepo)

	// Create test data - user with no read access
	docID := uuid.New()
	userID := uuid.New()
	user := &models.User{ID: userID, Role: models.RoleCommenter} // Commenter role - no read access

	// No mock expectations needed - authorization check happens first

	// Execute
	result, err := service.GetSteeringDocumentByID(docID, user)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, ErrUnauthorizedAccess, err)
}

func TestSteeringDocumentService_GetSteeringDocumentByID_AdminAccess(t *testing.T) {
	mockRepo := &MockSteeringDocumentRepository{}
	mockUserRepo := &MockSteeringUserRepository{}
	mockEpicRepo := &MockSteeringEpicRepository{}
	service := NewSteeringDocumentService(mockRepo, mockEpicRepo, mockUserRepo)

	// Create test data - admin accessing any document
	docID := uuid.New()
	creatorID := uuid.New()
	adminID := uuid.New()
	admin := &models.User{ID: adminID, Role: models.RoleAdministrator}
	doc := &models.SteeringDocument{
		ID:        docID,
		Title:     "Test Document",
		CreatorID: creatorID, // Different creator
	}

	// Mock expectations
	mockRepo.On("GetByID", docID).Return(doc, nil)

	// Execute
	result, err := service.GetSteeringDocumentByID(docID, admin)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, doc, result)
	mockRepo.AssertExpectations(t)
}

func TestSteeringDocumentService_UpdateSteeringDocument_Success(t *testing.T) {
	mockRepo := &MockSteeringDocumentRepository{}
	mockUserRepo := &MockSteeringUserRepository{}
	mockEpicRepo := &MockSteeringEpicRepository{}
	service := NewSteeringDocumentService(mockRepo, mockEpicRepo, mockUserRepo)

	// Create test data
	docID := uuid.New()
	userID := uuid.New()
	user := &models.User{ID: userID, Role: models.RoleUser}
	doc := &models.SteeringDocument{
		ID:        docID,
		Title:     "Original Title",
		CreatorID: userID,
	}

	req := UpdateSteeringDocumentRequest{
		Title:       steeringStringPtr("Updated Title"),
		Description: steeringStringPtr("Updated description"),
	}

	// Mock expectations
	mockRepo.On("GetByID", docID).Return(doc, nil)
	mockRepo.On("Update", mock.AnythingOfType("*models.SteeringDocument")).Return(nil)

	// Execute
	result, err := service.UpdateSteeringDocument(docID, req, user)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, *req.Title, result.Title)
	assert.Equal(t, req.Description, result.Description)
	mockRepo.AssertExpectations(t)
}

func TestSteeringDocumentService_UpdateSteeringDocument_AccessDenied(t *testing.T) {
	mockRepo := &MockSteeringDocumentRepository{}
	mockUserRepo := &MockSteeringUserRepository{}
	mockEpicRepo := &MockSteeringEpicRepository{}
	service := NewSteeringDocumentService(mockRepo, mockEpicRepo, mockUserRepo)

	// Create test data - user trying to update document they didn't create
	docID := uuid.New()
	creatorID := uuid.New()
	userID := uuid.New()
	user := &models.User{ID: userID, Role: models.RoleUser}
	doc := &models.SteeringDocument{
		ID:        docID,
		Title:     "Original Title",
		CreatorID: creatorID, // Different creator
	}

	req := UpdateSteeringDocumentRequest{
		Title: steeringStringPtr("Updated Title"),
	}

	// Mock expectations
	mockRepo.On("GetByID", docID).Return(doc, nil)

	// Execute
	result, err := service.UpdateSteeringDocument(docID, req, user)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, ErrUnauthorizedAccess, err)
	mockRepo.AssertExpectations(t)
}

func TestSteeringDocumentService_DeleteSteeringDocument_Success(t *testing.T) {
	mockRepo := &MockSteeringDocumentRepository{}
	mockUserRepo := &MockSteeringUserRepository{}
	mockEpicRepo := &MockSteeringEpicRepository{}
	service := NewSteeringDocumentService(mockRepo, mockEpicRepo, mockUserRepo)

	// Create test data
	docID := uuid.New()
	userID := uuid.New()
	user := &models.User{ID: userID, Role: models.RoleUser}
	doc := &models.SteeringDocument{
		ID:        docID,
		Title:     "Test Document",
		CreatorID: userID,
	}

	// Mock expectations
	mockRepo.On("GetByID", docID).Return(doc, nil)
	mockRepo.On("Delete", docID).Return(nil)

	// Execute
	err := service.DeleteSteeringDocument(docID, user)

	// Assert
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestSteeringDocumentService_LinkSteeringDocumentToEpic_Success(t *testing.T) {
	mockRepo := &MockSteeringDocumentRepository{}
	mockUserRepo := &MockSteeringUserRepository{}
	mockEpicRepo := &MockSteeringEpicRepository{}
	service := NewSteeringDocumentService(mockRepo, mockEpicRepo, mockUserRepo)

	// Create test data
	docID := uuid.New()
	epicID := uuid.New()
	userID := uuid.New()
	user := &models.User{ID: userID, Role: models.RoleUser}
	doc := &models.SteeringDocument{
		ID:        docID,
		Title:     "Test Document",
		CreatorID: userID,
	}

	// Mock expectations
	mockRepo.On("GetByID", docID).Return(doc, nil)
	mockUserRepo.On("Exists", userID).Return(true, nil)
	mockEpicRepo.On("Exists", epicID).Return(true, nil)
	mockRepo.On("LinkToEpic", docID, epicID).Return(nil)

	// Execute
	err := service.LinkSteeringDocumentToEpic(docID, epicID, user)

	// Assert
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockEpicRepo.AssertExpectations(t)
}

func TestSteeringDocumentService_LinkSteeringDocumentToEpic_EpicNotFound(t *testing.T) {
	mockRepo := &MockSteeringDocumentRepository{}
	mockUserRepo := &MockSteeringUserRepository{}
	mockEpicRepo := &MockSteeringEpicRepository{}
	service := NewSteeringDocumentService(mockRepo, mockEpicRepo, mockUserRepo)

	// Create test data
	docID := uuid.New()
	epicID := uuid.New()
	userID := uuid.New()
	user := &models.User{ID: userID, Role: models.RoleUser}
	doc := &models.SteeringDocument{
		ID:        docID,
		Title:     "Test Document",
		CreatorID: userID,
	}

	// Mock expectations
	mockRepo.On("GetByID", docID).Return(doc, nil)
	mockUserRepo.On("Exists", userID).Return(true, nil)
	mockEpicRepo.On("Exists", epicID).Return(false, nil)

	// Execute
	err := service.LinkSteeringDocumentToEpic(docID, epicID, user)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, ErrEpicNotFound, err)
	mockRepo.AssertExpectations(t)
	mockEpicRepo.AssertExpectations(t)
}

func TestSteeringDocumentService_ListSteeringDocuments_Success(t *testing.T) {
	mockRepo := &MockSteeringDocumentRepository{}
	mockUserRepo := &MockSteeringUserRepository{}
	mockEpicRepo := &MockSteeringEpicRepository{}
	service := NewSteeringDocumentService(mockRepo, mockEpicRepo, mockUserRepo)

	// Create test data
	userID := uuid.New()
	user := &models.User{ID: userID, Role: models.RoleUser}
	docs := []models.SteeringDocument{
		{ID: uuid.New(), Title: "Doc 1", CreatorID: userID},
		{ID: uuid.New(), Title: "Doc 2", CreatorID: userID},
	}

	filters := SteeringDocumentFilters{
		Limit:  10,
		Offset: 0,
	}

	expectedRepoFilters := repository.SteeringDocumentFilters{
		CreatorID: &userID, // Should be filtered for regular users
		Limit:     10,
		Offset:    0,
	}

	// Mock expectations
	mockRepo.On("ListWithFilters", expectedRepoFilters).Return(docs, int64(2), nil)

	// Execute
	result, total, err := service.ListSteeringDocuments(filters, user)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, docs, result)
	assert.Equal(t, int64(2), total)
	mockRepo.AssertExpectations(t)
}

func TestSteeringDocumentService_ListSteeringDocuments_AdminAccess(t *testing.T) {
	mockRepo := &MockSteeringDocumentRepository{}
	mockUserRepo := &MockSteeringUserRepository{}
	mockEpicRepo := &MockSteeringEpicRepository{}
	service := NewSteeringDocumentService(mockRepo, mockEpicRepo, mockUserRepo)

	// Create test data
	adminID := uuid.New()
	admin := &models.User{ID: adminID, Role: models.RoleAdministrator}
	docs := []models.SteeringDocument{
		{ID: uuid.New(), Title: "Doc 1", CreatorID: uuid.New()},
		{ID: uuid.New(), Title: "Doc 2", CreatorID: uuid.New()},
	}

	filters := SteeringDocumentFilters{
		Limit:  10,
		Offset: 0,
	}

	expectedRepoFilters := repository.SteeringDocumentFilters{
		// No CreatorID filter for admin
		Limit:  10,
		Offset: 0,
	}

	// Mock expectations
	mockRepo.On("ListWithFilters", expectedRepoFilters).Return(docs, int64(2), nil)

	// Execute
	result, total, err := service.ListSteeringDocuments(filters, admin)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, docs, result)
	assert.Equal(t, int64(2), total)
	mockRepo.AssertExpectations(t)
}

// Helper function for steering document tests
func steeringStringPtr(s string) *string {
	return &s
}

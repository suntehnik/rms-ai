package handlers

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"

	"product-requirements-management/internal/models"
	"product-requirements-management/internal/service"
)

// MockSteeringDocumentService is a mock implementation of SteeringDocumentService
type MockSteeringDocumentService struct {
	mock.Mock
}

func (m *MockSteeringDocumentService) CreateSteeringDocument(req service.CreateSteeringDocumentRequest, currentUser *models.User) (*models.SteeringDocument, error) {
	args := m.Called(req, currentUser)
	return args.Get(0).(*models.SteeringDocument), args.Error(1)
}

func (m *MockSteeringDocumentService) GetSteeringDocumentByID(id uuid.UUID, currentUser *models.User) (*models.SteeringDocument, error) {
	args := m.Called(id, currentUser)
	return args.Get(0).(*models.SteeringDocument), args.Error(1)
}

func (m *MockSteeringDocumentService) GetSteeringDocumentByReferenceID(referenceID string, currentUser *models.User) (*models.SteeringDocument, error) {
	args := m.Called(referenceID, currentUser)
	return args.Get(0).(*models.SteeringDocument), args.Error(1)
}

func (m *MockSteeringDocumentService) UpdateSteeringDocument(id uuid.UUID, req service.UpdateSteeringDocumentRequest, currentUser *models.User) (*models.SteeringDocument, error) {
	args := m.Called(id, req, currentUser)
	return args.Get(0).(*models.SteeringDocument), args.Error(1)
}

func (m *MockSteeringDocumentService) DeleteSteeringDocument(id uuid.UUID, currentUser *models.User) error {
	args := m.Called(id, currentUser)
	return args.Error(0)
}

func (m *MockSteeringDocumentService) ListSteeringDocuments(filters service.SteeringDocumentFilters, currentUser *models.User) ([]models.SteeringDocument, int64, error) {
	args := m.Called(filters, currentUser)
	return args.Get(0).([]models.SteeringDocument), args.Get(1).(int64), args.Error(2)
}

func (m *MockSteeringDocumentService) SearchSteeringDocuments(query string, currentUser *models.User) ([]models.SteeringDocument, error) {
	args := m.Called(query, currentUser)
	return args.Get(0).([]models.SteeringDocument), args.Error(1)
}

func (m *MockSteeringDocumentService) GetSteeringDocumentsByEpicID(epicID uuid.UUID, currentUser *models.User) ([]models.SteeringDocument, error) {
	args := m.Called(epicID, currentUser)
	return args.Get(0).([]models.SteeringDocument), args.Error(1)
}

func (m *MockSteeringDocumentService) LinkSteeringDocumentToEpic(steeringDocumentID, epicID uuid.UUID, currentUser *models.User) error {
	args := m.Called(steeringDocumentID, epicID, currentUser)
	return args.Error(0)
}

func (m *MockSteeringDocumentService) UnlinkSteeringDocumentFromEpic(steeringDocumentID, epicID uuid.UUID, currentUser *models.User) error {
	args := m.Called(steeringDocumentID, epicID, currentUser)
	return args.Error(0)
}

// MockUserRepository is a mock implementation of UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) GetByID(id uuid.UUID) (*models.User, error) {
	args := m.Called(id)
	return args.Get(0).(*models.User), args.Error(1)
}

// Add other required methods as stubs
func (m *MockUserRepository) Create(entity *models.User) error { return nil }
func (m *MockUserRepository) GetByReferenceID(referenceID string) (*models.User, error) {
	return nil, nil
}
func (m *MockUserRepository) Update(entity *models.User) error { return nil }
func (m *MockUserRepository) Delete(id uuid.UUID) error        { return nil }
func (m *MockUserRepository) List(filters map[string]interface{}, orderBy string, limit, offset int) ([]models.User, error) {
	return nil, nil
}
func (m *MockUserRepository) Count(filters map[string]interface{}) (int64, error)  { return 0, nil }
func (m *MockUserRepository) Exists(id uuid.UUID) (bool, error)                    { return false, nil }
func (m *MockUserRepository) ExistsByReferenceID(referenceID string) (bool, error) { return false, nil }
func (m *MockUserRepository) WithTransaction(fn func(*gorm.DB) error) error        { return nil }
func (m *MockUserRepository) GetDB() *gorm.DB                                      { return nil }
func (m *MockUserRepository) GetByUsername(username string) (*models.User, error)  { return nil, nil }
func (m *MockUserRepository) GetByEmail(email string) (*models.User, error)        { return nil, nil }
func (m *MockUserRepository) ExistsByUsername(username string) (bool, error)       { return false, nil }
func (m *MockUserRepository) ExistsByEmail(email string) (bool, error)             { return false, nil }

func TestNewSteeringDocumentHandler(t *testing.T) {
	mockService := &MockSteeringDocumentService{}
	mockUserRepo := &MockUserRepository{}

	handler := NewSteeringDocumentHandler(mockService, mockUserRepo)

	assert.NotNil(t, handler)
	assert.Equal(t, mockService, handler.steeringDocumentService)
	assert.Equal(t, mockUserRepo, handler.userRepo)
}

func TestSteeringDocumentHandler_CreateSteeringDocument_Success(t *testing.T) {
	mockService := &MockSteeringDocumentService{}
	mockUserRepo := &MockUserRepository{}
	handler := NewSteeringDocumentHandler(mockService, mockUserRepo)

	// Create test user
	user := &models.User{
		ID:   uuid.New(),
		Role: models.RoleUser,
	}

	// Create test request
	req := service.CreateSteeringDocumentRequest{
		Title:       "Test Document",
		Description: steeringHandlerStringPtr("Test description"),
	}

	// Create expected response
	expectedDoc := &models.SteeringDocument{
		ID:          uuid.New(),
		ReferenceID: "STD-001",
		Title:       req.Title,
		Description: req.Description,
		CreatorID:   user.ID,
	}

	// Mock expectations
	mockService.On("CreateSteeringDocument", req, user).Return(expectedDoc, nil)

	// Execute
	result, err := handler.steeringDocumentService.CreateSteeringDocument(req, user)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedDoc, result)
	mockService.AssertExpectations(t)
}

func TestSteeringDocumentHandler_CreateSteeringDocument_ValidationError(t *testing.T) {
	mockService := &MockSteeringDocumentService{}
	mockUserRepo := &MockUserRepository{}
	handler := NewSteeringDocumentHandler(mockService, mockUserRepo)

	user := &models.User{ID: uuid.New(), Role: models.RoleUser}
	req := service.CreateSteeringDocumentRequest{Title: ""} // Invalid request

	// Mock expectations
	mockService.On("CreateSteeringDocument", req, user).Return((*models.SteeringDocument)(nil), service.ErrValidation)

	// Execute
	result, err := handler.steeringDocumentService.CreateSteeringDocument(req, user)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, service.ErrValidation, err)
	mockService.AssertExpectations(t)
}

func TestSteeringDocumentHandler_GetSteeringDocumentByID_Success(t *testing.T) {
	mockService := &MockSteeringDocumentService{}
	mockUserRepo := &MockUserRepository{}
	handler := NewSteeringDocumentHandler(mockService, mockUserRepo)

	// Create test data
	docID := uuid.New()
	user := &models.User{ID: uuid.New(), Role: models.RoleUser}
	expectedDoc := &models.SteeringDocument{
		ID:          docID,
		ReferenceID: "STD-001",
		Title:       "Test Document",
		CreatorID:   user.ID,
	}

	// Mock expectations
	mockService.On("GetSteeringDocumentByID", docID, user).Return(expectedDoc, nil)

	// Execute
	result, err := handler.steeringDocumentService.GetSteeringDocumentByID(docID, user)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedDoc, result)
	mockService.AssertExpectations(t)
}

func TestSteeringDocumentHandler_GetSteeringDocumentByID_NotFound(t *testing.T) {
	mockService := &MockSteeringDocumentService{}
	mockUserRepo := &MockUserRepository{}
	handler := NewSteeringDocumentHandler(mockService, mockUserRepo)

	docID := uuid.New()
	user := &models.User{ID: uuid.New(), Role: models.RoleUser}

	// Mock expectations
	mockService.On("GetSteeringDocumentByID", docID, user).Return((*models.SteeringDocument)(nil), service.ErrSteeringDocumentNotFound)

	// Execute
	result, err := handler.steeringDocumentService.GetSteeringDocumentByID(docID, user)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, service.ErrSteeringDocumentNotFound, err)
	mockService.AssertExpectations(t)
}

func TestSteeringDocumentHandler_GetSteeringDocumentByReferenceID_Success(t *testing.T) {
	mockService := &MockSteeringDocumentService{}
	mockUserRepo := &MockUserRepository{}
	handler := NewSteeringDocumentHandler(mockService, mockUserRepo)

	// Create test data
	referenceID := "STD-001"
	user := &models.User{ID: uuid.New(), Role: models.RoleUser}
	expectedDoc := &models.SteeringDocument{
		ID:          uuid.New(),
		ReferenceID: referenceID,
		Title:       "Test Document",
		CreatorID:   user.ID,
	}

	// Mock expectations
	mockService.On("GetSteeringDocumentByReferenceID", referenceID, user).Return(expectedDoc, nil)

	// Execute
	result, err := handler.steeringDocumentService.GetSteeringDocumentByReferenceID(referenceID, user)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedDoc, result)
	mockService.AssertExpectations(t)
}

func TestSteeringDocumentHandler_UpdateSteeringDocument_Success(t *testing.T) {
	mockService := &MockSteeringDocumentService{}
	mockUserRepo := &MockUserRepository{}
	handler := NewSteeringDocumentHandler(mockService, mockUserRepo)

	// Create test data
	docID := uuid.New()
	user := &models.User{ID: uuid.New(), Role: models.RoleUser}
	req := service.UpdateSteeringDocumentRequest{
		Title:       steeringHandlerStringPtr("Updated Title"),
		Description: steeringHandlerStringPtr("Updated description"),
	}
	expectedDoc := &models.SteeringDocument{
		ID:          docID,
		ReferenceID: "STD-001",
		Title:       *req.Title,
		Description: req.Description,
		CreatorID:   user.ID,
	}

	// Mock expectations
	mockService.On("UpdateSteeringDocument", docID, req, user).Return(expectedDoc, nil)

	// Execute
	result, err := handler.steeringDocumentService.UpdateSteeringDocument(docID, req, user)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedDoc, result)
	mockService.AssertExpectations(t)
}

func TestSteeringDocumentHandler_UpdateSteeringDocument_AccessDenied(t *testing.T) {
	mockService := &MockSteeringDocumentService{}
	mockUserRepo := &MockUserRepository{}
	handler := NewSteeringDocumentHandler(mockService, mockUserRepo)

	docID := uuid.New()
	user := &models.User{ID: uuid.New(), Role: models.RoleUser}
	req := service.UpdateSteeringDocumentRequest{Title: steeringHandlerStringPtr("Updated Title")}

	// Mock expectations
	mockService.On("UpdateSteeringDocument", docID, req, user).Return((*models.SteeringDocument)(nil), service.ErrInsufficientPermissions)

	// Execute
	result, err := handler.steeringDocumentService.UpdateSteeringDocument(docID, req, user)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, service.ErrInsufficientPermissions, err)
	mockService.AssertExpectations(t)
}

func TestSteeringDocumentHandler_DeleteSteeringDocument_Success(t *testing.T) {
	mockService := &MockSteeringDocumentService{}
	mockUserRepo := &MockUserRepository{}
	handler := NewSteeringDocumentHandler(mockService, mockUserRepo)

	docID := uuid.New()
	user := &models.User{ID: uuid.New(), Role: models.RoleUser}

	// Mock expectations
	mockService.On("DeleteSteeringDocument", docID, user).Return(nil)

	// Execute
	err := handler.steeringDocumentService.DeleteSteeringDocument(docID, user)

	// Assert
	assert.NoError(t, err)
	mockService.AssertExpectations(t)
}

func TestSteeringDocumentHandler_ListSteeringDocuments_Success(t *testing.T) {
	mockService := &MockSteeringDocumentService{}
	mockUserRepo := &MockUserRepository{}
	handler := NewSteeringDocumentHandler(mockService, mockUserRepo)

	user := &models.User{ID: uuid.New(), Role: models.RoleUser}
	filters := service.SteeringDocumentFilters{
		Limit:  10,
		Offset: 0,
	}
	expectedDocs := []models.SteeringDocument{
		{ID: uuid.New(), Title: "Doc 1", CreatorID: user.ID},
		{ID: uuid.New(), Title: "Doc 2", CreatorID: user.ID},
	}
	expectedTotal := int64(2)

	// Mock expectations
	mockService.On("ListSteeringDocuments", filters, user).Return(expectedDocs, expectedTotal, nil)

	// Execute
	result, total, err := handler.steeringDocumentService.ListSteeringDocuments(filters, user)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedDocs, result)
	assert.Equal(t, expectedTotal, total)
	mockService.AssertExpectations(t)
}

func TestSteeringDocumentHandler_SearchSteeringDocuments_Success(t *testing.T) {
	mockService := &MockSteeringDocumentService{}
	mockUserRepo := &MockUserRepository{}
	handler := NewSteeringDocumentHandler(mockService, mockUserRepo)

	user := &models.User{ID: uuid.New(), Role: models.RoleUser}
	query := "test query"
	expectedDocs := []models.SteeringDocument{
		{ID: uuid.New(), Title: "Test Document", CreatorID: user.ID},
	}

	// Mock expectations
	mockService.On("SearchSteeringDocuments", query, user).Return(expectedDocs, nil)

	// Execute
	result, err := handler.steeringDocumentService.SearchSteeringDocuments(query, user)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedDocs, result)
	mockService.AssertExpectations(t)
}

func TestSteeringDocumentHandler_LinkSteeringDocumentToEpic_Success(t *testing.T) {
	mockService := &MockSteeringDocumentService{}
	mockUserRepo := &MockUserRepository{}
	handler := NewSteeringDocumentHandler(mockService, mockUserRepo)

	docID := uuid.New()
	epicID := uuid.New()
	user := &models.User{ID: uuid.New(), Role: models.RoleUser}

	// Mock expectations
	mockService.On("LinkSteeringDocumentToEpic", docID, epicID, user).Return(nil)

	// Execute
	err := handler.steeringDocumentService.LinkSteeringDocumentToEpic(docID, epicID, user)

	// Assert
	assert.NoError(t, err)
	mockService.AssertExpectations(t)
}

func TestSteeringDocumentHandler_UnlinkSteeringDocumentFromEpic_Success(t *testing.T) {
	mockService := &MockSteeringDocumentService{}
	mockUserRepo := &MockUserRepository{}
	handler := NewSteeringDocumentHandler(mockService, mockUserRepo)

	docID := uuid.New()
	epicID := uuid.New()
	user := &models.User{ID: uuid.New(), Role: models.RoleUser}

	// Mock expectations
	mockService.On("UnlinkSteeringDocumentFromEpic", docID, epicID, user).Return(nil)

	// Execute
	err := handler.steeringDocumentService.UnlinkSteeringDocumentFromEpic(docID, epicID, user)

	// Assert
	assert.NoError(t, err)
	mockService.AssertExpectations(t)
}

func TestSteeringDocumentHandler_GetSteeringDocumentsByEpicID_Success(t *testing.T) {
	mockService := &MockSteeringDocumentService{}
	mockUserRepo := &MockUserRepository{}
	handler := NewSteeringDocumentHandler(mockService, mockUserRepo)

	epicID := uuid.New()
	user := &models.User{ID: uuid.New(), Role: models.RoleUser}
	expectedDocs := []models.SteeringDocument{
		{ID: uuid.New(), Title: "Doc 1", CreatorID: user.ID},
		{ID: uuid.New(), Title: "Doc 2", CreatorID: user.ID},
	}

	// Mock expectations
	mockService.On("GetSteeringDocumentsByEpicID", epicID, user).Return(expectedDocs, nil)

	// Execute
	result, err := handler.steeringDocumentService.GetSteeringDocumentsByEpicID(epicID, user)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedDocs, result)
	mockService.AssertExpectations(t)
}

// Helper function for steering document handler tests
func steeringHandlerStringPtr(s string) *string {
	return &s
}

package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"product-requirements-management/internal/models"
	"product-requirements-management/internal/service"
)

// MockConfigService for testing status model handlers
type MockConfigServiceForStatusModel struct {
	mock.Mock
}

func (m *MockConfigServiceForStatusModel) CreateStatusModel(req service.CreateStatusModelRequest) (*models.StatusModel, error) {
	args := m.Called(req)
	return args.Get(0).(*models.StatusModel), args.Error(1)
}

func (m *MockConfigServiceForStatusModel) GetStatusModelByID(id uuid.UUID) (*models.StatusModel, error) {
	args := m.Called(id)
	return args.Get(0).(*models.StatusModel), args.Error(1)
}

func (m *MockConfigServiceForStatusModel) GetStatusModelByEntityTypeAndName(entityType models.EntityType, name string) (*models.StatusModel, error) {
	args := m.Called(entityType, name)
	return args.Get(0).(*models.StatusModel), args.Error(1)
}

func (m *MockConfigServiceForStatusModel) GetDefaultStatusModelByEntityType(entityType models.EntityType) (*models.StatusModel, error) {
	args := m.Called(entityType)
	return args.Get(0).(*models.StatusModel), args.Error(1)
}

func (m *MockConfigServiceForStatusModel) UpdateStatusModel(id uuid.UUID, req service.UpdateStatusModelRequest) (*models.StatusModel, error) {
	args := m.Called(id, req)
	return args.Get(0).(*models.StatusModel), args.Error(1)
}

func (m *MockConfigServiceForStatusModel) DeleteStatusModel(id uuid.UUID, force bool) error {
	args := m.Called(id, force)
	return args.Error(0)
}

func (m *MockConfigServiceForStatusModel) ListStatusModels(filters service.StatusModelFilters) ([]models.StatusModel, error) {
	args := m.Called(filters)
	return args.Get(0).([]models.StatusModel), args.Error(1)
}

func (m *MockConfigServiceForStatusModel) ListStatusModelsByEntityType(entityType models.EntityType) ([]models.StatusModel, error) {
	args := m.Called(entityType)
	return args.Get(0).([]models.StatusModel), args.Error(1)
}

func (m *MockConfigServiceForStatusModel) CreateStatus(req service.CreateStatusRequest) (*models.Status, error) {
	args := m.Called(req)
	return args.Get(0).(*models.Status), args.Error(1)
}

func (m *MockConfigServiceForStatusModel) GetStatusByID(id uuid.UUID) (*models.Status, error) {
	args := m.Called(id)
	return args.Get(0).(*models.Status), args.Error(1)
}

func (m *MockConfigServiceForStatusModel) UpdateStatus(id uuid.UUID, req service.UpdateStatusRequest) (*models.Status, error) {
	args := m.Called(id, req)
	return args.Get(0).(*models.Status), args.Error(1)
}

func (m *MockConfigServiceForStatusModel) DeleteStatus(id uuid.UUID, force bool) error {
	args := m.Called(id, force)
	return args.Error(0)
}

func (m *MockConfigServiceForStatusModel) ListStatusesByModel(statusModelID uuid.UUID) ([]models.Status, error) {
	args := m.Called(statusModelID)
	return args.Get(0).([]models.Status), args.Error(1)
}

func (m *MockConfigServiceForStatusModel) CreateStatusTransition(req service.CreateStatusTransitionRequest) (*models.StatusTransition, error) {
	args := m.Called(req)
	return args.Get(0).(*models.StatusTransition), args.Error(1)
}

func (m *MockConfigServiceForStatusModel) GetStatusTransitionByID(id uuid.UUID) (*models.StatusTransition, error) {
	args := m.Called(id)
	return args.Get(0).(*models.StatusTransition), args.Error(1)
}

func (m *MockConfigServiceForStatusModel) UpdateStatusTransition(id uuid.UUID, req service.UpdateStatusTransitionRequest) (*models.StatusTransition, error) {
	args := m.Called(id, req)
	return args.Get(0).(*models.StatusTransition), args.Error(1)
}

func (m *MockConfigServiceForStatusModel) DeleteStatusTransition(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockConfigServiceForStatusModel) ListStatusTransitionsByModel(statusModelID uuid.UUID) ([]models.StatusTransition, error) {
	args := m.Called(statusModelID)
	return args.Get(0).([]models.StatusTransition), args.Error(1)
}

func (m *MockConfigServiceForStatusModel) ValidateStatusTransition(entityType models.EntityType, fromStatus, toStatus string) error {
	args := m.Called(entityType, fromStatus, toStatus)
	return args.Error(0)
}

// Add the missing methods from ConfigService interface
func (m *MockConfigServiceForStatusModel) CreateRequirementType(req service.CreateRequirementTypeRequest) (*models.RequirementType, error) {
	args := m.Called(req)
	return args.Get(0).(*models.RequirementType), args.Error(1)
}

func (m *MockConfigServiceForStatusModel) GetRequirementTypeByID(id uuid.UUID) (*models.RequirementType, error) {
	args := m.Called(id)
	return args.Get(0).(*models.RequirementType), args.Error(1)
}

func (m *MockConfigServiceForStatusModel) GetRequirementTypeByName(name string) (*models.RequirementType, error) {
	args := m.Called(name)
	return args.Get(0).(*models.RequirementType), args.Error(1)
}

func (m *MockConfigServiceForStatusModel) UpdateRequirementType(id uuid.UUID, req service.UpdateRequirementTypeRequest) (*models.RequirementType, error) {
	args := m.Called(id, req)
	return args.Get(0).(*models.RequirementType), args.Error(1)
}

func (m *MockConfigServiceForStatusModel) DeleteRequirementType(id uuid.UUID, force bool) error {
	args := m.Called(id, force)
	return args.Error(0)
}

func (m *MockConfigServiceForStatusModel) ListRequirementTypes(filters service.RequirementTypeFilters) ([]models.RequirementType, error) {
	args := m.Called(filters)
	return args.Get(0).([]models.RequirementType), args.Error(1)
}

func (m *MockConfigServiceForStatusModel) CreateRelationshipType(req service.CreateRelationshipTypeRequest) (*models.RelationshipType, error) {
	args := m.Called(req)
	return args.Get(0).(*models.RelationshipType), args.Error(1)
}

func (m *MockConfigServiceForStatusModel) GetRelationshipTypeByID(id uuid.UUID) (*models.RelationshipType, error) {
	args := m.Called(id)
	return args.Get(0).(*models.RelationshipType), args.Error(1)
}

func (m *MockConfigServiceForStatusModel) GetRelationshipTypeByName(name string) (*models.RelationshipType, error) {
	args := m.Called(name)
	return args.Get(0).(*models.RelationshipType), args.Error(1)
}

func (m *MockConfigServiceForStatusModel) UpdateRelationshipType(id uuid.UUID, req service.UpdateRelationshipTypeRequest) (*models.RelationshipType, error) {
	args := m.Called(id, req)
	return args.Get(0).(*models.RelationshipType), args.Error(1)
}

func (m *MockConfigServiceForStatusModel) DeleteRelationshipType(id uuid.UUID, force bool) error {
	args := m.Called(id, force)
	return args.Error(0)
}

func (m *MockConfigServiceForStatusModel) ListRelationshipTypes(filters service.RelationshipTypeFilters) ([]models.RelationshipType, error) {
	args := m.Called(filters)
	return args.Get(0).([]models.RelationshipType), args.Error(1)
}

func (m *MockConfigServiceForStatusModel) ValidateRequirementType(typeID uuid.UUID) error {
	args := m.Called(typeID)
	return args.Error(0)
}

func (m *MockConfigServiceForStatusModel) ValidateRelationshipType(typeID uuid.UUID) error {
	args := m.Called(typeID)
	return args.Error(0)
}

func TestCreateStatusModel(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("successful creation", func(t *testing.T) {
		mockService := new(MockConfigServiceForStatusModel)
		handler := NewConfigHandler(mockService)

		req := service.CreateStatusModelRequest{
			EntityType:  models.EntityTypeEpic,
			Name:        "Test Status Model",
			Description: stringPtr("Test description"),
			IsDefault:   true,
		}

		expectedStatusModel := &models.StatusModel{
			ID:          uuid.New(),
			EntityType:  models.EntityTypeEpic,
			Name:        "Test Status Model",
			Description: stringPtr("Test description"),
			IsDefault:   true,
		}

		mockService.On("CreateStatusModel", req).Return(expectedStatusModel, nil)

		body, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/api/v1/config/status-models", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.CreateStatusModel(c)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response models.StatusModel
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, expectedStatusModel.Name, response.Name)
		assert.Equal(t, expectedStatusModel.EntityType, response.EntityType)

		mockService.AssertExpectations(t)
	})

	t.Run("invalid entity type", func(t *testing.T) {
		mockService := new(MockConfigServiceForStatusModel)
		handler := NewConfigHandler(mockService)

		req := service.CreateStatusModelRequest{
			EntityType: models.EntityType("invalid"),
			Name:       "Test Status Model",
		}

		mockService.On("CreateStatusModel", req).Return((*models.StatusModel)(nil), service.ErrInvalidEntityType)

		body, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/api/v1/config/status-models", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.CreateStatusModel(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		mockService.AssertExpectations(t)
	})

	t.Run("name already exists", func(t *testing.T) {
		mockService := new(MockConfigServiceForStatusModel)
		handler := NewConfigHandler(mockService)

		req := service.CreateStatusModelRequest{
			EntityType: models.EntityTypeEpic,
			Name:       "Existing Name",
		}

		mockService.On("CreateStatusModel", req).Return((*models.StatusModel)(nil), service.ErrStatusModelNameExists)

		body, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/api/v1/config/status-models", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.CreateStatusModel(c)

		assert.Equal(t, http.StatusConflict, w.Code)

		mockService.AssertExpectations(t)
	})
}

func TestGetStatusModel(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("successful retrieval", func(t *testing.T) {
		mockService := new(MockConfigServiceForStatusModel)
		handler := NewConfigHandler(mockService)

		id := uuid.New()
		expectedStatusModel := &models.StatusModel{
			ID:         id,
			EntityType: models.EntityTypeEpic,
			Name:       "Test Status Model",
		}

		mockService.On("GetStatusModelByID", id).Return(expectedStatusModel, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/api/v1/config/status-models/"+id.String(), nil)
		c.Params = []gin.Param{{Key: "id", Value: id.String()}}

		handler.GetStatusModel(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response models.StatusModel
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, expectedStatusModel.ID, response.ID)
		assert.Equal(t, expectedStatusModel.Name, response.Name)

		mockService.AssertExpectations(t)
	})

	t.Run("status model not found", func(t *testing.T) {
		mockService := new(MockConfigServiceForStatusModel)
		handler := NewConfigHandler(mockService)

		id := uuid.New()

		mockService.On("GetStatusModelByID", id).Return((*models.StatusModel)(nil), service.ErrStatusModelNotFound)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/api/v1/config/status-models/"+id.String(), nil)
		c.Params = []gin.Param{{Key: "id", Value: id.String()}}

		handler.GetStatusModel(c)

		assert.Equal(t, http.StatusNotFound, w.Code)

		mockService.AssertExpectations(t)
	})

	t.Run("invalid ID format", func(t *testing.T) {
		mockService := new(MockConfigServiceForStatusModel)
		handler := NewConfigHandler(mockService)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/api/v1/config/status-models/invalid-id", nil)
		c.Params = []gin.Param{{Key: "id", Value: "invalid-id"}}

		handler.GetStatusModel(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestListStatusModels(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("successful listing", func(t *testing.T) {
		mockService := new(MockConfigServiceForStatusModel)
		handler := NewConfigHandler(mockService)

		expectedModels := []models.StatusModel{
			{
				ID:         uuid.New(),
				EntityType: models.EntityTypeEpic,
				Name:       "Epic Workflow",
			},
			{
				ID:         uuid.New(),
				EntityType: models.EntityTypeUserStory,
				Name:       "User Story Workflow",
			},
		}

		filters := service.StatusModelFilters{}
		mockService.On("ListStatusModels", filters).Return(expectedModels, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/api/v1/config/status-models", nil)

		handler.ListStatusModels(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, float64(2), response["count"])

		mockService.AssertExpectations(t)
	})

	t.Run("filtering by entity type", func(t *testing.T) {
		mockService := new(MockConfigServiceForStatusModel)
		handler := NewConfigHandler(mockService)

		expectedModels := []models.StatusModel{
			{
				ID:         uuid.New(),
				EntityType: models.EntityTypeEpic,
				Name:       "Epic Workflow",
			},
		}

		filters := service.StatusModelFilters{
			EntityType: models.EntityTypeEpic,
		}
		mockService.On("ListStatusModels", filters).Return(expectedModels, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/api/v1/config/status-models?entity_type=epic", nil)

		handler.ListStatusModels(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, float64(1), response["count"])

		mockService.AssertExpectations(t)
	})
}

func TestGetDefaultStatusModel(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("successful retrieval", func(t *testing.T) {
		mockService := new(MockConfigServiceForStatusModel)
		handler := NewConfigHandler(mockService)

		expectedModel := &models.StatusModel{
			ID:         uuid.New(),
			EntityType: models.EntityTypeEpic,
			Name:       "Default Epic Workflow",
			IsDefault:  true,
		}

		mockService.On("GetDefaultStatusModelByEntityType", models.EntityTypeEpic).Return(expectedModel, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/api/v1/config/status-models/default/epic", nil)
		c.Params = []gin.Param{{Key: "entity_type", Value: "epic"}}

		handler.GetDefaultStatusModel(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response models.StatusModel
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, expectedModel.ID, response.ID)
		assert.True(t, response.IsDefault)

		mockService.AssertExpectations(t)
	})

	t.Run("default model not found", func(t *testing.T) {
		mockService := new(MockConfigServiceForStatusModel)
		handler := NewConfigHandler(mockService)

		mockService.On("GetDefaultStatusModelByEntityType", models.EntityTypeEpic).Return((*models.StatusModel)(nil), service.ErrStatusModelNotFound)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/api/v1/config/status-models/default/epic", nil)
		c.Params = []gin.Param{{Key: "entity_type", Value: "epic"}}

		handler.GetDefaultStatusModel(c)

		assert.Equal(t, http.StatusNotFound, w.Code)

		mockService.AssertExpectations(t)
	})
}


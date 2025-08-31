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

// MockConfigService for testing
type MockConfigService struct {
	mock.Mock
}

func (m *MockConfigService) CreateRequirementType(req service.CreateRequirementTypeRequest) (*models.RequirementType, error) {
	args := m.Called(req)
	return args.Get(0).(*models.RequirementType), args.Error(1)
}

func (m *MockConfigService) GetRequirementTypeByID(id uuid.UUID) (*models.RequirementType, error) {
	args := m.Called(id)
	return args.Get(0).(*models.RequirementType), args.Error(1)
}

func (m *MockConfigService) GetRequirementTypeByName(name string) (*models.RequirementType, error) {
	args := m.Called(name)
	return args.Get(0).(*models.RequirementType), args.Error(1)
}

func (m *MockConfigService) UpdateRequirementType(id uuid.UUID, req service.UpdateRequirementTypeRequest) (*models.RequirementType, error) {
	args := m.Called(id, req)
	return args.Get(0).(*models.RequirementType), args.Error(1)
}

func (m *MockConfigService) DeleteRequirementType(id uuid.UUID, force bool) error {
	args := m.Called(id, force)
	return args.Error(0)
}

func (m *MockConfigService) ListRequirementTypes(filters service.RequirementTypeFilters) ([]models.RequirementType, error) {
	args := m.Called(filters)
	return args.Get(0).([]models.RequirementType), args.Error(1)
}

func (m *MockConfigService) CreateRelationshipType(req service.CreateRelationshipTypeRequest) (*models.RelationshipType, error) {
	args := m.Called(req)
	return args.Get(0).(*models.RelationshipType), args.Error(1)
}

func (m *MockConfigService) GetRelationshipTypeByID(id uuid.UUID) (*models.RelationshipType, error) {
	args := m.Called(id)
	return args.Get(0).(*models.RelationshipType), args.Error(1)
}

func (m *MockConfigService) GetRelationshipTypeByName(name string) (*models.RelationshipType, error) {
	args := m.Called(name)
	return args.Get(0).(*models.RelationshipType), args.Error(1)
}

func (m *MockConfigService) UpdateRelationshipType(id uuid.UUID, req service.UpdateRelationshipTypeRequest) (*models.RelationshipType, error) {
	args := m.Called(id, req)
	return args.Get(0).(*models.RelationshipType), args.Error(1)
}

func (m *MockConfigService) DeleteRelationshipType(id uuid.UUID, force bool) error {
	args := m.Called(id, force)
	return args.Error(0)
}

func (m *MockConfigService) ListRelationshipTypes(filters service.RelationshipTypeFilters) ([]models.RelationshipType, error) {
	args := m.Called(filters)
	return args.Get(0).([]models.RelationshipType), args.Error(1)
}

func (m *MockConfigService) ValidateRequirementType(typeID uuid.UUID) error {
	args := m.Called(typeID)
	return args.Error(0)
}

func (m *MockConfigService) ValidateRelationshipType(typeID uuid.UUID) error {
	args := m.Called(typeID)
	return args.Error(0)
}

// Status Model methods
func (m *MockConfigService) CreateStatusModel(req service.CreateStatusModelRequest) (*models.StatusModel, error) {
	args := m.Called(req)
	return args.Get(0).(*models.StatusModel), args.Error(1)
}

func (m *MockConfigService) GetStatusModelByID(id uuid.UUID) (*models.StatusModel, error) {
	args := m.Called(id)
	return args.Get(0).(*models.StatusModel), args.Error(1)
}

func (m *MockConfigService) GetStatusModelByEntityTypeAndName(entityType models.EntityType, name string) (*models.StatusModel, error) {
	args := m.Called(entityType, name)
	return args.Get(0).(*models.StatusModel), args.Error(1)
}

func (m *MockConfigService) GetDefaultStatusModelByEntityType(entityType models.EntityType) (*models.StatusModel, error) {
	args := m.Called(entityType)
	return args.Get(0).(*models.StatusModel), args.Error(1)
}

func (m *MockConfigService) UpdateStatusModel(id uuid.UUID, req service.UpdateStatusModelRequest) (*models.StatusModel, error) {
	args := m.Called(id, req)
	return args.Get(0).(*models.StatusModel), args.Error(1)
}

func (m *MockConfigService) DeleteStatusModel(id uuid.UUID, force bool) error {
	args := m.Called(id, force)
	return args.Error(0)
}

func (m *MockConfigService) ListStatusModels(filters service.StatusModelFilters) ([]models.StatusModel, error) {
	args := m.Called(filters)
	return args.Get(0).([]models.StatusModel), args.Error(1)
}

func (m *MockConfigService) ListStatusModelsByEntityType(entityType models.EntityType) ([]models.StatusModel, error) {
	args := m.Called(entityType)
	return args.Get(0).([]models.StatusModel), args.Error(1)
}

// Status methods
func (m *MockConfigService) CreateStatus(req service.CreateStatusRequest) (*models.Status, error) {
	args := m.Called(req)
	return args.Get(0).(*models.Status), args.Error(1)
}

func (m *MockConfigService) GetStatusByID(id uuid.UUID) (*models.Status, error) {
	args := m.Called(id)
	return args.Get(0).(*models.Status), args.Error(1)
}

func (m *MockConfigService) UpdateStatus(id uuid.UUID, req service.UpdateStatusRequest) (*models.Status, error) {
	args := m.Called(id, req)
	return args.Get(0).(*models.Status), args.Error(1)
}

func (m *MockConfigService) DeleteStatus(id uuid.UUID, force bool) error {
	args := m.Called(id, force)
	return args.Error(0)
}

func (m *MockConfigService) ListStatusesByModel(statusModelID uuid.UUID) ([]models.Status, error) {
	args := m.Called(statusModelID)
	return args.Get(0).([]models.Status), args.Error(1)
}

// Status Transition methods
func (m *MockConfigService) CreateStatusTransition(req service.CreateStatusTransitionRequest) (*models.StatusTransition, error) {
	args := m.Called(req)
	return args.Get(0).(*models.StatusTransition), args.Error(1)
}

func (m *MockConfigService) GetStatusTransitionByID(id uuid.UUID) (*models.StatusTransition, error) {
	args := m.Called(id)
	return args.Get(0).(*models.StatusTransition), args.Error(1)
}

func (m *MockConfigService) UpdateStatusTransition(id uuid.UUID, req service.UpdateStatusTransitionRequest) (*models.StatusTransition, error) {
	args := m.Called(id, req)
	return args.Get(0).(*models.StatusTransition), args.Error(1)
}

func (m *MockConfigService) DeleteStatusTransition(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockConfigService) ListStatusTransitionsByModel(statusModelID uuid.UUID) ([]models.StatusTransition, error) {
	args := m.Called(statusModelID)
	return args.Get(0).([]models.StatusTransition), args.Error(1)
}

func (m *MockConfigService) ValidateStatusTransition(entityType models.EntityType, fromStatus, toStatus string) error {
	args := m.Called(entityType, fromStatus, toStatus)
	return args.Error(0)
}

func setupConfigTestRouter() (*gin.Engine, *MockConfigService) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	mockService := &MockConfigService{}
	handler := NewConfigHandler(mockService)
	
	v1 := router.Group("/api/v1")
	config := v1.Group("/config")
	{
		// Requirement Type routes
		requirementTypes := config.Group("/requirement-types")
		{
			requirementTypes.POST("", handler.CreateRequirementType)
			requirementTypes.GET("", handler.ListRequirementTypes)
			requirementTypes.GET("/:id", handler.GetRequirementType)
			requirementTypes.PUT("/:id", handler.UpdateRequirementType)
			requirementTypes.DELETE("/:id", handler.DeleteRequirementType)
		}

		// Relationship Type routes
		relationshipTypes := config.Group("/relationship-types")
		{
			relationshipTypes.POST("", handler.CreateRelationshipType)
			relationshipTypes.GET("", handler.ListRelationshipTypes)
			relationshipTypes.GET("/:id", handler.GetRelationshipType)
			relationshipTypes.PUT("/:id", handler.UpdateRelationshipType)
			relationshipTypes.DELETE("/:id", handler.DeleteRelationshipType)
		}
	}
	
	return router, mockService
}

func TestConfigHandler_CreateRequirementType(t *testing.T) {
	router, mockService := setupConfigTestRouter()

	t.Run("successful creation", func(t *testing.T) {
		req := service.CreateRequirementTypeRequest{
			Name:        "Functional",
			Description: stringPtrHandler("Functional requirements"),
		}

		expectedType := &models.RequirementType{
			ID:          uuid.New(),
			Name:        "Functional",
			Description: stringPtrHandler("Functional requirements"),
		}

		mockService.On("CreateRequirementType", req).Return(expectedType, nil)

		reqBody, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		httpReq, _ := http.NewRequest("POST", "/api/v1/config/requirement-types", bytes.NewBuffer(reqBody))
		httpReq.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusCreated, w.Code)
		
		var response models.RequirementType
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, expectedType.Name, response.Name)
		assert.Equal(t, *expectedType.Description, *response.Description)

		mockService.AssertExpectations(t)
	})

	t.Run("name already exists", func(t *testing.T) {
		req := service.CreateRequirementTypeRequest{
			Name: "Existing Type",
		}

		mockService.On("CreateRequirementType", req).Return((*models.RequirementType)(nil), service.ErrRequirementTypeNameExists)

		reqBody, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		httpReq, _ := http.NewRequest("POST", "/api/v1/config/requirement-types", bytes.NewBuffer(reqBody))
		httpReq.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusConflict, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Requirement type name already exists", response["error"])

		mockService.AssertExpectations(t)
	})

	t.Run("invalid request body", func(t *testing.T) {
		w := httptest.NewRecorder()
		httpReq, _ := http.NewRequest("POST", "/api/v1/config/requirement-types", bytes.NewBuffer([]byte("invalid json")))
		httpReq.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Invalid request body", response["error"])
	})
}

func TestConfigHandler_GetRequirementType(t *testing.T) {
	router, mockService := setupConfigTestRouter()

	t.Run("successful retrieval", func(t *testing.T) {
		id := uuid.New()
		expectedType := &models.RequirementType{
			ID:          id,
			Name:        "Functional",
			Description: stringPtrHandler("Functional requirements"),
		}

		mockService.On("GetRequirementTypeByID", id).Return(expectedType, nil)

		w := httptest.NewRecorder()
		httpReq, _ := http.NewRequest("GET", "/api/v1/config/requirement-types/"+id.String(), nil)

		router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var response models.RequirementType
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, expectedType.ID, response.ID)
		assert.Equal(t, expectedType.Name, response.Name)

		mockService.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		id := uuid.New()

		mockService.On("GetRequirementTypeByID", id).Return((*models.RequirementType)(nil), service.ErrRequirementTypeNotFound)

		w := httptest.NewRecorder()
		httpReq, _ := http.NewRequest("GET", "/api/v1/config/requirement-types/"+id.String(), nil)

		router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusNotFound, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Requirement type not found", response["error"])

		mockService.AssertExpectations(t)
	})

	t.Run("invalid ID format", func(t *testing.T) {
		w := httptest.NewRecorder()
		httpReq, _ := http.NewRequest("GET", "/api/v1/config/requirement-types/invalid-id", nil)

		router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Invalid requirement type ID format", response["error"])
	})
}

func TestConfigHandler_DeleteRequirementType(t *testing.T) {
	router, mockService := setupConfigTestRouter()

	t.Run("successful deletion", func(t *testing.T) {
		id := uuid.New()

		mockService.On("DeleteRequirementType", id, false).Return(nil)

		w := httptest.NewRecorder()
		httpReq, _ := http.NewRequest("DELETE", "/api/v1/config/requirement-types/"+id.String(), nil)

		router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusNoContent, w.Code)

		mockService.AssertExpectations(t)
	})

	t.Run("has associated requirements", func(t *testing.T) {
		id := uuid.New()

		mockService.On("DeleteRequirementType", id, false).Return(service.ErrRequirementTypeHasRequirements)

		w := httptest.NewRecorder()
		httpReq, _ := http.NewRequest("DELETE", "/api/v1/config/requirement-types/"+id.String(), nil)

		router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusConflict, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Requirement type has associated requirements and cannot be deleted", response["error"])

		mockService.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		id := uuid.New()

		mockService.On("DeleteRequirementType", id, false).Return(service.ErrRequirementTypeNotFound)

		w := httptest.NewRecorder()
		httpReq, _ := http.NewRequest("DELETE", "/api/v1/config/requirement-types/"+id.String(), nil)

		router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusNotFound, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Requirement type not found", response["error"])

		mockService.AssertExpectations(t)
	})
}

func TestConfigHandler_CreateRelationshipType(t *testing.T) {
	router, mockService := setupConfigTestRouter()

	t.Run("successful creation", func(t *testing.T) {
		req := service.CreateRelationshipTypeRequest{
			Name:        "depends_on",
			Description: stringPtrHandler("Dependency relationship"),
		}

		expectedType := &models.RelationshipType{
			ID:          uuid.New(),
			Name:        "depends_on",
			Description: stringPtrHandler("Dependency relationship"),
		}

		mockService.On("CreateRelationshipType", req).Return(expectedType, nil)

		reqBody, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		httpReq, _ := http.NewRequest("POST", "/api/v1/config/relationship-types", bytes.NewBuffer(reqBody))
		httpReq.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusCreated, w.Code)
		
		var response models.RelationshipType
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, expectedType.Name, response.Name)
		assert.Equal(t, *expectedType.Description, *response.Description)

		mockService.AssertExpectations(t)
	})

	t.Run("name already exists", func(t *testing.T) {
		req := service.CreateRelationshipTypeRequest{
			Name: "existing_type",
		}

		mockService.On("CreateRelationshipType", req).Return((*models.RelationshipType)(nil), service.ErrRelationshipTypeNameExists)

		reqBody, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		httpReq, _ := http.NewRequest("POST", "/api/v1/config/relationship-types", bytes.NewBuffer(reqBody))
		httpReq.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusConflict, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Relationship type name already exists", response["error"])

		mockService.AssertExpectations(t)
	})
}

func TestConfigHandler_ListRequirementTypes(t *testing.T) {
	t.Run("successful listing", func(t *testing.T) {
		router, mockService := setupConfigTestRouter()
		
		expectedTypes := []models.RequirementType{
			{
				ID:          uuid.New(),
				Name:        "Functional",
				Description: stringPtrHandler("Functional requirements"),
			},
			{
				ID:          uuid.New(),
				Name:        "Non-Functional",
				Description: stringPtrHandler("Non-functional requirements"),
			},
		}

		mockService.On("ListRequirementTypes", mock.AnythingOfType("service.RequirementTypeFilters")).Return(expectedTypes, nil)

		w := httptest.NewRecorder()
		httpReq, _ := http.NewRequest("GET", "/api/v1/config/requirement-types", nil)

		router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, float64(2), response["count"])
		
		types := response["requirement_types"].([]interface{})
		assert.Len(t, types, 2)

		mockService.AssertExpectations(t)
	})

	t.Run("with filters", func(t *testing.T) {
		router, mockService := setupConfigTestRouter()
		
		expectedTypes := []models.RequirementType{
			{
				ID:   uuid.New(),
				Name: "Functional",
			},
		}

		mockService.On("ListRequirementTypes", mock.AnythingOfType("service.RequirementTypeFilters")).Return(expectedTypes, nil)

		w := httptest.NewRecorder()
		httpReq, _ := http.NewRequest("GET", "/api/v1/config/requirement-types?limit=10&offset=0&order_by=name", nil)

		router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, float64(len(expectedTypes)), response["count"])

		mockService.AssertExpectations(t)
	})
}

// Helper function for string pointers
func stringPtrHandler(s string) *string {
	return &s
}
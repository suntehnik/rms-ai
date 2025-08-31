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

// MockRequirementService is a mock implementation of RequirementService
type MockRequirementService struct {
	mock.Mock
}

func (m *MockRequirementService) CreateRequirement(req service.CreateRequirementRequest) (*models.Requirement, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Requirement), args.Error(1)
}

func (m *MockRequirementService) GetRequirementByID(id uuid.UUID) (*models.Requirement, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Requirement), args.Error(1)
}

func (m *MockRequirementService) GetRequirementByReferenceID(referenceID string) (*models.Requirement, error) {
	args := m.Called(referenceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Requirement), args.Error(1)
}

func (m *MockRequirementService) UpdateRequirement(id uuid.UUID, req service.UpdateRequirementRequest) (*models.Requirement, error) {
	args := m.Called(id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Requirement), args.Error(1)
}

func (m *MockRequirementService) DeleteRequirement(id uuid.UUID, force bool) error {
	args := m.Called(id, force)
	return args.Error(0)
}

func (m *MockRequirementService) ListRequirements(filters service.RequirementFilters) ([]models.Requirement, error) {
	args := m.Called(filters)
	return args.Get(0).([]models.Requirement), args.Error(1)
}

func (m *MockRequirementService) GetRequirementWithRelationships(id uuid.UUID) (*models.Requirement, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Requirement), args.Error(1)
}

func (m *MockRequirementService) GetRequirementsByUserStory(userStoryID uuid.UUID) ([]models.Requirement, error) {
	args := m.Called(userStoryID)
	return args.Get(0).([]models.Requirement), args.Error(1)
}

func (m *MockRequirementService) ChangeRequirementStatus(id uuid.UUID, newStatus models.RequirementStatus) (*models.Requirement, error) {
	args := m.Called(id, newStatus)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Requirement), args.Error(1)
}

func (m *MockRequirementService) AssignRequirement(id uuid.UUID, assigneeID uuid.UUID) (*models.Requirement, error) {
	args := m.Called(id, assigneeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Requirement), args.Error(1)
}

func (m *MockRequirementService) CreateRelationship(req service.CreateRelationshipRequest) (*models.RequirementRelationship, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.RequirementRelationship), args.Error(1)
}

func (m *MockRequirementService) DeleteRelationship(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockRequirementService) GetRelationshipsByRequirement(requirementID uuid.UUID) ([]models.RequirementRelationship, error) {
	args := m.Called(requirementID)
	return args.Get(0).([]models.RequirementRelationship), args.Error(1)
}

func (m *MockRequirementService) SearchRequirements(searchText string) ([]models.Requirement, error) {
	args := m.Called(searchText)
	return args.Get(0).([]models.Requirement), args.Error(1)
}

func TestRequirementHandler_CreateRequirement(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("successful creation", func(t *testing.T) {
		mockService := new(MockRequirementService)
		handler := NewRequirementHandler(mockService)

		userStoryID := uuid.New()
		creatorID := uuid.New()
		typeID := uuid.New()

		reqBody := service.CreateRequirementRequest{
			UserStoryID: userStoryID,
			CreatorID:   creatorID,
			Priority:    models.PriorityHigh,
			TypeID:      typeID,
			Title:       "Test Requirement",
		}

		expectedRequirement := &models.Requirement{
			ID:          uuid.New(),
			UserStoryID: userStoryID,
			CreatorID:   creatorID,
			AssigneeID:  creatorID,
			Priority:    models.PriorityHigh,
			Status:      models.RequirementStatusDraft,
			TypeID:      typeID,
			Title:       "Test Requirement",
		}

		mockService.On("CreateRequirement", reqBody).Return(expectedRequirement, nil)

		// Create request
		jsonBody, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/api/v1/requirements", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		// Create response recorder
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		// Call handler
		handler.CreateRequirement(c)

		// Assertions
		assert.Equal(t, http.StatusCreated, w.Code)

		var response models.Requirement
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, expectedRequirement.ID, response.ID)
		assert.Equal(t, expectedRequirement.Title, response.Title)

		mockService.AssertExpectations(t)
	})

	t.Run("invalid request body", func(t *testing.T) {
		mockService := new(MockRequirementService)
		handler := NewRequirementHandler(mockService)

		// Create invalid request (missing required fields)
		req, _ := http.NewRequest("POST", "/api/v1/requirements", bytes.NewBuffer([]byte(`{"invalid": "json"}`)))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		handler.CreateRequirement(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response["error"], "Invalid request body")
	})

	t.Run("service error - user story not found", func(t *testing.T) {
		mockService := new(MockRequirementService)
		handler := NewRequirementHandler(mockService)

		reqBody := service.CreateRequirementRequest{
			UserStoryID: uuid.New(),
			CreatorID:   uuid.New(),
			Priority:    models.PriorityHigh,
			TypeID:      uuid.New(),
			Title:       "Test Requirement",
		}

		mockService.On("CreateRequirement", reqBody).Return(nil, service.ErrUserStoryNotFound)

		jsonBody, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/api/v1/requirements", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		handler.CreateRequirement(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "User story not found", response["error"])

		mockService.AssertExpectations(t)
	})
}

func TestRequirementHandler_GetRequirement(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("successful retrieval by UUID", func(t *testing.T) {
		mockService := new(MockRequirementService)
		handler := NewRequirementHandler(mockService)

		requirementID := uuid.New()
		expectedRequirement := &models.Requirement{
			ID:    requirementID,
			Title: "Test Requirement",
		}

		mockService.On("GetRequirementByID", requirementID).Return(expectedRequirement, nil)

		req, _ := http.NewRequest("GET", "/api/v1/requirements/"+requirementID.String(), nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = []gin.Param{{Key: "id", Value: requirementID.String()}}

		handler.GetRequirement(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response models.Requirement
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, expectedRequirement.ID, response.ID)
		assert.Equal(t, expectedRequirement.Title, response.Title)

		mockService.AssertExpectations(t)
	})

	t.Run("successful retrieval by reference ID", func(t *testing.T) {
		mockService := new(MockRequirementService)
		handler := NewRequirementHandler(mockService)

		referenceID := "REQ-001"
		expectedRequirement := &models.Requirement{
			ID:          uuid.New(),
			ReferenceID: referenceID,
			Title:       "Test Requirement",
		}

		mockService.On("GetRequirementByReferenceID", referenceID).Return(expectedRequirement, nil)

		req, _ := http.NewRequest("GET", "/api/v1/requirements/"+referenceID, nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = []gin.Param{{Key: "id", Value: referenceID}}

		handler.GetRequirement(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response models.Requirement
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, expectedRequirement.ID, response.ID)
		assert.Equal(t, expectedRequirement.ReferenceID, response.ReferenceID)

		mockService.AssertExpectations(t)
	})

	t.Run("requirement not found", func(t *testing.T) {
		mockService := new(MockRequirementService)
		handler := NewRequirementHandler(mockService)

		requirementID := uuid.New()

		mockService.On("GetRequirementByID", requirementID).Return(nil, service.ErrRequirementNotFound)

		req, _ := http.NewRequest("GET", "/api/v1/requirements/"+requirementID.String(), nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = []gin.Param{{Key: "id", Value: requirementID.String()}}

		handler.GetRequirement(c)

		assert.Equal(t, http.StatusNotFound, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Requirement not found", response["error"])

		mockService.AssertExpectations(t)
	})
}

func TestRequirementHandler_DeleteRequirement(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("successful deletion", func(t *testing.T) {
		mockService := new(MockRequirementService)
		handler := NewRequirementHandler(mockService)

		requirementID := uuid.New()

		mockService.On("DeleteRequirement", requirementID, false).Return(nil)

		req, _ := http.NewRequest("DELETE", "/api/v1/requirements/"+requirementID.String(), nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = []gin.Param{{Key: "id", Value: requirementID.String()}}

		handler.DeleteRequirement(c)

		assert.Equal(t, http.StatusNoContent, w.Code)

		mockService.AssertExpectations(t)
	})

	t.Run("force deletion", func(t *testing.T) {
		mockService := new(MockRequirementService)
		handler := NewRequirementHandler(mockService)

		requirementID := uuid.New()

		mockService.On("DeleteRequirement", requirementID, true).Return(nil)

		req, _ := http.NewRequest("DELETE", "/api/v1/requirements/"+requirementID.String()+"?force=true", nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = []gin.Param{{Key: "id", Value: requirementID.String()}}

		handler.DeleteRequirement(c)

		assert.Equal(t, http.StatusNoContent, w.Code)

		mockService.AssertExpectations(t)
	})

	t.Run("deletion blocked by relationships", func(t *testing.T) {
		mockService := new(MockRequirementService)
		handler := NewRequirementHandler(mockService)

		requirementID := uuid.New()

		mockService.On("DeleteRequirement", requirementID, false).Return(service.ErrRequirementHasRelationships)

		req, _ := http.NewRequest("DELETE", "/api/v1/requirements/"+requirementID.String(), nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = []gin.Param{{Key: "id", Value: requirementID.String()}}

		handler.DeleteRequirement(c)

		assert.Equal(t, http.StatusConflict, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response["error"], "relationships")
		assert.Contains(t, response["hint"], "force=true")

		mockService.AssertExpectations(t)
	})

	t.Run("invalid UUID format", func(t *testing.T) {
		mockService := new(MockRequirementService)
		handler := NewRequirementHandler(mockService)

		req, _ := http.NewRequest("DELETE", "/api/v1/requirements/invalid-uuid", nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = []gin.Param{{Key: "id", Value: "invalid-uuid"}}

		handler.DeleteRequirement(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Invalid requirement ID format", response["error"])
	})
}

func TestRequirementHandler_CreateRelationship(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("successful creation", func(t *testing.T) {
		mockService := new(MockRequirementService)
		handler := NewRequirementHandler(mockService)

		sourceID := uuid.New()
		targetID := uuid.New()
		relationshipTypeID := uuid.New()
		creatorID := uuid.New()

		reqBody := service.CreateRelationshipRequest{
			SourceRequirementID: sourceID,
			TargetRequirementID: targetID,
			RelationshipTypeID:  relationshipTypeID,
			CreatedBy:           creatorID,
		}

		expectedRelationship := &models.RequirementRelationship{
			ID:                  uuid.New(),
			SourceRequirementID: sourceID,
			TargetRequirementID: targetID,
			RelationshipTypeID:  relationshipTypeID,
			CreatedBy:           creatorID,
		}

		mockService.On("CreateRelationship", reqBody).Return(expectedRelationship, nil)

		jsonBody, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/api/v1/requirements/relationships", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		handler.CreateRelationship(c)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response models.RequirementRelationship
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, expectedRelationship.ID, response.ID)
		assert.Equal(t, expectedRelationship.SourceRequirementID, response.SourceRequirementID)

		mockService.AssertExpectations(t)
	})

	t.Run("circular relationship error", func(t *testing.T) {
		mockService := new(MockRequirementService)
		handler := NewRequirementHandler(mockService)

		reqBody := service.CreateRelationshipRequest{
			SourceRequirementID: uuid.New(),
			TargetRequirementID: uuid.New(),
			RelationshipTypeID:  uuid.New(),
			CreatedBy:           uuid.New(),
		}

		mockService.On("CreateRelationship", reqBody).Return(nil, service.ErrCircularRelationship)

		jsonBody, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/api/v1/requirements/relationships", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		handler.CreateRelationship(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response["error"], "same requirement")

		mockService.AssertExpectations(t)
	})
}

func TestRequirementHandler_SearchRequirements(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("successful search", func(t *testing.T) {
		mockService := new(MockRequirementService)
		handler := NewRequirementHandler(mockService)

		searchText := "test requirement"
		expectedRequirements := []models.Requirement{
			{
				ID:    uuid.New(),
				Title: "Test Requirement 1",
			},
			{
				ID:    uuid.New(),
				Title: "Test Requirement 2",
			},
		}

		mockService.On("SearchRequirements", searchText).Return(expectedRequirements, nil)

		req, _ := http.NewRequest("GET", "/api/v1/requirements/search?q="+searchText, nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		handler.SearchRequirements(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, float64(2), response["count"])
		assert.Equal(t, searchText, response["query"])

		requirements := response["requirements"].([]interface{})
		assert.Len(t, requirements, 2)

		mockService.AssertExpectations(t)
	})

	t.Run("missing search query", func(t *testing.T) {
		mockService := new(MockRequirementService)
		handler := NewRequirementHandler(mockService)

		req, _ := http.NewRequest("GET", "/api/v1/requirements/search", nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		handler.SearchRequirements(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response["error"], "Search query parameter 'q' is required")
	})
}
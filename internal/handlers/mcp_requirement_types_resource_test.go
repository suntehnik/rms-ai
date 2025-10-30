package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"

	"product-requirements-management/internal/models"
	"product-requirements-management/internal/service"
)

// MockRequirementTypeRepository for testing
type MockRequirementTypeRepository struct {
	mock.Mock
}

func (m *MockRequirementTypeRepository) Create(entity *models.RequirementType) error {
	args := m.Called(entity)
	return args.Error(0)
}

func (m *MockRequirementTypeRepository) GetByID(id uuid.UUID) (*models.RequirementType, error) {
	args := m.Called(id)
	return args.Get(0).(*models.RequirementType), args.Error(1)
}

func (m *MockRequirementTypeRepository) Update(entity *models.RequirementType) error {
	args := m.Called(entity)
	return args.Error(0)
}

func (m *MockRequirementTypeRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockRequirementTypeRepository) List(filter map[string]interface{}, orderBy string, limit, offset int) ([]models.RequirementType, error) {
	args := m.Called(filter, orderBy, limit, offset)
	return args.Get(0).([]models.RequirementType), args.Error(1)
}

func (m *MockRequirementTypeRepository) Count(filter map[string]interface{}) (int64, error) {
	args := m.Called(filter)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockRequirementTypeRepository) GetByName(name string) (*models.RequirementType, error) {
	args := m.Called(name)
	return args.Get(0).(*models.RequirementType), args.Error(1)
}

func (m *MockRequirementTypeRepository) ExistsByName(name string) (bool, error) {
	args := m.Called(name)
	return args.Get(0).(bool), args.Error(1)
}

func (m *MockRequirementTypeRepository) GetByReferenceID(referenceID string) (*models.RequirementType, error) {
	args := m.Called(referenceID)
	return args.Get(0).(*models.RequirementType), args.Error(1)
}

func (m *MockRequirementTypeRepository) GetByReferenceIDCaseInsensitive(referenceID string) (*models.RequirementType, error) {
	args := m.Called(referenceID)
	return args.Get(0).(*models.RequirementType), args.Error(1)
}

func (m *MockRequirementTypeRepository) Exists(id uuid.UUID) (bool, error) {
	args := m.Called(id)
	return args.Get(0).(bool), args.Error(1)
}

func (m *MockRequirementTypeRepository) ExistsByReferenceID(referenceID string) (bool, error) {
	args := m.Called(referenceID)
	return args.Get(0).(bool), args.Error(1)
}

func (m *MockRequirementTypeRepository) WithTransaction(fn func(*gorm.DB) error) error {
	args := m.Called(fn)
	return args.Error(0)
}

func (m *MockRequirementTypeRepository) GetDB() *gorm.DB {
	args := m.Called()
	return args.Get(0).(*gorm.DB)
}

func TestMCPHandler_RequirementTypesResource(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create mock requirement type repository
	mockRequirementTypeRepo := new(MockRequirementTypeRepository)

	// Mock requirement types data
	requirementTypes := []models.RequirementType{
		{
			ID:          uuid.New(),
			Name:        "Functional",
			Description: func() *string { s := "Functional requirements that describe what the system should do"; return &s }(),
		},
		{
			ID:   uuid.New(),
			Name: "Non-Functional",
			Description: func() *string {
				s := "Non-functional requirements that describe how the system should behave"
				return &s
			}(),
		},
	}

	// Setup mock expectations
	mockRequirementTypeRepo.On("List", mock.Anything, "name ASC", 1000, 0).Return(requirementTypes, nil)

	// Create resource service with requirement type provider
	logger := &logrus.Logger{}
	registry := service.NewResourceRegistry(logger)
	requirementTypeProvider := service.NewRequirementTypeResourceProvider(mockRequirementTypeRepo, logger)
	registry.RegisterProvider(requirementTypeProvider)
	resourceService := service.NewResourceService(registry, logger)

	// Create MCP handler
	handler := NewMCPHandler(nil, nil, nil, nil, nil, nil, nil, resourceService, mockRequirementTypeRepo)

	// Test resources/list to verify requirement types resource is included
	t.Run("resources_list_includes_requirement_types", func(t *testing.T) {
		requestBody := `{
			"jsonrpc": "2.0",
			"id": 1,
			"method": "resources/list"
		}`

		req := httptest.NewRequest("POST", "/api/v1/mcp", strings.NewReader(requestBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer test-token")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		// Add mock user to context (required for authentication)
		c.Set("user", &models.User{
			ID:       uuid.New(),
			Username: "testuser",
			Email:    "test@example.com",
			Role:     models.RoleUser,
		})

		handler.Process(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		// Check that the response contains resources
		result, ok := response["result"].(map[string]any)
		assert.True(t, ok)

		resources, ok := result["resources"].([]any)
		assert.True(t, ok)
		assert.NotEmpty(t, resources)

		// Check that requirements://requirements-types resource is included
		found := false
		for _, resource := range resources {
			resourceMap := resource.(map[string]any)
			if resourceMap["uri"] == "requirements://requirements-types" {
				found = true
				assert.Equal(t, "Requirement Types", resourceMap["name"])
				assert.Equal(t, "List of all supported requirement types in the system", resourceMap["description"])
				assert.Equal(t, "application/json", resourceMap["mimeType"])
				break
			}
		}
		assert.True(t, found, "requirements://requirements-types resource should be included in resources/list")
	})

	// Test resources/read for requirement types resource
	t.Run("resources_read_requirement_types", func(t *testing.T) {
		requestBody := `{
			"jsonrpc": "2.0",
			"id": 2,
			"method": "resources/read",
			"params": {
				"uri": "requirements://requirements-types"
			}
		}`

		req := httptest.NewRequest("POST", "/api/v1/mcp", strings.NewReader(requestBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer test-token")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		// Add mock user to context (required for authentication)
		c.Set("user", &models.User{
			ID:       uuid.New(),
			Username: "testuser",
			Email:    "test@example.com",
			Role:     models.RoleUser,
		})

		handler.Process(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		// Check that the response contains the requirement types data
		result, ok := response["result"].(map[string]any)
		assert.True(t, ok)

		contents, ok := result["contents"].([]any)
		assert.True(t, ok)
		assert.Len(t, contents, 1)

		content := contents[0].(map[string]any)
		assert.Equal(t, "requirements://requirements-types", content["uri"])
		assert.Equal(t, "application/json", content["mimeType"])

		// Parse the text content
		var textContent map[string]any
		err = json.Unmarshal([]byte(content["text"].(string)), &textContent)
		assert.NoError(t, err)

		// Verify the structure matches REQ-039 format
		requirementTypesData, ok := textContent["requirement_types"].([]any)
		assert.True(t, ok)
		assert.Len(t, requirementTypesData, 2)

		// Check first requirement type
		firstType := requirementTypesData[0].(map[string]any)
		assert.NotEmpty(t, firstType["id"])
		assert.Equal(t, "Functional", firstType["name"])
		assert.Equal(t, "Functional requirements that describe what the system should do", firstType["description"])

		// Check second requirement type
		secondType := requirementTypesData[1].(map[string]any)
		assert.NotEmpty(t, secondType["id"])
		assert.Equal(t, "Non-Functional", secondType["name"])
		assert.Equal(t, "Non-functional requirements that describe how the system should behave", secondType["description"])

		// Check count
		count, ok := textContent["count"].(float64)
		assert.True(t, ok)
		assert.Equal(t, float64(2), count)
	})

	mockRequirementTypeRepo.AssertExpectations(t)
}

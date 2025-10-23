package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"product-requirements-management/internal/handlers"
	"product-requirements-management/internal/models"
	"product-requirements-management/internal/repository"
	"product-requirements-management/internal/service"
)

func setupConfigIntegrationTest(t *testing.T) (*gin.Engine, *gorm.DB, service.ConfigService) {
	// Setup test database with SQL migrations
	testDatabase := SetupTestDatabase(t)
	db := testDatabase.DB

	// Initialize repositories
	requirementTypeRepo := repository.NewRequirementTypeRepository(db)
	relationshipTypeRepo := repository.NewRelationshipTypeRepository(db)
	requirementRepo := repository.NewRequirementRepository(db)
	requirementRelationshipRepo := repository.NewRequirementRelationshipRepository(db)

	// Initialize service
	configService := service.NewConfigService(
		requirementTypeRepo,
		relationshipTypeRepo,
		requirementRepo,
		requirementRelationshipRepo,
		nil, // statusModelRepo - not needed for this test
		nil, // statusRepo - not needed for this test
		nil, // statusTransitionRepo - not needed for this test
	)

	// Initialize handler
	configHandler := handlers.NewConfigHandler(configService)

	// Setup router
	gin.SetMode(gin.TestMode)
	router := gin.New()

	v1 := router.Group("/api/v1")
	config := v1.Group("/config")
	{
		// Requirement Type routes
		requirementTypes := config.Group("/requirement-types")
		{
			requirementTypes.POST("", configHandler.CreateRequirementType)
			requirementTypes.GET("", configHandler.ListRequirementTypes)
			requirementTypes.GET("/:id", configHandler.GetRequirementType)
			requirementTypes.PUT("/:id", configHandler.UpdateRequirementType)
			requirementTypes.DELETE("/:id", configHandler.DeleteRequirementType)
		}

		// Relationship Type routes
		relationshipTypes := config.Group("/relationship-types")
		{
			relationshipTypes.POST("", configHandler.CreateRelationshipType)
			relationshipTypes.GET("", configHandler.ListRelationshipTypes)
			relationshipTypes.GET("/:id", configHandler.GetRelationshipType)
			relationshipTypes.PUT("/:id", configHandler.UpdateRelationshipType)
			relationshipTypes.DELETE("/:id", configHandler.DeleteRelationshipType)
		}
	}

	return router, db, configService
}

func TestConfigIntegration_RequirementTypeLifecycle(t *testing.T) {
	router, db, _ := setupConfigIntegrationTest(t)

	t.Run("complete requirement type lifecycle", func(t *testing.T) {
		// 1. Create a new requirement type
		createReq := service.CreateRequirementTypeRequest{
			Name:        "Security",
			Description: stringPtr("Security-related requirements"),
		}

		reqBody, _ := json.Marshal(createReq)
		w := httptest.NewRecorder()
		httpReq, _ := http.NewRequest("POST", "/api/v1/config/requirement-types", bytes.NewBuffer(reqBody))
		httpReq.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusCreated, w.Code)

		var createdType models.RequirementType
		err := json.Unmarshal(w.Body.Bytes(), &createdType)
		require.NoError(t, err)
		assert.Equal(t, "Security", createdType.Name)
		assert.Equal(t, "Security-related requirements", *createdType.Description)
		assert.NotEqual(t, uuid.Nil, createdType.ID)

		// 2. Get the created requirement type
		w = httptest.NewRecorder()
		httpReq, _ = http.NewRequest("GET", "/api/v1/config/requirement-types/"+createdType.ID.String(), nil)

		router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusOK, w.Code)

		var retrievedType models.RequirementType
		err = json.Unmarshal(w.Body.Bytes(), &retrievedType)
		require.NoError(t, err)
		assert.Equal(t, createdType.ID, retrievedType.ID)
		assert.Equal(t, createdType.Name, retrievedType.Name)

		// 3. Update the requirement type
		updateReq := service.UpdateRequirementTypeRequest{
			Name:        stringPtr("Security & Privacy"),
			Description: stringPtr("Security and privacy-related requirements"),
		}

		reqBody, _ = json.Marshal(updateReq)
		w = httptest.NewRecorder()
		httpReq, _ = http.NewRequest("PUT", "/api/v1/config/requirement-types/"+createdType.ID.String(), bytes.NewBuffer(reqBody))
		httpReq.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusOK, w.Code)

		var updatedType models.RequirementType
		err = json.Unmarshal(w.Body.Bytes(), &updatedType)
		require.NoError(t, err)
		assert.Equal(t, "Security & Privacy", updatedType.Name)
		assert.Equal(t, "Security and privacy-related requirements", *updatedType.Description)

		// 4. List requirement types (should include default + created)
		w = httptest.NewRecorder()
		httpReq, _ = http.NewRequest("GET", "/api/v1/config/requirement-types", nil)

		router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusOK, w.Code)

		var listResponse map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &listResponse)
		require.NoError(t, err)

		count := listResponse["count"].(float64)
		assert.True(t, count >= 6) // 5 default + 1 created

		// 5. Delete the requirement type
		w = httptest.NewRecorder()
		httpReq, _ = http.NewRequest("DELETE", "/api/v1/config/requirement-types/"+createdType.ID.String(), nil)

		router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusNoContent, w.Code)

		// 6. Verify deletion - should return 404
		w = httptest.NewRecorder()
		httpReq, _ = http.NewRequest("GET", "/api/v1/config/requirement-types/"+createdType.ID.String(), nil)

		router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("prevent duplicate names", func(t *testing.T) {
		// Try to create a requirement type with existing name
		createReq := service.CreateRequirementTypeRequest{
			Name: "Functional", // This should already exist from default data
		}

		reqBody, _ := json.Marshal(createReq)
		w := httptest.NewRecorder()
		httpReq, _ := http.NewRequest("POST", "/api/v1/config/requirement-types", bytes.NewBuffer(reqBody))
		httpReq.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusConflict, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "Requirement type name already exists", response["error"])
	})

	t.Run("prevent deletion with associated requirements", func(t *testing.T) {
		// First, get a default requirement type
		var reqType models.RequirementType
		err := db.Where("name = ?", "Functional").First(&reqType).Error
		require.NoError(t, err)

		// Create a user for the requirement
		user := &models.User{
			ID:       uuid.New(),
			Username: "testuser",
			Email:    "test@example.com",
			Role:     "User",
		}
		err = db.Create(user).Error
		require.NoError(t, err)

		// Create an epic
		epic := &models.Epic{
			ID:          uuid.New(),
			CreatorID:   user.ID,
			AssigneeID:  user.ID,
			Priority:    models.PriorityMedium,
			Status:      models.EpicStatusDraft,
			Title:       "Test Epic",
			Description: stringPtr("Test epic description"),
		}
		err = db.Create(epic).Error
		require.NoError(t, err)

		// Create a user story
		userStory := &models.UserStory{
			ID:          uuid.New(),
			EpicID:      epic.ID,
			CreatorID:   user.ID,
			AssigneeID:  user.ID,
			Priority:    models.PriorityMedium,
			Status:      models.UserStoryStatusDraft,
			Title:       "Test User Story",
			Description: stringPtr("Test user story description"),
		}
		err = db.Create(userStory).Error
		require.NoError(t, err)

		// Create a requirement using the requirement type
		requirement := &models.Requirement{
			ID:          uuid.New(),
			UserStoryID: userStory.ID,
			CreatorID:   user.ID,
			AssigneeID:  user.ID,
			Priority:    models.PriorityMedium,
			Status:      models.RequirementStatusDraft,
			TypeID:      reqType.ID,
			Title:       "Test Requirement",
			Description: stringPtr("Test requirement description"),
		}
		err = db.Create(requirement).Error
		require.NoError(t, err)

		// Try to delete the requirement type
		w := httptest.NewRecorder()
		httpReq, _ := http.NewRequest("DELETE", "/api/v1/config/requirement-types/"+reqType.ID.String(), nil)

		router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusConflict, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "Requirement type has associated requirements and cannot be deleted", response["error"])
	})
}

func TestConfigIntegration_RelationshipTypeLifecycle(t *testing.T) {
	router, _, _ := setupConfigIntegrationTest(t)

	t.Run("complete relationship type lifecycle", func(t *testing.T) {
		// 1. Create a new relationship type
		createReq := service.CreateRelationshipTypeRequest{
			Name:        "implements",
			Description: stringPtr("Implementation relationship"),
		}

		reqBody, _ := json.Marshal(createReq)
		w := httptest.NewRecorder()
		httpReq, _ := http.NewRequest("POST", "/api/v1/config/relationship-types", bytes.NewBuffer(reqBody))
		httpReq.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusCreated, w.Code)

		var createdType models.RelationshipType
		err := json.Unmarshal(w.Body.Bytes(), &createdType)
		require.NoError(t, err)
		assert.Equal(t, "implements", createdType.Name)
		assert.Equal(t, "Implementation relationship", *createdType.Description)
		assert.NotEqual(t, uuid.Nil, createdType.ID)

		// 2. Get the created relationship type
		w = httptest.NewRecorder()
		httpReq, _ = http.NewRequest("GET", "/api/v1/config/relationship-types/"+createdType.ID.String(), nil)

		router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusOK, w.Code)

		var retrievedType models.RelationshipType
		err = json.Unmarshal(w.Body.Bytes(), &retrievedType)
		require.NoError(t, err)
		assert.Equal(t, createdType.ID, retrievedType.ID)
		assert.Equal(t, createdType.Name, retrievedType.Name)

		// 3. Update the relationship type
		updateReq := service.UpdateRelationshipTypeRequest{
			Name:        stringPtr("implements_feature"),
			Description: stringPtr("Feature implementation relationship"),
		}

		reqBody, _ = json.Marshal(updateReq)
		w = httptest.NewRecorder()
		httpReq, _ = http.NewRequest("PUT", "/api/v1/config/relationship-types/"+createdType.ID.String(), bytes.NewBuffer(reqBody))
		httpReq.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusOK, w.Code)

		var updatedType models.RelationshipType
		err = json.Unmarshal(w.Body.Bytes(), &updatedType)
		require.NoError(t, err)
		assert.Equal(t, "implements_feature", updatedType.Name)
		assert.Equal(t, "Feature implementation relationship", *updatedType.Description)

		// 4. List relationship types (should include default + created)
		w = httptest.NewRecorder()
		httpReq, _ = http.NewRequest("GET", "/api/v1/config/relationship-types", nil)

		router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusOK, w.Code)

		var listResponse map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &listResponse)
		require.NoError(t, err)

		count := listResponse["count"].(float64)
		assert.True(t, count >= 6) // 5 default + 1 created

		// 5. Delete the relationship type
		w = httptest.NewRecorder()
		httpReq, _ = http.NewRequest("DELETE", "/api/v1/config/relationship-types/"+createdType.ID.String(), nil)

		router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusNoContent, w.Code)

		// 6. Verify deletion - should return 404
		w = httptest.NewRecorder()
		httpReq, _ = http.NewRequest("GET", "/api/v1/config/relationship-types/"+createdType.ID.String(), nil)

		router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("prevent duplicate names", func(t *testing.T) {
		// Try to create a relationship type with existing name
		createReq := service.CreateRelationshipTypeRequest{
			Name: "depends_on", // This should already exist from default data
		}

		reqBody, _ := json.Marshal(createReq)
		w := httptest.NewRecorder()
		httpReq, _ := http.NewRequest("POST", "/api/v1/config/relationship-types", bytes.NewBuffer(reqBody))
		httpReq.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusConflict, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "Relationship type name already exists", response["error"])
	})
}

func TestConfigIntegration_ValidationIntegration(t *testing.T) {
	_, db, configService := setupConfigIntegrationTest(t)

	t.Run("validate requirement type integration", func(t *testing.T) {
		// Get a valid requirement type
		var reqType models.RequirementType
		err := db.Where("name = ?", "Functional").First(&reqType).Error
		require.NoError(t, err)

		// Test validation with valid type
		err = configService.ValidateRequirementType(reqType.ID)
		assert.NoError(t, err)

		// Test validation with invalid type
		invalidID := uuid.New()
		err = configService.ValidateRequirementType(invalidID)
		assert.Error(t, err)
		assert.Equal(t, service.ErrRequirementTypeNotFound, err)
	})

	t.Run("validate relationship type integration", func(t *testing.T) {
		// Get a valid relationship type
		var relType models.RelationshipType
		err := db.Where("name = ?", "depends_on").First(&relType).Error
		require.NoError(t, err)

		// Test validation with valid type
		err = configService.ValidateRelationshipType(relType.ID)
		assert.NoError(t, err)

		// Test validation with invalid type
		invalidID := uuid.New()
		err = configService.ValidateRelationshipType(invalidID)
		assert.Error(t, err)
		assert.Equal(t, service.ErrRelationshipTypeNotFound, err)
	})
}

func TestConfigIntegration_FilteringAndPagination(t *testing.T) {
	router, _, _ := setupConfigIntegrationTest(t)

	t.Run("requirement types filtering and pagination", func(t *testing.T) {
		// Test with limit
		w := httptest.NewRecorder()
		httpReq, _ := http.NewRequest("GET", "/api/v1/config/requirement-types?limit=2", nil)

		router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		types := response["requirement_types"].([]interface{})
		assert.True(t, len(types) <= 2)

		// Test with ordering
		w = httptest.NewRecorder()
		httpReq, _ = http.NewRequest("GET", "/api/v1/config/requirement-types?order_by=name", nil)

		router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusOK, w.Code)

		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		types = response["requirement_types"].([]interface{})
		assert.True(t, len(types) >= 5) // Should have at least the default types
	})

	t.Run("relationship types filtering and pagination", func(t *testing.T) {
		// Test with limit and offset
		w := httptest.NewRecorder()
		httpReq, _ := http.NewRequest("GET", "/api/v1/config/relationship-types?limit=3&offset=1", nil)

		router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		types := response["relationship_types"].([]interface{})
		assert.True(t, len(types) <= 3)
	})
}

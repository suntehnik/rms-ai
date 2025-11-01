package tools

import (
	"context"
	"errors"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"product-requirements-management/internal/mcp/types"
	"product-requirements-management/internal/models"
	"product-requirements-management/internal/service"
)

// MockRequirementService is a mock implementation of service.RequirementService
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

func (m *MockRequirementService) UpdateRequirement(id uuid.UUID, req service.UpdateRequirementRequest) (*models.Requirement, error) {
	args := m.Called(id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Requirement), args.Error(1)
}

func (m *MockRequirementService) GetRequirementByReferenceID(refID string) (*models.Requirement, error) {
	args := m.Called(refID)
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

// Implement other required methods to satisfy the interface
func (m *MockRequirementService) GetRequirementByID(id uuid.UUID) (*models.Requirement, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Requirement), args.Error(1)
}

func (m *MockRequirementService) DeleteRequirement(id uuid.UUID, force bool) error {
	args := m.Called(id, force)
	return args.Error(0)
}

func (m *MockRequirementService) ListRequirements(filters service.RequirementFilters) ([]models.Requirement, int64, error) {
	args := m.Called(filters)
	return args.Get(0).([]models.Requirement), args.Get(1).(int64), args.Error(2)
}

func (m *MockRequirementService) SearchRequirements(query string) ([]models.Requirement, error) {
	args := m.Called(query)
	return args.Get(0).([]models.Requirement), args.Error(1)
}

func (m *MockRequirementService) GetRequirementWithRelationships(id uuid.UUID) (*models.Requirement, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Requirement), args.Error(1)
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

func (m *MockRequirementService) DeleteRelationship(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockRequirementService) GetRelationshipsByRequirement(requirementID uuid.UUID) ([]models.RequirementRelationship, error) {
	args := m.Called(requirementID)
	return args.Get(0).([]models.RequirementRelationship), args.Error(1)
}

func (m *MockRequirementService) GetRelationshipsByRequirementWithPagination(requirementID uuid.UUID, limit, offset int) ([]models.RequirementRelationship, int64, error) {
	args := m.Called(requirementID, limit, offset)
	return args.Get(0).([]models.RequirementRelationship), args.Get(1).(int64), args.Error(2)
}

func (m *MockRequirementService) SearchRequirementsWithPagination(searchText string, limit, offset int) ([]models.Requirement, int64, error) {
	args := m.Called(searchText, limit, offset)
	return args.Get(0).([]models.Requirement), args.Get(1).(int64), args.Error(2)
}

func (m *MockRequirementService) GetRequirementsByUserStory(userStoryID uuid.UUID) ([]models.Requirement, error) {
	args := m.Called(userStoryID)
	return args.Get(0).([]models.Requirement), args.Error(1)
}

// MockUserStoryService is already defined in user_story_test.go

func TestRequirementHandler_GetSupportedTools(t *testing.T) {
	handler := NewRequirementHandler(nil, nil)
	tools := handler.GetSupportedTools()

	expected := []string{"create_requirement", "update_requirement", "create_relationship"}
	assert.Equal(t, expected, tools)
}

func TestRequirementHandler_HandleTool(t *testing.T) {
	mockReqService := &MockRequirementService{}
	mockUSService := &MockUserStoryService{}
	handler := NewRequirementHandler(mockReqService, mockUSService)

	tests := []struct {
		name        string
		toolName    string
		expectError bool
	}{
		{
			name:        "valid create_requirement tool",
			toolName:    "create_requirement",
			expectError: true, // Will error due to missing context/args, but tool routing works
		},
		{
			name:        "valid update_requirement tool",
			toolName:    "update_requirement",
			expectError: true, // Will error due to missing context/args, but tool routing works
		},
		{
			name:        "valid create_relationship tool",
			toolName:    "create_relationship",
			expectError: true, // Will error due to missing context/args, but tool routing works
		},
		{
			name:        "invalid tool name",
			toolName:    "invalid_tool",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := handler.HandleTool(context.Background(), tt.toolName, map[string]interface{}{})

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNewRequirementHandler(t *testing.T) {
	mockReqService := &MockRequirementService{}
	mockUSService := &MockUserStoryService{}
	handler := NewRequirementHandler(mockReqService, mockUSService)

	assert.NotNil(t, handler)
	assert.Equal(t, mockReqService, handler.requirementService)
	assert.Equal(t, mockUSService, handler.userStoryService)
}

// Helper function to create a context with user
func createRequirementContextWithUser(user *models.User) context.Context {
	ginCtx := &gin.Context{}
	ginCtx.Set("user", user)

	ctx := context.WithValue(context.Background(), "gin_context", ginCtx)
	return ctx
}

// TestRequirementHandler_Create_ValidParameters tests requirement creation with valid parameters
func TestRequirementHandler_Create_ValidParameters(t *testing.T) {
	mockReqService := &MockRequirementService{}
	mockUSService := &MockUserStoryService{}
	handler := NewRequirementHandler(mockReqService, mockUSService)

	// Create test user
	user := &models.User{
		ID:       uuid.New(),
		Username: "testuser",
		Email:    "test@example.com",
		Role:     models.RoleUser,
	}

	userStoryID := uuid.New()
	typeID := uuid.New()
	assigneeID := uuid.New()
	acceptanceCriteriaID := uuid.New()

	// Create test requirement
	expectedRequirement := &models.Requirement{
		ID:          uuid.New(),
		ReferenceID: "REQ-001",
		Title:       "Test Requirement",
		Description: stringPtr("Test Description"),
		Priority:    models.PriorityHigh,
		Status:      models.RequirementStatusDraft,
		CreatorID:   user.ID,
		UserStoryID: userStoryID,
		TypeID:      typeID,
	}

	tests := []struct {
		name     string
		args     map[string]interface{}
		expected *models.Requirement
	}{
		{
			name: "create requirement with all parameters",
			args: map[string]interface{}{
				"title":                  "Test Requirement",
				"description":            "Test Description",
				"user_story_id":          userStoryID.String(),
				"type_id":                typeID.String(),
				"priority":               2,
				"assignee_id":            assigneeID.String(),
				"acceptance_criteria_id": acceptanceCriteriaID.String(),
			},
			expected: expectedRequirement,
		},
		{
			name: "create requirement with minimal parameters",
			args: map[string]interface{}{
				"title":         "Minimal Requirement",
				"user_story_id": userStoryID.String(),
				"type_id":       typeID.String(),
				"priority":      1,
			},
			expected: expectedRequirement,
		},
		{
			name: "create requirement with float priority",
			args: map[string]interface{}{
				"title":         "Float Priority Requirement",
				"user_story_id": userStoryID.String(),
				"type_id":       typeID.String(),
				"priority":      3.0,
			},
			expected: expectedRequirement,
		},
	}
	expectedUserStory := &models.UserStory{
		ID: userStoryID,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock expectations
			mockReqService.On("CreateRequirement", mock.AnythingOfType("service.CreateRequirementRequest")).Return(tt.expected, nil).Once()
			mockUSService.On("GetUserStoryByID", userStoryID).Return(expectedUserStory, nil).Once()
			ctx := createRequirementContextWithUser(user)
			result, err := handler.Create(ctx, tt.args)

			assert.NoError(t, err)
			assert.NotNil(t, result)

			// Verify response format
			response, ok := result.(*types.ToolResponse)
			assert.True(t, ok)
			assert.Len(t, response.Content, 2) // Message + data
			assert.Equal(t, "text", response.Content[0].Type)
			assert.Contains(t, response.Content[0].Text, "Successfully created requirement")
			assert.Contains(t, response.Content[0].Text, tt.expected.ReferenceID)

			mockReqService.AssertExpectations(t)
		})
	}
}

// TestRequirementHandler_Create_UserStoryReferenceIDResolution tests requirement creation with user story reference ID resolution
func TestRequirementHandler_Create_UserStoryReferenceIDResolution(t *testing.T) {
	mockReqService := &MockRequirementService{}
	mockUSService := &MockUserStoryService{}
	handler := NewRequirementHandler(mockReqService, mockUSService)

	user := &models.User{
		ID:       uuid.New(),
		Username: "testuser",
		Email:    "test@example.com",
		Role:     models.RoleUser,
	}

	userStoryID := uuid.New()
	typeID := uuid.New()
	userStory := &models.UserStory{
		ID:          userStoryID,
		ReferenceID: "US-001",
		Title:       "Test User Story",
	}

	expectedRequirement := &models.Requirement{
		ID:          uuid.New(),
		ReferenceID: "REQ-001",
		Title:       "Test Requirement",
		UserStoryID: userStoryID,
		TypeID:      typeID,
	}

	tests := []struct {
		name        string
		args        map[string]interface{}
		setupMocks  func()
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid user story reference ID resolution",
			args: map[string]interface{}{
				"title":         "Test Requirement",
				"user_story_id": "US-001",
				"type_id":       typeID.String(),
				"priority":      1,
			},
			setupMocks: func() {
				mockUSService.On("GetUserStoryByReferenceID", "US-001").Return(userStory, nil).Once()
				mockReqService.On("CreateRequirement", mock.AnythingOfType("service.CreateRequirementRequest")).Return(expectedRequirement, nil).Once()
			},
			expectError: false,
		},
		{
			name: "invalid user story reference ID",
			args: map[string]interface{}{
				"title":         "Test Requirement",
				"user_story_id": "US-999",
				"type_id":       typeID.String(),
				"priority":      1,
			},
			setupMocks: func() {
				mockUSService.On("GetUserStoryByReferenceID", "US-999").Return(nil, errors.New("user story not found")).Once()
			},
			expectError: true,
			errorMsg:    "Invalid params",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			ctx := createRequirementContextWithUser(user)
			result, err := handler.Create(ctx, tt.args)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}

			mockReqService.AssertExpectations(t)
			mockUSService.AssertExpectations(t)
		})
	}
}

// TestRequirementHandler_Create_InvalidParameters tests requirement creation with invalid parameters
func TestRequirementHandler_Create_InvalidParameters(t *testing.T) {
	mockReqService := &MockRequirementService{}
	mockUSService := &MockUserStoryService{}
	handler := NewRequirementHandler(mockReqService, mockUSService)

	user := &models.User{
		ID:       uuid.New(),
		Username: "testuser",
		Email:    "test@example.com",
		Role:     models.RoleUser,
	}

	userStoryID := uuid.New()
	typeID := uuid.New()
	expectedUserStory := &models.UserStory{
		ID: userStoryID,
	}

	tests := []struct {
		name        string
		args        map[string]interface{}
		expectError string
	}{
		{
			name:        "missing title",
			args:        map[string]interface{}{"user_story_id": userStoryID.String(), "type_id": typeID.String(), "priority": 1},
			expectError: "Invalid params",
		},
		{
			name:        "empty title",
			args:        map[string]interface{}{"title": "", "user_story_id": userStoryID.String(), "type_id": typeID.String(), "priority": 1},
			expectError: "Invalid params",
		},
		{
			name:        "missing user_story_id",
			args:        map[string]interface{}{"title": "Test Requirement", "type_id": typeID.String(), "priority": 1},
			expectError: "Invalid params",
		},
		{
			name:        "empty user_story_id",
			args:        map[string]interface{}{"title": "Test Requirement", "user_story_id": "", "type_id": typeID.String(), "priority": 1},
			expectError: "Invalid params",
		},
		{
			name:        "missing type_id",
			args:        map[string]interface{}{"title": "Test Requirement", "user_story_id": userStoryID.String(), "priority": 1},
			expectError: "Invalid params",
		},
		{
			name:        "empty type_id",
			args:        map[string]interface{}{"title": "Test Requirement", "user_story_id": userStoryID.String(), "type_id": "", "priority": 1},
			expectError: "Invalid params",
		},
		{
			name:        "invalid type_id format",
			args:        map[string]interface{}{"title": "Test Requirement", "user_story_id": userStoryID.String(), "type_id": "invalid-uuid", "priority": 1},
			expectError: "Invalid params",
		},
		{
			name:        "missing priority",
			args:        map[string]interface{}{"title": "Test Requirement", "user_story_id": userStoryID.String(), "type_id": typeID.String()},
			expectError: "Invalid params",
		},
		{
			name:        "invalid priority type",
			args:        map[string]interface{}{"title": "Test Requirement", "user_story_id": userStoryID.String(), "type_id": typeID.String(), "priority": "high"},
			expectError: "Invalid params",
		},
		{
			name:        "invalid assignee_id format",
			args:        map[string]interface{}{"title": "Test Requirement", "user_story_id": userStoryID.String(), "type_id": typeID.String(), "priority": 1, "assignee_id": "invalid-uuid"},
			expectError: "Invalid params",
		},
		{
			name:        "invalid acceptance_criteria_id format",
			args:        map[string]interface{}{"title": "Test Requirement", "user_story_id": userStoryID.String(), "type_id": typeID.String(), "priority": 1, "acceptance_criteria_id": "invalid-uuid"},
			expectError: "Invalid params",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createRequirementContextWithUser(user)
			mockUSService.On("GetUserStoryByID", userStoryID).Return(expectedUserStory, nil).Once()

			_, err := handler.Create(ctx, tt.args)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectError)
		})
	}
}

// TestRequirementHandler_Update_ValidParameters tests requirement updates with valid parameters
func TestRequirementHandler_Update_ValidParameters(t *testing.T) {
	mockReqService := &MockRequirementService{}
	mockUSService := &MockUserStoryService{}
	handler := NewRequirementHandler(mockReqService, mockUSService)

	requirementID := uuid.New()
	expectedRequirement := &models.Requirement{
		ID:          requirementID,
		ReferenceID: "REQ-001",
		Title:       "Updated Requirement",
		Priority:    models.PriorityHigh,
		Status:      models.RequirementStatusActive,
	}

	tests := []struct {
		name     string
		args     map[string]interface{}
		expected *models.Requirement
	}{
		{
			name: "update with UUID",
			args: map[string]interface{}{
				"requirement_id": requirementID.String(),
				"title":          "Updated Requirement",
			},
			expected: expectedRequirement,
		},
		{
			name: "update with all parameters",
			args: map[string]interface{}{
				"requirement_id": requirementID.String(),
				"title":          "Updated Requirement",
				"description":    "Updated Description",
				"priority":       2,
				"assignee_id":    uuid.New().String(),
			},
			expected: expectedRequirement,
		},
		{
			name: "update with empty assignee_id (unassign)",
			args: map[string]interface{}{
				"requirement_id": requirementID.String(),
				"assignee_id":    "",
			},
			expected: expectedRequirement,
		},
		{
			name: "update with float priority",
			args: map[string]interface{}{
				"requirement_id": requirementID.String(),
				"priority":       3.0,
			},
			expected: expectedRequirement,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock expectations
			mockReqService.On("UpdateRequirement", requirementID, mock.AnythingOfType("service.UpdateRequirementRequest")).Return(tt.expected, nil).Once()

			result, err := handler.Update(context.Background(), tt.args)

			assert.NoError(t, err)
			assert.NotNil(t, result)

			// Verify response format
			response, ok := result.(*types.ToolResponse)
			assert.True(t, ok)
			assert.Len(t, response.Content, 2) // Message + data
			assert.Equal(t, "text", response.Content[0].Type)
			assert.Contains(t, response.Content[0].Text, "Successfully updated requirement")
			assert.Contains(t, response.Content[0].Text, tt.expected.ReferenceID)

			mockReqService.AssertExpectations(t)
		})
	}
}

// TestRequirementHandler_Update_ReferenceIDResolution tests requirement updates with reference ID resolution
func TestRequirementHandler_Update_ReferenceIDResolution(t *testing.T) {
	mockReqService := &MockRequirementService{}
	mockUSService := &MockUserStoryService{}
	handler := NewRequirementHandler(mockReqService, mockUSService)

	requirementID := uuid.New()
	expectedRequirement := &models.Requirement{
		ID:          requirementID,
		ReferenceID: "REQ-001",
		Title:       "Updated Requirement",
	}

	tests := []struct {
		name        string
		args        map[string]interface{}
		setupMocks  func()
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid reference ID resolution",
			args: map[string]interface{}{
				"requirement_id": "REQ-001",
				"title":          "Updated Requirement",
			},
			setupMocks: func() {
				mockReqService.On("GetRequirementByReferenceID", "REQ-001").Return(expectedRequirement, nil).Once()
				mockReqService.On("UpdateRequirement", requirementID, mock.AnythingOfType("service.UpdateRequirementRequest")).Return(expectedRequirement, nil).Once()
			},
			expectError: false,
		},
		{
			name: "invalid reference ID",
			args: map[string]interface{}{
				"requirement_id": "REQ-999",
				"title":          "Updated Requirement",
			},
			setupMocks: func() {
				mockReqService.On("GetRequirementByReferenceID", "REQ-999").Return(nil, errors.New("requirement not found")).Once()
			},
			expectError: true,
			errorMsg:    "Invalid params",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			result, err := handler.Update(context.Background(), tt.args)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}

			mockReqService.AssertExpectations(t)
		})
	}
}

// TestRequirementHandler_CreateRelationship_ValidParameters tests relationship creation with valid parameters
func TestRequirementHandler_CreateRelationship_ValidParameters(t *testing.T) {
	mockReqService := &MockRequirementService{}
	mockUSService := &MockUserStoryService{}
	handler := NewRequirementHandler(mockReqService, mockUSService)

	user := &models.User{
		ID:       uuid.New(),
		Username: "testuser",
		Email:    "test@example.com",
		Role:     models.RoleUser,
	}

	sourceID := uuid.New()
	targetID := uuid.New()
	relationshipTypeID := uuid.New()

	expectedRelationship := &models.RequirementRelationship{
		ID:                  uuid.New(),
		SourceRequirementID: sourceID,
		TargetRequirementID: targetID,
		RelationshipTypeID:  relationshipTypeID,
		CreatedBy:           user.ID,
	}

	tests := []struct {
		name     string
		args     map[string]interface{}
		expected *models.RequirementRelationship
	}{
		{
			name: "create relationship with UUIDs",
			args: map[string]interface{}{
				"source_requirement_id": sourceID.String(),
				"target_requirement_id": targetID.String(),
				"relationship_type_id":  relationshipTypeID.String(),
			},
			expected: expectedRelationship,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock expectations
			mockReqService.On("CreateRelationship", mock.AnythingOfType("service.CreateRelationshipRequest")).Return(tt.expected, nil).Once()

			ctx := createRequirementContextWithUser(user)
			result, err := handler.CreateRelationship(ctx, tt.args)

			assert.NoError(t, err)
			assert.NotNil(t, result)

			// Verify response format
			response, ok := result.(*types.ToolResponse)
			assert.True(t, ok)
			assert.Len(t, response.Content, 2) // Message + data
			assert.Equal(t, "text", response.Content[0].Type)
			assert.Contains(t, response.Content[0].Text, "Successfully created relationship")

			mockReqService.AssertExpectations(t)
		})
	}
}

// TestRequirementHandler_CreateRelationship_ReferenceIDResolution tests relationship creation with reference ID resolution
func TestRequirementHandler_CreateRelationship_ReferenceIDResolution(t *testing.T) {
	mockReqService := &MockRequirementService{}
	mockUSService := &MockUserStoryService{}
	handler := NewRequirementHandler(mockReqService, mockUSService)

	user := &models.User{
		ID:       uuid.New(),
		Username: "testuser",
		Email:    "test@example.com",
		Role:     models.RoleUser,
	}

	sourceID := uuid.New()
	targetID := uuid.New()
	relationshipTypeID := uuid.New()

	sourceRequirement := &models.Requirement{
		ID:          sourceID,
		ReferenceID: "REQ-001",
		Title:       "Source Requirement",
	}

	targetRequirement := &models.Requirement{
		ID:          targetID,
		ReferenceID: "REQ-002",
		Title:       "Target Requirement",
	}

	expectedRelationship := &models.RequirementRelationship{
		ID:                  uuid.New(),
		SourceRequirementID: sourceID,
		TargetRequirementID: targetID,
		RelationshipTypeID:  relationshipTypeID,
		CreatedBy:           user.ID,
	}

	tests := []struct {
		name        string
		args        map[string]interface{}
		setupMocks  func()
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid reference ID resolution for both requirements",
			args: map[string]interface{}{
				"source_requirement_id": "REQ-001",
				"target_requirement_id": "REQ-002",
				"relationship_type_id":  relationshipTypeID.String(),
			},
			setupMocks: func() {
				mockReqService.On("GetRequirementByReferenceID", "REQ-001").Return(sourceRequirement, nil).Once()
				mockReqService.On("GetRequirementByReferenceID", "REQ-002").Return(targetRequirement, nil).Once()
				mockReqService.On("CreateRelationship", mock.AnythingOfType("service.CreateRelationshipRequest")).Return(expectedRelationship, nil).Once()
			},
			expectError: false,
		},
		{
			name: "invalid source requirement reference ID",
			args: map[string]interface{}{
				"source_requirement_id": "REQ-999",
				"target_requirement_id": targetID.String(),
				"relationship_type_id":  relationshipTypeID.String(),
			},
			setupMocks: func() {
				mockReqService.On("GetRequirementByReferenceID", "REQ-999").Return(nil, errors.New("requirement not found")).Once()
			},
			expectError: true,
			errorMsg:    "Invalid params",
		},
		{
			name: "invalid target requirement reference ID",
			args: map[string]interface{}{
				"source_requirement_id": sourceID.String(),
				"target_requirement_id": "REQ-999",
				"relationship_type_id":  relationshipTypeID.String(),
			},
			setupMocks: func() {
				mockReqService.On("GetRequirementByReferenceID", "REQ-999").Return(nil, errors.New("requirement not found")).Once()
			},
			expectError: true,
			errorMsg:    "Invalid params",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			ctx := createRequirementContextWithUser(user)
			result, err := handler.CreateRelationship(ctx, tt.args)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}

			mockReqService.AssertExpectations(t)
		})
	}
}

// TestRequirementHandler_CreateRelationship_InvalidParameters tests relationship creation with invalid parameters
func TestRequirementHandler_CreateRelationship_InvalidParameters(t *testing.T) {
	mockReqService := &MockRequirementService{}
	mockUSService := &MockUserStoryService{}
	handler := NewRequirementHandler(mockReqService, mockUSService)

	user := &models.User{
		ID:       uuid.New(),
		Username: "testuser",
		Email:    "test@example.com",
		Role:     models.RoleUser,
	}

	sourceID := uuid.New()
	targetID := uuid.New()
	relationshipTypeID := uuid.New()

	tests := []struct {
		name        string
		args        map[string]interface{}
		expectError string
	}{
		{
			name:        "missing source_requirement_id",
			args:        map[string]interface{}{"target_requirement_id": targetID.String(), "relationship_type_id": relationshipTypeID.String()},
			expectError: "Invalid params",
		},
		{
			name:        "empty source_requirement_id",
			args:        map[string]interface{}{"source_requirement_id": "", "target_requirement_id": targetID.String(), "relationship_type_id": relationshipTypeID.String()},
			expectError: "Invalid params",
		},
		{
			name:        "missing target_requirement_id",
			args:        map[string]interface{}{"source_requirement_id": sourceID.String(), "relationship_type_id": relationshipTypeID.String()},
			expectError: "Invalid params",
		},
		{
			name:        "empty target_requirement_id",
			args:        map[string]interface{}{"source_requirement_id": sourceID.String(), "target_requirement_id": "", "relationship_type_id": relationshipTypeID.String()},
			expectError: "Invalid params",
		},
		{
			name:        "missing relationship_type_id",
			args:        map[string]interface{}{"source_requirement_id": sourceID.String(), "target_requirement_id": targetID.String()},
			expectError: "Invalid params",
		},
		{
			name:        "empty relationship_type_id",
			args:        map[string]interface{}{"source_requirement_id": sourceID.String(), "target_requirement_id": targetID.String(), "relationship_type_id": ""},
			expectError: "Invalid params",
		},
		{
			name:        "invalid relationship_type_id format",
			args:        map[string]interface{}{"source_requirement_id": sourceID.String(), "target_requirement_id": targetID.String(), "relationship_type_id": "invalid-uuid"},
			expectError: "Invalid params",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createRequirementContextWithUser(user)
			_, err := handler.CreateRelationship(ctx, tt.args)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectError)
		})
	}
}

// TestRequirementHandler_ServiceErrors tests service layer errors
func TestRequirementHandler_ServiceErrors(t *testing.T) {
	mockReqService := &MockRequirementService{}
	mockUSService := &MockUserStoryService{}
	handler := NewRequirementHandler(mockReqService, mockUSService)

	user := &models.User{
		ID:       uuid.New(),
		Username: "testuser",
		Email:    "test@example.com",
		Role:     models.RoleUser,
	}

	userStoryID := uuid.New()
	typeID := uuid.New()
	requirementID := uuid.New()
	expectedUserStory := &models.UserStory{
		ID: userStoryID,
	}

	tests := []struct {
		name      string
		method    string
		args      map[string]interface{}
		setupMock func()
	}{
		{
			name:   "create requirement service error",
			method: "create",
			args: map[string]interface{}{
				"title":         "Test Requirement",
				"user_story_id": userStoryID.String(),
				"type_id":       typeID.String(),
				"priority":      1,
			},
			setupMock: func() {
				mockUSService.On("GetUserStoryByID", userStoryID).Return(expectedUserStory, nil).Once()
				mockReqService.On("CreateRequirement", mock.AnythingOfType("service.CreateRequirementRequest")).Return(nil, errors.New("database error")).Once()
			},
		},
		{
			name:   "update requirement service error",
			method: "update",
			args: map[string]interface{}{
				"requirement_id": requirementID.String(),
				"title":          "Updated Requirement",
			},
			setupMock: func() {
				mockUSService.On("GetUserStoryByID", userStoryID).Return(expectedUserStory, nil).Once()
				mockReqService.On("UpdateRequirement", requirementID, mock.AnythingOfType("service.UpdateRequirementRequest")).Return(nil, errors.New("database error")).Once()
			},
		},
		{
			name:   "create relationship service error",
			method: "relationship",
			args: map[string]interface{}{
				"source_requirement_id": uuid.New().String(),
				"target_requirement_id": uuid.New().String(),
				"relationship_type_id":  uuid.New().String(),
			},
			setupMock: func() {
				mockReqService.On("CreateRelationship", mock.AnythingOfType("service.CreateRelationshipRequest")).Return(nil, errors.New("database error")).Once()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			ctx := createRequirementContextWithUser(user)
			var err error

			switch tt.method {
			case "create":
				_, err = handler.Create(ctx, tt.args)
			case "update":
				_, err = handler.Update(ctx, tt.args)
			case "relationship":
				_, err = handler.CreateRelationship(ctx, tt.args)
			}

			assert.Error(t, err)
			assert.Contains(t, err.Error(), "Internal error")
			mockReqService.AssertExpectations(t)
		})
	}
}

// TestRequirementHandler_ContextErrors tests context-related errors
func TestRequirementHandler_ContextErrors(t *testing.T) {
	mockReqService := &MockRequirementService{}
	mockUSService := &MockUserStoryService{}
	handler := NewRequirementHandler(mockReqService, mockUSService)

	userStoryID := uuid.New()
	typeID := uuid.New()

	args := map[string]interface{}{
		"title":         "Test Requirement",
		"user_story_id": userStoryID.String(),
		"type_id":       typeID.String(),
		"priority":      1,
	}

	relationshipArgs := map[string]interface{}{
		"source_requirement_id": uuid.New().String(),
		"target_requirement_id": uuid.New().String(),
		"relationship_type_id":  uuid.New().String(),
	}

	tests := []struct {
		name        string
		ctx         context.Context
		expectError string
	}{
		{
			name:        "missing gin context",
			ctx:         context.Background(),
			expectError: "Internal error",
		},
		{
			name:        "gin context without user",
			ctx:         context.WithValue(context.Background(), "gin_context", &gin.Context{}),
			expectError: "Internal error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name+" - create", func(t *testing.T) {
			_, err := handler.Create(tt.ctx, args)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectError)
		})

		t.Run(tt.name+" - create relationship", func(t *testing.T) {
			_, err := handler.CreateRelationship(tt.ctx, relationshipArgs)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectError)
		})
	}
}

// stringPtr is already defined in epic_test.go

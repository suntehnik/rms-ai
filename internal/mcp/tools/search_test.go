package tools

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"product-requirements-management/internal/mcp/types"
	"product-requirements-management/internal/models"
	"product-requirements-management/internal/service"
)

// MockSearchServiceForSearchHandler is a mock implementation specific to search handler tests
type MockSearchServiceForSearchHandler struct {
	mock.Mock
}

func (m *MockSearchServiceForSearchHandler) Search(ctx context.Context, options service.SearchOptions) (*service.SearchResponse, error) {
	args := m.Called(ctx, options)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.SearchResponse), args.Error(1)
}

func (m *MockSearchServiceForSearchHandler) SearchByReferenceID(ctx context.Context, referenceID string, entityTypes []string) (*service.SearchResponse, error) {
	args := m.Called(ctx, referenceID, entityTypes)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.SearchResponse), args.Error(1)
}

func (m *MockSearchServiceForSearchHandler) InvalidateCache(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func TestSearchHandler_GetSupportedTools(t *testing.T) {
	handler := NewSearchHandler(nil, nil)
	tools := handler.GetSupportedTools()

	expected := []string{"search_global", "search_requirements"}
	assert.Equal(t, expected, tools)
}

func TestSearchHandler_HandleTool(t *testing.T) {
	mockSearchService := &MockSearchServiceForSearchHandler{}
	mockRequirementService := &MockRequirementService{}
	handler := NewSearchHandler(mockSearchService, mockRequirementService)

	tests := []struct {
		name        string
		toolName    string
		expectError bool
	}{
		{
			name:        "valid search_global tool",
			toolName:    "search_global",
			expectError: true, // Will error due to missing query, but tool routing works
		},
		{
			name:        "valid search_requirements tool",
			toolName:    "search_requirements",
			expectError: true, // Will error due to missing query, but tool routing works
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

func TestSearchHandler_Global_ValidationErrors(t *testing.T) {
	mockSearchService := &MockSearchServiceForSearchHandler{}
	mockRequirementService := &MockRequirementService{}
	handler := NewSearchHandler(mockSearchService, mockRequirementService)

	tests := []struct {
		name string
		args map[string]interface{}
	}{
		{
			name: "missing query",
			args: map[string]interface{}{},
		},
		{
			name: "empty query",
			args: map[string]interface{}{"query": ""},
		},
		{
			name: "invalid query type",
			args: map[string]interface{}{"query": 123},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := handler.Global(context.Background(), tt.args)
			assert.Error(t, err)
		})
	}
}

func TestSearchHandler_Global_Success(t *testing.T) {
	mockSearchService := &MockSearchServiceForSearchHandler{}
	mockRequirementService := &MockRequirementService{}
	handler := NewSearchHandler(mockSearchService, mockRequirementService)

	// Mock successful search response
	expectedResponse := &service.SearchResponse{
		Results: []service.SearchResult{
			{
				ID:          uuid.New(),
				ReferenceID: "EP-001",
				Type:        "epic",
				Title:       "Test Epic",
				Description: "Test Description",
			},
		},
		Total:  1,
		Limit:  50,
		Offset: 0,
		Query:  "test",
	}

	mockSearchService.On("Search", mock.Anything, mock.MatchedBy(func(opts service.SearchOptions) bool {
		return opts.Query == "test" && opts.Limit == 50 && opts.Offset == 0
	})).Return(expectedResponse, nil)

	args := map[string]interface{}{
		"query": "test",
	}

	result, err := handler.Global(context.Background(), args)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	mockSearchService.AssertExpectations(t)
}

func TestSearchHandler_Global_ServiceError(t *testing.T) {
	mockSearchService := &MockSearchServiceForSearchHandler{}
	mockRequirementService := &MockRequirementService{}
	handler := NewSearchHandler(mockSearchService, mockRequirementService)

	mockSearchService.On("Search", mock.Anything, mock.Anything).Return(nil, errors.New("search service error"))

	args := map[string]interface{}{
		"query": "test",
	}

	_, err := handler.Global(context.Background(), args)
	assert.Error(t, err)

	mockSearchService.AssertExpectations(t)
}

func TestSearchHandler_Requirements_ValidationErrors(t *testing.T) {
	mockSearchService := &MockSearchServiceForSearchHandler{}
	mockRequirementService := &MockRequirementService{}
	handler := NewSearchHandler(mockSearchService, mockRequirementService)

	tests := []struct {
		name string
		args map[string]interface{}
	}{
		{
			name: "missing query",
			args: map[string]interface{}{},
		},
		{
			name: "empty query",
			args: map[string]interface{}{"query": ""},
		},
		{
			name: "invalid query type",
			args: map[string]interface{}{"query": 123},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := handler.Requirements(context.Background(), tt.args)
			assert.Error(t, err)
		})
	}
}

func TestSearchHandler_Requirements_Success(t *testing.T) {
	mockSearchService := &MockSearchServiceForSearchHandler{}
	mockRequirementService := &MockRequirementService{}
	handler := NewSearchHandler(mockSearchService, mockRequirementService)

	// Mock successful requirements search
	desc := "Test Description"
	expectedRequirements := []models.Requirement{
		{
			ID:          uuid.New(),
			ReferenceID: "REQ-001",
			Title:       "Test Requirement",
			Description: &desc,
		},
	}

	mockRequirementService.On("SearchRequirements", "test").Return(expectedRequirements, nil)

	args := map[string]interface{}{
		"query": "test",
	}

	result, err := handler.Requirements(context.Background(), args)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	mockRequirementService.AssertExpectations(t)
}

func TestSearchHandler_Requirements_ServiceError(t *testing.T) {
	mockSearchService := &MockSearchServiceForSearchHandler{}
	mockRequirementService := &MockRequirementService{}
	handler := NewSearchHandler(mockSearchService, mockRequirementService)

	mockRequirementService.On("SearchRequirements", "test").Return([]models.Requirement{}, errors.New("requirement service error"))

	args := map[string]interface{}{
		"query": "test",
	}

	_, err := handler.Requirements(context.Background(), args)
	assert.Error(t, err)

	mockRequirementService.AssertExpectations(t)
}

// TestSearchHandler_Global_WithEntityTypeFiltering tests global search with entity type filtering
func TestSearchHandler_Global_WithEntityTypeFiltering(t *testing.T) {
	mockSearchService := &MockSearchServiceForSearchHandler{}
	mockRequirementService := &MockRequirementService{}
	handler := NewSearchHandler(mockSearchService, mockRequirementService)

	// Test with specific entity types
	expectedResponse := &service.SearchResponse{
		Results: []service.SearchResult{
			{
				ID:          uuid.New(),
				ReferenceID: "EP-001",
				Type:        "epic",
				Title:       "Test Epic",
				Description: "Test Description",
			},
			{
				ID:          uuid.New(),
				ReferenceID: "US-001",
				Type:        "user_story",
				Title:       "Test User Story",
				Description: "Test Description",
			},
		},
		Total:  2,
		Limit:  50,
		Offset: 0,
		Query:  "test",
	}

	mockSearchService.On("Search", mock.Anything, mock.MatchedBy(func(opts service.SearchOptions) bool {
		return opts.Query == "test" &&
			len(opts.EntityTypes) == 2 &&
			opts.EntityTypes[0] == "epic" &&
			opts.EntityTypes[1] == "user_story"
	})).Return(expectedResponse, nil)

	args := map[string]interface{}{
		"query":        "test",
		"entity_types": []interface{}{"epic", "user_story"},
	}

	result, err := handler.Global(context.Background(), args)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	mockSearchService.AssertExpectations(t)
}

// TestSearchHandler_Global_WithPagination tests global search with pagination parameters
func TestSearchHandler_Global_WithPagination(t *testing.T) {
	mockSearchService := &MockSearchServiceForSearchHandler{}
	mockRequirementService := &MockRequirementService{}
	handler := NewSearchHandler(mockSearchService, mockRequirementService)

	expectedResponse := &service.SearchResponse{
		Results: []service.SearchResult{
			{
				ID:          uuid.New(),
				ReferenceID: "REQ-001",
				Type:        "requirement",
				Title:       "Test Requirement",
				Description: "Test Description",
			},
		},
		Total:  100,
		Limit:  10,
		Offset: 20,
		Query:  "pagination test",
	}

	mockSearchService.On("Search", mock.Anything, mock.MatchedBy(func(opts service.SearchOptions) bool {
		return opts.Query == "pagination test" &&
			opts.Limit == 10 &&
			opts.Offset == 20
	})).Return(expectedResponse, nil)

	args := map[string]interface{}{
		"query":  "pagination test",
		"limit":  float64(10),
		"offset": float64(20),
	}

	result, err := handler.Global(context.Background(), args)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	mockSearchService.AssertExpectations(t)
}

// TestSearchHandler_Global_ResultFormatting tests that search results are properly formatted
func TestSearchHandler_Global_ResultFormatting(t *testing.T) {
	mockSearchService := &MockSearchServiceForSearchHandler{}
	mockRequirementService := &MockRequirementService{}
	handler := NewSearchHandler(mockSearchService, mockRequirementService)

	expectedResponse := &service.SearchResponse{
		Results: []service.SearchResult{
			{
				ID:          uuid.New(),
				ReferenceID: "EP-001",
				Type:        "epic",
				Title:       "Test Epic",
				Description: "Test Description",
			},
		},
		Total:  1,
		Limit:  50,
		Offset: 0,
		Query:  "format test",
	}

	mockSearchService.On("Search", mock.Anything, mock.Anything).Return(expectedResponse, nil)

	args := map[string]interface{}{
		"query": "format test",
	}

	result, err := handler.Global(context.Background(), args)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Verify the result is properly formatted as ToolResponse
	toolResponse, ok := result.(*types.ToolResponse)
	assert.True(t, ok, "Result should be a ToolResponse")
	assert.Len(t, toolResponse.Content, 2, "Should have 2 content items")

	// Check first content item (summary message)
	assert.Equal(t, "text", toolResponse.Content[0].Type)
	assert.Contains(t, toolResponse.Content[0].Text, "Found 1 results for query 'format test'")

	// Check second content item (JSON data)
	assert.Equal(t, "text", toolResponse.Content[1].Type)
	assert.Contains(t, toolResponse.Content[1].Text, "\"query\": \"format test\"")
	assert.Contains(t, toolResponse.Content[1].Text, "\"total_count\": 1")

	mockSearchService.AssertExpectations(t)
}

// TestSearchHandler_Requirements_ResultFormatting tests that requirements search results are properly formatted
func TestSearchHandler_Requirements_ResultFormatting(t *testing.T) {
	mockSearchService := &MockSearchServiceForSearchHandler{}
	mockRequirementService := &MockRequirementService{}
	handler := NewSearchHandler(mockSearchService, mockRequirementService)

	desc := "Test Requirement Description"
	expectedRequirements := []models.Requirement{
		{
			ID:          uuid.New(),
			ReferenceID: "REQ-001",
			Title:       "Test Requirement 1",
			Description: &desc,
		},
		{
			ID:          uuid.New(),
			ReferenceID: "REQ-002",
			Title:       "Test Requirement 2",
			Description: &desc,
		},
	}

	mockRequirementService.On("SearchRequirements", "requirements test").Return(expectedRequirements, nil)

	args := map[string]interface{}{
		"query": "requirements test",
	}

	result, err := handler.Requirements(context.Background(), args)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Verify the result is properly formatted as ToolResponse
	toolResponse, ok := result.(*types.ToolResponse)
	assert.True(t, ok, "Result should be a ToolResponse")
	assert.Len(t, toolResponse.Content, 2, "Should have 2 content items")

	// Check first content item (summary message)
	assert.Equal(t, "text", toolResponse.Content[0].Type)
	assert.Contains(t, toolResponse.Content[0].Text, "Found 2 requirements matching query 'requirements test'")

	// Check second content item (JSON data)
	assert.Equal(t, "text", toolResponse.Content[1].Type)
	assert.Contains(t, toolResponse.Content[1].Text, "\"query\": \"requirements test\"")
	assert.Contains(t, toolResponse.Content[1].Text, "\"count\": 2")
	assert.Contains(t, toolResponse.Content[1].Text, "REQ-001")
	assert.Contains(t, toolResponse.Content[1].Text, "REQ-002")

	mockRequirementService.AssertExpectations(t)
}

// TestSearchHandler_Global_EmptyEntityTypes tests global search with empty entity types array
func TestSearchHandler_Global_EmptyEntityTypes(t *testing.T) {
	mockSearchService := &MockSearchServiceForSearchHandler{}
	mockRequirementService := &MockRequirementService{}
	handler := NewSearchHandler(mockSearchService, mockRequirementService)

	expectedResponse := &service.SearchResponse{
		Results: []service.SearchResult{},
		Total:   0,
		Limit:   50,
		Offset:  0,
		Query:   "empty test",
	}

	mockSearchService.On("Search", mock.Anything, mock.MatchedBy(func(opts service.SearchOptions) bool {
		return opts.Query == "empty test" && len(opts.EntityTypes) == 0
	})).Return(expectedResponse, nil)

	args := map[string]interface{}{
		"query":        "empty test",
		"entity_types": []interface{}{},
	}

	result, err := handler.Global(context.Background(), args)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	mockSearchService.AssertExpectations(t)
}

// TestSearchHandler_Requirements_EmptyResults tests requirements search with no results
func TestSearchHandler_Requirements_EmptyResults(t *testing.T) {
	mockSearchService := &MockSearchServiceForSearchHandler{}
	mockRequirementService := &MockRequirementService{}
	handler := NewSearchHandler(mockSearchService, mockRequirementService)

	mockRequirementService.On("SearchRequirements", "no results").Return([]models.Requirement{}, nil)

	args := map[string]interface{}{
		"query": "no results",
	}

	result, err := handler.Requirements(context.Background(), args)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Verify the result shows 0 results
	toolResponse, ok := result.(*types.ToolResponse)
	assert.True(t, ok, "Result should be a ToolResponse")
	assert.Contains(t, toolResponse.Content[0].Text, "Found 0 requirements matching query 'no results'")

	mockRequirementService.AssertExpectations(t)
}

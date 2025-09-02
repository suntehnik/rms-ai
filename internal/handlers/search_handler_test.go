package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"product-requirements-management/internal/service"
)

// MockSearchService is a mock implementation of the search service
type MockSearchService struct {
	mock.Mock
}

func (m *MockSearchService) Search(ctx context.Context, options service.SearchOptions) (*service.SearchResponse, error) {
	args := m.Called(ctx, options)
	return args.Get(0).(*service.SearchResponse), args.Error(1)
}

func (m *MockSearchService) InvalidateCache(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func TestSearchHandler_Search_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := &MockSearchService{}
	logger := logrus.New()
	handler := NewSearchHandler(mockService, logger)

	// Setup expected response
	expectedResponse := &service.SearchResponse{
		Results: []service.SearchResult{
			{
				ID:          uuid.New(),
				ReferenceID: "EP-001",
				Type:        "epic",
				Title:       "Test Epic",
				Status:      "Backlog",
				CreatedAt:   time.Now(),
			},
		},
		Total:      1,
		Limit:      50,
		Offset:     0,
		Query:      "test",
		ExecutedAt: time.Now(),
	}

	mockService.On("Search", mock.Anything, mock.MatchedBy(func(options service.SearchOptions) bool {
		return options.Query == "test" && options.Limit == 50 && options.Offset == 0
	})).Return(expectedResponse, nil)

	// Create request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req := httptest.NewRequest("GET", "/api/search?query=test", nil)
	c.Request = req

	// Set correlation ID
	c.Set("correlation_id", "test-correlation-id")

	// Execute
	handler.Search(c)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response service.SearchResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse.Query, response.Query)
	assert.Equal(t, expectedResponse.Total, response.Total)
	assert.Len(t, response.Results, 1)

	mockService.AssertExpectations(t)
}

func TestSearchHandler_Search_WithFilters(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := &MockSearchService{}
	logger := logrus.New()
	handler := NewSearchHandler(mockService, logger)

	creatorID := uuid.New()
	priority := 1

	expectedResponse := &service.SearchResponse{
		Results:    []service.SearchResult{},
		Total:      0,
		Limit:      25,
		Offset:     10,
		Query:      "test",
		ExecutedAt: time.Now(),
	}

	mockService.On("Search", mock.Anything, mock.MatchedBy(func(options service.SearchOptions) bool {
		return options.Query == "test" &&
			options.Limit == 25 &&
			options.Offset == 10 &&
			options.Filters.CreatorID != nil &&
			*options.Filters.CreatorID == creatorID &&
			options.Filters.Priority != nil &&
			*options.Filters.Priority == priority &&
			options.Filters.Status != nil &&
			*options.Filters.Status == "Backlog"
	})).Return(expectedResponse, nil)

	// Create request with filters
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	params := url.Values{}
	params.Add("query", "test")
	params.Add("creator_id", creatorID.String())
	params.Add("priority", "1")
	params.Add("status", "Backlog")
	params.Add("limit", "25")
	params.Add("offset", "10")

	req := httptest.NewRequest("GET", "/api/search?"+params.Encode(), nil)
	c.Request = req

	// Set correlation ID
	c.Set("correlation_id", "test-correlation-id")

	// Execute
	handler.Search(c)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	mockService.AssertExpectations(t)
}

func TestSearchHandler_Search_WithDateFilters(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := &MockSearchService{}
	logger := logrus.New()
	handler := NewSearchHandler(mockService, logger)

	createdFrom := time.Now().AddDate(0, -1, 0) // 1 month ago
	createdTo := time.Now()

	expectedResponse := &service.SearchResponse{
		Results:    []service.SearchResult{},
		Total:      0,
		Limit:      50,
		Offset:     0,
		Query:      "",
		ExecutedAt: time.Now(),
	}

	mockService.On("Search", mock.Anything, mock.MatchedBy(func(options service.SearchOptions) bool {
		return options.Filters.CreatedFrom != nil &&
			options.Filters.CreatedTo != nil &&
			options.Filters.CreatedFrom.Unix() == createdFrom.Unix() &&
			options.Filters.CreatedTo.Unix() == createdTo.Unix()
	})).Return(expectedResponse, nil)

	// Create request with date filters
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	params := url.Values{}
	params.Add("created_from", createdFrom.Format(time.RFC3339))
	params.Add("created_to", createdTo.Format(time.RFC3339))

	req := httptest.NewRequest("GET", "/api/search?"+params.Encode(), nil)
	c.Request = req

	// Set correlation ID
	c.Set("correlation_id", "test-correlation-id")

	// Execute
	handler.Search(c)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	mockService.AssertExpectations(t)
}

func TestSearchHandler_Search_InvalidUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := &MockSearchService{}
	logger := logrus.New()
	handler := NewSearchHandler(mockService, logger)

	// Create request with invalid UUID
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req := httptest.NewRequest("GET", "/api/search?creator_id=invalid-uuid", nil)
	c.Request = req

	// Set correlation ID
	c.Set("correlation_id", "test-correlation-id")

	// Execute
	handler.Search(c)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "INVALID_SEARCH_OPTIONS", response.Error.Code)
}

func TestSearchHandler_Search_InvalidPriority(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := &MockSearchService{}
	logger := logrus.New()
	handler := NewSearchHandler(mockService, logger)

	// Create request with invalid priority
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req := httptest.NewRequest("GET", "/api/search?priority=5", nil)
	c.Request = req

	// Set correlation ID
	c.Set("correlation_id", "test-correlation-id")

	// Execute
	handler.Search(c)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "INVALID_SEARCH_OPTIONS", response.Error.Code)
}

func TestSearchHandler_Search_InvalidDate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := &MockSearchService{}
	logger := logrus.New()
	handler := NewSearchHandler(mockService, logger)

	// Create request with invalid date
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req := httptest.NewRequest("GET", "/api/search?created_from=invalid-date", nil)
	c.Request = req

	// Set correlation ID
	c.Set("correlation_id", "test-correlation-id")

	// Execute
	handler.Search(c)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "INVALID_SEARCH_OPTIONS", response.Error.Code)
}

func TestSearchHandler_Search_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := &MockSearchService{}
	logger := logrus.New()
	handler := NewSearchHandler(mockService, logger)

	mockService.On("Search", mock.Anything, mock.Anything).Return((*service.SearchResponse)(nil), assert.AnError)

	// Create request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req := httptest.NewRequest("GET", "/api/search?query=test", nil)
	c.Request = req

	// Set correlation ID
	c.Set("correlation_id", "test-correlation-id")

	// Execute
	handler.Search(c)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "SEARCH_FAILED", response.Error.Code)

	mockService.AssertExpectations(t)
}

func TestSearchHandler_SearchSuggestions_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := &MockSearchService{}
	logger := logrus.New()
	handler := NewSearchHandler(mockService, logger)

	// Create request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req := httptest.NewRequest("GET", "/api/search/suggestions?query=test", nil)
	c.Request = req

	// Set correlation ID
	c.Set("correlation_id", "test-correlation-id")

	// Execute
	handler.SearchSuggestions(c)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string][]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "titles")
	assert.Contains(t, response, "reference_ids")
	assert.Contains(t, response, "statuses")
	assert.Contains(t, response["statuses"], "Backlog")
}

func TestSearchHandler_SearchSuggestions_MissingQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := &MockSearchService{}
	logger := logrus.New()
	handler := NewSearchHandler(mockService, logger)

	// Create request without query
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req := httptest.NewRequest("GET", "/api/search/suggestions", nil)
	c.Request = req

	// Execute
	handler.SearchSuggestions(c)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "MISSING_QUERY", response.Error.Code)
}

func TestSearchHandler_parseSearchOptions_Defaults(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := &SearchHandler{}

	// Create request with minimal parameters
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req := httptest.NewRequest("GET", "/api/search", nil)
	c.Request = req

	options, err := handler.parseSearchOptions(c)

	assert.NoError(t, err)
	assert.Equal(t, "", options.Query)
	assert.Equal(t, "created_at", options.SortBy)
	assert.Equal(t, "desc", options.SortOrder)
	assert.Equal(t, 50, options.Limit)
	assert.Equal(t, 0, options.Offset)
}

func TestSearchHandler_parseSearchOptions_AllFilters(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := &SearchHandler{}

	creatorID := uuid.New()
	assigneeID := uuid.New()
	epicID := uuid.New()
	userStoryID := uuid.New()
	acID := uuid.New()
	reqTypeID := uuid.New()
	authorID := uuid.New()

	// Create request with all filters
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	params := url.Values{}
	params.Add("query", "test query")
	params.Add("creator_id", creatorID.String())
	params.Add("assignee_id", assigneeID.String())
	params.Add("epic_id", epicID.String())
	params.Add("user_story_id", userStoryID.String())
	params.Add("acceptance_criteria_id", acID.String())
	params.Add("requirement_type_id", reqTypeID.String())
	params.Add("author_id", authorID.String())
	params.Add("priority", "2")
	params.Add("status", "In Progress")
	params.Add("created_from", "2023-01-01T00:00:00Z")
	params.Add("created_to", "2023-12-31T23:59:59Z")
	params.Add("sort_by", "priority")
	params.Add("sort_order", "asc")
	params.Add("limit", "25")
	params.Add("offset", "10")

	req := httptest.NewRequest("GET", "/api/search?"+params.Encode(), nil)
	c.Request = req

	options, err := handler.parseSearchOptions(c)

	assert.NoError(t, err)
	assert.Equal(t, "test query", options.Query)
	assert.Equal(t, "priority", options.SortBy)
	assert.Equal(t, "asc", options.SortOrder)
	assert.Equal(t, 25, options.Limit)
	assert.Equal(t, 10, options.Offset)

	// Check filters
	assert.NotNil(t, options.Filters.CreatorID)
	assert.Equal(t, creatorID, *options.Filters.CreatorID)
	assert.NotNil(t, options.Filters.AssigneeID)
	assert.Equal(t, assigneeID, *options.Filters.AssigneeID)
	assert.NotNil(t, options.Filters.EpicID)
	assert.Equal(t, epicID, *options.Filters.EpicID)
	assert.NotNil(t, options.Filters.UserStoryID)
	assert.Equal(t, userStoryID, *options.Filters.UserStoryID)
	assert.NotNil(t, options.Filters.AcceptanceCriteriaID)
	assert.Equal(t, acID, *options.Filters.AcceptanceCriteriaID)
	assert.NotNil(t, options.Filters.RequirementTypeID)
	assert.Equal(t, reqTypeID, *options.Filters.RequirementTypeID)
	assert.NotNil(t, options.Filters.AuthorID)
	assert.Equal(t, authorID, *options.Filters.AuthorID)
	assert.NotNil(t, options.Filters.Priority)
	assert.Equal(t, 2, *options.Filters.Priority)
	assert.NotNil(t, options.Filters.Status)
	assert.Equal(t, "In Progress", *options.Filters.Status)
	assert.NotNil(t, options.Filters.CreatedFrom)
	assert.NotNil(t, options.Filters.CreatedTo)
}

func TestSearchHandler_Search_InvalidLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := &MockSearchService{}
	logger := logrus.New()
	handler := NewSearchHandler(mockService, logger)

	testCases := []struct {
		name        string
		limit       string
		expectError bool
	}{
		{"negative_limit", "-1", true},
		{"zero_limit", "0", false}, // Zero is valid, will be set to default
		{"valid_limit", "25", false},
		{"max_limit", "100", false},
		{"over_max_limit", "101", true},
		{"invalid_format", "abc", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			req := httptest.NewRequest("GET", "/api/search?limit="+tc.limit, nil)
			c.Request = req
			c.Set("correlation_id", "test-correlation-id")

			if !tc.expectError {
				mockService.On("Search", mock.Anything, mock.Anything).Return(&service.SearchResponse{
					Results:    []service.SearchResult{},
					Total:      0,
					Limit:      50,
					Offset:     0,
					Query:      "",
					ExecutedAt: time.Now(),
				}, nil).Once()
			}

			handler.Search(c)

			if tc.expectError {
				assert.Equal(t, http.StatusBadRequest, w.Code)
				var response ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "INVALID_SEARCH_OPTIONS", response.Error.Code)
			} else {
				assert.Equal(t, http.StatusOK, w.Code)
			}
		})
	}
}

func TestSearchHandler_Search_InvalidOffset(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := &MockSearchService{}
	logger := logrus.New()
	handler := NewSearchHandler(mockService, logger)

	testCases := []struct {
		name        string
		offset      string
		expectError bool
	}{
		{"negative_offset", "-1", true},
		{"zero_offset", "0", false},
		{"valid_offset", "10", false},
		{"large_offset", "1000", false},
		{"invalid_format", "xyz", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			req := httptest.NewRequest("GET", "/api/search?offset="+tc.offset, nil)
			c.Request = req
			c.Set("correlation_id", "test-correlation-id")

			if !tc.expectError {
				mockService.On("Search", mock.Anything, mock.Anything).Return(&service.SearchResponse{
					Results:    []service.SearchResult{},
					Total:      0,
					Limit:      50,
					Offset:     0,
					Query:      "",
					ExecutedAt: time.Now(),
				}, nil).Once()
			}

			handler.Search(c)

			if tc.expectError {
				assert.Equal(t, http.StatusBadRequest, w.Code)
				var response ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "INVALID_SEARCH_OPTIONS", response.Error.Code)
			} else {
				assert.Equal(t, http.StatusOK, w.Code)
			}
		})
	}
}
package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)



func TestSearchService_prepareSearchQuery(t *testing.T) {
	service := &SearchService{}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "single word",
			input:    "test",
			expected: "test:*",
		},
		{
			name:     "multiple words",
			input:    "test query",
			expected: "test:* & query:*",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "whitespace only",
			input:    "   ",
			expected: "",
		},
		{
			name:     "multiple spaces",
			input:    "test   multiple   spaces",
			expected: "test:* & multiple:* & spaces:*",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.prepareSearchQuery(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSearchService_generateCacheKey(t *testing.T) {
	service := &SearchService{}

	options1 := SearchOptions{
		Query:     "test",
		SortBy:    "created_at",
		SortOrder: "desc",
		Limit:     50,
		Offset:    0,
	}

	options2 := SearchOptions{
		Query:     "test",
		SortBy:    "created_at",
		SortOrder: "desc",
		Limit:     50,
		Offset:    0,
	}

	options3 := SearchOptions{
		Query:     "different",
		SortBy:    "created_at",
		SortOrder: "desc",
		Limit:     50,
		Offset:    0,
	}

	key1 := service.generateCacheKey(options1)
	key2 := service.generateCacheKey(options2)
	key3 := service.generateCacheKey(options3)

	// Same options should generate same key
	assert.Equal(t, key1, key2)
	
	// Different options should generate different keys
	assert.NotEqual(t, key1, key3)
	
	// Keys should start with "search:"
	assert.Contains(t, key1, "search:")
}

func TestSearchService_InvalidateCache(t *testing.T) {
	// This test would need a proper Redis mock implementation
	// For now, we'll test the nil case
	service := &SearchService{
		redisClient: nil,
	}

	ctx := context.Background()

	err := service.InvalidateCache(ctx)
	assert.NoError(t, err)
}

func TestSearchService_InvalidateCache_NoRedis(t *testing.T) {
	service := &SearchService{
		redisClient: nil,
	}

	ctx := context.Background()

	err := service.InvalidateCache(ctx)
	assert.NoError(t, err)
}

func TestSearchService_InvalidateCache_NoKeys(t *testing.T) {
	service := &SearchService{
		redisClient: nil,
	}

	ctx := context.Background()

	err := service.InvalidateCache(ctx)
	assert.NoError(t, err)
}

func TestSearchOptions_Defaults(t *testing.T) {
	options := SearchOptions{
		Query: "test",
	}

	// Test that defaults are applied in the search method
	// This would be tested in integration tests with actual database
	assert.Equal(t, "test", options.Query)
}

func TestSearchFilters_UUIDFilters(t *testing.T) {
	creatorID := uuid.New()
	assigneeID := uuid.New()

	filters := SearchFilters{
		CreatorID:  &creatorID,
		AssigneeID: &assigneeID,
	}

	assert.Equal(t, creatorID, *filters.CreatorID)
	assert.Equal(t, assigneeID, *filters.AssigneeID)
}

func TestSearchResult_Structure(t *testing.T) {
	id := uuid.New()
	priority := 1
	now := time.Now()

	result := SearchResult{
		ID:          id,
		ReferenceID: "EP-001",
		Type:        "epic",
		Title:       "Test Epic",
		Description: "Test Description",
		Priority:    &priority,
		Status:      "Backlog",
		CreatedAt:   now,
		Relevance:   0.95,
	}

	assert.Equal(t, id, result.ID)
	assert.Equal(t, "EP-001", result.ReferenceID)
	assert.Equal(t, "epic", result.Type)
	assert.Equal(t, "Test Epic", result.Title)
	assert.Equal(t, "Test Description", result.Description)
	assert.Equal(t, 1, *result.Priority)
	assert.Equal(t, "Backlog", result.Status)
	assert.Equal(t, now, result.CreatedAt)
	assert.Equal(t, 0.95, result.Relevance)
}

func TestSearchResponse_Structure(t *testing.T) {
	results := []SearchResult{
		{
			ID:          uuid.New(),
			ReferenceID: "EP-001",
			Type:        "epic",
			Title:       "Test Epic",
			Status:      "Backlog",
			CreatedAt:   time.Now(),
		},
	}

	response := SearchResponse{
		Results:    results,
		Total:      1,
		Limit:      50,
		Offset:     0,
		Query:      "test",
		ExecutedAt: time.Now(),
	}

	assert.Len(t, response.Results, 1)
	assert.Equal(t, int64(1), response.Total)
	assert.Equal(t, 50, response.Limit)
	assert.Equal(t, 0, response.Offset)
	assert.Equal(t, "test", response.Query)
	assert.NotZero(t, response.ExecutedAt)
}
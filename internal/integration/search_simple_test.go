package integration

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"product-requirements-management/internal/repository"
	"product-requirements-management/internal/service"
)

func TestSearchIntegration_BasicFunctionality(t *testing.T) {
	// Setup test environment with PostgreSQL
	testDB := SetupTestDatabase(t)
	defer testDB.Cleanup(t)

	// Setup repositories
	repos := repository.NewRepositories(testDB.DB)

	// Setup search service
	searchService := service.NewSearchService(
		testDB.DB,
		nil, // No Redis for integration tests
		repos.Epic,
		repos.UserStory,
		repos.AcceptanceCriteria,
		repos.Requirement,
	)

	// Test basic search functionality
	options := service.SearchOptions{
		Query:     "",
		Limit:     50,
		Offset:    0,
		SortBy:    "created_at",
		SortOrder: "desc",
	}

	response, err := searchService.Search(context.Background(), options)
	require.NoError(t, err)
	assert.NotNil(t, response)

	// Should return empty results for empty database
	assert.Equal(t, int64(0), response.Total)
	assert.Equal(t, 0, len(response.Results))
	assert.Equal(t, "", response.Query)
	assert.Equal(t, 50, response.Limit)
	assert.Equal(t, 0, response.Offset)
	assert.NotZero(t, response.ExecutedAt)
}

func TestSearchIntegration_CacheInvalidation(t *testing.T) {
	// Setup test environment with PostgreSQL
	testDB := SetupTestDatabase(t)
	defer testDB.Cleanup(t)

	// Setup repositories
	repos := repository.NewRepositories(testDB.DB)

	// Setup search service without Redis
	searchService := service.NewSearchService(
		testDB.DB,
		nil, // No Redis
		repos.Epic,
		repos.UserStory,
		repos.AcceptanceCriteria,
		repos.Requirement,
	)

	// Test cache invalidation without Redis (should not fail)
	err := searchService.InvalidateCache(context.Background())
	assert.NoError(t, err)
}

func TestSearchIntegration_PrepareSearchQuery(t *testing.T) {
	// Setup test environment with PostgreSQL
	testDB := SetupTestDatabase(t)
	defer testDB.Cleanup(t)

	// Setup repositories
	repos := repository.NewRepositories(testDB.DB)

	// Setup search service
	searchService := service.NewSearchService(
		testDB.DB,
		nil,
		repos.Epic,
		repos.UserStory,
		repos.AcceptanceCriteria,
		repos.Requirement,
	)

	// Test filter-only search (no full-text search with SQLite)
	options := service.SearchOptions{
		Query:     "", // Empty query to avoid full-text search
		Limit:     10,
		Offset:    0,
		SortBy:    "created_at",
		SortOrder: "desc",
	}

	response, err := searchService.Search(context.Background(), options)
	require.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "", response.Query)
}

func TestSearchIntegration_FilterValidation(t *testing.T) {
	// Setup test environment with PostgreSQL
	testDB := SetupTestDatabase(t)
	defer testDB.Cleanup(t)

	// Setup repositories
	repos := repository.NewRepositories(testDB.DB)

	// Setup search service
	searchService := service.NewSearchService(
		testDB.DB,
		nil,
		repos.Epic,
		repos.UserStory,
		repos.AcceptanceCriteria,
		repos.Requirement,
	)

	// Test with various filter combinations
	priority := 1
	status := "Backlog"

	options := service.SearchOptions{
		Filters: service.SearchFilters{
			Priority: &priority,
			Status:   &status,
		},
		Limit:     25,
		Offset:    0,
		SortBy:    "priority",
		SortOrder: "asc",
	}

	response, err := searchService.Search(context.Background(), options)
	require.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, 25, response.Limit)
	assert.Equal(t, 0, response.Offset)
}

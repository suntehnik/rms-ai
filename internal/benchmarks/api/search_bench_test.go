package api

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"product-requirements-management/internal/benchmarks/helpers"
	"product-requirements-management/internal/benchmarks/setup"
	"product-requirements-management/internal/models"
)

// BenchmarkSearchFullText tests full-text search performance via search API endpoints
func BenchmarkSearchFullText(b *testing.B) {
	// Setup benchmark server
	server := setup.NewBenchmarkServer(b)
	defer server.Cleanup()

	// Start the server
	require.NoError(b, server.Start())

	// Seed medium dataset for search operations
	require.NoError(b, server.SeedMediumDataSet())

	// Create HTTP client
	client := helpers.NewBenchmarkClient(server.BaseURL)

	// Setup authentication
	authHelper := helpers.NewAuthHelper(server.Config.JWT.Secret)
	testUser := helpers.GetDefaultTestUser()
	require.NoError(b, authHelper.AuthenticateClient(client, testUser.ID, testUser.Username))

	b.ResetTimer()

	b.Run("SimpleKeywordSearch", func(b *testing.B) {
		searchTerms := []string{
			"epic",
			"user",
			"story",
			"requirement",
			"acceptance",
			"criteria",
			"test",
			"feature",
			"system",
			"performance",
		}

		for i := 0; i < b.N; i++ {
			searchTerm := searchTerms[i%len(searchTerms)]
			resp, err := client.GET(fmt.Sprintf("/api/v1/search?query=%s", url.QueryEscape(searchTerm)))
			require.NoError(b, err)
			require.Equal(b, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		}
	})

	b.Run("MultiWordSearch", func(b *testing.B) {
		searchQueries := []string{
			"user story epic",
			"acceptance criteria test",
			"requirement system performance",
			"epic feature development",
			"story acceptance requirement",
			"test system validation",
			"performance benchmark testing",
			"feature epic user story",
			"criteria acceptance validation",
			"system requirement analysis",
		}

		for i := 0; i < b.N; i++ {
			query := searchQueries[i%len(searchQueries)]
			resp, err := client.GET(fmt.Sprintf("/api/v1/search?query=%s", url.QueryEscape(query)))
			require.NoError(b, err)
			require.Equal(b, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		}
	})

	b.Run("ReferenceIDSearch", func(b *testing.B) {
		// Get some reference IDs to search for
		var epics []models.Epic
		require.NoError(b, server.DB.Limit(5).Find(&epics).Error)
		require.NotEmpty(b, epics)

		var userStories []models.UserStory
		require.NoError(b, server.DB.Limit(5).Find(&userStories).Error)
		require.NotEmpty(b, userStories)

		referenceIDs := make([]string, 0, len(epics)+len(userStories))
		for _, epic := range epics {
			referenceIDs = append(referenceIDs, epic.ReferenceID)
		}
		for _, us := range userStories {
			referenceIDs = append(referenceIDs, us.ReferenceID)
		}

		for i := 0; i < b.N; i++ {
			refID := referenceIDs[i%len(referenceIDs)]
			resp, err := client.GET(fmt.Sprintf("/api/v1/search?query=%s", url.QueryEscape(refID)))
			require.NoError(b, err)
			require.Equal(b, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		}
	})

	b.Run("PartialTextSearch", func(b *testing.B) {
		partialTerms := []string{
			"epi",
			"use",
			"sto",
			"req",
			"acc",
			"cri",
			"tes",
			"fea",
			"sys",
			"per",
		}

		for i := 0; i < b.N; i++ {
			term := partialTerms[i%len(partialTerms)]
			resp, err := client.GET(fmt.Sprintf("/api/v1/search?query=%s", url.QueryEscape(term)))
			require.NoError(b, err)
			require.Equal(b, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		}
	})
}

// BenchmarkSearchFiltering tests search filtering and pagination performance
func BenchmarkSearchFiltering(b *testing.B) {
	// Setup benchmark server
	server := setup.NewBenchmarkServer(b)
	defer server.Cleanup()

	// Start the server
	require.NoError(b, server.Start())

	// Seed medium dataset for filtering operations
	require.NoError(b, server.SeedMediumDataSet())

	// Create HTTP client
	client := helpers.NewBenchmarkClient(server.BaseURL)

	// Setup authentication
	authHelper := helpers.NewAuthHelper(server.Config.JWT.Secret)
	testUser := helpers.GetDefaultTestUser()
	require.NoError(b, authHelper.AuthenticateClient(client, testUser.ID, testUser.Username))

	// Get test data for filtering
	var users []models.User
	require.NoError(b, server.DB.Limit(3).Find(&users).Error)
	require.NotEmpty(b, users)

	var epics []models.Epic
	require.NoError(b, server.DB.Limit(3).Find(&epics).Error)
	require.NotEmpty(b, epics)

	b.ResetTimer()

	b.Run("FilterByStatus", func(b *testing.B) {
		statuses := []string{
			"Backlog",
			"Draft",
			"In Progress",
			"Done",
			"Active",
		}

		for i := 0; i < b.N; i++ {
			status := statuses[i%len(statuses)]
			resp, err := client.GET(fmt.Sprintf("/api/v1/search?status=%s", url.QueryEscape(status)))
			require.NoError(b, err)
			require.Equal(b, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		}
	})

	b.Run("FilterByPriority", func(b *testing.B) {
		priorities := []int{1, 2, 3, 4}

		for i := 0; i < b.N; i++ {
			priority := priorities[i%len(priorities)]
			resp, err := client.GET(fmt.Sprintf("/api/v1/search?priority=%d", priority))
			require.NoError(b, err)
			require.Equal(b, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		}
	})

	b.Run("FilterByCreator", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			user := users[i%len(users)]
			resp, err := client.GET(fmt.Sprintf("/api/v1/search?creator_id=%s", user.ID))
			require.NoError(b, err)
			require.Equal(b, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		}
	})

	b.Run("FilterByEpic", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			epic := epics[i%len(epics)]
			resp, err := client.GET(fmt.Sprintf("/api/v1/search?epic_id=%s", epic.ID))
			require.NoError(b, err)
			require.Equal(b, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		}
	})

	b.Run("FilterByDateRange", func(b *testing.B) {
		// Create date ranges for filtering
		now := time.Now()
		dateRanges := []struct {
			from, to string
		}{
			{
				from: now.AddDate(0, 0, -30).Format(time.RFC3339),
				to:   now.Format(time.RFC3339),
			},
			{
				from: now.AddDate(0, 0, -7).Format(time.RFC3339),
				to:   now.Format(time.RFC3339),
			},
			{
				from: now.AddDate(0, 0, -1).Format(time.RFC3339),
				to:   now.Format(time.RFC3339),
			},
		}

		for i := 0; i < b.N; i++ {
			dateRange := dateRanges[i%len(dateRanges)]
			resp, err := client.GET(fmt.Sprintf("/api/v1/search?created_from=%s&created_to=%s",
				url.QueryEscape(dateRange.from), url.QueryEscape(dateRange.to)))
			require.NoError(b, err)
			require.Equal(b, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		}
	})

	b.Run("CombinedFilters", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			user := users[i%len(users)]
			priority := (i % 4) + 1
			status := []string{"Backlog", "Draft", "In Progress", "Done"}[i%4]

			resp, err := client.GET(fmt.Sprintf("/api/v1/search?creator_id=%s&priority=%d&status=%s",
				user.ID, priority, url.QueryEscape(status)))
			require.NoError(b, err)
			require.Equal(b, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		}
	})
}

// BenchmarkSearchPagination tests search pagination performance
func BenchmarkSearchPagination(b *testing.B) {
	// Setup benchmark server
	server := setup.NewBenchmarkServer(b)
	defer server.Cleanup()

	// Start the server
	require.NoError(b, server.Start())

	// Seed medium dataset for pagination operations
	require.NoError(b, server.SeedMediumDataSet())

	// Create HTTP client
	client := helpers.NewBenchmarkClient(server.BaseURL)

	// Setup authentication
	authHelper := helpers.NewAuthHelper(server.Config.JWT.Secret)
	testUser := helpers.GetDefaultTestUser()
	require.NoError(b, authHelper.AuthenticateClient(client, testUser.ID, testUser.Username))

	b.ResetTimer()

	b.Run("PaginationWithLimit", func(b *testing.B) {
		limits := []int{5, 10, 20, 50}

		for i := 0; i < b.N; i++ {
			limit := limits[i%len(limits)]
			resp, err := client.GET(fmt.Sprintf("/api/v1/search?limit=%d", limit))
			require.NoError(b, err)
			require.Equal(b, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		}
	})

	b.Run("PaginationWithOffset", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			offset := (i % 10) * 5 // Vary offset for different pages
			resp, err := client.GET(fmt.Sprintf("/api/v1/search?limit=10&offset=%d", offset))
			require.NoError(b, err)
			require.Equal(b, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		}
	})

	b.Run("PaginationWithSorting", func(b *testing.B) {
		sortOptions := []struct {
			sortBy    string
			sortOrder string
		}{
			{"created_at", "desc"},
			{"created_at", "asc"},
			{"title", "asc"},
			{"title", "desc"},
			{"priority", "desc"},
			{"priority", "asc"},
		}

		for i := 0; i < b.N; i++ {
			sort := sortOptions[i%len(sortOptions)]
			offset := (i % 5) * 10
			resp, err := client.GET(fmt.Sprintf("/api/v1/search?limit=10&offset=%d&sort_by=%s&sort_order=%s",
				offset, sort.sortBy, sort.sortOrder))
			require.NoError(b, err)
			require.Equal(b, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		}
	})

	b.Run("LargePageSizes", func(b *testing.B) {
		pageSizes := []int{50, 75, 100}

		for i := 0; i < b.N; i++ {
			pageSize := pageSizes[i%len(pageSizes)]
			resp, err := client.GET(fmt.Sprintf("/api/v1/search?limit=%d", pageSize))
			require.NoError(b, err)
			require.Equal(b, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		}
	})
}

// BenchmarkSearchDatasetSizes tests search performance with varying dataset sizes
func BenchmarkSearchDatasetSizes(b *testing.B) {
	// Test with small dataset (100 records)
	b.Run("SmallDataset_100Records", func(b *testing.B) {
		benchmarkSearchWithDataset(b, "small")
	})

	// Test with medium dataset (1000 records)
	b.Run("MediumDataset_1000Records", func(b *testing.B) {
		benchmarkSearchWithDataset(b, "medium")
	})

	// Test with large dataset (10000 records)
	b.Run("LargeDataset_10000Records", func(b *testing.B) {
		benchmarkSearchWithDataset(b, "large")
	})
}

// benchmarkSearchWithDataset is a helper function to test search with different dataset sizes
func benchmarkSearchWithDataset(b *testing.B, datasetSize string) {
	// Setup benchmark server
	server := setup.NewBenchmarkServer(b)
	defer server.Cleanup()

	// Start the server
	require.NoError(b, server.Start())

	// Seed appropriate dataset based on size
	switch datasetSize {
	case "small":
		require.NoError(b, server.SeedSmallDataSet())
	case "medium":
		require.NoError(b, server.SeedMediumDataSet())
	case "large":
		require.NoError(b, server.SeedLargeDataSet())
	default:
		require.NoError(b, server.SeedMediumDataSet())
	}

	// Create HTTP client
	client := helpers.NewBenchmarkClient(server.BaseURL)

	// Setup authentication
	authHelper := helpers.NewAuthHelper(server.Config.JWT.Secret)
	testUser := helpers.GetDefaultTestUser()
	require.NoError(b, authHelper.AuthenticateClient(client, testUser.ID, testUser.Username))

	b.ResetTimer()

	// Test various search scenarios
	searchQueries := []string{
		"epic",
		"user story",
		"requirement test",
		"acceptance criteria",
		"system performance",
	}

	for i := 0; i < b.N; i++ {
		query := searchQueries[i%len(searchQueries)]
		resp, err := client.GET(fmt.Sprintf("/api/v1/search?query=%s&limit=20", url.QueryEscape(query)))
		require.NoError(b, err)
		require.Equal(b, http.StatusOK, resp.StatusCode)
		resp.Body.Close()
	}
}

// BenchmarkSearchResultRanking tests search result ranking and relevance scoring performance
func BenchmarkSearchResultRanking(b *testing.B) {
	// Setup benchmark server
	server := setup.NewBenchmarkServer(b)
	defer server.Cleanup()

	// Start the server
	require.NoError(b, server.Start())

	// Seed medium dataset for ranking operations
	require.NoError(b, server.SeedMediumDataSet())

	// Create HTTP client
	client := helpers.NewBenchmarkClient(server.BaseURL)

	// Setup authentication
	authHelper := helpers.NewAuthHelper(server.Config.JWT.Secret)
	testUser := helpers.GetDefaultTestUser()
	require.NoError(b, authHelper.AuthenticateClient(client, testUser.ID, testUser.Username))

	b.ResetTimer()

	b.Run("RelevanceScoring", func(b *testing.B) {
		// Test queries that should return results with different relevance scores
		relevanceQueries := []string{
			"epic feature development",
			"user story acceptance criteria",
			"requirement system validation",
			"test performance benchmark",
			"feature epic user story requirement",
		}

		for i := 0; i < b.N; i++ {
			query := relevanceQueries[i%len(relevanceQueries)]
			resp, err := client.GET(fmt.Sprintf("/api/v1/search?query=%s&limit=50", url.QueryEscape(query)))
			require.NoError(b, err)
			require.Equal(b, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		}
	})

	b.Run("SortByRelevance", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			resp, err := client.GET("/api/v1/search?query=epic&sort_by=created_at&sort_order=desc&limit=30")
			require.NoError(b, err)
			require.Equal(b, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		}
	})

	b.Run("ComplexRankingQueries", func(b *testing.B) {
		complexQueries := []string{
			"epic AND user story",
			"requirement OR acceptance",
			"system performance test",
			"feature development epic",
			"validation criteria requirement",
		}

		for i := 0; i < b.N; i++ {
			query := complexQueries[i%len(complexQueries)]
			resp, err := client.GET(fmt.Sprintf("/api/v1/search?query=%s&sort_by=title&sort_order=asc&limit=25",
				url.QueryEscape(query)))
			require.NoError(b, err)
			require.Equal(b, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		}
	})

	b.Run("RankingWithFilters", func(b *testing.B) {
		// Get test data for filtering
		var users []models.User
		require.NoError(b, server.DB.Limit(2).Find(&users).Error)
		require.NotEmpty(b, users)

		for i := 0; i < b.N; i++ {
			user := users[i%len(users)]
			priority := (i % 4) + 1

			resp, err := client.GET(fmt.Sprintf("/api/v1/search?query=epic&creator_id=%s&priority=%d&sort_by=created_at&sort_order=desc&limit=20",
				user.ID, priority))
			require.NoError(b, err)
			require.Equal(b, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		}
	})
}

// BenchmarkSearchSuggestions tests search suggestions API performance
func BenchmarkSearchSuggestions(b *testing.B) {
	// Setup benchmark server
	server := setup.NewBenchmarkServer(b)
	defer server.Cleanup()

	// Start the server
	require.NoError(b, server.Start())

	// Seed small dataset for suggestions
	require.NoError(b, server.SeedSmallDataSet())

	// Create HTTP client
	client := helpers.NewBenchmarkClient(server.BaseURL)

	// Setup authentication
	authHelper := helpers.NewAuthHelper(server.Config.JWT.Secret)
	testUser := helpers.GetDefaultTestUser()
	require.NoError(b, authHelper.AuthenticateClient(client, testUser.ID, testUser.Username))

	b.ResetTimer()

	b.Run("GetSearchSuggestions", func(b *testing.B) {
		partialQueries := []string{
			"ep",
			"us",
			"req",
			"acc",
			"test",
			"feat",
			"sys",
			"perf",
		}

		for i := 0; i < b.N; i++ {
			query := partialQueries[i%len(partialQueries)]
			resp, err := client.GET(fmt.Sprintf("/api/v1/search/suggestions?query=%s", url.QueryEscape(query)))
			require.NoError(b, err)
			require.Equal(b, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		}
	})

	b.Run("GetSearchSuggestionsWithLimit", func(b *testing.B) {
		limits := []int{5, 10, 15, 20}

		for i := 0; i < b.N; i++ {
			limit := limits[i%len(limits)]
			resp, err := client.GET(fmt.Sprintf("/api/v1/search/suggestions?query=epic&limit=%d", limit))
			require.NoError(b, err)
			require.Equal(b, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		}
	})
}

// BenchmarkSearchConcurrentOperations tests concurrent search operations
func BenchmarkSearchConcurrentOperations(b *testing.B) {
	// Setup benchmark server
	server := setup.NewBenchmarkServer(b)
	defer server.Cleanup()

	// Start the server
	require.NoError(b, server.Start())

	// Seed medium dataset for concurrent operations
	require.NoError(b, server.SeedMediumDataSet())

	// Create HTTP client
	client := helpers.NewBenchmarkClient(server.BaseURL)

	// Setup authentication
	authHelper := helpers.NewAuthHelper(server.Config.JWT.Secret)
	testUser := helpers.GetDefaultTestUser()
	require.NoError(b, authHelper.AuthenticateClient(client, testUser.ID, testUser.Username))

	b.ResetTimer()

	b.Run("ConcurrentSearchRequests", func(b *testing.B) {
		// Create search requests for parallel execution
		searchQueries := []string{
			"epic feature",
			"user story",
			"requirement test",
			"acceptance criteria",
			"system performance",
			"validation test",
			"development epic",
			"story requirement",
		}

		requests := make([]helpers.Request, b.N)
		for i := 0; i < b.N; i++ {
			query := searchQueries[i%len(searchQueries)]
			requests[i] = helpers.Request{
				Method: "GET",
				Path:   fmt.Sprintf("/api/v1/search?query=%s&limit=20", url.QueryEscape(query)),
			}
		}

		// Execute requests with limited concurrency
		concurrency := 10
		responses, err := client.RunParallelRequests(requests, concurrency)
		require.NoError(b, err)

		// Verify all requests succeeded
		for i, resp := range responses {
			require.NoError(b, resp.Error, "Request %d failed", i)
			require.Equal(b, http.StatusOK, resp.StatusCode, "Request %d returned wrong status", i)
		}
	})

	b.Run("ConcurrentMixedSearchOperations", func(b *testing.B) {
		// Mix of search and suggestions requests
		requests := make([]helpers.Request, b.N)
		for i := 0; i < b.N; i++ {
			if i%3 == 0 {
				// Search suggestions request
				requests[i] = helpers.Request{
					Method: "GET",
					Path:   "/api/v1/search/suggestions?query=epic",
				}
			} else {
				// Regular search request
				query := []string{"epic", "user story", "requirement"}[i%3]
				requests[i] = helpers.Request{
					Method: "GET",
					Path:   fmt.Sprintf("/api/v1/search?query=%s", url.QueryEscape(query)),
				}
			}
		}

		// Execute requests with limited concurrency
		concurrency := 15
		responses, err := client.RunParallelRequests(requests, concurrency)
		require.NoError(b, err)

		// Verify all requests succeeded
		for i, resp := range responses {
			require.NoError(b, resp.Error, "Request %d failed", i)
			require.Equal(b, http.StatusOK, resp.StatusCode, "Request %d returned wrong status", i)
		}
	})

	b.Run("ConcurrentFilteredSearches", func(b *testing.B) {
		// Get test data for filtering
		var users []models.User
		require.NoError(b, server.DB.Limit(3).Find(&users).Error)
		require.NotEmpty(b, users)

		requests := make([]helpers.Request, b.N)
		for i := 0; i < b.N; i++ {
			user := users[i%len(users)]
			priority := (i % 4) + 1
			status := []string{"Backlog", "Draft", "In Progress", "Done"}[i%4]

			requests[i] = helpers.Request{
				Method: "GET",
				Path: fmt.Sprintf("/api/v1/search?query=epic&creator_id=%s&priority=%d&status=%s&limit=10",
					user.ID, priority, url.QueryEscape(status)),
			}
		}

		// Execute requests with limited concurrency
		concurrency := 12
		responses, err := client.RunParallelRequests(requests, concurrency)
		require.NoError(b, err)

		// Verify all requests succeeded
		for i, resp := range responses {
			require.NoError(b, resp.Error, "Request %d failed", i)
			require.Equal(b, http.StatusOK, resp.StatusCode, "Request %d returned wrong status", i)
		}
	})
}

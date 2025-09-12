package api

import (
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"product-requirements-management/internal/benchmarks/helpers"
	"product-requirements-management/internal/benchmarks/setup"
	"product-requirements-management/internal/models"
)

// BenchmarkConcurrentSearchScalability tests search performance under concurrent load
func BenchmarkConcurrentSearchScalability(b *testing.B) {
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

	// Test different concurrency levels
	concurrencyLevels := []int{5, 10, 20, 50}

	for _, concurrency := range concurrencyLevels {
		b.Run(fmt.Sprintf("Concurrency_%d", concurrency), func(b *testing.B) {
			searchQueries := []string{
				"epic feature development",
				"user story acceptance",
				"requirement validation test",
				"system performance benchmark",
				"acceptance criteria validation",
				"feature epic user story",
				"test system requirement",
				"performance validation epic",
			}

			// Create requests for parallel execution
			requests := make([]helpers.Request, b.N)
			for i := 0; i < b.N; i++ {
				query := searchQueries[i%len(searchQueries)]
				requests[i] = helpers.Request{
					Method: "GET",
					Path:   fmt.Sprintf("/api/v1/search?query=%s&limit=20", url.QueryEscape(query)),
				}
			}

			// Execute requests with specified concurrency
			responses, err := client.RunParallelRequests(requests, concurrency)
			require.NoError(b, err)

			// Verify all requests succeeded
			successCount := 0
			for i, resp := range responses {
				if resp.Error == nil && resp.StatusCode == http.StatusOK {
					successCount++
				} else {
					b.Logf("Request %d failed: status=%d, error=%v", i, resp.StatusCode, resp.Error)
				}
			}

			// Report success rate
			successRate := float64(successCount) / float64(len(responses)) * 100
			b.ReportMetric(successRate, "success_rate_%")

			// Calculate average response time
			var totalDuration time.Duration
			for _, resp := range responses {
				totalDuration += resp.Duration
			}
			avgDuration := totalDuration / time.Duration(len(responses))
			b.ReportMetric(float64(avgDuration.Milliseconds()), "avg_response_ms")
		})
	}
}

// BenchmarkConcurrentMixedSearchWorkload tests mixed search operations under concurrent load
func BenchmarkConcurrentMixedSearchWorkload(b *testing.B) {
	// Setup benchmark server
	server := setup.NewBenchmarkServer(b)
	defer server.Cleanup()

	// Start the server
	require.NoError(b, server.Start())

	// Seed medium dataset for mixed workload operations
	require.NoError(b, server.SeedMediumDataSet())

	// Create HTTP client
	client := helpers.NewBenchmarkClient(server.BaseURL)

	// Setup authentication
	authHelper := helpers.NewAuthHelper(server.Config.JWT.Secret)
	testUser := helpers.GetDefaultTestUser()
	require.NoError(b, authHelper.AuthenticateClient(client, testUser.ID, testUser.Username))

	// Get test data for filtering
	var users []models.User
	require.NoError(b, server.DB.Limit(5).Find(&users).Error)
	require.NotEmpty(b, users)

	var epics []models.Epic
	require.NoError(b, server.DB.Limit(5).Find(&epics).Error)
	require.NotEmpty(b, epics)

	b.ResetTimer()

	b.Run("MixedSearchOperations", func(b *testing.B) {
		// Create a mix of different search operations
		requests := make([]helpers.Request, b.N)

		for i := 0; i < b.N; i++ {
			switch i % 6 {
			case 0:
				// Simple keyword search
				query := []string{"epic", "user", "story", "requirement", "test"}[i%5]
				requests[i] = helpers.Request{
					Method: "GET",
					Path:   fmt.Sprintf("/api/v1/search?query=%s", url.QueryEscape(query)),
				}
			case 1:
				// Multi-word search
				query := []string{
					"user story epic",
					"acceptance criteria test",
					"requirement system performance",
					"epic feature development",
				}[i%4]
				requests[i] = helpers.Request{
					Method: "GET",
					Path:   fmt.Sprintf("/api/v1/search?query=%s", url.QueryEscape(query)),
				}
			case 2:
				// Filtered search by status
				status := []string{"Backlog", "Draft", "In Progress", "Done"}[i%4]
				requests[i] = helpers.Request{
					Method: "GET",
					Path:   fmt.Sprintf("/api/v1/search?status=%s&limit=15", url.QueryEscape(status)),
				}
			case 3:
				// Filtered search by creator
				user := users[i%len(users)]
				requests[i] = helpers.Request{
					Method: "GET",
					Path:   fmt.Sprintf("/api/v1/search?creator_id=%s&limit=10", user.ID),
				}
			case 4:
				// Paginated search
				offset := (i % 10) * 5
				requests[i] = helpers.Request{
					Method: "GET",
					Path:   fmt.Sprintf("/api/v1/search?query=epic&limit=10&offset=%d", offset),
				}
			case 5:
				// Search suggestions
				query := []string{"ep", "us", "req", "acc", "test"}[i%5]
				requests[i] = helpers.Request{
					Method: "GET",
					Path:   fmt.Sprintf("/api/v1/search/suggestions?query=%s", url.QueryEscape(query)),
				}
			}
		}

		// Execute mixed workload with moderate concurrency
		concurrency := 15
		responses, err := client.RunParallelRequests(requests, concurrency)
		require.NoError(b, err)

		// Analyze results by operation type
		operationStats := make(map[string]struct {
			count    int
			success  int
			totalDur time.Duration
		})

		for i, resp := range responses {
			opType := getOperationType(requests[i].Path)
			stats := operationStats[opType]
			stats.count++
			stats.totalDur += resp.Duration
			if resp.Error == nil && resp.StatusCode == http.StatusOK {
				stats.success++
			}
			operationStats[opType] = stats
		}

		// Report metrics for each operation type
		for opType, stats := range operationStats {
			var successRate float64
			var avgDuration time.Duration
			if stats.count > 0 {
				successRate = float64(stats.success) / float64(stats.count) * 100
				avgDuration = stats.totalDur / time.Duration(stats.count)
			}
			b.Logf("%s: success_rate=%.1f%%, avg_duration=%v", opType, successRate, avgDuration)
		}
	})

	b.Run("ConcurrentComplexQueries", func(b *testing.B) {
		// Create complex search queries that stress the system
		requests := make([]helpers.Request, b.N)

		for i := 0; i < b.N; i++ {
			user := users[i%len(users)]
			epic := epics[i%len(epics)]
			priority := (i % 4) + 1
			status := []string{"Backlog", "Draft", "In Progress", "Done"}[i%4]

			// Complex query with multiple filters and sorting
			requests[i] = helpers.Request{
				Method: "GET",
				Path: fmt.Sprintf("/api/v1/search?query=epic feature&creator_id=%s&epic_id=%s&priority=%d&status=%s&sort_by=created_at&sort_order=desc&limit=25",
					user.ID, epic.ID, priority, url.QueryEscape(status)),
			}
		}

		// Execute complex queries with high concurrency
		concurrency := 20
		responses, err := client.RunParallelRequests(requests, concurrency)
		require.NoError(b, err)

		// Verify performance under complex query load
		successCount := 0
		var totalDuration time.Duration
		maxDuration := time.Duration(0)

		for _, resp := range responses {
			if resp.Error == nil && resp.StatusCode == http.StatusOK {
				successCount++
			}
			totalDuration += resp.Duration
			if resp.Duration > maxDuration {
				maxDuration = resp.Duration
			}
		}

		var successRate float64
		var avgDuration time.Duration
		if len(responses) > 0 {
			successRate = float64(successCount) / float64(len(responses)) * 100
			avgDuration = totalDuration / time.Duration(len(responses))
		}

		b.ReportMetric(successRate, "success_rate_%")
		b.ReportMetric(float64(avgDuration.Milliseconds()), "avg_response_ms")
		b.ReportMetric(float64(maxDuration.Milliseconds()), "max_response_ms")
	})
}

// BenchmarkConcurrentSearchWithDifferentDatasets tests concurrent search across different dataset sizes
func BenchmarkConcurrentSearchWithDifferentDatasets(b *testing.B) {
	datasets := []struct {
		name     string
		seedFunc func(*setup.BenchmarkServer) error
	}{
		{"SmallDataset", (*setup.BenchmarkServer).SeedSmallDataSet},
		{"MediumDataset", (*setup.BenchmarkServer).SeedMediumDataSet},
		{"LargeDataset", (*setup.BenchmarkServer).SeedLargeDataSet},
	}

	for _, dataset := range datasets {
		b.Run(dataset.name, func(b *testing.B) {
			// Setup benchmark server
			server := setup.NewBenchmarkServer(b)
			defer server.Cleanup()

			// Start the server
			require.NoError(b, server.Start())

			// Seed appropriate dataset
			require.NoError(b, dataset.seedFunc(server))

			// Create HTTP client
			client := helpers.NewBenchmarkClient(server.BaseURL)

			// Setup authentication
			authHelper := helpers.NewAuthHelper(server.Config.JWT.Secret)
			testUser := helpers.GetDefaultTestUser()
			require.NoError(b, authHelper.AuthenticateClient(client, testUser.ID, testUser.Username))

			b.ResetTimer()

			// Test concurrent search performance on this dataset
			searchQueries := []string{
				"epic feature development",
				"user story acceptance criteria",
				"requirement system validation",
				"test performance benchmark",
				"acceptance criteria validation",
			}

			requests := make([]helpers.Request, b.N)
			for i := 0; i < b.N; i++ {
				query := searchQueries[i%len(searchQueries)]
				limit := []int{10, 20, 30}[i%3]
				requests[i] = helpers.Request{
					Method: "GET",
					Path:   fmt.Sprintf("/api/v1/search?query=%s&limit=%d", url.QueryEscape(query), limit),
				}
			}

			// Execute with dataset-appropriate concurrency
			concurrency := 12
			if dataset.name == "LargeDataset" {
				concurrency = 8 // Reduce concurrency for large datasets
			}

			responses, err := client.RunParallelRequests(requests, concurrency)
			require.NoError(b, err)

			// Analyze performance characteristics
			successCount := 0
			var totalDuration time.Duration
			var responseSizes []int

			for _, resp := range responses {
				if resp.Error == nil && resp.StatusCode == http.StatusOK {
					successCount++
					responseSizes = append(responseSizes, len(resp.Body))
				}
				totalDuration += resp.Duration
			}

			var successRate float64
			var avgDuration time.Duration
			if len(responses) > 0 {
				successRate = float64(successCount) / float64(len(responses)) * 100
				avgDuration = totalDuration / time.Duration(len(responses))
			}

			// Calculate average response size
			var avgResponseSize float64
			if len(responseSizes) > 0 {
				totalSize := 0
				for _, size := range responseSizes {
					totalSize += size
				}
				avgResponseSize = float64(totalSize) / float64(len(responseSizes))
			}

			b.ReportMetric(successRate, "success_rate_%")
			b.ReportMetric(float64(avgDuration.Milliseconds()), "avg_response_ms")
			b.ReportMetric(avgResponseSize, "avg_response_bytes")
		})
	}
}

// BenchmarkConcurrentSearchStressTest performs stress testing with high concurrent load
func BenchmarkConcurrentSearchStressTest(b *testing.B) {
	// Setup benchmark server
	server := setup.NewBenchmarkServer(b)
	defer server.Cleanup()

	// Start the server
	require.NoError(b, server.Start())

	// Seed large dataset for stress testing
	require.NoError(b, server.SeedLargeDataSet())

	// Create multiple HTTP clients to simulate different users
	numClients := 5
	clients := make([]*helpers.BenchmarkClient, numClients)

	for i := 0; i < numClients; i++ {
		clients[i] = helpers.NewBenchmarkClient(server.BaseURL)

		// Setup authentication for each client
		authHelper := helpers.NewAuthHelper(server.Config.JWT.Secret)
		testUser := helpers.GetDefaultTestUser()
		require.NoError(b, authHelper.AuthenticateClient(clients[i], testUser.ID, testUser.Username))
	}

	b.ResetTimer()

	b.Run("HighConcurrencyStressTest", func(b *testing.B) {
		// Create a large number of concurrent requests
		totalRequests := b.N
		requestsPerClient := totalRequests / numClients

		var wg sync.WaitGroup
		results := make([][]helpers.Response, numClients)

		// Launch concurrent workers
		for clientIdx := 0; clientIdx < numClients; clientIdx++ {
			wg.Add(1)
			go func(idx int, client *helpers.BenchmarkClient) {
				defer wg.Done()

				// Create requests for this client
				requests := make([]helpers.Request, requestsPerClient)
				searchQueries := []string{
					"epic feature development system",
					"user story acceptance criteria validation",
					"requirement test performance benchmark",
					"system validation epic feature",
					"acceptance criteria requirement test",
				}

				for i := 0; i < requestsPerClient; i++ {
					query := searchQueries[i%len(searchQueries)]
					requests[i] = helpers.Request{
						Method: "GET",
						Path:   fmt.Sprintf("/api/v1/search?query=%s&limit=15", url.QueryEscape(query)),
					}
				}

				// Execute requests with high concurrency per client
				responses, err := client.RunParallelRequests(requests, 10)
				require.NoError(b, err)
				results[idx] = responses
			}(clientIdx, clients[clientIdx])
		}

		wg.Wait()

		// Aggregate results from all clients
		totalResponses := 0
		totalSuccess := 0
		var totalDuration time.Duration
		maxDuration := time.Duration(0)

		for _, clientResults := range results {
			for _, resp := range clientResults {
				totalResponses++
				if resp.Error == nil && resp.StatusCode == http.StatusOK {
					totalSuccess++
				}
				totalDuration += resp.Duration
				if resp.Duration > maxDuration {
					maxDuration = resp.Duration
				}
			}
		}

		var successRate float64
		var avgDuration time.Duration
		if totalResponses > 0 {
			successRate = float64(totalSuccess) / float64(totalResponses) * 100
			avgDuration = totalDuration / time.Duration(totalResponses)
		}

		b.ReportMetric(successRate, "success_rate_%")
		b.ReportMetric(float64(avgDuration.Milliseconds()), "avg_response_ms")
		b.ReportMetric(float64(maxDuration.Milliseconds()), "max_response_ms")
		b.ReportMetric(float64(totalResponses), "total_requests")
	})

	b.Run("SustainedLoadTest", func(b *testing.B) {
		// Test sustained load over time
		duration := 10 * time.Second
		if testing.Short() {
			duration = 2 * time.Second
		}

		var wg sync.WaitGroup
		stopChan := make(chan struct{})
		results := make(chan helpers.Response, 1000)

		// Start multiple workers
		for i := 0; i < numClients; i++ {
			wg.Add(1)
			go func(client *helpers.BenchmarkClient) {
				defer wg.Done()

				searchQueries := []string{
					"epic",
					"user story",
					"requirement",
					"acceptance criteria",
					"test validation",
				}

				queryIdx := 0
				for {
					select {
					case <-stopChan:
						return
					default:
						query := searchQueries[queryIdx%len(searchQueries)]
						queryIdx++

						start := time.Now()
						resp, err := client.GET(fmt.Sprintf("/api/v1/search?query=%s&limit=10", url.QueryEscape(query)))
						duration := time.Since(start)

						result := helpers.Response{
							Duration: duration,
							Error:    err,
						}

						if err == nil && resp != nil {
							result.StatusCode = resp.StatusCode
							resp.Body.Close()
						}

						select {
						case results <- result:
						default:
							// Channel full, skip this result
						}
					}
				}
			}(clients[i])
		}

		// Let the test run for the specified duration
		time.Sleep(duration)
		close(stopChan)
		wg.Wait()
		close(results)

		// Collect and analyze results
		var responses []helpers.Response
		for result := range results {
			responses = append(responses, result)
		}

		if len(responses) > 0 {
			successCount := 0
			var totalDuration time.Duration

			for _, resp := range responses {
				if resp.Error == nil && resp.StatusCode == http.StatusOK {
					successCount++
				}
				totalDuration += resp.Duration
			}

			var successRate float64
			var avgDuration time.Duration
			var requestsPerSecond float64
			if len(responses) > 0 {
				successRate = float64(successCount) / float64(len(responses)) * 100
				avgDuration = totalDuration / time.Duration(len(responses))
				requestsPerSecond = float64(len(responses)) / duration.Seconds()
			}

			b.ReportMetric(successRate, "success_rate_%")
			b.ReportMetric(float64(avgDuration.Milliseconds()), "avg_response_ms")
			b.ReportMetric(requestsPerSecond, "requests_per_second")
			b.ReportMetric(float64(len(responses)), "total_requests")
		}
	})
}

// getOperationType determines the type of search operation from the request path
func getOperationType(path string) string {
	if contains(path, "/suggestions") {
		return "suggestions"
	} else if contains(path, "creator_id=") {
		return "filtered_by_creator"
	} else if contains(path, "status=") {
		return "filtered_by_status"
	} else if contains(path, "offset=") {
		return "paginated"
	} else if contains(path, "query=") {
		return "keyword_search"
	}
	return "other"
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			(len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					containsSubstring(s, substr))))
}

// containsSubstring checks if s contains substr as a substring
func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

package api

import (
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"product-requirements-management/internal/benchmarks/helpers"
	"product-requirements-management/internal/benchmarks/setup"
	"product-requirements-management/internal/models"
	"product-requirements-management/internal/service"
)

// BenchmarkConcurrentCRUDOperations tests multiple simultaneous CRUD operations using parallel HTTP clients
func BenchmarkConcurrentCRUDOperations(b *testing.B) {
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

	// Get test data for operations
	var users []models.User
	require.NoError(b, server.DB.Limit(5).Find(&users).Error)
	require.NotEmpty(b, users)

	var epics []models.Epic
	require.NoError(b, server.DB.Limit(10).Find(&epics).Error)
	require.NotEmpty(b, epics)

	b.ResetTimer()

	// Test different concurrency levels for CRUD operations
	concurrencyLevels := []int{5, 10, 20, 30}

	for _, concurrency := range concurrencyLevels {
		b.Run(fmt.Sprintf("ConcurrentCRUD_Concurrency_%d", concurrency), func(b *testing.B) {
			// Create mixed CRUD requests for parallel execution
			requests := make([]helpers.Request, b.N)

			for i := 0; i < b.N; i++ {
				user := users[i%len(users)]
				epic := epics[i%len(epics)]

				switch i % 4 {
				case 0: // CREATE Epic
					description := fmt.Sprintf("Epic created in concurrent test %d", i)
					createReq := service.CreateEpicRequest{
						Title:       fmt.Sprintf("Concurrent Epic %d", i),
						Description: &description,
						CreatorID:   user.ID,
						Priority:    models.Priority((i % 4) + 1),
					}
					requests[i] = helpers.Request{
						Method: "POST",
						Path:   "/api/v1/epics",
						Body:   createReq,
					}
				case 1: // READ Epic
					requests[i] = helpers.Request{
						Method: "GET",
						Path:   fmt.Sprintf("/api/v1/epics/%s", epic.ID),
					}
				case 2: // UPDATE Epic
					title := fmt.Sprintf("Updated Concurrent Epic %d", i)
					description := fmt.Sprintf("Epic updated in concurrent test %d", i)
					priority := models.Priority(((i + 1) % 4) + 1)
					updateReq := service.UpdateEpicRequest{
						Title:       &title,
						Description: &description,
						Priority:    &priority,
					}
					requests[i] = helpers.Request{
						Method: "PUT",
						Path:   fmt.Sprintf("/api/v1/epics/%s", epic.ID),
						Body:   updateReq,
					}
				case 3: // LIST Epics
					requests[i] = helpers.Request{
						Method: "GET",
						Path:   "/api/v1/epics?limit=20&offset=0",
					}
				}
			}

			// Execute requests with specified concurrency
			responses, err := client.RunParallelRequests(requests, concurrency)
			require.NoError(b, err)

			// Analyze results by operation type
			operationStats := make(map[string]struct {
				count    int
				success  int
				totalDur time.Duration
			})

			for i, resp := range responses {
				opType := getCRUDOperationType(requests[i].Method, requests[i].Path)
				stats := operationStats[opType]
				stats.count++
				stats.totalDur += resp.Duration
				if resp.Error == nil && (resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated) {
					stats.success++
				}
				operationStats[opType] = stats
			}

			// Report metrics for each operation type
			totalSuccess := 0
			var totalDuration time.Duration
			for opType, stats := range operationStats {
				var successRate float64
				var avgDuration time.Duration
				if stats.count > 0 {
					successRate = float64(stats.success) / float64(stats.count) * 100
					avgDuration = stats.totalDur / time.Duration(stats.count)
				}
				totalSuccess += stats.success
				totalDuration += stats.totalDur
				b.Logf("%s: success_rate=%.1f%%, avg_duration=%v, count=%d",
					opType, successRate, avgDuration, stats.count)
			}

			// Report overall metrics
			overallSuccessRate := float64(totalSuccess) / float64(len(responses)) * 100
			avgOverallDuration := totalDuration / time.Duration(len(responses))

			b.ReportMetric(overallSuccessRate, "success_rate_%")
			b.ReportMetric(float64(avgOverallDuration.Milliseconds()), "avg_response_ms")
			b.ReportMetric(float64(concurrency), "concurrency_level")
		})
	}
}

// BenchmarkMixedReadWriteWorkload tests mixed read/write workload with concurrent request runners
func BenchmarkMixedReadWriteWorkload(b *testing.B) {
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

	// Get test data for operations
	var users []models.User
	require.NoError(b, server.DB.Limit(5).Find(&users).Error)
	require.NotEmpty(b, users)

	var epics []models.Epic
	require.NoError(b, server.DB.Limit(10).Find(&epics).Error)
	require.NotEmpty(b, epics)

	var userStories []models.UserStory
	require.NoError(b, server.DB.Limit(20).Find(&userStories).Error)
	require.NotEmpty(b, userStories)

	b.ResetTimer()

	// Test different read/write ratios
	workloadTypes := []struct {
		name      string
		readRatio float64
	}{
		{"ReadHeavy_80_20", 0.8},
		{"Balanced_50_50", 0.5},
		{"WriteHeavy_20_80", 0.2},
	}

	for _, workload := range workloadTypes {
		b.Run(workload.name, func(b *testing.B) {
			// Create mixed read/write requests based on ratio
			requests := make([]helpers.Request, b.N)

			for i := 0; i < b.N; i++ {
				user := users[i%len(users)]
				epic := epics[i%len(epics)]
				userStory := userStories[i%len(userStories)]

				// Determine if this should be a read or write operation
				isRead := float64(i%100)/100.0 < workload.readRatio

				if isRead {
					// Read operations (GET requests)
					switch i % 6 {
					case 0: // Read Epic
						requests[i] = helpers.Request{
							Method: "GET",
							Path:   fmt.Sprintf("/api/v1/epics/%s", epic.ID),
						}
					case 1: // List Epics
						requests[i] = helpers.Request{
							Method: "GET",
							Path:   "/api/v1/epics?limit=10",
						}
					case 2: // Read User Story
						requests[i] = helpers.Request{
							Method: "GET",
							Path:   fmt.Sprintf("/api/v1/user-stories/%s", userStory.ID),
						}
					case 3: // List User Stories
						requests[i] = helpers.Request{
							Method: "GET",
							Path:   fmt.Sprintf("/api/v1/epics/%s/user-stories", epic.ID),
						}
					case 4: // Search
						query := []string{"epic", "user", "story", "requirement", "test"}[i%5]
						requests[i] = helpers.Request{
							Method: "GET",
							Path:   fmt.Sprintf("/api/v1/search?query=%s&limit=10", query),
						}
					case 5: // List Requirements
						requests[i] = helpers.Request{
							Method: "GET",
							Path:   fmt.Sprintf("/api/v1/user-stories/%s/requirements", userStory.ID),
						}
					}
				} else {
					// Write operations (POST, PUT, DELETE requests)
					switch i % 6 {
					case 0: // Create Epic
						description := fmt.Sprintf("Epic created in mixed workload test %d", i)
						createReq := service.CreateEpicRequest{
							Title:       fmt.Sprintf("Mixed Workload Epic %d", i),
							Description: &description,
							CreatorID:   user.ID,
							Priority:    models.Priority((i % 4) + 1),
						}
						requests[i] = helpers.Request{
							Method: "POST",
							Path:   "/api/v1/epics",
							Body:   createReq,
						}
					case 1: // Update Epic
						title := fmt.Sprintf("Updated Mixed Workload Epic %d", i)
						description := fmt.Sprintf("Epic updated in mixed workload test %d", i)
						priority := models.Priority(((i + 1) % 4) + 1)
						updateReq := service.UpdateEpicRequest{
							Title:       &title,
							Description: &description,
							Priority:    &priority,
						}
						requests[i] = helpers.Request{
							Method: "PUT",
							Path:   fmt.Sprintf("/api/v1/epics/%s", epic.ID),
							Body:   updateReq,
						}
					case 2: // Create User Story
						description := fmt.Sprintf("User story created in mixed workload test %d", i)
						createUSReq := service.CreateUserStoryRequest{
							Title:       fmt.Sprintf("Mixed Workload User Story %d", i),
							Description: &description,
							EpicID:      epic.ID,
							CreatorID:   user.ID,
							Priority:    models.Priority((i % 4) + 1),
						}
						requests[i] = helpers.Request{
							Method: "POST",
							Path:   "/api/v1/user-stories",
							Body:   createUSReq,
						}
					case 3: // Update User Story
						title := fmt.Sprintf("Updated Mixed Workload User Story %d", i)
						description := fmt.Sprintf("User story updated in mixed workload test %d", i)
						priority := models.Priority(((i + 1) % 4) + 1)
						updateUSReq := service.UpdateUserStoryRequest{
							Title:       &title,
							Description: &description,
							Priority:    &priority,
						}
						requests[i] = helpers.Request{
							Method: "PUT",
							Path:   fmt.Sprintf("/api/v1/user-stories/%s", userStory.ID),
							Body:   updateUSReq,
						}
					case 4: // Create Requirement
						description := fmt.Sprintf("Requirement created in mixed workload test %d", i)
						// We'll need to get a valid TypeID from the database, for now use a placeholder
						typeID := uuid.New() // This should be replaced with actual requirement type lookup
						createReqReq := service.CreateRequirementRequest{
							Title:       fmt.Sprintf("Mixed Workload Requirement %d", i),
							Description: &description,
							UserStoryID: userStory.ID,
							CreatorID:   user.ID,
							Priority:    models.Priority((i % 4) + 1),
							TypeID:      typeID,
						}
						requests[i] = helpers.Request{
							Method: "POST",
							Path:   "/api/v1/requirements",
							Body:   createReqReq,
						}
					case 5: // Create Comment
						createCommentReq := service.CreateCommentRequest{
							Content:    fmt.Sprintf("Mixed workload comment %d", i),
							AuthorID:   user.ID,
							EntityType: models.EntityTypeEpic,
							EntityID:   epic.ID,
						}
						requests[i] = helpers.Request{
							Method: "POST",
							Path:   "/api/v1/comments",
							Body:   createCommentReq,
						}
					}
				}
			}

			// Execute mixed workload with moderate concurrency
			concurrency := 15
			responses, err := client.RunParallelRequests(requests, concurrency)
			require.NoError(b, err)

			// Analyze results by read/write type
			readStats := struct {
				count    int
				success  int
				totalDur time.Duration
			}{}
			writeStats := struct {
				count    int
				success  int
				totalDur time.Duration
			}{}

			for i, resp := range responses {
				isRead := isReadOperation(requests[i].Method)

				if isRead {
					readStats.count++
					readStats.totalDur += resp.Duration
					if resp.Error == nil && resp.StatusCode == http.StatusOK {
						readStats.success++
					}
				} else {
					writeStats.count++
					writeStats.totalDur += resp.Duration
					if resp.Error == nil && (resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated) {
						writeStats.success++
					}
				}
			}

			// Calculate and report metrics
			var readSuccessRate, writeSuccessRate float64
			var avgReadDuration, avgWriteDuration time.Duration

			if readStats.count > 0 {
				readSuccessRate = float64(readStats.success) / float64(readStats.count) * 100
				avgReadDuration = readStats.totalDur / time.Duration(readStats.count)
			}

			if writeStats.count > 0 {
				writeSuccessRate = float64(writeStats.success) / float64(writeStats.count) * 100
				avgWriteDuration = writeStats.totalDur / time.Duration(writeStats.count)
			}

			b.Logf("Read operations: success_rate=%.1f%%, avg_duration=%v, count=%d",
				readSuccessRate, avgReadDuration, readStats.count)
			b.Logf("Write operations: success_rate=%.1f%%, avg_duration=%v, count=%d",
				writeSuccessRate, avgWriteDuration, writeStats.count)

			// Report overall metrics
			totalSuccess := readStats.success + writeStats.success
			overallSuccessRate := float64(totalSuccess) / float64(len(responses)) * 100

			b.ReportMetric(overallSuccessRate, "success_rate_%")
			b.ReportMetric(readSuccessRate, "read_success_rate_%")
			b.ReportMetric(writeSuccessRate, "write_success_rate_%")
			b.ReportMetric(float64(avgReadDuration.Milliseconds()), "avg_read_ms")
			b.ReportMetric(float64(avgWriteDuration.Milliseconds()), "avg_write_ms")
		})
	}
}

// BenchmarkDatabaseConnectionPoolStress tests database connection pool under concurrent API load
func BenchmarkDatabaseConnectionPoolStress(b *testing.B) {
	// Setup benchmark server
	server := setup.NewBenchmarkServer(b)
	defer server.Cleanup()

	// Start the server
	require.NoError(b, server.Start())

	// Seed large dataset for connection pool stress testing
	require.NoError(b, server.SeedLargeDataSet())

	// Create multiple HTTP clients to simulate different users
	numClients := 10
	clients := make([]*helpers.BenchmarkClient, numClients)

	for i := 0; i < numClients; i++ {
		clients[i] = helpers.NewBenchmarkClient(server.BaseURL)

		// Setup authentication for each client
		authHelper := helpers.NewAuthHelper(server.Config.JWT.Secret)
		testUser := helpers.GetDefaultTestUser()
		require.NoError(b, authHelper.AuthenticateClient(clients[i], testUser.ID, testUser.Username))
	}

	// Get database connection stats before test
	sqlDB, err := server.DB.DB()
	require.NoError(b, err)
	initialStats := sqlDB.Stats()

	b.ResetTimer()

	b.Run("ConnectionPoolStressTest", func(b *testing.B) {
		// Create database-intensive requests
		totalRequests := b.N
		requestsPerClient := totalRequests / numClients

		var wg sync.WaitGroup
		results := make([][]helpers.Response, numClients)

		// Launch concurrent workers that stress the database connection pool
		for clientIdx := 0; clientIdx < numClients; clientIdx++ {
			wg.Add(1)
			go func(idx int, client *helpers.BenchmarkClient) {
				defer wg.Done()

				// Create database-intensive requests for this client
				requests := make([]helpers.Request, requestsPerClient)

				for i := 0; i < requestsPerClient; i++ {
					switch i % 8 {
					case 0: // Complex search query (joins multiple tables)
						requests[i] = helpers.Request{
							Method: "GET",
							Path:   "/api/v1/search?query=epic user story requirement&limit=50",
						}
					case 1: // List epics with user stories (joins)
						requests[i] = helpers.Request{
							Method: "GET",
							Path:   "/api/v1/epics?include_user_stories=true&limit=25",
						}
					case 2: // List user stories with requirements (joins)
						requests[i] = helpers.Request{
							Method: "GET",
							Path:   "/api/v1/user-stories?include_requirements=true&limit=25",
						}
					case 3: // Get epic with all related data (multiple joins)
						requests[i] = helpers.Request{
							Method: "GET",
							Path:   "/api/v1/epics?limit=1&include_user_stories=true",
						}
					case 4: // Search with complex filters (multiple WHERE clauses)
						requests[i] = helpers.Request{
							Method: "GET",
							Path:   "/api/v1/search?query=test&status=In Progress&priority=1&limit=20",
						}
					case 5: // List requirements with relationships (joins)
						requests[i] = helpers.Request{
							Method: "GET",
							Path:   "/api/v1/requirements?include_relationships=true&limit=20",
						}
					case 6: // Get comments with threading (recursive queries)
						requests[i] = helpers.Request{
							Method: "GET",
							Path:   "/api/v1/comments?entity_type=epic&limit=30",
						}
					case 7: // Navigation endpoint (multiple queries)
						requests[i] = helpers.Request{
							Method: "GET",
							Path:   "/api/v1/navigation",
						}
					}
				}

				// Execute requests with high concurrency to stress connection pool
				responses, err := client.RunParallelRequests(requests, 8)
				require.NoError(b, err)
				results[idx] = responses
			}(clientIdx, clients[clientIdx])
		}

		wg.Wait()

		// Get database connection stats after test
		finalStats := sqlDB.Stats()

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

		// Report performance metrics
		b.ReportMetric(successRate, "success_rate_%")
		b.ReportMetric(float64(avgDuration.Milliseconds()), "avg_response_ms")
		b.ReportMetric(float64(maxDuration.Milliseconds()), "max_response_ms")

		// Report database connection pool metrics
		b.ReportMetric(float64(finalStats.OpenConnections), "final_open_connections")
		b.ReportMetric(float64(finalStats.InUse), "final_in_use_connections")
		b.ReportMetric(float64(finalStats.Idle), "final_idle_connections")
		b.ReportMetric(float64(finalStats.WaitCount), "total_wait_count")
		b.ReportMetric(float64(finalStats.WaitDuration.Milliseconds()), "total_wait_duration_ms")

		// Log detailed connection pool statistics
		b.Logf("Initial DB Stats - Open: %d, InUse: %d, Idle: %d",
			initialStats.OpenConnections, initialStats.InUse, initialStats.Idle)
		b.Logf("Final DB Stats - Open: %d, InUse: %d, Idle: %d, WaitCount: %d, WaitDuration: %v",
			finalStats.OpenConnections, finalStats.InUse, finalStats.Idle,
			finalStats.WaitCount, finalStats.WaitDuration)
	})
}

// BenchmarkAPIEndpointScalability tests API endpoint scalability with parallel HTTP request execution
func BenchmarkAPIEndpointScalability(b *testing.B) {
	// Setup benchmark server
	server := setup.NewBenchmarkServer(b)
	defer server.Cleanup()

	// Start the server
	require.NoError(b, server.Start())

	// Seed medium dataset for scalability testing
	require.NoError(b, server.SeedMediumDataSet())

	// Create HTTP client
	client := helpers.NewBenchmarkClient(server.BaseURL)

	// Setup authentication
	authHelper := helpers.NewAuthHelper(server.Config.JWT.Secret)
	testUser := helpers.GetDefaultTestUser()
	require.NoError(b, authHelper.AuthenticateClient(client, testUser.ID, testUser.Username))

	b.ResetTimer()

	// Test scalability with increasing load levels
	loadLevels := []struct {
		name        string
		concurrency int
		duration    time.Duration
	}{
		{"LowLoad", 5, 5 * time.Second},
		{"MediumLoad", 15, 5 * time.Second},
		{"HighLoad", 30, 5 * time.Second},
		{"PeakLoad", 50, 3 * time.Second},
	}

	for _, load := range loadLevels {
		b.Run(load.name, func(b *testing.B) {
			if testing.Short() && load.name == "PeakLoad" {
				b.Skip("Skipping peak load test in short mode")
			}

			// Create a variety of API endpoint requests
			endpointTypes := []struct {
				name     string
				method   string
				pathFunc func(int) string
				bodyFunc func(int) interface{}
			}{
				{
					name:   "ListEpics",
					method: "GET",
					pathFunc: func(i int) string {
						return fmt.Sprintf("/api/v1/epics?limit=%d&offset=%d",
							10+i%10, (i%5)*10)
					},
				},
				{
					name:   "GetEpic",
					method: "GET",
					pathFunc: func(i int) string {
						// We'll need to get actual epic IDs, but for now use a pattern
						return "/api/v1/epics?limit=1"
					},
				},
				{
					name:   "SearchAPI",
					method: "GET",
					pathFunc: func(i int) string {
						queries := []string{"epic", "user story", "requirement", "test", "feature"}
						query := queries[i%len(queries)]
						return fmt.Sprintf("/api/v1/search?query=%s&limit=%d",
							query, 10+i%15)
					},
				},
				{
					name:   "ListUserStories",
					method: "GET",
					pathFunc: func(i int) string {
						return fmt.Sprintf("/api/v1/user-stories?limit=%d&offset=%d",
							15+i%10, (i%3)*15)
					},
				},
				{
					name:   "ListRequirements",
					method: "GET",
					pathFunc: func(i int) string {
						return fmt.Sprintf("/api/v1/requirements?limit=%d&offset=%d",
							20+i%10, (i%4)*20)
					},
				},
				{
					name:   "GetComments",
					method: "GET",
					pathFunc: func(i int) string {
						return fmt.Sprintf("/api/v1/comments?limit=%d&offset=%d",
							25+i%15, (i%6)*25)
					},
				},
			}

			// Run sustained load test
			var wg sync.WaitGroup
			stopChan := make(chan struct{})
			results := make(chan helpers.Response, 2000)

			// Start concurrent workers
			for i := 0; i < load.concurrency; i++ {
				wg.Add(1)
				go func(workerID int) {
					defer wg.Done()

					requestCount := 0
					for {
						select {
						case <-stopChan:
							return
						default:
							// Select endpoint type based on worker ID and request count
							endpointIdx := (workerID + requestCount) % len(endpointTypes)
							endpoint := endpointTypes[endpointIdx]

							// Create request
							path := endpoint.pathFunc(requestCount)
							var body interface{}
							if endpoint.bodyFunc != nil {
								body = endpoint.bodyFunc(requestCount)
							}

							// Execute request
							start := time.Now()
							var resp *http.Response
							var err error

							switch endpoint.method {
							case "GET":
								resp, err = client.GET(path)
							case "POST":
								resp, err = client.POST(path, body)
							case "PUT":
								resp, err = client.PUT(path, body)
							case "DELETE":
								resp, err = client.DELETE(path)
							}

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

							requestCount++
						}
					}
				}(i)
			}

			// Let the test run for the specified duration
			time.Sleep(load.duration)
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
				var responseTimes []time.Duration

				for _, resp := range responses {
					if resp.Error == nil && resp.StatusCode == http.StatusOK {
						successCount++
					}
					totalDuration += resp.Duration
					responseTimes = append(responseTimes, resp.Duration)
				}

				var successRate float64
				var avgDuration time.Duration
				var requestsPerSecond float64
				if len(responses) > 0 {
					successRate = float64(successCount) / float64(len(responses)) * 100
					avgDuration = totalDuration / time.Duration(len(responses))
					requestsPerSecond = float64(len(responses)) / load.duration.Seconds()
				}

				// Calculate percentiles
				p50, p95, p99 := calculatePercentiles(responseTimes)

				b.ReportMetric(successRate, "success_rate_%")
				b.ReportMetric(float64(avgDuration.Milliseconds()), "avg_response_ms")
				b.ReportMetric(float64(p50.Milliseconds()), "p50_response_ms")
				b.ReportMetric(float64(p95.Milliseconds()), "p95_response_ms")
				b.ReportMetric(float64(p99.Milliseconds()), "p99_response_ms")
				b.ReportMetric(requestsPerSecond, "requests_per_second")
				b.ReportMetric(float64(len(responses)), "total_requests")
				b.ReportMetric(float64(load.concurrency), "concurrency_level")

				b.Logf("Load: %s, Concurrency: %d, Total Requests: %d, Success Rate: %.1f%%, RPS: %.1f",
					load.name, load.concurrency, len(responses), successRate, requestsPerSecond)
			}
		})
	}
}

// BenchmarkConcurrentAccessReliability tests system reliability under concurrent access patterns
func BenchmarkConcurrentAccessReliability(b *testing.B) {
	// Setup benchmark server
	server := setup.NewBenchmarkServer(b)
	defer server.Cleanup()

	// Start the server
	require.NoError(b, server.Start())

	// Seed medium dataset for reliability testing
	require.NoError(b, server.SeedMediumDataSet())

	// Create multiple HTTP clients
	numClients := 8
	clients := make([]*helpers.BenchmarkClient, numClients)

	for i := 0; i < numClients; i++ {
		clients[i] = helpers.NewBenchmarkClient(server.BaseURL)

		// Setup authentication for each client
		authHelper := helpers.NewAuthHelper(server.Config.JWT.Secret)
		testUser := helpers.GetDefaultTestUser()
		require.NoError(b, authHelper.AuthenticateClient(clients[i], testUser.ID, testUser.Username))
	}

	b.ResetTimer()

	b.Run("ReliabilityUnderConcurrentLoad", func(b *testing.B) {
		// Test reliability patterns
		patterns := []struct {
			name        string
			concurrency int
			burstSize   int
			burstDelay  time.Duration
		}{
			{"SteadyLoad", 10, 0, 0},
			{"BurstLoad", 5, 20, 2 * time.Second},
			{"SpikeyLoad", 3, 30, 1 * time.Second},
		}

		for _, pattern := range patterns {
			b.Run(pattern.name, func(b *testing.B) {
				var wg sync.WaitGroup
				results := make(chan helpers.Response, 1000)
				stopChan := make(chan struct{})

				// Start base load workers
				for i := 0; i < pattern.concurrency; i++ {
					wg.Add(1)
					go func(client *helpers.BenchmarkClient) {
						defer wg.Done()

						requestCount := 0
						for {
							select {
							case <-stopChan:
								return
							default:
								// Execute a variety of requests
								var resp *http.Response
								var err error
								start := time.Now()

								switch requestCount % 5 {
								case 0:
									resp, err = client.GET("/api/v1/epics?limit=10")
								case 1:
									resp, err = client.GET("/api/v1/user-stories?limit=15")
								case 2:
									resp, err = client.GET("/api/v1/search?query=test&limit=10")
								case 3:
									resp, err = client.GET("/api/v1/requirements?limit=20")
								case 4:
									resp, err = client.GET("/api/v1/comments?limit=25")
								}

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
									// Channel full, skip
								}

								requestCount++

								// Small delay to simulate realistic usage
								time.Sleep(10 * time.Millisecond)
							}
						}
					}(clients[i%len(clients)])
				}

				// Add burst load if configured
				if pattern.burstSize > 0 {
					wg.Add(1)
					go func() {
						defer wg.Done()

						for {
							select {
							case <-stopChan:
								return
							default:
								// Create burst of requests
								burstRequests := make([]helpers.Request, pattern.burstSize)
								for i := 0; i < pattern.burstSize; i++ {
									burstRequests[i] = helpers.Request{
										Method: "GET",
										Path:   "/api/v1/search?query=burst&limit=5",
									}
								}

								// Execute burst
								client := clients[0]
								burstResponses, err := client.RunParallelRequests(burstRequests, pattern.burstSize)
								if err == nil {
									for _, resp := range burstResponses {
										select {
										case results <- resp:
										default:
											// Channel full, skip
										}
									}
								}

								// Wait before next burst
								time.Sleep(pattern.burstDelay)
							}
						}
					}()
				}

				// Run test for a fixed duration
				testDuration := 8 * time.Second
				if testing.Short() {
					testDuration = 3 * time.Second
				}

				time.Sleep(testDuration)
				close(stopChan)
				wg.Wait()
				close(results)

				// Analyze reliability metrics
				var responses []helpers.Response
				for result := range results {
					responses = append(responses, result)
				}

				if len(responses) > 0 {
					successCount := 0
					errorCount := 0
					timeoutCount := 0
					var totalDuration time.Duration

					for _, resp := range responses {
						if resp.Error != nil {
							errorCount++
							if resp.Duration > 5*time.Second {
								timeoutCount++
							}
						} else if resp.StatusCode == http.StatusOK {
							successCount++
						}
						totalDuration += resp.Duration
					}

					var successRate, errorRate, timeoutRate float64
					var avgDuration time.Duration
					if len(responses) > 0 {
						successRate = float64(successCount) / float64(len(responses)) * 100
						errorRate = float64(errorCount) / float64(len(responses)) * 100
						timeoutRate = float64(timeoutCount) / float64(len(responses)) * 100
						avgDuration = totalDuration / time.Duration(len(responses))
					}

					b.ReportMetric(successRate, "success_rate_%")
					b.ReportMetric(errorRate, "error_rate_%")
					b.ReportMetric(timeoutRate, "timeout_rate_%")
					b.ReportMetric(float64(avgDuration.Milliseconds()), "avg_response_ms")
					b.ReportMetric(float64(len(responses)), "total_requests")

					b.Logf("Pattern: %s, Success: %.1f%%, Errors: %.1f%%, Timeouts: %.1f%%, Avg Duration: %v",
						pattern.name, successRate, errorRate, timeoutRate, avgDuration)
				}
			})
		}
	})
}

// Helper functions

// getCRUDOperationType determines the CRUD operation type from HTTP method and path
func getCRUDOperationType(method, path string) string {
	switch method {
	case "POST":
		return "CREATE"
	case "GET":
		if contains(path, "?") || contains(path, "limit=") {
			return "LIST"
		}
		return "READ"
	case "PUT", "PATCH":
		return "UPDATE"
	case "DELETE":
		return "DELETE"
	default:
		return "OTHER"
	}
}

// isReadOperation determines if an HTTP method is a read operation
func isReadOperation(method string) bool {
	return method == "GET" || method == "HEAD" || method == "OPTIONS"
}

// calculatePercentiles calculates response time percentiles
func calculatePercentiles(durations []time.Duration) (p50, p95, p99 time.Duration) {
	if len(durations) == 0 {
		return 0, 0, 0
	}

	// Sort durations
	sorted := make([]time.Duration, len(durations))
	copy(sorted, durations)

	// Simple bubble sort for small arrays
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i] > sorted[j] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	// Calculate percentile indices
	p50Idx := int(float64(len(sorted)) * 0.50)
	p95Idx := int(float64(len(sorted)) * 0.95)
	p99Idx := int(float64(len(sorted)) * 0.99)

	// Ensure indices are within bounds
	if p50Idx >= len(sorted) {
		p50Idx = len(sorted) - 1
	}
	if p95Idx >= len(sorted) {
		p95Idx = len(sorted) - 1
	}
	if p99Idx >= len(sorted) {
		p99Idx = len(sorted) - 1
	}

	return sorted[p50Idx], sorted[p95Idx], sorted[p99Idx]
}

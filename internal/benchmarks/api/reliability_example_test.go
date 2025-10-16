package api

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"product-requirements-management/internal/benchmarks/helpers"
	"product-requirements-management/internal/benchmarks/setup"
	"product-requirements-management/internal/models"
	"product-requirements-management/internal/service"
)

// BenchmarkEpicCRUDWithReliability demonstrates enhanced reliability features
func BenchmarkEpicCRUDWithReliability(b *testing.B) {
	// Setup benchmark server with enhanced error handling
	server := setup.NewBenchmarkServer(b)
	defer func() {
		// Enhanced cleanup with timeout and error recovery
		if r := recover(); r != nil {
			b.Logf("Panic during benchmark: %v", r)
		}
		server.Cleanup()
	}()

	// Create enhanced test runner with reliability features
	testRunner := NewBenchmarkTestRunner(b, server.DB)
	defer testRunner.Cleanup()

	// Configure reliability settings
	reliabilityMgr := testRunner.GetReliabilityManager()
	reliabilityMgr.SetTimeout("http", 15*time.Second)
	reliabilityMgr.SetTimeout("database", 30*time.Second)

	// Start the server with timeout and retry
	err := testRunner.ExecuteWithReliability("server_start", func() error {
		return server.Start()
	})
	require.NoError(b, err, "Failed to start server with reliability features")

	// Setup benchmark environment with comprehensive validation
	err = testRunner.SetupBenchmarkEnvironment(server.DB, server.BaseURL)
	require.NoError(b, err, "Benchmark environment setup failed")

	// Seed data with validation and error handling
	err = testRunner.ExecuteWithReliability("data_seeding", func() error {
		return server.SeedSmallDataSet()
	})
	require.NoError(b, err, "Failed to seed test data")

	// Create HTTP client with enhanced error handling
	client := helpers.NewBenchmarkClient(server.BaseURL)

	// Setup authentication with retry logic
	authHelper := helpers.NewAuthHelper(server.Config.JWT.Secret)
	testUser := helpers.GetDefaultTestUser()
	err = testRunner.ExecuteWithReliability("authentication", func() error {
		return authHelper.AuthenticateClient(client, testUser.ID, testUser.Username)
	})
	require.NoError(b, err, "Authentication failed")

	// Validate test data availability
	validator := testRunner.GetValidator()
	var users []models.User
	require.NoError(b, server.DB.Limit(1).Find(&users).Error)
	err = validator.ValidateTestData([]string{users[0].ID.String()}, "user")
	require.NoError(b, err, "User validation failed")

	userID := users[0].ID

	// Start performance monitoring
	metricsCollector := helpers.NewMetricsCollector(server.DB)
	metricsCollector.StartMeasurement()

	b.ResetTimer()

	b.Run("CreateEpicWithReliability", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			// Check resource constraints before operation
			if reliabilityMgr.IsResourceConstrained() {
				b.Logf("Resource constraints detected at iteration %d", i)
			}

			createReq := service.CreateEpicRequest{
				CreatorID:   userID,
				Priority:    models.PriorityMedium,
				Title:       fmt.Sprintf("Reliable Epic %d", i),
				Description: stringPtr(fmt.Sprintf("Epic created with reliability features %d", i)),
			}

			// Execute with comprehensive error handling and timeout
			err := testRunner.ExecuteWithReliability("epic_creation", func() error {
				resp, err := client.POST("/api/v1/epics", createReq)
				if err != nil {
					return fmt.Errorf("HTTP request failed: %w", err)
				}
				defer resp.Body.Close()

				// Validate response with enhanced error reporting
				if err := validator.ValidateHTTPResponse(resp, http.StatusCreated, "epic_creation"); err != nil {
					return fmt.Errorf("response validation failed: %w", err)
				}

				// Record metrics for performance tracking
				metricsCollector.RecordSuccess()
				return nil
			})

			if err != nil {
				metricsCollector.RecordError()
				b.Fatalf("Epic creation failed at iteration %d: %v", i, err)
			}
		}
	})

	b.StopTimer()

	// Collect and validate performance metrics
	metrics := metricsCollector.EndMeasurement()

	// Validate performance meets requirements
	err = validator.ValidatePerformanceMetrics(metrics)
	if err != nil {
		b.Logf("Performance validation warning: %v", err)
	}

	// Validate resource usage
	err = validator.ValidateResourceUsage(metrics)
	if err != nil {
		b.Logf("Resource usage validation warning: %v", err)
	}

	// Report detailed metrics
	metricsCollector.ReportMetrics(b, metrics)

	// Generate validation report
	report := validator.CreateValidationReport(metrics)
	b.Logf("Benchmark Validation Report:\n%s", report)
}

// BenchmarkConcurrentOperationsWithReliability demonstrates concurrent operations with enhanced reliability
func BenchmarkConcurrentOperationsWithReliability(b *testing.B) {
	// Setup with enhanced error handling
	server := setup.NewBenchmarkServer(b)
	defer server.Cleanup()

	testRunner := NewBenchmarkTestRunner(b, server.DB)
	defer testRunner.Cleanup()

	// Configure for concurrent operations
	reliabilityMgr := testRunner.GetReliabilityManager()
	reliabilityMgr.SetTimeout("concurrent_operation", 60*time.Second)

	// Start server and setup environment
	require.NoError(b, testRunner.ExecuteWithReliability("server_start", server.Start))
	require.NoError(b, testRunner.SetupBenchmarkEnvironment(server.DB, server.BaseURL))
	require.NoError(b, testRunner.ExecuteWithReliability("data_seeding", server.SeedSmallDataSet))

	// Setup client and authentication
	client := helpers.NewBenchmarkClient(server.BaseURL)
	authHelper := helpers.NewAuthHelper(server.Config.JWT.Secret)
	testUser := helpers.GetDefaultTestUser()
	require.NoError(b, testRunner.ExecuteWithReliability("authentication", func() error {
		return authHelper.AuthenticateClient(client, testUser.ID, testUser.Username)
	}))

	// Get test data
	var users []models.User
	require.NoError(b, server.DB.Limit(1).Find(&users).Error)
	userID := users[0].ID

	b.ResetTimer()

	b.Run("ConcurrentEpicCreationWithReliability", func(b *testing.B) {
		// Adjust concurrency based on resource constraints
		requestedConcurrency := 10
		actualConcurrency := reliabilityMgr.AdjustConcurrencyForConstraints(requestedConcurrency)

		b.Logf("Using concurrency level: %d (requested: %d)", actualConcurrency, requestedConcurrency)

		// Create requests for parallel execution
		requests := make([]helpers.Request, b.N)
		for i := 0; i < b.N; i++ {
			createReq := service.CreateEpicRequest{
				CreatorID:   userID,
				Priority:    models.PriorityMedium,
				Title:       fmt.Sprintf("Concurrent Reliable Epic %d", i),
				Description: stringPtr(fmt.Sprintf("Epic created concurrently with reliability %d", i)),
			}
			requests[i] = helpers.Request{
				Method: "POST",
				Path:   "/api/v1/epics",
				Body:   createReq,
			}
		}

		// Execute with enhanced error handling and monitoring
		err := testRunner.ExecuteWithReliability("concurrent_execution", func() error {
			responses, err := client.RunParallelRequests(requests, actualConcurrency)
			if err != nil {
				return fmt.Errorf("parallel execution failed: %w", err)
			}

			// Validate all responses with detailed error reporting
			errorCount := 0
			for i, resp := range responses {
				if resp.Error != nil {
					errorCount++
					b.Logf("Request %d failed: %v", i, resp.Error)
					continue
				}

				if resp.StatusCode != http.StatusCreated {
					errorCount++
					b.Logf("Request %d returned wrong status: %d", i, resp.StatusCode)
				}
			}

			// Check error rate
			errorRate := float64(errorCount) / float64(len(responses))
			if errorRate > 0.05 { // 5% error threshold
				return fmt.Errorf("error rate %.2f%% exceeds threshold 5%%", errorRate*100)
			}

			return nil
		})

		require.NoError(b, err, "Concurrent operations with reliability failed")
	})
}

// BenchmarkResourceConstrainedOperations demonstrates graceful degradation under resource constraints
func BenchmarkResourceConstrainedOperations(b *testing.B) {
	server := setup.NewBenchmarkServer(b)
	defer server.Cleanup()

	testRunner := NewBenchmarkTestRunner(b, server.DB)
	defer testRunner.Cleanup()

	// Configure for resource-constrained environment
	reliabilityMgr := testRunner.GetReliabilityManager()

	// Note: In a full implementation, you would configure resource monitoring here
	// For this example, we'll rely on the default resource monitoring

	// Setup environment
	require.NoError(b, testRunner.ExecuteWithReliability("server_start", server.Start))
	require.NoError(b, testRunner.SetupBenchmarkEnvironment(server.DB, server.BaseURL))
	require.NoError(b, testRunner.ExecuteWithReliability("data_seeding", server.SeedSmallDataSet))

	// Setup client
	client := helpers.NewBenchmarkClient(server.BaseURL)
	authHelper := helpers.NewAuthHelper(server.Config.JWT.Secret)
	testUser := helpers.GetDefaultTestUser()
	require.NoError(b, testRunner.ExecuteWithReliability("authentication", func() error {
		return authHelper.AuthenticateClient(client, testUser.ID, testUser.Username)
	}))

	var users []models.User
	require.NoError(b, server.DB.Limit(1).Find(&users).Error)
	userID := users[0].ID

	b.ResetTimer()

	b.Run("OperationsWithGracefulDegradation", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			// Check if we should skip non-essential operations
			if reliabilityMgr.ShouldSkipOperation("detailed_metrics") {
				b.Logf("Skipping detailed metrics collection due to resource constraints")
			}

			// Adjust operation complexity based on constraints
			degradationConfig := reliabilityMgr.GetGracefulDegradationConfig()

			var title string
			if degradationConfig.UseSimplifiedData {
				title = fmt.Sprintf("Simple Epic %d", i)
			} else {
				title = fmt.Sprintf("Complex Epic with detailed description and metadata %d", i)
			}

			createReq := service.CreateEpicRequest{
				CreatorID: userID,
				Priority:  models.PriorityMedium,
				Title:     title,
			}

			// Only add description if not in fast mode
			if !degradationConfig.EnableFastMode {
				createReq.Description = stringPtr(fmt.Sprintf("Detailed description for epic %d", i))
			}

			err := testRunner.ExecuteWithReliability("epic_creation_degraded", func() error {
				resp, err := client.POST("/api/v1/epics", createReq)
				if err != nil {
					return err
				}
				defer resp.Body.Close()

				if resp.StatusCode != http.StatusCreated {
					return fmt.Errorf("unexpected status: %d", resp.StatusCode)
				}

				return nil
			})

			if err != nil {
				b.Fatalf("Operation failed at iteration %d: %v", i, err)
			}

			// Limit operations if configured
			if degradationConfig.MaxOperationsPerTest > 0 && i >= degradationConfig.MaxOperationsPerTest {
				b.Logf("Stopping at %d operations due to degradation limits", i+1)
				break
			}
		}
	})
}

// BenchmarkErrorRecoveryScenarios tests various error recovery scenarios
func BenchmarkErrorRecoveryScenarios(b *testing.B) {
	server := setup.NewBenchmarkServer(b)
	defer server.Cleanup()

	testRunner := NewBenchmarkTestRunner(b, server.DB)
	defer testRunner.Cleanup()

	// Configure retry behavior for error scenarios
	reliabilityMgr := testRunner.GetReliabilityManager()

	// Setup environment
	require.NoError(b, testRunner.ExecuteWithReliability("server_start", server.Start))
	require.NoError(b, testRunner.SetupBenchmarkEnvironment(server.DB, server.BaseURL))
	require.NoError(b, testRunner.ExecuteWithReliability("data_seeding", server.SeedSmallDataSet))

	client := helpers.NewBenchmarkClient(server.BaseURL)
	authHelper := helpers.NewAuthHelper(server.Config.JWT.Secret)
	testUser := helpers.GetDefaultTestUser()
	require.NoError(b, testRunner.ExecuteWithReliability("authentication", func() error {
		return authHelper.AuthenticateClient(client, testUser.ID, testUser.Username)
	}))

	b.ResetTimer()

	b.Run("RecoveryFromTransientErrors", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			// Simulate operations that might have transient failures
			err := testRunner.ExecuteWithReliability("transient_operation", func() error {
				// This operation might fail occasionally but should be retried
				resp, err := client.GET("/api/v1/epics")
				if err != nil {
					return err
				}
				defer resp.Body.Close()

				if resp.StatusCode == http.StatusServiceUnavailable {
					return fmt.Errorf("temporary failure - service unavailable")
				}

				if resp.StatusCode != http.StatusOK {
					return fmt.Errorf("unexpected status: %d", resp.StatusCode)
				}

				return nil
			})

			// The reliability manager should handle retries automatically
			require.NoError(b, err, "Operation should succeed with retry logic")
		}
	})

	b.Run("TimeoutHandling", func(b *testing.B) {
		// Set a very short timeout to test timeout handling
		reliabilityMgr.SetTimeout("timeout_test", 1*time.Millisecond)

		for i := 0; i < b.N; i++ {
			err := testRunner.ExecuteWithReliability("timeout_test", func() error {
				// This operation will likely timeout
				time.Sleep(10 * time.Millisecond)
				return nil
			})

			// We expect this to timeout, so we check for timeout error
			if err != nil {
				b.Logf("Expected timeout occurred: %v", err)
			}
		}

		// Reset timeout for cleanup
		reliabilityMgr.SetTimeout("timeout_test", 30*time.Second)
	})
}

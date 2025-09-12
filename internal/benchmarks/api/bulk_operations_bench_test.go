package api

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"product-requirements-management/internal/benchmarks/helpers"
	"product-requirements-management/internal/benchmarks/setup"
	"product-requirements-management/internal/models"
	"product-requirements-management/internal/service"
)

// BenchmarkBulkEpicCreation tests bulk Epic creation via API endpoints
func BenchmarkBulkEpicCreation(b *testing.B) {
	// Setup benchmark server
	server := setup.NewBenchmarkServer(b)
	defer server.Cleanup()

	// Start the server
	require.NoError(b, server.Start())

	// Seed small dataset for basic requirements
	require.NoError(b, server.SeedSmallDataSet())

	// Create HTTP client
	client := helpers.NewBenchmarkClient(server.BaseURL)

	// Setup authentication
	authHelper := helpers.NewAuthHelper(server.Config.JWT.Secret)
	testUser := helpers.GetDefaultTestUser()
	require.NoError(b, authHelper.AuthenticateClient(client, testUser.ID, testUser.Username))

	// Get a test user ID for epic creation
	var users []models.User
	require.NoError(b, server.DB.Limit(1).Find(&users).Error)
	require.NotEmpty(b, users)
	userID := users[0].ID

	b.ResetTimer()

	b.Run("BulkCreateEpics_10", func(b *testing.B) {
		batchSize := 10
		for i := 0; i < b.N; i++ {
			// Create batch of epic creation requests
			requests := make([]helpers.Request, batchSize)
			for j := 0; j < batchSize; j++ {
				createReq := service.CreateEpicRequest{
					CreatorID:   userID,
					Priority:    models.PriorityMedium,
					Title:       fmt.Sprintf("Bulk Epic %d-%d", i, j),
					Description: stringPtr(fmt.Sprintf("Bulk created epic %d in batch %d", j, i)),
				}
				requests[j] = helpers.Request{
					Method: "POST",
					Path:   "/api/v1/epics",
					Body:   createReq,
				}
			}

			// Execute batch requests with limited concurrency
			concurrency := 5
			responses, err := client.RunParallelRequests(requests, concurrency)
			require.NoError(b, err)

			// Verify all requests succeeded
			for idx, resp := range responses {
				require.NoError(b, resp.Error, "Request %d failed", idx)
				require.Equal(b, http.StatusCreated, resp.StatusCode, "Request %d returned wrong status", idx)
			}
		}
	})

	b.Run("BulkCreateEpics_50", func(b *testing.B) {
		batchSize := 50
		for i := 0; i < b.N; i++ {
			// Create batch of epic creation requests
			requests := make([]helpers.Request, batchSize)
			for j := 0; j < batchSize; j++ {
				createReq := service.CreateEpicRequest{
					CreatorID:   userID,
					Priority:    models.PriorityMedium,
					Title:       fmt.Sprintf("Bulk Epic %d-%d", i, j),
					Description: stringPtr(fmt.Sprintf("Bulk created epic %d in batch %d", j, i)),
				}
				requests[j] = helpers.Request{
					Method: "POST",
					Path:   "/api/v1/epics",
					Body:   createReq,
				}
			}

			// Execute batch requests with limited concurrency
			concurrency := 10
			responses, err := client.RunParallelRequests(requests, concurrency)
			require.NoError(b, err)

			// Verify all requests succeeded
			for idx, resp := range responses {
				require.NoError(b, resp.Error, "Request %d failed", idx)
				require.Equal(b, http.StatusCreated, resp.StatusCode, "Request %d returned wrong status", idx)
			}
		}
	})
}

// BenchmarkBulkUserStoryCreation tests bulk User Story creation via API endpoints
func BenchmarkBulkUserStoryCreation(b *testing.B) {
	// Setup benchmark server
	server := setup.NewBenchmarkServer(b)
	defer server.Cleanup()

	// Start the server
	require.NoError(b, server.Start())

	// Seed small dataset for basic requirements
	require.NoError(b, server.SeedSmallDataSet())

	// Create HTTP client
	client := helpers.NewBenchmarkClient(server.BaseURL)

	// Setup authentication
	authHelper := helpers.NewAuthHelper(server.Config.JWT.Secret)
	testUser := helpers.GetDefaultTestUser()
	require.NoError(b, authHelper.AuthenticateClient(client, testUser.ID, testUser.Username))

	// Get test data for user story creation
	var users []models.User
	require.NoError(b, server.DB.Limit(1).Find(&users).Error)
	require.NotEmpty(b, users)
	userID := users[0].ID

	var epics []models.Epic
	require.NoError(b, server.DB.Limit(1).Find(&epics).Error)
	require.NotEmpty(b, epics)
	epicID := epics[0].ID

	b.ResetTimer()

	b.Run("BulkCreateUserStories_25", func(b *testing.B) {
		batchSize := 25
		for i := 0; i < b.N; i++ {
			// Create batch of user story creation requests
			requests := make([]helpers.Request, batchSize)
			for j := 0; j < batchSize; j++ {
				createReq := service.CreateUserStoryRequest{
					EpicID:      epicID,
					CreatorID:   userID,
					Priority:    models.PriorityMedium,
					Title:       fmt.Sprintf("Bulk User Story %d-%d", i, j),
					Description: stringPtr(fmt.Sprintf("As a user, I want bulk feature %d-%d, so that I can test bulk performance", i, j)),
				}
				requests[j] = helpers.Request{
					Method: "POST",
					Path:   "/api/v1/user-stories",
					Body:   createReq,
				}
			}

			// Execute batch requests with limited concurrency
			concurrency := 8
			responses, err := client.RunParallelRequests(requests, concurrency)
			require.NoError(b, err)

			// Verify all requests succeeded
			for idx, resp := range responses {
				require.NoError(b, resp.Error, "Request %d failed", idx)
				require.Equal(b, http.StatusCreated, resp.StatusCode, "Request %d returned wrong status", idx)
			}
		}
	})
}

// BenchmarkBulkRequirementCreation tests bulk Requirement creation via API endpoints
func BenchmarkBulkRequirementCreation(b *testing.B) {
	// Setup benchmark server
	server := setup.NewBenchmarkServer(b)
	defer server.Cleanup()

	// Start the server
	require.NoError(b, server.Start())

	// Seed medium dataset for requirements
	require.NoError(b, server.SeedMediumDataSet())

	// Create HTTP client
	client := helpers.NewBenchmarkClient(server.BaseURL)

	// Setup authentication
	authHelper := helpers.NewAuthHelper(server.Config.JWT.Secret)
	testUser := helpers.GetDefaultTestUser()
	require.NoError(b, authHelper.AuthenticateClient(client, testUser.ID, testUser.Username))

	// Get test data for requirement creation
	var users []models.User
	require.NoError(b, server.DB.Limit(1).Find(&users).Error)
	require.NotEmpty(b, users)
	userID := users[0].ID

	var userStories []models.UserStory
	require.NoError(b, server.DB.Limit(1).Find(&userStories).Error)
	require.NotEmpty(b, userStories)
	userStoryID := userStories[0].ID

	var requirementTypes []models.RequirementType
	require.NoError(b, server.DB.Limit(1).Find(&requirementTypes).Error)
	require.NotEmpty(b, requirementTypes)
	requirementTypeID := requirementTypes[0].ID

	b.ResetTimer()

	b.Run("BulkCreateRequirements_50", func(b *testing.B) {
		batchSize := 50
		for i := 0; i < b.N; i++ {
			// Create batch of requirement creation requests
			requests := make([]helpers.Request, batchSize)
			for j := 0; j < batchSize; j++ {
				createReq := service.CreateRequirementRequest{
					UserStoryID: userStoryID,
					CreatorID:   userID,
					TypeID:      requirementTypeID,
					Priority:    models.PriorityMedium,
					Title:       fmt.Sprintf("Bulk Requirement %d-%d", i, j),
					Description: stringPtr(fmt.Sprintf("Bulk created requirement %d in batch %d", j, i)),
				}
				requests[j] = helpers.Request{
					Method: "POST",
					Path:   "/api/v1/requirements",
					Body:   createReq,
				}
			}

			// Execute batch requests with limited concurrency
			concurrency := 10
			responses, err := client.RunParallelRequests(requests, concurrency)
			require.NoError(b, err)

			// Verify all requests succeeded
			for idx, resp := range responses {
				require.NoError(b, resp.Error, "Request %d failed", idx)
				require.Equal(b, http.StatusCreated, resp.StatusCode, "Request %d returned wrong status", idx)
			}
		}
	})
}

// BenchmarkBulkEpicUpdates tests bulk Epic update operations via API endpoints
func BenchmarkBulkEpicUpdates(b *testing.B) {
	// Setup benchmark server
	server := setup.NewBenchmarkServer(b)
	defer server.Cleanup()

	// Start the server
	require.NoError(b, server.Start())

	// Seed medium dataset for update operations
	require.NoError(b, server.SeedMediumDataSet())

	// Create HTTP client
	client := helpers.NewBenchmarkClient(server.BaseURL)

	// Setup authentication
	authHelper := helpers.NewAuthHelper(server.Config.JWT.Secret)
	testUser := helpers.GetDefaultTestUser()
	require.NoError(b, authHelper.AuthenticateClient(client, testUser.ID, testUser.Username))

	// Get existing epics for update operations
	var epics []models.Epic
	require.NoError(b, server.DB.Limit(200).Find(&epics).Error)
	require.NotEmpty(b, epics)

	// Validate test data
	validator := NewBenchmarkDataValidator(b)
	validator.ValidateMinimumCount(len(epics), 50, "epics")

	b.ResetTimer()

	b.Run("BulkUpdateEpics_20", func(b *testing.B) {
		batchSize := 20
		for i := 0; i < b.N; i++ {
			// Create batch of epic update requests
			requests := make([]helpers.Request, batchSize)
			for j := 0; j < batchSize; j++ {
				epicIndex := safeIndex((i*batchSize)+j, len(epics))
				epic := epics[epicIndex]

				updateReq := service.UpdateEpicRequest{
					Title:       stringPtr(fmt.Sprintf("Bulk Updated Epic %d-%d", i, j)),
					Description: stringPtr(fmt.Sprintf("Bulk updated description %d-%d", i, j)),
				}
				requests[j] = helpers.Request{
					Method: "PUT",
					Path:   fmt.Sprintf("/api/v1/epics/%s", epic.ID),
					Body:   updateReq,
				}
			}

			// Execute batch requests with limited concurrency
			concurrency := 8
			responses, err := client.RunParallelRequests(requests, concurrency)
			require.NoError(b, err)

			// Verify all requests succeeded
			for idx, resp := range responses {
				require.NoError(b, resp.Error, "Request %d failed", idx)
				require.Equal(b, http.StatusOK, resp.StatusCode, "Request %d returned wrong status", idx)
			}
		}
	})
}

// BenchmarkBulkEpicDeletion tests bulk Epic deletion operations via API endpoints
func BenchmarkBulkEpicDeletion(b *testing.B) {
	// Setup benchmark server
	server := setup.NewBenchmarkServer(b)
	defer server.Cleanup()

	// Start the server
	require.NoError(b, server.Start())

	// Seed small dataset for basic requirements
	require.NoError(b, server.SeedSmallDataSet())

	// Create HTTP client
	client := helpers.NewBenchmarkClient(server.BaseURL)

	// Setup authentication
	authHelper := helpers.NewAuthHelper(server.Config.JWT.Secret)
	testUser := helpers.GetDefaultTestUser()
	require.NoError(b, authHelper.AuthenticateClient(client, testUser.ID, testUser.Username))

	// Get a test user ID for epic creation
	var users []models.User
	require.NoError(b, server.DB.Limit(1).Find(&users).Error)
	require.NotEmpty(b, users)
	userID := users[0].ID

	b.ResetTimer()

	b.Run("BulkDeleteEpics_15", func(b *testing.B) {
		batchSize := 15
		for i := 0; i < b.N; i++ {
			// First create epics to delete
			epicIDs := make([]uuid.UUID, batchSize)
			for j := 0; j < batchSize; j++ {
				createReq := service.CreateEpicRequest{
					CreatorID:   userID,
					Priority:    models.PriorityLow,
					Title:       fmt.Sprintf("Epic to Delete %d-%d", i, j),
					Description: stringPtr("Epic created for bulk deletion benchmark"),
				}

				resp, err := client.POST("/api/v1/epics", createReq)
				require.NoError(b, err)
				require.Equal(b, http.StatusCreated, resp.StatusCode)

				var epic models.Epic
				require.NoError(b, helpers.ParseJSONResponse(resp, &epic))
				epicIDs[j] = epic.ID
			}

			// Create batch of epic deletion requests
			requests := make([]helpers.Request, batchSize)
			for j := 0; j < batchSize; j++ {
				requests[j] = helpers.Request{
					Method: "DELETE",
					Path:   fmt.Sprintf("/api/v1/epics/%s", epicIDs[j]),
				}
			}

			// Execute batch deletion requests with limited concurrency
			concurrency := 6
			responses, err := client.RunParallelRequests(requests, concurrency)
			require.NoError(b, err)

			// Verify all requests succeeded
			for idx, resp := range responses {
				require.NoError(b, resp.Error, "Request %d failed", idx)
				require.Equal(b, http.StatusNoContent, resp.StatusCode, "Request %d returned wrong status", idx)
			}
		}
	})
}

// BenchmarkLargeListRetrieval tests large list retrieval performance via API endpoints
func BenchmarkLargeListRetrieval(b *testing.B) {
	// Setup benchmark server
	server := setup.NewBenchmarkServer(b)
	defer server.Cleanup()

	// Start the server
	require.NoError(b, server.Start())

	// Seed large dataset for list retrieval operations
	require.NoError(b, server.SeedLargeDataSet())

	// Create HTTP client
	client := helpers.NewBenchmarkClient(server.BaseURL)

	// Setup authentication
	authHelper := helpers.NewAuthHelper(server.Config.JWT.Secret)
	testUser := helpers.GetDefaultTestUser()
	require.NoError(b, authHelper.AuthenticateClient(client, testUser.ID, testUser.Username))

	b.ResetTimer()

	b.Run("RetrieveAllEpics", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			resp, err := client.GET("/api/v1/epics")
			require.NoError(b, err)
			require.Equal(b, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		}
	})

	b.Run("RetrieveAllUserStories", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			resp, err := client.GET("/api/v1/user-stories")
			require.NoError(b, err)
			require.Equal(b, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		}
	})

	b.Run("RetrieveAllRequirements", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			resp, err := client.GET("/api/v1/requirements")
			require.NoError(b, err)
			require.Equal(b, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		}
	})

	b.Run("RetrieveLargeEpicList_Limit100", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			resp, err := client.GET("/api/v1/epics?limit=100")
			require.NoError(b, err)
			require.Equal(b, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		}
	})

	b.Run("RetrieveLargeRequirementList_Limit500", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			resp, err := client.GET("/api/v1/requirements?limit=500")
			require.NoError(b, err)
			require.Equal(b, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		}
	})
}

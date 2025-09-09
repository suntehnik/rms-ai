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

// BenchmarkEpicCRUD tests Epic CRUD operations via HTTP endpoints
func BenchmarkEpicCRUD(b *testing.B) {
	// Setup benchmark server
	server := setup.NewBenchmarkServer(b)
	defer server.Cleanup()

	// Start the server
	require.NoError(b, server.Start())

	// Seed small dataset for CRUD operations
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

	b.Run("CreateEpic", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			createReq := service.CreateEpicRequest{
				CreatorID:   userID,
				Priority:    models.PriorityMedium,
				Title:       fmt.Sprintf("Benchmark Epic %d", i),
				Description: stringPtr(fmt.Sprintf("Description for benchmark epic %d", i)),
			}

			resp, err := client.POST("/api/v1/epics", createReq)
			require.NoError(b, err)
			require.Equal(b, http.StatusCreated, resp.StatusCode)
			resp.Body.Close()
		}
	})

	// Create an epic for read/update/delete operations
	createReq := service.CreateEpicRequest{
		CreatorID:   userID,
		Priority:    models.PriorityHigh,
		Title:       "Test Epic for CRUD",
		Description: stringPtr("Test epic for read/update/delete operations"),
	}

	resp, err := client.POST("/api/v1/epics", createReq)
	require.NoError(b, err)
	require.Equal(b, http.StatusCreated, resp.StatusCode)

	var createdEpic models.Epic
	require.NoError(b, helpers.ParseJSONResponse(resp, &createdEpic))

	b.Run("GetEpic", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			resp, err := client.GET(fmt.Sprintf("/api/v1/epics/%s", createdEpic.ID))
			require.NoError(b, err)
			require.Equal(b, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		}
	})

	b.Run("UpdateEpic", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			updateReq := service.UpdateEpicRequest{
				Title: stringPtr(fmt.Sprintf("Updated Epic %d", i)),
			}

			resp, err := client.PUT(fmt.Sprintf("/api/v1/epics/%s", createdEpic.ID), updateReq)
			require.NoError(b, err)
			require.Equal(b, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		}
	})

	b.Run("DeleteEpic", func(b *testing.B) {
		// Create epics to delete
		epicIDs := make([]uuid.UUID, b.N)
		for i := 0; i < b.N; i++ {
			createReq := service.CreateEpicRequest{
				CreatorID:   userID,
				Priority:    models.PriorityLow,
				Title:       fmt.Sprintf("Epic to Delete %d", i),
				Description: stringPtr("Epic created for deletion benchmark"),
			}

			resp, err := client.POST("/api/v1/epics", createReq)
			require.NoError(b, err)
			require.Equal(b, http.StatusCreated, resp.StatusCode)

			var epic models.Epic
			require.NoError(b, helpers.ParseJSONResponse(resp, &epic))
			epicIDs[i] = epic.ID
		}

		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			resp, err := client.DELETE(fmt.Sprintf("/api/v1/epics/%s", epicIDs[i]))
			require.NoError(b, err)
			require.Equal(b, http.StatusNoContent, resp.StatusCode)
			resp.Body.Close()
		}
	})
}

// BenchmarkEpicListing tests Epic listing and filtering performance
func BenchmarkEpicListing(b *testing.B) {
	// Setup benchmark server
	server := setup.NewBenchmarkServer(b)
	defer server.Cleanup()

	// Start the server
	require.NoError(b, server.Start())

	// Seed medium dataset for listing operations
	require.NoError(b, server.SeedMediumDataSet())

	// Create HTTP client
	client := helpers.NewBenchmarkClient(server.BaseURL)

	// Setup authentication
	authHelper := helpers.NewAuthHelper(server.Config.JWT.Secret)
	testUser := helpers.GetDefaultTestUser()
	require.NoError(b, authHelper.AuthenticateClient(client, testUser.ID, testUser.Username))

	b.ResetTimer()

	b.Run("ListAllEpics", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			resp, err := client.GET("/api/v1/epics")
			require.NoError(b, err)
			require.Equal(b, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		}
	})

	b.Run("ListEpicsWithLimit", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			resp, err := client.GET("/api/v1/epics?limit=10")
			require.NoError(b, err)
			require.Equal(b, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		}
	})

	b.Run("ListEpicsByStatus", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			resp, err := client.GET("/api/v1/epics?status=Backlog")
			require.NoError(b, err)
			require.Equal(b, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		}
	})

	b.Run("ListEpicsByPriority", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			resp, err := client.GET("/api/v1/epics?priority=2")
			require.NoError(b, err)
			require.Equal(b, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		}
	})

	b.Run("ListEpicsWithPagination", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			offset := (i % 10) * 5 // Vary offset for different pages
			resp, err := client.GET(fmt.Sprintf("/api/v1/epics?limit=5&offset=%d", offset))
			require.NoError(b, err)
			require.Equal(b, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		}
	})

	b.Run("ListEpicsWithOrdering", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			resp, err := client.GET("/api/v1/epics?order_by=created_at")
			require.NoError(b, err)
			require.Equal(b, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		}
	})
}

// BenchmarkEpicStatusChange tests Epic status change performance
func BenchmarkEpicStatusChange(b *testing.B) {
	// Setup benchmark server
	server := setup.NewBenchmarkServer(b)
	defer server.Cleanup()

	// Start the server
	require.NoError(b, server.Start())

	// Seed small dataset
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

	// Create epics for status change testing
	epicIDs := make([]uuid.UUID, b.N)
	for i := 0; i < b.N; i++ {
		createReq := service.CreateEpicRequest{
			CreatorID:   userID,
			Priority:    models.PriorityMedium,
			Title:       fmt.Sprintf("Epic for Status Change %d", i),
			Description: stringPtr("Epic created for status change benchmark"),
		}

		resp, err := client.POST("/api/v1/epics", createReq)
		require.NoError(b, err)
		require.Equal(b, http.StatusCreated, resp.StatusCode)

		var epic models.Epic
		require.NoError(b, helpers.ParseJSONResponse(resp, &epic))
		epicIDs[i] = epic.ID
	}

	b.ResetTimer()

	b.Run("ChangeEpicStatus", func(b *testing.B) {
		statuses := []models.EpicStatus{
			models.EpicStatusInProgress,
			models.EpicStatusDone,
			models.EpicStatusBacklog,
		}

		for i := 0; i < b.N; i++ {
			statusReq := map[string]interface{}{
				"status": statuses[i%len(statuses)],
			}

			resp, err := client.PATCH(fmt.Sprintf("/api/v1/epics/%s/status", epicIDs[i]), statusReq)
			require.NoError(b, err)
			require.Equal(b, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		}
	})
}

// BenchmarkEpicAssignment tests Epic assignment performance
func BenchmarkEpicAssignment(b *testing.B) {
	// Setup benchmark server
	server := setup.NewBenchmarkServer(b)
	defer server.Cleanup()

	// Start the server
	require.NoError(b, server.Start())

	// Seed small dataset
	require.NoError(b, server.SeedSmallDataSet())

	// Create HTTP client
	client := helpers.NewBenchmarkClient(server.BaseURL)

	// Setup authentication
	authHelper := helpers.NewAuthHelper(server.Config.JWT.Secret)
	testUser := helpers.GetDefaultTestUser()
	require.NoError(b, authHelper.AuthenticateClient(client, testUser.ID, testUser.Username))

	// Get test users for assignment
	var users []models.User
	require.NoError(b, server.DB.Limit(5).Find(&users).Error)
	require.NotEmpty(b, users)

	// Create epics for assignment testing
	epicIDs := make([]uuid.UUID, b.N)
	for i := 0; i < b.N; i++ {
		createReq := service.CreateEpicRequest{
			CreatorID:   users[0].ID,
			Priority:    models.PriorityMedium,
			Title:       fmt.Sprintf("Epic for Assignment %d", i),
			Description: stringPtr("Epic created for assignment benchmark"),
		}

		resp, err := client.POST("/api/v1/epics", createReq)
		require.NoError(b, err)
		require.Equal(b, http.StatusCreated, resp.StatusCode)

		var epic models.Epic
		require.NoError(b, helpers.ParseJSONResponse(resp, &epic))
		epicIDs[i] = epic.ID
	}

	b.ResetTimer()

	b.Run("AssignEpic", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			assigneeID := users[i%len(users)].ID
			assignReq := map[string]interface{}{
				"assignee_id": assigneeID,
			}

			resp, err := client.PATCH(fmt.Sprintf("/api/v1/epics/%s/assign", epicIDs[i]), assignReq)
			require.NoError(b, err)
			require.Equal(b, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		}
	})
}

// BenchmarkEpicWithUserStories tests Epic retrieval with user stories via API endpoints
func BenchmarkEpicWithUserStories(b *testing.B) {
	// Setup benchmark server
	server := setup.NewBenchmarkServer(b)
	defer server.Cleanup()

	// Start the server
	require.NoError(b, server.Start())

	// Seed medium dataset to have epics with user stories
	require.NoError(b, server.SeedMediumDataSet())

	// Create HTTP client
	client := helpers.NewBenchmarkClient(server.BaseURL)

	// Setup authentication
	authHelper := helpers.NewAuthHelper(server.Config.JWT.Secret)
	testUser := helpers.GetDefaultTestUser()
	require.NoError(b, authHelper.AuthenticateClient(client, testUser.ID, testUser.Username))

	// Get epics that have user stories
	var epics []models.Epic
	require.NoError(b, server.DB.Preload("UserStories").Find(&epics).Error)
	require.NotEmpty(b, epics)

	// Filter epics that actually have user stories
	var epicsWithStories []models.Epic
	for _, epic := range epics {
		if len(epic.UserStories) > 0 {
			epicsWithStories = append(epicsWithStories, epic)
		}
	}
	require.NotEmpty(b, epicsWithStories)

	b.ResetTimer()

	b.Run("GetEpicWithUserStories", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			epic := epicsWithStories[i%len(epicsWithStories)]
			resp, err := client.GET(fmt.Sprintf("/api/v1/epics/%s/user-stories", epic.ID))
			require.NoError(b, err)
			require.Equal(b, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		}
	})

	b.Run("GetEpicWithUserStoriesByReferenceID", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			epic := epicsWithStories[i%len(epicsWithStories)]
			resp, err := client.GET(fmt.Sprintf("/api/v1/epics/%s/user-stories", epic.ReferenceID))
			require.NoError(b, err)
			require.Equal(b, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		}
	})
}

// BenchmarkEpicConcurrentOperations tests concurrent Epic operations
func BenchmarkEpicConcurrentOperations(b *testing.B) {
	// Setup benchmark server
	server := setup.NewBenchmarkServer(b)
	defer server.Cleanup()

	// Start the server
	require.NoError(b, server.Start())

	// Seed small dataset
	require.NoError(b, server.SeedSmallDataSet())

	// Create HTTP client
	client := helpers.NewBenchmarkClient(server.BaseURL)

	// Setup authentication
	authHelper := helpers.NewAuthHelper(server.Config.JWT.Secret)
	testUser := helpers.GetDefaultTestUser()
	require.NoError(b, authHelper.AuthenticateClient(client, testUser.ID, testUser.Username))

	// Get a test user ID
	var users []models.User
	require.NoError(b, server.DB.Limit(1).Find(&users).Error)
	require.NotEmpty(b, users)
	userID := users[0].ID

	b.ResetTimer()

	b.Run("ConcurrentEpicCreation", func(b *testing.B) {
		// Create requests for parallel execution
		requests := make([]helpers.Request, b.N)
		for i := 0; i < b.N; i++ {
			createReq := service.CreateEpicRequest{
				CreatorID:   userID,
				Priority:    models.PriorityMedium,
				Title:       fmt.Sprintf("Concurrent Epic %d", i),
				Description: stringPtr(fmt.Sprintf("Epic created concurrently %d", i)),
			}
			requests[i] = helpers.Request{
				Method: "POST",
				Path:   "/api/v1/epics",
				Body:   createReq,
			}
		}

		// Execute requests with limited concurrency
		concurrency := 10
		responses, err := client.RunParallelRequests(requests, concurrency)
		require.NoError(b, err)

		// Verify all requests succeeded
		for i, resp := range responses {
			require.NoError(b, resp.Error, "Request %d failed", i)
			require.Equal(b, http.StatusCreated, resp.StatusCode, "Request %d returned wrong status", i)
		}
	})

	b.Run("ConcurrentEpicReads", func(b *testing.B) {
		// Get some existing epics
		var epics []models.Epic
		require.NoError(b, server.DB.Limit(10).Find(&epics).Error)
		require.NotEmpty(b, epics)

		// Create read requests for parallel execution
		requests := make([]helpers.Request, b.N)
		for i := 0; i < b.N; i++ {
			epic := epics[i%len(epics)]
			requests[i] = helpers.Request{
				Method: "GET",
				Path:   fmt.Sprintf("/api/v1/epics/%s", epic.ID),
			}
		}

		// Execute requests with limited concurrency
		concurrency := 20
		responses, err := client.RunParallelRequests(requests, concurrency)
		require.NoError(b, err)

		// Verify all requests succeeded
		for i, resp := range responses {
			require.NoError(b, resp.Error, "Request %d failed", i)
			require.Equal(b, http.StatusOK, resp.StatusCode, "Request %d returned wrong status", i)
		}
	})
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}
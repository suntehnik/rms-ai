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
	require.NotEmpty(b, users, "No users available for epic creation")
	userID := users[0].ID

	// Validate test data availability before proceeding
	if b.N <= 0 {
		b.Fatal("Invalid benchmark iteration count: b.N must be greater than 0")
	}

	// Create a reasonable number of epics for status change testing
	// Use a minimum of 10 epics or b.N, whichever is smaller, to avoid creating too many
	numEpics := b.N
	if numEpics > 100 {
		numEpics = 100 // Cap at 100 to avoid excessive setup time
	}
	if numEpics < 10 {
		numEpics = 10 // Minimum of 10 for reasonable testing
	}

	// Validate that we can create the required number of epics
	if numEpics <= 0 {
		b.Fatal("Invalid number of epics to create: must be greater than 0")
	}

	epicIDs := make([]uuid.UUID, 0, numEpics) // Use slice with capacity for better memory management

	// Create epics with proper error handling and validation
	for i := 0; i < numEpics; i++ {
		createReq := service.CreateEpicRequest{
			CreatorID:   userID,
			Priority:    models.PriorityMedium,
			Title:       fmt.Sprintf("Epic for Status Change %d", i),
			Description: stringPtr("Epic created for status change benchmark"),
		}

		resp, err := client.POST("/api/v1/epics", createReq)
		if err != nil {
			b.Fatalf("Failed to create epic %d: %v", i, err)
		}
		if resp.StatusCode != http.StatusCreated {
			resp.Body.Close()
			b.Fatalf("Failed to create epic %d: expected status %d, got %d", i, http.StatusCreated, resp.StatusCode)
		}

		var epic models.Epic
		if err := helpers.ParseJSONResponse(resp, &epic); err != nil {
			resp.Body.Close()
			b.Fatalf("Failed to parse epic response %d: %v", i, err)
		}

		// Validate that the epic ID is not nil
		if epic.ID == uuid.Nil {
			b.Fatalf("Created epic %d has nil UUID", i)
		}

		epicIDs = append(epicIDs, epic.ID)
	}

	// Validate test data using the validator
	validator := NewBenchmarkDataValidator(b)
	validator.ValidateUUIDs(epicIDs, "epic")
	validator.ValidateMinimumCount(len(epicIDs), 1, "epics")

	// Additional validation for test data availability
	if len(epicIDs) == 0 {
		b.Fatal("No epics were successfully created for benchmark testing")
	}

	b.ResetTimer()

	b.Run("ChangeEpicStatus", func(b *testing.B) {
		// Define available statuses for cycling
		statuses := []models.EpicStatus{
			models.EpicStatusInProgress,
			models.EpicStatusDone,
			models.EpicStatusBacklog,
		}

		// Validate that we have statuses to work with
		if len(statuses) == 0 {
			b.Fatal("No epic statuses available for testing")
		}

		// Validate that we have epics to work with before starting benchmark
		if len(epicIDs) == 0 {
			b.Fatal("No epic IDs available for status change benchmark")
		}

		for i := 0; i < b.N; i++ {
			// Use safe indexing with bounds checking to cycle through available epics
			epicIndex := safeIndex(i, len(epicIDs))
			statusIndex := safeIndex(i, len(statuses))

			// Additional bounds checking as defensive programming
			if epicIndex < 0 || epicIndex >= len(epicIDs) {
				b.Fatalf("Epic index out of bounds: %d (available: %d)", epicIndex, len(epicIDs))
			}
			if statusIndex < 0 || statusIndex >= len(statuses) {
				b.Fatalf("Status index out of bounds: %d (available: %d)", statusIndex, len(statuses))
			}

			// Validate epic ID before using it
			epicID := epicIDs[epicIndex]
			if epicID == uuid.Nil {
				b.Fatalf("Epic ID at index %d is nil", epicIndex)
			}

			statusReq := map[string]interface{}{
				"status": statuses[statusIndex],
			}

			resp, err := client.PATCH(fmt.Sprintf("/api/v1/epics/%s/status", epicID), statusReq)
			if err != nil {
				b.Fatalf("Failed to change epic status (iteration %d, epic %s): %v", i, epicID, err)
			}

			if resp.StatusCode != http.StatusOK {
				resp.Body.Close()
				b.Fatalf("Epic status change failed (iteration %d, epic %s): expected status %d, got %d",
					i, epicID, http.StatusOK, resp.StatusCode)
			}

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
	require.NotEmpty(b, users, "No users available for epic assignment")

	// Validate test data availability before proceeding
	if b.N <= 0 {
		b.Fatal("Invalid benchmark iteration count: b.N must be greater than 0")
	}

	// Create a reasonable number of epics for assignment testing
	numEpics := b.N
	if numEpics > 100 {
		numEpics = 100 // Cap at 100 to avoid excessive setup time
	}
	if numEpics < 10 {
		numEpics = 10 // Minimum of 10 for reasonable testing
	}

	// Validate that we can create the required number of epics
	if numEpics <= 0 {
		b.Fatal("Invalid number of epics to create: must be greater than 0")
	}

	epicIDs := make([]uuid.UUID, 0, numEpics) // Use slice with capacity for better memory management

	// Create epics with proper error handling and validation
	for i := 0; i < numEpics; i++ {
		createReq := service.CreateEpicRequest{
			CreatorID:   users[0].ID,
			Priority:    models.PriorityMedium,
			Title:       fmt.Sprintf("Epic for Assignment %d", i),
			Description: stringPtr("Epic created for assignment benchmark"),
		}

		resp, err := client.POST("/api/v1/epics", createReq)
		if err != nil {
			b.Fatalf("Failed to create epic %d: %v", i, err)
		}
		if resp.StatusCode != http.StatusCreated {
			resp.Body.Close()
			b.Fatalf("Failed to create epic %d: expected status %d, got %d", i, http.StatusCreated, resp.StatusCode)
		}

		var epic models.Epic
		if err := helpers.ParseJSONResponse(resp, &epic); err != nil {
			resp.Body.Close()
			b.Fatalf("Failed to parse epic response %d: %v", i, err)
		}

		// Validate that the epic ID is not nil
		if epic.ID == uuid.Nil {
			b.Fatalf("Created epic %d has nil UUID", i)
		}

		epicIDs = append(epicIDs, epic.ID)
	}

	// Validate test data using the validator
	validator := NewBenchmarkDataValidator(b)
	validator.ValidateUUIDs(epicIDs, "epic")
	validator.ValidateMinimumCount(len(epicIDs), 1, "epics")
	validator.ValidateMinimumCount(len(users), 1, "users")

	// Additional validation for test data availability
	if len(epicIDs) == 0 {
		b.Fatal("No epics were successfully created for benchmark testing")
	}
	if len(users) == 0 {
		b.Fatal("No users available for epic assignment")
	}

	b.ResetTimer()

	b.Run("AssignEpic", func(b *testing.B) {
		// Validate that we have data to work with before starting benchmark
		if len(epicIDs) == 0 {
			b.Fatal("No epic IDs available for assignment benchmark")
		}
		if len(users) == 0 {
			b.Fatal("No users available for assignment benchmark")
		}

		for i := 0; i < b.N; i++ {
			// Use safe indexing with bounds checking to cycle through available epics and users
			epicIndex := safeIndex(i, len(epicIDs))
			userIndex := safeIndex(i, len(users))

			// Additional bounds checking as defensive programming
			if epicIndex < 0 || epicIndex >= len(epicIDs) {
				b.Fatalf("Epic index out of bounds: %d (available: %d)", epicIndex, len(epicIDs))
			}
			if userIndex < 0 || userIndex >= len(users) {
				b.Fatalf("User index out of bounds: %d (available: %d)", userIndex, len(users))
			}

			// Validate IDs before using them
			epicID := epicIDs[epicIndex]
			if epicID == uuid.Nil {
				b.Fatalf("Epic ID at index %d is nil", epicIndex)
			}

			assigneeID := users[userIndex].ID
			if assigneeID == uuid.Nil {
				b.Fatalf("User ID at index %d is nil", userIndex)
			}

			assignReq := map[string]interface{}{
				"assignee_id": assigneeID,
			}

			resp, err := client.PATCH(fmt.Sprintf("/api/v1/epics/%s/assign", epicID), assignReq)
			if err != nil {
				b.Fatalf("Failed to assign epic (iteration %d, epic %s): %v", i, epicID, err)
			}

			if resp.StatusCode != http.StatusOK {
				resp.Body.Close()
				b.Fatalf("Epic assignment failed (iteration %d, epic %s): expected status %d, got %d",
					i, epicID, http.StatusOK, resp.StatusCode)
			}

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

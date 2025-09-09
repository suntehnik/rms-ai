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

// BenchmarkUserStoryCRUD tests User Story CRUD operations via HTTP endpoints
func BenchmarkUserStoryCRUD(b *testing.B) {
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

	b.Run("CreateUserStory", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			createReq := service.CreateUserStoryRequest{
				EpicID:      epicID,
				CreatorID:   userID,
				Priority:    models.PriorityMedium,
				Title:       fmt.Sprintf("Benchmark User Story %d", i),
				Description: stringPtr(fmt.Sprintf("As a user, I want benchmark feature %d, so that I can test performance", i)),
			}

			resp, err := client.POST("/api/v1/user-stories", createReq)
			require.NoError(b, err)
			require.Equal(b, http.StatusCreated, resp.StatusCode)
			resp.Body.Close()
		}
	})

	// Create a user story for read/update/delete operations
	createReq := service.CreateUserStoryRequest{
		EpicID:      epicID,
		CreatorID:   userID,
		Priority:    models.PriorityHigh,
		Title:       "Test User Story for CRUD",
		Description: stringPtr("As a tester, I want CRUD operations, so that I can benchmark performance"),
	}

	resp, err := client.POST("/api/v1/user-stories", createReq)
	require.NoError(b, err)
	require.Equal(b, http.StatusCreated, resp.StatusCode)

	var createdUserStory models.UserStory
	require.NoError(b, helpers.ParseJSONResponse(resp, &createdUserStory))

	b.Run("GetUserStory", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			resp, err := client.GET(fmt.Sprintf("/api/v1/user-stories/%s", createdUserStory.ID))
			require.NoError(b, err)
			require.Equal(b, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		}
	})

	b.Run("UpdateUserStory", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			updateReq := service.UpdateUserStoryRequest{
				Title: stringPtr(fmt.Sprintf("Updated User Story %d", i)),
			}

			resp, err := client.PUT(fmt.Sprintf("/api/v1/user-stories/%s", createdUserStory.ID), updateReq)
			require.NoError(b, err)
			require.Equal(b, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		}
	})

	b.Run("DeleteUserStory", func(b *testing.B) {
		// Create user stories to delete
		userStoryIDs := make([]uuid.UUID, b.N)
		for i := 0; i < b.N; i++ {
			createReq := service.CreateUserStoryRequest{
				EpicID:      epicID,
				CreatorID:   userID,
				Priority:    models.PriorityLow,
				Title:       fmt.Sprintf("User Story to Delete %d", i),
				Description: stringPtr("As a user, I want to be deleted, so that benchmarks can measure deletion performance"),
			}

			resp, err := client.POST("/api/v1/user-stories", createReq)
			require.NoError(b, err)
			require.Equal(b, http.StatusCreated, resp.StatusCode)

			var userStory models.UserStory
			require.NoError(b, helpers.ParseJSONResponse(resp, &userStory))
			userStoryIDs[i] = userStory.ID
		}

		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			resp, err := client.DELETE(fmt.Sprintf("/api/v1/user-stories/%s", userStoryIDs[i]))
			require.NoError(b, err)
			require.Equal(b, http.StatusNoContent, resp.StatusCode)
			resp.Body.Close()
		}
	})
}

// BenchmarkUserStoryListing tests User Story listing and filtering performance
func BenchmarkUserStoryListing(b *testing.B) {
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

	b.Run("ListAllUserStories", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			resp, err := client.GET("/api/v1/user-stories")
			require.NoError(b, err)
			require.Equal(b, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		}
	})

	b.Run("ListUserStoriesWithLimit", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			resp, err := client.GET("/api/v1/user-stories?limit=10")
			require.NoError(b, err)
			require.Equal(b, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		}
	})

	b.Run("ListUserStoriesByStatus", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			resp, err := client.GET("/api/v1/user-stories?status=Backlog")
			require.NoError(b, err)
			require.Equal(b, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		}
	})

	b.Run("ListUserStoriesByPriority", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			resp, err := client.GET("/api/v1/user-stories?priority=2")
			require.NoError(b, err)
			require.Equal(b, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		}
	})

	b.Run("ListUserStoriesWithPagination", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			offset := (i % 10) * 5 // Vary offset for different pages
			resp, err := client.GET(fmt.Sprintf("/api/v1/user-stories?limit=5&offset=%d", offset))
			require.NoError(b, err)
			require.Equal(b, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		}
	})

	b.Run("ListUserStoriesWithOrdering", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			resp, err := client.GET("/api/v1/user-stories?order_by=created_at")
			require.NoError(b, err)
			require.Equal(b, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		}
	})

	b.Run("ListUserStoriesByEpic", func(b *testing.B) {
		// Get an epic ID for filtering
		var epics []models.Epic
		require.NoError(b, server.DB.Limit(1).Find(&epics).Error)
		require.NotEmpty(b, epics)
		epicID := epics[0].ID

		for i := 0; i < b.N; i++ {
			resp, err := client.GET(fmt.Sprintf("/api/v1/user-stories?epic_id=%s", epicID))
			require.NoError(b, err)
			require.Equal(b, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		}
	})
}

// BenchmarkUserStoryStatusTransition tests User Story status transition performance
func BenchmarkUserStoryStatusTransition(b *testing.B) {
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

	// Get test data for user story creation
	var users []models.User
	require.NoError(b, server.DB.Limit(1).Find(&users).Error)
	require.NotEmpty(b, users)
	userID := users[0].ID

	var epics []models.Epic
	require.NoError(b, server.DB.Limit(1).Find(&epics).Error)
	require.NotEmpty(b, epics)
	epicID := epics[0].ID

	// Create user stories for status transition testing
	userStoryIDs := make([]uuid.UUID, b.N)
	for i := 0; i < b.N; i++ {
		createReq := service.CreateUserStoryRequest{
			EpicID:      epicID,
			CreatorID:   userID,
			Priority:    models.PriorityMedium,
			Title:       fmt.Sprintf("User Story for Status Change %d", i),
			Description: stringPtr("As a user, I want status changes, so that I can track progress"),
		}

		resp, err := client.POST("/api/v1/user-stories", createReq)
		require.NoError(b, err)
		require.Equal(b, http.StatusCreated, resp.StatusCode)

		var userStory models.UserStory
		require.NoError(b, helpers.ParseJSONResponse(resp, &userStory))
		userStoryIDs[i] = userStory.ID
	}

	b.ResetTimer()

	b.Run("ChangeUserStoryStatus", func(b *testing.B) {
		statuses := []models.UserStoryStatus{
			models.UserStoryStatusInProgress,
			models.UserStoryStatusDone,
			models.UserStoryStatusBacklog,
			models.UserStoryStatusDraft,
		}

		for i := 0; i < b.N; i++ {
			statusReq := map[string]interface{}{
				"status": statuses[i%len(statuses)],
			}

			resp, err := client.PATCH(fmt.Sprintf("/api/v1/user-stories/%s/status", userStoryIDs[i]), statusReq)
			require.NoError(b, err)
			require.Equal(b, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		}
	})
}

// BenchmarkUserStoryAssignment tests User Story assignment performance
func BenchmarkUserStoryAssignment(b *testing.B) {
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

	var epics []models.Epic
	require.NoError(b, server.DB.Limit(1).Find(&epics).Error)
	require.NotEmpty(b, epics)
	epicID := epics[0].ID

	// Create user stories for assignment testing
	userStoryIDs := make([]uuid.UUID, b.N)
	for i := 0; i < b.N; i++ {
		createReq := service.CreateUserStoryRequest{
			EpicID:      epicID,
			CreatorID:   users[0].ID,
			Priority:    models.PriorityMedium,
			Title:       fmt.Sprintf("User Story for Assignment %d", i),
			Description: stringPtr("As a user, I want to be assigned, so that ownership is clear"),
		}

		resp, err := client.POST("/api/v1/user-stories", createReq)
		require.NoError(b, err)
		require.Equal(b, http.StatusCreated, resp.StatusCode)

		var userStory models.UserStory
		require.NoError(b, helpers.ParseJSONResponse(resp, &userStory))
		userStoryIDs[i] = userStory.ID
	}

	b.ResetTimer()

	b.Run("AssignUserStory", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			assigneeID := users[i%len(users)].ID
			assignReq := map[string]interface{}{
				"assignee_id": assigneeID,
			}

			resp, err := client.PATCH(fmt.Sprintf("/api/v1/user-stories/%s/assign", userStoryIDs[i]), assignReq)
			require.NoError(b, err)
			require.Equal(b, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		}
	})
}

// BenchmarkUserStoryRelationshipManagement tests User Story relationship management via API endpoints
func BenchmarkUserStoryRelationshipManagement(b *testing.B) {
	// Setup benchmark server
	server := setup.NewBenchmarkServer(b)
	defer server.Cleanup()

	// Start the server
	require.NoError(b, server.Start())

	// Seed medium dataset to have user stories with relationships
	require.NoError(b, server.SeedMediumDataSet())

	// Create HTTP client
	client := helpers.NewBenchmarkClient(server.BaseURL)

	// Setup authentication
	authHelper := helpers.NewAuthHelper(server.Config.JWT.Secret)
	testUser := helpers.GetDefaultTestUser()
	require.NoError(b, authHelper.AuthenticateClient(client, testUser.ID, testUser.Username))

	// Get user stories that have acceptance criteria and requirements
	var userStories []models.UserStory
	require.NoError(b, server.DB.Preload("AcceptanceCriteria").Preload("Requirements").Find(&userStories).Error)
	require.NotEmpty(b, userStories)

	// Filter user stories that actually have relationships
	var userStoriesWithAC []models.UserStory
	var userStoriesWithReqs []models.UserStory
	for _, us := range userStories {
		if len(us.AcceptanceCriteria) > 0 {
			userStoriesWithAC = append(userStoriesWithAC, us)
		}
		if len(us.Requirements) > 0 {
			userStoriesWithReqs = append(userStoriesWithReqs, us)
		}
	}

	b.ResetTimer()

	if len(userStoriesWithAC) > 0 {
		b.Run("GetUserStoryWithAcceptanceCriteria", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				userStory := userStoriesWithAC[i%len(userStoriesWithAC)]
				resp, err := client.GET(fmt.Sprintf("/api/v1/user-stories/%s/acceptance-criteria", userStory.ID))
				require.NoError(b, err)
				require.Equal(b, http.StatusOK, resp.StatusCode)
				resp.Body.Close()
			}
		})
	}

	if len(userStoriesWithReqs) > 0 {
		b.Run("GetUserStoryWithRequirements", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				userStory := userStoriesWithReqs[i%len(userStoriesWithReqs)]
				resp, err := client.GET(fmt.Sprintf("/api/v1/user-stories/%s/requirements", userStory.ID))
				require.NoError(b, err)
				require.Equal(b, http.StatusOK, resp.StatusCode)
				resp.Body.Close()
			}
		})
	}

	b.Run("GetUserStoryByReferenceID", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			userStory := userStories[i%len(userStories)]
			resp, err := client.GET(fmt.Sprintf("/api/v1/user-stories/%s", userStory.ReferenceID))
			require.NoError(b, err)
			require.Equal(b, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		}
	})
}

// BenchmarkUserStoryConcurrentOperations tests concurrent User Story operations
func BenchmarkUserStoryConcurrentOperations(b *testing.B) {
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

	// Get test data
	var users []models.User
	require.NoError(b, server.DB.Limit(1).Find(&users).Error)
	require.NotEmpty(b, users)
	userID := users[0].ID

	var epics []models.Epic
	require.NoError(b, server.DB.Limit(1).Find(&epics).Error)
	require.NotEmpty(b, epics)
	epicID := epics[0].ID

	b.ResetTimer()

	b.Run("ConcurrentUserStoryCreation", func(b *testing.B) {
		// Create requests for parallel execution
		requests := make([]helpers.Request, b.N)
		for i := 0; i < b.N; i++ {
			createReq := service.CreateUserStoryRequest{
				EpicID:      epicID,
				CreatorID:   userID,
				Priority:    models.PriorityMedium,
				Title:       fmt.Sprintf("Concurrent User Story %d", i),
				Description: stringPtr(fmt.Sprintf("As a user, I want concurrent feature %d, so that I can test parallel performance", i)),
			}
			requests[i] = helpers.Request{
				Method: "POST",
				Path:   "/api/v1/user-stories",
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

	b.Run("ConcurrentUserStoryReads", func(b *testing.B) {
		// Get some existing user stories
		var userStories []models.UserStory
		require.NoError(b, server.DB.Limit(10).Find(&userStories).Error)
		require.NotEmpty(b, userStories)

		// Create read requests for parallel execution
		requests := make([]helpers.Request, b.N)
		for i := 0; i < b.N; i++ {
			userStory := userStories[i%len(userStories)]
			requests[i] = helpers.Request{
				Method: "GET",
				Path:   fmt.Sprintf("/api/v1/user-stories/%s", userStory.ID),
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


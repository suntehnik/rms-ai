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

// BenchmarkAcceptanceCriteriaCRUD tests Acceptance Criteria CRUD operations via HTTP endpoints
func BenchmarkAcceptanceCriteriaCRUD(b *testing.B) {
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

	// Get test data for acceptance criteria creation
	var users []models.User
	require.NoError(b, server.DB.Limit(1).Find(&users).Error)
	require.NotEmpty(b, users)
	authorID := users[0].ID

	var userStories []models.UserStory
	require.NoError(b, server.DB.Limit(1).Find(&userStories).Error)
	require.NotEmpty(b, userStories)
	userStoryID := userStories[0].ID

	b.ResetTimer()

	b.Run("CreateAcceptanceCriteria", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			createReq := service.CreateAcceptanceCriteriaRequest{
				AuthorID:    authorID,
				Description: fmt.Sprintf("WHEN user performs action %d THEN system SHALL respond with result %d", i, i),
			}

			resp, err := client.POST(fmt.Sprintf("/api/v1/user-stories/%s/acceptance-criteria", userStoryID), createReq)
			require.NoError(b, err)
			require.Equal(b, http.StatusCreated, resp.StatusCode)
			resp.Body.Close()
		}
	})

	// Create an acceptance criteria for read/update/delete operations
	createReq := service.CreateAcceptanceCriteriaRequest{
		AuthorID:    authorID,
		Description: "WHEN user submits form THEN system SHALL validate all required fields",
	}

	resp, err := client.POST(fmt.Sprintf("/api/v1/user-stories/%s/acceptance-criteria", userStoryID), createReq)
	require.NoError(b, err)
	require.Equal(b, http.StatusCreated, resp.StatusCode)

	var createdAcceptanceCriteria models.AcceptanceCriteria
	require.NoError(b, helpers.ParseJSONResponse(resp, &createdAcceptanceCriteria))

	b.Run("GetAcceptanceCriteria", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			resp, err := client.GET(fmt.Sprintf("/api/v1/acceptance-criteria/%s", createdAcceptanceCriteria.ID))
			require.NoError(b, err)
			require.Equal(b, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		}
	})

	b.Run("UpdateAcceptanceCriteria", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			updateReq := service.UpdateAcceptanceCriteriaRequest{
				Description: stringPtr(fmt.Sprintf("WHEN user updates data %d THEN system SHALL save changes %d", i, i)),
			}

			resp, err := client.PUT(fmt.Sprintf("/api/v1/acceptance-criteria/%s", createdAcceptanceCriteria.ID), updateReq)
			require.NoError(b, err)
			require.Equal(b, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		}
	})

	b.Run("DeleteAcceptanceCriteria", func(b *testing.B) {
		// Create acceptance criteria to delete
		acceptanceCriteriaIDs := make([]uuid.UUID, b.N)
		for i := 0; i < b.N; i++ {
			createReq := service.CreateAcceptanceCriteriaRequest{
				AuthorID:    authorID,
				Description: fmt.Sprintf("WHEN user deletes item %d THEN system SHALL remove it from database %d", i, i),
			}

			resp, err := client.POST(fmt.Sprintf("/api/v1/user-stories/%s/acceptance-criteria", userStoryID), createReq)
			require.NoError(b, err)
			require.Equal(b, http.StatusCreated, resp.StatusCode)

			var acceptanceCriteria models.AcceptanceCriteria
			require.NoError(b, helpers.ParseJSONResponse(resp, &acceptanceCriteria))
			acceptanceCriteriaIDs[i] = acceptanceCriteria.ID
		}

		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			resp, err := client.DELETE(fmt.Sprintf("/api/v1/acceptance-criteria/%s?force=true", acceptanceCriteriaIDs[i]))
			require.NoError(b, err)
			require.Equal(b, http.StatusNoContent, resp.StatusCode)
			resp.Body.Close()
		}
	})
}

// BenchmarkAcceptanceCriteriaListing tests Acceptance Criteria listing and filtering performance
func BenchmarkAcceptanceCriteriaListing(b *testing.B) {
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

	b.Run("ListAllAcceptanceCriteria", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			resp, err := client.GET("/api/v1/acceptance-criteria")
			require.NoError(b, err)
			require.Equal(b, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		}
	})

	b.Run("ListAcceptanceCriteriaWithLimit", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			resp, err := client.GET("/api/v1/acceptance-criteria?limit=10")
			require.NoError(b, err)
			require.Equal(b, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		}
	})

	b.Run("ListAcceptanceCriteriaWithPagination", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			offset := (i % 10) * 5 // Vary offset for different pages
			resp, err := client.GET(fmt.Sprintf("/api/v1/acceptance-criteria?limit=5&offset=%d", offset))
			require.NoError(b, err)
			require.Equal(b, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		}
	})

	b.Run("ListAcceptanceCriteriaWithOrdering", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			resp, err := client.GET("/api/v1/acceptance-criteria?order_by=created_at")
			require.NoError(b, err)
			require.Equal(b, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		}
	})

	b.Run("ListAcceptanceCriteriaByUserStory", func(b *testing.B) {
		// Get a user story ID for filtering
		var userStories []models.UserStory
		require.NoError(b, server.DB.Limit(1).Find(&userStories).Error)
		require.NotEmpty(b, userStories)
		userStoryID := userStories[0].ID

		for i := 0; i < b.N; i++ {
			resp, err := client.GET(fmt.Sprintf("/api/v1/acceptance-criteria?user_story_id=%s", userStoryID))
			require.NoError(b, err)
			require.Equal(b, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		}
	})

	b.Run("ListAcceptanceCriteriaByAuthor", func(b *testing.B) {
		// Get an author ID for filtering
		var users []models.User
		require.NoError(b, server.DB.Limit(1).Find(&users).Error)
		require.NotEmpty(b, users)
		authorID := users[0].ID

		for i := 0; i < b.N; i++ {
			resp, err := client.GET(fmt.Sprintf("/api/v1/acceptance-criteria?author_id=%s", authorID))
			require.NoError(b, err)
			require.Equal(b, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		}
	})
}

// BenchmarkAcceptanceCriteriaValidation tests Acceptance Criteria validation benchmarks
func BenchmarkAcceptanceCriteriaValidation(b *testing.B) {
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
	authorID := users[0].ID

	var userStories []models.UserStory
	require.NoError(b, server.DB.Limit(1).Find(&userStories).Error)
	require.NotEmpty(b, userStories)
	userStoryID := userStories[0].ID

	b.ResetTimer()

	b.Run("ValidateEARSFormat", func(b *testing.B) {
		// Test various EARS format acceptance criteria
		earsFormats := []string{
			"WHEN user clicks submit THEN system SHALL validate form",
			"IF user is authenticated THEN system SHALL allow access",
			"WHEN data is invalid THEN system SHALL display error message",
			"IF connection fails THEN system SHALL retry automatically",
			"WHEN user logs out THEN system SHALL clear session data",
		}

		for i := 0; i < b.N; i++ {
			description := earsFormats[safeIndex(i, len(earsFormats))]
			createReq := service.CreateAcceptanceCriteriaRequest{
				AuthorID:    authorID,
				Description: description,
			}

			resp, err := client.POST(fmt.Sprintf("/api/v1/user-stories/%s/acceptance-criteria", userStoryID), createReq)
			require.NoError(b, err)
			require.Equal(b, http.StatusCreated, resp.StatusCode)
			resp.Body.Close()
		}
	})

	b.Run("ValidateInvalidFormat", func(b *testing.B) {
		// Test invalid format handling (should still create but may not pass EARS validation)
		invalidFormats := []string{
			"User should be able to login",
			"The system needs to work fast",
			"Data must be secure",
			"Performance should be good",
			"Interface must be user-friendly",
		}

		for i := 0; i < b.N; i++ {
			description := invalidFormats[safeIndex(i, len(invalidFormats))]
			createReq := service.CreateAcceptanceCriteriaRequest{
				AuthorID:    authorID,
				Description: description,
			}

			resp, err := client.POST(fmt.Sprintf("/api/v1/user-stories/%s/acceptance-criteria", userStoryID), createReq)
			require.NoError(b, err)
			require.Equal(b, http.StatusCreated, resp.StatusCode)
			resp.Body.Close()
		}
	})

	b.Run("ValidateRequiredFields", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Test missing description (should fail validation)
			createReq := service.CreateAcceptanceCriteriaRequest{
				AuthorID: authorID,
				// Description intentionally omitted
			}

			resp, err := client.POST(fmt.Sprintf("/api/v1/user-stories/%s/acceptance-criteria", userStoryID), createReq)
			require.NoError(b, err)
			require.Equal(b, http.StatusBadRequest, resp.StatusCode)
			resp.Body.Close()
		}
	})

	b.Run("ValidateInvalidUserStory", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Test with non-existent user story ID
			invalidUserStoryID := uuid.New()
			createReq := service.CreateAcceptanceCriteriaRequest{
				AuthorID:    authorID,
				Description: "WHEN user performs action THEN system SHALL respond",
			}

			resp, err := client.POST(fmt.Sprintf("/api/v1/user-stories/%s/acceptance-criteria", invalidUserStoryID), createReq)
			require.NoError(b, err)
			require.Equal(b, http.StatusBadRequest, resp.StatusCode)
			resp.Body.Close()
		}
	})

	b.Run("ValidateInvalidAuthor", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Test with non-existent author ID
			invalidAuthorID := uuid.New()
			createReq := service.CreateAcceptanceCriteriaRequest{
				AuthorID:    invalidAuthorID,
				Description: "WHEN user performs action THEN system SHALL respond",
			}

			resp, err := client.POST(fmt.Sprintf("/api/v1/user-stories/%s/acceptance-criteria", userStoryID), createReq)
			require.NoError(b, err)
			require.Equal(b, http.StatusBadRequest, resp.StatusCode)
			resp.Body.Close()
		}
	})
}

// BenchmarkAcceptanceCriteriaRelationshipOperations tests Acceptance Criteria relationship operations via API endpoints
func BenchmarkAcceptanceCriteriaRelationshipOperations(b *testing.B) {
	// Setup benchmark server
	server := setup.NewBenchmarkServer(b)
	defer server.Cleanup()

	// Start the server
	require.NoError(b, server.Start())

	// Seed medium dataset to have acceptance criteria with relationships
	require.NoError(b, server.SeedMediumDataSet())

	// Create HTTP client
	client := helpers.NewBenchmarkClient(server.BaseURL)

	// Setup authentication
	authHelper := helpers.NewAuthHelper(server.Config.JWT.Secret)
	testUser := helpers.GetDefaultTestUser()
	require.NoError(b, authHelper.AuthenticateClient(client, testUser.ID, testUser.Username))

	// Get acceptance criteria that have requirements
	var acceptanceCriteria []models.AcceptanceCriteria
	require.NoError(b, server.DB.Preload("Requirements").Find(&acceptanceCriteria).Error)
	require.NotEmpty(b, acceptanceCriteria)

	// Get user stories with acceptance criteria
	var userStories []models.UserStory
	require.NoError(b, server.DB.Preload("AcceptanceCriteria").Find(&userStories).Error)
	require.NotEmpty(b, userStories)

	// Filter user stories that actually have acceptance criteria
	var userStoriesWithAC []models.UserStory
	for _, us := range userStories {
		if len(us.AcceptanceCriteria) > 0 {
			userStoriesWithAC = append(userStoriesWithAC, us)
		}
	}

	b.ResetTimer()

	b.Run("GetAcceptanceCriteriaByReferenceID", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			ac := acceptanceCriteria[i%len(acceptanceCriteria)]
			resp, err := client.GET(fmt.Sprintf("/api/v1/acceptance-criteria/%s", ac.ReferenceID))
			require.NoError(b, err)
			require.Equal(b, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		}
	})

	if len(userStoriesWithAC) > 0 {
		b.Run("GetAcceptanceCriteriaByUserStory", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				userStory := userStoriesWithAC[i%len(userStoriesWithAC)]
				resp, err := client.GET(fmt.Sprintf("/api/v1/user-stories/%s/acceptance-criteria", userStory.ID))
				require.NoError(b, err)
				require.Equal(b, http.StatusOK, resp.StatusCode)
				resp.Body.Close()
			}
		})
	}

	// Note: GetAcceptanceCriteriaByAuthor endpoint (/api/v1/users/:id/acceptance-criteria) 
	// is not implemented in routes, so we skip this benchmark

	// Test acceptance criteria with requirements relationships
	var acceptanceCriteriaWithReqs []models.AcceptanceCriteria
	for _, ac := range acceptanceCriteria {
		if len(ac.Requirements) > 0 {
			acceptanceCriteriaWithReqs = append(acceptanceCriteriaWithReqs, ac)
		}
	}

	if len(acceptanceCriteriaWithReqs) > 0 {
		b.Run("GetAcceptanceCriteriaWithRequirements", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				ac := acceptanceCriteriaWithReqs[i%len(acceptanceCriteriaWithReqs)]
				resp, err := client.GET(fmt.Sprintf("/api/v1/acceptance-criteria/%s", ac.ID))
				require.NoError(b, err)
				require.Equal(b, http.StatusOK, resp.StatusCode)
				resp.Body.Close()
			}
		})

		b.Run("DeleteAcceptanceCriteriaWithRequirements", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				ac := acceptanceCriteriaWithReqs[i%len(acceptanceCriteriaWithReqs)]
				// Test deletion without force (should fail due to requirements)
				resp, err := client.DELETE(fmt.Sprintf("/api/v1/acceptance-criteria/%s", ac.ID))
				require.NoError(b, err)
				require.Equal(b, http.StatusConflict, resp.StatusCode)
				resp.Body.Close()
			}
		})

		b.Run("ForceDeleteAcceptanceCriteriaWithRequirements", func(b *testing.B) {
			// Create acceptance criteria with requirements for force deletion testing
			var users []models.User
			require.NoError(b, server.DB.Limit(1).Find(&users).Error)
			require.NotEmpty(b, users)
			authorID := users[0].ID

			var userStories []models.UserStory
			require.NoError(b, server.DB.Limit(1).Find(&userStories).Error)
			require.NotEmpty(b, userStories)
			userStoryID := userStories[0].ID

			// Create acceptance criteria for force deletion
			acIDs := make([]uuid.UUID, b.N)
			for i := 0; i < b.N; i++ {
				createReq := service.CreateAcceptanceCriteriaRequest{
					AuthorID:    authorID,
					Description: fmt.Sprintf("WHEN user performs force delete test %d THEN system SHALL handle it %d", i, i),
				}

				resp, err := client.POST(fmt.Sprintf("/api/v1/user-stories/%s/acceptance-criteria", userStoryID), createReq)
				require.NoError(b, err)
				require.Equal(b, http.StatusCreated, resp.StatusCode)

				var ac models.AcceptanceCriteria
				require.NoError(b, helpers.ParseJSONResponse(resp, &ac))
				acIDs[i] = ac.ID
			}

			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				// Test force deletion (should succeed)
				resp, err := client.DELETE(fmt.Sprintf("/api/v1/acceptance-criteria/%s?force=true", acIDs[i]))
				require.NoError(b, err)
				require.Equal(b, http.StatusNoContent, resp.StatusCode)
				resp.Body.Close()
			}
		})
	}
}

// BenchmarkAcceptanceCriteriaConcurrentOperations tests concurrent Acceptance Criteria operations
func BenchmarkAcceptanceCriteriaConcurrentOperations(b *testing.B) {
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
	authorID := users[0].ID

	var userStories []models.UserStory
	require.NoError(b, server.DB.Limit(1).Find(&userStories).Error)
	require.NotEmpty(b, userStories)
	userStoryID := userStories[0].ID

	b.ResetTimer()

	b.Run("ConcurrentAcceptanceCriteriaCreation", func(b *testing.B) {
		// Create requests for parallel execution
		requests := make([]helpers.Request, b.N)
		for i := 0; i < b.N; i++ {
			createReq := service.CreateAcceptanceCriteriaRequest{
				AuthorID:    authorID,
				Description: fmt.Sprintf("WHEN user performs concurrent action %d THEN system SHALL handle it %d", i, i),
			}
			requests[i] = helpers.Request{
				Method: "POST",
				Path:   fmt.Sprintf("/api/v1/user-stories/%s/acceptance-criteria", userStoryID),
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

	b.Run("ConcurrentAcceptanceCriteriaReads", func(b *testing.B) {
		// Get some existing acceptance criteria
		var acceptanceCriteria []models.AcceptanceCriteria
		require.NoError(b, server.DB.Limit(10).Find(&acceptanceCriteria).Error)
		require.NotEmpty(b, acceptanceCriteria)

		// Create read requests for parallel execution
		requests := make([]helpers.Request, b.N)
		for i := 0; i < b.N; i++ {
			ac := acceptanceCriteria[i%len(acceptanceCriteria)]
			requests[i] = helpers.Request{
				Method: "GET",
				Path:   fmt.Sprintf("/api/v1/acceptance-criteria/%s", ac.ID),
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

	b.Run("ConcurrentAcceptanceCriteriaUpdates", func(b *testing.B) {
		// Create acceptance criteria for concurrent updates
		var acceptanceCriteria []models.AcceptanceCriteria
		numAC := 10 // Create 10 acceptance criteria for concurrent updates
		for i := 0; i < numAC; i++ {
			createReq := service.CreateAcceptanceCriteriaRequest{
				AuthorID:    authorID,
				Description: fmt.Sprintf("WHEN user performs update test %d THEN system SHALL handle it %d", i, i),
			}

			resp, err := client.POST(fmt.Sprintf("/api/v1/user-stories/%s/acceptance-criteria", userStoryID), createReq)
			require.NoError(b, err)
			require.Equal(b, http.StatusCreated, resp.StatusCode)

			var ac models.AcceptanceCriteria
			require.NoError(b, helpers.ParseJSONResponse(resp, &ac))
			acceptanceCriteria = append(acceptanceCriteria, ac)
		}

		// Create update requests for parallel execution
		requests := make([]helpers.Request, b.N)
		for i := 0; i < b.N; i++ {
			ac := acceptanceCriteria[i%len(acceptanceCriteria)]
			updateReq := service.UpdateAcceptanceCriteriaRequest{
				Description: stringPtr(fmt.Sprintf("WHEN user performs concurrent update %d THEN system SHALL save changes %d", i, i)),
			}
			requests[i] = helpers.Request{
				Method: "PUT",
				Path:   fmt.Sprintf("/api/v1/acceptance-criteria/%s", ac.ID),
				Body:   updateReq,
			}
		}

		// Execute requests with limited concurrency
		concurrency := 5
		responses, err := client.RunParallelRequests(requests, concurrency)
		require.NoError(b, err)

		// Verify all requests succeeded
		for i, resp := range responses {
			require.NoError(b, resp.Error, "Request %d failed", i)
			require.Equal(b, http.StatusOK, resp.StatusCode, "Request %d returned wrong status", i)
		}
	})
}
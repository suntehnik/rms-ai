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

// BenchmarkCommentCRUD tests Comment CRUD operations via HTTP endpoints
func BenchmarkCommentCRUD(b *testing.B) {
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

	// Get test data for comment creation
	var users []models.User
	require.NoError(b, server.DB.Limit(1).Find(&users).Error)
	require.NotEmpty(b, users)
	userID := users[0].ID

	var epics []models.Epic
	require.NoError(b, server.DB.Limit(1).Find(&epics).Error)
	require.NotEmpty(b, epics)
	epicID := epics[0].ID

	b.ResetTimer()

	b.Run("CreateComment", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			createReq := service.CreateCommentRequest{
				EntityType: models.EntityTypeEpic,
				EntityID:   epicID,
				AuthorID:   userID,
				Content:    fmt.Sprintf("Benchmark comment %d", i),
			}

			resp, err := client.POST(fmt.Sprintf("/api/v1/epics/%s/comments", epicID), createReq)
			require.NoError(b, err)
			require.Equal(b, http.StatusCreated, resp.StatusCode)
			resp.Body.Close()
		}
	})

	// Create a comment for read/update/delete operations
	createReq := service.CreateCommentRequest{
		EntityType: models.EntityTypeEpic,
		EntityID:   epicID,
		AuthorID:   userID,
		Content:    "Test comment for CRUD operations",
	}

	resp, err := client.POST(fmt.Sprintf("/api/v1/epics/%s/comments", epicID), createReq)
	require.NoError(b, err)
	require.Equal(b, http.StatusCreated, resp.StatusCode)

	var createdComment service.CommentResponse
	require.NoError(b, helpers.ParseJSONResponse(resp, &createdComment))

	b.Run("GetComment", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			resp, err := client.GET(fmt.Sprintf("/api/v1/comments/%s", createdComment.ID))
			require.NoError(b, err)
			require.Equal(b, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		}
	})

	b.Run("UpdateComment", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			updateReq := service.UpdateCommentRequest{
				Content: fmt.Sprintf("Updated comment content %d", i),
			}

			resp, err := client.PUT(fmt.Sprintf("/api/v1/comments/%s", createdComment.ID), updateReq)
			require.NoError(b, err)
			require.Equal(b, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		}
	})

	b.Run("DeleteComment", func(b *testing.B) {
		// Create comments to delete
		commentIDs := make([]uuid.UUID, b.N)
		for i := 0; i < b.N; i++ {
			createReq := service.CreateCommentRequest{
				EntityType: models.EntityTypeEpic,
				EntityID:   epicID,
				AuthorID:   userID,
				Content:    fmt.Sprintf("Comment to delete %d", i),
			}

			resp, err := client.POST(fmt.Sprintf("/api/v1/epics/%s/comments", epicID), createReq)
			require.NoError(b, err)
			require.Equal(b, http.StatusCreated, resp.StatusCode)

			var comment service.CommentResponse
			require.NoError(b, helpers.ParseJSONResponse(resp, &comment))
			commentIDs[i] = comment.ID
		}

		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			resp, err := client.DELETE(fmt.Sprintf("/api/v1/comments/%s", commentIDs[i]))
			require.NoError(b, err)
			require.Equal(b, http.StatusNoContent, resp.StatusCode)
			resp.Body.Close()
		}
	})
}

// BenchmarkCommentThreading tests comment threading operations performance
func BenchmarkCommentThreading(b *testing.B) {
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

	// Create parent comments for threading tests
	parentCommentIDs := make([]uuid.UUID, 10)
	for i := 0; i < 10; i++ {
		createReq := service.CreateCommentRequest{
			EntityType: models.EntityTypeEpic,
			EntityID:   epicID,
			AuthorID:   userID,
			Content:    fmt.Sprintf("Parent comment %d", i),
		}

		resp, err := client.POST(fmt.Sprintf("/api/v1/epics/%s/comments", epicID), createReq)
		require.NoError(b, err)
		require.Equal(b, http.StatusCreated, resp.StatusCode)

		var comment service.CommentResponse
		require.NoError(b, helpers.ParseJSONResponse(resp, &comment))
		parentCommentIDs[i] = comment.ID
	}

	b.ResetTimer()

	b.Run("CreateCommentReply", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			parentID := parentCommentIDs[i%len(parentCommentIDs)]
			createReq := service.CreateCommentRequest{
				AuthorID: userID,
				Content:  fmt.Sprintf("Reply to comment %d", i),
			}

			resp, err := client.POST(fmt.Sprintf("/api/v1/comments/%s/replies", parentID), createReq)
			require.NoError(b, err)
			require.Equal(b, http.StatusCreated, resp.StatusCode)
			resp.Body.Close()
		}
	})

	b.Run("GetThreadedComments", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			resp, err := client.GET(fmt.Sprintf("/api/v1/epics/%s/comments?threaded=true", epicID))
			require.NoError(b, err)
			require.Equal(b, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		}
	})

	b.Run("GetCommentReplies", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			parentID := parentCommentIDs[i%len(parentCommentIDs)]
			resp, err := client.GET(fmt.Sprintf("/api/v1/comments/%s/replies", parentID))
			require.NoError(b, err)
			require.Equal(b, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		}
	})
}

// BenchmarkCommentResolution tests comment resolution and status operations
func BenchmarkCommentResolution(b *testing.B) {
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

	// Create comments for resolution testing
	commentIDs := make([]uuid.UUID, 20)
	for i := 0; i < 20; i++ {
		createReq := service.CreateCommentRequest{
			EntityType: models.EntityTypeEpic,
			EntityID:   epicID,
			AuthorID:   userID,
			Content:    fmt.Sprintf("Comment for resolution %d", i),
		}

		resp, err := client.POST(fmt.Sprintf("/api/v1/epics/%s/comments", epicID), createReq)
		require.NoError(b, err)
		require.Equal(b, http.StatusCreated, resp.StatusCode)

		var comment service.CommentResponse
		require.NoError(b, helpers.ParseJSONResponse(resp, &comment))
		commentIDs[i] = comment.ID
	}

	b.ResetTimer()

	b.Run("ResolveComment", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			commentID := commentIDs[i%len(commentIDs)]
			resp, err := client.POST(fmt.Sprintf("/api/v1/comments/%s/resolve", commentID), nil)
			require.NoError(b, err)
			require.Equal(b, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		}
	})

	b.Run("UnresolveComment", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			commentID := commentIDs[i%len(commentIDs)]
			resp, err := client.POST(fmt.Sprintf("/api/v1/comments/%s/unresolve", commentID), nil)
			require.NoError(b, err)
			require.Equal(b, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		}
	})

	b.Run("GetCommentsByStatus", func(b *testing.B) {
		statuses := []string{"resolved", "unresolved"}
		for i := 0; i < b.N; i++ {
			status := statuses[i%len(statuses)]
			resp, err := client.GET(fmt.Sprintf("/api/v1/comments/status/%s", status))
			require.NoError(b, err)
			require.Equal(b, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		}
	})

	b.Run("GetCommentsWithStatusFilter", func(b *testing.B) {
		statuses := []string{"resolved", "unresolved"}
		for i := 0; i < b.N; i++ {
			status := statuses[i%len(statuses)]
			resp, err := client.GET(fmt.Sprintf("/api/v1/epics/%s/comments?status=%s", epicID, status))
			require.NoError(b, err)
			require.Equal(b, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		}
	})
}

// BenchmarkInlineComments tests inline comment performance via API endpoints
func BenchmarkInlineComments(b *testing.B) {
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

	b.Run("CreateInlineComment", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			createReq := service.CreateCommentRequest{
				AuthorID:          userID,
				Content:           fmt.Sprintf("Inline comment %d", i),
				LinkedText:        stringPtr("sample text"),
				TextPositionStart: intPtr(0),
				TextPositionEnd:   intPtr(11),
			}

			resp, err := client.POST(fmt.Sprintf("/api/v1/epics/%s/comments/inline", epicID), createReq)
			require.NoError(b, err)
			require.Equal(b, http.StatusCreated, resp.StatusCode)
			resp.Body.Close()
		}
	})

	b.Run("GetInlineComments", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			resp, err := client.GET(fmt.Sprintf("/api/v1/epics/%s/comments?inline=true", epicID))
			require.NoError(b, err)
			require.Equal(b, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		}
	})

	b.Run("GetVisibleInlineComments", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			resp, err := client.GET(fmt.Sprintf("/api/v1/epics/%s/comments/inline/visible", epicID))
			require.NoError(b, err)
			require.Equal(b, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		}
	})

	b.Run("ValidateInlineComments", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			validateReq := map[string]interface{}{
				"new_description": fmt.Sprintf("Updated description %d with new content", i),
			}

			resp, err := client.POST(fmt.Sprintf("/api/v1/epics/%s/comments/inline/validate", epicID), validateReq)
			require.NoError(b, err)
			require.Equal(b, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		}
	})
}

// BenchmarkCommentEntityOperations tests comment operations across different entity types
func BenchmarkCommentEntityOperations(b *testing.B) {
	// Setup benchmark server
	server := setup.NewBenchmarkServer(b)
	defer server.Cleanup()

	// Start the server
	require.NoError(b, server.Start())

	// Seed medium dataset to have various entity types
	require.NoError(b, server.SeedMediumDataSet())

	// Create HTTP client
	client := helpers.NewBenchmarkClient(server.BaseURL)

	// Setup authentication
	authHelper := helpers.NewAuthHelper(server.Config.JWT.Secret)
	testUser := helpers.GetDefaultTestUser()
	require.NoError(b, authHelper.AuthenticateClient(client, testUser.ID, testUser.Username))

	// Get test data for different entity types
	var users []models.User
	require.NoError(b, server.DB.Limit(1).Find(&users).Error)
	require.NotEmpty(b, users)
	userID := users[0].ID

	var epics []models.Epic
	require.NoError(b, server.DB.Limit(1).Find(&epics).Error)
	require.NotEmpty(b, epics)

	var userStories []models.UserStory
	require.NoError(b, server.DB.Limit(1).Find(&userStories).Error)
	require.NotEmpty(b, userStories)

	var requirements []models.Requirement
	require.NoError(b, server.DB.Limit(1).Find(&requirements).Error)
	require.NotEmpty(b, requirements)

	// Define entity test cases
	entityTests := []struct {
		name       string
		entityType models.EntityType
		entityID   uuid.UUID
		endpoint   string
	}{
		{"Epic", models.EntityTypeEpic, epics[0].ID, "epics"},
		{"UserStory", models.EntityTypeUserStory, userStories[0].ID, "user-stories"},
		{"Requirement", models.EntityTypeRequirement, requirements[0].ID, "requirements"},
	}

	b.ResetTimer()

	for _, entityTest := range entityTests {
		b.Run(fmt.Sprintf("CreateComment_%s", entityTest.name), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				createReq := service.CreateCommentRequest{
					AuthorID: userID,
					Content:  fmt.Sprintf("Comment on %s %d", entityTest.name, i),
				}

				resp, err := client.POST(fmt.Sprintf("/api/v1/%s/%s/comments", entityTest.endpoint, entityTest.entityID), createReq)
				require.NoError(b, err)
				require.Equal(b, http.StatusCreated, resp.StatusCode)
				resp.Body.Close()
			}
		})

		b.Run(fmt.Sprintf("GetComments_%s", entityTest.name), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				resp, err := client.GET(fmt.Sprintf("/api/v1/%s/%s/comments", entityTest.endpoint, entityTest.entityID))
				require.NoError(b, err)
				require.Equal(b, http.StatusOK, resp.StatusCode)
				resp.Body.Close()
			}
		})
	}
}

// BenchmarkCommentConcurrentOperations tests concurrent comment operations
func BenchmarkCommentConcurrentOperations(b *testing.B) {
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

	b.Run("ConcurrentCommentCreation", func(b *testing.B) {
		// Create requests for parallel execution
		requests := make([]helpers.Request, b.N)
		for i := 0; i < b.N; i++ {
			createReq := service.CreateCommentRequest{
				AuthorID: userID,
				Content:  fmt.Sprintf("Concurrent comment %d", i),
			}
			requests[i] = helpers.Request{
				Method: "POST",
				Path:   fmt.Sprintf("/api/v1/epics/%s/comments", epicID),
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

	b.Run("ConcurrentCommentReads", func(b *testing.B) {
		// Create read requests for parallel execution
		requests := make([]helpers.Request, b.N)
		for i := 0; i < b.N; i++ {
			requests[i] = helpers.Request{
				Method: "GET",
				Path:   fmt.Sprintf("/api/v1/epics/%s/comments", epicID),
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

// intPtr creates a pointer to an int value
func intPtr(i int) *int {
	return &i
}

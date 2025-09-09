package api

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"product-requirements-management/internal/benchmarks/helpers"
	"product-requirements-management/internal/benchmarks/setup"
	"product-requirements-management/internal/models"
	"product-requirements-management/internal/service"
)

// BenchmarkRequirementCRUD tests Requirement CRUD operations via HTTP endpoints
func BenchmarkRequirementCRUD(b *testing.B) {
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
	typeID := requirementTypes[0].ID

	b.ResetTimer()

	b.Run("CreateRequirement", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			createReq := service.CreateRequirementRequest{
				UserStoryID: userStoryID,
				CreatorID:   userID,
				Priority:    models.PriorityMedium,
				TypeID:      typeID,
				Title:       fmt.Sprintf("Benchmark Requirement %d", i),
				Description: stringPtr(fmt.Sprintf("Description for benchmark requirement %d", i)),
			}

			resp, err := client.POST("/api/v1/requirements", createReq)
			require.NoError(b, err)
			require.Equal(b, http.StatusCreated, resp.StatusCode)
			resp.Body.Close()
		}
	})

	// Create a requirement for read/update/delete operations
	createReq := service.CreateRequirementRequest{
		UserStoryID: userStoryID,
		CreatorID:   userID,
		Priority:    models.PriorityHigh,
		TypeID:      typeID,
		Title:       "Test Requirement for CRUD",
		Description: stringPtr("Test requirement for read/update/delete operations"),
	}

	resp, err := client.POST("/api/v1/requirements", createReq)
	require.NoError(b, err)
	require.Equal(b, http.StatusCreated, resp.StatusCode)

	var createdRequirement models.Requirement
	require.NoError(b, helpers.ParseJSONResponse(resp, &createdRequirement))

	b.Run("GetRequirement", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			resp, err := client.GET(fmt.Sprintf("/api/v1/requirements/%s", createdRequirement.ID))
			require.NoError(b, err)
			require.Equal(b, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		}
	})

	b.Run("UpdateRequirement", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			updateReq := service.UpdateRequirementRequest{
				Title: stringPtr(fmt.Sprintf("Updated Requirement %d", i)),
			}

			resp, err := client.PUT(fmt.Sprintf("/api/v1/requirements/%s", createdRequirement.ID), updateReq)
			require.NoError(b, err)
			require.Equal(b, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		}
	})

	b.Run("DeleteRequirement", func(b *testing.B) {
		// Create requirements to delete
		requirementIDs := make([]uuid.UUID, b.N)
		for i := 0; i < b.N; i++ {
			createReq := service.CreateRequirementRequest{
				UserStoryID: userStoryID,
				CreatorID:   userID,
				Priority:    models.PriorityLow,
				TypeID:      typeID,
				Title:       fmt.Sprintf("Requirement to Delete %d", i),
				Description: stringPtr("Requirement created for deletion benchmark"),
			}

			resp, err := client.POST("/api/v1/requirements", createReq)
			require.NoError(b, err)
			require.Equal(b, http.StatusCreated, resp.StatusCode)

			var requirement models.Requirement
			require.NoError(b, helpers.ParseJSONResponse(resp, &requirement))
			requirementIDs[i] = requirement.ID
		}

		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			resp, err := client.DELETE(fmt.Sprintf("/api/v1/requirements/%s", requirementIDs[i]))
			require.NoError(b, err)
			require.Equal(b, http.StatusNoContent, resp.StatusCode)
			resp.Body.Close()
		}
	})
}

// BenchmarkRequirementListing tests Requirement listing and filtering performance
func BenchmarkRequirementListing(b *testing.B) {
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

	b.Run("ListAllRequirements", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			resp, err := client.GET("/api/v1/requirements")
			require.NoError(b, err)
			require.Equal(b, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		}
	})

	b.Run("ListRequirementsWithLimit", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			resp, err := client.GET("/api/v1/requirements?limit=10")
			require.NoError(b, err)
			require.Equal(b, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		}
	})

	b.Run("ListRequirementsByStatus", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			resp, err := client.GET("/api/v1/requirements?status=Draft")
			require.NoError(b, err)
			require.Equal(b, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		}
	})

	b.Run("ListRequirementsByPriority", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			resp, err := client.GET("/api/v1/requirements?priority=2")
			require.NoError(b, err)
			require.Equal(b, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		}
	})

	b.Run("ListRequirementsWithPagination", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			offset := (i % 10) * 5 // Vary offset for different pages
			resp, err := client.GET(fmt.Sprintf("/api/v1/requirements?limit=5&offset=%d", offset))
			require.NoError(b, err)
			require.Equal(b, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		}
	})

	b.Run("ListRequirementsWithOrdering", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			resp, err := client.GET("/api/v1/requirements?order_by=created_at")
			require.NoError(b, err)
			require.Equal(b, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		}
	})

	b.Run("ListRequirementsByUserStory", func(b *testing.B) {
		// Get a user story ID for filtering
		var userStories []models.UserStory
		require.NoError(b, server.DB.Limit(1).Find(&userStories).Error)
		require.NotEmpty(b, userStories)
		userStoryID := userStories[0].ID

		for i := 0; i < b.N; i++ {
			resp, err := client.GET(fmt.Sprintf("/api/v1/requirements?user_story_id=%s", userStoryID))
			require.NoError(b, err)
			require.Equal(b, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		}
	})

	b.Run("ListRequirementsByType", func(b *testing.B) {
		// Get a requirement type ID for filtering
		var requirementTypes []models.RequirementType
		require.NoError(b, server.DB.Limit(1).Find(&requirementTypes).Error)
		require.NotEmpty(b, requirementTypes)
		typeID := requirementTypes[0].ID

		for i := 0; i < b.N; i++ {
			resp, err := client.GET(fmt.Sprintf("/api/v1/requirements?type_id=%s", typeID))
			require.NoError(b, err)
			require.Equal(b, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		}
	})
}

// BenchmarkRequirementRelationshipManagement tests Requirement relationship management benchmarks
func BenchmarkRequirementRelationshipManagement(b *testing.B) {
	// Setup benchmark server
	server := setup.NewBenchmarkServer(b)
	defer server.Cleanup()

	// Start the server
	require.NoError(b, server.Start())

	// Seed medium dataset to have requirements with relationships
	require.NoError(b, server.SeedMediumDataSet())

	// Create HTTP client
	client := helpers.NewBenchmarkClient(server.BaseURL)

	// Setup authentication
	authHelper := helpers.NewAuthHelper(server.Config.JWT.Secret)
	testUser := helpers.GetDefaultTestUser()
	require.NoError(b, authHelper.AuthenticateClient(client, testUser.ID, testUser.Username))

	// Get requirements that have relationships
	var requirements []models.Requirement
	require.NoError(b, server.DB.Preload("SourceRelationships").Preload("TargetRelationships").Find(&requirements).Error)
	require.NotEmpty(b, requirements)

	// Filter requirements that actually have relationships
	var requirementsWithRelationships []models.Requirement
	for _, req := range requirements {
		if len(req.SourceRelationships) > 0 || len(req.TargetRelationships) > 0 {
			requirementsWithRelationships = append(requirementsWithRelationships, req)
		}
	}

	b.ResetTimer()

	b.Run("GetRequirementWithRelationships", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			requirement := requirements[i%len(requirements)]
			resp, err := client.GET(fmt.Sprintf("/api/v1/requirements/%s/relationships", requirement.ID))
			require.NoError(b, err)
			require.Equal(b, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		}
	})

	if len(requirementsWithRelationships) > 0 {
		b.Run("GetRequirementRelationships", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				requirement := requirementsWithRelationships[i%len(requirementsWithRelationships)]
				resp, err := client.GET(fmt.Sprintf("/api/v1/requirements/%s/relationships", requirement.ID))
				require.NoError(b, err)
				require.Equal(b, http.StatusOK, resp.StatusCode)
				resp.Body.Close()
			}
		})
	}

	b.Run("GetRequirementByReferenceID", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			requirement := requirements[i%len(requirements)]
			resp, err := client.GET(fmt.Sprintf("/api/v1/requirements/%s", requirement.ReferenceID))
			require.NoError(b, err)
			require.Equal(b, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		}
	})

	// Test creating relationships
	if len(requirements) >= 2 {
		b.Run("CreateRequirementRelationship", func(b *testing.B) {
			// Get relationship types
			var relationshipTypes []models.RelationshipType
			require.NoError(b, server.DB.Limit(1).Find(&relationshipTypes).Error)
			require.NotEmpty(b, relationshipTypes)
			relationshipTypeID := relationshipTypes[0].ID

			// Get user for creating relationships
			var users []models.User
			require.NoError(b, server.DB.Limit(1).Find(&users).Error)
			require.NotEmpty(b, users)
			userID := users[0].ID

			for i := 0; i < b.N; i++ {
				sourceReq := requirements[i%len(requirements)]
				targetReq := requirements[(i+1)%len(requirements)]

				// Skip if same requirement
				if sourceReq.ID == targetReq.ID {
					continue
				}

				createRelReq := service.CreateRelationshipRequest{
					SourceRequirementID: sourceReq.ID,
					TargetRequirementID: targetReq.ID,
					RelationshipTypeID:  relationshipTypeID,
					CreatedBy:           userID,
				}

				resp, err := client.POST("/api/v1/requirements/relationships", createRelReq)
				require.NoError(b, err)
				// Accept both 201 (created) and 409 (conflict for duplicate)
				require.True(b, resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusConflict)
				resp.Body.Close()
			}
		})
	}
}

// BenchmarkRequirementTypeAndStatusOperations tests Requirement type and status operations via API endpoints
func BenchmarkRequirementTypeAndStatusOperations(b *testing.B) {
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
	require.NoError(b, server.DB.Find(&requirementTypes).Error)
	require.NotEmpty(b, requirementTypes)

	// Create requirements for status change testing
	requirementIDs := make([]uuid.UUID, b.N)
	for i := 0; i < b.N; i++ {
		createReq := service.CreateRequirementRequest{
			UserStoryID: userStoryID,
			CreatorID:   userID,
			Priority:    models.PriorityMedium,
			TypeID:      requirementTypes[0].ID,
			Title:       fmt.Sprintf("Requirement for Status Change %d", i),
			Description: stringPtr("Requirement created for status change benchmark"),
		}

		resp, err := client.POST("/api/v1/requirements", createReq)
		require.NoError(b, err)
		require.Equal(b, http.StatusCreated, resp.StatusCode)

		var requirement models.Requirement
		require.NoError(b, helpers.ParseJSONResponse(resp, &requirement))
		requirementIDs[i] = requirement.ID
	}

	b.ResetTimer()

	b.Run("ChangeRequirementStatus", func(b *testing.B) {
		statuses := []models.RequirementStatus{
			models.RequirementStatusActive,
			models.RequirementStatusObsolete,
			models.RequirementStatusDraft,
		}

		for i := 0; i < b.N; i++ {
			statusReq := map[string]interface{}{
				"status": statuses[i%len(statuses)],
			}

			requirementID := requirementIDs[i%len(requirementIDs)]
			resp, err := client.PATCH(fmt.Sprintf("/api/v1/requirements/%s/status", requirementID), statusReq)
			require.NoError(b, err)
			require.Equal(b, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		}
	})

	b.Run("AssignRequirement", func(b *testing.B) {
		// Get multiple users for assignment
		var assignUsers []models.User
		require.NoError(b, server.DB.Limit(5).Find(&assignUsers).Error)
		require.NotEmpty(b, assignUsers)

		for i := 0; i < b.N; i++ {
			assigneeID := assignUsers[i%len(assignUsers)].ID
			assignReq := map[string]interface{}{
				"assignee_id": assigneeID,
			}

			requirementID := requirementIDs[i%len(requirementIDs)]
			resp, err := client.PATCH(fmt.Sprintf("/api/v1/requirements/%s/assign", requirementID), assignReq)
			require.NoError(b, err)
			require.Equal(b, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		}
	})

	b.Run("SearchRequirements", func(b *testing.B) {
		searchQueries := []string{
			"benchmark",
			"requirement",
			"test",
			"status",
			"change",
		}

		for i := 0; i < b.N; i++ {
			query := searchQueries[i%len(searchQueries)]
			resp, err := client.GET(fmt.Sprintf("/api/v1/requirements/search?q=%s", query))
			require.NoError(b, err)
			require.Equal(b, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		}
	})

	// Test filtering by different requirement types
	if len(requirementTypes) > 1 {
		b.Run("FilterRequirementsByType", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				reqType := requirementTypes[i%len(requirementTypes)]
				resp, err := client.GET(fmt.Sprintf("/api/v1/requirements?type_id=%s", reqType.ID))
				require.NoError(b, err)
				require.Equal(b, http.StatusOK, resp.StatusCode)
				resp.Body.Close()
			}
		})
	}
}

// BenchmarkRequirementConcurrentOperations tests concurrent Requirement operations
func BenchmarkRequirementConcurrentOperations(b *testing.B) {
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

	var userStories []models.UserStory
	require.NoError(b, server.DB.Limit(1).Find(&userStories).Error)
	require.NotEmpty(b, userStories)
	userStoryID := userStories[0].ID

	var requirementTypes []models.RequirementType
	require.NoError(b, server.DB.Limit(1).Find(&requirementTypes).Error)
	require.NotEmpty(b, requirementTypes)
	typeID := requirementTypes[0].ID

	b.ResetTimer()

	b.Run("ConcurrentRequirementCreation", func(b *testing.B) {
		// Create requests for parallel execution
		requests := make([]helpers.Request, b.N)
		for i := 0; i < b.N; i++ {
			createReq := service.CreateRequirementRequest{
				UserStoryID: userStoryID,
				CreatorID:   userID,
				Priority:    models.PriorityMedium,
				TypeID:      typeID,
				Title:       fmt.Sprintf("Concurrent Requirement %d-%d", time.Now().UnixNano(), i),
				Description: stringPtr(fmt.Sprintf("Requirement created concurrently %d at %d", i, time.Now().UnixNano())),
			}
			requests[i] = helpers.Request{
				Method: "POST",
				Path:   "/api/v1/requirements",
				Body:   createReq,
			}
		}

		// Execute requests with limited concurrency
		concurrency := 5 // Reduce concurrency to avoid database conflicts
		responses, err := client.RunParallelRequests(requests, concurrency)
		require.NoError(b, err)

		// Verify all requests succeeded
		for i, resp := range responses {
			require.NoError(b, resp.Error, "Request %d failed", i)
			if resp.StatusCode != http.StatusCreated {
				b.Logf("Request %d failed with status %d, body: %s", i, resp.StatusCode, string(resp.Body))
			}
			require.Equal(b, http.StatusCreated, resp.StatusCode, "Request %d returned wrong status", i)
		}
	})

	b.Run("ConcurrentRequirementReads", func(b *testing.B) {
		// Get some existing requirements
		var requirements []models.Requirement
		require.NoError(b, server.DB.Limit(10).Find(&requirements).Error)
		require.NotEmpty(b, requirements)

		// Create read requests for parallel execution
		requests := make([]helpers.Request, b.N)
		for i := 0; i < b.N; i++ {
			requirement := requirements[i%len(requirements)]
			requests[i] = helpers.Request{
				Method: "GET",
				Path:   fmt.Sprintf("/api/v1/requirements/%s", requirement.ID),
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


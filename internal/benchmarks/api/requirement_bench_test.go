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

// BenchmarkRequirementStatusChange tests Requirement status change performance
func BenchmarkRequirementStatusChange(b *testing.B) {
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
	require.NoError(b, server.DB.Limit(1).Find(&requirementTypes).Error)
	require.NotEmpty(b, requirementTypes)
	typeID := requirementTypes[0].ID

	// Create a reasonable number of requirements for status change testing
	numRequirements := b.N
	if numRequirements > 100 {
		numRequirements = 100 // Cap at 100 to avoid excessive setup time
	}
	if numRequirements < 10 {
		numRequirements = 10 // Minimum of 10 for reasonable testing
	}

	requirementIDs := make([]uuid.UUID, numRequirements)
	for i := 0; i < numRequirements; i++ {
		createReq := service.CreateRequirementRequest{
			UserStoryID: userStoryID,
			CreatorID:   userID,
			Priority:    models.PriorityMedium,
			TypeID:      typeID,
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

	// Validate test data
	validator := NewBenchmarkDataValidator(b)
	validator.ValidateUUIDs(requirementIDs, "requirement")

	b.ResetTimer()

	b.Run("ChangeRequirementStatus", func(b *testing.B) {
		statuses := []models.RequirementStatus{
			models.RequirementStatusActive,
			models.RequirementStatusObsolete,
			models.RequirementStatusDraft,
		}

		for i := 0; i < b.N; i++ {
			// Use safe indexing to cycle through available requirements
			requirementIndex := safeIndex(i, len(requirementIDs))
			
			statusReq := map[string]interface{}{
				"status": statuses[safeIndex(i, len(statuses))],
			}

			requirementID := requirementIDs[requirementIndex]
			resp, err := client.PATCH(fmt.Sprintf("/api/v1/requirements/%s/status", requirementID), statusReq)
			require.NoError(b, err)
			require.Equal(b, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		}
	})
}

// BenchmarkRequirementAssignment tests Requirement assignment performance
func BenchmarkRequirementAssignment(b *testing.B) {
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

	var userStories []models.UserStory
	require.NoError(b, server.DB.Limit(1).Find(&userStories).Error)
	require.NotEmpty(b, userStories)
	userStoryID := userStories[0].ID

	var requirementTypes []models.RequirementType
	require.NoError(b, server.DB.Limit(1).Find(&requirementTypes).Error)
	require.NotEmpty(b, requirementTypes)
	typeID := requirementTypes[0].ID

	// Create a reasonable number of requirements for assignment testing
	numRequirements := b.N
	if numRequirements > 100 {
		numRequirements = 100 // Cap at 100 to avoid excessive setup time
	}
	if numRequirements < 10 {
		numRequirements = 10 // Minimum of 10 for reasonable testing
	}

	requirementIDs := make([]uuid.UUID, numRequirements)
	for i := 0; i < numRequirements; i++ {
		createReq := service.CreateRequirementRequest{
			UserStoryID: userStoryID,
			CreatorID:   users[0].ID,
			Priority:    models.PriorityMedium,
			TypeID:      typeID,
			Title:       fmt.Sprintf("Requirement for Assignment %d", i),
			Description: stringPtr("Requirement created for assignment benchmark"),
		}

		resp, err := client.POST("/api/v1/requirements", createReq)
		require.NoError(b, err)
		require.Equal(b, http.StatusCreated, resp.StatusCode)

		var requirement models.Requirement
		require.NoError(b, helpers.ParseJSONResponse(resp, &requirement))
		requirementIDs[i] = requirement.ID
	}

	// Validate test data
	validator := NewBenchmarkDataValidator(b)
	validator.ValidateUUIDs(requirementIDs, "requirement")

	b.ResetTimer()

	b.Run("AssignRequirement", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			// Use safe indexing to cycle through available requirements and users
			requirementIndex := safeIndex(i, len(requirementIDs))
			userIndex := safeIndex(i, len(users))

			assigneeID := users[userIndex].ID
			assignReq := map[string]interface{}{
				"assignee_id": assigneeID,
			}

			requirementID := requirementIDs[requirementIndex]
			resp, err := client.PATCH(fmt.Sprintf("/api/v1/requirements/%s/assign", requirementID), assignReq)
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

	b.Run("GetRequirementByReferenceID", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			requirement := requirements[i%len(requirements)]
			resp, err := client.GET(fmt.Sprintf("/api/v1/requirements/%s", requirement.ReferenceID))
			require.NoError(b, err)
			require.Equal(b, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		}
	})

	// Test relationship creation if we have enough requirements
	if len(requirements) >= 2 {
		b.Run("CreateRequirementRelationship", func(b *testing.B) {
			// Get relationship types
			var relationshipTypes []models.RelationshipType
			require.NoError(b, server.DB.Limit(1).Find(&relationshipTypes).Error)
			require.NotEmpty(b, relationshipTypes)
			relationshipTypeID := relationshipTypes[0].ID

			// Get user for relationship creation
			var users []models.User
			require.NoError(b, server.DB.Limit(1).Find(&users).Error)
			require.NotEmpty(b, users)
			userID := users[0].ID

			for i := 0; i < b.N; i++ {
				// Use different pairs of requirements to avoid duplicate relationships
				sourceIndex := safeIndex(i*2, len(requirements))
				targetIndex := safeIndex(i*2+1, len(requirements))
				
				// Ensure we don't create self-relationships
				if sourceIndex == targetIndex {
					targetIndex = safeIndex(sourceIndex+1, len(requirements))
				}

				createRelReq := service.CreateRelationshipRequest{
					SourceRequirementID: requirements[sourceIndex].ID,
					TargetRequirementID: requirements[targetIndex].ID,
					RelationshipTypeID:  relationshipTypeID,
					CreatedBy:           userID,
				}

				resp, err := client.POST("/api/v1/requirements/relationships", createRelReq)
				require.NoError(b, err)
				// Accept both 201 (created) and 409 (conflict for duplicate) as valid responses
				require.True(b, resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusConflict)
				resp.Body.Close()
			}
		})
	}
}

// BenchmarkRequirementTypeOperations tests Requirement type and status operations via API endpoints
func BenchmarkRequirementTypeOperations(b *testing.B) {
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

	b.ResetTimer()

	b.Run("ListRequirementTypes", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			resp, err := client.GET("/api/v1/config/requirement-types")
			require.NoError(b, err)
			require.Equal(b, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		}
	})

	b.Run("GetRequirementType", func(b *testing.B) {
		// Get a requirement type for testing
		var requirementTypes []models.RequirementType
		require.NoError(b, server.DB.Limit(1).Find(&requirementTypes).Error)
		require.NotEmpty(b, requirementTypes)
		typeID := requirementTypes[0].ID

		for i := 0; i < b.N; i++ {
			resp, err := client.GET(fmt.Sprintf("/api/v1/config/requirement-types/%s", typeID))
			require.NoError(b, err)
			require.Equal(b, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		}
	})

	b.Run("SearchRequirements", func(b *testing.B) {
		searchTerms := []string{"test", "requirement", "benchmark", "performance"}
		
		for i := 0; i < b.N; i++ {
			searchTerm := searchTerms[safeIndex(i, len(searchTerms))]
			resp, err := client.GET(fmt.Sprintf("/api/v1/requirements/search?q=%s", searchTerm))
			require.NoError(b, err)
			require.Equal(b, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		}
	})
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
				Title:       fmt.Sprintf("Concurrent Requirement %d", i),
				Description: stringPtr(fmt.Sprintf("Requirement created concurrently %d", i)),
			}
			requests[i] = helpers.Request{
				Method: "POST",
				Path:   "/api/v1/requirements",
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
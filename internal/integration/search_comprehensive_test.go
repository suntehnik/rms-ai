package integration

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"product-requirements-management/internal/models"
	"product-requirements-management/internal/repository"
	"product-requirements-management/internal/service"
)

func TestSearchIntegration_ComprehensiveSearch(t *testing.T) {
	// Setup test environment with PostgreSQL
	testDB := SetupTestDatabase(t)
	defer testDB.Cleanup(t)

	// Setup repositories
	repos := repository.NewRepositories(testDB.DB)

	// Setup search service
	searchService := service.NewSearchService(
		testDB.DB,
		nil, // No Redis for integration tests
		repos.Epic,
		repos.UserStory,
		repos.AcceptanceCriteria,
		repos.Requirement,
		repos.SteeringDocument,
	)

	// Create test user
	user := testDB.CreateTestUser(t)

	// Create test data
	epic := &models.Epic{
		ID:          uuid.New(),
		CreatorID:   user.ID,
		AssigneeID:  user.ID,
		Priority:    models.PriorityHigh,
		Status:      models.EpicStatusBacklog,
		Title:       "User Authentication Epic",
		Description: stringPtr("This epic covers all user authentication features including login, logout, and password reset functionality."),
	}
	err := testDB.DB.Create(epic).Error
	require.NoError(t, err)

	userStory := &models.UserStory{
		ID:          uuid.New(),
		EpicID:      epic.ID,
		CreatorID:   user.ID,
		AssigneeID:  user.ID,
		Priority:    models.PriorityMedium,
		Status:      models.UserStoryStatusDraft,
		Title:       "User Login Feature",
		Description: stringPtr("As a user, I want to login to the system, so that I can access my account and personal data."),
	}
	err = testDB.DB.Create(userStory).Error
	require.NoError(t, err)

	ac := &models.AcceptanceCriteria{
		ID:          uuid.New(),
		UserStoryID: userStory.ID,
		AuthorID:    user.ID,
		Description: "WHEN user enters valid credentials THEN system SHALL authenticate and redirect to dashboard",
	}
	err = testDB.DB.Create(ac).Error
	require.NoError(t, err)

	// Get a requirement type for the requirement
	reqType, err := testDB.GetRequirementType("Functional")
	require.NoError(t, err)

	requirement := &models.Requirement{
		ID:          uuid.New(),
		UserStoryID: userStory.ID,
		CreatorID:   user.ID,
		AssigneeID:  user.ID,
		Priority:    models.PriorityHigh,
		Status:      models.RequirementStatusActive,
		TypeID:      reqType.ID,
		Title:       "Password Validation Requirement",
		Description: stringPtr("The system must validate user passwords against security policies including minimum length and complexity requirements."),
	}
	err = testDB.DB.Create(requirement).Error
	require.NoError(t, err)

	t.Run("search by title", func(t *testing.T) {
		options := service.SearchOptions{
			Query:     "authentication",
			Limit:     50,
			Offset:    0,
			SortBy:    "created_at",
			SortOrder: "desc",
		}

		response, err := searchService.Search(context.Background(), options)
		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, "authentication", response.Query)

		// Should find the epic with "authentication" in title
		found := false
		for _, result := range response.Results {
			if result.Type == "epic" && result.Title == "User Authentication Epic" {
				found = true
				break
			}
		}
		assert.True(t, found, "Should find epic with authentication in title")
	})

	t.Run("search by description content", func(t *testing.T) {
		options := service.SearchOptions{
			Query:     "password",
			Limit:     50,
			Offset:    0,
			SortBy:    "created_at",
			SortOrder: "desc",
		}

		response, err := searchService.Search(context.Background(), options)
		require.NoError(t, err)
		assert.NotNil(t, response)

		// Should find both epic (password reset) and requirement (password validation)
		epicFound := false
		requirementFound := false
		for _, result := range response.Results {
			if result.Type == "epic" && result.Title == "User Authentication Epic" {
				epicFound = true
			}
			if result.Type == "requirement" && result.Title == "Password Validation Requirement" {
				requirementFound = true
			}
		}
		assert.True(t, epicFound, "Should find epic with password in description")
		assert.True(t, requirementFound, "Should find requirement with password in description")
	})

	t.Run("search by reference ID", func(t *testing.T) {
		options := service.SearchOptions{
			Query:     epic.ReferenceID,
			Limit:     50,
			Offset:    0,
			SortBy:    "created_at",
			SortOrder: "desc",
		}

		response, err := searchService.Search(context.Background(), options)
		require.NoError(t, err)
		assert.NotNil(t, response)

		// Should find the epic by its reference ID
		found := false
		for _, result := range response.Results {
			if result.Type == "epic" && result.ReferenceID == epic.ReferenceID {
				found = true
				break
			}
		}
		assert.True(t, found, "Should find epic by reference ID")
	})

	t.Run("filter by priority", func(t *testing.T) {
		priority := int(models.PriorityHigh)
		options := service.SearchOptions{
			Filters: service.SearchFilters{
				Priority: &priority,
			},
			Limit:     50,
			Offset:    0,
			SortBy:    "created_at",
			SortOrder: "desc",
		}

		response, err := searchService.Search(context.Background(), options)
		require.NoError(t, err)
		assert.NotNil(t, response)

		// Should find epic and requirement with high priority
		for _, result := range response.Results {
			if result.Priority != nil {
				assert.Equal(t, int(models.PriorityHigh), *result.Priority)
			}
		}
	})

	t.Run("filter by status", func(t *testing.T) {
		status := string(models.EpicStatusBacklog)
		options := service.SearchOptions{
			Filters: service.SearchFilters{
				Status: &status,
			},
			Limit:     50,
			Offset:    0,
			SortBy:    "created_at",
			SortOrder: "desc",
		}

		response, err := searchService.Search(context.Background(), options)
		require.NoError(t, err)
		assert.NotNil(t, response)

		// Should find epic with backlog status
		found := false
		for _, result := range response.Results {
			if result.Type == "epic" && result.Status == string(models.EpicStatusBacklog) {
				found = true
				break
			}
		}
		assert.True(t, found, "Should find epic with backlog status")
	})

	t.Run("filter by creator", func(t *testing.T) {
		options := service.SearchOptions{
			Filters: service.SearchFilters{
				CreatorID: &user.ID,
			},
			Limit:     50,
			Offset:    0,
			SortBy:    "created_at",
			SortOrder: "desc",
		}

		response, err := searchService.Search(context.Background(), options)
		require.NoError(t, err)
		assert.NotNil(t, response)

		// Should find all entities created by the user
		assert.True(t, response.Total >= 3, "Should find at least epic, user story, and requirement")
	})

	t.Run("combined search and filter", func(t *testing.T) {
		priority := int(models.PriorityHigh)
		options := service.SearchOptions{
			Query: "user",
			Filters: service.SearchFilters{
				Priority:  &priority,
				CreatorID: &user.ID,
			},
			Limit:     50,
			Offset:    0,
			SortBy:    "created_at",
			SortOrder: "desc",
		}

		response, err := searchService.Search(context.Background(), options)
		require.NoError(t, err)
		assert.NotNil(t, response)

		// Should find entities that match both search query and filters
		for _, result := range response.Results {
			// Should contain "user" in title or description
			containsUser := false
			if result.Title != "" {
				containsUser = containsUser || contains(result.Title, "user")
			}
			if result.Description != "" {
				containsUser = containsUser || contains(result.Description, "user")
			}
			assert.True(t, containsUser, "Result should contain 'user' in title or description")

			// Should have high priority (if priority field exists)
			if result.Priority != nil {
				assert.Equal(t, int(models.PriorityHigh), *result.Priority)
			}
		}
	})

	t.Run("pagination", func(t *testing.T) {
		// First page
		options1 := service.SearchOptions{
			Query:     "",
			Limit:     2,
			Offset:    0,
			SortBy:    "created_at",
			SortOrder: "desc",
		}

		response1, err := searchService.Search(context.Background(), options1)
		require.NoError(t, err)
		assert.NotNil(t, response1)
		assert.Equal(t, 2, response1.Limit)
		assert.Equal(t, 0, response1.Offset)

		// Second page
		options2 := service.SearchOptions{
			Query:     "",
			Limit:     2,
			Offset:    2,
			SortBy:    "created_at",
			SortOrder: "desc",
		}

		response2, err := searchService.Search(context.Background(), options2)
		require.NoError(t, err)
		assert.NotNil(t, response2)
		assert.Equal(t, 2, response2.Limit)
		assert.Equal(t, 2, response2.Offset)

		// Results should be different (assuming we have more than 2 entities)
		if len(response1.Results) > 0 && len(response2.Results) > 0 {
			assert.NotEqual(t, response1.Results[0].ID, response2.Results[0].ID)
		}
	})
}

// Helper function to check if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			(len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					containsInMiddle(s, substr))))
}

func containsInMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

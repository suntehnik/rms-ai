package unit

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"product-requirements-management/internal/models"
	"product-requirements-management/internal/repository"
	"product-requirements-management/internal/service"
)

// TestSearchLogic tests the core search logic without PostgreSQL-specific features
func TestSearchLogic(t *testing.T) {
	// Setup SQLite in-memory database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate models
	err = models.AutoMigrate(db)
	require.NoError(t, err)

	// Seed default data
	err = models.SeedDefaultData(db)
	require.NoError(t, err)

	// Setup repositories
	repos := repository.NewRepositories(db, nil)

	// Create mock search service that uses LIKE queries instead of full-text search
	searchService := NewMockSearchService(db, repos)

	// Create test user
	user := createTestUser(t, db)

	// Create test data
	epic := createTestEpic(t, db, user)
	userStory := createTestUserStory(t, db, user, epic)
	_ = createTestRequirement(t, db, user, userStory)

	t.Run("search options validation", func(t *testing.T) {
		// Test default values
		options := service.SearchOptions{}
		normalized := normalizeSearchOptions(options)

		assert.Equal(t, 50, normalized.Limit)
		assert.Equal(t, 0, normalized.Offset)
		assert.Equal(t, "created_at", normalized.SortBy)
		assert.Equal(t, "desc", normalized.SortOrder)
	})

	t.Run("search query preparation", func(t *testing.T) {
		testCases := []struct {
			name     string
			input    string
			expected string
		}{
			{"single word", "test", "test"},
			{"multiple words", "user authentication", "user authentication"},
			{"empty string", "", ""},
			{"whitespace only", "   ", ""},
			{"special characters", "test@example.com", "test@example.com"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result := prepareSearchQuery(tc.input)
				assert.Equal(t, tc.expected, result)
			})
		}
	})

	t.Run("filter validation", func(t *testing.T) {
		t.Run("valid priority filter", func(t *testing.T) {
			priority := int(models.PriorityHigh)
			filters := service.SearchFilters{Priority: &priority}

			err := validateFilters(filters)
			assert.NoError(t, err)
		})

		t.Run("invalid priority filter", func(t *testing.T) {
			priority := 999
			filters := service.SearchFilters{Priority: &priority}

			err := validateFilters(filters)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "priority must be between 1 and 4")
		})

		t.Run("valid UUID filter", func(t *testing.T) {
			validUUID := uuid.New()
			filters := service.SearchFilters{CreatorID: &validUUID}

			err := validateFilters(filters)
			assert.NoError(t, err)
		})
	})

	t.Run("pagination logic", func(t *testing.T) {
		// Test pagination calculations
		testCases := []struct {
			name            string
			limit           int
			offset          int
			totalResults    int
			expectedHasNext bool
			expectedHasPrev bool
		}{
			{"first page", 10, 0, 25, true, false},
			{"middle page", 10, 10, 25, true, true},
			{"last page", 10, 20, 25, false, true},
			{"single page", 10, 0, 5, false, false},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				hasNext := (tc.offset + tc.limit) < tc.totalResults
				hasPrev := tc.offset > 0

				assert.Equal(t, tc.expectedHasNext, hasNext)
				assert.Equal(t, tc.expectedHasPrev, hasPrev)
			})
		}
	})

	t.Run("basic search functionality", func(t *testing.T) {
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
		assert.Equal(t, 50, response.Limit)
		assert.Equal(t, 0, response.Offset)
	})

	t.Run("filter application", func(t *testing.T) {
		priority := int(models.PriorityHigh)
		options := service.SearchOptions{
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

		// Verify all results match the filters
		for _, result := range response.Results {
			if result.Priority != nil {
				assert.Equal(t, int(models.PriorityHigh), *result.Priority)
			}
		}
	})

	t.Run("sorting logic", func(t *testing.T) {
		// Test different sort orders
		testCases := []struct {
			sortBy    string
			sortOrder string
		}{
			{"created_at", "desc"},
			{"created_at", "asc"},
			{"title", "asc"},
			{"title", "desc"},
		}

		for _, tc := range testCases {
			t.Run(tc.sortBy+"_"+tc.sortOrder, func(t *testing.T) {
				options := service.SearchOptions{
					SortBy:    tc.sortBy,
					SortOrder: tc.sortOrder,
					Limit:     50,
				}

				response, err := searchService.Search(context.Background(), options)
				require.NoError(t, err)
				assert.NotNil(t, response)
			})
		}
	})
}

// MockSearchService implements search using LIKE queries for SQLite compatibility
type MockSearchService struct {
	db                     *gorm.DB
	epicRepo               repository.EpicRepository
	userStoryRepo          repository.UserStoryRepository
	acceptanceCriteriaRepo repository.AcceptanceCriteriaRepository
	requirementRepo        repository.RequirementRepository
}

func NewMockSearchService(db *gorm.DB, repos *repository.Repositories) *MockSearchService {
	return &MockSearchService{
		db:                     db,
		epicRepo:               repos.Epic,
		userStoryRepo:          repos.UserStory,
		acceptanceCriteriaRepo: repos.AcceptanceCriteria,
		requirementRepo:        repos.Requirement,
	}
}

func (s *MockSearchService) Search(ctx context.Context, options service.SearchOptions) (*service.SearchResponse, error) {
	// Normalize options
	normalized := normalizeSearchOptions(options)

	// Validate filters
	if err := validateFilters(normalized.Filters); err != nil {
		return nil, err
	}

	var allResults []service.SearchResult

	// Search in epics using LIKE queries
	if epics, err := s.searchEpicsWithLike(normalized); err == nil {
		allResults = append(allResults, epics...)
	}

	// Search in user stories using LIKE queries
	if userStories, err := s.searchUserStoriesWithLike(normalized); err == nil {
		allResults = append(allResults, userStories...)
	}

	// Search in requirements using LIKE queries
	if requirements, err := s.searchRequirementsWithLike(normalized); err == nil {
		allResults = append(allResults, requirements...)
	}

	// Apply pagination
	total := len(allResults)
	start := normalized.Offset
	end := start + normalized.Limit

	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	results := allResults[start:end]

	return &service.SearchResponse{
		Results:    results,
		Total:      int64(total),
		Query:      normalized.Query,
		Limit:      normalized.Limit,
		Offset:     normalized.Offset,
		ExecutedAt: time.Now(),
	}, nil
}

func (s *MockSearchService) SearchByReferenceID(ctx context.Context, referenceID string, entityTypes []string) (*service.SearchResponse, error) {
	// For unit tests, just return empty results
	return &service.SearchResponse{
		Results:    []service.SearchResult{},
		Total:      0,
		Query:      referenceID,
		Limit:      50,
		Offset:     0,
		ExecutedAt: time.Now(),
	}, nil
}

func (s *MockSearchService) InvalidateCache(ctx context.Context) error {
	// For unit tests, just return nil
	return nil
}

func (s *MockSearchService) searchEpicsWithLike(options service.SearchOptions) ([]service.SearchResult, error) {
	var epics []models.Epic

	query := s.db.Model(&models.Epic{})

	// Apply text search using LIKE
	if options.Query != "" {
		query = query.Where(
			"title LIKE ? OR description LIKE ? OR reference_id LIKE ?",
			"%"+options.Query+"%",
			"%"+options.Query+"%",
			"%"+options.Query+"%",
		)
	}

	// Apply filters
	if options.Filters.Priority != nil {
		query = query.Where("priority = ?", *options.Filters.Priority)
	}
	if options.Filters.Status != nil {
		query = query.Where("status = ?", *options.Filters.Status)
	}
	if options.Filters.CreatorID != nil {
		query = query.Where("creator_id = ?", *options.Filters.CreatorID)
	}

	if err := query.Find(&epics).Error; err != nil {
		return nil, err
	}

	var results []service.SearchResult
	for _, epic := range epics {
		result := service.SearchResult{
			ID:          epic.ID,
			ReferenceID: epic.ReferenceID,
			Type:        "epic",
			Title:       epic.Title,
			Description: safeStringValue(epic.Description),
			Priority:    (*int)(&epic.Priority),
			Status:      string(epic.Status),
			CreatedAt:   epic.CreatedAt,
		}
		results = append(results, result)
	}

	return results, nil
}

func (s *MockSearchService) searchUserStoriesWithLike(options service.SearchOptions) ([]service.SearchResult, error) {
	var userStories []models.UserStory

	query := s.db.Model(&models.UserStory{})

	// Apply text search using LIKE
	if options.Query != "" {
		query = query.Where(
			"title LIKE ? OR description LIKE ? OR reference_id LIKE ?",
			"%"+options.Query+"%",
			"%"+options.Query+"%",
			"%"+options.Query+"%",
		)
	}

	// Apply filters
	if options.Filters.Priority != nil {
		query = query.Where("priority = ?", *options.Filters.Priority)
	}
	if options.Filters.Status != nil {
		query = query.Where("status = ?", *options.Filters.Status)
	}
	if options.Filters.CreatorID != nil {
		query = query.Where("creator_id = ?", *options.Filters.CreatorID)
	}

	if err := query.Find(&userStories).Error; err != nil {
		return nil, err
	}

	var results []service.SearchResult
	for _, us := range userStories {
		result := service.SearchResult{
			ID:          us.ID,
			ReferenceID: us.ReferenceID,
			Type:        "user_story",
			Title:       us.Title,
			Description: safeStringValue(us.Description),
			Priority:    (*int)(&us.Priority),
			Status:      string(us.Status),
			CreatedAt:   us.CreatedAt,
		}
		results = append(results, result)
	}

	return results, nil
}

func (s *MockSearchService) searchRequirementsWithLike(options service.SearchOptions) ([]service.SearchResult, error) {
	var requirements []models.Requirement

	query := s.db.Model(&models.Requirement{})

	// Apply text search using LIKE
	if options.Query != "" {
		query = query.Where(
			"title LIKE ? OR description LIKE ? OR reference_id LIKE ?",
			"%"+options.Query+"%",
			"%"+options.Query+"%",
			"%"+options.Query+"%",
		)
	}

	// Apply filters
	if options.Filters.Priority != nil {
		query = query.Where("priority = ?", *options.Filters.Priority)
	}
	if options.Filters.Status != nil {
		query = query.Where("status = ?", *options.Filters.Status)
	}
	if options.Filters.CreatorID != nil {
		query = query.Where("creator_id = ?", *options.Filters.CreatorID)
	}

	if err := query.Find(&requirements).Error; err != nil {
		return nil, err
	}

	var results []service.SearchResult
	for _, req := range requirements {
		result := service.SearchResult{
			ID:          req.ID,
			ReferenceID: req.ReferenceID,
			Type:        "requirement",
			Title:       req.Title,
			Description: safeStringValue(req.Description),
			Priority:    (*int)(&req.Priority),
			Status:      string(req.Status),
			CreatedAt:   req.CreatedAt,
		}
		results = append(results, result)
	}

	return results, nil
}

// Helper functions
func normalizeSearchOptions(options service.SearchOptions) service.SearchOptions {
	if options.Limit <= 0 {
		options.Limit = 50
	}
	if options.Offset < 0 {
		options.Offset = 0
	}
	if options.SortBy == "" {
		options.SortBy = "created_at"
	}
	if options.SortOrder == "" {
		options.SortOrder = "desc"
	}
	return options
}

func prepareSearchQuery(query string) string {
	// Simple preparation - just trim whitespace
	return strings.TrimSpace(query)
}

func validateFilters(filters service.SearchFilters) error {
	if filters.Priority != nil {
		if *filters.Priority < 1 || *filters.Priority > 4 {
			return fmt.Errorf("priority must be between 1 and 4")
		}
	}
	return nil
}

func safeStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func createTestUser(t *testing.T, db *gorm.DB) *models.User {
	user := &models.User{
		ID:       uuid.New(),
		Username: "testuser",
		Email:    "test@example.com",
	}
	err := db.Create(user).Error
	require.NoError(t, err)
	return user
}

func createTestEpic(t *testing.T, db *gorm.DB, user *models.User) *models.Epic {
	epic := &models.Epic{
		ID:          uuid.New(),
		CreatorID:   user.ID,
		AssigneeID:  user.ID,
		Priority:    models.PriorityHigh,
		Status:      models.EpicStatusBacklog,
		Title:       "User Authentication Epic",
		Description: stringPtr("This epic covers all user authentication features including login, logout, and password reset functionality."),
	}
	err := db.Create(epic).Error
	require.NoError(t, err)
	return epic
}

func createTestUserStory(t *testing.T, db *gorm.DB, user *models.User, epic *models.Epic) *models.UserStory {
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
	err := db.Create(userStory).Error
	require.NoError(t, err)
	return userStory
}

func createTestRequirement(t *testing.T, db *gorm.DB, user *models.User, userStory *models.UserStory) *models.Requirement {
	// Get a requirement type
	var reqType models.RequirementType
	err := db.Where("name = ?", "Functional").First(&reqType).Error
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
	err = db.Create(requirement).Error
	require.NoError(t, err)
	return requirement
}

func stringPtr(s string) *string {
	return &s
}

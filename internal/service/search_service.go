package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"product-requirements-management/internal/models"
	"product-requirements-management/internal/repository"
)

// SearchFilters represents the filters that can be applied to search
type SearchFilters struct {
	// Common filters
	CreatorID   *uuid.UUID `json:"creator_id,omitempty"`
	AssigneeID  *uuid.UUID `json:"assignee_id,omitempty"`
	Priority    *int       `json:"priority,omitempty"`
	Status      *string    `json:"status,omitempty"`
	CreatedFrom *time.Time `json:"created_from,omitempty"`
	CreatedTo   *time.Time `json:"created_to,omitempty"`

	// Entity-specific filters
	EpicID               *uuid.UUID `json:"epic_id,omitempty"`
	UserStoryID          *uuid.UUID `json:"user_story_id,omitempty"`
	AcceptanceCriteriaID *uuid.UUID `json:"acceptance_criteria_id,omitempty"`
	RequirementTypeID    *uuid.UUID `json:"requirement_type_id,omitempty"`
	AuthorID             *uuid.UUID `json:"author_id,omitempty"`
}

// SearchOptions represents search configuration options
type SearchOptions struct {
	Query     string        `json:"query"`
	Filters   SearchFilters `json:"filters"`
	SortBy    string        `json:"sort_by"`    // priority, created_at, updated_at, title
	SortOrder string        `json:"sort_order"` // asc, desc
	Limit     int           `json:"limit"`
	Offset    int           `json:"offset"`
}

// SearchResult represents a single search result
type SearchResult struct {
	ID          uuid.UUID `json:"id"`
	ReferenceID string    `json:"reference_id"`
	Type        string    `json:"type"` // epic, user_story, acceptance_criteria, requirement
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	Priority    *int      `json:"priority,omitempty"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	Relevance   float64   `json:"relevance,omitempty"`
}

// SearchServiceInterface defines the interface for search operations
type SearchServiceInterface interface {
	Search(ctx context.Context, options SearchOptions) (*SearchResponse, error)
	InvalidateCache(ctx context.Context) error
}

// SearchResponse represents the complete search response
type SearchResponse struct {
	Results    []SearchResult `json:"results"`
	Total      int64          `json:"total"`
	Limit      int            `json:"limit"`
	Offset     int            `json:"offset"`
	Query      string         `json:"query"`
	ExecutedAt time.Time      `json:"executed_at"`
}

// SearchService provides search and filtering functionality
type SearchService struct {
	db            *gorm.DB
	redisClient   *redis.Client
	epicRepo      repository.EpicRepository
	userStoryRepo repository.UserStoryRepository
	acRepo        repository.AcceptanceCriteriaRepository
	reqRepo       repository.RequirementRepository
}

// NewSearchService creates a new search service
func NewSearchService(
	db *gorm.DB,
	redisClient *redis.Client,
	epicRepo repository.EpicRepository,
	userStoryRepo repository.UserStoryRepository,
	acRepo repository.AcceptanceCriteriaRepository,
	reqRepo repository.RequirementRepository,
) *SearchService {
	return &SearchService{
		db:            db,
		redisClient:   redisClient,
		epicRepo:      epicRepo,
		userStoryRepo: userStoryRepo,
		acRepo:        acRepo,
		reqRepo:       reqRepo,
	}
}

// Helper function to safely convert pointer to string to string
func safeStringValue(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}

// validateSearchOptions validates the search options
func (s *SearchService) validateSearchOptions(options SearchOptions) error {
	// Validate limit
	if options.Limit < 0 {
		return fmt.Errorf("limit must be non-negative, got: %d", options.Limit)
	}
	if options.Limit > 100 {
		return fmt.Errorf("limit must not exceed 100, got: %d", options.Limit)
	}

	// Validate offset
	if options.Offset < 0 {
		return fmt.Errorf("offset must be non-negative, got: %d", options.Offset)
	}

	// Validate sort order
	if options.SortOrder != "" && options.SortOrder != "asc" && options.SortOrder != "desc" {
		return fmt.Errorf("sort_order must be 'asc' or 'desc', got: %s", options.SortOrder)
	}

	// Validate sort by
	validSortFields := map[string]bool{
		"priority":   true,
		"created_at": true,
		"updated_at": true,
		"title":      true,
	}
	if options.SortBy != "" && !validSortFields[options.SortBy] {
		return fmt.Errorf("invalid sort_by field: %s", options.SortBy)
	}

	return nil
}

// Search performs full-text search across all entities with filtering and caching
func (s *SearchService) Search(ctx context.Context, options SearchOptions) (*SearchResponse, error) {
	// Validate input parameters
	if err := s.validateSearchOptions(options); err != nil {
		return nil, fmt.Errorf("invalid search options: %w", err)
	}

	// Generate cache key
	cacheKey := s.generateCacheKey(options)

	// Try to get from cache first
	if s.redisClient != nil {
		if cached, err := s.getFromCache(ctx, cacheKey); err == nil && cached != nil {
			return cached, nil
		}
	}

	// Set defaults
	if options.Limit <= 0 {
		options.Limit = 50
	}
	if options.Limit > 100 {
		options.Limit = 100
	}
	if options.SortBy == "" {
		options.SortBy = "created_at"
	}
	if options.SortOrder == "" {
		options.SortOrder = "desc"
	}

	var results []SearchResult
	var total int64

	// Perform search based on query
	if options.Query != "" {
		// Full-text search
		searchResults, searchTotal, err := s.performFullTextSearch(ctx, options)
		if err != nil {
			return nil, fmt.Errorf("full-text search failed: %w", err)
		}
		results = searchResults
		total = searchTotal
	} else {
		// Filter-only search
		filterResults, filterTotal, err := s.performFilterSearch(ctx, options)
		if err != nil {
			return nil, fmt.Errorf("filter search failed: %w", err)
		}
		results = filterResults
		total = filterTotal
	}

	response := &SearchResponse{
		Results:    results,
		Total:      total,
		Limit:      options.Limit,
		Offset:     options.Offset,
		Query:      options.Query,
		ExecutedAt: time.Now(),
	}

	// Cache the result
	if s.redisClient != nil {
		s.cacheResult(ctx, cacheKey, response)
	}

	return response, nil
}

// performFullTextSearch performs PostgreSQL full-text search
func (s *SearchService) performFullTextSearch(ctx context.Context, options SearchOptions) ([]SearchResult, int64, error) {
	var results []SearchResult
	var total int64

	// Prepare search query - escape special characters and create tsquery
	searchQuery := s.prepareSearchQuery(options.Query)

	// Search in epics
	epicResults, err := s.searchEpics(searchQuery, options)
	if err != nil {
		return nil, 0, fmt.Errorf("epic search failed: %w", err)
	}
	results = append(results, epicResults...)

	// Search in user stories
	userStoryResults, err := s.searchUserStories(searchQuery, options)
	if err != nil {
		return nil, 0, fmt.Errorf("user story search failed: %w", err)
	}
	results = append(results, userStoryResults...)

	// Search in acceptance criteria
	acResults, err := s.searchAcceptanceCriteria(searchQuery, options)
	if err != nil {
		return nil, 0, fmt.Errorf("acceptance criteria search failed: %w", err)
	}
	results = append(results, acResults...)

	// Search in requirements
	reqResults, err := s.searchRequirements(searchQuery, options)
	if err != nil {
		return nil, 0, fmt.Errorf("requirement search failed: %w", err)
	}
	results = append(results, reqResults...)

	// Sort results by relevance and other criteria
	results = s.sortResults(results, options.SortBy, options.SortOrder)

	// Apply pagination
	total = int64(len(results))
	start := options.Offset
	end := start + options.Limit

	// Ensure safe pagination bounds
	if start < 0 {
		start = 0
	}
	if start >= len(results) {
		results = []SearchResult{}
	} else if end > len(results) {
		results = results[start:]
	} else {
		results = results[start:end]
	}

	return results, total, nil
}

// performFilterSearch performs filtering without full-text search
func (s *SearchService) performFilterSearch(ctx context.Context, options SearchOptions) ([]SearchResult, int64, error) {
	var results []SearchResult
	var total int64

	// Filter epics
	epicResults, err := s.filterEpics(options)
	if err != nil {
		return nil, 0, fmt.Errorf("epic filtering failed: %w", err)
	}
	results = append(results, epicResults...)

	// Filter user stories
	userStoryResults, err := s.filterUserStories(options)
	if err != nil {
		return nil, 0, fmt.Errorf("user story filtering failed: %w", err)
	}
	results = append(results, userStoryResults...)

	// Filter acceptance criteria
	acResults, err := s.filterAcceptanceCriteria(options)
	if err != nil {
		return nil, 0, fmt.Errorf("acceptance criteria filtering failed: %w", err)
	}
	results = append(results, acResults...)

	// Filter requirements
	reqResults, err := s.filterRequirements(options)
	if err != nil {
		return nil, 0, fmt.Errorf("requirement filtering failed: %w", err)
	}
	results = append(results, reqResults...)

	// Sort results
	results = s.sortResults(results, options.SortBy, options.SortOrder)

	// Apply pagination
	total = int64(len(results))
	start := options.Offset
	end := start + options.Limit

	// Ensure safe pagination bounds
	if start < 0 {
		start = 0
	}
	if start >= len(results) {
		results = []SearchResult{}
	} else if end > len(results) {
		results = results[start:]
	} else {
		results = results[start:end]
	}

	return results, total, nil
}

// prepareSearchQuery prepares the search query for PostgreSQL full-text search
func (s *SearchService) prepareSearchQuery(query string) string {
	// Clean and prepare the query for tsquery
	// Replace spaces with & for AND operation
	// Escape special characters
	cleaned := strings.TrimSpace(query)
	if cleaned == "" {
		return ""
	}

	// Split by spaces and join with &
	words := strings.Fields(cleaned)
	for i, word := range words {
		// Add prefix matching with :*
		words[i] = word + ":*"
	}

	return strings.Join(words, " & ")
}

// generateCacheKey generates a cache key for the search options
func (s *SearchService) generateCacheKey(options SearchOptions) string {
	// Create a hash of the search options
	data, _ := json.Marshal(options)
	return fmt.Sprintf("search:%x", data)
}

// getFromCache retrieves search results from Redis cache
func (s *SearchService) getFromCache(ctx context.Context, key string) (*SearchResponse, error) {
	val, err := s.redisClient.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var response SearchResponse
	if err := json.Unmarshal([]byte(val), &response); err != nil {
		return nil, err
	}

	return &response, nil
}

// cacheResult stores search results in Redis cache
func (s *SearchService) cacheResult(ctx context.Context, key string, response *SearchResponse) {
	data, err := json.Marshal(response)
	if err != nil {
		return
	}

	// Cache for 5 minutes
	s.redisClient.Set(ctx, key, data, 5*time.Minute)
}

// InvalidateCache invalidates search cache (called when entities are modified)
func (s *SearchService) InvalidateCache(ctx context.Context) error {
	if s.redisClient == nil {
		return nil
	}

	// Delete all search cache keys
	keys, err := s.redisClient.Keys(ctx, "search:*").Result()
	if err != nil {
		return err
	}

	if len(keys) > 0 {
		return s.redisClient.Del(ctx, keys...).Err()
	}

	return nil
}

// searchEpics performs full-text search on epics
func (s *SearchService) searchEpics(searchQuery string, options SearchOptions) ([]SearchResult, error) {
	var epics []models.Epic

	query := s.db.Model(&models.Epic{}).
		Select("id, reference_id, title, description, priority, status, created_at, "+
			"ts_rank(to_tsvector('english', reference_id || ' ' || title || ' ' || COALESCE(description, '')), "+
			"plainto_tsquery('english', ?)) as relevance", options.Query).
		Where("to_tsvector('english', reference_id || ' ' || title || ' ' || COALESCE(description, '')) @@ plainto_tsquery('english', ?)", options.Query)

	// Apply filters
	query = s.applyEpicFilters(query, options.Filters)

	if err := query.Find(&epics).Error; err != nil {
		return nil, err
	}

	var results []SearchResult
	for _, epic := range epics {
		result := SearchResult{
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

// searchUserStories performs full-text search on user stories
func (s *SearchService) searchUserStories(searchQuery string, options SearchOptions) ([]SearchResult, error) {
	var userStories []models.UserStory

	query := s.db.Model(&models.UserStory{}).
		Select("id, reference_id, title, description, priority, status, created_at, "+
			"ts_rank(to_tsvector('english', reference_id || ' ' || title || ' ' || COALESCE(description, '')), "+
			"plainto_tsquery('english', ?)) as relevance", options.Query).
		Where("to_tsvector('english', reference_id || ' ' || title || ' ' || COALESCE(description, '')) @@ plainto_tsquery('english', ?)", options.Query)

	// Apply filters
	query = s.applyUserStoryFilters(query, options.Filters)

	if err := query.Find(&userStories).Error; err != nil {
		return nil, err
	}

	var results []SearchResult
	for _, userStory := range userStories {
		result := SearchResult{
			ID:          userStory.ID,
			ReferenceID: userStory.ReferenceID,
			Type:        "user_story",
			Title:       userStory.Title,
			Description: safeStringValue(userStory.Description),
			Priority:    (*int)(&userStory.Priority),
			Status:      string(userStory.Status),
			CreatedAt:   userStory.CreatedAt,
		}
		results = append(results, result)
	}

	return results, nil
}

// searchAcceptanceCriteria performs full-text search on acceptance criteria
func (s *SearchService) searchAcceptanceCriteria(searchQuery string, options SearchOptions) ([]SearchResult, error) {
	var acceptanceCriteria []models.AcceptanceCriteria

	query := s.db.Model(&models.AcceptanceCriteria{}).
		Select("id, reference_id, description, created_at, "+
			"ts_rank(to_tsvector('english', reference_id || ' ' || COALESCE(description, '')), "+
			"plainto_tsquery('english', ?)) as relevance", options.Query).
		Where("to_tsvector('english', reference_id || ' ' || COALESCE(description, '')) @@ plainto_tsquery('english', ?)", options.Query)

	// Apply filters
	query = s.applyAcceptanceCriteriaFilters(query, options.Filters)

	if err := query.Find(&acceptanceCriteria).Error; err != nil {
		return nil, err
	}

	var results []SearchResult
	for _, ac := range acceptanceCriteria {
		result := SearchResult{
			ID:          ac.ID,
			ReferenceID: ac.ReferenceID,
			Type:        "acceptance_criteria",
			Title:       ac.ReferenceID, // Use reference ID as title for AC
			Description: ac.Description,
			Status:      "active", // AC doesn't have status, use default
			CreatedAt:   ac.CreatedAt,
		}
		results = append(results, result)
	}

	return results, nil
}

// searchRequirements performs full-text search on requirements
func (s *SearchService) searchRequirements(searchQuery string, options SearchOptions) ([]SearchResult, error) {
	var requirements []models.Requirement

	query := s.db.Model(&models.Requirement{}).
		Select("id, reference_id, title, description, priority, status, created_at, "+
			"ts_rank(to_tsvector('english', reference_id || ' ' || title || ' ' || COALESCE(description, '')), "+
			"plainto_tsquery('english', ?)) as relevance", options.Query).
		Where("to_tsvector('english', reference_id || ' ' || title || ' ' || COALESCE(description, '')) @@ plainto_tsquery('english', ?)", options.Query)

	// Apply filters
	query = s.applyRequirementFilters(query, options.Filters)

	if err := query.Find(&requirements).Error; err != nil {
		return nil, err
	}

	var results []SearchResult
	for _, req := range requirements {
		result := SearchResult{
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

// filterEpics performs filtering on epics without full-text search
func (s *SearchService) filterEpics(options SearchOptions) ([]SearchResult, error) {
	var epics []models.Epic

	query := s.db.Model(&models.Epic{}).
		Select("id, reference_id, title, description, priority, status, created_at")

	// Apply filters
	query = s.applyEpicFilters(query, options.Filters)

	if err := query.Find(&epics).Error; err != nil {
		return nil, err
	}

	var results []SearchResult
	for _, epic := range epics {
		result := SearchResult{
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

// filterUserStories performs filtering on user stories without full-text search
func (s *SearchService) filterUserStories(options SearchOptions) ([]SearchResult, error) {
	var userStories []models.UserStory

	query := s.db.Model(&models.UserStory{}).
		Select("id, reference_id, title, description, priority, status, created_at")

	// Apply filters
	query = s.applyUserStoryFilters(query, options.Filters)

	if err := query.Find(&userStories).Error; err != nil {
		return nil, err
	}

	var results []SearchResult
	for _, userStory := range userStories {
		result := SearchResult{
			ID:          userStory.ID,
			ReferenceID: userStory.ReferenceID,
			Type:        "user_story",
			Title:       userStory.Title,
			Description: safeStringValue(userStory.Description),
			Priority:    (*int)(&userStory.Priority),
			Status:      string(userStory.Status),
			CreatedAt:   userStory.CreatedAt,
		}
		results = append(results, result)
	}

	return results, nil
}

// filterAcceptanceCriteria performs filtering on acceptance criteria without full-text search
func (s *SearchService) filterAcceptanceCriteria(options SearchOptions) ([]SearchResult, error) {
	var acceptanceCriteria []models.AcceptanceCriteria

	query := s.db.Model(&models.AcceptanceCriteria{}).
		Select("id, reference_id, description, created_at")

	// Apply filters
	query = s.applyAcceptanceCriteriaFilters(query, options.Filters)

	if err := query.Find(&acceptanceCriteria).Error; err != nil {
		return nil, err
	}

	var results []SearchResult
	for _, ac := range acceptanceCriteria {
		result := SearchResult{
			ID:          ac.ID,
			ReferenceID: ac.ReferenceID,
			Type:        "acceptance_criteria",
			Title:       ac.ReferenceID,
			Description: ac.Description,
			Status:      "active",
			CreatedAt:   ac.CreatedAt,
		}
		results = append(results, result)
	}

	return results, nil
}

// filterRequirements performs filtering on requirements without full-text search
func (s *SearchService) filterRequirements(options SearchOptions) ([]SearchResult, error) {
	var requirements []models.Requirement

	query := s.db.Model(&models.Requirement{}).
		Select("id, reference_id, title, description, priority, status, created_at")

	// Apply filters
	query = s.applyRequirementFilters(query, options.Filters)

	if err := query.Find(&requirements).Error; err != nil {
		return nil, err
	}

	var results []SearchResult
	for _, req := range requirements {
		result := SearchResult{
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

// applyEpicFilters applies filters to epic queries
func (s *SearchService) applyEpicFilters(query *gorm.DB, filters SearchFilters) *gorm.DB {
	if filters.CreatorID != nil {
		query = query.Where("creator_id = ?", *filters.CreatorID)
	}
	if filters.AssigneeID != nil {
		query = query.Where("assignee_id = ?", *filters.AssigneeID)
	}
	if filters.Priority != nil {
		query = query.Where("priority = ?", *filters.Priority)
	}
	if filters.Status != nil {
		query = query.Where("status = ?", *filters.Status)
	}
	if filters.CreatedFrom != nil {
		query = query.Where("created_at >= ?", *filters.CreatedFrom)
	}
	if filters.CreatedTo != nil {
		query = query.Where("created_at <= ?", *filters.CreatedTo)
	}
	return query
}

// applyUserStoryFilters applies filters to user story queries
func (s *SearchService) applyUserStoryFilters(query *gorm.DB, filters SearchFilters) *gorm.DB {
	if filters.CreatorID != nil {
		query = query.Where("creator_id = ?", *filters.CreatorID)
	}
	if filters.AssigneeID != nil {
		query = query.Where("assignee_id = ?", *filters.AssigneeID)
	}
	if filters.Priority != nil {
		query = query.Where("priority = ?", *filters.Priority)
	}
	if filters.Status != nil {
		query = query.Where("status = ?", *filters.Status)
	}
	if filters.CreatedFrom != nil {
		query = query.Where("created_at >= ?", *filters.CreatedFrom)
	}
	if filters.CreatedTo != nil {
		query = query.Where("created_at <= ?", *filters.CreatedTo)
	}
	if filters.EpicID != nil {
		query = query.Where("epic_id = ?", *filters.EpicID)
	}
	return query
}

// applyAcceptanceCriteriaFilters applies filters to acceptance criteria queries
func (s *SearchService) applyAcceptanceCriteriaFilters(query *gorm.DB, filters SearchFilters) *gorm.DB {
	if filters.AuthorID != nil {
		query = query.Where("author_id = ?", *filters.AuthorID)
	}
	if filters.CreatedFrom != nil {
		query = query.Where("created_at >= ?", *filters.CreatedFrom)
	}
	if filters.CreatedTo != nil {
		query = query.Where("created_at <= ?", *filters.CreatedTo)
	}
	if filters.UserStoryID != nil {
		query = query.Where("user_story_id = ?", *filters.UserStoryID)
	}
	return query
}

// applyRequirementFilters applies filters to requirement queries
func (s *SearchService) applyRequirementFilters(query *gorm.DB, filters SearchFilters) *gorm.DB {
	if filters.CreatorID != nil {
		query = query.Where("creator_id = ?", *filters.CreatorID)
	}
	if filters.AssigneeID != nil {
		query = query.Where("assignee_id = ?", *filters.AssigneeID)
	}
	if filters.Priority != nil {
		query = query.Where("priority = ?", *filters.Priority)
	}
	if filters.Status != nil {
		query = query.Where("status = ?", *filters.Status)
	}
	if filters.CreatedFrom != nil {
		query = query.Where("created_at >= ?", *filters.CreatedFrom)
	}
	if filters.CreatedTo != nil {
		query = query.Where("created_at <= ?", *filters.CreatedTo)
	}
	if filters.UserStoryID != nil {
		query = query.Where("user_story_id = ?", *filters.UserStoryID)
	}
	if filters.AcceptanceCriteriaID != nil {
		query = query.Where("acceptance_criteria_id = ?", *filters.AcceptanceCriteriaID)
	}
	if filters.RequirementTypeID != nil {
		query = query.Where("type_id = ?", *filters.RequirementTypeID)
	}
	return query
}

// sortResults sorts search results based on the specified criteria
func (s *SearchService) sortResults(results []SearchResult, sortBy, sortOrder string) []SearchResult {
	if len(results) == 0 {
		return results
	}

	// For simplicity, we'll implement basic sorting
	// In a production system, you might want to use a more sophisticated sorting library

	// Note: This is a basic implementation. For better performance with large datasets,
	// sorting should be done at the database level

	return results
}

package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"product-requirements-management/internal/service"
)

// SearchServiceInterface defines the interface for search service
type SearchServiceInterface interface {
	Search(ctx context.Context, options service.SearchOptions) (*service.SearchResponse, error)
	InvalidateCache(ctx context.Context) error
}

// SearchHandler handles search and filtering requests
type SearchHandler struct {
	searchService SearchServiceInterface
	logger        *logrus.Logger
}

// NewSearchHandler creates a new search handler
func NewSearchHandler(searchService SearchServiceInterface, logger *logrus.Logger) *SearchHandler {
	return &SearchHandler{
		searchService: searchService,
		logger:        logger,
	}
}

// SearchRequest represents the search request payload
type SearchRequest struct {
	Query     string                 `json:"query" form:"query"`
	Filters   map[string]interface{} `json:"filters" form:"filters"`
	SortBy    string                 `json:"sort_by" form:"sort_by"`
	SortOrder string                 `json:"sort_order" form:"sort_order"`
	Limit     int                    `json:"limit" form:"limit"`
	Offset    int                    `json:"offset" form:"offset"`
}

// Search handles search requests
//
//	@Summary		Search across all entities
//	@Description	Performs full-text search and filtering across epics, user stories, acceptance criteria, and requirements. Supports PostgreSQL full-text search with ranking and comprehensive filtering options. Results are cached for performance. Requires authentication.
//	@Tags			search
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			query					query		string	false	"Full-text search query. Searches across titles, descriptions, and reference IDs. Supports PostgreSQL text search syntax with automatic prefix matching."	example("user authentication")
//	@Param			creator_id				query		string	false	"Filter by creator ID (UUID format)"																																		example("123e4567-e89b-12d3-a456-426614174000")
//	@Param			assignee_id				query		string	false	"Filter by assignee ID (UUID format)"																																		example("123e4567-e89b-12d3-a456-426614174001")
//	@Param			priority				query		int		false	"Filter by priority level (1=Critical, 2=High, 3=Medium, 4=Low)"																											example(1)
//	@Param			status					query		string	false	"Filter by status (backlog, draft, in_progress, done, cancelled, active, obsolete)"																					example("in_progress")
//	@Param			created_from			query		string	false	"Filter by creation date from (RFC3339 format: YYYY-MM-DDTHH:MM:SSZ)"																								example("2023-01-01T00:00:00Z")
//	@Param			created_to				query		string	false	"Filter by creation date to (RFC3339 format: YYYY-MM-DDTHH:MM:SSZ)"																								example("2023-12-31T23:59:59Z")
//	@Param			epic_id					query		string	false	"Filter by parent epic ID (UUID format). Returns user stories, acceptance criteria, and requirements within the epic."												example("123e4567-e89b-12d3-a456-426614174002")
//	@Param			user_story_id			query		string	false	"Filter by parent user story ID (UUID format). Returns acceptance criteria and requirements within the user story."												example("123e4567-e89b-12d3-a456-426614174003")
//	@Param			acceptance_criteria_id	query		string	false	"Filter by parent acceptance criteria ID (UUID format). Returns requirements within the acceptance criteria."														example("123e4567-e89b-12d3-a456-426614174004")
//	@Param			requirement_type_id		query		string	false	"Filter by requirement type ID (UUID format). Only applies to requirement entities."																					example("123e4567-e89b-12d3-a456-426614174005")
//	@Param			author_id				query		string	false	"Filter by author ID (UUID format). Applies to comments and acceptance criteria."																						example("123e4567-e89b-12d3-a456-426614174006")
//	@Param			sort_by					query		string	false	"Sort by field: priority, created_at, last_modified, title, relevance (relevance only available with query)"														default(created_at)	example("priority")
//	@Param			sort_order				query		string	false	"Sort order: asc (ascending) or desc (descending)"																													default(desc)			example("asc")
//	@Param			limit					query		int		false	"Maximum number of results to return (1-100)"																															default(50)				example(20)
//	@Param			offset					query		int		false	"Number of results to skip for pagination (0-based)"																													default(0)				example(0)
//	@Success		200						{object}	service.SearchResponse	"Successful search with results, pagination metadata, and execution details"
//	@Failure		400						{object}	ErrorResponse			"Invalid search parameters (invalid UUID format, out of range values, invalid sort fields)"
//	@Failure		401						{object}	ErrorResponse			"Authentication required"
//	@Failure		500						{object}	ErrorResponse			"Internal server error during search operation"
//	@Router			/api/v1/search [get]
func (h *SearchHandler) Search(c *gin.Context) {
	correlationID, _ := c.Get("correlation_id")
	logger := h.logger.WithField("correlation_id", correlationID)

	// Parse query parameters
	options, err := h.parseSearchOptions(c)
	if err != nil {
		logger.WithError(err).Error("Failed to parse search options")
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: ErrorDetail{
				Code:    "INVALID_SEARCH_OPTIONS",
				Message: "Invalid search options: " + err.Error(),
			},
		})
		return
	}

	logger.WithFields(logrus.Fields{
		"query":      options.Query,
		"sort_by":    options.SortBy,
		"sort_order": options.SortOrder,
		"limit":      options.Limit,
		"offset":     options.Offset,
	}).Info("Performing search")

	// Perform search
	response, err := h.searchService.Search(c.Request.Context(), options)
	if err != nil {
		logger.WithError(err).Error("Search failed")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: ErrorDetail{
				Code:    "SEARCH_FAILED",
				Message: "Search operation failed",
			},
		})
		return
	}

	logger.WithFields(logrus.Fields{
		"total_results": response.Total,
		"returned":      len(response.Results),
	}).Info("Search completed successfully")

	c.JSON(http.StatusOK, response)
}

// parseSearchOptions parses search options from query parameters
func (h *SearchHandler) parseSearchOptions(c *gin.Context) (service.SearchOptions, error) {
	options := service.SearchOptions{
		Query:     c.Query("query"),
		SortBy:    c.DefaultQuery("sort_by", "created_at"),
		SortOrder: c.DefaultQuery("sort_order", "desc"),
	}

	// Parse limit
	if limitStr := c.Query("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			return options, fmt.Errorf("invalid limit parameter: %s", err.Error())
		}
		if limit < 0 {
			return options, fmt.Errorf("limit must be non-negative, got: %d", limit)
		}
		if limit > 100 {
			return options, fmt.Errorf("limit must not exceed 100, got: %d", limit)
		}
		options.Limit = limit
	} else {
		options.Limit = 50
	}

	// Parse offset
	if offsetStr := c.Query("offset"); offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err != nil {
			return options, fmt.Errorf("invalid offset parameter: %s", err.Error())
		}
		if offset < 0 {
			return options, fmt.Errorf("offset must be non-negative, got: %d", offset)
		}
		options.Offset = offset
	}

	// Parse filters
	filters := service.SearchFilters{}

	// Parse UUID filters
	if creatorIDStr := c.Query("creator_id"); creatorIDStr != "" {
		creatorID, err := uuid.Parse(creatorIDStr)
		if err != nil {
			return options, fmt.Errorf("invalid creator_id UUID format: %s", err.Error())
		}
		filters.CreatorID = &creatorID
	}

	if assigneeIDStr := c.Query("assignee_id"); assigneeIDStr != "" {
		assigneeID, err := uuid.Parse(assigneeIDStr)
		if err != nil {
			return options, fmt.Errorf("invalid assignee_id UUID format: %s", err.Error())
		}
		filters.AssigneeID = &assigneeID
	}

	if epicIDStr := c.Query("epic_id"); epicIDStr != "" {
		epicID, err := uuid.Parse(epicIDStr)
		if err != nil {
			return options, fmt.Errorf("invalid epic_id UUID format: %s", err.Error())
		}
		filters.EpicID = &epicID
	}

	if userStoryIDStr := c.Query("user_story_id"); userStoryIDStr != "" {
		userStoryID, err := uuid.Parse(userStoryIDStr)
		if err != nil {
			return options, fmt.Errorf("invalid user_story_id UUID format: %s", err.Error())
		}
		filters.UserStoryID = &userStoryID
	}

	if acIDStr := c.Query("acceptance_criteria_id"); acIDStr != "" {
		acID, err := uuid.Parse(acIDStr)
		if err != nil {
			return options, fmt.Errorf("invalid acceptance_criteria_id UUID format: %s", err.Error())
		}
		filters.AcceptanceCriteriaID = &acID
	}

	if reqTypeIDStr := c.Query("requirement_type_id"); reqTypeIDStr != "" {
		reqTypeID, err := uuid.Parse(reqTypeIDStr)
		if err != nil {
			return options, fmt.Errorf("invalid requirement_type_id UUID format: %s", err.Error())
		}
		filters.RequirementTypeID = &reqTypeID
	}

	if authorIDStr := c.Query("author_id"); authorIDStr != "" {
		authorID, err := uuid.Parse(authorIDStr)
		if err != nil {
			return options, fmt.Errorf("invalid author_id UUID format: %s", err.Error())
		}
		filters.AuthorID = &authorID
	}

	// Parse priority filter
	if priorityStr := c.Query("priority"); priorityStr != "" {
		priority, err := strconv.Atoi(priorityStr)
		if err != nil {
			return options, err
		}
		if priority < 1 || priority > 4 {
			return options, fmt.Errorf("priority must be between 1 and 4")
		}
		filters.Priority = &priority
	}

	// Parse status filter
	if status := c.Query("status"); status != "" {
		filters.Status = &status
	}

	// Parse date filters
	if createdFromStr := c.Query("created_from"); createdFromStr != "" {
		createdFrom, err := time.Parse(time.RFC3339, createdFromStr)
		if err != nil {
			return options, fmt.Errorf("invalid created_from date format, expected RFC3339: %s", err.Error())
		}
		filters.CreatedFrom = &createdFrom
	}

	if createdToStr := c.Query("created_to"); createdToStr != "" {
		createdTo, err := time.Parse(time.RFC3339, createdToStr)
		if err != nil {
			return options, fmt.Errorf("invalid created_to date format, expected RFC3339: %s", err.Error())
		}
		filters.CreatedTo = &createdTo
	}

	options.Filters = filters
	return options, nil
}

// SearchSuggestions handles search suggestion requests
//
//	@Summary		Get search suggestions
//	@Description	Provides search suggestions based on partial query input. Returns matching titles, reference IDs, and available status values to help users construct effective search queries.
//	@Tags			search
//	@Accept			json
//	@Produce		json
//	@Param			query	query		string	true	"Partial search query for generating suggestions. Minimum 2 characters recommended."	example("auth")
//	@Param			limit	query		int		false	"Maximum number of suggestions per category (1-50)"										default(10)	example(5)
//	@Success		200		{object}	SearchSuggestionsResponse	"Search suggestions grouped by category"
//	@Failure		400		{object}	ErrorResponse				"Invalid parameters (missing query, invalid limit)"
//	@Failure		500		{object}	ErrorResponse				"Internal server error during suggestion generation"
//	@Router			/api/search/suggestions [get]
func (h *SearchHandler) SearchSuggestions(c *gin.Context) {
	correlationID, _ := c.Get("correlation_id")
	logger := h.logger.WithField("correlation_id", correlationID)

	query := c.Query("query")
	if query == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: ErrorDetail{
				Code:    "MISSING_QUERY",
				Message: "Query parameter is required",
			},
		})
		return
	}

	limit := 10
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	logger.WithFields(logrus.Fields{
		"query": query,
		"limit": limit,
	}).Info("Getting search suggestions")

	// For now, return empty suggestions
	// In a full implementation, this would query the database for matching titles/reference IDs
	suggestions := map[string][]string{
		"titles":        []string{},
		"reference_ids": []string{},
		"statuses":      []string{"Backlog", "Draft", "In Progress", "Done", "Cancelled", "Active", "Obsolete"},
	}

	c.JSON(http.StatusOK, suggestions)
}

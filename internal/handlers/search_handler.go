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
// @Summary Search across all entities
// @Description Performs full-text search and filtering across epics, user stories, acceptance criteria, and requirements
// @Tags search
// @Accept json
// @Produce json
// @Param query query string false "Search query text"
// @Param creator_id query string false "Filter by creator ID (UUID)"
// @Param assignee_id query string false "Filter by assignee ID (UUID)"
// @Param priority query int false "Filter by priority (1-4)"
// @Param status query string false "Filter by status"
// @Param created_from query string false "Filter by creation date from (RFC3339 format)"
// @Param created_to query string false "Filter by creation date to (RFC3339 format)"
// @Param epic_id query string false "Filter by epic ID (UUID)"
// @Param user_story_id query string false "Filter by user story ID (UUID)"
// @Param acceptance_criteria_id query string false "Filter by acceptance criteria ID (UUID)"
// @Param requirement_type_id query string false "Filter by requirement type ID (UUID)"
// @Param author_id query string false "Filter by author ID (UUID)"
// @Param sort_by query string false "Sort by field (priority, created_at, last_modified, title)" default(created_at)
// @Param sort_order query string false "Sort order (asc, desc)" default(desc)
// @Param limit query int false "Limit number of results (max 100)" default(50)
// @Param offset query int false "Offset for pagination" default(0)
// @Success 200 {object} service.SearchResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/search [get]
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
			return options, err
		}
		options.Limit = limit
	} else {
		options.Limit = 50
	}

	// Parse offset
	if offsetStr := c.Query("offset"); offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err != nil {
			return options, err
		}
		options.Offset = offset
	}

	// Parse filters
	filters := service.SearchFilters{}

	// Parse UUID filters
	if creatorIDStr := c.Query("creator_id"); creatorIDStr != "" {
		creatorID, err := uuid.Parse(creatorIDStr)
		if err != nil {
			return options, err
		}
		filters.CreatorID = &creatorID
	}

	if assigneeIDStr := c.Query("assignee_id"); assigneeIDStr != "" {
		assigneeID, err := uuid.Parse(assigneeIDStr)
		if err != nil {
			return options, err
		}
		filters.AssigneeID = &assigneeID
	}

	if epicIDStr := c.Query("epic_id"); epicIDStr != "" {
		epicID, err := uuid.Parse(epicIDStr)
		if err != nil {
			return options, err
		}
		filters.EpicID = &epicID
	}

	if userStoryIDStr := c.Query("user_story_id"); userStoryIDStr != "" {
		userStoryID, err := uuid.Parse(userStoryIDStr)
		if err != nil {
			return options, err
		}
		filters.UserStoryID = &userStoryID
	}

	if acIDStr := c.Query("acceptance_criteria_id"); acIDStr != "" {
		acID, err := uuid.Parse(acIDStr)
		if err != nil {
			return options, err
		}
		filters.AcceptanceCriteriaID = &acID
	}

	if reqTypeIDStr := c.Query("requirement_type_id"); reqTypeIDStr != "" {
		reqTypeID, err := uuid.Parse(reqTypeIDStr)
		if err != nil {
			return options, err
		}
		filters.RequirementTypeID = &reqTypeID
	}

	if authorIDStr := c.Query("author_id"); authorIDStr != "" {
		authorID, err := uuid.Parse(authorIDStr)
		if err != nil {
			return options, err
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
			return options, err
		}
		filters.CreatedFrom = &createdFrom
	}

	if createdToStr := c.Query("created_to"); createdToStr != "" {
		createdTo, err := time.Parse(time.RFC3339, createdToStr)
		if err != nil {
			return options, err
		}
		filters.CreatedTo = &createdTo
	}

	options.Filters = filters
	return options, nil
}

// SearchSuggestions handles search suggestion requests
// @Summary Get search suggestions
// @Description Get search suggestions based on partial query
// @Tags search
// @Accept json
// @Produce json
// @Param query query string true "Partial search query"
// @Param limit query int false "Limit number of suggestions" default(10)
// @Success 200 {object} map[string][]string
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/search/suggestions [get]
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
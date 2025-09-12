package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"product-requirements-management/internal/models"
	"product-requirements-management/internal/service"
)

// NavigationHandler handles HTTP requests for hierarchical navigation
type NavigationHandler struct {
	navigationService service.NavigationService
}

// NewNavigationHandler creates a new navigation handler instance
func NewNavigationHandler(navigationService service.NavigationService) *NavigationHandler {
	return &NavigationHandler{
		navigationService: navigationService,
	}
}

// GetHierarchy handles GET /api/v1/hierarchy
func (h *NavigationHandler) GetHierarchy(c *gin.Context) {
	var filters service.HierarchyFilters

	// Parse query parameters for filtering
	if creatorID := c.Query("creator_id"); creatorID != "" {
		if id, err := uuid.Parse(creatorID); err == nil {
			filters.CreatorID = &id
		}
	}

	if assigneeID := c.Query("assignee_id"); assigneeID != "" {
		if id, err := uuid.Parse(assigneeID); err == nil {
			filters.AssigneeID = &id
		}
	}

	if status := c.Query("status"); status != "" {
		filters.Status = &status
	}

	if priority := c.Query("priority"); priority != "" {
		if p, err := strconv.Atoi(priority); err == nil && p >= 1 && p <= 4 {
			prio := models.Priority(p)
			filters.Priority = &prio
		}
	}

	// Parse sorting parameters
	if orderBy := c.Query("order_by"); orderBy != "" {
		filters.OrderBy = orderBy
	}

	if orderDir := c.Query("order_dir"); orderDir != "" {
		filters.OrderDirection = orderDir
	}

	// Parse pagination parameters
	if limit := c.Query("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil && l > 0 {
			filters.Limit = l
		}
	}

	if offset := c.Query("offset"); offset != "" {
		if o, err := strconv.Atoi(offset); err == nil && o >= 0 {
			filters.Offset = o
		}
	}

	// Parse expansion parameters
	if expand := c.Query("expand"); expand != "" {
		filters.Expand = expand
	}

	hierarchy, err := h.navigationService.GetHierarchy(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get hierarchy",
		})
		return
	}

	c.JSON(http.StatusOK, hierarchy)
}

// GetEpicHierarchy handles GET /api/v1/hierarchy/epics/:id
func (h *NavigationHandler) GetEpicHierarchy(c *gin.Context) {
	idParam := c.Param("id")

	// Try to parse as UUID first, then as reference ID
	var epicID uuid.UUID
	var err error

	if id, parseErr := uuid.Parse(idParam); parseErr == nil {
		epicID = id
	} else {
		// Need to resolve reference ID to UUID
		epic, resolveErr := h.navigationService.GetEpicByReferenceID(idParam)
		if resolveErr != nil {
			if errors.Is(resolveErr, service.ErrEpicNotFound) {
				c.JSON(http.StatusNotFound, gin.H{
					"error": "Epic not found",
				})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Failed to resolve epic reference ID",
				})
			}
			return
		}
		epicID = epic.ID
	}

	// Parse expansion parameters
	expand := c.Query("expand")
	if expand == "" {
		expand = "user_stories,requirements" // Default expansion
	}

	// Parse sorting parameters
	orderBy := c.Query("order_by")
	orderDirection := c.Query("order_dir")

	epicHierarchy, err := h.navigationService.GetEpicHierarchy(epicID, expand, orderBy, orderDirection)
	if err != nil {
		if errors.Is(err, service.ErrEpicNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Epic not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to get epic hierarchy",
			})
		}
		return
	}

	c.JSON(http.StatusOK, epicHierarchy)
}

// GetUserStoryHierarchy handles GET /api/v1/hierarchy/user-stories/:id
func (h *NavigationHandler) GetUserStoryHierarchy(c *gin.Context) {
	idParam := c.Param("id")

	// Try to parse as UUID first, then as reference ID
	var userStoryID uuid.UUID
	var err error

	if id, parseErr := uuid.Parse(idParam); parseErr == nil {
		userStoryID = id
	} else {
		// Need to resolve reference ID to UUID
		userStory, resolveErr := h.navigationService.GetUserStoryByReferenceID(idParam)
		if resolveErr != nil {
			if errors.Is(resolveErr, service.ErrUserStoryNotFound) {
				c.JSON(http.StatusNotFound, gin.H{
					"error": "User story not found",
				})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Failed to resolve user story reference ID",
				})
			}
			return
		}
		userStoryID = userStory.ID
	}

	// Parse expansion parameters
	expand := c.Query("expand")
	if expand == "" {
		expand = "requirements,acceptance_criteria" // Default expansion
	}

	// Parse sorting parameters
	orderBy := c.Query("order_by")
	orderDirection := c.Query("order_dir")

	userStoryHierarchy, err := h.navigationService.GetUserStoryHierarchy(userStoryID, expand, orderBy, orderDirection)
	if err != nil {
		if errors.Is(err, service.ErrUserStoryNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "User story not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to get user story hierarchy",
			})
		}
		return
	}

	c.JSON(http.StatusOK, userStoryHierarchy)
}

// GetEntityPath handles GET /api/v1/hierarchy/path/:entity_type/:id
func (h *NavigationHandler) GetEntityPath(c *gin.Context) {
	entityType := c.Param("entity_type")
	idParam := c.Param("id")

	// Validate entity type
	if !isValidEntityType(entityType) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid entity type",
		})
		return
	}

	// Try to parse as UUID first, then as reference ID
	var entityID uuid.UUID
	var err error

	if id, parseErr := uuid.Parse(idParam); parseErr == nil {
		entityID = id
	} else {
		// Need to resolve reference ID to UUID based on entity type
		entityID, err = h.navigationService.ResolveReferenceID(entityType, idParam)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Entity not found",
			})
			return
		}
	}

	path, err := h.navigationService.GetEntityPath(entityType, entityID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get entity path",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"path": path,
	})
}

// isValidEntityType checks if the entity type is valid for navigation
func isValidEntityType(entityType string) bool {
	validTypes := []string{"epic", "user_story", "acceptance_criteria", "requirement"}
	for _, validType := range validTypes {
		if entityType == validType {
			return true
		}
	}
	return false
}

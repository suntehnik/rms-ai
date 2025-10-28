package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ListResponse represents a standardized paginated response format
// All list endpoints should return data in this format for consistency
// @Description Standardized paginated response format used by all list endpoints
// @Description This ensures consistent pagination handling across the entire API
type ListResponse[T any] struct {
	// The actual data items returned by the endpoint
	Data []T `json:"data" swaggertype:"array"`
	// Total number of items available (not just in this page)
	TotalCount int64 `json:"total_count" example:"150" minimum:"0"`
	// Number of items per page (maximum items returned in this response)
	Limit int `json:"limit" example:"50" minimum:"1" maximum:"100"`
	// Number of items skipped from the beginning (for pagination)
	Offset int `json:"offset" example:"0" minimum:"0"`
}

// PaginationParams represents standard pagination parameters
// @Description Standard pagination parameters used across all list endpoints
type PaginationParams struct {
	// Maximum number of items to return per page
	Limit int `json:"limit" form:"limit" binding:"min=1,max=100" example:"50" minimum:"1" maximum:"100"`
	// Number of items to skip from the beginning
	Offset int `json:"offset" form:"offset" binding:"min=0" example:"0" minimum:"0"`
}

// SetDefaults sets default values for pagination parameters
func (p *PaginationParams) SetDefaults() {
	if p.Limit == 0 {
		p.Limit = 50 // Default page size
	}
	if p.Offset < 0 {
		p.Offset = 0
	}
}

// SendListResponse sends a standardized list response
// Helper function to send consistent paginated responses across all list endpoints
func SendListResponse[T any](c *gin.Context, data []T, totalCount int64, limit, offset int) {
	response := ListResponse[T]{
		Data:       data,
		TotalCount: totalCount,
		Limit:      limit,
		Offset:     offset,
	}
	c.JSON(http.StatusOK, response)
}

// Note: ErrorResponse and ErrorDetail types are defined in deletion_handler.go
// We'll use those existing types for consistency

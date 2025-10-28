package handlers

import (
	"net/http"
	"product-requirements-management/internal/models"
	"product-requirements-management/internal/service"

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

// Concrete types for Swagger documentation
// These types are used only for Swagger documentation since swag doesn't support generics

// AcceptanceCriteriaListResponse represents a paginated list of acceptance criteria
// @Description Standardized paginated response for acceptance criteria
type AcceptanceCriteriaListResponse struct {
	// The acceptance criteria items returned by the endpoint
	Data []models.AcceptanceCriteria `json:"data"`
	// Total number of acceptance criteria available (not just in this page)
	TotalCount int64 `json:"total_count" example:"150" minimum:"0"`
	// Number of items per page (maximum items returned in this response)
	Limit int `json:"limit" example:"50" minimum:"1" maximum:"100"`
	// Number of items skipped from the beginning (for pagination)
	Offset int `json:"offset" example:"0" minimum:"0"`
}

// RequirementRelationshipListResponse represents a paginated list of requirement relationships
// @Description Standardized paginated response for requirement relationships
type RequirementRelationshipListResponse struct {
	// The requirement relationship items returned by the endpoint
	Data []models.RequirementRelationship `json:"data"`
	// Total number of relationships available (not just in this page)
	TotalCount int64 `json:"total_count" example:"150" minimum:"0"`
	// Number of items per page (maximum items returned in this response)
	Limit int `json:"limit" example:"50" minimum:"1" maximum:"100"`
	// Number of items skipped from the beginning (for pagination)
	Offset int `json:"offset" example:"0" minimum:"0"`
}

// RequirementListResponse represents a paginated list of requirements
// @Description Standardized paginated response for requirements
type RequirementListResponse struct {
	// The requirement items returned by the endpoint
	Data []models.Requirement `json:"data"`
	// Total number of requirements available (not just in this page)
	TotalCount int64 `json:"total_count" example:"150" minimum:"0"`
	// Number of items per page (maximum items returned in this response)
	Limit int `json:"limit" example:"50" minimum:"1" maximum:"100"`
	// Number of items skipped from the beginning (for pagination)
	Offset int `json:"offset" example:"0" minimum:"0"`
}

// SteeringDocumentListResponse represents a paginated list of steering documents
// @Description Standardized paginated response for steering documents
type SteeringDocumentListResponse struct {
	// The steering document items returned by the endpoint
	Data []models.SteeringDocument `json:"data"`
	// Total number of steering documents available (not just in this page)
	TotalCount int64 `json:"total_count" example:"150" minimum:"0"`
	// Number of items per page (maximum items returned in this response)
	Limit int `json:"limit" example:"50" minimum:"1" maximum:"100"`
	// Number of items skipped from the beginning (for pagination)
	Offset int `json:"offset" example:"0" minimum:"0"`
}

// CommentListResponse represents a paginated list of comments
// @Description Standardized paginated response for comments
type CommentListResponse struct {
	// The comment items returned by the endpoint
	Data []service.CommentResponse `json:"data"`
	// Total number of comments available (not just in this page)
	TotalCount int64 `json:"total_count" example:"150" minimum:"0"`
	// Number of items per page (maximum items returned in this response)
	Limit int `json:"limit" example:"50" minimum:"1" maximum:"100"`
	// Number of items skipped from the beginning (for pagination)
	Offset int `json:"offset" example:"0" minimum:"0"`
}

// PersonalAccessTokenListResponse represents a paginated list of personal access tokens
// @Description Standardized paginated response for personal access tokens
type PersonalAccessTokenListResponse struct {
	// The personal access token items returned by the endpoint
	Data []models.PersonalAccessToken `json:"data"`
	// Total number of tokens available (not just in this page)
	TotalCount int64 `json:"total_count" example:"150" minimum:"0"`
	// Number of items per page (maximum items returned in this response)
	Limit int `json:"limit" example:"50" minimum:"1" maximum:"100"`
	// Number of items skipped from the beginning (for pagination)
	Offset int `json:"offset" example:"0" minimum:"0"`
}

// RequirementTypeListResponse represents a paginated list of requirement types
// @Description Standardized paginated response for requirement types
type RequirementTypeListResponse struct {
	// The requirement type items returned by the endpoint
	Data []models.RequirementType `json:"data"`
	// Total number of requirement types available (not just in this page)
	TotalCount int64 `json:"total_count" example:"150" minimum:"0"`
	// Number of items per page (maximum items returned in this response)
	Limit int `json:"limit" example:"50" minimum:"1" maximum:"100"`
	// Number of items skipped from the beginning (for pagination)
	Offset int `json:"offset" example:"0" minimum:"0"`
}

// RelationshipTypeListResponse represents a paginated list of relationship types
// @Description Standardized paginated response for relationship types
type RelationshipTypeListResponse struct {
	// The relationship type items returned by the endpoint
	Data []models.RelationshipType `json:"data"`
	// Total number of relationship types available (not just in this page)
	TotalCount int64 `json:"total_count" example:"150" minimum:"0"`
	// Number of items per page (maximum items returned in this response)
	Limit int `json:"limit" example:"50" minimum:"1" maximum:"100"`
	// Number of items skipped from the beginning (for pagination)
	Offset int `json:"offset" example:"0" minimum:"0"`
}

// StatusModelListResponse represents a paginated list of status models
// @Description Standardized paginated response for status models
type StatusModelListResponse struct {
	// The status model items returned by the endpoint
	Data []models.StatusModel `json:"data"`
	// Total number of status models available (not just in this page)
	TotalCount int64 `json:"total_count" example:"150" minimum:"0"`
	// Number of items per page (maximum items returned in this response)
	Limit int `json:"limit" example:"50" minimum:"1" maximum:"100"`
	// Number of items skipped from the beginning (for pagination)
	Offset int `json:"offset" example:"0" minimum:"0"`
}

// StatusListResponse represents a paginated list of statuses
// @Description Standardized paginated response for statuses
type StatusListResponse struct {
	// The status items returned by the endpoint
	Data []models.Status `json:"data"`
	// Total number of statuses available (not just in this page)
	TotalCount int64 `json:"total_count" example:"150" minimum:"0"`
	// Number of items per page (maximum items returned in this response)
	Limit int `json:"limit" example:"50" minimum:"1" maximum:"100"`
	// Number of items skipped from the beginning (for pagination)
	Offset int `json:"offset" example:"0" minimum:"0"`
}

// StatusTransitionListResponse represents a paginated list of status transitions
// @Description Standardized paginated response for status transitions
type StatusTransitionListResponse struct {
	// The status transition items returned by the endpoint
	Data []models.StatusTransition `json:"data"`
	// Total number of status transitions available (not just in this page)
	TotalCount int64 `json:"total_count" example:"150" minimum:"0"`
	// Number of items per page (maximum items returned in this response)
	Limit int `json:"limit" example:"50" minimum:"1" maximum:"100"`
	// Number of items skipped from the beginning (for pagination)
	Offset int `json:"offset" example:"0" minimum:"0"`
}

// Note: ErrorResponse and ErrorDetail types are defined in deletion_handler.go
// We'll use those existing types for consistency

package docs

import (
	"time"

	"github.com/google/uuid"
)

// APIResponse represents a standard API response wrapper
// @Description Standard API response wrapper for all endpoints
type APIResponse struct {
	Data    any          `json:"data,omitempty" example:"{}"`
	Message string       `json:"message,omitempty" example:"Operation completed successfully"`
	Error   *ErrorDetail `json:"error,omitempty"`
} // @name APIResponse

// ErrorResponse represents a standard error response
// @Description Standard error response structure
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
} // @name ErrorResponse

// ErrorDetail represents detailed error information
// @Description Detailed error information with code and message
type ErrorDetail struct {
	Code    string `json:"code" example:"VALIDATION_ERROR"`
	Message string `json:"message" example:"Invalid input provided"`
} // @name ErrorDetail

// ValidationErrorResponse represents validation error details
// @Description Validation error response with field-specific errors
type ValidationErrorResponse struct {
	Error *ValidationErrorDetail `json:"error"`
} // @name ValidationErrorResponse

// ValidationErrorDetail represents detailed validation error information
// @Description Detailed validation error with field-specific messages
type ValidationErrorDetail struct {
	Code    string            `json:"code" example:"VALIDATION_ERROR"`
	Message string            `json:"message" example:"Validation failed"`
	Fields  map[string]string `json:"fields,omitempty" example:"{\"title\":\"Title is required\",\"priority\":\"Priority must be between 1 and 4\"}"`
} // @name ValidationErrorDetail

// PaginationMeta represents pagination metadata for list responses
// @Description Pagination metadata for paginated responses
type PaginationMeta struct {
	Limit  int `json:"limit" example:"50"`
	Offset int `json:"offset" example:"0"`
	Total  int `json:"total" example:"150"`
	Count  int `json:"count" example:"50"`
} // @name PaginationMeta

// PaginatedResponse represents a paginated list response
// @Description Standard paginated response wrapper
type PaginatedResponse struct {
	Data       any             `json:"data"`
	Pagination *PaginationMeta `json:"pagination"`
	Message    string          `json:"message,omitempty" example:"Data retrieved successfully"`
} // @name PaginatedResponse

// SuccessResponse represents a simple success response
// @Description Simple success response for operations that don't return data
type SuccessResponse struct {
	Success bool   `json:"success" example:"true"`
	Message string `json:"message" example:"Operation completed successfully"`
} // @name SuccessResponse

// IDResponse represents a response containing a newly created entity ID
// @Description Response containing the ID of a newly created entity
type IDResponse struct {
	ID      uuid.UUID `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Message string    `json:"message" example:"Entity created successfully"`
} // @name IDResponse

// StatusResponse represents a response for status change operations
// @Description Response for status change operations
type StatusResponse struct {
	ID        uuid.UUID `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Status    string    `json:"status" example:"in_progress"`
	Message   string    `json:"message" example:"Status updated successfully"`
	UpdatedAt time.Time `json:"updated_at" example:"2023-01-01T12:00:00Z"`
} // @name StatusResponse

// AssignmentResponse represents a response for assignment operations
// @Description Response for assignment operations
type AssignmentResponse struct {
	ID         uuid.UUID  `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	AssigneeID *uuid.UUID `json:"assignee_id,omitempty" example:"456e7890-e89b-12d3-a456-426614174001"`
	Message    string     `json:"message" example:"Assignment updated successfully"`
	UpdatedAt  time.Time  `json:"updated_at" example:"2023-01-01T12:00:00Z"`
} // @name AssignmentResponse

// HealthCheckResponse represents a health check response
// @Description Health check response
type HealthCheckResponse struct {
	Status  string `json:"status" example:"healthy"`
	Service string `json:"service" example:"product-requirements-management"`
	Version string `json:"version,omitempty" example:"1.0.0"`
} // @name HealthCheckResponse

// DetailedHealthCheckResponse represents a detailed health check response
// @Description Detailed health check response with dependency status
type DetailedHealthCheckResponse struct {
	Status  string                 `json:"status" example:"healthy"`
	Service string                 `json:"service" example:"product-requirements-management"`
	Message string                 `json:"message,omitempty" example:"All systems operational"`
	Checks  map[string]HealthCheck `json:"checks,omitempty"`
} // @name DetailedHealthCheckResponse

// HealthCheck represents individual health check status
// @Description Individual health check status for a dependency
type HealthCheck struct {
	Status  string `json:"status" example:"healthy"`
	Message string `json:"message,omitempty" example:"Connection successful"`
} // @name HealthCheck

// Note: SearchResponse is defined in internal/service/search_service.go
// This is a documentation reference that aligns with the existing service structure
// SearchResponseDoc represents a search response for Swagger documentation
// @Description Search response with results and metadata from search service
type SearchResponseDoc struct {
	Results    []any     `json:"results" example:"[{\"id\":\"123e4567-e89b-12d3-a456-426614174000\",\"type\":\"epic\",\"title\":\"User Authentication\"}]"`
	Total      int64     `json:"total" example:"25"`
	Limit      int       `json:"limit" example:"50"`
	Offset     int       `json:"offset" example:"0"`
	Query      string    `json:"query" example:"user authentication"`
	ExecutedAt time.Time `json:"executed_at" example:"2023-01-01T12:00:00Z"`
} // @name SearchResponse

// BulkOperationResponse represents a response for bulk operations
// @Description Response for bulk operations with success and failure counts
type BulkOperationResponse struct {
	TotalProcessed int         `json:"total_processed" example:"100"`
	Successful     int         `json:"successful" example:"95"`
	Failed         int         `json:"failed" example:"5"`
	Errors         []BulkError `json:"errors,omitempty"`
	Message        string      `json:"message" example:"Bulk operation completed"`
} // @name BulkOperationResponse

// BulkError represents an error in bulk operations
// @Description Error information for failed items in bulk operations
type BulkError struct {
	Index int    `json:"index" example:"3"`
	ID    string `json:"id,omitempty" example:"123e4567-e89b-12d3-a456-426614174000"`
	Error string `json:"error" example:"Validation failed: title is required"`
	Code  string `json:"code" example:"VALIDATION_ERROR"`
} // @name BulkError

// Common HTTP Status Code Examples for Swagger Documentation

// BadRequestError represents a 400 Bad Request error
// @Description Bad Request - Invalid input or malformed request
type BadRequestError ErrorResponse

// UnauthorizedError represents a 401 Unauthorized error
// @Description Unauthorized - Authentication required or invalid
type UnauthorizedError ErrorResponse

// ForbiddenError represents a 403 Forbidden error
// @Description Forbidden - Insufficient permissions
type ForbiddenError ErrorResponse

// NotFoundError represents a 404 Not Found error
// @Description Not Found - Requested resource does not exist
type NotFoundError ErrorResponse

// ConflictError represents a 409 Conflict error
// @Description Conflict - Resource conflict or business rule violation
type ConflictError ErrorResponse

// UnprocessableEntityError represents a 422 Unprocessable Entity error
// @Description Unprocessable Entity - Validation failed
type UnprocessableEntityError ValidationErrorResponse

// InternalServerError represents a 500 Internal Server Error
// @Description Internal Server Error - Unexpected server error
type InternalServerError ErrorResponse

// ServiceUnavailableError represents a 503 Service Unavailable error
// @Description Service Unavailable - Service temporarily unavailable
type ServiceUnavailableError ErrorResponse

// UserStoryRelationshipExamples provides examples of hierarchical data retrieval patterns
// @Description Examples of complex relationship queries for user stories

// UserStoryWithAcceptanceCriteriaExample represents a user story with acceptance criteria
// @Description Example response for GET /api/v1/user-stories/{id}/acceptance-criteria showing hierarchical data
type UserStoryWithAcceptanceCriteriaExample struct {
	ID                 uuid.UUID `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	ReferenceID        string    `json:"reference_id" example:"US-001"`
	Title              string    `json:"title" example:"User Login with Email and Password"`
	AcceptanceCriteria []struct {
		ID          uuid.UUID `json:"id" example:"456e7890-e89b-12d3-a456-426614174001"`
		ReferenceID string    `json:"reference_id" example:"AC-001"`
		Title       string    `json:"title" example:"Valid email format validation"`
		Description string    `json:"description" example:"WHEN user enters email THEN system SHALL validate email format"`
	} `json:"acceptance_criteria"`
} // @name UserStoryWithAcceptanceCriteriaExample

// UserStoryWithRequirementsExample represents a user story with requirements
// @Description Example response for GET /api/v1/user-stories/{id}/requirements showing hierarchical data
type UserStoryWithRequirementsExample struct {
	ID           uuid.UUID `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	ReferenceID  string    `json:"reference_id" example:"US-001"`
	Title        string    `json:"title" example:"User Login with Email and Password"`
	Requirements []struct {
		ID          uuid.UUID `json:"id" example:"789e0123-e89b-12d3-a456-426614174002"`
		ReferenceID string    `json:"reference_id" example:"REQ-001"`
		Title       string    `json:"title" example:"Email validation service integration"`
		Description string    `json:"description" example:"The system must integrate with email validation service to verify email format and domain validity"`
		Priority    int       `json:"priority" example:"2"`
		Status      string    `json:"status" example:"Draft"`
	} `json:"requirements"`
} // @name UserStoryWithRequirementsExample

// UserStoryListResponse represents a paginated list of user stories
// @Description Example response for GET /api/v1/user-stories with filtering and pagination
type UserStoryListResponse struct {
	UserStories []struct {
		ID          uuid.UUID `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
		ReferenceID string    `json:"reference_id" example:"US-001"`
		EpicID      uuid.UUID `json:"epic_id" example:"456e7890-e89b-12d3-a456-426614174001"`
		Title       string    `json:"title" example:"User Login with Email and Password"`
		Status      string    `json:"status" example:"Backlog"`
		Priority    int       `json:"priority" example:"2"`
		CreatedAt   time.Time `json:"created_at" example:"2023-01-15T10:30:00Z"`
	} `json:"user_stories"`
	Count int `json:"count" example:"25"`
} // @name UserStoryListResponse

// NestedUserStoryCreationExample represents nested creation within epic context
// @Description Example request for POST /api/v1/epics/{id}/user-stories showing nested resource creation
type NestedUserStoryCreationExample struct {
	CreatorID   uuid.UUID `json:"creator_id" example:"123e4567-e89b-12d3-a456-426614174001"`
	AssigneeID  uuid.UUID `json:"assignee_id,omitempty" example:"123e4567-e89b-12d3-a456-426614174002"`
	Priority    int       `json:"priority" example:"2"`
	Title       string    `json:"title" example:"User Profile Management"`
	Description string    `json:"description" example:"As a registered user, I want to manage my profile information, so that I can keep my account details up to date."`
} // @name NestedUserStoryCreationExample

// RequirementRelationshipExamples provides examples of requirement relationship operations
// @Description Examples of requirement relationship creation and management

// RequirementWithRelationshipsExample represents a requirement with all its relationships
// @Description Example response for GET /api/v1/requirements/{id}/relationships showing complete relationship view
type RequirementWithRelationshipsExample struct {
	ID                  uuid.UUID `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	ReferenceID         string    `json:"reference_id" example:"REQ-001"`
	Title               string    `json:"title" example:"User authentication validation"`
	Description         string    `json:"description" example:"The system must validate user credentials against the database"`
	Status              string    `json:"status" example:"draft"`
	Priority            int       `json:"priority" example:"2"`
	SourceRelationships []struct {
		ID                  uuid.UUID `json:"id" example:"456e7890-e89b-12d3-a456-426614174001"`
		TargetRequirementID uuid.UUID `json:"target_requirement_id" example:"789e0123-e89b-12d3-a456-426614174002"`
		RelationshipType    string    `json:"relationship_type" example:"depends_on"`
		TargetRequirement   struct {
			ID          uuid.UUID `json:"id" example:"789e0123-e89b-12d3-a456-426614174002"`
			ReferenceID string    `json:"reference_id" example:"REQ-002"`
			Title       string    `json:"title" example:"Database connection setup"`
		} `json:"target_requirement"`
	} `json:"source_relationships"`
	TargetRelationships []struct {
		ID                  uuid.UUID `json:"id" example:"abc1234d-e89b-12d3-a456-426614174003"`
		SourceRequirementID uuid.UUID `json:"source_requirement_id" example:"def5678e-e89b-12d3-a456-426614174004"`
		RelationshipType    string    `json:"relationship_type" example:"blocks"`
		SourceRequirement   struct {
			ID          uuid.UUID `json:"id" example:"def5678e-e89b-12d3-a456-426614174004"`
			ReferenceID string    `json:"reference_id" example:"REQ-003"`
			Title       string    `json:"title" example:"User interface design"`
		} `json:"source_requirement"`
	} `json:"target_relationships"`
} // @name RequirementWithRelationshipsExample

// RequirementListResponse represents a paginated list of requirements
// @Description Example response for GET /api/v1/requirements with filtering and pagination
type RequirementListResponse struct {
	Requirements []struct {
		ID                   uuid.UUID  `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
		ReferenceID          string     `json:"reference_id" example:"REQ-001"`
		UserStoryID          uuid.UUID  `json:"user_story_id" example:"456e7890-e89b-12d3-a456-426614174001"`
		AcceptanceCriteriaID *uuid.UUID `json:"acceptance_criteria_id,omitempty" example:"789e0123-e89b-12d3-a456-426614174002"`
		Title                string     `json:"title" example:"User authentication validation"`
		Description          string     `json:"description" example:"The system must validate user credentials against the database"`
		Status               string     `json:"status" example:"draft"`
		Priority             int        `json:"priority" example:"2"`
		CreatedAt            time.Time  `json:"created_at" example:"2023-01-15T10:30:00Z"`
	} `json:"requirements"`
	Count int `json:"count" example:"15"`
} // @name RequirementListResponse

// RequirementRelationshipCreationExample represents relationship creation request
// @Description Example request for POST /api/v1/requirements/relationships showing relationship creation
type RequirementRelationshipCreationExample struct {
	SourceRequirementID uuid.UUID `json:"source_requirement_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	TargetRequirementID uuid.UUID `json:"target_requirement_id" example:"456e7890-e89b-12d3-a456-426614174001"`
	RelationshipTypeID  uuid.UUID `json:"relationship_type_id" example:"789e0123-e89b-12d3-a456-426614174002"`
	CreatedBy           uuid.UUID `json:"created_by" example:"abc1234d-e89b-12d3-a456-426614174003"`
} // @name RequirementRelationshipCreationExample

// RequirementStatusChangeExample represents status change request
// @Description Example request for PATCH /api/v1/requirements/{id}/status showing status transition
type RequirementStatusChangeExample struct {
	Status string `json:"status" example:"in_review"`
} // @name RequirementStatusChangeExample

// RequirementAssignmentExample represents assignment request
// @Description Example request for PATCH /api/v1/requirements/{id}/assign showing assignment operation
type RequirementAssignmentExample struct {
	AssigneeID uuid.UUID `json:"assignee_id" example:"123e4567-e89b-12d3-a456-426614174001"`
} // @name RequirementAssignmentExample

// AcceptanceCriteriaExamples provides examples of acceptance criteria operations
// @Description Examples of acceptance criteria creation and management

// AcceptanceCriteriaListResponse represents a paginated list of acceptance criteria
// @Description Example response for GET /api/v1/acceptance-criteria with filtering and pagination
type AcceptanceCriteriaListResponse struct {
	AcceptanceCriteria []struct {
		ID          uuid.UUID `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
		ReferenceID string    `json:"reference_id" example:"AC-001"`
		UserStoryID uuid.UUID `json:"user_story_id" example:"456e7890-e89b-12d3-a456-426614174001"`
		AuthorID    uuid.UUID `json:"author_id" example:"789e0123-e89b-12d3-a456-426614174002"`
		Description string    `json:"description" example:"WHEN user enters valid email and password THEN system SHALL authenticate user and redirect to dashboard"`
		CreatedAt   time.Time `json:"created_at" example:"2023-01-15T10:30:00Z"`
	} `json:"acceptance_criteria"`
	Count int `json:"count" example:"8"`
} // @name AcceptanceCriteriaListResponse

// AcceptanceCriteriaCreationExample represents acceptance criteria creation request
// @Description Example request for POST /api/v1/user-stories/{id}/acceptance-criteria showing nested creation
type AcceptanceCriteriaCreationExample struct {
	AuthorID    uuid.UUID `json:"author_id" example:"123e4567-e89b-12d3-a456-426614174001"`
	Description string    `json:"description" example:"WHEN user enters valid email and password THEN system SHALL authenticate user and redirect to dashboard"`
} // @name AcceptanceCriteriaCreationExample

// AcceptanceCriteriaUpdateExample represents acceptance criteria update request
// @Description Example request for PUT /api/v1/acceptance-criteria/{id} showing update operation
type AcceptanceCriteriaUpdateExample struct {
	Description string `json:"description" example:"WHEN user enters valid email and password THEN system SHALL authenticate user, log the session, and redirect to dashboard"`
} // @name AcceptanceCriteriaUpdateExample

// RequirementSearchResponse represents search results for requirements
// @Description Example response for GET /api/v1/requirements/search showing full-text search results
type RequirementSearchResponse struct {
	Requirements []struct {
		ID          uuid.UUID `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
		ReferenceID string    `json:"reference_id" example:"REQ-001"`
		Title       string    `json:"title" example:"User authentication validation"`
		Description string    `json:"description" example:"The system must validate user credentials against the database"`
		Status      string    `json:"status" example:"draft"`
		Priority    int       `json:"priority" example:"2"`
		Relevance   float64   `json:"relevance,omitempty" example:"0.85"`
	} `json:"requirements"`
	Count int    `json:"count" example:"5"`
	Query string `json:"query" example:"authentication validation"`
} // @name RequirementSearchResponse

// CircularDependencyPreventionExample represents circular dependency error
// @Description Example error response when attempting to create circular relationships
type CircularDependencyPreventionExample struct {
	Error string `json:"error" example:"Cannot create relationship between the same requirement"`
	Code  string `json:"code" example:"CIRCULAR_RELATIONSHIP"`
} // @name CircularDependencyPreventionExample

// DuplicateRelationshipExample represents duplicate relationship error
// @Description Example error response when attempting to create duplicate relationships
type DuplicateRelationshipExample struct {
	Error string `json:"error" example:"Relationship already exists"`
	Code  string `json:"code" example:"DUPLICATE_RELATIONSHIP"`
} // @name DuplicateRelationshipExample

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

// SearchSuggestionsResponse represents search suggestions grouped by category
// @Description Search suggestions response with categorized suggestions for query completion
type SearchSuggestionsResponse struct {
	Titles       []string `json:"titles" example:"User Authentication System,User Profile Management,Authentication Service"`
	ReferenceIDs []string `json:"reference_ids" example:"EP-001,US-042,REQ-123,AC-089"`
	Statuses     []string `json:"statuses" example:"Backlog,Draft,In Progress,Done,Cancelled,Active,Obsolete"`
} // @name SearchSuggestionsResponse

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
// Comment System Documentation Schemas

// CommentCreationRequest represents a request to create a comment
// @Description Request structure for creating comments (both general and inline)
type CommentCreationRequest struct {
	AuthorID          uuid.UUID  `json:"author_id" binding:"required" example:"123e4567-e89b-12d3-a456-426614174000"`
	Content           string     `json:"content" binding:"required" example:"This requirement needs clarification on the authentication flow."`
	ParentCommentID   *uuid.UUID `json:"parent_comment_id,omitempty" example:"456e7890-e89b-12d3-a456-426614174001"`
	LinkedText        *string    `json:"linked_text,omitempty" example:"OAuth 2.0 authentication flow"`
	TextPositionStart *int       `json:"text_position_start,omitempty" example:"45"`
	TextPositionEnd   *int       `json:"text_position_end,omitempty" example:"73"`
} // @name CommentCreationRequest

// CommentUpdateRequest represents a request to update a comment
// @Description Request structure for updating comment content
type CommentUpdateRequest struct {
	Content string `json:"content" binding:"required" example:"Updated comment content with additional clarification."`
} // @name CommentUpdateRequest

// CommentResponse represents a comment in API responses
// @Description Complete comment information including author details and thread context
type CommentResponse struct {
	ID              uuid.UUID  `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	EntityType      string     `json:"entity_type" example:"epic"`
	EntityID        uuid.UUID  `json:"entity_id" example:"456e7890-e89b-12d3-a456-426614174001"`
	ParentCommentID *uuid.UUID `json:"parent_comment_id,omitempty" example:"789e0123-e89b-12d3-a456-426614174002"`
	AuthorID        uuid.UUID  `json:"author_id" example:"abc1234d-e89b-12d3-a456-426614174003"`
	Author          *struct {
		ID       uuid.UUID `json:"id" example:"abc1234d-e89b-12d3-a456-426614174003"`
		Username string    `json:"username" example:"john.doe"`
		Email    string    `json:"email" example:"john.doe@example.com"`
	} `json:"author,omitempty"`
	CreatedAt         string            `json:"created_at" example:"2023-01-01T00:00:00Z"`
	UpdatedAt         string            `json:"updated_at" example:"2023-01-02T12:30:00Z"`
	Content           string            `json:"content" example:"This requirement needs clarification on the authentication flow."`
	IsResolved        bool              `json:"is_resolved" example:"false"`
	LinkedText        *string           `json:"linked_text,omitempty" example:"OAuth 2.0 authentication flow"`
	TextPositionStart *int              `json:"text_position_start,omitempty" example:"45"`
	TextPositionEnd   *int              `json:"text_position_end,omitempty" example:"73"`
	Replies           []CommentResponse `json:"replies,omitempty"`
	IsInline          bool              `json:"is_inline" example:"true"`
	IsReply           bool              `json:"is_reply" example:"false"`
	Depth             int               `json:"depth" example:"0"`
} // @name CommentResponse

// CommentListResponse represents a list of comments with metadata
// @Description Response structure for comment list endpoints
type CommentListResponse struct {
	Comments []CommentResponse `json:"comments"`
	Count    int               `json:"count" example:"5"`
} // @name CommentListResponse

// CommentsByStatusResponse represents comments filtered by status
// @Description Response structure for comments filtered by resolution status
type CommentsByStatusResponse struct {
	Comments []CommentResponse `json:"comments"`
	Count    int               `json:"count" example:"3"`
	Status   string            `json:"status" example:"unresolved"`
} // @name CommentsByStatusResponse

// InlineCommentCreationExample represents an inline comment creation request
// @Description Example request for creating an inline comment with text position data
type InlineCommentCreationExample struct {
	AuthorID          uuid.UUID `json:"author_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Content           string    `json:"content" example:"This authentication method needs to specify which OAuth 2.0 flow to use."`
	LinkedText        string    `json:"linked_text" example:"OAuth 2.0 authentication flow"`
	TextPositionStart int       `json:"text_position_start" example:"45"`
	TextPositionEnd   int       `json:"text_position_end" example:"73"`
} // @name InlineCommentCreationExample

// ThreadedCommentExample represents a threaded comment structure
// @Description Example of threaded comments showing parent-child relationships
type ThreadedCommentExample struct {
	ID      uuid.UUID `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Content string    `json:"content" example:"This requirement needs clarification."`
	IsReply bool      `json:"is_reply" example:"false"`
	Depth   int       `json:"depth" example:"0"`
	Replies []struct {
		ID      uuid.UUID `json:"id" example:"456e7890-e89b-12d3-a456-426614174001"`
		Content string    `json:"content" example:"I agree, specifically about the authentication flow."`
		IsReply bool      `json:"is_reply" example:"true"`
		Depth   int       `json:"depth" example:"1"`
		Replies []struct {
			ID      uuid.UUID `json:"id" example:"789e0123-e89b-12d3-a456-426614174002"`
			Content string    `json:"content" example:"Let me provide more details on this."`
			IsReply bool      `json:"is_reply" example:"true"`
			Depth   int       `json:"depth" example:"2"`
		} `json:"replies"`
	} `json:"replies"`
} // @name ThreadedCommentExample

// InlineCommentValidationRequest represents a request to validate inline comments
// @Description Request structure for validating inline comments after text changes
type InlineCommentValidationRequest struct {
	NewDescription string `json:"new_description" binding:"required" example:"Updated entity description with modified text content that may affect inline comment positions."`
} // @name InlineCommentValidationRequest

// CommentResolutionWorkflowExample represents comment resolution workflow
// @Description Example showing comment resolution and unresolve operations
type CommentResolutionWorkflowExample struct {
	ID         uuid.UUID `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Content    string    `json:"content" example:"This issue has been addressed in the latest revision."`
	IsResolved bool      `json:"is_resolved" example:"true"`
	UpdatedAt  string    `json:"updated_at" example:"2023-01-02T15:45:00Z"`
} // @name CommentResolutionWorkflowExample

// InlineCommentPositionExample represents inline comment text positioning
// @Description Example showing how inline comments are positioned within text
type InlineCommentPositionExample struct {
	OriginalText      string `json:"original_text" example:"The system must implement OAuth 2.0 authentication flow for secure user login."`
	LinkedText        string `json:"linked_text" example:"OAuth 2.0 authentication flow"`
	TextPositionStart int    `json:"text_position_start" example:"32"`
	TextPositionEnd   int    `json:"text_position_end" example:"60"`
	CommentContent    string `json:"comment_content" example:"Which specific OAuth 2.0 flow should be implemented? Authorization Code, Implicit, or Client Credentials?"`
} // @name InlineCommentPositionExample

// Configuration Management Response Types

// RequirementTypeListResponse represents a list of requirement types
// @Description Response structure for requirement type list endpoints
type RequirementTypeListResponse struct {
	RequirementTypes []struct {
		ID          uuid.UUID `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
		Name        string    `json:"name" example:"Functional Requirement"`
		Description *string   `json:"description,omitempty" example:"Requirements that specify what the system should do"`
		CreatedAt   time.Time `json:"created_at" example:"2023-01-01T00:00:00Z"`
		UpdatedAt   time.Time `json:"updated_at" example:"2023-01-01T00:00:00Z"`
	} `json:"requirement_types"`
	Count int `json:"count" example:"5"`
} // @name RequirementTypeListResponse

// RelationshipTypeListResponse represents a list of relationship types
// @Description Response structure for relationship type list endpoints
type RelationshipTypeListResponse struct {
	RelationshipTypes []struct {
		ID          uuid.UUID `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
		Name        string    `json:"name" example:"depends_on"`
		Description *string   `json:"description,omitempty" example:"Indicates that one requirement depends on another"`
		CreatedAt   time.Time `json:"created_at" example:"2023-01-01T00:00:00Z"`
		UpdatedAt   time.Time `json:"updated_at" example:"2023-01-01T00:00:00Z"`
	} `json:"relationship_types"`
	Count int `json:"count" example:"8"`
} // @name RelationshipTypeListResponse

// StatusModelListResponse represents a list of status models
// @Description Response structure for status model list endpoints
type StatusModelListResponse struct {
	StatusModels []struct {
		ID          uuid.UUID `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
		EntityType  string    `json:"entity_type" example:"epic"`
		Name        string    `json:"name" example:"Default Epic Workflow"`
		Description *string   `json:"description,omitempty" example:"Standard workflow for epic management"`
		IsDefault   bool      `json:"is_default" example:"true"`
		CreatedAt   time.Time `json:"created_at" example:"2023-01-01T00:00:00Z"`
		UpdatedAt   time.Time `json:"updated_at" example:"2023-01-01T00:00:00Z"`
	} `json:"status_models"`
	Count int `json:"count" example:"4"`
} // @name StatusModelListResponse

// StatusListResponse represents a list of statuses
// @Description Response structure for status list endpoints
type StatusListResponse struct {
	Statuses []struct {
		ID            uuid.UUID `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
		StatusModelID uuid.UUID `json:"status_model_id" example:"456e7890-e89b-12d3-a456-426614174001"`
		Name          string    `json:"name" example:"Backlog"`
		Description   *string   `json:"description,omitempty" example:"Items waiting to be started"`
		Color         *string   `json:"color,omitempty" example:"#6B7280"`
		IsInitial     bool      `json:"is_initial" example:"true"`
		IsFinal       bool      `json:"is_final" example:"false"`
		Order         int       `json:"order" example:"1"`
		CreatedAt     time.Time `json:"created_at" example:"2023-01-01T00:00:00Z"`
		UpdatedAt     time.Time `json:"updated_at" example:"2023-01-01T00:00:00Z"`
	} `json:"statuses"`
	Count int `json:"count" example:"6"`
} // @name StatusListResponse

// StatusTransitionListResponse represents a list of status transitions
// @Description Response structure for status transition list endpoints
type StatusTransitionListResponse struct {
	Transitions []struct {
		ID            uuid.UUID `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
		StatusModelID uuid.UUID `json:"status_model_id" example:"456e7890-e89b-12d3-a456-426614174001"`
		FromStatusID  uuid.UUID `json:"from_status_id" example:"789e0123-e89b-12d3-a456-426614174002"`
		ToStatusID    uuid.UUID `json:"to_status_id" example:"abc1234d-e89b-12d3-a456-426614174003"`
		Name          *string   `json:"name,omitempty" example:"Start Work"`
		Description   *string   `json:"description,omitempty" example:"Transition from backlog to in progress"`
		FromStatus    struct {
			ID   uuid.UUID `json:"id" example:"789e0123-e89b-12d3-a456-426614174002"`
			Name string    `json:"name" example:"Backlog"`
		} `json:"from_status"`
		ToStatus struct {
			ID   uuid.UUID `json:"id" example:"abc1234d-e89b-12d3-a456-426614174003"`
			Name string    `json:"name" example:"In Progress"`
		} `json:"to_status"`
		CreatedAt time.Time `json:"created_at" example:"2023-01-01T00:00:00Z"`
		UpdatedAt time.Time `json:"updated_at" example:"2023-01-01T00:00:00Z"`
	} `json:"transitions"`
	Count int `json:"count" example:"12"`
} // @name StatusTransitionListResponse

// Configuration Examples and Workflows

// StatusWorkflowExample represents a complete status workflow configuration
// @Description Example showing a complete status workflow with statuses and transitions
type StatusWorkflowExample struct {
	StatusModel struct {
		ID          uuid.UUID `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
		EntityType  string    `json:"entity_type" example:"epic"`
		Name        string    `json:"name" example:"Epic Workflow"`
		Description string    `json:"description" example:"Standard workflow for epic lifecycle management"`
		IsDefault   bool      `json:"is_default" example:"true"`
	} `json:"status_model"`
	Statuses []struct {
		ID        uuid.UUID `json:"id" example:"456e7890-e89b-12d3-a456-426614174001"`
		Name      string    `json:"name" example:"Backlog"`
		IsInitial bool      `json:"is_initial" example:"true"`
		IsFinal   bool      `json:"is_final" example:"false"`
		Order     int       `json:"order" example:"1"`
	} `json:"statuses"`
	Transitions []struct {
		ID           uuid.UUID `json:"id" example:"789e0123-e89b-12d3-a456-426614174002"`
		FromStatusID uuid.UUID `json:"from_status_id" example:"456e7890-e89b-12d3-a456-426614174001"`
		ToStatusID   uuid.UUID `json:"to_status_id" example:"abc1234d-e89b-12d3-a456-426614174003"`
		Name         string    `json:"name" example:"Start Epic"`
	} `json:"transitions"`
} // @name StatusWorkflowExample

// RequirementTypeUsageExample represents requirement type usage patterns
// @Description Example showing how requirement types are used to categorize requirements
type RequirementTypeUsageExample struct {
	RequirementType struct {
		ID          uuid.UUID `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
		Name        string    `json:"name" example:"Security Requirement"`
		Description string    `json:"description" example:"Requirements related to system security and data protection"`
	} `json:"requirement_type"`
	UsageCount int `json:"usage_count" example:"15"`
	Examples   []struct {
		ID          uuid.UUID `json:"id" example:"456e7890-e89b-12d3-a456-426614174001"`
		ReferenceID string    `json:"reference_id" example:"REQ-042"`
		Title       string    `json:"title" example:"User authentication must use OAuth 2.0"`
	} `json:"examples"`
} // @name RequirementTypeUsageExample

// RelationshipTypeUsageExample represents relationship type usage patterns
// @Description Example showing how relationship types define requirement dependencies
type RelationshipTypeUsageExample struct {
	RelationshipType struct {
		ID          uuid.UUID `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
		Name        string    `json:"name" example:"depends_on"`
		Description string    `json:"description" example:"Source requirement depends on target requirement"`
	} `json:"relationship_type"`
	UsageCount int `json:"usage_count" example:"8"`
	Examples   []struct {
		ID                  uuid.UUID `json:"id" example:"456e7890-e89b-12d3-a456-426614174001"`
		SourceRequirementID uuid.UUID `json:"source_requirement_id" example:"789e0123-e89b-12d3-a456-426614174002"`
		TargetRequirementID uuid.UUID `json:"target_requirement_id" example:"abc1234d-e89b-12d3-a456-426614174003"`
		SourceTitle         string    `json:"source_title" example:"User login validation"`
		TargetTitle         string    `json:"target_title" example:"Database connection setup"`
	} `json:"examples"`
} // @name RelationshipTypeUsageExample

// Authentication and Authorization Documentation Schemas

// JWTTokenExample represents a JWT token structure for documentation
// @Description JWT token structure and format used for authentication
type JWTTokenExample struct {
	Token     string    `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMTIzZTQ1NjctZTg5Yi0xMmQzLWE0NTYtNDI2NjE0MTc0MDAwIiwidXNlcm5hbWUiOiJqb2huX2RvZSIsInJvbGUiOiJVc2VyIiwiZXhwIjoxNjQwOTk1MjAwLCJpYXQiOjE2NDA5OTE2MDAsIm5iZiI6MTY0MDk5MTYwMH0.signature"`
	TokenType string    `json:"token_type" example:"Bearer"`
	ExpiresIn int       `json:"expires_in" example:"3600"`
	ExpiresAt time.Time `json:"expires_at" example:"2023-01-02T12:30:00Z"`
} // @name JWTTokenExample

// JWTClaimsExample represents JWT token claims structure
// @Description JWT token claims containing user information and permissions
type JWTClaimsExample struct {
	UserID   string `json:"user_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Username string `json:"username" example:"john_doe"`
	Role     string `json:"role" example:"User"`
	Exp      int64  `json:"exp" example:"1640995200"`
	Iat      int64  `json:"iat" example:"1640991600"`
	Nbf      int64  `json:"nbf" example:"1640991600"`
} // @name JWTClaimsExample

// AuthenticationErrorResponse represents authentication-related errors
// @Description Standard error responses for authentication failures
type AuthenticationErrorResponse struct {
	Error string `json:"error" example:"Authentication required"`
	Code  string `json:"code" example:"AUTHENTICATION_REQUIRED"`
} // @name AuthenticationErrorResponse

// AuthorizationErrorResponse represents authorization-related errors
// @Description Standard error responses for authorization failures
type AuthorizationErrorResponse struct {
	Error string `json:"error" example:"Insufficient permissions"`
	Code  string `json:"code" example:"INSUFFICIENT_PERMISSIONS"`
} // @name AuthorizationErrorResponse

// TokenExpiredErrorResponse represents expired token errors
// @Description Error response when JWT token has expired
type TokenExpiredErrorResponse struct {
	Error string `json:"error" example:"Token expired"`
	Code  string `json:"code" example:"TOKEN_EXPIRED"`
} // @name TokenExpiredErrorResponse

// InvalidTokenErrorResponse represents invalid token errors
// @Description Error response when JWT token is invalid or malformed
type InvalidTokenErrorResponse struct {
	Error string `json:"error" example:"Invalid token"`
	Code  string `json:"code" example:"INVALID_TOKEN"`
} // @name InvalidTokenErrorResponse

// RoleBasedAccessExample represents role-based access control patterns
// @Description Examples of role-based access control for different user roles
type RoleBasedAccessExample struct {
	Administrator struct {
		Description string   `json:"description" example:"Full system access including user and configuration management"`
		Permissions []string `json:"permissions" example:"[\"create_users\",\"manage_config\",\"delete_entities\",\"view_all\",\"edit_all\"]"`
	} `json:"administrator"`
	User struct {
		Description string   `json:"description" example:"Can create, edit, and delete entities but cannot manage users or system configuration"`
		Permissions []string `json:"permissions" example:"[\"create_entities\",\"edit_entities\",\"delete_entities\",\"view_all\",\"comment\"]"`
	} `json:"user"`
	Commenter struct {
		Description string   `json:"description" example:"Can only add comments and view entities, limited editing capabilities"`
		Permissions []string `json:"permissions" example:"[\"view_all\",\"comment\",\"limited_edit\"]"`
	} `json:"commenter"`
} // @name RoleBasedAccessExample

// SecurityBestPracticesExample represents security best practices documentation
// @Description Security best practices for API usage and token management
type SecurityBestPracticesExample struct {
	TokenManagement struct {
		StorageRecommendation string `json:"storage_recommendation" example:"Store JWT tokens securely in httpOnly cookies or secure storage"`
		ExpirationHandling    string `json:"expiration_handling" example:"Implement automatic token refresh before expiration"`
		TransmissionSecurity  string `json:"transmission_security" example:"Always use HTTPS for token transmission"`
	} `json:"token_management"`
	AuthenticationFlow struct {
		LoginProcess    string `json:"login_process" example:"POST /auth/login with username/password to receive JWT token"`
		TokenUsage      string `json:"token_usage" example:"Include 'Authorization: Bearer <token>' header in all authenticated requests"`
		LogoutProcess   string `json:"logout_process" example:"Client-side token removal (server-side blacklisting optional)"`
		RefreshStrategy string `json:"refresh_strategy" example:"Implement token refresh mechanism before expiration"`
	} `json:"authentication_flow"`
	ErrorHandling struct {
		UnauthorizedResponse string `json:"unauthorized_response" example:"Return 401 for missing or invalid authentication"`
		ForbiddenResponse    string `json:"forbidden_response" example:"Return 403 for insufficient permissions"`
		TokenExpiredResponse string `json:"token_expired_response" example:"Return 401 with specific error code for expired tokens"`
	} `json:"error_handling"`
} // @name SecurityBestPracticesExample

// AuthenticationHeaderExample represents authentication header format
// @Description Examples of proper authentication header formatting
type AuthenticationHeaderExample struct {
	HeaderName   string `json:"header_name" example:"Authorization"`
	HeaderFormat string `json:"header_format" example:"Bearer <jwt_token>"`
	Example      string `json:"example" example:"Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	Description  string `json:"description" example:"Include the word 'Bearer' followed by a space and the JWT token"`
} // @name AuthenticationHeaderExample

// PermissionMatrixExample represents permission matrix for different operations
// @Description Permission matrix showing which roles can perform which operations
type PermissionMatrixExample struct {
	Operations map[string]struct {
		Administrator bool   `json:"administrator"`
		User          bool   `json:"user"`
		Commenter     bool   `json:"commenter"`
		Description   string `json:"description"`
	} `json:"operations"`
} // @name PermissionMatrixExample

// SecuritySchemeDocumentation represents the security scheme documentation
// @Description Complete security scheme documentation for the API
type SecuritySchemeDocumentation struct {
	Type        string `json:"type" example:"apiKey"`
	In          string `json:"in" example:"header"`
	Name        string `json:"name" example:"Authorization"`
	Description string `json:"description" example:"JWT token authentication. Format: 'Bearer {token}'"`
	TokenFormat string `json:"token_format" example:"JWT"`
	Algorithm   string `json:"algorithm" example:"HS256"`
	ExpiresIn   string `json:"expires_in" example:"1 hour"`
} // @name SecuritySchemeDocumentation

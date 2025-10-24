# Design Document

## Overview

This design document outlines the architectural changes needed to address API inconsistencies in the Product Requirements Management System. The solution focuses on standardizing response formats, improving handler naming conventions, ensuring proper entity preloading, and creating consistent error messaging across all API endpoints.

## Architecture

### Current Architecture Issues

The current architecture has several inconsistencies:

1. **Response Format Inconsistency**: Different handlers return different response structures for list operations
2. **Handler Naming Inconsistency**: Multiple handlers exist for similar operations (e.g., comment creation)
3. **Missing Entity Preloading**: To-one relationships are not consistently preloaded
4. **Error Message Inconsistency**: Different handlers use different error message formats

### Target Architecture

The target architecture will implement:

1. **Standardized Response Layer**: All list endpoints return consistent paginated responses
2. **Unified Handler Pattern**: Single handlers for similar operations with parameterized routing
3. **Default Entity Preloading**: Automatic preloading of to-one relationships
4. **Centralized Error Handling**: Standardized error message templates

## Components and Interfaces

### 1. Standardized Response Format

#### Standard List Response Structure
```go
type ListResponse struct {
    Data       interface{} `json:"data"`
    TotalCount int64       `json:"total_count"`
    Limit      int         `json:"limit"`
    Offset     int         `json:"offset"`
}
```

#### Pagination Parameters
```go
type PaginationParams struct {
    Limit  int `json:"limit" form:"limit" binding:"min=1,max=100"`
    Offset int `json:"offset" form:"offset" binding:"min=0"`
}

func (p *PaginationParams) SetDefaults() {
    if p.Limit == 0 {
        p.Limit = 50
    }
    if p.Offset < 0 {
        p.Offset = 0
    }
}
```

### 2. Handler Consolidation Pattern

#### Unified Comment Handler
```go
// Single handler for all comment creation
func CreateComment(c *gin.Context) {
    entityType := c.Param("entity_type") // epic, user_story, acceptance_criteria, requirement
    entityID := c.Param("entity_id")
    
    // Validate entity type and ID
    // Create comment using unified service method
}
```

#### Unified Resource Creation Pattern
```go
// Single handler for nested resource creation
func CreateAcceptanceCriteria(c *gin.Context) {
    userStoryID := c.Param("user_story_id") // from URL path
    // Handle creation with parent validation
}
```

### 3. Entity Preloading Strategy

#### Service Layer Preloading
```go
// Epic service with default preloading
func (s *EpicService) ListEpics(params ListParams) ([]models.Epic, int64, error) {
    return s.repository.ListEpics(params, []string{"Creator", "Assignee"})
}

// User Story service with default preloading
func (s *UserStoryService) ListUserStories(params ListParams) ([]models.UserStory, int64, error) {
    return s.repository.ListUserStories(params, []string{"Creator", "Assignee", "Epic"})
}
```

#### Repository Layer Preloading
```go
func (r *EpicRepository) ListEpics(params ListParams, preloads []string) ([]models.Epic, int64, error) {
    query := r.db.Model(&models.Epic{})
    
    // Apply preloads
    for _, preload := range preloads {
        query = query.Preload(preload)
    }
    
    // Apply pagination and return results with count
}
```

### 4. Centralized Error Handling

#### Error Message Templates
```go
// internal/service/errors.go
package service

import "fmt"

type ErrorMessages struct{}

func (e *ErrorMessages) NotFound(entityType string) string {
    return fmt.Sprintf("%s not found", entityType)
}

func (e *ErrorMessages) InvalidID(entityType string) string {
    return fmt.Sprintf("Invalid %s ID format", entityType)
}

func (e *ErrorMessages) DeletionConflict(entityType string) string {
    return fmt.Sprintf("Cannot delete %s due to dependencies. Use force=true to override.", entityType)
}

var Errors = &ErrorMessages{}
```

## Data Models

### Enhanced Repository Interface
```go
type Repository interface {
    // Standard list method with preloading support
    List(params ListParams, preloads []string) (interface{}, int64, error)
    
    // Standard get method with preloading support
    Get(id string, preloads []string) (interface{}, error)
}

type ListParams struct {
    Limit    int
    Offset   int
    Filters  map[string]interface{}
    OrderBy  string
}
```

### Response Helper Functions
```go
// internal/handlers/response.go
package handlers

func SendListResponse(c *gin.Context, data interface{}, totalCount int64, limit, offset int) {
    response := ListResponse{
        Data:       data,
        TotalCount: totalCount,
        Limit:      limit,
        Offset:     offset,
    }
    c.JSON(http.StatusOK, response)
}

func SendErrorResponse(c *gin.Context, statusCode int, message string) {
    c.JSON(statusCode, gin.H{"error": message})
}
```

## Error Handling

### Standardized Error Response Format
```go
type ErrorResponse struct {
    Error string `json:"error"`
}

type ValidationErrorResponse struct {
    Error            string            `json:"error"`
    ValidationErrors map[string]string `json:"validation_errors,omitempty"`
}
```

### Error Handling Middleware
```go
func ErrorHandlingMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Next()
        
        // Handle any errors that occurred during request processing
        if len(c.Errors) > 0 {
            err := c.Errors.Last()
            // Convert to standardized error response
        }
    }
}
```

## Testing Strategy

### Unit Testing Approach

1. **Handler Tests**: Test each handler with standardized response format
2. **Service Tests**: Test preloading functionality and pagination
3. **Repository Tests**: Test pagination implementation and preloading
4. **Error Message Tests**: Test standardized error message generation

### Integration Testing Approach

1. **API Response Format Tests**: Verify all list endpoints return consistent format
2. **Preloading Tests**: Verify all entities include expected relationships
3. **Error Consistency Tests**: Verify error messages follow standard format
4. **Pagination Tests**: Verify pagination works correctly across all endpoints

### Test Examples

#### Handler Test Example
```go
func TestListEpics_StandardResponseFormat(t *testing.T) {
    // Setup test data
    // Call handler
    // Assert response format matches ListResponse structure
    // Assert preloaded entities are included
}
```

#### Service Test Example
```go
func TestEpicService_ListEpics_PreloadsRelationships(t *testing.T) {
    // Setup test data with relationships
    // Call service method
    // Assert Creator and Assignee are preloaded
}
```

## Implementation Plan

### Phase 1: Response Format Standardization
1. Create `ListResponse` struct and helper functions
2. Update all list handlers to use standardized response format
3. Update corresponding service and repository methods for pagination

### Phase 2: Entity Preloading Implementation
1. Update service methods to include default preloads
2. Modify repository methods to support preloading parameters
3. Test all endpoints to ensure relationships are included

### Phase 3: Handler Consolidation
1. Create unified comment creation handler
2. Consolidate acceptance criteria and requirement creation handlers
3. Update routing to use consolidated handlers

### Phase 4: Error Message Standardization
1. Create centralized error message templates
2. Update all handlers to use standardized error messages
3. Implement error handling middleware

### Phase 5: Testing and Validation
1. Write comprehensive tests for all changes
2. Validate API consistency across all endpoints
3. Update API documentation to reflect changes

## Migration Considerations

### Backward Compatibility
- All changes maintain backward compatibility with existing API contracts
- Response format changes are additive (adding fields, not removing)
- Error message changes improve consistency without breaking functionality

### Performance Impact
- Preloading relationships may increase query complexity but reduces API calls
- Pagination implementation may require database query optimization
- Overall performance should improve due to reduced round trips

### Deployment Strategy
- Changes can be deployed incrementally by handler/endpoint
- No database schema changes required
- API documentation should be updated simultaneously with deployment
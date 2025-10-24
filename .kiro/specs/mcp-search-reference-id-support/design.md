# Design Document: MCP Search Reference ID Support

## Overview

This design document outlines the implementation of reference ID search support for the MCP (Model Context Protocol) search functionality. The enhancement will enable the search_global tool to handle reference ID queries (e.g., "US-119", "EP-006") and support both direct entity lookup and hierarchical child entity searches based on entity type filtering.

## Architecture

### Current Search Flow
```
MCP Agent → handleSearchGlobal → SearchService.Search → performFullTextSearch/performFilterSearch → Database Queries
```

### Enhanced Search Flow
```
MCP Agent → handleSearchGlobal → SearchService.Search → detectReferenceIDPattern → 
  ├─ Direct Reference ID Search (no entity_types filter)
  └─ Hierarchical Reference ID Search (with entity_types filter)
```

## Components and Interfaces

### 1. Reference ID Pattern Detection

**New Component: ReferenceIDDetector**
```go
type ReferenceIDDetector struct{}

type ReferenceIDPattern struct {
    IsReferenceID bool
    EntityType    string  // "epic", "user_story", "requirement", "acceptance_criteria", "steering_document"
    Number        string  // The numeric part (e.g., "119" from "US-119")
    OriginalQuery string
}

func (d *ReferenceIDDetector) DetectPattern(query string) ReferenceIDPattern
```

**Pattern Matching Rules:**
- Epic: `EP-\d+` (e.g., EP-006)
- User Story: `US-\d+` (e.g., US-119)  
- Requirement: `REQ-\d+` (e.g., REQ-045)
- Acceptance Criteria: `AC-\d+` (e.g., AC-023)
- Steering Document: `STD-\d+` (e.g., STD-012)

### 2. Enhanced Search Service

**Modified SearchService Interface:**
```go
type SearchServiceInterface interface {
    Search(ctx context.Context, options SearchOptions) (*SearchResponse, error)
    SearchByReferenceID(ctx context.Context, referenceID string, entityTypes []string) (*SearchResponse, error)
    InvalidateCache(ctx context.Context) error
}
```

**New SearchOptions Fields:**
```go
type SearchOptions struct {
    Query       string        `json:"query"`
    EntityTypes []string      `json:"entity_types,omitempty"` // New field
    Filters     SearchFilters `json:"filters"`
    SortBy      string        `json:"sort_by"`
    SortOrder   string        `json:"sort_order"`
    Limit       int           `json:"limit"`
    Offset      int           `json:"offset"`
}
```

### 3. Reference ID Search Logic

**Direct Reference ID Search (No Entity Types Filter):**
```go
func (s *SearchService) searchByDirectReferenceID(ctx context.Context, pattern ReferenceIDPattern) (*SearchResponse, error) {
    // Search for the exact entity with the matching reference ID
    // Return the single matching entity
}
```

**Hierarchical Reference ID Search (With Entity Types Filter):**
```go
func (s *SearchService) searchByHierarchicalReferenceID(ctx context.Context, pattern ReferenceIDPattern, entityTypes []string) (*SearchResponse, error) {
    // Find the parent entity by reference ID
    // Search for child entities of the specified types
    // Return all matching child entities
}
```

### 4. Database Query Optimization

**New Repository Methods:**
```go
// Epic Repository
func (r *EpicRepository) GetByReferenceID(referenceID string) (*models.Epic, error)
func (r *EpicRepository) GetUserStoriesByEpicReferenceID(referenceID string) ([]models.UserStory, error)

// User Story Repository  
func (r *UserStoryRepository) GetByReferenceID(referenceID string) (*models.UserStory, error)
func (r *UserStoryRepository) GetRequirementsByUserStoryReferenceID(referenceID string) ([]models.Requirement, error)
func (r *UserStoryRepository) GetAcceptanceCriteriaByUserStoryReferenceID(referenceID string) ([]models.AcceptanceCriteria, error)

// Steering Document Repository
func (r *SteeringDocumentRepository) GetByReferenceID(referenceID string) (*models.SteeringDocument, error)
func (r *SteeringDocumentRepository) GetEpicsBySteeringDocumentReferenceID(referenceID string) ([]models.Epic, error)
```

## Data Models

### Enhanced Search Result
```go
type SearchResult struct {
    ID          uuid.UUID `json:"id"`
    ReferenceID string    `json:"reference_id"`
    Type        string    `json:"type"`
    Title       string    `json:"title"`
    Description string    `json:"description,omitempty"`
    Priority    *int      `json:"priority,omitempty"`
    Status      string    `json:"status"`
    CreatedAt   time.Time `json:"created_at"`
    Relevance   float64   `json:"relevance,omitempty"`
    ParentID    *string   `json:"parent_id,omitempty"`    // New field for hierarchical searches
    ParentType  *string   `json:"parent_type,omitempty"`  // New field for hierarchical searches
}
```

### Search Context
```go
type SearchContext struct {
    IsReferenceIDSearch bool
    ReferenceIDPattern  ReferenceIDPattern
    EntityTypes         []string
    IsHierarchical      bool
}
```

## Error Handling

### Reference ID Not Found
```go
type ReferenceIDNotFoundError struct {
    ReferenceID string
    EntityType  string
}

func (e ReferenceIDNotFoundError) Error() string {
    return fmt.Sprintf("reference ID %s not found for entity type %s", e.ReferenceID, e.EntityType)
}
```

### Invalid Entity Type Filter
```go
type InvalidEntityTypeError struct {
    EntityType     string
    ParentType     string
    ValidTypes     []string
}

func (e InvalidEntityTypeError) Error() string {
    return fmt.Sprintf("entity type %s is not a valid child type for parent %s. Valid types: %v", 
        e.EntityType, e.ParentType, e.ValidTypes)
}
```

## Testing Strategy

### Unit Tests

**Reference ID Pattern Detection Tests:**
```go
func TestReferenceIDDetector_DetectPattern(t *testing.T) {
    tests := []struct {
        query    string
        expected ReferenceIDPattern
    }{
        {"US-119", ReferenceIDPattern{IsReferenceID: true, EntityType: "user_story", Number: "119"}},
        {"EP-006", ReferenceIDPattern{IsReferenceID: true, EntityType: "epic", Number: "006"}},
        {"us-119", ReferenceIDPattern{IsReferenceID: true, EntityType: "user_story", Number: "119"}},
        {"random text", ReferenceIDPattern{IsReferenceID: false}},
    }
}
```

**Direct Reference ID Search Tests:**
```go
func TestSearchService_SearchByDirectReferenceID(t *testing.T) {
    // Test direct lookup of US-119 returns only that user story
    // Test direct lookup of EP-006 returns only that epic
    // Test case-insensitive matching
}
```

**Hierarchical Reference ID Search Tests:**
```go
func TestSearchService_SearchByHierarchicalReferenceID(t *testing.T) {
    // Test EP-006 with entity_types ["user_story"] returns all user stories of that epic
    // Test US-119 with entity_types ["requirement"] returns all requirements of that user story
    // Test invalid combinations return appropriate errors
}
```

### Integration Tests

**MCP Handler Integration Tests:**
```go
func TestToolsHandler_HandleSearchGlobal_ReferenceID(t *testing.T) {
    // Test full MCP search flow with reference ID queries
    // Test entity type filtering with reference IDs
    // Test error handling for invalid reference IDs
}
```

### End-to-End Tests

**MCP Agent Simulation Tests:**
```go
func TestMCPAgent_ReferenceIDSearch(t *testing.T) {
    // Simulate actual MCP agent usage patterns
    // Test all documented usage patterns from requirements
}
```

## Implementation Plan

### Phase 1: Core Reference ID Detection
1. Implement `ReferenceIDDetector` component
2. Add reference ID pattern matching logic
3. Add unit tests for pattern detection

### Phase 2: Direct Reference ID Search
1. Implement direct reference ID lookup methods in repositories
2. Add `searchByDirectReferenceID` method to SearchService
3. Integrate with existing search flow
4. Add unit and integration tests

### Phase 3: Hierarchical Reference ID Search
1. Implement hierarchical lookup methods in repositories
2. Add `searchByHierarchicalReferenceID` method to SearchService
3. Add entity type validation logic
4. Add comprehensive tests

### Phase 4: MCP Handler Integration
1. Modify `handleSearchGlobal` to support entity_types parameter
2. Integrate reference ID search logic
3. Update error handling and response formatting
4. Add end-to-end tests

### Phase 5: Performance Optimization
1. Add database indexes for reference ID lookups
2. Implement caching for reference ID searches
3. Add performance benchmarks
4. Optimize query performance

## Performance Considerations

### Database Indexes
```sql
-- Ensure reference_id columns have indexes for fast lookups
CREATE INDEX CONCURRENTLY idx_epics_reference_id ON epics(reference_id);
CREATE INDEX CONCURRENTLY idx_user_stories_reference_id ON user_stories(reference_id);
CREATE INDEX CONCURRENTLY idx_requirements_reference_id ON requirements(reference_id);
CREATE INDEX CONCURRENTLY idx_acceptance_criteria_reference_id ON acceptance_criteria(reference_id);
CREATE INDEX CONCURRENTLY idx_steering_documents_reference_id ON steering_documents(reference_id);

-- Indexes for hierarchical lookups
CREATE INDEX CONCURRENTLY idx_user_stories_epic_id ON user_stories(epic_id);
CREATE INDEX CONCURRENTLY idx_requirements_user_story_id ON requirements(user_story_id);
CREATE INDEX CONCURRENTLY idx_acceptance_criteria_user_story_id ON acceptance_criteria(user_story_id);
```

### Caching Strategy
- Cache reference ID to UUID mappings for 10 minutes
- Cache hierarchical relationship queries for 5 minutes
- Invalidate cache when entities are created/updated/deleted

### Query Optimization
- Use direct reference ID lookups instead of full-text search when pattern is detected
- Batch hierarchical queries to minimize database round trips
- Implement query result pagination for large hierarchical results

## Security Considerations

### Access Control
- Maintain existing user authentication and authorization
- Ensure reference ID searches respect user permissions
- Validate entity type filters to prevent unauthorized access

### Input Validation
- Sanitize reference ID patterns to prevent injection attacks
- Validate entity type parameters against allowed values
- Implement rate limiting for reference ID searches

## Monitoring and Observability

### Metrics
- Track reference ID search usage patterns
- Monitor search performance for reference ID vs text searches
- Track cache hit rates for reference ID lookups

### Logging
- Log reference ID search patterns and results
- Log hierarchical search operations
- Log performance metrics for optimization

This design provides a comprehensive approach to implementing reference ID search support while maintaining performance, security, and maintainability of the existing search system.
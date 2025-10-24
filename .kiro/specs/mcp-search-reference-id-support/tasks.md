# Implementation Plan

- [x] 1. Implement reference ID pattern detection
  - Create ReferenceIDDetector component with pattern matching logic
  - Add support for all entity types (EP, US, REQ, AC, STD patterns)
  - Implement case-insensitive matching for reference IDs
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 3.1, 3.2, 3.3, 3.4, 3.5_

- [x] 1.1 Write unit tests for reference ID pattern detection
  - Test all reference ID patterns (EP-XXX, US-XXX, REQ-XXX, AC-XXX, STD-XXX)
  - Test case-insensitive matching
  - Test invalid patterns return false
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 3.1, 3.2, 3.3, 3.4, 3.5_

- [x] 2. Enhance SearchService with entity types support
  - Add EntityTypes field to SearchOptions struct
  - Modify Search method to handle entity type filtering
  - Add validation for entity type parameters
  - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5, 2.6_

- [ ]* 2.1 Write unit tests for SearchOptions validation
  - Test valid entity type combinations
  - Test invalid entity type parameters
  - Test empty entity types handling
  - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5, 2.6_

- [x] 3. Implement direct reference ID search functionality
  - Add searchByDirectReferenceID method to SearchService
  - Implement repository methods for reference ID lookups
  - Add case-insensitive reference ID matching in database queries
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 4.1, 4.2, 4.3_

- [x] 3.1 Add repository methods for direct reference ID lookups
  - Implement GetByReferenceID methods for all entity repositories
  - Add case-insensitive database queries using ILIKE
  - Handle not found cases appropriately
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5_

- [ ]* 3.2 Write unit tests for direct reference ID search
  - Test exact reference ID matches for all entity types
  - Test case-insensitive matching
  - Test not found scenarios
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 3.1, 3.2, 3.3, 3.4, 3.5_

- [ ] 4. Implement hierarchical reference ID search functionality
  - Add searchByHierarchicalReferenceID method to SearchService
  - Implement repository methods for hierarchical lookups
  - Add entity type validation for parent-child relationships
  - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5, 2.6_

- [ ] 4.1 Add repository methods for hierarchical lookups
  - Implement GetUserStoriesByEpicReferenceID method
  - Implement GetRequirementsByUserStoryReferenceID method
  - Implement GetAcceptanceCriteriaByUserStoryReferenceID method
  - Implement GetEpicsBySteeringDocumentReferenceID method
  - _Requirements: 2.2, 2.3, 2.4, 2.5, 2.6_

- [ ]* 4.2 Write unit tests for hierarchical reference ID search
  - Test epic to user stories lookup
  - Test user story to requirements lookup
  - Test user story to acceptance criteria lookup
  - Test steering document to epics lookup
  - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5, 2.6_

- [ ] 5. Integrate reference ID search with MCP handler
  - Modify handleSearchGlobal to support entity_types parameter
  - Add reference ID pattern detection to search flow
  - Route to appropriate search method based on pattern and entity types
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 2.1, 2.2, 2.3, 2.4, 2.5, 2.6_

- [ ] 5.1 Update MCP tool schema for entity_types parameter
  - Add entity_types parameter to search_global tool schema
  - Update parameter validation in MCP tools handler
  - Ensure backward compatibility with existing searches
  - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5, 2.6_

- [ ]* 5.2 Write integration tests for MCP handler
  - Test search_global with reference ID queries
  - Test entity_types parameter handling
  - Test error responses for invalid reference IDs
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 2.1, 2.2, 2.3, 2.4, 2.5, 2.6_

- [ ] 6. Implement search result prioritization for reference IDs
  - Add relevance scoring for exact reference ID matches
  - Ensure reference ID matches appear first in results
  - Maintain existing relevance scoring for text searches
  - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.5_

- [ ]* 6.1 Write unit tests for search result prioritization
  - Test reference ID matches have highest relevance
  - Test mixed search results ordering
  - Test relevance scoring consistency
  - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.5_

- [ ] 7. Add performance optimizations for reference ID searches
  - Implement database indexes for reference ID columns
  - Add caching for reference ID to UUID mappings
  - Optimize hierarchical lookup queries
  - _Requirements: 5.1, 5.2, 5.3, 5.4, 5.5_

- [ ] 7.1 Create database migration for reference ID indexes
  - Add indexes on reference_id columns for all entity tables
  - Add indexes on foreign key columns for hierarchical lookups
  - Ensure indexes are created concurrently to avoid downtime
  - _Requirements: 5.1, 5.2, 5.3_

- [ ]* 7.2 Write performance tests for reference ID searches
  - Benchmark direct reference ID lookups
  - Benchmark hierarchical reference ID searches
  - Compare performance with existing text searches
  - _Requirements: 5.1, 5.2, 5.3, 5.4, 5.5_

- [ ] 8. Add comprehensive error handling
  - Implement ReferenceIDNotFoundError for missing reference IDs
  - Add InvalidEntityTypeError for invalid entity type combinations
  - Update error responses in MCP handler
  - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5, 2.6_

- [ ]* 8.1 Write unit tests for error handling
  - Test reference ID not found scenarios
  - Test invalid entity type combinations
  - Test error message formatting
  - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5, 2.6_

- [ ] 9. Update search caching for reference ID searches
  - Extend cache key generation to include entity types
  - Add cache invalidation for reference ID searches
  - Implement separate cache TTL for reference ID lookups
  - _Requirements: 5.4, 5.5_

- [ ]* 9.1 Write unit tests for search caching
  - Test cache key generation with entity types
  - Test cache invalidation scenarios
  - Test cache hit/miss behavior
  - _Requirements: 5.4, 5.5_

- [ ] 10. Add end-to-end tests for MCP agent scenarios
  - Test all documented usage patterns from requirements
  - Simulate real MCP agent search behavior
  - Validate complete search flow from MCP request to response
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 2.1, 2.2, 2.3, 2.4, 2.5, 2.6_

- [ ]* 10.1 Write comprehensive end-to-end tests
  - Test "US-119" direct lookup
  - Test "EP-006" with entity_types ["user_story"] hierarchical lookup
  - Test case-insensitive searches
  - Test error scenarios
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 2.1, 2.2, 2.3, 2.4, 2.5, 2.6, 3.1, 3.2, 3.3, 3.4, 3.5_
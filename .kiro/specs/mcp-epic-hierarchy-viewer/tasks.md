# Implementation Plan: MCP Epic Hierarchy Viewer

## Task List

- [x] 1. Add repository method for complete hierarchy retrieval
  - Add `GetCompleteHierarchy(id uuid.UUID)` method to `EpicRepository`
  - Use GORM nested preloading for UserStories, Requirements, and AcceptanceCriteria
  - Maintain natural ordering with `Order("created_at ASC")`
  - _Requirements: REQ-089_

- [x] 2. Enhance Epic Service with hierarchy retrieval
  - Add `GetEpicWithCompleteHierarchy(id uuid.UUID)` method to `EpicService` interface
  - Implement method in `epicService` - calls `epicRepo.GetCompleteHierarchy()`
  - Handle errors and map repository errors to service errors
  - Return complete Epic model with all nested entities preloaded
  - _Requirements: REQ-089_

- [x] 3. Add epic_hierarchy tool to existing EpicHandler
  - [x] 3.1 Update EpicHandler in `internal/mcp/tools/epic.go`
    - Add `ToolEpicHierarchy` constant
    - Update `GetSupportedTools()` to include `ToolEpicHierarchy`
    - Update `HandleTool()` switch to route `epic_hierarchy` calls
    - _Requirements: REQ-089_

  - [x] 3.2 Implement GetHierarchy method
    - Add `GetHierarchy(ctx context.Context, args map[string]interface{})` method
    - Validate and parse epic parameter (reference ID or UUID)
    - Call `epicService.GetEpicWithCompleteHierarchy()` to retrieve data
    - Format output using `formatTree()` method
    - Return MCP response with formatted tree
    - _Requirements: REQ-089_

  - [x] 3.3 Implement tree formatting methods (private methods in EpicHandler)
    - Implement `formatTree(epic *models.Epic) string` method
    - Implement `formatUserStory()` method with proper indentation
    - Implement `formatRequirement()` method for requirement display
    - Implement `formatAcceptanceCriteria()` method with truncation
    - Implement `truncateDescription()` helper for 80-char limit with UTF-8 support
    - Handle empty states (no user stories, no requirements, no acceptance criteria)
    - _Requirements: REQ-089, REQ-091_

  - [x] 3.4 Implement error handling
    - Return JSON-RPC error for invalid epic reference ID format
    - Return JSON-RPC error for epic not found
    - Return JSON-RPC error for database errors
    - Reuse existing error handling patterns from other EpicHandler methods
    - _Requirements: REQ-090_

- [x] 4. Add tool schema definition
  - Add `epic_hierarchy` tool definition to `internal/mcp/schemas/tools.go`
  - Define input schema with `epic` parameter (string, EP-XXX or UUID)
  - Add description and examples
  - _Requirements: REQ-089, REQ-090_

- [x] 5. Write unit tests
  - [x] 5.1 Test repository method
    - Test `GetCompleteHierarchy()` with complete hierarchy
    - Test with empty user stories
    - Test with missing requirements/acceptance criteria
    - Test error cases (epic not found)
    - _Requirements: REQ-089, REQ-090_

  - [x] 5.2 Test service method
    - Test `GetEpicWithCompleteHierarchy()` returns complete epic
    - Test error handling (epic not found)
    - Test repository error mapping
    - _Requirements: REQ-089, REQ-090_

  - [x] 5.3 Test handler formatting methods
    - Test `formatTree()` with complete hierarchy
    - Test empty state messages
    - Test acceptance criteria truncation (80 chars)
    - Test UTF-8 character handling (Cyrillic, emoji)
    - Test first sentence extraction
    - Test proper indentation and tree characters
    - _Requirements: REQ-089, REQ-091_

  - [x] 5.4 Test EpicHandler.GetHierarchy method
    - Test valid reference ID input (EP-XXX)
    - Test valid UUID input
    - Test invalid reference ID format
    - Test epic not found scenario
    - Test JSON-RPC response format
    - Test integration with existing EpicHandler
    - _Requirements: REQ-089, REQ-090_

- [x] 6. Write integration tests
  - Create test epic with full hierarchy (epic → user stories → requirements + acceptance criteria)
  - Call `epic_hierarchy` tool via MCP handler
  - Verify output format matches specification
  - Test with PostgreSQL database
  - Verify single query execution (no N+1)
  - _Requirements: REQ-089, REQ-091_

- [ ]* 7. Verify tool registration
  - Verify `epic_hierarchy` tool appears in `tools/list` response (should work automatically via EpicHandler)
  - No changes needed to `internal/handlers/mcp_handler.go` (EpicHandler already registered)
  - _Requirements: REQ-089_

- [ ]* 8. Manual testing and validation
  - Test with real epic data from Spexus
  - Verify ASCII tree characters render correctly
  - Test with large hierarchies (50+ entities)
  - Verify performance (< 100ms for typical hierarchies)
  - Test with Unicode titles and descriptions
  - _Requirements: REQ-089, REQ-091_

- [x] 9. Add steering documents support to hierarchy
  - [x] 9.1 Update repository method to preload steering documents
    - Add `Preload("SteeringDocuments")` to `GetCompleteHierarchy()` method
    - Load steering documents via many-to-many relationship
    - Maintain natural ordering with `Order("created_at ASC")`
    - _Requirements: REQ-089_

  - [x] 9.2 Implement steering document formatting
    - Add `formatSteeringDocument()` method for steering document display (no status/priority)
    - Update `formatTree()` to display steering documents first, then user stories (both at same indentation level)
    - Handle empty states (no steering documents or user stories)
    - _Requirements: REQ-089, REQ-091_

  - [x] 9.3 Update tool schema description
    - Update `epic_hierarchy` tool description to mention steering documents in hierarchy
    - _Requirements: REQ-089_

  - [x] 9.4 Test steering documents functionality
    - Test repository preloads steering documents via many-to-many relationship
    - Test `formatSteeringDocument()` displays without status/priority
    - Test steering documents appear before user stories at same indentation level
    - Test empty state messages (no steering documents or user stories)
    - Test steering document description truncation (80 chars)
    - Test first sentence extraction for steering documents
    - _Requirements: REQ-089, REQ-091_

  - [x] 9.5 Integration test with steering documents
    - Create test epic with steering documents linked via many-to-many relationship
    - Verify output format includes steering documents
    - Verify query execution (6 queries total with steering documents)
    - _Requirements: REQ-089, REQ-091_

  - [x] 9.6 Manual validation with steering documents
    - Test with real epic data including steering documents
    - Verify steering documents appear before user stories at same level
    - Test with Unicode titles and descriptions in steering documents
    - _Requirements: REQ-089, REQ-091_

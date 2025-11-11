# Design Document: MCP Epic Hierarchy Viewer

## Overview

The MCP Epic Hierarchy Viewer is a new MCP tool that provides a compact, text-based visualization of an epic's complete hierarchical structure. This tool enables AI coding agents to quickly understand the full context of an epic including its steering documents, user stories, requirements, and acceptance criteria without making multiple API calls or manually navigating through entities.

The hierarchy structure is: **Epic → [Steering Documents, User Stories] → [Requirements, Acceptance Criteria]** where Steering Documents and User Stories are displayed at the same hierarchical level under the Epic. Requirements and Acceptance Criteria are displayed at the same hierarchical level under each User Story. Requirements are listed first, followed by Acceptance Criteria.

The tool will be implemented as `epic_hierarchy` and will follow the existing MCP tool architecture patterns established in the Spexus system.

### Database Schema Validation

The hierarchy **Epic → [Steering Documents, User Stories] → [Requirements, Acceptance Criteria]** is validated against the database schema:

✅ **Epic → Steering Documents**: Many-to-many relationship through `epic_steering_documents` junction table
- One epic can have many steering documents
- One steering document can be linked to many epics
- Indexed on both `epic_id` and `steering_document_id` for efficient queries

✅ **Epic → User Stories**: `user_stories.epic_id` → `epics.id` (ON DELETE CASCADE)
- One epic can have many user stories
- Indexed on `epic_id` for efficient queries

✅ **User Story → Requirements**: `requirements.user_story_id` → `user_stories.id` (ON DELETE CASCADE)
- One user story can have many requirements
- Indexed on `user_story_id` for efficient queries

✅ **User Story → Acceptance Criteria**: `acceptance_criteria.user_story_id` → `user_stories.id` (ON DELETE CASCADE)
- One user story can have many acceptance criteria
- **REQUIRED**: Every acceptance criteria MUST belong to a user story (`user_story_id NOT NULL`)
- Indexed on `user_story_id` for efficient queries

✅ **Requirement → Acceptance Criteria** (Optional Link): `requirements.acceptance_criteria_id` → `acceptance_criteria.id` (ON DELETE SET NULL)
- One requirement can be optionally linked to ONE acceptance criteria
- One acceptance criteria can be linked to MULTIPLE requirements
- This is a **many-to-one** relationship from Requirement perspective
- Indexed on `acceptance_criteria_id` for efficient queries
- **Note**: This optional link is NOT used in the hierarchy display

**Hierarchy Display Strategy**:
- Steering Documents and User Stories are displayed at the same level under Epic
- Steering Documents are listed first (all steering documents linked to the epic)
- User Stories are listed second (all user stories for the epic)
- Requirements and Acceptance Criteria are displayed at the same level under User Story
- Requirements are listed first (all requirements for the user story)
- Acceptance Criteria are listed second (all acceptance criteria for the user story)
- The optional `requirements.acceptance_criteria_id` link is ignored for display purposes

## Architecture

### Component Structure

The implementation follows the established MCP tool architecture with these key components:

```
internal/mcp/tools/
├── epic_hierarchy.go          # New: Epic hierarchy tool handler
├── epic.go                    # Existing: Epic CRUD operations
└── handler.go                 # Existing: Tool routing and registration

internal/mcp/schemas/
└── tools.go                   # Update: Add epic_hierarchy tool definition

internal/service/
├── epic_service.go            # Existing: Epic business logic
└── epic_hierarchy_service.go  # New: Hierarchy formatting service

internal/repository/
└── epic_repository.go         # Existing: Epic data access (may need enhancements)
```

### Design Decisions

#### 1. Separate Service Layer for Formatting
**Decision**: Create a dedicated `EpicHierarchyService` for tree formatting logic rather than embedding it in the tool handler.

**Rationale**:
- Separates data retrieval from presentation logic
- Makes the formatting logic testable independently
- Allows reuse of formatting logic in other contexts (e.g., CLI tools, reports)
- Follows single responsibility principle

#### 2. Single Database Query with Preloading
**Decision**: Retrieve the entire hierarchy in one database query using GORM's preloading capabilities.

**Rationale**:
- Minimizes database round trips (N+1 query problem)
- Improves performance for large hierarchies
- Leverages existing GORM relationship definitions
- Consistent with existing patterns in the codebase

#### 3. ASCII Tree Characters
**Decision**: Use standard ASCII tree characters (├──, └──, │) with 2-space indentation.

**Rationale**:
- Universal compatibility across terminals and text editors
- Clear visual hierarchy representation
- Consistent with common CLI tool conventions (e.g., `tree` command)
- Easy to parse visually for both humans and AI agents

#### 4. Reference ID as Primary Input
**Decision**: Accept only reference IDs (EP-XXX format) as input, not UUIDs.

**Rationale**:
- More user-friendly for AI agents and humans
- Consistent with MCP tool patterns in the system
- Reference IDs are the primary way users identify entities
- Simplifies the tool interface

#### 5. Truncation Strategy for Acceptance Criteria
**Decision**: Display only the first sentence of acceptance criteria descriptions, truncated to 80 characters.

**Rationale**:
- Keeps output compact and scannable
- Prevents overwhelming the agent with too much detail
- First sentence typically contains the most important information
- 80 characters is a standard terminal width consideration

## Components and Interfaces

### 1. Enhance Existing EpicHandler

Add the `epic_hierarchy` tool to the existing `EpicHandler` in `internal/mcp/tools/epic.go`:

```go
// EpicHandler already exists - add new method and update GetSupportedTools()
type EpicHandler struct {
    epicService     service.EpicService
    userService     service.UserService
    statusValidator validation.StatusValidator
}

// Update GetSupportedTools to include epic_hierarchy
func (h *EpicHandler) GetSupportedTools() []string {
    return []string{
        ToolCreateEpic,
        ToolUpdateEpic,
        ToolListEpics,
        ToolEpicHierarchy, // NEW
    }
}

// Update HandleTool to route epic_hierarchy calls
func (h *EpicHandler) HandleTool(ctx context.Context, toolName string, args map[string]interface{}) (interface{}, error) {
    switch toolName {
    case ToolListEpics:
        return h.List(ctx, args)
    case ToolCreateEpic:
        return h.Create(ctx, args)
    case ToolUpdateEpic:
        return h.Update(ctx, args)
    case ToolEpicHierarchy: // NEW
        return h.GetHierarchy(ctx, args)
    default:
        return nil, jsonrpc.NewMethodNotFoundError(fmt.Sprintf("Unknown Epic tool: %s", toolName))
    }
}

// NEW: GetHierarchy handles the epic_hierarchy tool
func (h *EpicHandler) GetHierarchy(ctx context.Context, args map[string]interface{}) (interface{}, error)

// NEW: Private formatting methods
func (h *EpicHandler) formatTree(epic *models.Epic) string
func (h *EpicHandler) formatUserStory(builder *strings.Builder, us models.UserStory, isLast bool, indent string)
func (h *EpicHandler) formatRequirement(builder *strings.Builder, req models.Requirement, indent string)
func (h *EpicHandler) formatAcceptanceCriteria(builder *strings.Builder, ac models.AcceptanceCriteria, isLast bool, indent string)
func (h *EpicHandler) truncateDescription(desc string, maxLength int) string
```

**Responsibilities**:
- Add `epic_hierarchy` tool to existing `EpicHandler`
- Validate input parameters (epic reference ID)
- Parse reference ID or UUID (reuse existing helper functions)
- Call service layer to retrieve epic with complete hierarchy
- **Format the hierarchy as ASCII tree** (presentation logic)
- Handle empty states (no user stories, no requirements, etc.)
- Apply truncation rules for acceptance criteria
- Handle errors and return JSON-RPC compliant responses
- Follow existing patterns from other methods in `epic.go`

**Rationale**:
- Reuses existing `EpicHandler` - no new handler needed
- All epic-related tools in one place
- Shares common dependencies (epicService, userService)
- Consistent with existing architecture
- Tool handler is presentation layer - responsible for output formatting

### 2. Epic Service Enhancement

The existing `EpicService` will be enhanced with a new method to retrieve complete hierarchy:

```go
// Add to EpicService interface
type EpicService interface {
    // ... existing methods ...
    GetEpicWithCompleteHierarchy(id uuid.UUID) (*models.Epic, error)
}

// Implementation in epicService
func (s *epicService) GetEpicWithCompleteHierarchy(id uuid.UUID) (*models.Epic, error) {
    epic, err := s.epicRepo.GetCompleteHierarchy(id)
    if err != nil {
        if errors.Is(err, repository.ErrNotFound) {
            return nil, ErrEpicNotFound
        }
        return nil, fmt.Errorf("failed to get epic hierarchy: %w", err)
    }
    return epic, nil
}
```

**Responsibilities**:
- Retrieve epic with all nested entities using `GetCompleteHierarchy()`
- Return the complete Epic model with preloaded relationships
- Handle repository errors and map to service errors
- **No formatting logic** - just data retrieval

**Rationale**:
- Reuses existing `EpicService` - no need for separate hierarchy service
- Service layer focuses on business logic and data access only
- Formatting is presentation concern - belongs in tool handler
- Follows single responsibility and separation of concerns principles

### 3. Repository Enhancements

#### Required Repository Changes

**Note**: The `AcceptanceCriteriaRepository` currently has `GetByUserStory()` but NOT `GetByRequirement()`. However, this is not needed because we use GORM's nested preloading which automatically handles the relationship through `Requirement.AcceptanceCriteria`.

#### Option A: Single Query with Nested Preloading (Recommended)

Add a new method to `EpicRepository` to retrieve the complete hierarchy in one query:

```go
// GetCompleteHierarchy retrieves an epic with all nested entities preloaded
func (r *epicRepository) GetCompleteHierarchy(id uuid.UUID) (*models.Epic, error) {
    var epic models.Epic
    err := r.GetDB().
        Preload("SteeringDocuments", func(db *gorm.DB) *gorm.DB {
            return db.Order("created_at ASC")
        }).
        Preload("UserStories", func(db *gorm.DB) *gorm.DB {
            return db.Order("created_at ASC")
        }).
        Preload("UserStories.Requirements", func(db *gorm.DB) *gorm.DB {
            return db.Order("created_at ASC")
        }).
        Preload("UserStories.Requirements.Type").
        Preload("UserStories.AcceptanceCriteria", func(db *gorm.DB) *gorm.DB {
            return db.Order("created_at ASC")
        }).
        Where("id = ?", id).
        First(&epic).Error
    
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, ErrNotFound
        }
        return nil, r.handleDBError(err)
    }
    return &epic, nil
}
```

**Note**: We preload `UserStories.Requirements` and `UserStories.AcceptanceCriteria` separately since they are both direct children of UserStory, not nested under each other.

**Advantages**:
- Single database query (optimal performance)
- Leverages GORM's built-in preloading
- Maintains natural ordering with `Order()` clauses
- No N+1 query problem

**SQL Generated** (approximately):
```sql
-- Query 1: Get epic
SELECT * FROM epics WHERE id = ?;

-- Query 2: Get all steering documents for this epic (through junction table)
SELECT steering_documents.* FROM steering_documents
INNER JOIN epic_steering_documents ON steering_documents.id = epic_steering_documents.steering_document_id
WHERE epic_steering_documents.epic_id = ? ORDER BY created_at ASC;

-- Query 3: Get all user stories for this epic
SELECT * FROM user_stories WHERE epic_id = ? ORDER BY created_at ASC;

-- Query 4: Get all requirements for these user stories
SELECT * FROM requirements WHERE user_story_id IN (?, ?, ...) ORDER BY created_at ASC;

-- Query 5: Get all requirement types
SELECT * FROM requirement_types WHERE id IN (?, ?, ...);

-- Query 6: Get all acceptance criteria for these user stories
SELECT * FROM acceptance_criteria WHERE user_story_id IN (?, ?, ...) ORDER BY created_at ASC;
```

#### Option B: Recursive Service Calls (Alternative)

Use existing repository methods with recursive calls in the service layer:

```go
func (s *epicHierarchyService) GetHierarchyTree(epicID uuid.UUID) (string, error) {
    // 1. Get epic
    epic, err := s.epicRepo.GetByID(epicID)
    if err != nil {
        return "", err
    }
    
    // 2. Get user stories for epic
    userStories, err := s.userStoryRepo.GetByEpic(epicID)
    if err != nil {
        return "", err
    }
    
    // 3. For each user story, get requirements
    for i := range userStories {
        requirements, err := s.requirementRepo.GetByUserStory(userStories[i].ID)
        if err != nil {
            return "", err
        }
        userStories[i].Requirements = requirements
        
        // 4. For each requirement, get acceptance criteria
        for j := range requirements {
            criteria, err := s.acceptanceCriteriaRepo.GetByRequirement(requirements[j].ID)
            if err != nil {
                return "", err
            }
            requirements[j].AcceptanceCriteria = criteria
        }
    }
    
    epic.UserStories = userStories
    return s.formatTree(epic), nil
}
```

**Disadvantages**:
- Multiple database queries (N+1 problem)
- Slower performance for large hierarchies
- More complex error handling

**Decision**: Use **Option A** (nested preloading) for optimal performance and simplicity.

### 4. Tool Schema Definition

Add to `internal/mcp/schemas/tools.go`:

```go
{
    Name:        "epic_hierarchy",
    Title:       "View Epic Hierarchy",
    Description: "Display the complete hierarchical structure of an epic including user stories, requirements, and acceptance criteria in a compact ASCII tree format",
    InputSchema: map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "epic": map[string]interface{}{
                "type":        "string",
                "description": "Epic reference ID (e.g., EP-001) or UUID",
                "pattern":     "^(EP-\\d+|[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12})$",
            },
        },
        "required": []string{"epic"},
    },
}
```

## Data Models

### Existing Model Relationships

The implementation leverages existing GORM model relationships:

```go
// Epic model (internal/models/epic.go)
type Epic struct {
    ID                 uuid.UUID
    ReferenceID        string
    Title              string
    Status             EpicStatus
    Priority           Priority
    SteeringDocuments  []SteeringDocument `gorm:"many2many:epic_steering_documents"` // Many-to-many
    UserStories        []UserStory        `gorm:"foreignKey:EpicID"`                 // One-to-many
}

// SteeringDocument model (internal/models/steering_document.go)
type SteeringDocument struct {
    ID          uuid.UUID
    ReferenceID string
    Title       string
    Description string
    Epics       []Epic `gorm:"many2many:epic_steering_documents"` // Many-to-many
}

// UserStory model (internal/models/user_story.go)
type UserStory struct {
    ID                  uuid.UUID
    ReferenceID         string
    Title               string
    Status              UserStoryStatus
    Priority            Priority
    EpicID              uuid.UUID
    Requirements        []Requirement        `gorm:"foreignKey:UserStoryID"` // One-to-many
    AcceptanceCriteria  []AcceptanceCriteria `gorm:"foreignKey:UserStoryID"` // One-to-many
}

// Requirement model (internal/models/requirement.go)
type Requirement struct {
    ID                   uuid.UUID
    ReferenceID          string
    Title                string
    Status               RequirementStatus
    Priority             Priority
    UserStoryID          uuid.UUID
    AcceptanceCriteriaID *uuid.UUID              // Optional foreign key
    AcceptanceCriteria   *AcceptanceCriteria     `gorm:"foreignKey:AcceptanceCriteriaID"` // Many-to-one (optional)
    TypeID               uuid.UUID
    Type                 RequirementType         `gorm:"foreignKey:TypeID"` // Many-to-one
}

// AcceptanceCriteria model (internal/models/acceptance_criteria.go)
type AcceptanceCriteria struct {
    ID           uuid.UUID
    ReferenceID  string
    Description  string
    UserStoryID  uuid.UUID
    Requirements []Requirement `gorm:"foreignKey:AcceptanceCriteriaID"` // One-to-many
}
```

**Important Note**: The relationship between `Requirement` and `AcceptanceCriteria` is **many-to-one** (or one-to-many from AC perspective):
- One requirement can be linked to ZERO or ONE acceptance criteria (optional)
- One acceptance criterion can be linked to MULTIPLE requirements
- Database schema: `requirements.acceptance_criteria_id` → `acceptance_criteria.id` (ON DELETE SET NULL)
- GORM handles this with `Preload("AcceptanceCriteria")` on Requirement model

### Hierarchy Node Structure

The service will work with the existing models but organize them hierarchically:

```
Epic
├── Status: EpicStatus (Backlog, Draft, In Progress, Done, Cancelled)
├── Priority: Priority (1-4)
├── SteeringDocuments []SteeringDocument (displayed first, same level as UserStories)
│   ├── ReferenceID: string (STD-XXX)
│   ├── Title: string
│   └── Description: string (first sentence, max 80 chars)
└── UserStories []UserStory (displayed second, same level as SteeringDocuments)
    ├── Status: UserStoryStatus
    ├── Priority: Priority
    ├── Requirements []Requirement (displayed first)
    │   ├── Status: RequirementStatus (Draft, Active, Obsolete)
    │   ├── Priority: Priority
    │   └── Type: RequirementType
    └── AcceptanceCriteria []AcceptanceCriteria (displayed second)
        └── Description: string (first sentence, max 80 chars)
```

**Display Order**:
1. All Steering Documents for the Epic (at same level as User Stories)
2. All User Stories for the Epic (at same level as Steering Documents)
3. Within each User Story:
   - All Requirements for the User Story
   - All Acceptance Criteria for the User Story

### Output Format Structure

```
EP-XXX [Status] [Priority] Title
│
├── STD-XXX Title
├── STD-XXX Title
│
├─┬ US-XXX [Status] [Priority] Title
│ │
│ ├── REQ-XXX [Status] [Priority] Title
│ ├── REQ-XXX [Status] [Priority] Title
│ │
│ └── AC-XXX — First sentence of description...
│
└─┬ US-XXX [Status] [Priority] Title
  │
  ├── No requirements
  │
  └── No acceptance criteria
```

**Important Design Decisions**: 
- Steering Documents and User Stories are displayed at the SAME hierarchical level under Epic
- Steering Documents are listed FIRST (all steering documents linked to the epic)
- User Stories are listed SECOND (all user stories for the epic)
- Steering Documents do NOT have status or priority indicators (they don't have these attributes)
- Requirements and Acceptance Criteria are displayed at the SAME hierarchical level under User Story
- Requirements are listed FIRST (all requirements for the user story)
- Acceptance Criteria are listed SECOND (all acceptance criteria for the user story)
- The optional `requirements.acceptance_criteria_id` link is IGNORED for display purposes

**Rationale**:
- Reflects the actual database structure (both SteeringDocuments and UserStories belong to Epic)
- Steering Documents provide context before diving into implementation details
- Simpler and clearer hierarchy
- No duplication or confusion about placement
- All entities for an Epic are visible at appropriate levels
- Easier to understand the complete scope of an Epic

## Error Handling

### Error Types and Responses

1. **Epic Not Found**
   - Condition: Reference ID doesn't exist
   - Response: JSON-RPC error with message "Epic EP-XXX not found"
   - HTTP Status: 400 (via JSON-RPC error mapping)

2. **Invalid Reference ID Format**
   - Condition: Input doesn't match EP-XXX pattern or valid UUID
   - Response: JSON-RPC error with message "Invalid epic reference ID format"
   - HTTP Status: 400

3. **Database Errors**
   - Condition: Database connection or query failures
   - Response: JSON-RPC error with generic message (don't expose internals)
   - HTTP Status: 500

4. **Authentication Errors**
   - Condition: Missing or invalid JWT token
   - Response: JSON-RPC error "Authentication required"
   - HTTP Status: 401

### Error Handling Pattern

Follow the existing pattern from `epic.go`:

```go
if err != nil {
    if errors.Is(err, service.ErrEpicNotFound) {
        return nil, jsonrpc.NewInvalidParamsError("Epic not found")
    }
    return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to retrieve hierarchy: %v", err))
}
```

## Testing Strategy

### Unit Tests

1. **Handler Tests** (`epic_hierarchy_test.go`)
   - Test valid reference ID input
   - Test invalid reference ID formats
   - Test epic not found scenarios
   - Test error handling and JSON-RPC response format

2. **Service Tests** (`epic_hierarchy_service_test.go`)
   - Test tree formatting with complete hierarchy
   - Test empty states (no user stories, no requirements)
   - Test acceptance criteria truncation
   - Test Unicode character handling
   - Test proper indentation and tree characters

3. **Repository Tests** (if new methods added)
   - Test complete hierarchy preloading
   - Test query performance with nested entities

### Integration Tests

1. **End-to-End MCP Tool Test**
   - Create test epic with full hierarchy
   - Call epic_hierarchy tool via MCP handler
   - Verify output format matches specification
   - Test with real database (PostgreSQL)

2. **Performance Tests**
   - Test with large hierarchies (100+ entities)
   - Measure query execution time
   - Verify single query execution (no N+1)

### Test Data Setup

Use existing test helpers from `internal/integration/test_helpers.go`:
- `CreateTestEpic()`
- `CreateTestUserStory()`
- `CreateTestRequirement()`
- `CreateTestAcceptanceCriteria()`

## Implementation Details

### Tree Formatting Algorithm (in EpicHandler)

The formatting logic is implemented in the existing `EpicHandler`, not in the service layer:

```go
// GetHierarchy implementation (new method in EpicHandler)
func (h *EpicHandler) GetHierarchy(ctx context.Context, args map[string]interface{}) (interface{}, error) {
    // 1. Validate and parse epic ID
    epicIDStr, ok := getStringArg(args, "epic")
    if !ok || epicIDStr == "" {
        return nil, jsonrpc.NewInvalidParamsError("Missing or invalid 'epic' argument")
    }
    
    // 2. Parse UUID or reference ID
    epicID, err := parseUUIDOrReferenceID(epicIDStr, func(refID string) (interface{}, error) {
        return h.epicService.GetEpicByReferenceID(refID)
    })
    if err != nil {
        return nil, jsonrpc.NewInvalidParamsError("Invalid 'epic': not a valid UUID or reference ID")
    }
    
    // 3. Retrieve epic with complete hierarchy
    epic, err := h.epicService.GetEpicWithCompleteHierarchy(epicID)
    if err != nil {
        if errors.Is(err, service.ErrEpicNotFound) {
            return nil, jsonrpc.NewInvalidParamsError("Epic not found")
        }
        return nil, jsonrpc.NewInternalError(fmt.Sprintf("Failed to retrieve hierarchy: %v", err))
    }
    
    // 4. Format as ASCII tree
    treeOutput := h.formatTree(epic)
    
    // 5. Return MCP response
    return types.CreateDataResponse(treeOutput, nil), nil
}

func (h *EpicHandler) formatTree(epic *models.Epic) string {
    var builder strings.Builder
    
    // Epic root node
    builder.WriteString(fmt.Sprintf("%s [%s] [P%d] %s\n",
        epic.ReferenceID, epic.Status, epic.Priority, epic.Title))
    
    if len(epic.SteeringDocuments) == 0 && len(epic.UserStories) == 0 {
        builder.WriteString("│\n")
        builder.WriteString("└── No steering documents or user stories attached\n")
        return builder.String()
    }
    
    builder.WriteString("│\n")
    
    // Display steering documents first (at same level as user stories)
    for i, std := range epic.SteeringDocuments {
        h.formatSteeringDocument(&builder, std)
    }
    
    // Display user stories second (at same level as steering documents)
    for i, us := range epic.UserStories {
        isLastUS := i == len(epic.UserStories)-1
        h.formatUserStory(&builder, us, isLastUS, "")
    }
    
    return builder.String()
}

func (h *EpicHandler) formatSteeringDocument(builder *strings.Builder, std models.SteeringDocument) {
    // Steering documents don't have status or priority
    builder.WriteString(fmt.Sprintf("├── %s %s\n",
        std.ReferenceID, std.Title))
}

func (h *EpicHandler) formatUserStory(builder *strings.Builder, us models.UserStory, isLast bool, indent string) {
    // User story prefix
    prefix := "├─┬"
    if isLast {
        prefix = "└─┬"
    }
    
    builder.WriteString(fmt.Sprintf("%s %s [%s] [P%d] %s\n",
        prefix, us.ReferenceID, us.Status, us.Priority, us.Title))
    
    childIndent := "│ "
    if isLast {
        childIndent = "  "
    }
    builder.WriteString(fmt.Sprintf("%s│\n", childIndent))
    
    // Display requirements first
    if len(us.Requirements) == 0 {
        builder.WriteString(fmt.Sprintf("%s├── No requirements\n", childIndent))
    } else {
        for i, req := range us.Requirements {
            h.formatRequirement(builder, req, childIndent)
        }
    }
    
    // Display acceptance criteria second
    if len(us.AcceptanceCriteria) == 0 {
        builder.WriteString(fmt.Sprintf("%s│\n", childIndent))
        builder.WriteString(fmt.Sprintf("%s└── No acceptance criteria\n", childIndent))
    } else {
        builder.WriteString(fmt.Sprintf("%s│\n", childIndent))
        for i, ac := range us.AcceptanceCriteria {
            isLastAC := i == len(us.AcceptanceCriteria)-1
            h.formatAcceptanceCriteria(builder, ac, isLastAC, childIndent)
        }
    }
}

func (h *EpicHandler) formatRequirement(builder *strings.Builder, req models.Requirement, indent string) {
    builder.WriteString(fmt.Sprintf("%s├── %s [%s] [P%d] %s\n",
        indent, req.ReferenceID, req.Status, req.Priority, req.Title))
}

func (h *EpicHandler) formatAcceptanceCriteria(builder *strings.Builder, ac models.AcceptanceCriteria, isLast bool, indent string) {
    prefix := "├──"
    if isLast {
        prefix = "└──"
    }
    
    truncatedDesc := h.truncateDescription(ac.Description, 80)
    builder.WriteString(fmt.Sprintf("%s%s %s — %s\n",
        indent, prefix, ac.ReferenceID, truncatedDesc))
}

func (h *EpicHandler) truncateDescription(desc string, maxLength int) string {
    // Extract first sentence
    sentences := strings.SplitN(desc, ".", 2)
    firstSentence := strings.TrimSpace(sentences[0])
    
    // Handle case where there's no period (single sentence)
    if firstSentence == "" && len(desc) > 0 {
        firstSentence = desc
    }
    
    // Truncate to max length (accounting for UTF-8)
    runes := []rune(firstSentence)
    if len(runes) > maxLength {
        return string(runes[:maxLength-3]) + "..."
    }
    
    return firstSentence
}
```

**Key Implementation Details**:

1. **Indentation Management**: Uses `indent` parameter to track current indentation level
2. **Tree Characters**: 
   - `├─┬` for non-last user story with children
   - `└─┬` for last user story with children
   - `├──` for requirements (all use same prefix)
   - `├──` for non-last acceptance criteria
   - `└──` for last acceptance criteria
   - `│` for vertical continuation lines
3. **Display Order**: Requirements first, then acceptance criteria
4. **UTF-8 Handling**: Uses `[]rune` for accurate character counting (not bytes)
5. **Empty State Messages**: Clear messages for missing requirements/acceptance criteria

### Acceptance Criteria Truncation

The truncation logic is integrated into the `formatRequirement` method (see Tree Formatting Algorithm above).

**Key Points**:
- Extracts first sentence by splitting on period (`.`)
- Handles descriptions without periods (single sentence)
- Uses `[]rune` for UTF-8 character counting (not byte length)
- Appends `...` if truncated
- Maximum 80 characters as specified in requirements

### Unicode Handling

The implementation uses Go's native UTF-8 string handling:
- `strings.Builder` for efficient string concatenation
- `[]rune` conversion for accurate character counting (not byte length)
- Properly handles Cyrillic, emoji, and other multi-byte UTF-8 characters
- Example: "Просмотр иерархии эпика" is counted as 23 characters, not 43 bytes

## Performance Considerations

### Database Query Optimization

1. **Single Query with Nested Preloading**
   - Use GORM's nested `Preload()` to fetch all related entities
   - Executes 6 queries total (1 for epic + 5 for nested entities including steering documents)
   - Avoid N+1 query problem completely
   - Expected query time: < 100ms for typical hierarchies (< 50 user stories)

2. **Query Execution Plan**
   ```sql
   -- Query 1: Get epic
   SELECT * FROM epics WHERE id = ?;
   
   -- Query 2: Get all steering documents for this epic (through junction table)
   SELECT steering_documents.* FROM steering_documents
   INNER JOIN epic_steering_documents ON steering_documents.id = epic_steering_documents.steering_document_id
   WHERE epic_steering_documents.epic_id = ? ORDER BY created_at ASC;
   
   -- Query 3: Get all user stories for this epic
   SELECT * FROM user_stories WHERE epic_id = ? ORDER BY created_at ASC;
   
   -- Query 4: Get all requirements for these user stories (IN clause)
   SELECT * FROM requirements WHERE user_story_id IN (?, ?, ...) ORDER BY created_at ASC;
   
   -- Query 5: Get all acceptance criteria for these user stories (IN clause)
   SELECT * FROM acceptance_criteria WHERE user_story_id IN (?, ?, ...) ORDER BY created_at ASC;
   
   -- Query 6: Get all requirement types (IN clause)
   SELECT * FROM requirement_types WHERE id IN (?, ?, ...);
   ```

3. **Indexing Requirements**
   - `epics.id` - Primary key (already indexed)
   - `epic_steering_documents.epic_id` - Junction table foreign key (already indexed)
   - `epic_steering_documents.steering_document_id` - Junction table foreign key (already indexed)
   - `user_stories.epic_id` - Foreign key (already indexed)
   - `requirements.user_story_id` - Foreign key (already indexed)
   - `acceptance_criteria.user_story_id` - Foreign key (already indexed)
   - `requirements.type_id` - Foreign key (already indexed)

4. **Performance Characteristics**
   - Small hierarchy (< 10 user stories): ~20-50ms
   - Medium hierarchy (10-50 user stories): ~50-100ms
   - Large hierarchy (50-100 user stories): ~100-200ms
   - Very large hierarchy (> 100 user stories): Consider pagination or depth limiting

5. **Memory Usage**
   - Typical hierarchy: ~1-5MB in memory
   - Large hierarchy (100 entities): ~10-20MB
   - GORM efficiently manages memory with pointer references

### Memory Considerations

- For large hierarchies (100+ entities), memory usage should be < 10MB
- Use `strings.Builder` for efficient string concatenation
- Avoid creating intermediate string copies

### Caching Strategy

- No caching at the tool level (data freshness is important)
- Rely on database query cache for repeated queries
- Consider adding caching in future if performance issues arise

## Security Considerations

### Authentication and Authorization

1. **JWT Token Validation**
   - All MCP tool calls require valid JWT token
   - Handled by existing middleware in `mcp_handler.go`

2. **User Permissions**
   - No additional permission checks needed (read-only operation)
   - Users can view any epic they have access to
   - Follows existing access control patterns

3. **Input Validation**
   - Validate reference ID format (EP-XXX or UUID)
   - Sanitize input to prevent injection attacks
   - Use parameterized queries (GORM handles this)

### Data Exposure

- Only expose data user has permission to view
- Don't expose internal IDs or sensitive metadata
- Follow existing patterns from other MCP tools

## Integration with Existing System

### MCP Handler Registration

Update `internal/handlers/mcp_handler.go`:

```go
func NewMCPHandler(...) *MCPHandler {
    // ... existing code ...
    
    // Register epic_hierarchy tool
    toolsHandler.RegisterHandler("epic_hierarchy", epicHierarchyHandler)
    
    // ... existing code ...
}
```

### Tool Discovery

The tool will automatically appear in `tools/list` response once registered in `schemas/tools.go`.

### Logging and Monitoring

Follow existing patterns:
- Log tool invocations with correlation ID
- Log errors with appropriate context
- Track performance metrics (execution time)
- Use existing `MCPLogger` for structured logging

## Future Enhancements

### Potential Improvements (Out of Scope for Initial Implementation)

1. **Filtering Options**
   - Filter by status (e.g., only show active requirements)
   - Filter by priority (e.g., only show P1 and P2)
   - Filter by assignee

2. **Output Formats**
   - JSON format for programmatic consumption
   - Markdown format for documentation
   - HTML format for web display

3. **Depth Control**
   - Limit hierarchy depth (e.g., only show user stories, not requirements)
   - Configurable detail level

4. **Relationship Visualization**
   - Show requirement relationships (depends_on, blocks, etc.)
   - Show linked acceptance criteria

5. **Performance Optimization**
   - Add caching layer for frequently accessed hierarchies
   - Implement pagination for very large hierarchies

## Dependencies

### Existing Dependencies
- `github.com/google/uuid` - UUID handling
- `gorm.io/gorm` - Database ORM
- Existing internal packages:
  - `internal/models` - Data models
  - `internal/repository` - Data access
  - `internal/service` - Business logic
  - `internal/jsonrpc` - JSON-RPC error handling
  - `internal/mcp/types` - MCP response types

### No New External Dependencies Required

## Deployment Considerations

### Backward Compatibility
- New tool, no breaking changes to existing functionality
- Existing MCP tools continue to work unchanged

### Database Migrations
- No database schema changes required
- Uses existing tables and relationships

### Configuration
- No new configuration parameters needed
- Uses existing database connection and MCP settings

### Rollout Strategy
1. Deploy code with new tool
2. Tool automatically available via MCP protocol
3. Update MCP client documentation
4. No downtime required

## Acceptance Criteria Mapping

This design addresses all requirements from the requirements document:

### Requirement 1: Display Epic Hierarchy Structure
- ✅ ASCII tree characters (├──, └──, │)
- ✅ Reference ID, status, priority, and title on each line (status/priority omitted for steering documents)
- ✅ 2-space indentation per level
- ✅ Natural ordering preserved
- ✅ Steering documents and user stories at same level under epic
- ✅ Steering documents displayed first, then user stories

### Requirement 2: Handle Empty and Missing Data
- ✅ "No steering documents or user stories attached" message when both are empty
- ✅ Display steering documents normally when present (no "No steering documents" message)
- ✅ Display user stories normally when present (no "No user stories" message)
- ✅ "No requirements" message for user stories without requirements
- ✅ "No acceptance criteria" message for user stories without acceptance criteria
- ✅ "Epic EP-XXX not found" error

### Requirement 3: Display Steering Documents
- ✅ Steering document reference ID (STD-XXX) and title
- ✅ No status or priority indicators for steering documents
- ✅ Description truncation to 80 characters with "..."
- ✅ Proper tree indentation at same level as user stories
- ✅ First sentence extraction for descriptions
- ✅ All steering documents displayed before user stories

### Requirement 4: Display Acceptance Criteria Details
- ✅ AC reference ID prefix
- ✅ First sentence extraction
- ✅ 80-character truncation with "..."
- ✅ Proper tree indentation at same level as requirements
- ✅ All acceptance criteria for user story are displayed (not linked to specific requirements)

### Requirement 5: Provide Error Handling
- ✅ Human-readable error messages
- ✅ JSON-RPC error responses
- ✅ Proper error codes

### Technical Constraints
- ✅ Implemented as MCP tool
- ✅ Uses existing Spexus MCP API
- ✅ Reference ID support
- ✅ Unicode character handling
- ✅ Go coding standards

## Conclusion

This design provides a comprehensive, performant, and maintainable solution for the MCP Epic Hierarchy Viewer. It follows established patterns in the codebase, addresses all requirements, and provides a solid foundation for future enhancements.

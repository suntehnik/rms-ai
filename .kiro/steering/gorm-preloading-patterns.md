# GORM Preloading Patterns

## Overview

GORM preloading is used to load related entities efficiently, avoiding the N+1 query problem. This guide covers common patterns used in the Spexus codebase.

## Basic Preloading

### Single Level Preload

Load one level of relationships:

```go
var epic models.Epic
err := db.Preload("UserStories").Where("id = ?", id).First(&epic).Error
```

**SQL Generated**:
```sql
-- Query 1: Get epic
SELECT * FROM epics WHERE id = ?;

-- Query 2: Get user stories
SELECT * FROM user_stories WHERE epic_id = ?;
```

### Multiple Preloads

Load multiple relationships at the same level:

```go
var epic models.Epic
err := db.
    Preload("Creator").
    Preload("Assignee").
    Preload("UserStories").
    Where("id = ?", id).
    First(&epic).Error
```

## Nested Preloading for Hierarchies

### Two-Level Hierarchy

Load Epic → UserStories → Requirements:

```go
var epic models.Epic
err := db.
    Preload("UserStories").
    Preload("UserStories.Requirements").
    Where("id = ?", id).
    First(&epic).Error
```

**SQL Generated**:
```sql
-- Query 1: Get epic
SELECT * FROM epics WHERE id = ?;

-- Query 2: Get user stories
SELECT * FROM user_stories WHERE epic_id = ?;

-- Query 3: Get requirements
SELECT * FROM requirements WHERE user_story_id IN (?, ?, ...);
```

### Three-Level Hierarchy

Load Epic → UserStories → Requirements → AcceptanceCriteria:

```go
var epic models.Epic
err := db.
    Preload("UserStories").
    Preload("UserStories.Requirements").
    Preload("UserStories.Requirements.AcceptanceCriteria").
    Where("id = ?", id).
    First(&epic).Error
```

### Parallel Nested Preloads

Load multiple branches at the same level:

```go
var epic models.Epic
err := db.
    Preload("UserStories").
    Preload("UserStories.Requirements").        // Branch 1
    Preload("UserStories.AcceptanceCriteria").  // Branch 2 (parallel to Requirements)
    Where("id = ?", id).
    First(&epic).Error
```

**SQL Generated**:
```sql
-- Query 1: Get epic
SELECT * FROM epics WHERE id = ?;

-- Query 2: Get user stories
SELECT * FROM user_stories WHERE epic_id = ?;

-- Query 3: Get requirements
SELECT * FROM requirements WHERE user_story_id IN (?, ?, ...);

-- Query 4: Get acceptance criteria
SELECT * FROM acceptance_criteria WHERE user_story_id IN (?, ?, ...);
```

## Preload with Conditions

### Ordering

Apply ordering to preloaded relationships:

```go
var epic models.Epic
err := db.
    Preload("UserStories", func(db *gorm.DB) *gorm.DB {
        return db.Order("created_at ASC")
    }).
    Preload("UserStories.Requirements", func(db *gorm.DB) *gorm.DB {
        return db.Order("priority ASC, created_at ASC")
    }).
    Where("id = ?", id).
    First(&epic).Error
```

### Filtering

Filter preloaded relationships:

```go
var epic models.Epic
err := db.
    Preload("UserStories", func(db *gorm.DB) *gorm.DB {
        return db.Where("status = ?", "In Progress")
    }).
    Where("id = ?", id).
    First(&epic).Error
```

### Combining Conditions

```go
var epic models.Epic
err := db.
    Preload("UserStories", func(db *gorm.DB) *gorm.DB {
        return db.
            Where("status IN ?", []string{"Backlog", "In Progress"}).
            Order("priority ASC, created_at ASC")
    }).
    Where("id = ?", id).
    First(&epic).Error
```

## Complete Hierarchy Pattern

### Epic with Full Hierarchy

This is the recommended pattern for loading complete epic hierarchy:

```go
func (r *epicRepository) GetCompleteHierarchy(id uuid.UUID) (*models.Epic, error) {
    var epic models.Epic
    err := r.GetDB().
        // Load user stories with ordering
        Preload("UserStories", func(db *gorm.DB) *gorm.DB {
            return db.Order("created_at ASC")
        }).
        // Load requirements for each user story
        Preload("UserStories.Requirements", func(db *gorm.DB) *gorm.DB {
            return db.Order("created_at ASC")
        }).
        // Load requirement types
        Preload("UserStories.Requirements.Type").
        // Load acceptance criteria for each user story
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

**Query Execution**:
- Total queries: 5 (not N+1)
- Query 1: Epic
- Query 2: UserStories
- Query 3: Requirements
- Query 4: RequirementTypes
- Query 5: AcceptanceCriteria

**Performance**:
- Small hierarchy (< 10 user stories): ~20-50ms
- Medium hierarchy (10-50 user stories): ~50-100ms
- Large hierarchy (50-100 user stories): ~100-200ms

## Common Patterns in Codebase

### List with Includes

Used in list endpoints to optionally include related entities:

```go
func (r *epicRepository) ListWithIncludes(filters map[string]interface{}, includes []string, orderBy string, limit, offset int) ([]models.Epic, error) {
    query := r.GetDB().Model(&models.Epic{})
    
    // Apply includes based on request
    for _, include := range includes {
        switch include {
        case "creator":
            query = query.Preload("Creator")
        case "assignee":
            query = query.Preload("Assignee")
        case "user_stories":
            query = query.Preload("UserStories")
        case "comments":
            query = query.Preload("Comments")
        }
    }
    
    // Apply filters, ordering, pagination
    for key, value := range filters {
        query = query.Where(key+" = ?", value)
    }
    
    if orderBy != "" {
        query = query.Order(orderBy)
    }
    
    if limit > 0 {
        query = query.Limit(limit)
    }
    if offset > 0 {
        query = query.Offset(offset)
    }
    
    var epics []models.Epic
    if err := query.Find(&epics).Error; err != nil {
        return nil, r.handleDBError(err)
    }
    
    return epics, nil
}
```

### Get with Preloads

Used in get-by-id endpoints to always include related entities:

```go
func (r *requirementRepository) GetByIDWithPreloads(id uuid.UUID) (*models.Requirement, error) {
    var requirement models.Requirement
    if err := r.GetDB().
        Preload("Creator").
        Preload("Assignee").
        Preload("UserStory").
        Preload("AcceptanceCriteria").
        Preload("Type").
        Where("id = ?", id).First(&requirement).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, ErrNotFound
        }
        return nil, r.handleDBError(err)
    }
    return &requirement, nil
}
```

## Performance Tips

### DO: Use Preload for Related Entities

✅ **Good** - Single query per relationship level:
```go
db.Preload("UserStories").Preload("UserStories.Requirements").Find(&epics)
// Executes 3 queries total
```

❌ **Bad** - N+1 queries:
```go
db.Find(&epics)
for _, epic := range epics {
    db.Where("epic_id = ?", epic.ID).Find(&epic.UserStories)
    // Executes 1 + N queries
}
```

### DO: Order Preloaded Relationships

✅ **Good** - Consistent ordering:
```go
Preload("UserStories", func(db *gorm.DB) *gorm.DB {
    return db.Order("created_at ASC")
})
```

### DO: Limit Preload Depth

✅ **Good** - Load only what you need:
```go
// For list view - minimal preloads
Preload("Creator").Preload("Assignee")

// For detail view - full preloads
Preload("Creator").Preload("Assignee").Preload("UserStories").Preload("Comments")
```

### DON'T: Preload Everything Always

❌ **Bad** - Unnecessary data loading:
```go
// Loading full hierarchy when you only need epic title
db.
    Preload("UserStories").
    Preload("UserStories.Requirements").
    Preload("UserStories.AcceptanceCriteria").
    Preload("Comments").
    Where("id = ?", id).First(&epic)
```

## Testing Preloads

### Verify Preload Execution

```go
func TestGetCompleteHierarchy(t *testing.T) {
    // Setup test data
    epic := createTestEpic(t, db)
    us := createTestUserStory(t, db, epic.ID)
    req := createTestRequirement(t, db, us.ID)
    ac := createTestAcceptanceCriteria(t, db, us.ID)
    
    // Execute
    result, err := repo.GetCompleteHierarchy(epic.ID)
    
    // Verify
    require.NoError(t, err)
    assert.NotNil(t, result)
    assert.Len(t, result.UserStories, 1)
    assert.Len(t, result.UserStories[0].Requirements, 1)
    assert.Len(t, result.UserStories[0].AcceptanceCriteria, 1)
}
```

### Check Query Count

Use database query logging to verify N+1 queries are avoided:

```go
// Enable query logging in tests
db = db.Debug()

// Execute query
result, err := repo.GetCompleteHierarchy(epicID)

// Check logs - should see 5 queries, not N+1
```

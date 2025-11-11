# Database Entity Relationships

## Entity Relationship Diagram

```
Epic (1) ──────< (N) UserStory
                      │
                      ├──< (N) Requirement
                      │         └── (0..1) AcceptanceCriteria (optional link)
                      │
                      └──< (N) AcceptanceCriteria (required link)
```

## Core Relationships

### Epic → UserStories
- **Type**: One-to-Many
- **Foreign Key**: `user_stories.epic_id` → `epics.id`
- **Delete Behavior**: CASCADE
- **Required**: Yes (`epic_id NOT NULL`)
- **Indexed**: Yes (`idx_user_stories_epic`)

```sql
CREATE TABLE user_stories (
    epic_id UUID NOT NULL REFERENCES epics(id) ON DELETE CASCADE,
    ...
);
```

### UserStory → Requirements
- **Type**: One-to-Many
- **Foreign Key**: `requirements.user_story_id` → `user_stories.id`
- **Delete Behavior**: CASCADE
- **Required**: Yes (`user_story_id NOT NULL`)
- **Indexed**: Yes (`idx_requirements_user_story`)

```sql
CREATE TABLE requirements (
    user_story_id UUID NOT NULL REFERENCES user_stories(id) ON DELETE CASCADE,
    ...
);
```

### UserStory → AcceptanceCriteria
- **Type**: One-to-Many
- **Foreign Key**: `acceptance_criteria.user_story_id` → `user_stories.id`
- **Delete Behavior**: CASCADE
- **Required**: Yes (`user_story_id NOT NULL`)
- **Indexed**: Yes (`idx_acceptance_criteria_user_story`)

```sql
CREATE TABLE acceptance_criteria (
    user_story_id UUID NOT NULL REFERENCES user_stories(id) ON DELETE CASCADE,
    ...
);
```

**Important**: Every AcceptanceCriteria MUST belong to a UserStory. This is the primary relationship.

### Requirement → AcceptanceCriteria (Optional Link)
- **Type**: Many-to-One (optional)
- **Foreign Key**: `requirements.acceptance_criteria_id` → `acceptance_criteria.id`
- **Delete Behavior**: SET NULL
- **Required**: No (`acceptance_criteria_id` can be NULL)
- **Indexed**: Yes (`idx_requirements_acceptance_criteria`)

```sql
CREATE TABLE requirements (
    acceptance_criteria_id UUID REFERENCES acceptance_criteria(id) ON DELETE SET NULL,
    ...
);
```

**Important**: This is an OPTIONAL link. A Requirement MAY be linked to one AcceptanceCriteria, but doesn't have to be.

## Hierarchy Implications

### For Display/Visualization

When displaying hierarchy, there are two valid approaches:

**Approach 1: Requirement-centric** (optional link used)
```
Epic
└── UserStory
    └── Requirement
        └── AcceptanceCriteria (if linked)
```

**Approach 2: UserStory-centric** (primary relationships)
```
Epic
└── UserStory
    ├── Requirements (all for this user story)
    └── AcceptanceCriteria (all for this user story)
```

The second approach is preferred because:
- Shows all AcceptanceCriteria (not just linked ones)
- Reflects actual database structure
- Simpler to implement (no need to check optional link)

### For Data Loading

Use GORM nested preloading to load complete hierarchy:

```go
epic, err := db.
    Preload("UserStories").
    Preload("UserStories.Requirements").
    Preload("UserStories.AcceptanceCriteria").
    Where("id = ?", id).
    First(&epic).Error
```

This loads:
- Epic
- All UserStories for the Epic
- All Requirements for each UserStory
- All AcceptanceCriteria for each UserStory

## Reference IDs

Each entity has a human-readable reference ID:

| Entity | Format | Example | Sequence |
|--------|--------|---------|----------|
| Epic | EP-XXX | EP-021 | epic_ref_seq |
| UserStory | US-XXX | US-064 | user_story_ref_seq |
| Requirement | REQ-XXX | REQ-089 | requirement_ref_seq |
| AcceptanceCriteria | AC-XXX | AC-021 | acceptance_criteria_ref_seq |

Reference IDs are:
- Unique across the system
- Auto-generated on creation
- Indexed for fast lookup
- Used in MCP tools and API

## Common Queries

### Get Epic with Complete Hierarchy
```go
var epic models.Epic
err := db.
    Preload("UserStories", func(db *gorm.DB) *gorm.DB {
        return db.Order("created_at ASC")
    }).
    Preload("UserStories.Requirements", func(db *gorm.DB) *gorm.DB {
        return db.Order("created_at ASC")
    }).
    Preload("UserStories.AcceptanceCriteria", func(db *gorm.DB) *gorm.DB {
        return db.Order("created_at ASC")
    }).
    Where("id = ?", epicID).
    First(&epic).Error
```

### Get UserStory with Requirements and AcceptanceCriteria
```go
var userStory models.UserStory
err := db.
    Preload("Requirements").
    Preload("AcceptanceCriteria").
    Where("id = ?", userStoryID).
    First(&userStory).Error
```

### Get Requirement with Optional AcceptanceCriteria Link
```go
var requirement models.Requirement
err := db.
    Preload("AcceptanceCriteria"). // May be nil
    Where("id = ?", requirementID).
    First(&requirement).Error
```

## Deletion Behavior

### Cascade Deletes

When you delete an Epic:
1. All UserStories are deleted (CASCADE)
2. All Requirements for those UserStories are deleted (CASCADE)
3. All AcceptanceCriteria for those UserStories are deleted (CASCADE)

When you delete a UserStory:
1. All Requirements are deleted (CASCADE)
2. All AcceptanceCriteria are deleted (CASCADE)

### SET NULL Behavior

When you delete an AcceptanceCriteria:
1. Any Requirements linked to it have `acceptance_criteria_id` set to NULL
2. The Requirements themselves are NOT deleted

## Performance Considerations

### Indexes

All foreign keys are indexed:
- `idx_user_stories_epic` on `user_stories.epic_id`
- `idx_requirements_user_story` on `requirements.user_story_id`
- `idx_acceptance_criteria_user_story` on `acceptance_criteria.user_story_id`
- `idx_requirements_acceptance_criteria` on `requirements.acceptance_criteria_id`

### Query Optimization

For hierarchical queries:
- Use nested Preload to avoid N+1 queries
- Typical hierarchy query executes 4-5 SQL queries total (not N+1)
- Add `Order()` clauses to Preload for consistent ordering

### Expected Performance

- Small hierarchy (< 10 user stories): ~20-50ms
- Medium hierarchy (10-50 user stories): ~50-100ms
- Large hierarchy (50-100 user stories): ~100-200ms

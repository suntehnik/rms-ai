---
inclusion: always
---

# Reference ID Generation Strategy

## Current State Analysis

### Entities with ReferenceID

All major entities in the system have a `ReferenceID` field that serves as a human-readable identifier:

- **Epic**: `EP-001`, `EP-002`, etc.
- **UserStory**: `US-001`, `US-002`, etc.  
- **Requirement**: `REQ-001`, `REQ-002`, etc.
- **AcceptanceCriteria**: `AC-001`, `AC-002`, etc.

### Current Implementation Inconsistencies

#### Sophisticated Approach (Epic & UserStory)
- Uses PostgreSQL advisory locks (`pg_try_advisory_xact_lock`) for atomic generation
- Different lock keys per entity type (Epic: 2147483647, UserStory: 2147483646)
- Falls back to UUID-based IDs when lock acquisition fails
- Handles PostgreSQL vs SQLite differences properly
- Thread-safe and handles concurrent operations

#### Basic Approach (Requirement & AcceptanceCriteria)  
- Uses simple `COUNT(*)` in BeforeCreate hook
- No concurrency protection
- Vulnerable to race conditions under load
- Can generate duplicate reference IDs in concurrent scenarios
- No fallback mechanism

### Problems with Current State

1. **Race Conditions**: Requirement and AcceptanceCriteria can generate duplicate IDs under concurrent load
2. **Inconsistent Behavior**: Different entities behave differently under the same conditions
3. **Maintenance Burden**: Multiple strategies to maintain and test
4. **Reliability Issues**: Basic approach can fail in production scenarios

## Desired Strategy

### Unified Approach
All entities should use the same reference ID generation strategy:

1. **PostgreSQL Production**: Use advisory locks with entity-specific lock keys
2. **SQLite Testing**: Use simple count method (acceptable for single-threaded tests)
3. **Fallback**: UUID-based reference IDs when lock acquisition fails
4. **Centralized Logic**: Extract common logic into reusable utility functions

### Lock Key Assignment
- Epic: 2147483647 (existing)
- UserStory: 2147483646 (existing)  
- Requirement: 2147483645 (new)
- AcceptanceCriteria: 2147483644 (new)

### Implementation Pattern
```go
// Generate reference ID if not set
if entity.ReferenceID == "" {
    if tx.Dialector.Name() == "postgres" {
        // Use PostgreSQL advisory lock
        lockKey := int64(ENTITY_SPECIFIC_LOCK_KEY)
        var lockAcquired bool
        if err := tx.Raw("SELECT pg_try_advisory_xact_lock(?)", lockKey).Scan(&lockAcquired).Error; err != nil {
            return fmt.Errorf("failed to acquire advisory lock: %w", err)
        }
        
        if !lockAcquired {
            // Fallback to UUID-based ID
            entity.ReferenceID = fmt.Sprintf("PREFIX-%s", uuid.New().String()[:8])
        } else {
            // Generate sequential ID
            var count int64
            if err := tx.Model(&EntityType{}).Count(&count).Error; err != nil {
                return fmt.Errorf("failed to count entities: %w", err)
            }
            entity.ReferenceID = fmt.Sprintf("PREFIX-%03d", count+1)
        }
    } else {
        // For non-PostgreSQL (SQLite in tests)
        var count int64
        if err := tx.Model(&EntityType{}).Count(&count).Error; err != nil {
            return fmt.Errorf("failed to count entities: %w", err)
        }
        entity.ReferenceID = fmt.Sprintf("PREFIX-%03d", count+1)
    }
}
```

## Benefits of Unified Strategy

1. **Consistency**: All entities behave the same way
2. **Reliability**: Proper concurrency handling for all entities
3. **Maintainability**: Single strategy to maintain and test
4. **Predictability**: Developers know what to expect across all entities
5. **Scalability**: Handles concurrent operations properly in production
# Design Document

## Overview

This design addresses the critical bug in Epic reference ID generation that causes e2e test failures due to duplicate key constraint violations. The solution involves implementing application-level reference ID generation in the Epic model's BeforeCreate hook, consistent with the patterns already established in AcceptanceCriteria and Requirement models.

## Architecture

### Current Problem
- GORM is inserting empty strings for `reference_id` field instead of letting database generate values
- Database sequences are not being used due to GORM behavior
- Multiple epics created with empty `reference_id` cause unique constraint violations
- E2E tests fail specifically in cache invalidation scenarios

### Solution Architecture
- Implement application-level reference ID generation in Epic model BeforeCreate hook
- Use database count method to determine next sequence number (consistent with existing models)
- Generate reference IDs in format "EP-001", "EP-002", etc.
- Ensure thread-safe generation to prevent race conditions

## Components and Interfaces

### Epic Model Changes
```go
// BeforeCreate hook in Epic model with PostgreSQL advisory locks for distributed safety
func (e *Epic) BeforeCreate(tx *gorm.DB) error {
    if e.ID == uuid.Nil {
        e.ID = uuid.New()
    }
    if e.Status == "" {
        e.Status = EpicStatusBacklog
    }
    
    // Generate reference ID if not set
    if e.ReferenceID == "" {
        // Use PostgreSQL advisory lock for atomic reference ID generation
        // Lock key: hash of "epic_reference_id" to avoid conflicts with other entities
        lockKey := int64(2147483647) // Fixed key for epic reference ID generation
        
        // Acquire advisory lock (automatically released at transaction end)
        var lockAcquired bool
        if err := tx.Raw("SELECT pg_try_advisory_xact_lock(?)", lockKey).Scan(&lockAcquired).Error; err != nil {
            return fmt.Errorf("failed to acquire advisory lock: %w", err)
        }
        
        if !lockAcquired {
            // If lock not acquired, fall back to UUID-based ID
            e.ReferenceID = fmt.Sprintf("EP-%s", uuid.New().String()[:8])
        } else {
            // Lock acquired, safely generate sequential reference ID
            var count int64
            if err := tx.Model(&Epic{}).Count(&count).Error; err != nil {
                return fmt.Errorf("failed to count epics: %w", err)
            }
            e.ReferenceID = fmt.Sprintf("EP-%03d", count+1)
        }
    }
    
    return nil
}
```

### Required Import Addition
- Add `fmt` import to Epic model for string formatting

### Consistency with Existing Models
- Follow the same pattern used in AcceptanceCriteria and Requirement models
- Use database count method for sequence generation
- Maintain the same reference ID format pattern

## Data Models

### Epic Model Updates
- Add reference ID generation logic to BeforeCreate hook
- Ensure ReferenceID field is properly populated before database insertion
- Maintain existing field structure and relationships

### Database Schema
- No changes required to database schema
- Existing unique constraint on reference_id remains
- Database sequences can be removed in future cleanup (not part of this fix)

## Error Handling

### Distributed Environment Synchronization
- **PostgreSQL Advisory Locks**: Use `pg_try_advisory_xact_lock()` for atomic reference ID generation
- **Transaction-Scoped Locks**: Advisory locks automatically released at transaction end
- **Non-Blocking Approach**: Use `pg_try_advisory_xact_lock()` to avoid deadlocks
- **Fallback Strategy**: If lock not acquired, fall back to UUID-based reference IDs
- **Cross-Instance Safety**: Advisory locks work across multiple application instances

### Duplicate Key Prevention
- Use PostgreSQL advisory locks to ensure atomic reference ID generation
- Single attempt with lock acquisition - no retry loops needed
- Database-level unique constraints as final safety net
- Immediate fallback to UUID-based IDs if lock not acquired

### Concurrency Strategies
1. **Advisory Lock Approach**: Use PostgreSQL's built-in distributed locking mechanism
2. **Non-Blocking Lock**: `pg_try_advisory_xact_lock()` prevents application blocking
3. **Graceful Degradation**: Immediate fallback to UUID-based IDs if lock unavailable
4. **Transaction Safety**: Locks automatically released at transaction commit/rollback

### Validation
- Ensure reference ID is not empty before database insertion
- Validate reference ID format matches expected pattern
- Maintain existing validation for other Epic fields
- Add error handling for reference ID generation failures

## Testing Strategy

### Unit Tests
- Test Epic reference ID generation in isolation
- Verify format matches expected pattern (EP-001, EP-002, etc.)
- Test BeforeCreate hook behavior with various input scenarios
- Test concurrent creation scenarios for race conditions

### Integration Tests
- Verify Epic creation through service layer generates correct reference IDs
- Test multiple Epic creation in sequence
- Ensure no duplicate reference IDs are generated

### E2E Tests
- Verify cache invalidation test passes after fix
- Test complete Epic creation workflow through API
- Ensure no database constraint violations occur

### Performance Tests
- Test concurrent Epic creation scenarios
- Verify no significant performance impact from count queries
- Ensure reference ID generation scales appropriately
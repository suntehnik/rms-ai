# Architecture Decision Record (ADR)

## ADR-001: Epic Reference ID Generation Strategy

### Status
Proposed

### Context
The current Epic model relies on database-level default values for reference ID generation using PostgreSQL sequences. However, GORM is inserting empty strings for the `reference_id` field instead of allowing the database to generate values using the configured sequence. This causes unique constraint violations when multiple epics are created, particularly evident in e2e tests where the cache invalidation test fails.

### Problem Statement
- E2E tests fail with duplicate key constraint violation on `idx_epics_reference_id`
- GORM inserts empty strings for `reference_id` instead of using database defaults
- Database sequences defined in migrations are not being utilized
- AcceptanceCriteria and Requirement models already use application-level generation successfully

### Decision
We will implement application-level reference ID generation in the Epic model's BeforeCreate hook, following the same pattern already established in AcceptanceCriteria and Requirement models.

### Rationale

#### Why Application-Level Generation?
1. **Consistency**: AcceptanceCriteria and Requirement models already use this pattern successfully
2. **GORM Compatibility**: Avoids GORM's behavior of overriding database defaults with empty values
3. **Predictability**: Application has full control over reference ID generation
4. **Testability**: Easier to test and mock reference ID generation logic

#### Why PostgreSQL Advisory Locks with Count Method?
1. **Existing Pattern**: Builds on proven count method from AcceptanceCriteria and Requirement models
2. **Distributed Safety**: PostgreSQL advisory locks provide true distributed synchronization
3. **Built-in Solution**: Uses PostgreSQL's native locking mechanism, no external dependencies
4. **Transaction Safety**: Locks automatically released at transaction end
5. **Non-Blocking**: `pg_try_advisory_xact_lock()` prevents application deadlocks
6. **Graceful Degradation**: Falls back to UUID-based IDs if lock unavailable

#### Why Not Database Sequences?
1. **GORM Limitation**: GORM doesn't properly handle PostgreSQL default values with sequences
2. **Complexity**: Would require custom GORM hooks or raw SQL for proper sequence handling
3. **Inconsistency**: Other models don't use sequences, creating maintenance burden
4. **Migration Overhead**: Would require significant changes to existing working patterns

### Implementation Details

#### Epic Model Changes
```go
func (e *Epic) BeforeCreate(tx *gorm.DB) error {
    if e.ID == uuid.Nil {
        e.ID = uuid.New()
    }
    if e.Status == "" {
        e.Status = EpicStatusBacklog
    }
    
    // Generate reference ID using PostgreSQL advisory locks for distributed safety
    if e.ReferenceID == "" {
        // Use PostgreSQL advisory lock for atomic reference ID generation
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

#### Required Changes
1. Add `fmt` import to Epic model
2. Implement BeforeCreate hook with reference ID generation
3. Follow exact pattern from AcceptanceCriteria and Requirement models
4. Add unit tests for reference ID generation

### Consequences

#### Positive
- **Immediate Fix**: Resolves e2e test failures immediately
- **Consistency**: Aligns Epic model with existing successful patterns
- **Maintainability**: Single approach across all models reduces complexity
- **Reliability**: Proven pattern reduces risk of regression

#### Negative
- **Performance**: Additional advisory lock query on each Epic creation
- **PostgreSQL Dependency**: Relies on PostgreSQL-specific advisory lock functionality
- **Fallback Behavior**: May generate UUID-based IDs when locks are contended
- **Database Sequences**: Existing sequences become unused (can be cleaned up later)

#### Neutral
- **Migration Path**: No database schema changes required
- **Backward Compatibility**: Existing epics with reference IDs remain unchanged
- **Testing**: Requires additional unit tests but improves overall test coverage

### Alternatives Considered

#### 1. Fix Database Sequence Usage
- **Pros**: Uses intended PostgreSQL feature, potentially better performance
- **Cons**: Complex GORM integration, inconsistent with existing models, higher risk
- **Decision**: Rejected due to complexity and inconsistency

#### 2. External ID Generation Service
- **Pros**: Centralized ID generation, better concurrency handling
- **Cons**: Additional complexity, external dependency, overkill for current needs
- **Decision**: Rejected as over-engineering for current requirements

#### 3. Simple Retry Logic Without Locks
- **Pros**: Simple implementation, no database-specific features
- **Cons**: Race conditions still possible, multiple database queries
- **Decision**: Rejected in favor of PostgreSQL advisory locks for better reliability

#### 4. UUID-Based Reference IDs
- **Pros**: Guaranteed uniqueness, no sequence management
- **Cons**: Not human-readable, breaks existing format expectations
- **Decision**: Used as fallback strategy when advisory locks unavailable

### Monitoring and Validation

#### Success Criteria
1. E2E cache invalidation test passes consistently
2. No duplicate key constraint violations in Epic creation
3. Reference IDs follow expected format (EP-001, EP-002, etc.)
4. Unit tests verify reference ID generation logic

#### Rollback Plan
If issues arise, the change can be easily reverted by:
1. Removing the reference ID generation logic from BeforeCreate hook
2. Reverting to empty ReferenceID field (will still cause original issue)
3. Implementing alternative solution from considered options

### Future Considerations

#### Distributed Environment Enhancements
1. **Advisory Lock Optimization**: Fine-tune lock keys and timeout strategies
2. **Database Sequences with GORM Fix**: Investigate GORM plugins for proper sequence handling
3. **Multi-Database Support**: Adapt advisory lock approach for other database systems
4. **Event-Driven ID Generation**: Use message queues for centralized ID generation service

#### Potential Improvements
1. **Performance Monitoring**: Track retry frequency and fallback usage in production
2. **Database Cleanup**: Remove unused sequences in future migration
3. **Caching Strategies**: Cache highest reference ID to reduce count queries
4. **Load Testing**: Validate behavior under high concurrent Epic creation loads

#### Migration Strategy
This change is backward compatible and requires no data migration. Existing epics retain their reference IDs, and new epics will use the application-generated IDs.

### References
- Task 27 in product-requirements-management spec
- AcceptanceCriteria model BeforeCreate implementation
- Requirement model BeforeCreate implementation
- E2E test failure logs showing duplicate key constraint violation
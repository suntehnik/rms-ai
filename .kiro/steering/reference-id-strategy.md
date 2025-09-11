---
inclusion: always
---

# Reference ID Generation Strategy

## Current Implementation (Unified Static Selection Approach)

### Entities with ReferenceID

All major entities in the system have a `ReferenceID` field that serves as a human-readable identifier:

- **Epic**: `EP-001`, `EP-002`, etc.
- **UserStory**: `US-001`, `US-002`, etc.  
- **Requirement**: `REQ-001`, `REQ-002`, etc.
- **AcceptanceCriteria**: `AC-001`, `AC-002`, etc.

### Unified Implementation Strategy

All entities now use a consistent, static selection approach with two generator types:

#### Production Generator (PostgreSQLReferenceIDGenerator)
- **Location**: `internal/models/reference_id.go`
- **Usage**: Production, integration tests, and e2e tests
- **Features**:
  - PostgreSQL advisory locks (`pg_try_advisory_xact_lock`) for atomic generation
  - Entity-specific lock keys for concurrent safety
  - UUID fallback when lock acquisition fails
  - Automatic PostgreSQL vs SQLite detection
  - Thread-safe and handles concurrent operations

#### Test Generator (TestReferenceIDGenerator)  
- **Location**: `internal/models/reference_id_test.go` (test files only)
- **Usage**: Unit tests only
- **Features**:
  - Simple internal counter without database dependencies
  - No advisory locks (optimized for single-threaded unit tests)
  - Thread-safe with mutex for concurrent test scenarios
  - Helper methods for test setup (Reset, SetCounter, GetCounter)
  - Not included in production builds

### Static Selection Benefits

1. **Compile-time Selection**: Generator type is determined at compile time, not runtime
2. **No Runtime Overhead**: No conditional logic to select generators during execution
3. **Clear Separation**: Production and test code are completely isolated
4. **Predictable Behavior**: System behavior is deterministic based on build context
5. **Type Safety**: Compile-time guarantees about which generator is available

## Implementation Architecture

### Generator Interface
```go
type ReferenceIDGenerator interface {
    Generate(tx *gorm.DB, model interface{}) (string, error)
}
```

### Production Generator Structure
```go
type PostgreSQLReferenceIDGenerator struct {
    lockKey int64  // PostgreSQL advisory lock key
    prefix  string // Entity prefix (EP, US, REQ, AC)
}
```

### Test Generator Structure (Test Files Only)
```go
type TestReferenceIDGenerator struct {
    prefix  string    // Entity prefix (EP, US, REQ, AC)
    counter int64     // Internal counter for sequential IDs
    mutex   sync.Mutex // Thread-safety for concurrent tests
}
```

### Lock Key Assignments
- **Epic**: 2147483647 (existing, maintained for backward compatibility)
- **UserStory**: 2147483646 (existing, maintained for backward compatibility)  
- **Requirement**: 2147483645 (newly assigned)
- **AcceptanceCriteria**: 2147483644 (newly assigned)

### Entity Integration Pattern
Each entity model contains a package-level generator instance:

```go
// In requirement.go
var requirementGenerator = NewPostgreSQLReferenceIDGenerator(2147483645, "REQ")

func (r *Requirement) BeforeCreate(tx *gorm.DB) error {
    // ... other logic ...
    
    if r.ReferenceID == "" {
        referenceID, err := requirementGenerator.Generate(tx, &Requirement{})
        if err != nil {
            return err
        }
        r.ReferenceID = referenceID
    }
    
    return nil
}
```

## Generator Selection Guidelines

### When to Use Production Generator (PostgreSQLReferenceIDGenerator)
- **Production environments**: Always use for live systems
- **Integration tests**: Use when testing with PostgreSQL database features
- **E2E tests**: Use when testing complete application workflows
- **Performance tests**: Use when testing concurrent operations and load scenarios

### When to Use Test Generator (TestReferenceIDGenerator)
- **Unit tests**: Use for fast, isolated business logic testing
- **Mock scenarios**: Use when you need predictable, controllable reference IDs
- **Test setup**: Use when you need to reset or manipulate reference ID sequences
- **SQLite tests**: Use when testing with in-memory SQLite databases

### Compile-time Selection Implementation

The system uses static selection based on file context:

1. **Production Code**: Only has access to `PostgreSQLReferenceIDGenerator`
2. **Test Files**: Can import and use `TestReferenceIDGenerator` from `reference_id_test.go`
3. **Build Separation**: Test generator is excluded from production builds automatically
4. **Type Safety**: Compile-time guarantees prevent using wrong generator in wrong context

## Benefits of Static Selection Approach

1. **Performance**: No runtime overhead for generator selection
2. **Clarity**: Clear separation between production and test code
3. **Safety**: Compile-time prevention of using test code in production
4. **Maintainability**: Single strategy with environment-appropriate implementations
5. **Consistency**: All entities use the same pattern with their specific generators
6. **Testability**: Fast unit tests with predictable behavior
7. **Reliability**: Production-grade concurrency handling where needed

## Migration Notes

This implementation maintains backward compatibility:
- Existing reference IDs remain unchanged
- Lock keys for Epic and UserStory are preserved
- Sequential numbering continues from current maximum values
- No API changes required
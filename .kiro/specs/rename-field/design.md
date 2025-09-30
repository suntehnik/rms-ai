# Design Document: Field Renaming from `last_modified` to `updated_at`

**Status:** `Draft`
**Date:** `2025-09-28`
**Author:** `Kiro`

## Overview

This design document outlines the comprehensive approach for renaming the `last_modified` field to `updated_at` across the Product Requirements Management System. The change aims to achieve naming consistency throughout the system, as some entities (users, comments) already use `updated_at` while core business entities (epics, user_stories, acceptance_criteria, requirements) use `last_modified`.

### Scope of Change

The renaming affects four core business entities:
- **Epics** (`epics` table)
- **User Stories** (`user_stories` table) 
- **Acceptance Criteria** (`acceptance_criteria` table)
- **Requirements** (`requirements` table)

## Architecture

### Current State Analysis

**Database Layer:**
- Tables use `last_modified TIMESTAMP WITH TIME ZONE` columns
- Automatic update triggers are configured for these columns
- Full-text search indexes reference the `last_modified` field
- Database indexes exist on `last_modified` columns for performance

**Application Layer:**
- Go models use `LastModified time.Time` fields with JSON tag `json:"last_modified"`
- GORM `BeforeUpdate` hooks update the `LastModified` field
- Repository layer queries and filters use `LastModified` field references
- Service layer business logic operates on `LastModified` fields
- API handlers serialize/deserialize using `last_modified` JSON keys

**API Layer:**
- OpenAPI specification defines `last_modified` properties in schemas
- All API responses include `last_modified` timestamps
- Client applications expect `last_modified` field in JSON responses

### Target State Design

**Database Layer:**
- Rename columns to `updated_at` while maintaining all existing constraints and indexes
- Update trigger names to reflect new column names for clarity
- Preserve all existing functionality and performance characteristics

**Application Layer:**
- Rename Go struct fields from `LastModified` to `UpdatedAt`
- Update JSON tags to `json:"updated_at"`
- Update GORM database tags to `db:"updated_at"`
- Modify `BeforeUpdate` hooks to set `UpdatedAt` field
- Update all repository queries and service logic

**API Layer:**
- Update OpenAPI schemas to use `updated_at` property names
- Ensure API responses serialize with `updated_at` keys
- Regenerate all documentation artifacts

## Components and Interfaces

### Database Migration Strategy

**Migration File Structure:**
```
migrations/000004_rename_last_modified_to_updated_at.up.sql
migrations/000004_rename_last_modified_to_updated_at.down.sql
```

**Up Migration Design:**
```sql
-- Rename columns in all affected tables
ALTER TABLE epics RENAME COLUMN last_modified TO updated_at;
ALTER TABLE user_stories RENAME COLUMN last_modified TO updated_at;
ALTER TABLE acceptance_criteria RENAME COLUMN last_modified TO updated_at;
ALTER TABLE requirements RENAME COLUMN last_modified TO updated_at;

-- Update trigger names for clarity (optional but recommended)
DROP TRIGGER IF EXISTS update_epics_last_modified ON epics;
DROP TRIGGER IF EXISTS update_user_stories_last_modified ON user_stories;
DROP TRIGGER IF EXISTS update_acceptance_criteria_last_modified ON acceptance_criteria;
DROP TRIGGER IF EXISTS update_requirements_last_modified ON requirements;

CREATE TRIGGER update_epics_updated_at BEFORE UPDATE ON epics 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_user_stories_updated_at BEFORE UPDATE ON user_stories 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_acceptance_criteria_updated_at BEFORE UPDATE ON acceptance_criteria 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_requirements_updated_at BEFORE UPDATE ON requirements 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
```

**Down Migration Design:**
```sql
-- Reverse the column renames
ALTER TABLE epics RENAME COLUMN updated_at TO last_modified;
ALTER TABLE user_stories RENAME COLUMN updated_at TO last_modified;
ALTER TABLE acceptance_criteria RENAME COLUMN updated_at TO last_modified;
ALTER TABLE requirements RENAME COLUMN updated_at TO last_modified;

-- Restore original trigger names
DROP TRIGGER IF EXISTS update_epics_updated_at ON epics;
DROP TRIGGER IF EXISTS update_user_stories_updated_at ON user_stories;
DROP TRIGGER IF EXISTS update_acceptance_criteria_updated_at ON acceptance_criteria;
DROP TRIGGER IF EXISTS update_requirements_updated_at ON requirements;

CREATE TRIGGER update_epics_last_modified BEFORE UPDATE ON epics 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_user_stories_last_modified BEFORE UPDATE ON user_stories 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_acceptance_criteria_last_modified BEFORE UPDATE ON acceptance_criteria 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_requirements_last_modified BEFORE UPDATE ON requirements 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
```

### Model Layer Changes

**Field Renaming Pattern:**
```go
// Before
LastModified time.Time `json:"last_modified" db:"last_modified"`

// After  
UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
```

**BeforeUpdate Hook Updates:**
```go
// Before
func (e *Epic) BeforeUpdate(tx *gorm.DB) error {
    e.LastModified = time.Now().UTC()
    return nil
}

// After
func (e *Epic) UpdatedAt(tx *gorm.DB) error {
    e.UpdatedAt = time.Now().UTC()
    return nil
}
```

**Custom JSON Marshaling Updates:**
The Epic model's custom `MarshalJSON` method needs updating to use the new field name in the result map.

### Repository Layer Impact

**Query Updates:**
All repository methods that reference `LastModified` fields in queries, ordering, or filtering need updates:
- Order by clauses: `ORDER BY last_modified` → `ORDER BY updated_at`
- Filter conditions involving the timestamp field
- Any raw SQL queries that reference the column

### Service Layer Impact

**Business Logic Updates:**
- Update any service methods that access or manipulate the `LastModified` field
- Ensure timestamp comparison logic uses the new `UpdatedAt` field
- Update any caching keys or logic that incorporates the field name

### API Documentation Updates

**OpenAPI Schema Changes:**
```yaml
# Before
Epic:
  required: [id, reference_id, title, status, priority, creator_id, created_at, last_modified]
  properties:
    last_modified:
      type: string
      format: date-time

# After
Epic:
  required: [id, reference_id, title, status, priority, creator_id, created_at, updated_at]
  properties:
    updated_at:
      type: string
      format: date-time
```

## Data Models

### Affected Model Structures

**Epic Model:**
- Field: `LastModified` → `UpdatedAt`
- JSON tag: `json:"last_modified"` → `json:"updated_at"`
- Database tag: `db:"last_modified"` → `db:"updated_at"`

**UserStory Model:**
- Field: `LastModified` → `UpdatedAt`
- JSON tag: `json:"last_modified"` → `json:"updated_at"`
- Database tag: `db:"last_modified"` → `db:"updated_at"`

**AcceptanceCriteria Model:**
- Field: `LastModified` → `UpdatedAt`
- JSON tag: `json:"last_modified"` → `json:"updated_at"`
- Database tag: `db:"last_modified"` → `db:"updated_at"`

**Requirement Model:**
- Field: `LastModified` → `UpdatedAt`
- JSON tag: `json:"last_modified"` → `json:"updated_at"`
- Database tag: `db:"last_modified"` → `db:"updated_at"`

### Data Consistency Considerations

**Timestamp Preservation:**
- All existing timestamp values will be preserved during column rename
- No data loss or corruption expected
- Automatic update triggers continue to function normally

**Index Preservation:**
- PostgreSQL `ALTER TABLE RENAME COLUMN` preserves all indexes automatically
- Full-text search indexes remain functional
- Performance characteristics unchanged

## Error Handling

### Migration Error Scenarios

**Database Migration Failures:**
- **Column rename conflicts:** If any application code is running during migration, it may fail with "column does not exist" errors
- **Trigger recreation failures:** If triggers cannot be dropped/recreated due to dependencies
- **Index conflicts:** Unlikely but possible if custom indexes exist with conflicting names

**Mitigation Strategies:**
- Coordinate migration deployment with application deployment
- Test migration thoroughly in staging environment
- Prepare rollback plan using down migration
- Monitor application logs during deployment

### Application Error Scenarios

**Compilation Errors:**
- Go compiler will catch all field reference errors
- Systematic approach ensures all references are updated
- Use IDE refactoring tools to minimize manual errors

**Runtime Errors:**
- JSON serialization/deserialization should work seamlessly after field updates
- Database queries will fail if any references to old column names remain
- Test coverage should catch any missed references

### API Compatibility Errors

**Breaking Change Impact:**
- Existing API clients will receive `updated_at` instead of `last_modified`
- Client applications must be updated to handle new field name
- No backward compatibility provided (clean break approach)

**Client Update Strategy:**
- Document breaking change in API changelog
- Provide migration guide for client developers
- Consider API versioning if gradual migration needed

## Testing Strategy

### Database Testing

**Migration Testing:**
- Test up migration in isolated database environment
- Verify all data preserved and accessible
- Test down migration for complete rollback capability
- Validate trigger functionality after migration

**Performance Testing:**
- Verify query performance unchanged after column rename
- Test full-text search functionality
- Validate index usage in query plans

### Application Testing

**Unit Testing:**
- Update all model tests to use new field names
- Test GORM model operations (create, update, query)
- Verify JSON serialization/deserialization
- Test custom marshaling logic

**Integration Testing:**
- Test complete CRUD operations through repository layer
- Verify service layer business logic
- Test API endpoints with new field names
- Validate search functionality

**End-to-End Testing:**
- Test complete user workflows
- Verify API responses contain correct field names
- Test timestamp update behavior
- Validate documentation accuracy

### Regression Testing

**Existing Functionality:**
- Run complete test suite to ensure no regressions
- Test all timestamp-related functionality
- Verify audit trail and change tracking
- Test backup and restore procedures

## Design Decisions and Rationales

### 1. Column Rename vs. New Column Approach

**Decision:** Use `ALTER TABLE RENAME COLUMN` approach
**Rationale:** 
- Preserves all existing data without migration
- Maintains all indexes and constraints automatically
- Simpler implementation with lower risk
- No storage overhead or data duplication
- Clean, atomic operation

**Alternative Considered:** Add new column, migrate data, drop old column
**Rejected Because:** More complex, higher risk, temporary storage overhead

### 2. Breaking Change vs. Backward Compatibility

**Decision:** Implement as breaking change without backward compatibility
**Rationale:**
- Cleaner codebase without compatibility shims
- Simpler implementation and maintenance
- Aligns with system's current development phase
- Forces consistent adoption across all clients

**Alternative Considered:** Support both field names temporarily
**Rejected Because:** Adds complexity, potential for confusion, maintenance burden

### 3. Trigger Rename Strategy

**Decision:** Rename triggers to match new column names
**Rationale:**
- Improves code maintainability and clarity
- Prevents confusion in database administration
- Minimal additional effort with clear benefit
- Follows naming consistency principle

**Alternative Considered:** Keep existing trigger names
**Rejected Because:** Inconsistent naming could cause confusion

### 4. Migration Timing Strategy

**Decision:** Single atomic migration for all tables
**Rationale:**
- Ensures consistency across all entities
- Simpler deployment coordination
- Reduces number of breaking changes
- Lower overall risk than multiple migrations

**Alternative Considered:** Separate migrations per table
**Rejected Because:** Multiple breaking changes, coordination complexity

### 5. Documentation Update Strategy

**Decision:** Regenerate all documentation from OpenAPI specification
**Rationale:**
- Ensures consistency across all documentation formats
- Leverages existing automation
- Reduces manual error potential
- Maintains single source of truth

## Implementation Considerations

### Deployment Coordination

**Application Deployment:**
- Deploy database migration first
- Deploy updated application code immediately after
- Monitor for any runtime errors
- Have rollback plan ready

**Client Communication:**
- Notify all API consumers of breaking change
- Provide clear migration timeline
- Document exact field name changes
- Offer support during transition

### Performance Impact

**Database Performance:**
- Column rename is instant operation in PostgreSQL
- No performance degradation expected
- All indexes preserved automatically
- Query plans remain optimal

**Application Performance:**
- No runtime performance impact
- JSON serialization performance unchanged
- Memory usage identical

### Monitoring and Validation

**Post-Deployment Validation:**
- Verify API responses contain `updated_at` fields
- Check database triggers are functioning
- Validate timestamp updates on entity modifications
- Confirm documentation accuracy

**Error Monitoring:**
- Monitor application logs for field reference errors
- Track API error rates for client compatibility issues
- Watch database logs for query failures
- Set up alerts for unusual error patterns

This design provides a comprehensive, low-risk approach to achieving field naming consistency across the Product Requirements Management System while maintaining all existing functionality and performance characteristics.
# Core Data Models Implementation

This document describes the implementation of the core data models and database schema for the Product Requirements Management System.

## Overview

The implementation includes:
- GORM models for all core entities
- Dual ID system (UUID + human-readable reference IDs)
- Database migrations with proper indexes
- Model validation and business logic
- Comprehensive test coverage

## Implemented Models

### 1. User Model (`internal/models/user.go`)
- **Purpose**: System users with role-based access control
- **Key Features**:
  - Role-based permissions (Administrator, User, Commenter)
  - Permission checking methods (`CanEdit()`, `CanDelete()`, etc.)
  - Secure password hash storage
- **Relationships**: One-to-many with all other entities as creator/assignee

### 2. Epic Model (`internal/models/epic.go`)
- **Purpose**: Top-level functional blocks
- **Key Features**:
  - Priority system (1-4: Critical, High, Medium, Low)
  - Status management (Backlog, Draft, In Progress, Done, Cancelled)
  - Dual ID system with auto-generated reference IDs (EP-001, EP-002, etc.)
- **Relationships**: One-to-many with UserStories

### 3. UserStory Model (`internal/models/user_story.go`)
- **Purpose**: Detailed functionality within epics
- **Key Features**:
  - Template validation for "As [role], I want [function], so that [goal]" format
  - Same priority and status system as epics
  - Reference ID format: US-001, US-002, etc.
- **Relationships**: 
  - Belongs to Epic
  - One-to-many with AcceptanceCriteria and Requirements

### 4. AcceptanceCriteria Model (`internal/models/acceptance_criteria.go`)
- **Purpose**: Readiness conditions for user stories
- **Key Features**:
  - EARS format validation (WHEN/IF...THEN...SHALL)
  - Reference ID format: AC-001, AC-002, etc.
- **Relationships**: 
  - Belongs to UserStory
  - One-to-many with Requirements (optional link)

### 5. Requirement Model (`internal/models/requirement.go`)
- **Purpose**: Detailed technical requirements
- **Key Features**:
  - Configurable requirement types
  - Status management (Draft, Active, Obsolete)
  - Optional link to acceptance criteria
  - Reference ID format: REQ-001, REQ-002, etc.
- **Relationships**:
  - Belongs to UserStory
  - Optional link to AcceptanceCriteria
  - Many-to-many relationships with other Requirements

### 6. RequirementType Model (`internal/models/requirement_type.go`)
- **Purpose**: Configurable dictionary for requirement types
- **Default Types**: Functional, Non-Functional, Business Rule, Interface, Data
- **Admin Configurable**: Yes

### 7. RelationshipType Model (`internal/models/relationship_type.go`)
- **Purpose**: Configurable dictionary for requirement relationships
- **Default Types**: depends_on, blocks, relates_to, conflicts_with, derives_from
- **Admin Configurable**: Yes

### 8. RequirementRelationship Model (`internal/models/requirement_relationship.go`)
- **Purpose**: Links between requirements with typed relationships
- **Key Features**:
  - Prevents self-referencing relationships
  - Unique constraint on source+target+type combinations
- **Relationships**: Links two Requirements via RelationshipType

### 9. Comment Model (`internal/models/comment.go`)
- **Purpose**: General and inline comments on all entities
- **Key Features**:
  - Supports threading (parent-child relationships)
  - Inline comments with text position tracking
  - Resolution status management
  - Polymorphic association with all commentable entities

## Database Schema Features

### Dual ID System
- **UUID**: Internal system identifier for relationships and APIs
- **Reference ID**: Human-readable identifier (EP-001, US-001, etc.)
- **Sequences**: PostgreSQL sequences ensure unique reference IDs
- **Functions**: Helper functions for reference ID generation

### Indexes and Performance
- **Primary Indexes**: UUID and reference ID indexes on all entities
- **Foreign Key Indexes**: All relationship columns indexed
- **Composite Indexes**: Multi-column indexes for common query patterns
- **Full-Text Search**: GIN indexes for search functionality
- **Performance Optimized**: Query patterns analyzed and optimized

### Data Integrity
- **Foreign Key Constraints**: Proper referential integrity
- **Check Constraints**: Priority validation (1-4), status validation
- **Unique Constraints**: Reference IDs, usernames, emails
- **Cascade Rules**: Proper deletion behavior (CASCADE vs RESTRICT)

## Validation and Business Logic

### Priority System
- **Values**: 1 (Critical), 2 (High), 3 (Medium), 4 (Low)
- **Validation**: Database check constraints and model validation
- **Helper Methods**: `GetPriorityString()`, `ValidatePriority()`

### Status Management
- **Epic Statuses**: Backlog, Draft, In Progress, Done, Cancelled
- **UserStory Statuses**: Same as Epic
- **Requirement Statuses**: Draft, Active, Obsolete
- **Transition Rules**: All transitions allowed by default (configurable)

### Template Validation
- **User Stories**: Validates "As...I want...so that" format
- **Acceptance Criteria**: Validates EARS format (WHEN/IF...THEN...SHALL)
- **Comments**: Validates inline comment text positions

### Role-Based Permissions
- **Administrator**: Full system access, user management, configuration
- **User**: Create, edit, delete entities and comments
- **Commenter**: View entities, create comments only

## Testing

### Unit Tests (`internal/models/models_test.go`)
- **Coverage**: All models and validation methods
- **Database**: In-memory SQLite for fast testing
- **Assertions**: Comprehensive test cases for all functionality

### Integration Tests (`internal/models/integration_test.go`)
- **Database**: Real PostgreSQL connection
- **Features**: Dual ID system, relationships, default data
- **Tag**: `integration` tag for selective running

### Verification Script (`scripts/verify_models.go`)
- **Purpose**: End-to-end verification of model functionality
- **Features**: Creates complete entity hierarchy, tests relationships
- **Usage**: `go run scripts/verify_models.go`

## Migration Files

### 000001_initial_schema.up.sql
- **Purpose**: Creates all tables, indexes, and constraints
- **Features**: Complete schema with PostgreSQL-specific features
- **Sequences**: Reference ID generation sequences
- **Triggers**: Auto-update timestamps

### 000002_add_reference_id_functions.up.sql
- **Purpose**: Adds helper functions for GORM compatibility
- **Features**: Reference ID generation functions
- **Indexes**: Additional performance indexes
- **Compatibility**: Ensures GORM models work with existing schema

## Usage Examples

### Creating Entities
```go
// Create user
user := models.User{
    Username: "john.doe",
    Email: "john@example.com",
    Role: models.RoleUser,
}
db.Create(&user)

// Create epic with auto-generated reference ID
epic := models.Epic{
    CreatorID: user.ID,
    AssigneeID: user.ID,
    Priority: models.PriorityHigh,
    Title: "User Authentication",
}
db.Create(&epic) // ReferenceID will be auto-generated (e.g., "EP-001")
```

### Querying with Relationships
```go
// Load epic with all related data
var epic models.Epic
db.Preload("UserStories.AcceptanceCriteria").
   Preload("UserStories.Requirements").
   Where("reference_id = ?", "EP-001").
   First(&epic)
```

### Finding by Either ID Type
```go
// Find by UUID
db.Where("id = ?", uuid).First(&epic)

// Find by reference ID
db.Where("reference_id = ?", "EP-001").First(&epic)
```

## Configuration

### Default Data Seeding
- **Automatic**: Default requirement and relationship types created on startup
- **Idempotent**: Safe to run multiple times
- **Configurable**: Admin can modify types after creation

### Database Connection
- **Auto-Migration**: Models automatically migrated on startup
- **Connection Pooling**: Optimized connection settings
- **Health Checks**: Database connectivity monitoring

## Next Steps

The models are now ready for:
1. **Repository Layer**: Data access patterns and query optimization
2. **Service Layer**: Business logic and validation rules
3. **API Layer**: REST endpoints and request/response handling
4. **Authentication**: JWT integration and role enforcement
5. **Testing**: Integration with actual PostgreSQL database

All models follow the requirements specified in the design document and are fully tested and verified.
# Project Structure & Organization

## Directory Layout
```
├── cmd/                    # Application entry points
│   ├── server/            # Main HTTP server
│   └── migrate/           # Database migration tool
├── internal/              # Private application code
│   ├── auth/              # Authentication & authorization
│   ├── config/            # Configuration management
│   ├── database/          # Database connection & health
│   ├── logger/            # Structured logging setup
│   ├── models/            # GORM models & business logic
│   └── server/            # HTTP server & routing
│       ├── middleware/    # HTTP middleware
│       └── routes/        # Route definitions
├── migrations/            # Database migration files
├── docs/                  # Documentation
├── scripts/               # Utility scripts
├── pkg/                   # Public/reusable packages (future)
└── bin/                   # Compiled binaries
```

## Package Organization Principles
- **cmd/**: Executable entry points, minimal logic
- **internal/**: Private packages, not importable by external projects
- **pkg/**: Public packages for potential reuse (currently empty)
- **migrations/**: Sequential SQL migration files with up/down scripts

## Model Structure
- **Base Models**: User, Epic, UserStory, AcceptanceCriteria, Requirement
- **Dictionary Models**: RequirementType, RelationshipType (configurable)
- **Relationship Models**: RequirementRelationship, Comment
- **Reference IDs**: Auto-generated human-readable IDs (EP-001, US-001, etc.)

## Database Schema Patterns
- UUID primary keys with reference ID sequences
- Soft foreign key constraints with CASCADE/RESTRICT policies
- Full-text search indexes on searchable content
- Audit fields: created_at, updated_at, last_modified
- Status enums with database constraints

## Testing Structure
- Unit tests alongside source files (*_test.go)
- Integration tests in models package
- Test utilities and mocks in testify framework
- Database tests use SQLite for speed

## Configuration Hierarchy
1. Environment variables (highest priority)
2. Default values in config package
3. Required validation for sensitive values (JWT_SECRET)
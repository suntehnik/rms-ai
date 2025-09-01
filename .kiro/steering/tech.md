# Technology Stack & Build System

## Core Technologies
- **Language**: Go 1.24.5
- **Web Framework**: Gin (HTTP router and middleware)
- **Database**: PostgreSQL 12+ (primary), Redis 6+ (caching/sessions)
- **ORM**: GORM v2 with PostgreSQL driver
- **Authentication**: JWT tokens with golang-jwt/jwt/v5
- **Logging**: Structured logging with Logrus (JSON/text formats)
- **Migrations**: golang-migrate/v4
- **Testing**: testify for assertions and mocking, testcontainers for integration tests

## Architecture Patterns
- **Clean Architecture**: Separation of concerns with internal packages
- **Configuration**: Environment-based config with sensible defaults
- **Middleware Chain**: Logger → Recovery → CORS → Routes
- **Database**: Connection pooling, health checks, graceful shutdown
- **Error Handling**: Structured errors with correlation IDs

## Build Commands
```bash
# Development
make dev              # Run with debug logging
make deps             # Install/update dependencies
make fmt              # Format code

# Building
make build            # Build server binary to bin/server
make run              # Run directly with go run

# Testing
make test             # Run all tests (unit → integration → e2e)
make test-unit        # Run unit tests (SQLite, fast)
make test-integration # Run integration tests (PostgreSQL)
make test-e2e         # Run end-to-end tests (PostgreSQL)
make test-fast        # Run only fast unit tests
make test-ci          # Run tests suitable for CI/CD
make test-coverage    # Generate coverage reports for all tests
make test-compile     # Check if tests compile
make test-debug       # Run tests in debug mode
make test-race        # Run tests with race detection

# Database
make migrate-up       # Apply migrations
make migrate-down     # Rollback last migration
make migrate-version  # Check migration status

# Docker Development
make docker-up        # Start PostgreSQL + Redis
make docker-down      # Stop containers
make dev-setup        # Full dev environment setup

# Git Commands
gh pr create           # Create pull request
```

## Code Standards
- Use structured logging with correlation IDs
- Environment-based configuration (never hardcode secrets)
- Graceful shutdown with 30s timeout
- Database transactions for multi-table operations
- UUID primary keys with human-readable reference IDs
- Full-text search indexes for searchable entities
- Use git flow and merge requests with github for each task
## Testing Strategy

### Test Types & Database Usage
- **Unit Tests**: SQLite (in-memory) - Fast, isolated, no external dependencies
- **Integration Tests**: PostgreSQL (testcontainers) - Real database features, full-text search
- **E2E Tests**: PostgreSQL (testcontainers) - Complete application stack testing

### Test Execution Order
1. **Unit Tests** - Quick validation of business logic
2. **Integration Tests** - Database integration and service layer testing  
3. **E2E Tests** - Full application workflow testing

### Database Strategy
- **SQLite for Unit Tests**: Fast setup, no external dependencies, perfect for business logic testing
- **PostgreSQL for Integration/E2E**: Real production environment, full-text search, advanced features
- **Testcontainers**: Automatic PostgreSQL container management for consistent test environments
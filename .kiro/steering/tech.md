# Technology Stack & Build System

## Core Technologies
- **Language**: Go 1.24.5
- **Web Framework**: Gin (HTTP router and middleware)
- **Database**: PostgreSQL 12+ (primary), Redis 6+ (caching/sessions)
- **ORM**: GORM v2 with PostgreSQL driver
- **Authentication**: JWT tokens with golang-jwt/jwt/v5
- **Logging**: Structured logging with Logrus (JSON/text formats)
- **Migrations**: golang-migrate/v4
- **Testing**: testify for assertions and mocking

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
make test             # Run all tests with verbose output

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
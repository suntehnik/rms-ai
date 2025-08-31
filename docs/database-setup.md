# Database Setup Guide

This document describes how to set up and manage the database for the Product Requirements Management System.

## Prerequisites

- PostgreSQL 12+ 
- Redis 6+
- Go 1.21+

## Database Configuration

The application uses environment variables for database configuration. Copy `.env.example` to `.env` and update the values:

```bash
cp .env.example .env
```

### PostgreSQL Configuration

```bash
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=requirements_db
DB_SSLMODE=disable
```

### Redis Configuration

```bash
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0
```

## Local Development Setup

### Using Docker Compose (Recommended)

Create a `docker-compose.yml` file for local development:

```yaml
version: '3.8'
services:
  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: requirements_db
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data

volumes:
  postgres_data:
  redis_data:
```

Start the services:

```bash
docker-compose up -d
```

### Manual Installation

#### PostgreSQL

**macOS (using Homebrew):**
```bash
brew install postgresql
brew services start postgresql
createdb requirements_db
```

**Ubuntu/Debian:**
```bash
sudo apt update
sudo apt install postgresql postgresql-contrib
sudo systemctl start postgresql
sudo -u postgres createdb requirements_db
```

#### Redis

**macOS (using Homebrew):**
```bash
brew install redis
brew services start redis
```

**Ubuntu/Debian:**
```bash
sudo apt update
sudo apt install redis-server
sudo systemctl start redis-server
```

## Database Migrations

The application uses golang-migrate for database schema management.

### Running Migrations

```bash
# Run all pending migrations
make migrate-up

# Or using go run directly
go run cmd/migrate/main.go -up
```

### Rolling Back Migrations

```bash
# Rollback the last migration
make migrate-down

# Or using go run directly
go run cmd/migrate/main.go -down
```

### Check Migration Status

```bash
# Check current migration version
make migrate-version

# Or using go run directly
go run cmd/migrate/main.go -version
```

### Creating New Migrations

Use the migrate CLI tool to create new migration files:

```bash
# Install migrate CLI
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Create a new migration
migrate create -ext sql -dir migrations -seq add_new_table
```

This will create two files:
- `migrations/XXXXXX_add_new_table.up.sql` - for applying the migration
- `migrations/XXXXXX_add_new_table.down.sql` - for rolling back the migration

## Database Schema

The initial schema includes the following tables:

### Core Tables
- `users` - User accounts and roles
- `epics` - Epic-level requirements
- `user_stories` - User stories within epics
- `acceptance_criteria` - Acceptance criteria for user stories
- `requirements` - Detailed requirements

### Configuration Tables
- `requirement_types` - Configurable requirement types
- `relationship_types` - Configurable relationship types
- `requirement_relationships` - Relationships between requirements

### Supporting Tables
- `comments` - Comments on all entity types

### Key Features
- **UUID Primary Keys**: All tables use UUIDs for primary keys
- **Reference IDs**: Human-readable reference IDs (EP-001, US-001, etc.)
- **Full-Text Search**: PostgreSQL full-text search indexes
- **Audit Fields**: Created/modified timestamps on all entities
- **Soft Constraints**: Application-level constraint enforcement

## Connection Pooling

The application configures PostgreSQL connection pooling with the following settings:

- **Max Idle Connections**: 10
- **Max Open Connections**: 100
- **Connection Max Lifetime**: 1 hour

These settings can be adjusted in `internal/database/database.go` based on your load requirements.

## Health Checks

The application provides several health check endpoints:

- `GET /health` - Basic health status
- `GET /health/deep` - Detailed health including database connectivity
- `GET /ready` - Kubernetes readiness probe
- `GET /live` - Kubernetes liveness probe

## Monitoring

Database connections are monitored through:

- Connection pool statistics
- Health check endpoints
- Structured logging of database operations
- Error tracking and alerting

## Backup and Recovery

For production deployments, ensure you have:

1. **Regular PostgreSQL backups**:
   ```bash
   pg_dump requirements_db > backup.sql
   ```

2. **Redis persistence** configured (RDB or AOF)

3. **Point-in-time recovery** setup for PostgreSQL

4. **Monitoring** of backup processes

## Troubleshooting

### Common Issues

1. **Connection Refused**
   - Ensure PostgreSQL/Redis services are running
   - Check host/port configuration
   - Verify firewall settings

2. **Authentication Failed**
   - Check username/password in configuration
   - Verify PostgreSQL user permissions

3. **Migration Errors**
   - Check migration file syntax
   - Ensure database user has DDL permissions
   - Review migration logs for specific errors

4. **Performance Issues**
   - Monitor connection pool usage
   - Check for long-running queries
   - Review index usage

### Logs

Database operations are logged with structured logging. Check application logs for:

- Connection establishment/failures
- Migration execution
- Health check results
- Query performance issues

### Testing

Run database tests (requires test database):

```bash
# Set up test database
createdb requirements_test_db

# Run tests
DB_NAME=requirements_test_db go test ./internal/database/... -v
```
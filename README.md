# Product Requirements Management System

A Go-based web application for managing product requirements through a hierarchical structure: Epics → User Stories → Requirements.

## Project Structure

```
├── cmd/
│   └── server/          # Application entry point
├── internal/
│   ├── config/          # Configuration management
│   ├── logger/          # Structured logging
│   └── server/          # HTTP server and routing
│       ├── middleware/  # HTTP middleware
│       └── routes/      # Route definitions
├── migrations/          # Database migration files (future)
├── pkg/                 # Reusable packages (future)
├── bin/                 # Compiled binaries
└── Makefile            # Build and development commands
```

## Getting Started

### Prerequisites

- Go 1.21 or higher
- PostgreSQL 12+
- Redis 6+

### Database Setup

See [Database Setup Guide](docs/database-setup.md) for detailed instructions on setting up PostgreSQL and Redis.

Quick start with Docker:
```bash
# Create docker-compose.yml (see database setup guide)
docker-compose up -d

# Run migrations
make migrate-up
```

### Environment Configuration

Copy the example environment file and configure your settings:

```bash
cp .env.example .env
```

Edit `.env` with your configuration values, especially:
- `JWT_SECRET`: Set a secure secret key for JWT tokens
- Database and Redis connection details

### Production Initialization

For fresh installations, use the initialization service to set up the database and create an admin user:

```bash
# Build the initialization binary
make build-init

# Set required environment variables
export DB_HOST=localhost
export DB_USER=postgres
export DB_PASSWORD=your-password
export DB_NAME=requirements_db
export JWT_SECRET=your-secret-key
export DEFAULT_ADMIN_PASSWORD=secure-admin-password

# Run initialization (only works on empty databases)
make init

# Or run directly
./bin/init
```

**Safety Note**: The initialization service will only run on completely empty databases to prevent accidental data corruption.

### Running the Application

#### Development Mode
```bash
make dev
```

#### Production Mode
```bash
make build
JWT_SECRET=your-secret ./bin/server
```

#### Using Make Commands
```bash
make deps           # Install dependencies
make build          # Build the application
make build-init     # Build initialization binary
make run            # Run the application
make init           # Run initialization service
make test           # Run tests
make clean          # Clean build artifacts
make fmt            # Format code
make migrate-up     # Run database migrations
make migrate-down   # Rollback last migration
make migrate-version # Check migration status
```

## API Endpoints

### Health Checks
- `GET /health` - Basic health check
- `GET /health/deep` - Detailed health check with database connectivity
- `GET /ready` - Readiness probe (includes database health)
- `GET /live` - Liveness probe

### API v1 (Placeholder endpoints)
- `GET /api/v1/epics` - Epics management (to be implemented)
- `GET /api/v1/user-stories` - User stories management (to be implemented)
- `GET /api/v1/requirements` - Requirements management (to be implemented)

## Configuration

The application uses environment variables for configuration:

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVER_HOST` | `0.0.0.0` | Server bind address |
| `SERVER_PORT` | `8080` | Server port |
| `JWT_SECRET` | - | JWT signing secret (required) |
| `LOG_LEVEL` | `info` | Log level (debug, info, warn, error) |
| `LOG_FORMAT` | `json` | Log format (json, text) |
| `DB_HOST` | `localhost` | Database host |
| `DB_PORT` | `5432` | Database port |
| `DB_USER` | `postgres` | Database user |
| `DB_PASSWORD` | - | Database password |
| `DB_NAME` | `requirements_db` | Database name |
| `REDIS_HOST` | `localhost` | Redis host |
| `REDIS_PORT` | `6379` | Redis port |
| `DEFAULT_ADMIN_PASSWORD` | - | Admin password for initialization |

## Logging

The application uses structured logging with logrus:
- JSON format for production
- Text format for development
- Configurable log levels
- Request correlation IDs
- Comprehensive error context

## Middleware

- **Logger**: HTTP request logging with correlation IDs
- **Recovery**: Panic recovery with structured error responses
- **CORS**: Cross-origin resource sharing configuration

## Development

### Completed Features
- ✅ Basic server setup with Gin framework
- ✅ Configuration management with environment variables
- ✅ Structured logging with correlation IDs
- ✅ Database connection management (PostgreSQL + Redis)
- ✅ Database migration system
- ✅ Health checks with database connectivity
- ✅ Connection pooling and monitoring

### Upcoming Tasks
- Database models and GORM integration
- Authentication and authorization
- Business logic for requirements management
- Search and filtering capabilities
- Comment system
- API documentation

## Testing

```bash
make test
```

Tests will be added in subsequent development tasks.
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
- PostgreSQL (for future database tasks)
- Redis (for future caching tasks)

### Environment Configuration

Copy the example environment file and configure your settings:

```bash
cp .env.example .env
```

Edit `.env` with your configuration values, especially:
- `JWT_SECRET`: Set a secure secret key for JWT tokens
- Database and Redis connection details (for future tasks)

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
make deps    # Install dependencies
make build   # Build the application
make run     # Run the application
make test    # Run tests
make clean   # Clean build artifacts
make fmt     # Format code
```

## API Endpoints

### Health Checks
- `GET /health` - Basic health check
- `GET /health/deep` - Detailed health check with dependencies
- `GET /ready` - Readiness probe
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

This is the foundation setup. Subsequent tasks will implement:
- Database models and migrations
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
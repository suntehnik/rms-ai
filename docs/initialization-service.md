# Production Initialization Service

## Overview

The production initialization service is a standalone binary that safely sets up fresh installations of the product requirements management system. It performs environment validation, database setup, migrations, and admin user creation with comprehensive safety checks to prevent execution on existing installations.

## Quick Start

### Prerequisites

- PostgreSQL 12+ database server
- Redis 6+ server (for main application)
- Go 1.24.5+ (for building from source)

### Basic Usage

1. **Build the initialization binary:**
   ```bash
   make build-init
   ```

2. **Set required environment variables:**
   ```bash
   export DB_HOST=localhost
   export DB_PORT=5432
   export DB_USER=app_user
   export DB_PASSWORD=your_secure_password
   export DB_NAME=requirements_db
   export JWT_SECRET=your_jwt_secret
   export DEFAULT_ADMIN_PASSWORD=secure_admin_password
   ```

3. **Run initialization:**
   ```bash
   ./bin/init
   ```

## Environment Configuration

### Required Environment Variables

The initialization service requires the following environment variables to be set:

#### Database Configuration
```bash
# PostgreSQL connection parameters
DB_HOST=your-database-host          # Database server hostname
DB_PORT=5432                        # Database server port
DB_USER=app_user                    # Database username
DB_PASSWORD=secure_password         # Database password
DB_NAME=requirements_db             # Database name
DB_SSLMODE=require                  # SSL mode (require for production)
```

#### Application Configuration
```bash
# JWT secret for token signing (required by main app)
JWT_SECRET=your-very-secure-jwt-secret-key

# Admin user password
DEFAULT_ADMIN_PASSWORD=secure_admin_password
```

#### Optional Configuration
```bash
# Redis configuration (used by main application)
REDIS_HOST=localhost                # Redis server hostname
REDIS_PORT=6379                     # Redis server port
REDIS_PASSWORD=redis_password       # Redis password (if required)

# Logging configuration
LOG_LEVEL=info                      # Log level (debug, info, warn, error)
LOG_FORMAT=json                     # Log format (json, text)
```

### Production Environment Example

```bash
#!/bin/bash
# production-init.env - Production environment configuration

# Database Configuration
export DB_HOST=prod-postgres.company.com
export DB_PORT=5432
export DB_USER=requirements_app
export DB_PASSWORD=super_secure_production_password
export DB_NAME=requirements_production
export DB_SSLMODE=require

# Redis Configuration
export REDIS_HOST=prod-redis.company.com
export REDIS_PORT=6379
export REDIS_PASSWORD=redis_production_password

# JWT Configuration
export JWT_SECRET=production_jwt_secret_key_minimum_32_characters

# Admin Configuration
export DEFAULT_ADMIN_PASSWORD=admin_production_password_change_after_first_login

# Logging Configuration
export LOG_LEVEL=info
export LOG_FORMAT=json
```

## Usage Examples

### Development Environment

```bash
# Load development environment
source .env.dev.local

# Build and run initialization
make build-init
./bin/init
```

### Docker Environment

```bash
# Using docker-compose for database
make docker-up

# Set environment variables
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=postgres
export DB_NAME=requirements_dev
export JWT_SECRET=dev_jwt_secret
export DEFAULT_ADMIN_PASSWORD=admin123

# Run initialization
make init
```

### Production Deployment

```bash
# Load production environment
source /etc/requirements-app/production.env

# Verify environment
./bin/init --dry-run  # (if implemented)

# Run initialization
./bin/init
```

## Safety Features

### Database Safety Checks

The initialization service includes multiple safety mechanisms:

1. **Empty Database Verification**: Checks that no data exists in critical tables
2. **Connection Validation**: Verifies database connectivity before proceeding
3. **Environment Validation**: Ensures all required configuration is present
4. **Atomic Operations**: Uses database transactions for consistency

### Protected Tables

The service checks for existing data in these tables:
- `users` - User accounts
- `epics` - Epic records
- `user_stories` - User story records
- `requirements` - Requirement records
- `acceptance_criteria` - Acceptance criteria records
- `comments` - Comment records

If any data is found, initialization will abort with an error message.

## Initialization Process

### Step-by-Step Flow

1. **Environment Validation**
   - Verify all required environment variables are set
   - Validate configuration values

2. **Database Connection**
   - Establish connection to PostgreSQL
   - Verify database accessibility

3. **Safety Checks**
   - Check if database is empty
   - Report any existing data found

4. **Migration Execution**
   - Run all pending database migrations
   - Verify schema integrity

5. **Admin User Creation**
   - Create admin user with configured password
   - Assign Administrator role
   - Log successful creation

6. **Completion**
   - Log initialization summary
   - Display next steps

### Expected Output

```
INFO[2024-01-15T10:30:00Z] Starting production initialization service
INFO[2024-01-15T10:30:00Z] Environment validation completed successfully
INFO[2024-01-15T10:30:01Z] Database connection established
INFO[2024-01-15T10:30:01Z] Safety check: Database is empty, proceeding with initialization
INFO[2024-01-15T10:30:02Z] Running database migrations...
INFO[2024-01-15T10:30:05Z] Migrations completed successfully (3 applied)
INFO[2024-01-15T10:30:05Z] Creating admin user...
INFO[2024-01-15T10:30:06Z] Admin user 'admin' created successfully
INFO[2024-01-15T10:30:06Z] Initialization completed successfully

Next steps:
1. Start the main application server
2. Login with username 'admin' and the configured password
3. Change the admin password after first login
4. Configure additional users and permissions as needed
```

## Integration with Deployment Processes

### CI/CD Pipeline Integration

#### GitLab CI Example

```yaml
# .gitlab-ci.yml
stages:
  - build
  - deploy
  - initialize

build-init:
  stage: build
  script:
    - make build-init
  artifacts:
    paths:
      - bin/init
    expire_in: 1 hour

deploy-production:
  stage: deploy
  script:
    - # Deploy application code
    - # Update database connection strings
    - # Deploy initialization binary

initialize-production:
  stage: initialize
  script:
    - source /etc/requirements-app/production.env
    - ./bin/init
  when: manual
  only:
    - main
```

#### GitHub Actions Example

```yaml
# .github/workflows/deploy.yml
name: Deploy to Production

on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Build initialization binary
        run: make build-init
        
      - name: Deploy to production
        run: |
          # Deploy application and initialization binary
          scp bin/init production-server:/opt/requirements-app/
          
      - name: Run initialization
        run: |
          ssh production-server "cd /opt/requirements-app && ./init"
        env:
          DB_HOST: ${{ secrets.DB_HOST }}
          DB_PASSWORD: ${{ secrets.DB_PASSWORD }}
          JWT_SECRET: ${{ secrets.JWT_SECRET }}
          DEFAULT_ADMIN_PASSWORD: ${{ secrets.DEFAULT_ADMIN_PASSWORD }}
```

### Docker Deployment

#### Dockerfile for Initialization

```dockerfile
# Dockerfile.init
FROM golang:1.24.5-alpine AS builder

WORKDIR /app
COPY . .
RUN make build-init

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/bin/init .
COPY --from=builder /app/migrations ./migrations

CMD ["./init"]
```

#### Docker Compose for Production

```yaml
# docker-compose.production.yml
version: '3.8'

services:
  postgres:
    image: postgres:12
    environment:
      POSTGRES_DB: requirements_production
      POSTGRES_USER: requirements_app
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data

  redis:
    image: redis:6-alpine
    command: redis-server --requirepass ${REDIS_PASSWORD}

  init:
    build:
      context: .
      dockerfile: Dockerfile.init
    environment:
      DB_HOST: postgres
      DB_PORT: 5432
      DB_USER: requirements_app
      DB_PASSWORD: ${DB_PASSWORD}
      DB_NAME: requirements_production
      JWT_SECRET: ${JWT_SECRET}
      DEFAULT_ADMIN_PASSWORD: ${DEFAULT_ADMIN_PASSWORD}
    depends_on:
      - postgres
      - redis
    restart: "no"

  app:
    build: .
    ports:
      - "8080:8080"
    environment:
      DB_HOST: postgres
      REDIS_HOST: redis
      # ... other environment variables
    depends_on:
      - init
    restart: unless-stopped

volumes:
  postgres_data:
```

### Kubernetes Deployment

#### Job for Initialization

```yaml
# k8s/init-job.yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: requirements-init
spec:
  template:
    spec:
      containers:
      - name: init
        image: requirements-app:init-latest
        env:
        - name: DB_HOST
          valueFrom:
            secretKeyRef:
              name: db-credentials
              key: host
        - name: DB_PASSWORD
          valueFrom:
            secretKeyRef:
              name: db-credentials
              key: password
        - name: JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: app-secrets
              key: jwt-secret
        - name: DEFAULT_ADMIN_PASSWORD
          valueFrom:
            secretKeyRef:
              name: app-secrets
              key: admin-password
      restartPolicy: Never
  backoffLimit: 3
```

### Ansible Playbook Integration

```yaml
# ansible/initialize.yml
---
- name: Initialize Requirements Management System
  hosts: production
  vars:
    app_dir: /opt/requirements-app
  
  tasks:
    - name: Copy initialization binary
      copy:
        src: ../bin/init
        dest: "{{ app_dir }}/init"
        mode: '0755'
    
    - name: Set environment variables
      template:
        src: production.env.j2
        dest: "{{ app_dir }}/.env"
        mode: '0600'
    
    - name: Run initialization
      shell: |
        cd {{ app_dir }}
        source .env
        ./init
      register: init_result
    
    - name: Display initialization result
      debug:
        var: init_result.stdout_lines
```

## Build System Integration

### Makefile Targets

The initialization service integrates with the existing build system:

```makefile
# Build initialization binary
build-init: ## Build initialization binary
	@echo "Building initialization service..."
	go build -ldflags="-s -w" -o bin/init cmd/init/main.go
	@echo "Initialization binary built: bin/init"

# Run initialization (requires build-init)
init: build-init ## Run initialization service
	@echo "Running initialization service..."
	./bin/init

# Clean initialization binary
clean-init: ## Remove initialization binary
	rm -f bin/init

# Add to existing clean target
clean: clean-init
	# ... existing clean commands

.PHONY: build-init init clean-init
```

### Development Workflow

```bash
# Full development setup
make dev-setup          # Start PostgreSQL + Redis
make build-init         # Build initialization binary
make init              # Run initialization
make dev               # Start development server
```

## Security Considerations

### Password Security

1. **Environment Variables**: Always use environment variables for passwords
2. **Secure Storage**: Store production passwords in secure secret management systems
3. **Password Rotation**: Change default admin password after first login
4. **Strong Passwords**: Use strong, unique passwords for production

### Database Security

1. **SSL/TLS**: Always use SSL connections in production (`DB_SSLMODE=require`)
2. **Network Security**: Restrict database access to application servers only
3. **User Privileges**: Use dedicated database user with minimal required privileges
4. **Connection Limits**: Configure appropriate connection limits

### Access Control

1. **Admin Role**: Default admin user has full system access
2. **Role Management**: Configure additional users with appropriate roles after initialization
3. **Authentication**: JWT tokens provide secure authentication
4. **Session Management**: Redis provides secure session storage

## Monitoring and Logging

### Log Output

The initialization service provides structured logging:

```json
{
  "level": "info",
  "msg": "Starting production initialization service",
  "time": "2024-01-15T10:30:00Z",
  "correlation_id": "init-abc123"
}
```

### Monitoring Integration

#### Prometheus Metrics (Future Enhancement)

```go
// Example metrics that could be added
var (
    initializationDuration = prometheus.NewHistogram(...)
    initializationSuccess = prometheus.NewCounter(...)
    initializationErrors = prometheus.NewCounterVec(...)
)
```

#### Health Checks

```bash
# Verify initialization completed successfully
curl -f http://localhost:8080/health || echo "Application not ready"
```

## Next Steps

After successful initialization:

1. **Start Main Application**: Deploy and start the main requirements management server
2. **Initial Login**: Login with username `admin` and configured password
3. **Security Setup**: Change admin password and configure additional security settings
4. **User Management**: Create additional users with appropriate roles
5. **System Configuration**: Configure system settings and preferences
6. **Data Import**: Import any existing requirements data if needed
7. **Testing**: Perform functional testing to verify system operation
8. **Monitoring**: Set up monitoring and alerting for the production system

## Support

For additional support:
- Check the troubleshooting guide: `docs/initialization-troubleshooting.md`
- Review application logs for detailed error information
- Consult the main application documentation for post-initialization setup
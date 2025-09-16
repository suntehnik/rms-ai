# Environment Configuration Examples

## Overview

This document provides comprehensive examples of environment variable configurations for different deployment scenarios of the production initialization service.

## Development Environment

### Local Development with Docker

```bash
# .env.dev.local - Development environment with Docker PostgreSQL
# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=requirements_dev
DB_SSLMODE=disable

# Redis Configuration
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

# JWT Configuration
JWT_SECRET=dev_jwt_secret_key_for_development_only

# Admin Configuration
DEFAULT_ADMIN_PASSWORD=admin123

# Logging Configuration
LOG_LEVEL=debug
LOG_FORMAT=text
```

### Local Development with External Database

```bash
# .env.dev.external - Development with external PostgreSQL
# Database Configuration
DB_HOST=dev-postgres.company.local
DB_PORT=5432
DB_USER=dev_user
DB_PASSWORD=dev_password_123
DB_NAME=requirements_dev
DB_SSLMODE=prefer

# Redis Configuration
REDIS_HOST=dev-redis.company.local
REDIS_PORT=6379
REDIS_PASSWORD=dev_redis_pass

# JWT Configuration
JWT_SECRET=development_jwt_secret_minimum_32_chars

# Admin Configuration
DEFAULT_ADMIN_PASSWORD=dev_admin_password

# Logging Configuration
LOG_LEVEL=debug
LOG_FORMAT=json
```

## Testing Environments

### Unit Testing Environment

```bash
# .env.test - Unit testing configuration
# Database Configuration (SQLite for unit tests)
DB_HOST=localhost
DB_PORT=5432
DB_USER=test_user
DB_PASSWORD=test_password
DB_NAME=requirements_test
DB_SSLMODE=disable

# Redis Configuration
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

# JWT Configuration
JWT_SECRET=test_jwt_secret_for_unit_tests_only

# Admin Configuration
DEFAULT_ADMIN_PASSWORD=test_admin_123

# Logging Configuration
LOG_LEVEL=warn
LOG_FORMAT=json
```

### Integration Testing Environment

```bash
# .env.integration - Integration testing with testcontainers
# Database Configuration
DB_HOST=localhost
DB_PORT=5433  # Different port to avoid conflicts
DB_USER=integration_user
DB_PASSWORD=integration_password_123
DB_NAME=requirements_integration
DB_SSLMODE=disable

# Redis Configuration
REDIS_HOST=localhost
REDIS_PORT=6380  # Different port to avoid conflicts
REDIS_PASSWORD=integration_redis_pass

# JWT Configuration
JWT_SECRET=integration_jwt_secret_minimum_32_characters

# Admin Configuration
DEFAULT_ADMIN_PASSWORD=integration_admin_password

# Logging Configuration
LOG_LEVEL=info
LOG_FORMAT=json
```

### End-to-End Testing Environment

```bash
# .env.e2e - End-to-end testing configuration
# Database Configuration
DB_HOST=e2e-postgres
DB_PORT=5432
DB_USER=e2e_user
DB_PASSWORD=e2e_secure_password_123
DB_NAME=requirements_e2e
DB_SSLMODE=require

# Redis Configuration
REDIS_HOST=e2e-redis
REDIS_PORT=6379
REDIS_PASSWORD=e2e_redis_secure_password

# JWT Configuration
JWT_SECRET=e2e_jwt_secret_for_testing_minimum_32_chars

# Admin Configuration
DEFAULT_ADMIN_PASSWORD=e2e_admin_secure_password

# Logging Configuration
LOG_LEVEL=info
LOG_FORMAT=json
```

## Staging Environment

### Staging Environment Configuration

```bash
# .env.staging - Staging environment configuration
# Database Configuration
DB_HOST=staging-postgres.company.com
DB_PORT=5432
DB_USER=requirements_staging
DB_PASSWORD=staging_secure_password_change_me
DB_NAME=requirements_staging
DB_SSLMODE=require

# Redis Configuration
REDIS_HOST=staging-redis.company.com
REDIS_PORT=6379
REDIS_PASSWORD=staging_redis_secure_password

# JWT Configuration
JWT_SECRET=staging_jwt_secret_key_minimum_32_characters_secure

# Admin Configuration
DEFAULT_ADMIN_PASSWORD=staging_admin_secure_password_change_after_init

# Logging Configuration
LOG_LEVEL=info
LOG_FORMAT=json

# Optional: Additional staging-specific configuration
ENVIRONMENT=staging
```

## Production Environments

### Single Server Production

```bash
# .env.production - Single server production configuration
# Database Configuration
DB_HOST=prod-postgres.company.com
DB_PORT=5432
DB_USER=requirements_prod
DB_PASSWORD=super_secure_production_password_2024
DB_NAME=requirements_production
DB_SSLMODE=require

# Redis Configuration
REDIS_HOST=prod-redis.company.com
REDIS_PORT=6379
REDIS_PASSWORD=redis_production_secure_password_2024

# JWT Configuration
JWT_SECRET=production_jwt_secret_key_very_secure_minimum_32_characters

# Admin Configuration
DEFAULT_ADMIN_PASSWORD=admin_production_password_change_immediately_after_init

# Logging Configuration
LOG_LEVEL=info
LOG_FORMAT=json

# Production-specific configuration
ENVIRONMENT=production
```

### High Availability Production

```bash
# .env.production.ha - High availability production configuration
# Database Configuration (with connection pooling)
DB_HOST=prod-postgres-cluster.company.com
DB_PORT=5432
DB_USER=requirements_prod_ha
DB_PASSWORD=ha_production_password_very_secure_2024
DB_NAME=requirements_production
DB_SSLMODE=require
DB_MAX_CONNECTIONS=100
DB_MAX_IDLE_CONNECTIONS=10
DB_CONNECTION_LIFETIME=3600

# Redis Configuration (with clustering)
REDIS_HOST=prod-redis-cluster.company.com
REDIS_PORT=6379
REDIS_PASSWORD=redis_cluster_password_very_secure_2024
REDIS_CLUSTER_MODE=true

# JWT Configuration
JWT_SECRET=ha_production_jwt_secret_very_secure_minimum_32_characters

# Admin Configuration
DEFAULT_ADMIN_PASSWORD=ha_admin_production_password_change_immediately

# Logging Configuration
LOG_LEVEL=info
LOG_FORMAT=json

# High availability specific configuration
ENVIRONMENT=production
HA_MODE=true
```

### Multi-Region Production

```bash
# .env.production.us-east - US East region production
# Database Configuration
DB_HOST=us-east-postgres.company.com
DB_PORT=5432
DB_USER=requirements_prod_us_east
DB_PASSWORD=us_east_production_password_secure_2024
DB_NAME=requirements_production_us_east
DB_SSLMODE=require

# Redis Configuration
REDIS_HOST=us-east-redis.company.com
REDIS_PORT=6379
REDIS_PASSWORD=us_east_redis_password_secure_2024

# JWT Configuration
JWT_SECRET=us_east_jwt_secret_production_minimum_32_characters

# Admin Configuration
DEFAULT_ADMIN_PASSWORD=us_east_admin_password_change_after_init

# Logging Configuration
LOG_LEVEL=info
LOG_FORMAT=json

# Region-specific configuration
ENVIRONMENT=production
REGION=us-east-1
```

## Cloud Provider Configurations

### AWS Configuration

```bash
# .env.aws - AWS deployment configuration
# Database Configuration (RDS)
DB_HOST=requirements-prod.cluster-xyz.us-east-1.rds.amazonaws.com
DB_PORT=5432
DB_USER=requirements_app
DB_PASSWORD=aws_rds_password_secure_2024
DB_NAME=requirements_production
DB_SSLMODE=require

# Redis Configuration (ElastiCache)
REDIS_HOST=requirements-prod.abc123.cache.amazonaws.com
REDIS_PORT=6379
REDIS_PASSWORD=  # ElastiCache may not require password with VPC

# JWT Configuration (from AWS Secrets Manager)
JWT_SECRET=aws_jwt_secret_from_secrets_manager_minimum_32_chars

# Admin Configuration (from AWS Secrets Manager)
DEFAULT_ADMIN_PASSWORD=aws_admin_password_from_secrets_manager

# Logging Configuration
LOG_LEVEL=info
LOG_FORMAT=json

# AWS-specific configuration
ENVIRONMENT=production
CLOUD_PROVIDER=aws
AWS_REGION=us-east-1
```

### Google Cloud Platform Configuration

```bash
# .env.gcp - GCP deployment configuration
# Database Configuration (Cloud SQL)
DB_HOST=10.0.0.3  # Private IP for Cloud SQL
DB_PORT=5432
DB_USER=requirements-app
DB_PASSWORD=gcp_cloudsql_password_secure_2024
DB_NAME=requirements-production
DB_SSLMODE=require

# Redis Configuration (Memorystore)
REDIS_HOST=10.0.0.4  # Private IP for Memorystore
REDIS_PORT=6379
REDIS_PASSWORD=gcp_memorystore_password_secure_2024

# JWT Configuration (from Secret Manager)
JWT_SECRET=gcp_jwt_secret_from_secret_manager_minimum_32_chars

# Admin Configuration (from Secret Manager)
DEFAULT_ADMIN_PASSWORD=gcp_admin_password_from_secret_manager

# Logging Configuration
LOG_LEVEL=info
LOG_FORMAT=json

# GCP-specific configuration
ENVIRONMENT=production
CLOUD_PROVIDER=gcp
GCP_PROJECT=requirements-prod-project
GCP_REGION=us-central1
```

### Azure Configuration

```bash
# .env.azure - Azure deployment configuration
# Database Configuration (Azure Database for PostgreSQL)
DB_HOST=requirements-prod.postgres.database.azure.com
DB_PORT=5432
DB_USER=requirements_app@requirements-prod
DB_PASSWORD=azure_postgres_password_secure_2024
DB_NAME=requirements_production
DB_SSLMODE=require

# Redis Configuration (Azure Cache for Redis)
REDIS_HOST=requirements-prod.redis.cache.windows.net
REDIS_PORT=6380  # SSL port for Azure Redis
REDIS_PASSWORD=azure_redis_password_secure_2024

# JWT Configuration (from Azure Key Vault)
JWT_SECRET=azure_jwt_secret_from_keyvault_minimum_32_chars

# Admin Configuration (from Azure Key Vault)
DEFAULT_ADMIN_PASSWORD=azure_admin_password_from_keyvault

# Logging Configuration
LOG_LEVEL=info
LOG_FORMAT=json

# Azure-specific configuration
ENVIRONMENT=production
CLOUD_PROVIDER=azure
AZURE_SUBSCRIPTION_ID=12345678-1234-1234-1234-123456789012
AZURE_RESOURCE_GROUP=requirements-prod-rg
AZURE_REGION=eastus
```

## Container Orchestration Configurations

### Docker Compose Configuration

```yaml
# docker-compose.production.yml
version: '3.8'

services:
  init:
    image: requirements-app:init-latest
    environment:
      # Database Configuration
      DB_HOST: postgres
      DB_PORT: 5432
      DB_USER: requirements_app
      DB_PASSWORD: ${DB_PASSWORD}
      DB_NAME: requirements_production
      DB_SSLMODE: require
      
      # Redis Configuration
      REDIS_HOST: redis
      REDIS_PORT: 6379
      REDIS_PASSWORD: ${REDIS_PASSWORD}
      
      # JWT Configuration
      JWT_SECRET: ${JWT_SECRET}
      
      # Admin Configuration
      DEFAULT_ADMIN_PASSWORD: ${DEFAULT_ADMIN_PASSWORD}
      
      # Logging Configuration
      LOG_LEVEL: info
      LOG_FORMAT: json
    depends_on:
      - postgres
      - redis
    restart: "no"

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
    volumes:
      - redis_data:/data

volumes:
  postgres_data:
  redis_data:
```

### Kubernetes Configuration

```yaml
# k8s/init-configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: init-config
data:
  DB_HOST: "postgres-service"
  DB_PORT: "5432"
  DB_USER: "requirements_app"
  DB_NAME: "requirements_production"
  DB_SSLMODE: "require"
  REDIS_HOST: "redis-service"
  REDIS_PORT: "6379"
  LOG_LEVEL: "info"
  LOG_FORMAT: "json"
  ENVIRONMENT: "production"

---
# k8s/init-secret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: init-secrets
type: Opaque
stringData:
  DB_PASSWORD: "kubernetes_postgres_password_secure_2024"
  REDIS_PASSWORD: "kubernetes_redis_password_secure_2024"
  JWT_SECRET: "kubernetes_jwt_secret_minimum_32_characters_secure"
  DEFAULT_ADMIN_PASSWORD: "kubernetes_admin_password_change_after_init"

---
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
        envFrom:
        - configMapRef:
            name: init-config
        - secretRef:
            name: init-secrets
      restartPolicy: Never
  backoffLimit: 3
```

## Security Best Practices

### Environment Variable Security

```bash
# Use secure methods to set sensitive environment variables

# 1. From secure file (with restricted permissions)
chmod 600 /etc/requirements-app/secrets.env
source /etc/requirements-app/secrets.env

# 2. From external secret management
export JWT_SECRET=$(aws secretsmanager get-secret-value --secret-id prod/jwt-secret --query SecretString --output text)
export DEFAULT_ADMIN_PASSWORD=$(aws secretsmanager get-secret-value --secret-id prod/admin-password --query SecretString --output text)

# 3. From HashiCorp Vault
export JWT_SECRET=$(vault kv get -field=jwt_secret secret/requirements-app/prod)
export DEFAULT_ADMIN_PASSWORD=$(vault kv get -field=admin_password secret/requirements-app/prod)
```

### Password Generation Examples

```bash
# Generate secure passwords for production

# Generate JWT secret (base64 encoded)
JWT_SECRET=$(openssl rand -base64 32)

# Generate admin password (alphanumeric with symbols)
DEFAULT_ADMIN_PASSWORD=$(openssl rand -base64 24 | tr -d "=+/" | cut -c1-20)

# Generate database password (strong random password)
DB_PASSWORD=$(openssl rand -base64 32 | tr -d "=+/" | cut -c1-25)

# Generate Redis password
REDIS_PASSWORD=$(openssl rand -base64 24 | tr -d "=+/" | cut -c1-20)
```

## Validation Scripts

### Environment Validation Script

```bash
#!/bin/bash
# validate-env.sh - Validate environment configuration

set -e

echo "Validating environment configuration..."

# Required variables
required_vars=(
    "DB_HOST"
    "DB_PORT"
    "DB_USER"
    "DB_PASSWORD"
    "DB_NAME"
    "JWT_SECRET"
    "DEFAULT_ADMIN_PASSWORD"
)

# Optional variables with defaults
optional_vars=(
    "DB_SSLMODE:prefer"
    "REDIS_HOST:localhost"
    "REDIS_PORT:6379"
    "LOG_LEVEL:info"
    "LOG_FORMAT:json"
)

# Validate required variables
for var in "${required_vars[@]}"; do
    if [ -z "${!var}" ]; then
        echo "ERROR: Required variable $var is not set"
        exit 1
    fi
    echo "✓ $var is set"
done

# Check optional variables and set defaults
for var_default in "${optional_vars[@]}"; do
    var="${var_default%:*}"
    default="${var_default#*:}"
    if [ -z "${!var}" ]; then
        export "$var"="$default"
        echo "ℹ $var not set, using default: $default"
    else
        echo "✓ $var is set: ${!var}"
    fi
done

# Validate JWT secret length
if [ ${#JWT_SECRET} -lt 32 ]; then
    echo "ERROR: JWT_SECRET must be at least 32 characters long"
    exit 1
fi
echo "✓ JWT_SECRET length is sufficient"

# Validate database port is numeric
if ! [[ "$DB_PORT" =~ ^[0-9]+$ ]]; then
    echo "ERROR: DB_PORT must be numeric"
    exit 1
fi
echo "✓ DB_PORT is numeric"

# Test database connectivity (optional)
if command -v pg_isready &> /dev/null; then
    if pg_isready -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" &> /dev/null; then
        echo "✓ Database is reachable"
    else
        echo "⚠ Database connectivity test failed (this may be expected)"
    fi
fi

echo "Environment validation completed successfully"
```

### Configuration Template Generator

```bash
#!/bin/bash
# generate-config.sh - Generate configuration template

ENVIRONMENT=${1:-development}

cat > ".env.${ENVIRONMENT}" << EOF
# ${ENVIRONMENT^} Environment Configuration
# Generated on $(date)

# Database Configuration
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5432}
DB_USER=${DB_USER:-requirements_${ENVIRONMENT}}
DB_PASSWORD=${DB_PASSWORD:-CHANGE_ME}
DB_NAME=${DB_NAME:-requirements_${ENVIRONMENT}}
DB_SSLMODE=${DB_SSLMODE:-prefer}

# Redis Configuration
REDIS_HOST=${REDIS_HOST:-localhost}
REDIS_PORT=${REDIS_PORT:-6379}
REDIS_PASSWORD=${REDIS_PASSWORD:-}

# JWT Configuration
JWT_SECRET=${JWT_SECRET:-GENERATE_SECURE_32_CHAR_SECRET}

# Admin Configuration
DEFAULT_ADMIN_PASSWORD=${DEFAULT_ADMIN_PASSWORD:-CHANGE_ME}

# Logging Configuration
LOG_LEVEL=${LOG_LEVEL:-info}
LOG_FORMAT=${LOG_FORMAT:-json}

# Environment
ENVIRONMENT=${ENVIRONMENT}
EOF

echo "Configuration template generated: .env.${ENVIRONMENT}"
echo "Please update the CHANGE_ME and GENERATE_* values with secure values"
```

## Usage Examples

### Loading Environment Files

```bash
# Load environment file
source .env.production

# Verify environment is loaded
./validate-env.sh

# Run initialization
./bin/init
```

### Environment-Specific Initialization

```bash
# Development
source .env.dev.local && ./bin/init

# Staging
source .env.staging && ./bin/init

# Production
source .env.production && ./bin/init
```

### Automated Deployment Script

```bash
#!/bin/bash
# deploy-init.sh - Automated deployment with environment selection

ENVIRONMENT=${1:-production}
ENV_FILE=".env.${ENVIRONMENT}"

if [ ! -f "$ENV_FILE" ]; then
    echo "Error: Environment file $ENV_FILE not found"
    exit 1
fi

echo "Loading environment: $ENVIRONMENT"
source "$ENV_FILE"

echo "Validating environment..."
./validate-env.sh

echo "Building initialization binary..."
make build-init

echo "Running initialization..."
./bin/init

echo "Initialization completed for environment: $ENVIRONMENT"
```

This comprehensive set of environment configuration examples covers all major deployment scenarios and provides the foundation for secure, reliable initialization service deployment across different environments.
# Production Deployment Guide

This guide covers deploying the Product Requirements Management System to production using Docker.

## Prerequisites

- Docker Engine 20.10+
- Docker Compose 2.0+
- At least 2GB RAM
- At least 10GB disk space
- Linux/macOS/Windows with WSL2

## Quick Start

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd product-requirements-management
   ```

2. **Configure environment**
   ```bash
   cp .env.prod.template .env.prod
   # Edit .env.prod with your production values
   ```

3. **Deploy**
   ```bash
   ./scripts/deploy-prod.sh
   ```

## Detailed Setup

### 1. Environment Configuration

Copy the template and configure your production environment:

```bash
cp .env.prod.template .env.prod
```

Edit `.env.prod` with secure values:

```bash
# Database Configuration
DB_PASSWORD=your_secure_database_password_here

# Redis Configuration (optional)
REDIS_PASSWORD=your_redis_password_here

# JWT Configuration - MUST be changed
JWT_SECRET=your_jwt_secret_key_here_minimum_32_characters

# Default Admin Password - Change after first login
DEFAULT_ADMIN_PASSWORD=your_secure_admin_password_here

# MCP Personal Access Token
PAT_TOKEN=mcp_pat_your_secure_token_here
```

### 2. Security Considerations

**Critical Security Steps:**

1. **Change default passwords** - Never use default passwords in production
2. **Generate secure JWT secret** - Use at least 32 random characters
3. **Use strong database passwords** - Include special characters, numbers, and mixed case
4. **Enable SSL/TLS** - Configure HTTPS for production (see SSL section)
5. **Configure firewall** - Only expose necessary ports
6. **Regular updates** - Keep Docker images and dependencies updated

### 3. SSL/HTTPS Configuration (Recommended)

For production with HTTPS:

1. **Obtain SSL certificates** (Let's Encrypt, commercial CA, or self-signed)

2. **Place certificates in nginx/ssl/**
   ```bash
   mkdir -p nginx/ssl
   cp your-certificate.crt nginx/ssl/certificate.crt
   cp your-private-key.key nginx/ssl/private.key
   ```

3. **Update nginx configuration**
   - Uncomment HTTPS server block in `nginx/nginx.conf`
   - Update `server_name` with your domain
   - Configure SSL paths

4. **Update environment variables**
   ```bash
   SSL_CERT_PATH=/etc/nginx/ssl/certificate.crt
   SSL_KEY_PATH=/etc/nginx/ssl/private.key
   ```

### 4. Database Configuration

#### Using Built-in PostgreSQL (Default)
The docker-compose includes PostgreSQL. Data is persisted in Docker volumes.

#### Using External Database
To use an external PostgreSQL database:

1. **Update .env.prod**
   ```bash
   DB_HOST=your_external_db_host
   DB_PORT=5432
   DB_USER=your_db_user
   DB_PASSWORD=your_db_password
   DB_NAME=your_db_name
   DB_SSLMODE=require
   ```

2. **Remove PostgreSQL service** from docker-compose.prod.yml

3. **Ensure database exists** and user has necessary permissions

### 5. Deployment Options

#### Option A: Automated Deployment (Recommended)
```bash
./scripts/deploy-prod.sh
```

#### Option B: Manual Deployment
```bash
# Build images
docker-compose -f docker-compose.prod.yml build

# Start services
docker-compose -f docker-compose.prod.yml up -d

# Run migrations
docker exec requirements-app ./migrate -up
```

#### Option C: Without Nginx
If you have an external reverse proxy:

```bash
# Start only app and databases
docker-compose -f docker-compose.prod.yml up -d app postgres redis
```

## Management Commands

### Service Management
```bash
# View status
./scripts/deploy-prod.sh status

# View logs
./scripts/deploy-prod.sh logs

# Stop services
./scripts/deploy-prod.sh stop

# Restart services
./scripts/deploy-prod.sh restart

# Create backup
./scripts/deploy-prod.sh backup
```

### Manual Commands
```bash
# View running containers
docker ps

# View logs
docker-compose -f docker-compose.prod.yml logs -f

# Execute commands in container
docker exec -it requirements-app ./migrate -version

# Access database
docker exec -it requirements-postgres psql -U postgres -d requirements_db

# Access Redis
docker exec -it requirements-redis redis-cli
```

## Monitoring and Health Checks

### Health Endpoints
- **Application**: `http://localhost:8080/ready`
- **Liveness**: `http://localhost:8080/live`

### Container Health
```bash
# Check container health
docker ps

# View detailed health status
docker inspect requirements-app | grep -A 10 Health
```

### Logs
```bash
# Application logs
docker logs requirements-app

# Database logs
docker logs requirements-postgres

# Nginx logs
docker logs requirements-nginx

# All logs
docker-compose -f docker-compose.prod.yml logs
```

## Backup and Recovery

### Automated Backup
```bash
./scripts/deploy-prod.sh backup
```

Creates backups in `backups/YYYYMMDD_HHMMSS/`:
- `database.sql` - Database dump
- `postgres-data.tar.gz` - PostgreSQL data volume
- `redis-data.tar.gz` - Redis data volume

### Manual Backup
```bash
# Database backup
docker exec requirements-postgres pg_dump -U postgres requirements_db > backup.sql

# Volume backup
docker run --rm -v requirements-management_postgres-data:/data -v $(pwd):/backup alpine tar czf /backup/postgres-backup.tar.gz -C /data .
```

### Recovery
```bash
# Restore database
docker exec -i requirements-postgres psql -U postgres requirements_db < backup.sql

# Restore volume
docker run --rm -v requirements-management_postgres-data:/data -v $(pwd):/backup alpine tar xzf /backup/postgres-backup.tar.gz -C /data
```

## Performance Tuning

### Database Optimization
The PostgreSQL container is configured with optimized settings in `scripts/init-db.sql`:
- Connection limits
- Memory settings
- Checkpoint configuration

### Application Optimization
Production environment variables:
- `GIN_MODE=release` - Optimized Gin framework
- `LOG_LEVEL=info` - Reduced logging
- `COMPRESSION_ENABLED=true` - Response compression
- `CACHE_ENABLED=true` - Redis caching

### Nginx Optimization
- Gzip compression enabled
- Connection keep-alive
- Rate limiting configured
- Static file caching

## Scaling

### Horizontal Scaling
To run multiple application instances:

1. **Use external database and Redis**
2. **Update docker-compose.prod.yml**
   ```yaml
   app:
     deploy:
       replicas: 3
   ```
3. **Configure load balancer** (Nginx upstream or external)

### Vertical Scaling
Adjust resource limits in docker-compose.prod.yml:
```yaml
app:
  deploy:
    resources:
      limits:
        cpus: '2.0'
        memory: 2G
      reservations:
        cpus: '1.0'
        memory: 1G
```

## Troubleshooting

### Common Issues

#### Application Won't Start
```bash
# Check logs
docker logs requirements-app

# Check environment variables
docker exec requirements-app env | grep -E "(DB_|REDIS_|JWT_)"

# Test database connection
docker exec requirements-app ./migrate -version
```

#### Database Connection Issues
```bash
# Check PostgreSQL status
docker exec requirements-postgres pg_isready -U postgres

# Check network connectivity
docker exec requirements-app ping postgres

# Verify credentials
docker exec requirements-postgres psql -U postgres -d requirements_db -c "SELECT version();"
```

#### Performance Issues
```bash
# Check resource usage
docker stats

# Check database performance
docker exec requirements-postgres psql -U postgres -d requirements_db -c "SELECT * FROM pg_stat_activity;"

# Check Redis performance
docker exec requirements-redis redis-cli info stats
```

### Log Analysis
```bash
# Application errors
docker logs requirements-app 2>&1 | grep -i error

# Database slow queries
docker logs requirements-postgres 2>&1 | grep -i "slow"

# Nginx access patterns
docker logs requirements-nginx | tail -100
```

## Security Checklist

- [ ] Changed all default passwords
- [ ] Generated secure JWT secret (32+ characters)
- [ ] Configured HTTPS/SSL certificates
- [ ] Enabled firewall rules
- [ ] Set up regular backups
- [ ] Configured log rotation
- [ ] Updated Docker images to latest versions
- [ ] Reviewed and configured security headers
- [ ] Set up monitoring and alerting
- [ ] Documented access credentials securely

## Maintenance

### Regular Tasks
- **Weekly**: Check logs for errors
- **Monthly**: Update Docker images
- **Quarterly**: Review security settings
- **As needed**: Backup before major changes

### Updates
```bash
# Pull latest images
docker-compose -f docker-compose.prod.yml pull

# Rebuild and restart
./scripts/deploy-prod.sh
```

## Support

For issues and questions:
1. Check logs first
2. Review this documentation
3. Check GitHub issues
4. Contact system administrator

## Architecture Overview

```
Internet → Nginx (80/443) → Application (8080) → PostgreSQL (5432)
                                              → Redis (6379)
```

**Components:**
- **Nginx**: Reverse proxy, SSL termination, rate limiting
- **Application**: Go web server with Gin framework
- **PostgreSQL**: Primary database with full-text search
- **Redis**: Caching and session storage

**Data Flow:**
1. Client requests → Nginx
2. Nginx → Application (with security headers)
3. Application → PostgreSQL/Redis
4. Response → Client (compressed, cached)

This deployment provides a production-ready, scalable, and secure environment for the Product Requirements Management System.
# Initialization Service Troubleshooting Guide

## Common Issues and Solutions

### Configuration Issues

#### Missing Environment Variables

**Error:**
```
FATAL[2024-01-15T10:30:00Z] Configuration validation failed: missing required environment variable: DB_HOST
```

**Solution:**
Ensure all required environment variables are set:
```bash
# Check current environment
env | grep -E "(DB_|JWT_|DEFAULT_ADMIN_)"

# Set missing variables
export DB_HOST=your-database-host
export DB_PORT=5432
export DB_USER=app_user
export DB_PASSWORD=your_password
export DB_NAME=requirements_db
export JWT_SECRET=your_jwt_secret
export DEFAULT_ADMIN_PASSWORD=admin_password
```

#### Invalid JWT Secret

**Error:**
```
FATAL[2024-01-15T10:30:00Z] JWT_SECRET must be at least 32 characters long
```

**Solution:**
Generate a secure JWT secret:
```bash
# Generate a secure random string
export JWT_SECRET=$(openssl rand -base64 32)

# Or use a manually created secure string
export JWT_SECRET="your-very-secure-jwt-secret-key-minimum-32-characters"
```

#### Missing Admin Password

**Error:**
```
FATAL[2024-01-15T10:30:00Z] DEFAULT_ADMIN_PASSWORD environment variable is required
```

**Solution:**
Set a secure admin password:
```bash
export DEFAULT_ADMIN_PASSWORD="secure_admin_password_123"
```

### Database Connection Issues

#### Database Connection Failed

**Error:**
```
ERROR[2024-01-15T10:30:01Z] Failed to connect to database: dial tcp 127.0.0.1:5432: connect: connection refused
```

**Possible Causes and Solutions:**

1. **Database server not running:**
   ```bash
   # Check if PostgreSQL is running
   sudo systemctl status postgresql
   
   # Start PostgreSQL if needed
   sudo systemctl start postgresql
   
   # For Docker:
   docker ps | grep postgres
   make docker-up  # If using project's docker-compose
   ```

2. **Wrong connection parameters:**
   ```bash
   # Verify connection parameters
   psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME
   
   # Test connection manually
   pg_isready -h $DB_HOST -p $DB_PORT
   ```

3. **Network/firewall issues:**
   ```bash
   # Test network connectivity
   telnet $DB_HOST $DB_PORT
   
   # Check firewall rules
   sudo ufw status  # Ubuntu
   sudo firewall-cmd --list-all  # CentOS/RHEL
   ```

#### Authentication Failed

**Error:**
```
ERROR[2024-01-15T10:30:01Z] Failed to connect to database: pq: password authentication failed for user "app_user"
```

**Solution:**
1. **Verify credentials:**
   ```bash
   # Test credentials manually
   psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME
   ```

2. **Check PostgreSQL authentication configuration:**
   ```bash
   # Check pg_hba.conf
   sudo cat /etc/postgresql/12/main/pg_hba.conf
   
   # Ensure appropriate authentication method is configured
   # Example line for password authentication:
   # host    all             all             0.0.0.0/0               md5
   ```

#### Database Does Not Exist

**Error:**
```
ERROR[2024-01-15T10:30:01Z] Failed to connect to database: pq: database "requirements_db" does not exist
```

**Solution:**
Create the database:
```bash
# Connect as superuser and create database
sudo -u postgres psql
CREATE DATABASE requirements_db;
CREATE USER app_user WITH PASSWORD 'your_password';
GRANT ALL PRIVILEGES ON DATABASE requirements_db TO app_user;
\q

# Or using command line
sudo -u postgres createdb requirements_db
sudo -u postgres psql -c "CREATE USER app_user WITH PASSWORD 'your_password';"
sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE requirements_db TO app_user;"
```

#### SSL Connection Issues

**Error:**
```
ERROR[2024-01-15T10:30:01Z] Failed to connect to database: pq: SSL is not enabled on the server
```

**Solutions:**

1. **For development (disable SSL):**
   ```bash
   export DB_SSLMODE=disable
   ```

2. **For production (enable SSL on server):**
   ```bash
   # Edit postgresql.conf
   ssl = on
   ssl_cert_file = 'server.crt'
   ssl_key_file = 'server.key'
   
   # Restart PostgreSQL
   sudo systemctl restart postgresql
   ```

### Safety Check Failures

#### Database Not Empty

**Error:**
```
ERROR[2024-01-15T10:30:02Z] Safety check failed: Database is not empty. Found data in tables: users (5 records), epics (12 records)
```

**Explanation:**
The initialization service detected existing data and aborted to prevent data corruption.

**Solutions:**

1. **For development (clear database):**
   ```bash
   # Connect to database and clear data
   psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME
   
   -- Clear all data (CAUTION: This deletes all data!)
   TRUNCATE users, epics, user_stories, requirements, acceptance_criteria, comments CASCADE;
   
   -- Or drop and recreate database
   DROP DATABASE requirements_db;
   CREATE DATABASE requirements_db;
   ```

2. **For production (verify correct database):**
   ```bash
   # Ensure you're connecting to the correct database
   echo "Current database: $DB_NAME"
   
   # If wrong database, update environment variable
   export DB_NAME=requirements_production_new
   ```

3. **For migration scenarios:**
   ```bash
   # If you need to migrate existing data, use a different approach
   # This initialization service is only for fresh installations
   # Consider using database migration tools for existing installations
   ```

### Migration Issues

#### Migration Execution Failed

**Error:**
```
ERROR[2024-01-15T10:30:03Z] Migration failed: migration 000001_initial_schema.up.sql failed: syntax error at or near "CONSTRAINT"
```

**Solutions:**

1. **Check migration files:**
   ```bash
   # Verify migration files exist and are readable
   ls -la migrations/
   
   # Check migration file syntax
   cat migrations/000001_initial_schema.up.sql
   ```

2. **Check database permissions:**
   ```bash
   # Ensure user has CREATE privileges
   psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME
   \du  -- List users and privileges
   
   -- Grant necessary privileges if needed
   GRANT CREATE ON DATABASE requirements_db TO app_user;
   ```

3. **Manual migration verification:**
   ```bash
   # Test migration manually
   psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f migrations/000001_initial_schema.up.sql
   ```

#### Migration Version Conflict

**Error:**
```
ERROR[2024-01-15T10:30:03Z] Migration version conflict: expected version 0, found version 3
```

**Solution:**
Check migration state and reset if necessary:
```bash
# Check current migration version
psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME
SELECT * FROM schema_migrations;

# If database should be empty, drop migration table
DROP TABLE IF EXISTS schema_migrations;
```

### Admin User Creation Issues

#### User Creation Failed

**Error:**
```
ERROR[2024-01-15T10:30:05Z] Failed to create admin user: duplicate key value violates unique constraint "users_username_key"
```

**Solution:**
Check if admin user already exists:
```bash
psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME
SELECT username, role FROM users WHERE username = 'admin';

-- If user exists but initialization should proceed, remove the user
DELETE FROM users WHERE username = 'admin';
```

#### Password Hashing Failed

**Error:**
```
ERROR[2024-01-15T10:30:05Z] Failed to hash admin password: bcrypt: password length exceeds 72 bytes
```

**Solution:**
Use a shorter password:
```bash
# Password should be less than 72 characters
export DEFAULT_ADMIN_PASSWORD="secure_admin_pass_123"
```

### Permission Issues

#### Binary Not Executable

**Error:**
```
bash: ./bin/init: Permission denied
```

**Solution:**
```bash
# Make binary executable
chmod +x bin/init

# Or rebuild with correct permissions
make build-init
```

#### File System Permissions

**Error:**
```
ERROR[2024-01-15T10:30:00Z] Failed to read migration files: open migrations/000001_initial_schema.up.sql: permission denied
```

**Solution:**
```bash
# Fix migration file permissions
chmod 644 migrations/*.sql

# Ensure directory is readable
chmod 755 migrations/
```

### Resource Issues

#### Out of Memory

**Error:**
```
ERROR[2024-01-15T10:30:03Z] Failed to execute migration: out of memory
```

**Solutions:**

1. **Increase available memory:**
   ```bash
   # Check current memory usage
   free -h
   
   # For Docker containers, increase memory limit
   docker run --memory=2g your-init-container
   ```

2. **Optimize database configuration:**
   ```bash
   # Reduce PostgreSQL memory usage temporarily
   # Edit postgresql.conf
   shared_buffers = 128MB
   work_mem = 4MB
   ```

#### Disk Space Issues

**Error:**
```
ERROR[2024-01-15T10:30:03Z] Failed to write to database: no space left on device
```

**Solution:**
```bash
# Check disk space
df -h

# Clean up unnecessary files
docker system prune  # For Docker environments
sudo apt autoremove  # For Ubuntu systems

# Move database to larger partition if needed
```

### Network Issues

#### Timeout Errors

**Error:**
```
ERROR[2024-01-15T10:30:10Z] Database connection timeout after 30 seconds
```

**Solutions:**

1. **Increase timeout:**
   ```bash
   # Add connection timeout to database URL
   export DB_TIMEOUT=60  # If supported by configuration
   ```

2. **Check network latency:**
   ```bash
   # Test network latency to database
   ping $DB_HOST
   
   # Test database response time
   time pg_isready -h $DB_HOST -p $DB_PORT
   ```

#### DNS Resolution Issues

**Error:**
```
ERROR[2024-01-15T10:30:01Z] Failed to resolve database host: no such host
```

**Solution:**
```bash
# Test DNS resolution
nslookup $DB_HOST
dig $DB_HOST

# Use IP address instead of hostname
export DB_HOST=192.168.1.100

# Check /etc/hosts for local resolution
cat /etc/hosts
```

## Debugging Techniques

### Enable Debug Logging

```bash
# Set debug log level for detailed output
export LOG_LEVEL=debug

# Run initialization with debug logging
./bin/init
```

### Database Connection Testing

```bash
# Test database connection independently
psql "postgresql://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME?sslmode=$DB_SSLMODE"

# Test with connection string
export DATABASE_URL="postgresql://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME?sslmode=$DB_SSLMODE"
psql $DATABASE_URL
```

### Environment Variable Verification

```bash
# Create a script to verify all required variables
cat > check-env.sh << 'EOF'
#!/bin/bash
required_vars=(
    "DB_HOST"
    "DB_PORT" 
    "DB_USER"
    "DB_PASSWORD"
    "DB_NAME"
    "JWT_SECRET"
    "DEFAULT_ADMIN_PASSWORD"
)

for var in "${required_vars[@]}"; do
    if [ -z "${!var}" ]; then
        echo "ERROR: $var is not set"
        exit 1
    else
        echo "OK: $var is set"
    fi
done
echo "All required environment variables are set"
EOF

chmod +x check-env.sh
./check-env.sh
```

### Manual Step-by-Step Testing

```bash
# Test each component individually

# 1. Test configuration loading
go run cmd/init/main.go --validate-config  # If implemented

# 2. Test database connection
go run -c "
package main
import (
    \"internal/database\"
    \"internal/config\"
)
func main() {
    cfg := config.Load()
    db, err := database.Connect(cfg)
    if err != nil {
        panic(err)
    }
    println(\"Database connection successful\")
}"

# 3. Test migration files
ls -la migrations/
head -n 10 migrations/*.sql
```

## Getting Help

### Log Analysis

When reporting issues, include:

1. **Full error logs** with timestamps
2. **Environment configuration** (without sensitive values)
3. **Database version** and configuration
4. **System information** (OS, architecture)
5. **Steps to reproduce** the issue

### Useful Commands for Support

```bash
# System information
uname -a
go version
psql --version

# Environment (sanitized)
env | grep -E "(DB_|LOG_|JWT_)" | sed 's/=.*/=***/'

# Database status
pg_isready -h $DB_HOST -p $DB_PORT
psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "SELECT version();"

# Application logs (last 50 lines)
tail -n 50 /var/log/requirements-init.log
```

### Contact Information

For additional support:
- Check the main documentation: `docs/initialization-service.md`
- Review application architecture: `docs/database-setup.md`
- Consult security guidelines: `docs/security-guide.md`
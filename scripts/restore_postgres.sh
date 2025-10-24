#!/bin/bash

# PostgreSQL Database Restore Script with .env support
# Usage: ./restore_postgres.sh <backup_file> [env_file_path]

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to load .env file
load_env() {
    local env_file="$1"
    
    if [[ -f "$env_file" ]]; then
        print_info "Loading environment variables from: $env_file"
        
        # Export variables from .env file, ignoring comments and empty lines
        set -a  # automatically export all variables
        source <(grep -v '^#' "$env_file" | grep -v '^$' | sed 's/^/export /')
        set +a  # stop automatically exporting
        
        print_success "Environment variables loaded successfully"
    else
        print_error "Environment file not found: $env_file"
        exit 1
    fi
}

# Function to validate required environment variables
validate_env() {
    local required_vars=("DB_HOST" "DB_USER" "DB_NAME")
    local missing_vars=()
    
    for var in "${required_vars[@]}"; do
        if [[ -z "${!var}" ]]; then
            missing_vars+=("$var")
        fi
    done
    
    if [[ ${#missing_vars[@]} -gt 0 ]]; then
        print_error "Missing required environment variables:"
        for var in "${missing_vars[@]}"; do
            echo "  - $var"
        done
        print_info "Please ensure these variables are set in your .env file"
        exit 1
    fi
    
    # Set PGPASSWORD if DB_PASSWORD is provided
    if [[ -n "$DB_PASSWORD" ]]; then
        export PGPASSWORD="$DB_PASSWORD"
        print_info "Database password loaded from DB_PASSWORD variable"
    fi
}

# Function to find PostgreSQL container
find_postgres_container() {
    local container_name=""
    
    # Try to find container by name patterns
    if [[ -n "$CONTAINER_NAME" ]]; then
        container_name="$CONTAINER_NAME"
    elif [[ -n "$DB_HOST" ]] && [[ "$DB_HOST" != "localhost" ]] && [[ "$DB_HOST" != "127.0.0.1" ]]; then
        container_name="$DB_HOST"
    else
        # Try common PostgreSQL container names
        local common_names=("postgres" "postgresql" "db" "database")
        for name in "${common_names[@]}"; do
            if docker ps --format "table {{.Names}}" | grep -q "^${name}$"; then
                container_name="$name"
                break
            fi
        done
    fi
    
    # Verify container exists and is running
    if [[ -n "$container_name" ]] && docker ps --format "table {{.Names}}" | grep -q "^${container_name}$"; then
        echo "$container_name"
    else
        print_error "PostgreSQL container not found or not running"
        print_info "Available containers:"
        docker ps --format "table {{.Names}}\t{{.Image}}\t{{.Status}}"
        exit 1
    fi
}

# Function to detect backup format
detect_backup_format() {
    local backup_file="$1"
    
    if [[ -d "$backup_file" ]]; then
        echo "directory"
    elif [[ "$backup_file" == *.sql.gz ]]; then
        echo "sql_compressed"
    elif [[ "$backup_file" == *.sql ]]; then
        echo "sql"
    elif [[ "$backup_file" == *.dump ]]; then
        echo "custom"
    else
        print_error "Cannot detect backup format for: $backup_file"
        exit 1
    fi
}

# Function to restore database
restore_backup() {
    local container_name="$1"
    local backup_file="$2"
    local backup_format="$3"
    
    print_info "Restoring database: $DB_NAME"
    print_info "Container: $container_name"
    print_info "User: $DB_USER"
    print_info "Backup file: $backup_file"
    print_info "Format: $backup_format"
    
    # Ask for confirmation
    if [[ "${FORCE_RESTORE:-false}" != "true" ]]; then
        print_warning "This will overwrite the existing database: $DB_NAME"
        read -p "Are you sure you want to continue? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            print_info "Restore cancelled by user"
            exit 0
        fi
    fi
    
    case "$backup_format" in
        "sql")
            if [[ -n "$DB_PASSWORD" ]]; then
                docker exec -e PGPASSWORD="$DB_PASSWORD" -i "$container_name" psql \
                    -h "${DB_HOST:-localhost}" \
                    -p "${DB_PORT:-5432}" \
                    -U "$DB_USER" \
                    -d "$DB_NAME" \
                    < "$backup_file"
            else
                docker exec -i "$container_name" psql \
                    -h "${DB_HOST:-localhost}" \
                    -p "${DB_PORT:-5432}" \
                    -U "$DB_USER" \
                    -d "$DB_NAME" \
                    < "$backup_file"
            fi
            ;;
        "sql_compressed")
            if [[ -n "$DB_PASSWORD" ]]; then
                gunzip -c "$backup_file" | docker exec -e PGPASSWORD="$DB_PASSWORD" -i "$container_name" psql \
                    -h "${DB_HOST:-localhost}" \
                    -p "${DB_PORT:-5432}" \
                    -U "$DB_USER" \
                    -d "$DB_NAME"
            else
                gunzip -c "$backup_file" | docker exec -i "$container_name" psql \
                    -h "${DB_HOST:-localhost}" \
                    -p "${DB_PORT:-5432}" \
                    -U "$DB_USER" \
                    -d "$DB_NAME"
            fi
            ;;
        "custom")
            # Copy backup file to container first
            local temp_file="/tmp/restore_$(basename "$backup_file")"
            docker cp "$backup_file" "${container_name}:${temp_file}"
            
            if [[ -n "$DB_PASSWORD" ]]; then
                docker exec -e PGPASSWORD="$DB_PASSWORD" "$container_name" pg_restore \
                    -h "${DB_HOST:-localhost}" \
                    -p "${DB_PORT:-5432}" \
                    -U "$DB_USER" \
                    -d "$DB_NAME" \
                    --clean \
                    --if-exists \
                    --verbose \
                    "$temp_file"
            else
                docker exec "$container_name" pg_restore \
                    -h "${DB_HOST:-localhost}" \
                    -p "${DB_PORT:-5432}" \
                    -U "$DB_USER" \
                    -d "$DB_NAME" \
                    --clean \
                    --if-exists \
                    --verbose \
                    "$temp_file"
            fi
            
            # Cleanup temp file
            docker exec "$container_name" rm -f "$temp_file"
            ;;
        "directory")
            # Copy directory to container first
            local temp_dir="/tmp/restore_$(date +%s)"
            docker cp "$backup_file" "${container_name}:${temp_dir}"
            
            if [[ -n "$DB_PASSWORD" ]]; then
                docker exec -e PGPASSWORD="$DB_PASSWORD" "$container_name" pg_restore \
                    -h "${DB_HOST:-localhost}" \
                    -p "${DB_PORT:-5432}" \
                    -U "$DB_USER" \
                    -d "$DB_NAME" \
                    --clean \
                    --if-exists \
                    --verbose \
                    "$temp_dir"
            else
                docker exec "$container_name" pg_restore \
                    -h "${DB_HOST:-localhost}" \
                    -p "${DB_PORT:-5432}" \
                    -U "$DB_USER" \
                    -d "$DB_NAME" \
                    --clean \
                    --if-exists \
                    --verbose \
                    "$temp_dir"
            fi
            
            # Cleanup temp directory
            docker exec "$container_name" rm -rf "$temp_dir"
            ;;
        *)
            print_error "Unsupported backup format: $backup_format"
            exit 1
            ;;
    esac
}

# Function to verify restore
verify_restore() {
    local container_name="$1"
    
    print_info "Verifying database restore..."
    
    # Check if database exists and has tables
    local table_count
    if [[ -n "$DB_PASSWORD" ]]; then
        table_count=$(docker exec -e PGPASSWORD="$DB_PASSWORD" "$container_name" psql \
            -h "${DB_HOST:-localhost}" \
            -p "${DB_PORT:-5432}" \
            -U "$DB_USER" \
            -d "$DB_NAME" \
            -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public';" | tr -d ' ')
    else
        table_count=$(docker exec "$container_name" psql \
            -h "${DB_HOST:-localhost}" \
            -p "${DB_PORT:-5432}" \
            -U "$DB_USER" \
            -d "$DB_NAME" \
            -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public';" | tr -d ' ')
    fi
    
    if [[ "$table_count" -gt 0 ]]; then
        print_success "Database restored successfully with $table_count tables"
        return 0
    else
        print_warning "Database restored but no tables found in public schema"
        return 1
    fi
}

# Main function
main() {
    local backup_file="$1"
    local env_file="${2:-.env}"
    
    # Check if backup file is provided
    if [[ -z "$backup_file" ]]; then
        print_error "Backup file not specified"
        print_info "Usage: $0 <backup_file> [env_file_path]"
        exit 1
    fi
    
    # Check if backup file exists
    if [[ ! -f "$backup_file" ]] && [[ ! -d "$backup_file" ]]; then
        print_error "Backup file not found: $backup_file"
        exit 1
    fi
    
    print_info "Starting PostgreSQL restore process..."
    
    # Load environment variables
    load_env "$env_file"
    
    # Validate required variables
    validate_env
    
    # Find PostgreSQL container
    local container_name
    container_name=$(find_postgres_container)
    print_success "Found PostgreSQL container: $container_name"
    
    # Detect backup format
    local backup_format
    backup_format=$(detect_backup_format "$backup_file")
    
    # Restore backup
    restore_backup "$container_name" "$backup_file" "$backup_format"
    
    # Verify restore
    if verify_restore "$container_name"; then
        print_success "Restore process completed successfully!"
    else
        print_warning "Restore completed but verification had issues"
    fi
}

# Show usage if help is requested
if [[ "$1" == "-h" ]] || [[ "$1" == "--help" ]]; then
    cat << EOF
PostgreSQL Database Restore Script

Usage: $0 <backup_file> [env_file_path]

Arguments:
  backup_file      Path to backup file (.sql, .sql.gz, .dump, or directory)
  env_file_path    Path to .env file (default: .env)

Required Environment Variables:
  DB_HOST          Database host (container name or localhost)
  DB_USER          Database username
  DB_NAME          Database name

Optional Environment Variables:
  DB_PORT          Database port (default: 5432)
  DB_PASSWORD      Database password (if authentication required)
  CONTAINER_NAME   Specific container name to use
  FORCE_RESTORE    Skip confirmation prompt: true/false (default: false)

Supported Backup Formats:
  .sql             Plain SQL format
  .sql.gz          Compressed SQL format
  .dump            PostgreSQL custom format
  directory        PostgreSQL directory format

Examples:
  $0 backup.sql
  $0 backup.sql.gz .env.production
  $0 backup.dump
  $0 backup_dir/

EOF
    exit 0
fi

# Run main function
main "$@"
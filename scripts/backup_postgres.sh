#!/bin/bash

# PostgreSQL Database Backup Script with .env support
# Usage: ./backup_postgres.sh [env_file_path]

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

# Function to create backup directory
create_backup_dir() {
    local backup_dir="${BACKUP_DIR:-./backups}"
    
    if [[ ! -d "$backup_dir" ]]; then
        print_info "Creating backup directory: $backup_dir"
        mkdir -p "$backup_dir"
    fi
    
    echo "$backup_dir"
}

# Function to generate backup filename
generate_backup_filename() {
    local db_name="$1"
    local backup_dir="$2"
    local timestamp=$(date +%Y%m%d_%H%M%S)
    local hostname="${DB_HOST:-localhost}"
    
    echo "${backup_dir}/backup_${hostname}_${db_name}_${timestamp}"
}

# Function to create database backup
create_backup() {
    local container_name="$1"
    local backup_file="$2"
    local backup_format="${BACKUP_FORMAT:-sql}"
    
    print_info "Creating backup of database: $DB_NAME"
    print_info "Container: $container_name"
    print_info "User: $DB_USER"
    print_info "Format: $backup_format"
    
    case "$backup_format" in
        "sql")
            if [[ -n "$DB_PASSWORD" ]]; then
                docker exec -e PGPASSWORD="$DB_PASSWORD" -t "$container_name" pg_dump \
                    -h "${DB_HOST:-localhost}" \
                    -p "${DB_PORT:-5432}" \
                    -U "$DB_USER" \
                    -d "$DB_NAME" \
                    --clean \
                    --if-exists \
                    --verbose \
                    > "${backup_file}.sql"
            else
                docker exec -t "$container_name" pg_dump \
                    -h "${DB_HOST:-localhost}" \
                    -p "${DB_PORT:-5432}" \
                    -U "$DB_USER" \
                    -d "$DB_NAME" \
                    --clean \
                    --if-exists \
                    --verbose \
                    > "${backup_file}.sql"
            fi
            
            # Compress the backup
            if [[ "${COMPRESS_BACKUP:-true}" == "true" ]]; then
                print_info "Compressing backup..."
                gzip "${backup_file}.sql"
                backup_file="${backup_file}.sql.gz"
            else
                backup_file="${backup_file}.sql"
            fi
            ;;
        "custom")
            if [[ -n "$DB_PASSWORD" ]]; then
                docker exec -e PGPASSWORD="$DB_PASSWORD" -t "$container_name" pg_dump \
                    -h "${DB_HOST:-localhost}" \
                    -p "${DB_PORT:-5432}" \
                    -U "$DB_USER" \
                    -d "$DB_NAME" \
                    --format=custom \
                    --clean \
                    --if-exists \
                    --verbose \
                    > "${backup_file}.dump"
            else
                docker exec -t "$container_name" pg_dump \
                    -h "${DB_HOST:-localhost}" \
                    -p "${DB_PORT:-5432}" \
                    -U "$DB_USER" \
                    -d "$DB_NAME" \
                    --format=custom \
                    --clean \
                    --if-exists \
                    --verbose \
                    > "${backup_file}.dump"
            fi
            backup_file="${backup_file}.dump"
            ;;
        "directory")
            # Create directory format backup inside container, then copy out
            local temp_dir="/tmp/pg_backup_$(date +%s)"
            docker exec "$container_name" mkdir -p "$temp_dir"
            
            if [[ -n "$DB_PASSWORD" ]]; then
                docker exec -e PGPASSWORD="$DB_PASSWORD" "$container_name" pg_dump \
                    -h "${DB_HOST:-localhost}" \
                    -p "${DB_PORT:-5432}" \
                    -U "$DB_USER" \
                    -d "$DB_NAME" \
                    --format=directory \
                    --clean \
                    --if-exists \
                    --verbose \
                    --file="$temp_dir"
            else
                docker exec "$container_name" pg_dump \
                    -h "${DB_HOST:-localhost}" \
                    -p "${DB_PORT:-5432}" \
                    -U "$DB_USER" \
                    -d "$DB_NAME" \
                    --format=directory \
                    --clean \
                    --if-exists \
                    --verbose \
                    --file="$temp_dir"
            fi
            
            # Copy directory from container
            docker cp "${container_name}:${temp_dir}" "${backup_file}_dir"
            docker exec "$container_name" rm -rf "$temp_dir"
            backup_file="${backup_file}_dir"
            ;;
        *)
            print_error "Unsupported backup format: $backup_format"
            print_info "Supported formats: sql, custom, directory"
            exit 1
            ;;
    esac
    
    echo "$backup_file"
}

# Function to verify backup
verify_backup() {
    local backup_file="$1"
    
    if [[ -f "$backup_file" ]] || [[ -d "$backup_file" ]]; then
        local size=$(du -sh "$backup_file" | cut -f1)
        print_success "Backup created successfully: $backup_file"
        print_info "Backup size: $size"
        return 0
    else
        print_error "Backup verification failed: $backup_file not found"
        return 1
    fi
}

# Function to cleanup old backups
cleanup_old_backups() {
    local backup_dir="$1"
    local retention_days="${BACKUP_RETENTION_DAYS:-7}"
    
    if [[ "$retention_days" -gt 0 ]]; then
        print_info "Cleaning up backups older than $retention_days days..."
        find "$backup_dir" -name "backup_*" -type f -mtime +$retention_days -delete
        find "$backup_dir" -name "backup_*_dir" -type d -mtime +$retention_days -exec rm -rf {} +
        print_success "Cleanup completed"
    fi
}

# Main function
main() {
    local env_file="${1:-.env}"
    
    print_info "Starting PostgreSQL backup process..."
    
    # Load environment variables
    load_env "$env_file"
    
    # Validate required variables
    validate_env
    
    # Find PostgreSQL container
    local container_name
    container_name=$(find_postgres_container)
    print_success "Found PostgreSQL container: $container_name"
    
    # Create backup directory
    local backup_dir
    backup_dir=$(create_backup_dir)
    
    # Generate backup filename
    local backup_file
    backup_file=$(generate_backup_filename "$DB_NAME" "$backup_dir")
    
    # Create backup
    local final_backup_file
    final_backup_file=$(create_backup "$container_name" "$backup_file")
    
    # Verify backup
    if verify_backup "$final_backup_file"; then
        # Cleanup old backups
        cleanup_old_backups "$backup_dir"
        
        print_success "Backup process completed successfully!"
        print_info "Backup location: $final_backup_file"
    else
        print_error "Backup process failed!"
        exit 1
    fi
}

# Show usage if help is requested
if [[ "$1" == "-h" ]] || [[ "$1" == "--help" ]]; then
    cat << EOF
PostgreSQL Database Backup Script

Usage: $0 [env_file_path]

Arguments:
  env_file_path    Path to .env file (default: .env)

Required Environment Variables:
  DB_HOST          Database host (container name or localhost)
  DB_USER          Database username
  DB_NAME          Database name

Optional Environment Variables:
  DB_PORT              Database port (default: 5432)
  DB_PASSWORD          Database password (if authentication required)
  CONTAINER_NAME       Specific container name to use
  BACKUP_DIR           Backup directory (default: ./backups)
  BACKUP_FORMAT        Backup format: sql, custom, directory (default: sql)
  COMPRESS_BACKUP      Compress SQL backups: true/false (default: true)
  BACKUP_RETENTION_DAYS Number of days to keep backups (default: 7, 0 to disable)

Example .env file:
  DB_HOST=postgres
  DB_USER=myuser
  DB_NAME=mydatabase
  DB_PORT=5432
  DB_PASSWORD=mypassword
  BACKUP_FORMAT=custom
  COMPRESS_BACKUP=true
  BACKUP_RETENTION_DAYS=14

EOF
    exit 0
fi

# Run main function
main "$@"
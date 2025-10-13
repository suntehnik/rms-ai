#!/bin/bash

# Production deployment script for Product Requirements Management System
# This script handles the complete deployment process with safety checks

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
COMPOSE_FILE="docker-compose.prod.yml"
ENV_FILE=".env.prod"
BACKUP_DIR="backups/$(date +%Y%m%d_%H%M%S)"

# Functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

check_prerequisites() {
    log_info "Checking prerequisites..."
    
    # Check if Docker is installed and running
    if ! command -v docker &> /dev/null; then
        log_error "Docker is not installed"
        exit 1
    fi
    
    if ! docker info &> /dev/null; then
        log_error "Docker is not running"
        exit 1
    fi
    
    # Check if Docker Compose is available
    if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
        log_error "Docker Compose is not installed"
        exit 1
    fi
    
    # Check if environment file exists
    if [[ ! -f "$ENV_FILE" ]]; then
        log_error "Environment file $ENV_FILE not found"
        log_info "Please copy .env.prod.template to $ENV_FILE and configure it"
        exit 1
    fi
    
    log_success "Prerequisites check passed"
}

backup_data() {
    log_info "Creating backup..."
    
    mkdir -p "$BACKUP_DIR"
    
    # Backup database if container is running
    if docker ps | grep -q requirements-postgres; then
        log_info "Backing up database..."
        docker exec requirements-postgres pg_dump -U postgres requirements_db > "$BACKUP_DIR/database.sql"
        log_success "Database backup created: $BACKUP_DIR/database.sql"
    else
        log_warning "PostgreSQL container not running, skipping database backup"
    fi
    
    # Backup volumes
    if docker volume ls | grep -q requirements-management_postgres-data; then
        log_info "Backing up PostgreSQL data volume..."
        docker run --rm -v requirements-management_postgres-data:/data -v "$(pwd)/$BACKUP_DIR":/backup alpine tar czf /backup/postgres-data.tar.gz -C /data .
        log_success "PostgreSQL data backup created: $BACKUP_DIR/postgres-data.tar.gz"
    fi
    
    if docker volume ls | grep -q requirements-management_redis-data; then
        log_info "Backing up Redis data volume..."
        docker run --rm -v requirements-management_redis-data:/data -v "$(pwd)/$BACKUP_DIR":/backup alpine tar czf /backup/redis-data.tar.gz -C /data .
        log_success "Redis data backup created: $BACKUP_DIR/redis-data.tar.gz"
    fi
}

build_images() {
    log_info "Building Docker images..."
    
    # Use docker-compose or docker compose based on availability
    if command -v docker-compose &> /dev/null; then
        COMPOSE_CMD="docker-compose"
    else
        COMPOSE_CMD="docker compose"
    fi
    
    $COMPOSE_CMD -f "$COMPOSE_FILE" build --no-cache
    log_success "Docker images built successfully"
}

deploy_services() {
    log_info "Deploying services..."
    
    # Stop existing services
    log_info "Stopping existing services..."
    $COMPOSE_CMD -f "$COMPOSE_FILE" down
    
    # Start services
    log_info "Starting services..."
    $COMPOSE_CMD -f "$COMPOSE_FILE" up -d
    
    log_success "Services deployed successfully"
}

wait_for_services() {
    log_info "Waiting for services to be ready..."
    
    # Wait for database
    log_info "Waiting for PostgreSQL..."
    timeout=60
    while ! docker exec requirements-postgres pg_isready -U postgres -d requirements_db &> /dev/null; do
        sleep 2
        timeout=$((timeout - 2))
        if [[ $timeout -le 0 ]]; then
            log_error "PostgreSQL failed to start within 60 seconds"
            exit 1
        fi
    done
    log_success "PostgreSQL is ready"
    
    # Wait for Redis
    log_info "Waiting for Redis..."
    timeout=30
    while ! docker exec requirements-redis redis-cli ping &> /dev/null; do
        sleep 2
        timeout=$((timeout - 2))
        if [[ $timeout -le 0 ]]; then
            log_error "Redis failed to start within 30 seconds"
            exit 1
        fi
    done
    log_success "Redis is ready"
    
    # Wait for application
    log_info "Waiting for application..."
    timeout=60
    while ! curl -f http://localhost:8080/ready &> /dev/null; do
        sleep 2
        timeout=$((timeout - 2))
        if [[ $timeout -le 0 ]]; then
            log_error "Application failed to start within 60 seconds"
            exit 1
        fi
    done
    log_success "Application is ready"
}

run_migrations() {
    log_info "Running database migrations..."
    
    # Run migrations using the migrate tool in the container
    docker exec requirements-app ./migrate -up
    
    log_success "Database migrations completed"
}

verify_deployment() {
    log_info "Verifying deployment..."
    
    # Check if all containers are running
    if ! docker ps | grep -q requirements-app; then
        log_error "Application container is not running"
        exit 1
    fi
    
    if ! docker ps | grep -q requirements-postgres; then
        log_error "PostgreSQL container is not running"
        exit 1
    fi
    
    if ! docker ps | grep -q requirements-redis; then
        log_error "Redis container is not running"
        exit 1
    fi
    
    # Check application health
    if ! curl -f http://localhost:8080/ready &> /dev/null; then
        log_error "Application health check failed"
        exit 1
    fi
    
    log_success "Deployment verification passed"
}

show_status() {
    log_info "Deployment Status:"
    echo ""
    
    # Show container status
    docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
    echo ""
    
    # Show application info
    log_info "Application URL: http://localhost:8080"
    log_info "Health Check: http://localhost:8080/ready"
    log_info "API Documentation: http://localhost:8080/swagger/index.html (if enabled)"
    echo ""
    
    # Show logs command
    log_info "To view logs: $COMPOSE_CMD -f $COMPOSE_FILE logs -f"
    log_info "To stop services: $COMPOSE_CMD -f $COMPOSE_FILE down"
}

cleanup_old_images() {
    log_info "Cleaning up old Docker images..."
    
    # Remove dangling images
    docker image prune -f
    
    log_success "Old images cleaned up"
}

# Main deployment process
main() {
    log_info "Starting production deployment..."
    echo ""
    
    check_prerequisites
    
    # Ask for confirmation
    read -p "This will deploy the application to production. Continue? (y/N): " -n 1 -r
    echo ""
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        log_info "Deployment cancelled"
        exit 0
    fi
    
    backup_data
    build_images
    deploy_services
    wait_for_services
    run_migrations
    verify_deployment
    cleanup_old_images
    
    echo ""
    log_success "Production deployment completed successfully!"
    echo ""
    show_status
}

# Handle script arguments
case "${1:-deploy}" in
    "deploy")
        main
        ;;
    "backup")
        backup_data
        ;;
    "status")
        show_status
        ;;
    "logs")
        $COMPOSE_CMD -f "$COMPOSE_FILE" logs -f
        ;;
    "stop")
        log_info "Stopping services..."
        $COMPOSE_CMD -f "$COMPOSE_FILE" down
        log_success "Services stopped"
        ;;
    "restart")
        log_info "Restarting services..."
        $COMPOSE_CMD -f "$COMPOSE_FILE" restart
        log_success "Services restarted"
        ;;
    *)
        echo "Usage: $0 {deploy|backup|status|logs|stop|restart}"
        echo ""
        echo "Commands:"
        echo "  deploy  - Full production deployment (default)"
        echo "  backup  - Create backup of data"
        echo "  status  - Show deployment status"
        echo "  logs    - Show service logs"
        echo "  stop    - Stop all services"
        echo "  restart - Restart all services"
        exit 1
        ;;
esac
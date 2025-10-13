#!/bin/bash

# Deploy from Registry script for Product Requirements Management System
# This script pulls Docker images from GitHub Container Registry and deploys them

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
GITHUB_REGISTRY="ghcr.io"
GITHUB_USERNAME="${GITHUB_USERNAME:-$(git remote get-url origin | sed -n 's/.*github\.com[:/]\([^/]*\)\/.*/\1/p')}"
REPOSITORY_NAME="${REPOSITORY_NAME:-$(git remote get-url origin | sed -n 's/.*github\.com[:/][^/]*\/\([^.]*\)\.git/\1/p')}"
IMAGE_NAME="${GITHUB_REGISTRY}/${GITHUB_USERNAME}/${REPOSITORY_NAME}"
VERSION="${VERSION:-latest}"

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
    if command -v docker-compose &> /dev/null; then
        COMPOSE_CMD="docker-compose"
    elif docker compose version &> /dev/null; then
        COMPOSE_CMD="docker compose"
    else
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

authenticate_github() {
    log_info "Authenticating with GitHub Container Registry..."
    
    if [[ -n "$GITHUB_TOKEN" ]]; then
        echo "$GITHUB_TOKEN" | docker login ghcr.io -u "$GITHUB_USERNAME" --password-stdin
        log_success "Authenticated with GitHub Container Registry using token"
    else
        log_warning "GITHUB_TOKEN not set, trying existing authentication"
        # Try to pull without explicit login (may use existing credentials)
    fi
}

load_deployment_info() {
    if [[ -f "deployment-info.json" ]]; then
        log_info "Loading deployment information from deployment-info.json..."
        
        # Extract information from deployment-info.json if available
        if command -v jq &> /dev/null; then
            IMAGE_NAME=$(jq -r '.image' deployment-info.json)
            VERSION=$(jq -r '.version' deployment-info.json)
            log_info "Using image: ${IMAGE_NAME}:${VERSION}"
        else
            log_warning "jq not available, using default image configuration"
        fi
    fi
}

pull_image() {
    log_info "Pulling Docker image from GitHub Container Registry..."
    
    local image_tag="${IMAGE_NAME}:${VERSION}"
    log_info "Pulling: $image_tag"
    
    docker pull "$image_tag"
    log_success "Image pulled successfully: $image_tag"
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

update_compose_file() {
    log_info "Updating docker-compose file with new image..."
    
    # Create a temporary compose file with the specific image
    local temp_compose="docker-compose.prod.temp.yml"
    local image_tag="${IMAGE_NAME}:${VERSION}"
    
    # Replace the build context with the specific image
    sed "s|build:|#build:|g; s|context: \.|#context: .|g; s|dockerfile: Dockerfile|#dockerfile: Dockerfile|g" "$COMPOSE_FILE" > "$temp_compose"
    sed -i.bak "s|image: requirements-management:latest|image: ${image_tag}|g" "$temp_compose"
    
    # Use the temporary compose file
    COMPOSE_FILE="$temp_compose"
    
    log_success "Compose file updated with image: $image_tag"
}

deploy_services() {
    log_info "Deploying services..."
    
    # Load environment variables
    if [[ -f "$ENV_FILE" ]]; then
        log_info "Loading environment variables from $ENV_FILE"
        set -a
        source "$ENV_FILE"
        set +a
    fi
    
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

run_init() {
    log_info "Running database initialization..."
    
    # Run initialization using the init tool in the container
    if docker exec requirements-app test -f ./init; then
        docker exec requirements-app ./init
        log_success "Database initialization completed"
    else
        log_warning "Init binary not found in container, skipping initialization"
    fi
}

run_migrations() {
    log_info "Running database migrations..."
    
    # Run migrations using the migrate tool in the container
    if docker exec requirements-app test -f ./migrate; then
        docker exec requirements-app ./migrate -up
        log_success "Database migrations completed"
    else
        log_error "Migrate binary not found in container"
        exit 1
    fi
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
    
    # Show current image info
    if docker ps --format "{{.Image}}" | grep -q "$IMAGE_NAME"; then
        local current_image=$(docker ps --format "{{.Names}}\t{{.Image}}" | grep requirements-app | cut -f2)
        log_info "Current image: $current_image"
    fi
    
    # Show logs command
    log_info "To view logs: $COMPOSE_CMD -f docker-compose.prod.yml logs -f"
    log_info "To stop services: $COMPOSE_CMD -f docker-compose.prod.yml down"
}

cleanup_temp_files() {
    log_info "Cleaning up temporary files..."
    
    # Remove temporary compose file
    if [[ -f "docker-compose.prod.temp.yml" ]]; then
        rm -f docker-compose.prod.temp.yml
        rm -f docker-compose.prod.temp.yml.bak
    fi
    
    # Remove dangling images
    docker image prune -f
    
    log_success "Cleanup completed"
}

# Main deployment process
main() {
    log_info "Starting deployment from GitHub Container Registry..."
    echo ""
    
    check_prerequisites
    load_deployment_info
    
    log_info "Image to deploy: ${IMAGE_NAME}:${VERSION}"
    
    # Ask for confirmation
    read -p "Deploy ${IMAGE_NAME}:${VERSION} to production? (y/N): " -n 1 -r
    echo ""
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        log_info "Deployment cancelled"
        exit 0
    fi
    
    authenticate_github
    pull_image
    backup_data
    update_compose_file
    deploy_services
    wait_for_services
    run_migrations
    verify_deployment
    cleanup_temp_files
    
    echo ""
    log_success "Deployment completed successfully!"
    echo ""
    show_status
}

# Handle script arguments
case "${1:-deploy}" in
    "deploy")
        main
        ;;
    "backup")
        check_prerequisites
        backup_data
        ;;
    "init")
        check_prerequisites
        run_init
        ;;
    "migrate")
        check_prerequisites
        run_migrations
        ;;
    "update")
        check_prerequisites
        load_deployment_info
        authenticate_github
        pull_image
        log_info "Restarting services with new image..."
        $COMPOSE_CMD -f docker-compose.prod.yml restart app
        wait_for_services
        verify_deployment
        log_success "Update completed successfully!"
        ;;
    "restart")
        check_prerequisites
        log_info "Restarting services..."
        $COMPOSE_CMD -f docker-compose.prod.yml restart
        wait_for_services
        verify_deployment
        log_success "Services restarted successfully!"
        ;;
    "status")
        show_status
        ;;
    "logs")
        $COMPOSE_CMD -f docker-compose.prod.yml logs -f
        ;;
    "stop")
        log_info "Stopping services..."
        $COMPOSE_CMD -f docker-compose.prod.yml down
        log_success "Services stopped"
        ;;
    "pull")
        check_prerequisites
        load_deployment_info
        authenticate_github
        pull_image
        ;;
    *)
        echo "Usage: $0 {deploy|backup|init|migrate|update|restart|status|logs|stop|pull}"
        echo ""
        echo "Commands:"
        echo "  deploy   - Full deployment from registry (default)"
        echo "  backup   - Create backup of data"
        echo "  init     - Run database initialization (./bin/init)"
        echo "  migrate  - Run database migrations"
        echo "  update   - Pull new image and restart app container"
        echo "  restart  - Restart all services"
        echo "  status   - Show deployment status"
        echo "  logs     - Show service logs"
        echo "  stop     - Stop all services"
        echo "  pull     - Pull image from registry"
        echo ""
        echo "Environment Variables:"
        echo "  GITHUB_TOKEN     - GitHub Personal Access Token"
        echo "  GITHUB_USERNAME  - GitHub username (auto-detected from git)"
        echo "  REPOSITORY_NAME  - Repository name (default: product-requirements-management)"
        echo "  VERSION          - Image version (default: latest)"
        echo ""
        echo "Examples:"
        echo "  VERSION=v1.0.0 ./scripts/deploy-from-registry.sh deploy"
        echo "  ./scripts/deploy-from-registry.sh update"
        echo "  ./scripts/deploy-from-registry.sh migrate"
        exit 1
        ;;
esac
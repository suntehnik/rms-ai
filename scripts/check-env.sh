#!/bin/bash

# Environment variables check script
# This script validates that all required environment variables are set

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
ENV_FILE="${ENV_FILE:-.env.prod}"

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

check_env_file() {
    log_info "Checking environment file: $ENV_FILE"
    
    if [[ ! -f "$ENV_FILE" ]]; then
        log_error "Environment file $ENV_FILE not found"
        log_info "Please copy .env.prod.template to $ENV_FILE and configure it"
        exit 1
    fi
    
    log_success "Environment file found"
}

load_env_file() {
    log_info "Loading environment variables from $ENV_FILE"
    
    # Load environment variables
    set -a
    source "$ENV_FILE"
    set +a
    
    log_success "Environment variables loaded"
}

check_required_vars() {
    log_info "Checking required environment variables..."
    
    local missing_vars=()
    local weak_vars=()
    
    # Required variables
    local required_vars=(
        "DB_PASSWORD"
        "JWT_SECRET"
        "DEFAULT_ADMIN_PASSWORD"
        "PAT_TOKEN"
    )
    
    for var in "${required_vars[@]}"; do
        if [[ -z "${!var}" ]]; then
            missing_vars+=("$var")
        elif [[ "${!var}" == *"your_"* ]] || [[ "${!var}" == *"here"* ]]; then
            weak_vars+=("$var")
        fi
    done
    
    # Check for missing variables
    if [[ ${#missing_vars[@]} -gt 0 ]]; then
        log_error "Missing required environment variables:"
        for var in "${missing_vars[@]}"; do
            echo "  - $var"
        done
        return 1
    fi
    
    # Check for weak/template variables
    if [[ ${#weak_vars[@]} -gt 0 ]]; then
        log_warning "Variables still contain template values (need to be changed):"
        for var in "${weak_vars[@]}"; do
            echo "  - $var"
        done
        return 1
    fi
    
    log_success "All required variables are set"
}

check_variable_strength() {
    log_info "Checking variable strength..."
    
    local weak_passwords=()
    
    # Check JWT_SECRET length
    if [[ ${#JWT_SECRET} -lt 32 ]]; then
        weak_passwords+=("JWT_SECRET (too short, minimum 32 characters)")
    fi
    
    # Check DB_PASSWORD strength
    if [[ ${#DB_PASSWORD} -lt 12 ]]; then
        weak_passwords+=("DB_PASSWORD (too short, minimum 12 characters)")
    fi
    
    # Check DEFAULT_ADMIN_PASSWORD strength
    if [[ ${#DEFAULT_ADMIN_PASSWORD} -lt 8 ]]; then
        weak_passwords+=("DEFAULT_ADMIN_PASSWORD (too short, minimum 8 characters)")
    fi
    
    # Check PAT_TOKEN format
    if [[ ! "$PAT_TOKEN" =~ ^mcp_pat_ ]]; then
        weak_passwords+=("PAT_TOKEN (should start with 'mcp_pat_')")
    fi
    
    if [[ ${#weak_passwords[@]} -gt 0 ]]; then
        log_warning "Weak security settings detected:"
        for item in "${weak_passwords[@]}"; do
            echo "  - $item"
        done
        return 1
    fi
    
    log_success "All variables meet security requirements"
}

check_optional_vars() {
    log_info "Checking optional variables..."
    
    # Optional variables with recommendations
    if [[ -z "$REDIS_PASSWORD" ]]; then
        log_warning "REDIS_PASSWORD not set - Redis will run without authentication"
    else
        log_success "REDIS_PASSWORD is set"
    fi
}

generate_secure_values() {
    log_info "Generating secure example values..."
    
    echo ""
    echo "Example secure values (copy to your $ENV_FILE):"
    echo ""
    
    # Generate JWT_SECRET
    local jwt_secret=$(openssl rand -base64 32 | tr -d '\n')
    echo "JWT_SECRET=$jwt_secret"
    
    # Generate DB_PASSWORD
    local db_password=$(openssl rand -base64 16 | tr -d '\n')
    echo "DB_PASSWORD=$db_password"
    
    # Generate DEFAULT_ADMIN_PASSWORD
    local admin_password=$(openssl rand -base64 12 | tr -d '\n')
    echo "DEFAULT_ADMIN_PASSWORD=$admin_password"
    
    # Generate PAT_TOKEN
    local pat_token="mcp_pat_$(openssl rand -base64 32 | tr -d '\n' | tr '+/' '-_')"
    echo "PAT_TOKEN=$pat_token"
    
    # Generate REDIS_PASSWORD
    local redis_password=$(openssl rand -base64 16 | tr -d '\n')
    echo "REDIS_PASSWORD=$redis_password"
    
    echo ""
}

show_docker_compose_check() {
    log_info "Testing Docker Compose configuration..."
    
    # Test docker-compose config
    if docker-compose -f docker-compose.prod.yml config > /dev/null 2>&1; then
        log_success "Docker Compose configuration is valid"
    else
        log_error "Docker Compose configuration has errors"
        echo "Run: docker-compose -f docker-compose.prod.yml config"
        return 1
    fi
}

# Main check process
main() {
    log_info "Starting environment variables check..."
    echo ""
    
    local errors=0
    
    check_env_file || ((errors++))
    
    if [[ $errors -eq 0 ]]; then
        load_env_file || ((errors++))
        check_required_vars || ((errors++))
        check_variable_strength || ((errors++))
        check_optional_vars
        show_docker_compose_check || ((errors++))
    fi
    
    echo ""
    
    if [[ $errors -gt 0 ]]; then
        log_error "Environment check failed with $errors error(s)"
        echo ""
        generate_secure_values
        exit 1
    else
        log_success "Environment check passed successfully!"
        log_info "Your production environment is ready for deployment"
    fi
}

# Handle script arguments
case "${1:-check}" in
    "check")
        main
        ;;
    "generate")
        generate_secure_values
        ;;
    "test-compose")
        check_env_file
        load_env_file
        show_docker_compose_check
        ;;
    *)
        echo "Usage: $0 {check|generate|test-compose}"
        echo ""
        echo "Commands:"
        echo "  check        - Check environment variables (default)"
        echo "  generate     - Generate secure example values"
        echo "  test-compose - Test Docker Compose configuration"
        echo ""
        echo "Environment Variables:"
        echo "  ENV_FILE     - Environment file to check (default: .env.prod)"
        echo ""
        echo "Examples:"
        echo "  ./scripts/check-env.sh"
        echo "  ENV_FILE=.env.staging ./scripts/check-env.sh"
        echo "  ./scripts/check-env.sh generate"
        exit 1
        ;;
esac
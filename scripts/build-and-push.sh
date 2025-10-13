#!/bin/bash

# Build and Push script for Product Requirements Management System
# This script builds Docker images and pushes them to GitHub Container Registry

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
GITHUB_REGISTRY="ghcr.io"
GITHUB_USERNAME="${GITHUB_USERNAME:-$(git remote get-url origin | sed -n 's/.*github\.com[:/]\([^/]*\)\/.*/\1/p')}"
REPOSITORY_NAME="${REPOSITORY_NAME:-$(git remote get-url origin | sed -n 's/.*github\.com[:/][^/]*\/\([^.]*\)\.git/\1/p')}"
IMAGE_NAME="${GITHUB_REGISTRY}/${GITHUB_USERNAME}/${REPOSITORY_NAME}"
VERSION="${VERSION:-$(git rev-parse --short HEAD)}"
LATEST_TAG="latest"

# Multi-platform build configuration
PLATFORMS="${PLATFORMS:-linux/amd64,linux/arm64}"
BUILDER_NAME="multiarch-builder"
USE_BUILDX="${USE_BUILDX:-true}"

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
    
    # Check Docker version for buildx support
    DOCKER_VERSION=$(docker version --format '{{.Client.Version}}' | cut -d. -f1-2)
    log_info "Docker version: $DOCKER_VERSION"
    
    # Check if buildx is available and enabled
    if [[ "$USE_BUILDX" == "true" ]]; then
        if ! docker buildx version &> /dev/null; then
            log_warning "Docker buildx not available, falling back to regular build"
            USE_BUILDX="false"
        else
            log_info "Docker buildx available for multi-platform builds"
        fi
    fi
    
    # Check if git is available
    if ! command -v git &> /dev/null; then
        log_error "Git is not installed"
        exit 1
    fi
    
    # Check if we're in a git repository
    if ! git rev-parse --git-dir &> /dev/null; then
        log_error "Not in a git repository"
        exit 1
    fi
    
    # Check if GitHub token is available
    if [[ -z "$GITHUB_TOKEN" ]]; then
        log_warning "GITHUB_TOKEN environment variable not set"
        log_info "You may need to authenticate manually with: docker login ghcr.io"
    fi
    
    log_success "Prerequisites check passed"
}

authenticate_github() {
    log_info "Authenticating with GitHub Container Registry..."
    
    if [[ -n "$GITHUB_TOKEN" ]]; then
        echo "$GITHUB_TOKEN" | docker login ghcr.io -u "$GITHUB_USERNAME" --password-stdin
        log_success "Authenticated with GitHub Container Registry using token"
    else
        log_info "Please authenticate manually:"
        docker login ghcr.io
    fi
}

get_version_info() {
    log_info "Getting version information..."
    
    # Get git commit hash
    GIT_COMMIT=$(git rev-parse HEAD)
    GIT_SHORT_COMMIT=$(git rev-parse --short HEAD)
    
    # Get git branch
    GIT_BRANCH=$(git rev-parse --abbrev-ref HEAD)
    
    # Get git tag if exists
    GIT_TAG=$(git describe --tags --exact-match 2>/dev/null || echo "")
    
    # Determine version
    if [[ -n "$GIT_TAG" ]]; then
        VERSION="$GIT_TAG"
    else
        VERSION="$GIT_BRANCH-$GIT_SHORT_COMMIT"
    fi
    
    # Clean version for Docker tag (replace invalid characters)
    VERSION=$(echo "$VERSION" | sed 's/[^a-zA-Z0-9._-]/-/g')
    
    log_info "Git commit: $GIT_COMMIT"
    log_info "Git branch: $GIT_BRANCH"
    log_info "Git tag: ${GIT_TAG:-none}"
    log_info "Docker version tag: $VERSION"
}

setup_buildx() {
    if [[ "$USE_BUILDX" == "true" ]]; then
        log_info "Setting up Docker buildx for multi-platform builds..."
        
        # Create builder if it doesn't exist
        if ! docker buildx inspect "$BUILDER_NAME" &> /dev/null; then
            log_info "Creating buildx builder: $BUILDER_NAME"
            docker buildx create --name "$BUILDER_NAME" --driver docker-container --bootstrap
        fi
        
        # Use the builder
        docker buildx use "$BUILDER_NAME"
        
        # Inspect builder to show supported platforms
        log_info "Available platforms:"
        docker buildx inspect --bootstrap | grep "Platforms:" || true
        
        log_success "Buildx setup completed"
    fi
}

build_image() {
    if [[ "$USE_BUILDX" == "true" ]]; then
        build_multiplatform_image
    else
        build_single_platform_image
    fi
}

build_single_platform_image() {
    log_info "Building single-platform Docker image..."
    log_warning "Building for current platform only (no cross-compilation)"
    
    # Build image with multiple tags for current platform
    docker build \
        --build-arg GIT_COMMIT="$GIT_COMMIT" \
        --build-arg GIT_BRANCH="$GIT_BRANCH" \
        --build-arg BUILD_DATE="$(date -u +'%Y-%m-%dT%H:%M:%SZ')" \
        --build-arg VERSION="$VERSION" \
        -t "${IMAGE_NAME}:${VERSION}" \
        -t "${IMAGE_NAME}:${LATEST_TAG}" \
        .
    
    log_success "Single-platform Docker image built successfully"
    log_info "Image tags:"
    log_info "  ${IMAGE_NAME}:${VERSION}"
    log_info "  ${IMAGE_NAME}:${LATEST_TAG}"
}

build_multiplatform_image() {
    log_info "Building multi-platform Docker image..."
    log_info "Target platforms: $PLATFORMS"
    
    # Build and push multi-platform image
    docker buildx build \
        --platform "$PLATFORMS" \
        --build-arg GIT_COMMIT="$GIT_COMMIT" \
        --build-arg GIT_BRANCH="$GIT_BRANCH" \
        --build-arg BUILD_DATE="$(date -u +'%Y-%m-%dT%H:%M:%SZ')" \
        --build-arg VERSION="$VERSION" \
        -t "${IMAGE_NAME}:${VERSION}" \
        -t "${IMAGE_NAME}:${LATEST_TAG}" \
        --push \
        .
    
    log_success "Multi-platform Docker image built and pushed successfully"
    log_info "Platforms: $PLATFORMS"
    log_info "Image tags:"
    log_info "  ${IMAGE_NAME}:${VERSION}"
    log_info "  ${IMAGE_NAME}:${LATEST_TAG}"
}

push_image() {
    # Skip push if using buildx (already pushed during build)
    if [[ "$USE_BUILDX" == "true" ]]; then
        log_info "Images already pushed during buildx build"
        return 0
    fi
    
    log_info "Pushing Docker image to GitHub Container Registry..."
    
    # Push version tag
    docker push "${IMAGE_NAME}:${VERSION}"
    log_success "Pushed ${IMAGE_NAME}:${VERSION}"
    
    # Push latest tag
    docker push "${IMAGE_NAME}:${LATEST_TAG}"
    log_success "Pushed ${IMAGE_NAME}:${LATEST_TAG}"
}

cleanup_local_images() {
    log_info "Cleaning up local images..."
    
    # Remove dangling images
    docker image prune -f
    
    # Clean buildx cache if using buildx
    if [[ "$USE_BUILDX" == "true" ]]; then
        log_info "Cleaning buildx cache..."
        docker buildx prune -f || true
    fi
    
    log_success "Local cleanup completed"
}

cleanup_buildx() {
    if [[ "$USE_BUILDX" == "true" ]]; then
        log_info "Cleaning up buildx builder..."
        docker buildx rm "$BUILDER_NAME" || true
        log_success "Buildx builder cleaned up"
    fi
}

show_image_info() {
    log_info "Image Information:"
    echo ""
    log_info "Registry: $GITHUB_REGISTRY"
    log_info "Repository: ${GITHUB_USERNAME}/${REPOSITORY_NAME}"
    log_info "Image: ${IMAGE_NAME}"
    log_info "Version: $VERSION"
    log_info "Latest: ${IMAGE_NAME}:${LATEST_TAG}"
    log_info "Build type: $([ "$USE_BUILDX" == "true" ] && echo "Multi-platform" || echo "Single-platform")"
    if [[ "$USE_BUILDX" == "true" ]]; then
        log_info "Platforms: $PLATFORMS"
    fi
    echo ""
    log_info "Pull commands:"
    log_info "  docker pull ${IMAGE_NAME}:${VERSION}"
    log_info "  docker pull ${IMAGE_NAME}:${LATEST_TAG}"
    echo ""
    log_info "Platform-specific pull (if multi-platform):"
    log_info "  docker pull --platform linux/amd64 ${IMAGE_NAME}:${VERSION}"
    log_info "  docker pull --platform linux/arm64 ${IMAGE_NAME}:${VERSION}"
    echo ""
    log_info "GitHub Packages URL:"
    log_info "  https://github.com/${GITHUB_USERNAME}/${REPOSITORY_NAME}/pkgs/container/${REPOSITORY_NAME}"
}

generate_deployment_info() {
    log_info "Generating deployment information..."
    
    # Create deployment info file
    cat > deployment-info.json << EOF
{
  "image": "${IMAGE_NAME}",
  "version": "${VERSION}",
  "latest_tag": "${IMAGE_NAME}:${LATEST_TAG}",
  "version_tag": "${IMAGE_NAME}:${VERSION}",
  "git_commit": "${GIT_COMMIT}",
  "git_branch": "${GIT_BRANCH}",
  "git_tag": "${GIT_TAG}",
  "build_date": "$(date -u +'%Y-%m-%dT%H:%M:%SZ')",
  "registry": "${GITHUB_REGISTRY}",
  "repository": "${GITHUB_USERNAME}/${REPOSITORY_NAME}",
  "build_type": "$([ "$USE_BUILDX" == "true" ] && echo "multi-platform" || echo "single-platform")",
  "platforms": "${PLATFORMS}",
  "buildx_enabled": ${USE_BUILDX}
}
EOF
    
    log_success "Deployment info saved to deployment-info.json"
}

# Main build and push process
main() {
    log_info "Starting build and push process..."
    echo ""
    
    check_prerequisites
    get_version_info
    
    # Ask for confirmation
    read -p "Build and push ${IMAGE_NAME}:${VERSION} to GitHub Registry? (y/N): " -n 1 -r
    echo ""
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        log_info "Build cancelled"
        exit 0
    fi
    
    authenticate_github
    setup_buildx
    build_image
    push_image
    generate_deployment_info
    cleanup_local_images
    
    echo ""
    log_success "Build and push completed successfully!"
    echo ""
    show_image_info
}

# Handle script arguments
case "${1:-build}" in
    "build")
        main
        ;;
    "info")
        get_version_info
        show_image_info
        ;;
    "auth")
        authenticate_github
        ;;
    "cleanup")
        cleanup_local_images
        ;;
    "cleanup-buildx")
        cleanup_buildx
        ;;
    "setup-buildx")
        setup_buildx
        ;;
    *)
        echo "Usage: $0 {build|info|auth|cleanup|cleanup-buildx|setup-buildx}"
        echo ""
        echo "Commands:"
        echo "  build         - Build and push Docker image (default)"
        echo "  info          - Show image information"
        echo "  auth          - Authenticate with GitHub Registry"
        echo "  cleanup       - Clean up local Docker images"
        echo "  cleanup-buildx - Remove buildx builder"
        echo "  setup-buildx  - Setup buildx builder for multi-platform builds"
        echo ""
        echo "Environment Variables:"
        echo "  GITHUB_TOKEN     - GitHub Personal Access Token"
        echo "  GITHUB_USERNAME  - GitHub username (auto-detected from git)"
        echo "  REPOSITORY_NAME  - Repository name (default: product-requirements-management)"
        echo "  VERSION          - Image version (auto-detected from git)"
        echo "  PLATFORMS        - Target platforms (default: linux/amd64,linux/arm64)"
        echo "  USE_BUILDX       - Enable buildx for multi-platform (default: true)"
        echo ""
        echo "Platform Examples:"
        echo "  PLATFORMS=linux/amd64 ./scripts/build-and-push.sh          # x86_64 only"
        echo "  PLATFORMS=linux/arm64 ./scripts/build-and-push.sh          # ARM64 only"
        echo "  PLATFORMS=linux/amd64,linux/arm64 ./scripts/build-and-push.sh  # Both (default)"
        echo "  USE_BUILDX=false ./scripts/build-and-push.sh               # Single platform"
        echo ""
        echo "Examples:"
        echo "  GITHUB_TOKEN=ghp_xxx ./scripts/build-and-push.sh"
        echo "  VERSION=v1.0.0 PLATFORMS=linux/amd64 ./scripts/build-and-push.sh"
        exit 1
        ;;
esac
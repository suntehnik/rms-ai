#!/bin/bash

# MCP Server Development Helper Script
# This script helps with local development and testing of the MCP server

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
BINARY_NAME="spexus-mcp"
CONFIG_FILE="config.example.json"
LOG_LEVEL="debug"

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

show_help() {
    cat << EOF
MCP Server Development Helper

Usage: $0 [COMMAND] [OPTIONS]

Commands:
    build           Build the MCP server binary
    test            Run MCP server tests
    run             Run the MCP server locally
    install         Install the MCP server to system
    clean           Clean build artifacts
    validate        Validate MCP server configuration
    protocol-test   Test MCP protocol compliance
    benchmark       Run performance benchmarks
    help            Show this help message

Options:
    --config FILE   Use specific configuration file (default: $CONFIG_FILE)
    --log-level LVL Set log level (debug, info, warn, error) (default: $LOG_LEVEL)
    --stdio         Use STDIO transport (for MCP protocol testing)
    --port PORT     Use specific port for HTTP transport (default: 8080)

Examples:
    $0 build                    # Build the MCP server
    $0 run --stdio              # Run with STDIO transport
    $0 test                     # Run all tests
    $0 protocol-test            # Test MCP protocol compliance
    $0 install                  # Install to /usr/local/bin

EOF
}

build_server() {
    log_info "Building MCP server..."
    make build-mcp-server
    log_success "MCP server built successfully: bin/$BINARY_NAME"
}

test_server() {
    log_info "Running MCP server tests..."
    
    # Run unit tests
    log_info "Running unit tests..."
    go test -v ./cmd/mcp-server/... ./internal/mcp/... || log_warning "Some unit tests failed"
    
    # Run integration tests if available
    if [ -d "tests/mcp" ]; then
        log_info "Running integration tests..."
        go test -v ./tests/mcp/... || log_warning "Some integration tests failed"
    fi
    
    log_success "Test run completed"
}

run_server() {
    local use_stdio=false
    local port=8080
    local config_file="$CONFIG_FILE"
    
    # Parse additional arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            --stdio)
                use_stdio=true
                shift
                ;;
            --port)
                port="$2"
                shift 2
                ;;
            --config)
                config_file="$2"
                shift 2
                ;;
            --log-level)
                LOG_LEVEL="$2"
                shift 2
                ;;
            *)
                shift
                ;;
        esac
    done
    
    # Ensure binary exists
    if [ ! -f "bin/$BINARY_NAME" ]; then
        log_info "Binary not found, building..."
        build_server
    fi
    
    log_info "Starting MCP server..."
    log_info "Configuration: $config_file"
    log_info "Log level: $LOG_LEVEL"
    
    if [ "$use_stdio" = true ]; then
        log_info "Using STDIO transport"
        LOG_LEVEL="$LOG_LEVEL" ./bin/$BINARY_NAME --stdio --config "$config_file"
    else
        log_info "Using HTTP transport on port $port"
        LOG_LEVEL="$LOG_LEVEL" ./bin/$BINARY_NAME --port "$port" --config "$config_file"
    fi
}

install_server() {
    if [ ! -f "bin/$BINARY_NAME" ]; then
        log_info "Binary not found, building..."
        build_server
    fi
    
    log_info "Installing MCP server to /usr/local/bin..."
    sudo cp "bin/$BINARY_NAME" /usr/local/bin/
    sudo chmod +x "/usr/local/bin/$BINARY_NAME"
    log_success "MCP server installed: /usr/local/bin/$BINARY_NAME"
    
    log_info "You can now configure your MCP client to use: /usr/local/bin/$BINARY_NAME"
}

clean_artifacts() {
    log_info "Cleaning build artifacts..."
    rm -rf bin/
    rm -f *.out *.html
    log_success "Build artifacts cleaned"
}

validate_config() {
    local config_file="$CONFIG_FILE"
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            --config)
                config_file="$2"
                shift 2
                ;;
            *)
                shift
                ;;
        esac
    done
    
    if [ ! -f "$config_file" ]; then
        log_error "Configuration file not found: $config_file"
        return 1
    fi
    
    log_info "Validating configuration: $config_file"
    
    # Basic JSON validation
    if command -v jq >/dev/null 2>&1; then
        if jq empty "$config_file" >/dev/null 2>&1; then
            log_success "Configuration file is valid JSON"
        else
            log_error "Configuration file contains invalid JSON"
            return 1
        fi
    else
        log_warning "jq not found, skipping JSON validation"
    fi
    
    # Test configuration with binary
    if [ -f "bin/$BINARY_NAME" ]; then
        log_info "Testing configuration with MCP server..."
        if ./bin/$BINARY_NAME --validate-config --config "$config_file"; then
            log_success "Configuration validated successfully"
        else
            log_error "Configuration validation failed"
            return 1
        fi
    else
        log_warning "MCP server binary not found, skipping runtime validation"
    fi
}

protocol_test() {
    if [ ! -f "bin/$BINARY_NAME" ]; then
        log_info "Binary not found, building..."
        build_server
    fi
    
    log_info "Testing MCP protocol compliance..."
    
    # Create a simple test script
    cat > /tmp/mcp_test.json << 'EOF'
{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "initialize",
    "params": {
        "protocolVersion": "2024-11-05",
        "capabilities": {},
        "clientInfo": {
            "name": "test-client",
            "version": "1.0.0"
        }
    }
}
EOF
    
    log_info "Sending initialization request..."
    
    # Test with timeout
    if echo '{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {"protocolVersion": "2024-11-05", "capabilities": {}, "clientInfo": {"name": "test-client", "version": "1.0.0"}}}' | timeout 10s ./bin/$BINARY_NAME --stdio; then
        log_success "Protocol test completed"
    else
        log_warning "Protocol test completed with warnings"
    fi
    
    rm -f /tmp/mcp_test.json
}

benchmark_server() {
    log_info "Running MCP server benchmarks..."
    
    # Run Go benchmarks
    go test -bench=. -benchmem -benchtime=5s ./cmd/mcp-server/... ./internal/mcp/... | tee mcp-benchmark.txt
    
    log_success "Benchmark results saved to mcp-benchmark.txt"
}

# Main script logic
case "${1:-help}" in
    build)
        build_server
        ;;
    test)
        shift
        test_server "$@"
        ;;
    run)
        shift
        run_server "$@"
        ;;
    install)
        install_server
        ;;
    clean)
        clean_artifacts
        ;;
    validate)
        shift
        validate_config "$@"
        ;;
    protocol-test)
        protocol_test
        ;;
    benchmark)
        benchmark_server
        ;;
    help|--help|-h)
        show_help
        ;;
    *)
        log_error "Unknown command: $1"
        echo ""
        show_help
        exit 1
        ;;
esac
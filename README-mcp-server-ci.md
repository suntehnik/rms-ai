# MCP Server CI/CD & Development Guide

This document describes the Continuous Integration/Continuous Deployment (CI/CD) setup and development workflow for the Requirements Management MCP Server.

## 🚀 GitHub Actions Workflows

### 1. MCP Server Build & Publish (`mcp-server-build.yml`)

**Triggers:**
- Push to `main` or `develop` branches (when MCP server files change)
- Pull requests to `main` or `develop` branches
- GitHub releases
- Manual workflow dispatch

**Features:**
- **Multi-platform builds**: Linux (AMD64, ARM64), macOS (AMD64, ARM64), Windows (AMD64)
- **Automated testing**: Runs fast unit tests on Linux AMD64 builds
- **Version management**: Automatic versioning based on git commits and releases
- **Artifact publishing**: Uploads binaries to GitHub artifacts with checksums
- **Release creation**: Automatically creates GitHub releases with all platform binaries
- **Docker support**: Builds and publishes Docker images to GitHub Container Registry

**Artifacts produced:**
- `spexus-mcp-linux-amd64.tar.gz`
- `spexus-mcp-linux-arm64.tar.gz`
- `spexus-mcp-darwin-amd64.tar.gz`
- `spexus-mcp-darwin-arm64.tar.gz`
- `spexus-mcp-windows-amd64.exe.zip`
- SHA256 checksums for all binaries
- Docker image: `ghcr.io/[repository]/spexus-mcp`

### 2. MCP Server Testing (`mcp-server-test.yml`)

**Triggers:**
- Push to `main` or `develop` branches (when MCP server files change)
- Pull requests
- Manual workflow dispatch

**Test Suites:**
- **Unit Tests**: Fast tests with coverage reporting
- **Integration Tests**: Tests with PostgreSQL database
- **Protocol Compliance**: MCP protocol specification compliance tests
- **Security Scanning**: Gosec security analysis and vulnerability checking
- **Cross-Platform Compatibility**: Tests on Ubuntu, macOS, and Windows
- **Performance Tests**: Benchmarks (triggered on main branch or with `[perf]` in commit message)

## 📦 Installation & Usage

### From GitHub Releases

1. **Download the latest release:**
   ```bash
   # Linux AMD64
   wget https://github.com/[your-repo]/releases/latest/download/spexus-mcp-linux-amd64.tar.gz
   tar -xzf spexus-mcp-linux-amd64.tar.gz
   
   # macOS ARM64 (Apple Silicon)
   wget https://github.com/[your-repo]/releases/latest/download/spexus-mcp-darwin-arm64.tar.gz
   tar -xzf spexus-mcp-darwin-arm64.tar.gz
   ```

2. **Install the binary:**
   ```bash
   chmod +x spexus-mcp
   sudo mv spexus-mcp /usr/local/bin/
   ```

3. **Verify installation:**
   ```bash
   spexus-mcp --version
   ```

### From Docker

```bash
# Pull the latest image
docker pull ghcr.io/[your-repo]/spexus-mcp:main

# Run the MCP server
docker run -p 8080:8080 ghcr.io/[your-repo]/spexus-mcp:main
```

### From Source

```bash
# Clone the repository
git clone https://github.com/[your-repo].git
cd [your-repo]

# Build and install
make build-mcp-server
make install-mcp-server
```

## 🛠️ Development Workflow

### Local Development Script

Use the provided development helper script:

```bash
# Build the MCP server
./scripts/mcp-dev.sh build

# Run tests
./scripts/mcp-dev.sh test

# Run the server locally with STDIO transport
./scripts/mcp-dev.sh run --stdio

# Run the server with HTTP transport
./scripts/mcp-dev.sh run --port 8080

# Validate configuration
./scripts/mcp-dev.sh validate --config config.example.json

# Test MCP protocol compliance
./scripts/mcp-dev.sh protocol-test

# Run performance benchmarks
./scripts/mcp-dev.sh benchmark

# Install to system
./scripts/mcp-dev.sh install

# Clean build artifacts
./scripts/mcp-dev.sh clean
```

### Manual Development Commands

```bash
# Build MCP server
make build-mcp-server

# Install to system
make install-mcp-server

# Run tests
make test-fast

# Run with development settings
LOG_LEVEL=debug ./bin/requirements-mcp-server --stdio
```

## 🔧 Configuration

### MCP Client Configuration

Configure your MCP client (like Claude Desktop) to use the server:

```json
{
  "mcpServers": {
    "spexus": {
      "command": "/usr/local/bin/spexus-mcp",
      "args": ["--config", "/path/to/config.json"],
      "env": {
        "LOG_LEVEL": "info"
      }
    }
  }
}
```

### Server Configuration

Create a configuration file based on `config.example.json`:

```json
{
  "database": {
    "url": "postgres://user:pass@localhost:5432/requirements?sslmode=disable"
  },
  "server": {
    "port": 8080,
    "host": "localhost"
  },
  "logging": {
    "level": "info",
    "format": "json"
  },
  "mcp": {
    "protocol_version": "2024-11-05",
    "capabilities": {
      "tools": true,
      "resources": true,
      "prompts": true
    }
  }
}
```

## 🚀 Release Process

### Automatic Releases

1. **Development releases**: Automatically created on pushes to `main` branch
2. **Tagged releases**: Created when you publish a GitHub release
3. **Manual releases**: Use workflow dispatch to create releases

### Manual Release Creation

1. **Via GitHub UI:**
   - Go to Actions → MCP Server Build & Publish
   - Click "Run workflow"
   - Set "Create a GitHub release" to true
   - Click "Run workflow"

2. **Via Git tags:**
   ```bash
   git tag v1.0.0
   git push origin v1.0.0
   # Then create a GitHub release from the tag
   ```

### Release Assets

Each release includes:
- Multi-platform binaries (Linux, macOS, Windows)
- SHA256 checksums for verification
- Docker image published to GitHub Container Registry
- Automated release notes with installation instructions

## 🧪 Testing Strategy

### Test Types

1. **Unit Tests**: Fast, isolated tests using SQLite
2. **Integration Tests**: Database integration with PostgreSQL
3. **Protocol Compliance**: MCP specification compliance
4. **Security Tests**: Vulnerability scanning and security analysis
5. **Performance Tests**: Benchmarks and load testing
6. **Cross-Platform Tests**: Compatibility across operating systems

### Running Tests Locally

```bash
# All tests
./scripts/mcp-dev.sh test

# Specific test types
go test -v ./cmd/mcp-server/...           # Unit tests
go test -v ./tests/mcp/...                # Integration tests
go test -bench=. ./cmd/mcp-server/...     # Benchmarks
```

### CI/CD Test Execution

- **Every commit**: Unit tests and basic validation
- **Pull requests**: Full test suite including integration tests
- **Main branch**: All tests including performance benchmarks
- **Releases**: Complete test suite with security scanning

## 📊 Monitoring & Metrics

### Build Status

Monitor build status through:
- GitHub Actions dashboard
- Build badges (add to main README)
- Automated notifications on failures

### Test Coverage

- Unit test coverage uploaded to Codecov
- Coverage reports available as GitHub artifacts
- Coverage trends tracked over time

### Performance Monitoring

- Benchmark results stored as artifacts
- Performance regression detection
- Load testing results for capacity planning

## 🔒 Security

### Security Scanning

- **Gosec**: Static security analysis
- **govulncheck**: Vulnerability database checking
- **Dependency scanning**: Automated dependency vulnerability checks
- **SARIF reports**: Security findings uploaded to GitHub Security tab

### Security Best Practices

- Minimal Docker images with security updates
- No secrets in code or configuration files
- Secure defaults in configuration
- Regular dependency updates

## 🐛 Troubleshooting

### Common Issues

1. **Build failures:**
   ```bash
   # Check Go version
   go version
   
   # Clean and rebuild
   ./scripts/mcp-dev.sh clean
   ./scripts/mcp-dev.sh build
   ```

2. **Test failures:**
   ```bash
   # Run tests with verbose output
   go test -v ./cmd/mcp-server/...
   
   # Check database connectivity
   make docker-up
   ```

3. **Protocol issues:**
   ```bash
   # Test protocol compliance
   ./scripts/mcp-dev.sh protocol-test
   
   # Check server logs
   LOG_LEVEL=debug ./bin/requirements-mcp-server --stdio
   ```

### Getting Help

- Check GitHub Issues for known problems
- Review GitHub Actions logs for build failures
- Use the development script for common tasks
- Consult the main project documentation

## 📚 Additional Resources

- [MCP Specification](https://modelcontextprotocol.io/specification/)
- [Project Documentation](./docs/)
- [API Documentation](./docs/generated/)
- [Contributing Guidelines](./CONTRIBUTING.md)
- [Security Policy](./SECURITY.md)

---

This CI/CD setup ensures reliable, secure, and efficient delivery of the MCP server across multiple platforms while maintaining high code quality through comprehensive testing.
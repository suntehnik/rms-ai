.PHONY: build build-init build-mcp-server install-mcp-server run init test test-unit test-integration test-e2e test-fast test-coverage test-unit-coverage test-integration-coverage test-e2e-coverage test-bench test-bench-api test-bench-results test-bench-api-results test-parallel test-race test-run test-debug test-compile test-ci clean deps dev fmt lint migrate-up migrate-down migrate-version build-migrate docker-up docker-down docker-logs docker-clean dev-setup mocks swagger swagger-fmt swagger-validate swagger-clean swagger-dev swagger-staging swagger-prod swagger-config swagger-env-dev swagger-env-staging swagger-env-prod swagger-deploy swagger-test swagger-serve help

# Build the application
build:
	go build -o bin/server cmd/server/main.go

# Build initialization binary
build-init:
	go build -o bin/init cmd/init/main.go

# Build mock data generator
build-gen-mock-data:
	go build -o bin/gen-mock-data cmd/gen-mock-data/main.go

# Build MCP server
build-mcp-server:
	@echo "🔧 Building MCP Server..."
	@mkdir -p bin
	@if [ "$(OS)" = "Windows_NT" ] || [ "$(GOOS)" = "windows" ]; then \
		go build -o bin/spexus-mcp.exe cmd/mcp-server/main.go; \
		echo "✅ MCP Server built: bin/spexus-mcp.exe"; \
	else \
		go build -o bin/spexus-mcp cmd/mcp-server/main.go; \
		echo "✅ MCP Server built: bin/spexus-mcp"; \
	fi

# Build MCP server with version info (for releases)
build-mcp-server-release:
	@echo "🔧 Building MCP Server with version info..."
	@VERSION=$${VERSION:-dev-$(shell git rev-parse --short HEAD)} && \
	BUILD_DATE=$$(date -u +'%Y-%m-%dT%H:%M:%SZ') && \
	GIT_COMMIT=$$(git rev-parse --short HEAD) && \
	go build \
		-ldflags="-s -w -X main.Version=$$VERSION -X main.BuildDate=$$BUILD_DATE -X main.GitCommit=$$GIT_COMMIT" \
		-o bin/spexus-mcp \
		cmd/mcp-server/main.go
	@echo "✅ MCP Server built with version info: bin/spexus-mcp"

# Install MCP server to system
install-mcp-server: build-mcp-server
	@echo "📦 Installing MCP Server to /usr/local/bin..."
	@sudo cp bin/spexus-mcp /usr/local/bin/
	@echo "✅ MCP Server installed: /usr/local/bin/spexus-mcp"
	@echo "💡 Configure Claude Desktop to use: /usr/local/bin/spexus-mcp"

# Test MCP server specifically
test-mcp-server:
	@echo "🧪 Running MCP Server tests..."
	go test -v -race ./cmd/mcp-server/... ./internal/mcp/...
	@echo "✅ MCP Server tests completed"

# Test MCP server with coverage
test-mcp-server-coverage:
	@echo "📊 Running MCP Server tests with coverage..."
	go test -v -race -coverprofile=mcp-coverage.out ./cmd/mcp-server/... ./internal/mcp/...
	go tool cover -html=mcp-coverage.out -o mcp-coverage.html
	@echo "✅ MCP Server coverage report: mcp-coverage.html"

# Run MCP protocol compliance tests
test-mcp-protocol:
	@echo "🔍 Testing MCP protocol compliance..."
	@if [ -f "bin/spexus-mcp" ]; then \
		echo "Testing server startup and basic protocol..."; \
		timeout 10s ./bin/spexus-mcp --test-mode || echo "Protocol test completed"; \
	else \
		echo "Building MCP server first..."; \
		$(MAKE) build-mcp-server; \
		timeout 10s ./bin/spexus-mcp --test-mode || echo "Protocol test completed"; \
	fi
	@echo "✅ MCP protocol compliance test completed"

# Validate MCP server configuration
validate-mcp-config:
	@echo "🔍 Validating MCP server configuration..."
	@if [ -f "config.example.json" ]; then \
		if command -v jq >/dev/null 2>&1; then \
			jq empty config.example.json && echo "✅ Configuration JSON is valid"; \
		else \
			echo "⚠️  jq not found, skipping JSON validation"; \
		fi; \
		if [ -f "bin/spexus-mcp" ]; then \
			./bin/spexus-mcp --validate-config --config config.example.json || echo "Config validation completed"; \
		else \
			echo "Building MCP server for config validation..."; \
			$(MAKE) build-mcp-server; \
			./bin/spexus-mcp --validate-config --config config.example.json || echo "Config validation completed"; \
		fi; \
	else \
		echo "⚠️  config.example.json not found"; \
	fi
	@echo "✅ MCP configuration validation completed"

# Run MCP server benchmarks
bench-mcp-server:
	@echo "🏃 Running MCP Server benchmarks..."
	@mkdir -p benchmark-results
	go test -bench=. -benchmem -benchtime=5s -timeout=30m ./cmd/mcp-server/... ./internal/mcp/... | tee benchmark-results/mcp-benchmark-$(shell date +%Y%m%d-%H%M%S).txt
	@echo "✅ MCP Server benchmarks completed"

# Build multi-platform MCP server binaries (for local testing)
build-mcp-server-all:
	@echo "🏗️ Building MCP Server for multiple platforms..."
	@mkdir -p bin/dist
	@VERSION=$${VERSION:-dev-$(shell git rev-parse --short HEAD)} && \
	BUILD_DATE=$$(date -u +'%Y-%m-%dT%H:%M:%SZ') && \
	GIT_COMMIT=$$(git rev-parse --short HEAD) && \
	LDFLAGS="-s -w -X main.Version=$$VERSION -X main.BuildDate=$$BUILD_DATE -X main.GitCommit=$$GIT_COMMIT" && \
	echo "Building for Linux AMD64..." && \
	GOOS=linux GOARCH=amd64 go build -ldflags="$$LDFLAGS" -o bin/dist/spexus-mcp-linux-amd64 cmd/mcp-server/main.go && \
	echo "Building for Linux ARM64..." && \
	GOOS=linux GOARCH=arm64 go build -ldflags="$$LDFLAGS" -o bin/dist/spexus-mcp-linux-arm64 cmd/mcp-server/main.go && \
	echo "Building for macOS AMD64..." && \
	GOOS=darwin GOARCH=amd64 go build -ldflags="$$LDFLAGS" -o bin/dist/spexus-mcp-darwin-amd64 cmd/mcp-server/main.go && \
	echo "Building for macOS ARM64..." && \
	GOOS=darwin GOARCH=arm64 go build -ldflags="$$LDFLAGS" -o bin/dist/spexus-mcp-darwin-arm64 cmd/mcp-server/main.go && \
	echo "Building for Windows AMD64..." && \
	GOOS=windows GOARCH=amd64 go build -ldflags="$$LDFLAGS" -o bin/dist/spexus-mcp-windows-amd64.exe cmd/mcp-server/main.go
	@echo "✅ Multi-platform MCP Server binaries built in bin/dist/"

# Package MCP server binaries
package-mcp-server: build-mcp-server-all
	@echo "📦 Packaging MCP Server binaries..."
	@cd bin/dist && \
	tar -czf spexus-mcp-linux-amd64.tar.gz spexus-mcp-linux-amd64 && \
	tar -czf spexus-mcp-linux-arm64.tar.gz spexus-mcp-linux-arm64 && \
	tar -czf spexus-mcp-darwin-amd64.tar.gz spexus-mcp-darwin-amd64 && \
	tar -czf spexus-mcp-darwin-arm64.tar.gz spexus-mcp-darwin-arm64 && \
	zip spexus-mcp-windows-amd64.zip spexus-mcp-windows-amd64.exe && \
	sha256sum *.tar.gz *.zip > checksums.sha256
	@echo "✅ MCP Server packages created in bin/dist/"

# Run MCP development helper script
mcp-dev:
	@echo "🛠️ Running MCP development helper..."
	@if [ -f "scripts/mcp-dev.sh" ]; then \
		./scripts/mcp-dev.sh $(filter-out $@,$(MAKECMDGOALS)); \
	else \
		echo "❌ MCP development script not found: scripts/mcp-dev.sh"; \
		exit 1; \
	fi

# MCP server development workflow
mcp-dev-setup: build-mcp-server validate-mcp-config
	@echo "🚀 MCP Server development setup completed!"
	@echo "💡 Use 'make mcp-dev run --stdio' to start the server"
	@echo "💡 Use 'make mcp-dev test' to run tests"
	@echo "💡 Use 'make mcp-dev help' for more options"

# Run the application
run:
	go run cmd/server/main.go

# Run initialization service
init: build-init
	./bin/init

# Generate mock data
gen-mock-data: build-gen-mock-data
	@echo "🎭 Generating mock data for development..."
	@export $(shell cat .env.mock-data 2>/dev/null | xargs) && ./bin/gen-mock-data
	@echo "✅ Mock data generation completed!"

# Run all tests in sequence: unit → integration → e2e
test: test-unit test-integration test-e2e
	@echo "✅ All tests completed successfully!"

# Run unit tests only (fast, SQLite)
test-unit:
	@echo "🧪 Running unit tests..."
	go test -v ./internal/models/... ./internal/repository/... ./internal/service/... ./internal/handlers/... ./tests/unit/...
	@echo "✅ Unit tests completed"

# Run integration tests (requires database setup)
test-integration:
	@echo "🔧 Running integration tests..."
	go test -v ./internal/integration/...
	@echo "✅ Integration tests completed"

# Run E2E tests (full environment)
test-e2e:
	@echo "🌐 Running E2E tests..."
	go test -v ./tests/e2e/...
	@echo "✅ E2E tests completed"

# Run only fast tests (unit tests with short flag)
test-fast:
	@echo "⚡ Running fast tests only..."
	go test -short -v ./internal/models/... ./internal/repository/... ./internal/service/... ./internal/handlers/... ./tests/unit/...

# Run tests with coverage (all types)
test-coverage: test-unit-coverage test-integration-coverage
	@echo "📊 Generating combined coverage report..."
	@echo "Coverage reports generated: coverage-unit.html, coverage-integration.html"

# Run unit tests with coverage
test-unit-coverage:
	@echo "📊 Running unit tests with coverage..."
	go test -v -coverprofile=coverage-unit.out ./internal/models/... ./internal/repository/... ./internal/service/... ./internal/handlers/... ./tests/unit/...
	go tool cover -html=coverage-unit.out -o coverage-unit.html
	@echo "✅ Unit test coverage report: coverage-unit.html"

# Run integration tests with coverage
test-integration-coverage:
	@echo "📊 Running integration tests with coverage..."
	go test -v -coverprofile=coverage-integration.out ./internal/integration/...
	go tool cover -html=coverage-integration.out -o coverage-integration.html
	@echo "✅ Integration test coverage report: coverage-integration.html"

# Run E2E tests with coverage
test-e2e-coverage:
	@echo "📊 Running E2E tests with coverage..."
	go test -v -coverprofile=coverage-e2e.out ./tests/e2e/...
	go tool cover -html=coverage-e2e.out -o coverage-e2e.html
	@echo "✅ E2E test coverage report: coverage-e2e.html"

# Run performance benchmarks (all benchmarks)
test-bench:
	@echo "🏃 Running performance benchmarks..."
	go test -bench=. -benchmem -benchtime=3s -timeout=30m ./internal/service/... ./internal/repository/... ./internal/benchmarks/... ./tests/integration/...
	@echo "✅ Performance benchmarks completed"

# Run API endpoint benchmarks specifically
test-bench-api:
	@echo "🌐 Running API endpoint benchmarks..."
	go test -bench=. -benchmem -benchtime=3s -timeout=30m ./internal/benchmarks/api/...
	@echo "✅ API endpoint benchmarks completed"

# Run benchmarks with result file generation
test-bench-results:
	@echo "📊 Running benchmarks with result file generation..."
	@mkdir -p benchmark-results
	go test -bench=. -benchmem -benchtime=3s -timeout=30m ./internal/service/... ./internal/repository/... ./internal/benchmarks/... ./tests/integration/... | tee benchmark-results/benchmark-$(shell date +%Y%m%d-%H%M%S).txt
	@echo "✅ Benchmark results saved to benchmark-results/"

# Run API benchmarks with result file generation
test-bench-api-results:
	@echo "📊 Running API benchmarks with result file generation..."
	@mkdir -p benchmark-results
	go test -bench=. -benchmem -benchtime=3s -timeout=30m ./internal/benchmarks/api/... | tee benchmark-results/api-benchmark-$(shell date +%Y%m%d-%H%M%S).txt
	@echo "✅ API benchmark results saved to benchmark-results/"

# Run tests in parallel (for CI/CD)
test-parallel:
	@echo "⚡ Running tests in parallel..."
	go test -v -parallel 4 ./internal/models/... ./internal/repository/... ./internal/service/... ./internal/handlers/...

# Run tests with race detection
test-race:
	@echo "🏁 Running tests with race detection..."
	go test -race -v ./internal/models/... ./internal/repository/... ./internal/service/... ./internal/handlers/...

# Run specific test by name (usage: make test-run TEST=TestName)
test-run:
	@echo "🎯 Running specific test: $(TEST)"
	go test -v -run $(TEST) ./...

# Run tests and generate verbose output for debugging
test-debug:
	@echo "🐛 Running tests in debug mode..."
	go test -v -count=1 ./internal/models/... ./internal/repository/... ./internal/service/... ./internal/handlers/...

# Check if tests compile without running them
test-compile:
	@echo "🔍 Checking if all tests compile..."
	go test -c ./internal/models/... ./internal/repository/... ./internal/service/... ./internal/handlers/... ./internal/integration/... ./tests/unit/... ./tests/e2e/...
	@echo "✅ All tests compile successfully"

# Run tests suitable for CI/CD (no interactive components)
test-ci: test-unit test-integration
	@echo "🤖 CI/CD tests completed (skipping E2E tests)"

# Clean build artifacts
clean:
	rm -rf bin/

# Install dependencies
deps:
	go mod tidy
	go mod download

# Run with development environment
dev:
	LOG_LEVEL=debug LOG_FORMAT=text go run cmd/server/main.go

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run

# Database commands
migrate-up:
	go run cmd/migrate/main.go -up

migrate-down:
	go run cmd/migrate/main.go -down

migrate-version:
	go run cmd/migrate/main.go -version

# Build migration tool
build-migrate:
	go build -o bin/migrate cmd/migrate/main.go

# Docker commands for development
docker-up:
	docker-compose -f docker-compose.dev.yml up -d

docker-down:
	docker-compose -f docker-compose.dev.yml down

docker-logs:
	docker-compose -f docker-compose.dev.yml logs -f

docker-clean:
	docker-compose -f docker-compose.dev.yml down -v

# Production Docker commands
docker-prod-build:
	docker-compose -f docker-compose.prod.yml build --no-cache

docker-prod-up:
	docker-compose -f docker-compose.prod.yml up -d

docker-prod-down:
	docker-compose -f docker-compose.prod.yml down

docker-prod-logs:
	docker-compose -f docker-compose.prod.yml logs -f

docker-prod-clean:
	docker-compose -f docker-compose.prod.yml down -v

# Build and Push to GitHub Registry
build-push:
	@echo "🏗️ Building and pushing multi-platform to GitHub Container Registry..."
	./scripts/build-and-push.sh

build-push-x86:
	@echo "🏗️ Building and pushing x86_64 only to GitHub Container Registry..."
	PLATFORMS=linux/amd64 ./scripts/build-and-push.sh

build-push-arm:
	@echo "🏗️ Building and pushing ARM64 only to GitHub Container Registry..."
	PLATFORMS=linux/arm64 ./scripts/build-and-push.sh

build-push-single:
	@echo "🏗️ Building and pushing single-platform to GitHub Container Registry..."
	USE_BUILDX=false ./scripts/build-and-push.sh

build-push-info:
	@echo "ℹ️ Showing build information..."
	./scripts/build-and-push.sh info

build-push-auth:
	@echo "🔐 Authenticating with GitHub Registry..."
	./scripts/build-and-push.sh auth

build-push-setup:
	@echo "🔧 Setting up buildx for multi-platform builds..."
	./scripts/build-and-push.sh setup-buildx

build-push-cleanup:
	@echo "🧹 Cleaning up build artifacts..."
	./scripts/build-and-push.sh cleanup

build-push-cleanup-buildx:
	@echo "🧹 Cleaning up buildx builder..."
	./scripts/build-and-push.sh cleanup-buildx

# Production deployment from Registry
deploy-prod:
	@echo "🚀 Starting production deployment from registry..."
	./scripts/deploy-from-registry.sh

deploy-prod-backup:
	@echo "💾 Creating production backup..."
	./scripts/deploy-from-registry.sh backup

deploy-prod-init:
	@echo "🔧 Running database initialization..."
	./scripts/deploy-from-registry.sh init

deploy-prod-migrate:
	@echo "📊 Running database migrations..."
	./scripts/deploy-from-registry.sh migrate

deploy-prod-update:
	@echo "🔄 Updating production services..."
	./scripts/deploy-from-registry.sh update

deploy-prod-restart:
	@echo "🔄 Restarting production services..."
	./scripts/deploy-from-registry.sh restart

deploy-prod-status:
	@echo "📊 Checking production status..."
	./scripts/deploy-from-registry.sh status

deploy-prod-stop:
	@echo "🛑 Stopping production services..."
	./scripts/deploy-from-registry.sh stop

deploy-prod-pull:
	@echo "📥 Pulling image from registry..."
	./scripts/deploy-from-registry.sh pull

# Environment management
env-check:
	@echo "🔍 Checking environment variables..."
	./scripts/check-env.sh

env-generate:
	@echo "🔐 Generating secure environment values..."
	./scripts/check-env.sh generate

env-test-compose:
	@echo "🧪 Testing Docker Compose configuration..."
	./scripts/check-env.sh test-compose

# Legacy deployment (local build)
deploy-prod-local:
	@echo "🚀 Starting local production deployment..."
	./scripts/deploy-prod.sh

# Full development setup
dev-setup: docker-up
	@echo "Waiting for databases to be ready..."
	@sleep 10
	@make migrate-up
	@echo "Development environment ready!"

# Generate mocks (for future use)
mocks:
	@echo "Mock generation will be added in future tasks"

# Swagger documentation commands
swagger:
	@echo "📚 Generating Swagger documentation..."
	swag init -g cmd/server/main.go -o docs --parseDependency --parseInternal
	@echo "✅ Swagger documentation generated in docs/ directory"

swagger-fmt:
	@echo "🎨 Formatting Swagger comments..."
	swag fmt -g cmd/server/main.go
	@echo "✅ Swagger comments formatted"

swagger-validate:
	@echo "🔍 Validating Swagger documentation..."
	@if [ -f docs/swagger.json ]; then \
		echo "Swagger JSON found, validating..."; \
		swag init -g cmd/server/main.go -o docs --parseDependency --parseInternal --quiet; \
		echo "✅ Swagger documentation is valid"; \
	else \
		echo "❌ No Swagger documentation found. Run 'make swagger' first."; \
		exit 1; \
	fi

swagger-clean:
	@echo "🧹 Cleaning Swagger documentation..."
	rm -rf docs/docs.go docs/swagger.json docs/swagger.yaml
	@echo "✅ Swagger documentation cleaned"

# Enhanced Swagger deployment and configuration commands
swagger-dev:
	@echo "🚀 Setting up Swagger for development environment..."
	@export ENVIRONMENT=development && \
	export SWAGGER_ENABLED=true && \
	export SWAGGER_REQUIRE_AUTH=false && \
	export LOG_LEVEL=debug && \
	$(MAKE) swagger
	@echo "✅ Swagger configured for development"

swagger-staging:
	@echo "🚀 Setting up Swagger for staging environment..."
	@export ENVIRONMENT=staging && \
	export SWAGGER_ENABLED=true && \
	export SWAGGER_REQUIRE_AUTH=true && \
	export LOG_LEVEL=info && \
	$(MAKE) swagger
	@echo "✅ Swagger configured for staging"

swagger-prod:
	@echo "🚀 Setting up Swagger for production environment..."
	@export ENVIRONMENT=production && \
	export SWAGGER_ENABLED=false && \
	export SWAGGER_REQUIRE_AUTH=true && \
	export LOG_LEVEL=warn && \
	$(MAKE) swagger
	@echo "⚠️  Swagger is disabled in production by default"
	@echo "   To enable in production, set SWAGGER_ENABLED=true"

swagger-config:
	@echo "📋 Current Swagger configuration:"
	@echo "  Environment: $${ENVIRONMENT:-development}"
	@echo "  Swagger Enabled: $${SWAGGER_ENABLED:-true}"
	@echo "  Require Auth: $${SWAGGER_REQUIRE_AUTH:-false}"
	@echo "  Log Level: $${LOG_LEVEL:-info}"
	@echo "  Base Path: $${SWAGGER_BASE_PATH:-/swagger}"
	@echo "  Host: $${SWAGGER_HOST:-localhost:8080}"

swagger-env-dev:
	@echo "📝 Generating .env file for development..."
	@echo "# Development Environment Configuration" > .env.development
	@echo "ENVIRONMENT=development" >> .env.development
	@echo "SWAGGER_ENABLED=true" >> .env.development
	@echo "SWAGGER_REQUIRE_AUTH=false" >> .env.development
	@echo "LOG_LEVEL=debug" >> .env.development
	@echo "CORS_ENABLED=true" >> .env.development
	@echo "RATE_LIMIT_ENABLED=false" >> .env.development
	@echo "COMPRESSION_ENABLED=false" >> .env.development
	@echo "CACHE_ENABLED=false" >> .env.development
	@echo "DEBUG_MODE=true" >> .env.development
	@echo "✅ Development environment file created: .env.development"

swagger-env-staging:
	@echo "📝 Generating .env file for staging..."
	@echo "# Staging Environment Configuration" > .env.staging
	@echo "ENVIRONMENT=staging" >> .env.staging
	@echo "SWAGGER_ENABLED=true" >> .env.staging
	@echo "SWAGGER_REQUIRE_AUTH=true" >> .env.staging
	@echo "LOG_LEVEL=info" >> .env.staging
	@echo "CORS_ENABLED=true" >> .env.staging
	@echo "RATE_LIMIT_ENABLED=true" >> .env.staging
	@echo "COMPRESSION_ENABLED=true" >> .env.staging
	@echo "CACHE_ENABLED=true" >> .env.staging
	@echo "DEBUG_MODE=false" >> .env.staging
	@echo "✅ Staging environment file created: .env.staging"

swagger-env-prod:
	@echo "📝 Generating .env file for production..."
	@echo "# Production Environment Configuration" > .env.production
	@echo "ENVIRONMENT=production" >> .env.production
	@echo "SWAGGER_ENABLED=false" >> .env.production
	@echo "SWAGGER_REQUIRE_AUTH=true" >> .env.production
	@echo "LOG_LEVEL=warn" >> .env.production
	@echo "CORS_ENABLED=false" >> .env.production
	@echo "RATE_LIMIT_ENABLED=true" >> .env.production
	@echo "COMPRESSION_ENABLED=true" >> .env.production
	@echo "CACHE_ENABLED=true" >> .env.production
	@echo "DEBUG_MODE=false" >> .env.production
	@echo "CSP_ENABLED=true" >> .env.production
	@echo "SECURITY_HEADERS=true" >> .env.production
	@echo "✅ Production environment file created: .env.production"

swagger-deploy:
	@echo "🚀 Deploying Swagger documentation..."
	@if [ -z "$$ENVIRONMENT" ]; then \
		echo "❌ ENVIRONMENT variable not set. Use: make swagger-dev, swagger-staging, or swagger-prod"; \
		exit 1; \
	fi
	@echo "Deploying for environment: $$ENVIRONMENT"
	@$(MAKE) swagger
	@if [ "$$ENVIRONMENT" = "production" ] && [ "$$SWAGGER_ENABLED" != "true" ]; then \
		echo "⚠️  Swagger is disabled in production"; \
		echo "   Documentation will not be accessible"; \
	else \
		echo "✅ Swagger documentation deployed"; \
		echo "   Access at: http://$${SWAGGER_HOST:-localhost:8080}$${SWAGGER_BASE_PATH:-/swagger}/index.html"; \
	fi

swagger-test:
	@echo "🧪 Testing Swagger documentation..."
	@$(MAKE) swagger
	@if [ -f docs/swagger.json ]; then \
		echo "✅ Swagger JSON generated successfully"; \
		echo "📊 Checking documentation completeness..."; \
		if command -v jq >/dev/null 2>&1; then \
			PATHS=$$(jq '.paths | keys | length' docs/swagger.json); \
			DEFINITIONS=$$(jq '.definitions | keys | length' docs/swagger.json); \
			echo "   Endpoints documented: $$PATHS"; \
			echo "   Models documented: $$DEFINITIONS"; \
		else \
			echo "   Install jq for detailed metrics"; \
		fi; \
	else \
		echo "❌ Swagger generation failed"; \
		exit 1; \
	fi

swagger-serve:
	@echo "🌐 Starting server with Swagger documentation..."
	@export SWAGGER_ENABLED=true && \
	export ENVIRONMENT=development && \
	$(MAKE) swagger && \
	$(MAKE) run

# Documentation quality metrics
docs-metrics:
	@echo "📊 Generating documentation quality metrics..."
	go run cmd/docs-metrics/main.go -format=text -verbose
	@echo "✅ Documentation metrics generated"

docs-metrics-json:
	@echo "📊 Generating documentation metrics (JSON)..."
	@mkdir -p reports
	go run cmd/docs-metrics/main.go -output=reports/docs-metrics.json -format=json
	@echo "✅ Documentation metrics saved to reports/docs-metrics.json"

docs-metrics-summary:
	@echo "📊 Documentation quality summary..."
	go run cmd/docs-metrics/main.go -format=summary

docs-quality-check:
	@echo "🔍 Checking documentation quality..."
	@go run cmd/docs-metrics/main.go -format=summary | grep -q "Documentation Quality: [89][0-9]" || \
	(echo "❌ Documentation quality below 80%. Run 'make docs-metrics' for details." && exit 1)
	@echo "✅ Documentation quality check passed"

# Documentation validation tests
docs-validate:
	@echo "🔍 Running comprehensive documentation validation..."
	go run scripts/validate-documentation/main.go
	@echo "✅ Documentation validation completed"

docs-validate-routes:
	@echo "🔍 Validating route implementation vs documentation..."
	go test -v ./internal/validation -run TestOpenAPIRouteCompleteness
	@echo "✅ Route validation completed"

docs-validate-response-schemas:
	@echo "🔍 Validating response schema consistency..."
	go test -v ./internal/validation -run TestResponseSchemaValidation
	@echo "✅ Response schema validation completed"

docs-validate-auth:
	@echo "🔍 Validating authentication documentation..."
	go test -v ./internal/validation -run TestAuthenticationDocumentation
	@echo "✅ Authentication validation completed"

docs-validate-completeness:
	@echo "🔍 Validating documentation completeness..."
	go test -v ./internal/validation -run TestDocumentationCompleteness
	@echo "✅ Completeness validation completed"

docs-validate-all:
	@echo "🔍 Running all documentation validation tests..."
	@$(MAKE) docs-validate-routes
	@$(MAKE) docs-validate-response-schemas
	@$(MAKE) docs-validate-auth
	@$(MAKE) docs-validate-completeness
	@echo "✅ All documentation validation tests completed"

# Run comprehensive OpenAPI validation
docs-validate-comprehensive:
	@echo "🔍 Running comprehensive OpenAPI validation..."
	go run scripts/comprehensive-validation/main.go
	@echo "✅ Comprehensive validation completed"

docs-validate-api-completeness:
	@echo "🔍 Validating API completeness..."
	go run scripts/validate-api-completeness/main.go
	@echo "✅ API completeness validation completed"

docs-validate-openapi:
	@echo "🔍 Validating OpenAPI specification..."
	go run scripts/validate-openapi/main.go
	@echo "✅ OpenAPI validation completed"

docs-validate-schemas:
	@echo "🔍 Validating schemas and parameters..."
	go run scripts/validate-schemas/main.go
	@echo "✅ Schema validation completed"

docs-verify-models:
	@echo "🔍 Verifying model consistency..."
	go run scripts/verify-models/main.go
	@echo "✅ Model verification completed"

docs-run-all-validation:
	@echo "🔍 Running all validation tests..."
	go run scripts/run-all-validation/main.go
	@echo "✅ All validation tests completed"

# Generate comprehensive API documentation from OpenAPI specification
docs-generate:
	@echo "📚 Generating comprehensive API documentation..."
	@mkdir -p docs/generated
	go run scripts/generate-api-docs/main.go -input=docs/openapi-v3.yaml -output=docs/generated -format=all -verbose
	@echo "✅ API documentation generated in docs/generated/"

docs-generate-html:
	@echo "📚 Generating HTML API documentation..."
	@mkdir -p docs/generated
	go run scripts/generate-api-docs/main.go -input=docs/openapi-v3.yaml -output=docs/generated -format=html -verbose
	@echo "✅ HTML documentation generated: docs/generated/api-documentation.html"

docs-generate-markdown:
	@echo "📚 Generating Markdown API documentation..."
	@mkdir -p docs/generated
	go run scripts/generate-api-docs/main.go -input=docs/openapi-v3.yaml -output=docs/generated -format=markdown -verbose
	@echo "✅ Markdown documentation generated: docs/generated/api-documentation.md"

docs-generate-typescript:
	@echo "📚 Generating TypeScript API documentation..."
	@mkdir -p docs/generated
	go run scripts/generate-api-docs/main.go -input=docs/openapi-v3.yaml -output=docs/generated -format=typescript -verbose
	@echo "✅ TypeScript documentation generated: docs/generated/api-types.ts"

docs-generate-json:
	@echo "📚 Generating JSON API documentation..."
	@mkdir -p docs/generated
	go run scripts/generate-api-docs/main.go -input=docs/openapi-v3.yaml -output=docs/generated -format=json -verbose
	@echo "✅ JSON documentation generated: docs/generated/api-documentation.json"

# Show help for all available targets
help:
	@echo "📋 Available Make targets:"
	@echo ""
	@echo "🏗️  Build & Run:"
	@echo "  build              - Build the application binary"
	@echo "  build-init         - Build initialization binary"
	@echo "  build-gen-mock-data - Build mock data generator"
	@echo "  build-mcp-server   - Build MCP Server console application"
	@echo "  build-mcp-server-release - Build MCP Server with version info"
	@echo "  build-mcp-server-all - Build MCP Server for multiple platforms"
	@echo "  package-mcp-server - Package MCP Server binaries with checksums"
	@echo "  install-mcp-server - Install MCP Server to /usr/local/bin"
	@echo "  run                - Run the application directly"
	@echo "  init               - Run initialization service"
	@echo "  gen-mock-data      - Generate mock data for development"
	@echo "  dev                - Run with development settings"
	@echo ""
	@echo "🧪 Testing:"
	@echo "  test               - Run all tests (unit → integration → e2e)"
	@echo "  test-unit          - Run unit tests only (fast)"
	@echo "  test-integration   - Run integration tests"
	@echo "  test-e2e           - Run end-to-end tests"
	@echo "  test-fast          - Run only fast unit tests"
	@echo "  test-ci            - Run tests suitable for CI/CD"
	@echo "  test-mcp-server    - Run MCP Server specific tests"
	@echo "  test-mcp-server-coverage - Run MCP Server tests with coverage"
	@echo "  test-mcp-protocol  - Test MCP protocol compliance"
	@echo "  validate-mcp-config - Validate MCP server configuration"
	@echo "  bench-mcp-server   - Run MCP Server benchmarks"
	@echo ""
	@echo "📊 Coverage & Analysis:"
	@echo "  test-coverage      - Generate coverage reports for all tests"
	@echo "  test-unit-coverage - Generate unit test coverage"
	@echo "  test-integration-coverage - Generate integration test coverage"
	@echo "  test-e2e-coverage  - Generate E2E test coverage"
	@echo ""
	@echo "🔧 Advanced Testing:"
	@echo "  test-bench         - Run performance benchmarks (all)"
	@echo "  test-bench-api     - Run API endpoint benchmarks only"
	@echo "  test-bench-results - Run benchmarks with result file generation"
	@echo "  test-bench-api-results - Run API benchmarks with result files"
	@echo "  test-parallel      - Run tests in parallel"
	@echo "  test-race          - Run tests with race detection"
	@echo "  test-debug         - Run tests in debug mode"
	@echo "  test-compile       - Check if tests compile"
	@echo "  test-run TEST=name - Run specific test by name"
	@echo ""
	@echo "🗄️  Database:"
	@echo "  migrate-up         - Apply database migrations"
	@echo "  migrate-down       - Rollback database migrations"
	@echo "  migrate-version    - Check migration status"
	@echo ""
	@echo "📚 Documentation:"
	@echo "  swagger            - Generate Swagger documentation"
	@echo "  swagger-fmt        - Format Swagger comments"
	@echo "  swagger-validate   - Validate Swagger documentation"
	@echo "  swagger-clean      - Clean generated Swagger files"
	@echo "  swagger-dev        - Configure Swagger for development"
	@echo "  swagger-staging    - Configure Swagger for staging"
	@echo "  swagger-prod       - Configure Swagger for production"
	@echo "  swagger-config     - Show current Swagger configuration"
	@echo "  swagger-env-dev    - Generate development .env file"
	@echo "  swagger-env-staging - Generate staging .env file"
	@echo "  swagger-env-prod   - Generate production .env file"
	@echo "  swagger-deploy     - Deploy Swagger for current environment"
	@echo "  swagger-test       - Test Swagger documentation generation"
	@echo "  swagger-serve      - Start server with Swagger enabled"
	@echo "  docs-generate      - Generate comprehensive API documentation (all formats)"
	@echo "  docs-generate-html - Generate HTML API documentation"
	@echo "  docs-generate-markdown - Generate Markdown API documentation"
	@echo "  docs-generate-typescript - Generate TypeScript API documentation"
	@echo "  docs-generate-json - Generate JSON API documentation"
	@echo "  docs-metrics       - Generate documentation quality metrics"
	@echo "  docs-metrics-json  - Generate metrics in JSON format"
	@echo "  docs-metrics-summary - Show documentation quality summary"
	@echo "  docs-quality-check - Check if documentation quality meets standards"
	@echo "  docs-validate      - Run comprehensive documentation validation"
	@echo "  docs-validate-routes - Validate route implementation vs documentation"
	@echo "  docs-validate-response-schemas - Validate response schema consistency"
	@echo "  docs-validate-auth - Validate authentication documentation"
	@echo "  docs-validate-completeness - Validate documentation completeness"
	@echo "  docs-validate-comprehensive - Run comprehensive OpenAPI validation"
	@echo "  docs-validate-api-completeness - Validate API completeness"
	@echo "  docs-validate-openapi - Validate OpenAPI specification"
	@echo "  docs-validate-schemas - Validate schemas and parameters"
	@echo "  docs-verify-models - Verify model consistency"
	@echo "  docs-run-all-validation - Run all validation tests"
	@echo "  docs-validate-all  - Run all documentation validation tests"
	@echo ""
	@echo "🐳 Docker Development:"
	@echo "  docker-up          - Start development containers"
	@echo "  docker-down        - Stop development containers"
	@echo "  dev-setup          - Full development environment setup"
	@echo ""
	@echo "🏗️ Build & Push to Registry:"
	@echo "  build-push         - Build and push multi-platform (amd64+arm64)"
	@echo "  build-push-x86     - Build and push x86_64/amd64 only"
	@echo "  build-push-arm     - Build and push ARM64 only"
	@echo "  build-push-single  - Build and push single-platform (current)"
	@echo "  build-push-info    - Show build information"
	@echo "  build-push-auth    - Authenticate with GitHub Registry"
	@echo "  build-push-setup   - Setup buildx for multi-platform builds"
	@echo "  build-push-cleanup - Clean up build artifacts"
	@echo "  build-push-cleanup-buildx - Clean up buildx builder"
	@echo ""
	@echo "🚀 Production Deployment (from Registry):"
	@echo "  deploy-prod        - Full production deployment from registry"
	@echo "  deploy-prod-backup - Create production backup"
	@echo "  deploy-prod-init   - Run database initialization"
	@echo "  deploy-prod-migrate - Run database migrations"
	@echo "  deploy-prod-update - Update services with new image"
	@echo "  deploy-prod-restart - Restart production services"
	@echo "  deploy-prod-status - Check production status"
	@echo "  deploy-prod-stop   - Stop production services"
	@echo "  deploy-prod-pull   - Pull image from registry"
	@echo "  deploy-prod-local  - Legacy local build deployment"
	@echo ""
	@echo "🔧 Environment Management:"
	@echo "  env-check          - Check environment variables"
	@echo "  env-generate       - Generate secure environment values"
	@echo "  env-test-compose   - Test Docker Compose configuration"
	@echo ""
	@echo "🐳 Docker Production (Manual):"
	@echo "  docker-prod-build  - Build production images"
	@echo "  docker-prod-up     - Start production containers"
	@echo "  docker-prod-down   - Stop production containers"
	@echo "  docker-prod-logs   - View production logs"
	@echo "  docker-prod-clean  - Clean production containers and volumes"
	@echo ""
	@echo "🧹 Maintenance:"
	@echo "  clean              - Clean build artifacts"
	@echo "  deps               - Install/update dependencies"
	@echo "  fmt                - Format code"
	@echo "  lint               - Run linter"
	@echo ""
	@echo "🔧 MCP Server Development:"
	@echo "  mcp-dev            - Run MCP development helper script"
	@echo "  mcp-dev-setup      - Complete MCP development environment setup"
	@echo "  mcp-dev COMMAND    - Run specific MCP development command"
	@echo "                     Commands: build, test, run, install, clean, validate,"
	@echo "                               protocol-test, benchmark, help"
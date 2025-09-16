.PHONY: build build-init run init test test-unit test-integration test-e2e test-fast test-coverage test-unit-coverage test-integration-coverage test-e2e-coverage test-bench test-bench-api test-bench-results test-bench-api-results test-parallel test-race test-run test-debug test-compile test-ci clean deps dev fmt lint migrate-up migrate-down migrate-version build-migrate docker-up docker-down docker-logs docker-clean dev-setup mocks swagger swagger-fmt swagger-validate swagger-clean swagger-dev swagger-staging swagger-prod swagger-config swagger-env-dev swagger-env-staging swagger-env-prod swagger-deploy swagger-test swagger-serve help

# Build the application
build:
	go build -o bin/server cmd/server/main.go

# Build initialization binary
build-init:
	go build -o bin/init cmd/init/main.go

# Run the application
run:
	go run cmd/server/main.go

# Run initialization service
init: build-init
	./bin/init

# Run all tests in sequence: unit → integration → e2e
test: test-unit test-integration test-e2e
	@echo "✅ All tests completed successfully!"

# Run unit tests only (fast, SQLite)
test-unit:
	@echo "🧪 Running unit tests..."
	go test -v ./internal/models/... ./internal/repository/... ./internal/service/... ./tests/unit/...
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
	go test -short -v ./internal/models/... ./internal/repository/... ./internal/service/... ./tests/unit/...

# Run tests with coverage (all types)
test-coverage: test-unit-coverage test-integration-coverage
	@echo "📊 Generating combined coverage report..."
	@echo "Coverage reports generated: coverage-unit.html, coverage-integration.html"

# Run unit tests with coverage
test-unit-coverage:
	@echo "📊 Running unit tests with coverage..."
	go test -v -coverprofile=coverage-unit.out ./internal/models/... ./internal/repository/... ./internal/service/... ./tests/unit/...
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
	go test -v -parallel 4 ./internal/models/... ./internal/repository/... ./internal/service/...

# Run tests with race detection
test-race:
	@echo "🏁 Running tests with race detection..."
	go test -race -v ./internal/models/... ./internal/repository/... ./internal/service/...

# Run specific test by name (usage: make test-run TEST=TestName)
test-run:
	@echo "🎯 Running specific test: $(TEST)"
	go test -v -run $(TEST) ./...

# Run tests and generate verbose output for debugging
test-debug:
	@echo "🐛 Running tests in debug mode..."
	go test -v -count=1 ./internal/models/... ./internal/repository/... ./internal/service/...

# Check if tests compile without running them
test-compile:
	@echo "🔍 Checking if all tests compile..."
	go test -c ./internal/models/... ./internal/repository/... ./internal/service/... ./internal/integration/... ./tests/unit/... ./tests/e2e/...
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

# Show help for all available targets
help:
	@echo "📋 Available Make targets:"
	@echo ""
	@echo "🏗️  Build & Run:"
	@echo "  build              - Build the application binary"
	@echo "  build-init         - Build initialization binary"
	@echo "  run                - Run the application directly"
	@echo "  init               - Run initialization service"
	@echo "  dev                - Run with development settings"
	@echo ""
	@echo "🧪 Testing:"
	@echo "  test               - Run all tests (unit → integration → e2e)"
	@echo "  test-unit          - Run unit tests only (fast)"
	@echo "  test-integration   - Run integration tests"
	@echo "  test-e2e           - Run end-to-end tests"
	@echo "  test-fast          - Run only fast unit tests"
	@echo "  test-ci            - Run tests suitable for CI/CD"
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
	@echo "  docs-metrics       - Generate documentation quality metrics"
	@echo "  docs-metrics-json  - Generate metrics in JSON format"
	@echo "  docs-metrics-summary - Show documentation quality summary"
	@echo "  docs-quality-check - Check if documentation quality meets standards"
	@echo ""
	@echo "🐳 Docker:"
	@echo "  docker-up          - Start development containers"
	@echo "  docker-down        - Stop development containers"
	@echo "  dev-setup          - Full development environment setup"
	@echo ""
	@echo "🧹 Maintenance:"
	@echo "  clean              - Clean build artifacts"
	@echo "  deps               - Install/update dependencies"
	@echo "  fmt                - Format code"
	@echo "  lint               - Run linter"
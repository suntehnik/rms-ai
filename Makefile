.PHONY: build run test clean deps

# Build the application
build:
	go build -o bin/server cmd/server/main.go

# Run the application
run:
	go run cmd/server/main.go

# Run tests
test:
	go test -v ./...

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
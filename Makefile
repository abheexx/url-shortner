.PHONY: help build test clean dev migrate-up migrate-down lint bench docker-build docker-run docker-stop

# Default target
help:
	@echo "Available commands:"
	@echo "  build        - Build the application"
	@echo "  test         - Run tests"
	@echo "  clean        - Clean build artifacts"
	@echo "  dev          - Run the application locally"
	@echo "  migrate-up   - Run database migrations up"
	@echo "  migrate-down - Rollback database migrations"
	@echo "  lint         - Run linter"
	@echo "  bench        - Run benchmarks"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-run   - Run with Docker Compose"
	@echo "  docker-stop  - Stop Docker Compose services"

# Build the application
build:
	@echo "Building application..."
	go build -o bin/urlshortener ./cmd/api

# Run tests
test:
	@echo "Running tests..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/ coverage.out coverage.html

# Run the application locally
dev:
	@echo "Starting application in development mode..."
	go run ./cmd/api

# Run database migrations up
migrate-up:
	@echo "Running database migrations up..."
	migrate -path migrations -database "postgres://urlshortener:password@localhost:5432/urlshortener?sslmode=disable" up

# Rollback database migrations
migrate-down:
	@echo "Rolling back database migrations..."
	migrate -path migrations -database "postgres://urlshortener:password@localhost:5432/urlshortener?sslmode=disable" down

# Run linter
lint:
	@echo "Running linter..."
	golangci-lint run

# Run benchmarks
bench:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -t urlshortener:latest .

# Run with Docker Compose
docker-run:
	@echo "Starting services with Docker Compose..."
	docker-compose -f deploy/docker-compose/docker-compose.yml up -d

# Stop Docker Compose services
docker-stop:
	@echo "Stopping Docker Compose services..."
	docker-compose -f deploy/docker-compose/docker-compose.yml down

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# Generate mock files (if using mockgen)
mocks:
	@echo "Generating mocks..."
	@if command -v mockgen > /dev/null; then \
		mockgen -source=internal/repo/interface.go -destination=internal/repo/mocks.go -package=repo; \
		mockgen -source=internal/cache/interface.go -destination=internal/cache/mocks.go -package=cache; \
	else \
		echo "mockgen not found. Install with: go install github.com/golang/mock/mockgen@latest"; \
	fi

# Run load tests with k6
load-test:
	@echo "Running load tests with k6..."
	@if command -v k6 > /dev/null; then \
		k6 run tests/load/load-test.js; \
	else \
		echo "k6 not found. Install from: https://k6.io/docs/getting-started/installation/"; \
	fi

# Seed database with sample data
seed:
	@echo "Seeding database with sample data..."
	@if [ -f "scripts/seed.sql" ]; then \
		psql -h localhost -U urlshortener -d urlshortener -f scripts/seed.sql; \
	else \
		echo "Seed script not found: scripts/seed.sql"; \
	fi

# Show application logs
logs:
	@echo "Showing application logs..."
	docker-compose -f deploy/docker-compose/docker-compose.yml logs -f api

# Show all service logs
logs-all:
	@echo "Showing all service logs..."
	docker-compose -f deploy/docker-compose/docker-compose.yml logs -f

# Check service status
status:
	@echo "Checking service status..."
	docker-compose -f deploy/docker-compose/docker-compose.yml ps

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...
	goimports -w .

# Vet code
vet:
	@echo "Vetting code..."
	go vet ./...

# Check for security vulnerabilities
security:
	@echo "Checking for security vulnerabilities..."
	@if command -v gosec > /dev/null; then \
		gosec ./...; \
	else \
		echo "gosec not found. Install with: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest"; \
	fi

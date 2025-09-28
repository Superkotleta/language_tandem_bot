# Profile Service Makefile

.PHONY: help build run test clean deps migrate-up migrate-down docker-build docker-run

# Default target
help:
	@echo "Available targets:"
	@echo "  build        - Build the profile service"
	@echo "  run          - Run the profile service"
	@echo "  test         - Run tests"
	@echo "  clean        - Clean build artifacts"
	@echo "  deps         - Download dependencies"
	@echo "  migrate-up   - Run database migrations up"
	@echo "  migrate-down - Run database migrations down"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-run   - Run Docker container"

# Build the service
build:
	@echo "Building profile service..."
	go build -o bin/profile ./cmd/profile

# Run the service
run:
	@echo "Running profile service..."
	go run ./cmd/profile

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

# Run database migrations up
migrate-up:
	@echo "Running database migrations up..."
	migrate -path ./migrations/profile -database "postgres://profile_rw:profile_pwd@localhost:5432/languagebot?sslmode=disable" up

# Run database migrations down
migrate-down:
	@echo "Running database migrations down..."
	migrate -path ./migrations/profile -database "postgres://profile_rw:profile_pwd@localhost:5432/languagebot?sslmode=disable" down

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -t profile-service .

# Run Docker container
docker-run:
	@echo "Running Docker container..."
	docker run -p 8081:8081 \
		-e DATABASE_URL="postgres://profile_rw:profile_pwd@host.docker.internal:5432/languagebot?sslmode=disable" \
		-e HTTP_PORT="8081" \
		profile-service

# Development setup
dev-setup: deps
	@echo "Setting up development environment..."
	@echo "Make sure PostgreSQL is running on localhost:5432"
	@echo "Make sure the database 'languagebot' exists"
	@echo "Run 'make migrate-up' to set up the database schema"

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Lint code
lint:
	@echo "Linting code..."
	golangci-lint run

# Generate API documentation
docs:
	@echo "Generating API documentation..."
	swag init -g cmd/profile/main.go -o api/

# Install development tools
install-tools:
	@echo "Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/swaggo/swag/cmd/swag@latest
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

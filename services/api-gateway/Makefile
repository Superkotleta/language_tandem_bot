# API Gateway Makefile

.PHONY: help deps build run test clean docker-build docker-run

# Default target
help:
	@echo "Available targets:"
	@echo "  deps        - Install dependencies"
	@echo "  build       - Build the service"
	@echo "  run         - Run the service"
	@echo "  test        - Run tests"
	@echo "  clean       - Clean build artifacts"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-run  - Run with Docker"

# Install dependencies
deps:
	go mod download
	go mod tidy

# Build the service
build:
	go build -o bin/api-gateway ./cmd/api-gateway

# Run the service
run:
	go run ./cmd/api-gateway

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -rf bin/

# Build Docker image
docker-build:
	docker build -t api-gateway:latest .

# Run with Docker
docker-run:
	docker run -p 8080:8080 \
		-e PROFILE_SERVICE_URL=http://host.docker.internal:8081 \
		-e BOT_SERVICE_URL=http://host.docker.internal:8082 \
		api-gateway:latest

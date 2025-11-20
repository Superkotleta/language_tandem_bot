.PHONY: help build test clean docker-build docker-run lint format deps proto swagger

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build all services
	@echo "Building all services..."
	@for service in bot matcher profile; do \
		echo "Building $$service..."; \
		cd services/$$service && go build ./... && cd ../..; \
	done

test: ## Run tests for all services
	@echo "Running tests..."
	@for service in bot matcher profile; do \
		echo "Testing $$service..."; \
		cd services/$$service && go test ./... -v && cd ../..; \
	done

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@find . -name "*.exe" -delete
	@find . -name "*.exe~" -delete
	@find . -name "*.dll" -delete
	@find . -name "*.so" -delete
	@find . -name "*.dylib" -delete
	@find . -name "*.test" -delete
	@find . -name "*.out" -delete
	@find . -name "main" -delete

docker-build: ## Build Docker images
	@echo "Building Docker images..."
	@for service in bot matcher profile; do \
		echo "Building $$service image..."; \
		docker build -t language-exchange-$$service:latest services/$$service; \
	done

docker-run: ## Run services with Docker Compose
	@echo "Starting services with Docker Compose..."
	docker-compose up -d

lint: ## Run linter
	@echo "Running golangci-lint..."
	@golangci-lint run

format: ## Format code
	@echo "Formatting code..."
	@go fmt ./...
	@gofmt -s -w .

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

proto: ## Generate Go code from proto files
	@echo "Generating Go code from proto files..."
	@protoc --go_out=. --go-grpc_out=. api/proto/*.proto

swagger: ## Generate Swagger documentation for bot service
	@echo "Generating Swagger documentation..."
	@cd services/bot && ~/go/bin/swag init -g cmd/bot/main.go

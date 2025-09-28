# Language Exchange Bot - Microservices Makefile
# Основные команды для управления микросервисами

# Переменные
COMPOSE_FILE := docker-compose.mvp.yml
SERVICES := profile-service bot-service api-gateway

# Help - main command
.PHONY: help
help: ## Show help for commands
	@echo "Language Exchange Bot - Microservices"
	@echo "===================================="
	@echo ""
	@echo " Main Commands:"
	@echo "   make run              - Start all services"
	@echo "   make stop             - Stop all services"
	@echo "   make restart          - Restart all services"
	@echo "   make status           - Show services status"
	@echo ""
	@echo " Development:"
	@echo "   make build            - Build all services"
	@echo "   make rebuild          - Rebuild all services"
	@echo "   make rebuild-profile  - Rebuild Profile Service"
	@echo "   make rebuild-bot      - Rebuild Bot Service"
	@echo "   make rebuild-gateway  - Rebuild API Gateway"
	@echo ""
	@echo " Monitoring:"
	@echo "   make logs             - Show logs for all services"
	@echo "   make logs-profile     - Show Profile Service logs"
	@echo "   make logs-bot         - Show Bot Service logs"
	@echo "   make logs-gateway     - Show API Gateway logs"
	@echo "   make health           - Check services health"
	@echo ""
	@echo " Telegram Bot:"
	@echo "   make test-bot         - Send test message to bot"
	@echo "   make bot-info         - Get bot information"
	@echo ""
	@echo " Utilities:"
	@echo "   make clean            - Clean containers and images"
	@echo "   make fix-deps         - Fix Go dependencies"
	@echo "   make test             - Run tests"
	@echo ""
	@echo " Additional Commands:"
	@echo "   make help             - Show this help"
	@echo "   make help-verbose     - Show detailed help"
	@echo ""

# Detailed help
.PHONY: help-verbose
help-verbose: ## Show detailed help
	@echo "Language Exchange Bot - Detailed Help"
	@echo "====================================="
	@echo ""
	@echo "Project Structure:"
	@echo "  services/profile/     - Profile Service (port 8081)"
	@echo "  services/bot/         - Bot Service (port 8082)"
	@echo "  services/api-gateway/ - API Gateway (port 8080)"
	@echo ""
	@echo "Environment Variables:"
	@echo "  TELEGRAM_TOKEN       - Telegram bot token"
	@echo "  ADMIN_CHAT_IDS       - Admin chat IDs"
	@echo "  DATABASE_URL         - PostgreSQL database URL"
	@echo ""
	@echo "Service Ports:"
	@echo "  8080 - API Gateway"
	@echo "  8081 - Profile Service"
	@echo "  8082 - Bot Service"
	@echo "  5432 - PostgreSQL"
	@echo ""
	@echo "Quick Start:"
	@echo "  1. make run          - Start all services"
	@echo "  2. make health       - Check services health"
	@echo "  3. Send /start to bot in Telegram"
	@echo ""

# Main commands
.PHONY: run
run: ## Start all services
	@echo "Starting all services..."
	docker-compose -f $(COMPOSE_FILE) up -d
	@echo "Services started!"

.PHONY: stop
stop: ## Stop all services
	@echo "Stopping all services..."
	docker-compose -f $(COMPOSE_FILE) down
	@echo "Services stopped!"

.PHONY: restart
restart: stop run ## Restart all services

.PHONY: status
status: ## Show services status
	@echo "Services status:"
	docker-compose -f $(COMPOSE_FILE) ps

# Build and rebuild
.PHONY: build
build: ## Build all services
	@echo "Building all services..."
	docker-compose -f $(COMPOSE_FILE) build
	@echo "Build completed!"

.PHONY: rebuild
rebuild: ## Rebuild all services
	@echo "Rebuilding all services..."
	docker-compose -f $(COMPOSE_FILE) down
	docker-compose -f $(COMPOSE_FILE) build --no-cache
	docker-compose -f $(COMPOSE_FILE) up -d
	@echo "Rebuild completed!"

.PHONY: rebuild-profile
rebuild-profile: ## Rebuild Profile Service
	@echo "Rebuilding Profile Service..."
	docker-compose -f $(COMPOSE_FILE) build --no-cache profile-service
	docker-compose -f $(COMPOSE_FILE) up -d profile-service
	@echo "Profile Service rebuilt!"

.PHONY: rebuild-bot
rebuild-bot: ## Rebuild Bot Service
	@echo "Rebuilding Bot Service..."
	docker-compose -f $(COMPOSE_FILE) build --no-cache bot-service
	docker-compose -f $(COMPOSE_FILE) up -d bot-service
	@echo "Bot Service rebuilt!"

.PHONY: rebuild-gateway
rebuild-gateway: ## Rebuild API Gateway
	@echo "Rebuilding API Gateway..."
	docker-compose -f $(COMPOSE_FILE) build --no-cache api-gateway
	docker-compose -f $(COMPOSE_FILE) up -d api-gateway
	@echo "API Gateway rebuilt!"

# Logs
.PHONY: logs
logs: ## Show logs for all services
	docker-compose -f $(COMPOSE_FILE) logs -f

.PHONY: logs-profile
logs-profile: ## Show Profile Service logs
	docker-compose -f $(COMPOSE_FILE) logs -f profile-service

.PHONY: logs-bot
logs-bot: ## Show Bot Service logs
	docker-compose -f $(COMPOSE_FILE) logs -f bot-service

.PHONY: logs-gateway
logs-gateway: ## Show API Gateway logs
	docker-compose -f $(COMPOSE_FILE) logs -f api-gateway

# Health check
.PHONY: health
health: ## Check services health
	@echo "Checking services health..."
	@echo ""
	@echo "Profile Service (8081):"
	@powershell -Command "try { Invoke-WebRequest -Uri 'http://localhost:8081/healthz' -UseBasicParsing | Out-Null; Write-Host 'OK' } catch { Write-Host 'FAIL' }"
	@echo "Bot Service (8082):"
	@powershell -Command "try { Invoke-WebRequest -Uri 'http://localhost:8082/healthz' -UseBasicParsing | Out-Null; Write-Host 'OK' } catch { Write-Host 'FAIL' }"
	@echo "API Gateway (8080):"
	@powershell -Command "try { Invoke-WebRequest -Uri 'http://localhost:8080/healthz' -UseBasicParsing | Out-Null; Write-Host 'OK' } catch { Write-Host 'FAIL' }"

# Telegram Bot commands

.PHONY: test-bot
test-bot: ## Send test message to bot
	@echo "Sending test message..."
	@echo "Please send /start to your bot in Telegram to test it manually"
	@echo "Or use: curl -X POST https://api.telegram.org/bot<YOUR_TOKEN>/sendMessage -d 'chat_id=<CHAT_ID>&text=Hello'"

.PHONY: bot-info
bot-info: ## Get bot information
	@echo "Bot information:"
	@echo "Please check your .env file for TELEGRAM_TOKEN"
	@echo "Or use: curl https://api.telegram.org/bot<YOUR_TOKEN>/getMe"

# Utilities
.PHONY: clean
clean: ## Clean containers and images
	@echo "Cleaning containers and images..."
	docker-compose -f $(COMPOSE_FILE) down --volumes --remove-orphans
	docker system prune -f
	@echo "Cleanup completed!"

.PHONY: fix-deps
fix-deps: ## Fix Go dependencies
	@echo "Fixing Go dependencies..."
	@echo "Please run 'go mod tidy' in each service directory manually:"
	@echo "  cd services/profile && go mod tidy"
	@echo "  cd services/bot && go mod tidy"
	@echo "  cd services/api-gateway && go mod tidy"

.PHONY: test
test: ## Run tests
	@echo "Running tests..."
	@echo "Please run tests manually in each service directory:"
	@echo "  cd services/profile && go test ./..."
	@echo "  cd services/bot && go test ./..."
	@echo "  cd services/api-gateway && go test ./..."

# Show help by default
.DEFAULT_GOAL := help

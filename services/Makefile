SHELL := /bin/bash

.PHONY: up down restart logs psql-root psql-profile psql-matcher demo-rbac rebuild clean-db full-restart optimized-dev optimized-prod optimized-setup optimized-help

# Main commands
up:
	cd deploy && docker compose up -d --build

down:
	cd deploy && docker compose down -v

restart: down up

rebuild:
	cd deploy && docker compose build --no-cache

clean-db:
	cd deploy && docker compose down -v
	docker volume rm $$(docker volume ls -q --filter name=language_exchange_bot) 2>nul || echo "No volumes to remove"
	cd deploy && docker compose up -d --build

full-restart: down clean-db up

# Logs and monitoring
logs:
	cd deploy && docker compose logs -f --tail=200

logs-bot:
	cd deploy && docker compose logs -f bot --tail=100

logs-db:
	cd deploy && docker compose logs -f postgres --tail=100

# Database access
psql-root:
	cd deploy && docker compose exec -e PGPASSWORD=$$POSTGRES_PASSWORD postgres psql -U $$POSTGRES_USER $$POSTGRES_DB

psql-profile:
	cd deploy && docker compose exec -e PGPASSWORD=$$PROFILE_DB_PASS postgres psql "postgresql://profile_rw:$$PROFILE_DB_PASS@postgres:5432/$$POSTGRES_DB?sslmode=disable"

psql-matcher:
	cd deploy && docker compose exec -e PGPASSWORD=$$MATCHING_DB_PASS postgres psql "postgresql://matching_rw:$$MATCHING_DB_PASS@postgres:5432/$$POSTGRES_DB?sslmode=disable"

# Demonstrations and tests
demo-rbac:
	@echo "== matching_ro can read profile.users, but cannot write =="
	cd deploy && docker compose exec -e PGPASSWORD=$$MATCHING_RO_DB_PASS postgres psql "postgresql://matching_ro:$$MATCHING_RO_DB_PASS@postgres:5432/$$POSTGRES_DB?sslmode=disable" -c "SET search_path=profile,public; SELECT count(*) FROM users;"
	cd deploy && docker compose exec -e PGPASSWORD=$$MATCHING_RO_DB_PASS postgres psql "postgresql://matching_ro:$$MATCHING_RO_DB_PASS@postgres:5432/$$POSTGRES_DB?sslmode=disable" -c "SET search_path=profile,public; INSERT INTO users(telegram_id) VALUES (9999); " || true

# Convenient development commands
rebuild-and-up: rebuild up

clean-logs:
	cd deploy && docker compose logs --no-log-prefix | head -n 10000 > ../logs/docker-compose.log
	@echo "Logs saved to logs/docker-compose.log"

check-health:
	cd deploy && docker compose ps
	@echo "=== Container Status ==="
	-cd deploy && docker compose exec postgres pg_isready -U postgres -d postgres || echo "PostgreSQL not ready"
	@echo "=== Waiting for PostgreSQL to be ready ==="
	@echo "If containers are running but PostgreSQL is not ready, wait 30-60 seconds and try again."

help:
	@echo "Available commands:"
	@echo ""
	@echo "=== Main commands (v1.0) ==="
	@echo "  up              - Start containers"
	@echo "  down            - Stop containers and remove volumes"
	@echo "  restart         - Restart containers"
	@echo "  rebuild         - Rebuild images without cache"
	@echo "  clean-db        - Clean database volumes and restart"
	@echo "  full-restart    - Full restart with clean database"
	@echo "  rebuild-and-up  - Rebuild and start"
	@echo "  logs            - Show all container logs"
	@echo "  logs-bot        - Show only bot logs"
	@echo "  logs-db         - Show only database logs"
	@echo "  clean-logs      - Save logs to file"
	@echo "  check-health    - Check container status"
	@echo "  psql-root       - Connect to main database"
	@echo "  psql-profile    - Connect to profile schema"
	@echo "  psql-matcher    - Connect to matching schema"
	@echo "  demo-rbac       - Show RBAC demonstration"
	@echo ""
	@echo "=== Optimized version (v2.0) ==="
	@echo "  optimized-dev           - Start for development (with ngrok)"
	@echo "  optimized-prod          - Start for production"
	@echo "  optimized-setup         - Full development setup"
	@echo "  optimized-ngrok         - Setup ngrok"
	@echo "  optimized-webhook       - Setup webhook"
	@echo "  optimized-stop-ngrok    - Stop ngrok"
	@echo "  optimized-monitor       - Monitoring"
	@echo "  optimized-logs          - Logs"
	@echo "  optimized-health        - Health check"
	@echo "  optimized-down          - Stop"
	@echo "  optimized-restart       - Restart"
	@echo "  optimized-backup        - Create backup"
	@echo "  optimized-production-guide - Production instructions"
	@echo "  optimized-help          - Optimized version help"
	@echo ""
	@echo "=== Windows commands ==="
	@echo "  win-logs-bot    - Show bot logs (PowerShell)"
	@echo "  win-logs-db     - Show database logs (PowerShell)"
	@echo "  win-check-emojis - Check emoji loading"
	@echo "  win-clean-and-restart - Clean and restart (PowerShell)"
	@echo "  win-clean-all   - Full clean (PowerShell)"
	@echo "  win-diagnose    - Network diagnostics"
	@echo ""
	@echo "  help            - Show this help"

# Legacy aliases for compatibility
migrate-db: full-restart
db-recreate: clean-db

# Windows PowerShell compatibility commands
win-logs-bot:
	powershell -Command "cd deploy; docker compose logs -f bot --tail=100"

win-logs-db:
	powershell -Command "cd deploy; docker compose logs -f postgres --tail=50"

win-check-emojis:
	@echo "=== Checking if emojis are loaded in database ==="
	powershell -Command "cd deploy; docker compose exec postgres psql -U postgres -d languagebot -c \"SELECT t.name FROM interest_translations t JOIN interests i ON t.interest_id=i.id WHERE t.language_code='ru' AND i.key_name='movies_tv';\""

win-clean-and-restart:
	powershell -Command "cd deploy; docker compose down -v"
	powershell -Command "docker volume prune -f"
	powershell -Command "docker network prune -f"
	powershell -Command "cd deploy; docker compose up -d --build"

win-clean-all:
	powershell -Command "cd deploy; docker compose down -v"
	powershell -Command "docker volume prune -f --filter label=com.docker.compose.project=language_exchange_bot"
	powershell -Command "docker network prune -f"
	powershell -Command "docker system prune -f"
	powershell -Command "cd deploy; docker compose up -d --build"

win-diagnose:
	@echo "=== Network and IP address check ==="
	docker network ls
	docker network inspect deploy_app-network || echo "Network not found"
	docker container ls -a --filter name=pg
	docker container inspect pg --format "{{ range .NetworkSettings.Networks }}{{ .IPAddress }}{{ end }}" || echo "IP info not available"
	@echo "=== If IP is not shown, try: ==="
	@echo "docker logs pg | findstr -c:database"
	@echo "docker logs deploy-bot-1 | tail -n 10"

# ===========================================
# Optimized version (v2.0)
# ===========================================

# Development with ngrok
optimized-dev: ## Start optimized version for development
	@echo "Starting optimized version for development..."
	cd deploy && make -f Makefile.optimized dev-setup

# Production deployment
optimized-prod: ## Start optimized version for production
	@echo "Starting optimized version for production..."
	cd deploy && make -f Makefile.optimized prod

# Full development setup
optimized-setup: ## Full setup of optimized version
	@echo "Full setup of optimized version..."
	cd deploy && make -f Makefile.optimized dev-setup

# Setup ngrok
optimized-ngrok: ## Setup ngrok for development
	@echo "Setting up ngrok..."
	cd deploy && make -f Makefile.optimized ngrok-setup

# Setup webhook
optimized-webhook: ## Setup webhook in Telegram
	@echo "Setting up webhook..."
	cd deploy && make -f Makefile.optimized webhook-setup

# Stop ngrok
optimized-stop-ngrok: ## Stop ngrok
	@echo "Stopping ngrok..."
	cd deploy && make -f Makefile.optimized ngrok-stop

# Monitoring optimized version
optimized-monitor: ## Monitor optimized version
	@echo "Monitoring optimized version..."
	cd deploy && make -f Makefile.optimized monitor

# Logs optimized version
optimized-logs: ## Logs optimized version
	@echo "Logs optimized version..."
	cd deploy && make -f Makefile.optimized logs

# Health check optimized version
optimized-health: ## Health check optimized version
	@echo "Health check optimized version..."
	cd deploy && make -f Makefile.optimized health

# Stop optimized version
optimized-down: ## Stop optimized version
	@echo "Stopping optimized version..."
	cd deploy && make -f Makefile.optimized down

# Restart optimized version
optimized-restart: ## Restart optimized version
	@echo "Restarting optimized version..."
	cd deploy && make -f Makefile.optimized restart

# Backup optimized version
optimized-backup: ## Create backup of optimized version
	@echo "Creating backup..."
	cd deploy && make -f Makefile.optimized db-backup

# Production instructions
optimized-production-guide: ## Show production instructions
	@echo "Production deployment instructions..."
	cd deploy && make -f Makefile.optimized production-setup

# Help for optimized version
optimized-help: ## Help for optimized version
	@echo "Help for optimized version..."
	cd deploy && make -f Makefile.optimized help

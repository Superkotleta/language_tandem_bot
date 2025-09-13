SHELL := /bin/sh

.PHONY: up down restart logs psql-root psql-profile psql-matcher demo-rbac rebuild clean-db full-restart

# Основные команды
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

# Логи и мониторинг
logs:
	cd deploy && docker compose logs -f --tail=200

logs-bot:
	cd deploy && docker compose logs -f bot --tail=100

logs-db:
	cd deploy && docker compose logs -f postgres --tail=100

# Доступ к базам данных
psql-root:
	cd deploy && docker compose exec -e PGPASSWORD=$$POSTGRES_PASSWORD postgres psql -U $$POSTGRES_USER $$POSTGRES_DB

psql-profile:
	cd deploy && docker compose exec -e PGPASSWORD=$$PROFILE_DB_PASS postgres psql "postgresql://profile_rw:$$PROFILE_DB_PASS@postgres:5432/$$POSTGRES_DB?sslmode=disable"

psql-matcher:
	cd deploy && docker compose exec -e PGPASSWORD=$$MATCHING_DB_PASS postgres psql "postgresql://matching_rw:$$MATCHING_DB_PASS@postgres:5432/$$POSTGRES_DB?sslmode=disable"

# Демонстрации и тесты
demo-rbac:
	@echo "== matching_ro can read profile.users, but cannot write =="
	cd deploy && docker compose exec -e PGPASSWORD=$$MATCHING_RO_DB_PASS postgres psql "postgresql://matching_ro:$$MATCHING_RO_DB_PASS@postgres:5432/$$POSTGRES_DB?sslmode=disable" -c "SET search_path=profile,public; SELECT count(*) FROM users;"
	cd deploy && docker compose exec -e PGPASSWORD=$$MATCHING_RO_DB_PASS postgres psql "postgresql://matching_ro:$$MATCHING_RO_DB_PASS@postgres:5432/$$POSTGRES_DB?sslmode=disable" -c "SET search_path=profile,public; INSERT INTO users(telegram_id) VALUES (9999); " || true

# Удобные команды для разработки
rebuild-and-up: rebuild up

clean-logs:
	cd deploy && docker compose logs --no-log-prefix | head -n 10000 > ../logs/docker-compose.log
	@echo "Logs saved to logs/docker-compose.log"

check-health:
	cd deploy && docker compose ps
	@echo "=== Container Status ==="
	cd deploy && docker compose exec postgres pg_isready -U $$(cat ../.env | grep POSTGRES_USER | cut -d '=' -f2) -d $$(cat ../.env | grep POSTGRES_DB | cut -d '=' -f2) || echo "PostgreSQL not ready"

help:
	@echo "Available commands:"
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
	@echo "  help            - Show this help"

# Legacy aliases for compatibility
migrate-db: full-restart
db-recreate: clean-db

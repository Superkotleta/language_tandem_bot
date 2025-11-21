.PHONY: up down logs migrate ps restart clean lint lint-gentle lint-fast lint-enhanced lint-strict fmt vet fix install help

# Справка
help:
	@echo "=== Docker команды ==="
	@echo "  make up           - Запуск всех сервисов"
	@echo "  make down         - Остановка"
	@echo "  make logs         - Просмотр логов"
	@echo "  make migrate      - Применение миграций"
	@echo "  make ps           - Статус контейнеров"
	@echo ""
	@echo "=== Линтер команды ==="
	@echo "  make install      - Установить golangci-lint"
	@echo "  make lint         - Запуск линтера (мягкая конфигурация)"
	@echo "  make lint-fast    - Быстрая проверка"
	@echo "  make lint-enhanced - Улучшенная проверка"
	@echo "  make lint-strict  - Строгая проверка"
	@echo "  make fmt          - Форматирование кода"
	@echo "  make vet          - Go vet"
	@echo "  make fix          - Автоисправление"
	@echo ""
	@echo "Подробнее: cd linter && make help"

# Запуск всех сервисов
up:
	docker-compose up -d --build

# Остановка всех сервисов
down:
	docker-compose down

# Просмотр логов всех сервисов
logs:
	docker-compose logs -f

# Просмотр логов только бота
logs-bot:
	docker-compose logs -f bot

# Просмотр логов только БД
logs-db:
	docker-compose logs -f db

# Применение миграций
migrate:
	@echo "Применение миграций..."
	docker-compose exec -T bot /bin/sh -c "psql \$$DATABASE_URL < /app/migrations/000001_create_users_table.up.sql"
	docker-compose exec -T bot /bin/sh -c "psql \$$DATABASE_URL < /app/migrations/000002_create_reference_tables.up.sql"
	docker-compose exec -T bot /bin/sh -c "psql \$$DATABASE_URL < /app/migrations/000003_seed_data.up.sql"
	@echo "Миграции применены успешно!"

# Статус контейнеров
ps:
	docker-compose ps

# Перезапуск всех сервисов
restart:
	docker-compose restart

# Перезапуск только бота
restart-bot:
	docker-compose restart bot

# Остановка и удаление контейнеров, сетей, volumes
clean:
	docker-compose down -v

# Подключение к БД через psql
db-shell:
	docker-compose exec db psql -U postgres -d language_exchange

# Просмотр переменных окружения
env:
	docker-compose config

# === ЛИНТЕР КОМАНДЫ ===

# Установка golangci-lint
install:
	@cd linter && $(MAKE) install

# Мягкая конфигурация (рекомендуется)
lint:
	@cd linter && $(MAKE) lint-gentle

# Алиас для lint
lint-gentle:
	@cd linter && $(MAKE) lint-gentle

# Быстрая конфигурация
lint-fast:
	@cd linter && $(MAKE) lint-fast

# Улучшенная конфигурация
lint-enhanced:
	@cd linter && $(MAKE) lint-enhanced

# Строгая конфигурация
lint-strict:
	@cd linter && $(MAKE) lint-strict

# Форматирование кода
fmt:
	@cd linter && $(MAKE) fmt

# Go vet
vet:
	@cd linter && $(MAKE) vet

# Автоисправление
fix:
	@cd linter && $(MAKE) fix

# Проверка перед коммитом
pre-commit:
	@cd linter && $(MAKE) pre-commit


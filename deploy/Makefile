.PHONY: up down logs migrate ps restart clean

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


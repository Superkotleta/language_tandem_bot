# Инструкция по запуску бота

## Подготовка базы данных

1. Создай базу данных PostgreSQL
2. Примени миграции (последовательно):

```bash
# 1. Создание таблицы пользователей (UUID, основные поля)
psql $DATABASE_URL < migrations/000001_create_users_table.up.sql

# 2. Создание справочников (языки, категории, интересы)
psql $DATABASE_URL < migrations/000002_create_reference_tables.up.sql

# 3. Наполнение базы начальными данными (список языков и интересов)
psql $DATABASE_URL < migrations/000003_seed_data.up.sql
```

## Настройка переменных окружения

Создай файл `.env` или экспортируй переменные:

```bash
export DATABASE_URL="postgres://user:password@localhost:5432/dbname?sslmode=disable"
export TELEGRAM_TOKEN="your_bot_token_from_botfather"
export LOCALES_PATH="./locales"  # Опционально, по умолчанию ./locales
```

## Запуск

```bash
# Если переменные в .env
source .env

# Запуск через go run
go run cmd/bot/main.go

# ИЛИ сборка и запуск бинарника
go build -o bin/bot cmd/bot/main.go
./bin/bot
```

## Что работает сейчас

✅ Бот запускается и подключается к БД
✅ При `/start` регистрирует пользователя (создает UUID профиль)
✅ Полная схема БД с поддержкой JSONB-переводов
✅ Заполненные справочники языков и интересов
✅ Локализация (ru, en, es, zh)
✅ Graceful shutdown (Ctrl+C)

## Что НЕ работает (ещё не реализовано)

❌ Wizard заполнения профиля (пошаговый выбор языка/интересов)
❌ Просмотр профиля
❌ Редактирование профиля
❌ Поиск партнёров
❌ Система матчинга

## Структура проекта

```
cmd/bot/main.go              # Точка входа
internal/
  ├── config/                # Загрузка конфигурации
  ├── domain/                # Бизнес-сущности (User, Language, Interest, Category)
  ├── repository/            # Работа с БД (pgx/v5)
  │     ├── user_repository.go       # Работа с пользователями
  │     └── reference_repository.go  # Справочники (языки, интересы)
  ├── service/               # Бизнес-логика
  ├── delivery/telegram/     # Telegram bot хэндлер
  ├── pkg/i18n/              # Система локализации
  └── ui/                    # UI компоненты (keyboards, messages)
```

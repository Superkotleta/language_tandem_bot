# Инструкция по запуску бота

## Вариант 1: Запуск через Docker (Рекомендуется)

### Быстрый старт

1. Перейди в папку `deploy`:
```bash
cd deploy
```

2. Скопируй и настрой переменные окружения:
```bash
cp env.example .env
```

3. Отредактируй `.env`, добавь свой токен бота:
```bash
TELEGRAM_TOKEN=ваш_токен_от_BotFather
```

4. Запусти все сервисы (БД + pgAdmin + бот):
```bash
make up
```

5. Примени миграции:
```bash
make migrate
```

6. Проверь логи бота:
```bash
make logs-bot
```

### Управление сервисами

```bash
make down          # Остановить все контейнеры
make restart-bot   # Перезапустить только бота
make logs          # Просмотр логов всех сервисов
make ps            # Статус контейнеров
make db-shell      # Подключиться к PostgreSQL
make clean         # Удалить контейнеры и volumes (ОСТОРОЖНО!)
```

### Доступ к pgAdmin

- URL: http://localhost:5050
- Email/Password: указаны в `.env` (по умолчанию: `admin@localhost.local` / `admin`)

Для подключения к БД в pgAdmin:
- Host: `db` (имя контейнера)
- Port: `5432`
- Database: `language_exchange`
- Username/Password: из `.env`

---

## Вариант 2: Локальный запуск (для разработки)

### Подготовка базы данных

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

### Настройка переменных окружения

Создай файл `.env` или экспортируй переменные:

```bash
export DATABASE_URL="postgres://user:password@localhost:5432/dbname?sslmode=disable"
export TELEGRAM_TOKEN="your_bot_token_from_botfather"
export LOCALES_PATH="./locales"  # Опционально, по умолчанию ./locales
```

### Запуск

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

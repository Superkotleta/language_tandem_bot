# Deploy Configuration

Эта папка содержит все необходимое для развертывания бота через Docker.

## Быстрый старт

1. Скопируй и настрой переменные окружения:
```bash
cp env.example .env
# Отредактируй .env, добавь TELEGRAM_TOKEN
```

2. Запусти сервисы:
```bash
make up
```

3. Примени миграции:
```bash
make migrate
```

4. Проверь логи:
```bash
make logs-bot
```

## Файлы

- `Dockerfile` - Образ бота (Go 1.25 + Alpine)
- `docker-compose.yml` - Конфигурация сервисов (PostgreSQL 17 + pgAdmin + Bot)
- `Makefile` - Команды для управления
- `env.example` - Шаблон переменных окружения
- `linter/` - Система линтинга с 4 конфигурациями

## Сервисы

### PostgreSQL 17
- Порт: `5432`
- Данные хранятся в volume `postgres_data`

### pgAdmin
- URL: http://localhost:5050
- Для подключения к БД используй host: `db`

### Bot
- Автоматически подключается к БД
- Логи ротируются (max 10MB, 3 файла)

## Команды Makefile

### Docker команды

```bash
make up            # Запуск всех сервисов
make down          # Остановка
make logs          # Просмотр логов всех сервисов
make logs-bot      # Логи только бота
make migrate       # Применение миграций
make restart-bot   # Перезапуск бота
make db-shell      # Подключение к PostgreSQL
make ps            # Статус контейнеров
make clean         # Удаление контейнеров и данных (ОСТОРОЖНО!)
```

### Линтер команды

```bash
make lint          # Мягкая проверка (рекомендуется)
make lint-fast     # Быстрая проверка
make lint-enhanced # Улучшенная проверка
make lint-strict   # Строгая проверка
make fmt           # Форматирование кода
make vet           # Go vet
make fix           # Автоисправление
make pre-commit    # Проверка перед коммитом
```

Подробнее о линтере: [linter/README.md](linter/README.md)

## Логи

Логи всех контейнеров настроены на автоматическую ротацию:
- Максимальный размер файла: 10 MB
- Количество хранимых файлов: 3

## Разработка

Для локальной разработки с автоматической перезагрузкой можно примонтировать исходники:

```yaml
# Добавь в docker-compose.yml в сервис bot:
volumes:
  - ../cmd:/app/cmd:ro
  - ../internal:/app/internal:ro
```

Затем используй `air` или аналогичный инструмент для hot-reload.


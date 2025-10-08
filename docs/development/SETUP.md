# 🛠️ Development Setup Guide

Подробное руководство по настройке окружения разработки для Language Exchange Bot.

## 📋 Предварительные требования

### Системные требования
- **OS**: Linux/macOS/Windows (рекомендуется Linux)
- **RAM**: Минимум 4GB, рекомендуется 8GB+
- **Disk**: 10GB свободного места
- **CPU**: 2+ ядра

### Обязательное ПО
- **Go 1.25+** - [Установка](https://golang.org/doc/install)
- **Docker 20.10+** - [Установка](https://docs.docker.com/get-docker/)
- **Docker Compose 2.0+** - [Установка](https://docs.docker.com/compose/install/)
- **Git 2.30+** - [Установка](https://git-scm.com/downloads)

### Опциональное ПО
- **PostgreSQL 14+** - для локальной разработки
- **Redis 6+** - для локального кэширования
- **VS Code** - рекомендуемый редактор
- **Postman** - для тестирования API
- **ngrok** - для webhook тестирования

## 🚀 Пошаговая настройка

### 1. Клонирование репозитория

```bash
# Клонируем репозиторий
git clone https://github.com/your-org/language-tandem-bot.git
cd language-tandem-bot

# Проверяем версию Go
go version
# Должно быть: go version go1.25.x linux/amd64
```

### 2. Настройка Go workspace

```bash
# Инициализируем Go workspace
go work init

# Добавляем модули в workspace
go work use ./services/bot
go work use ./services/matcher
go work use ./services/profile

# Проверяем workspace
go work list
```

### 3. Настройка переменных окружения

```bash
# Копируем пример конфигурации
cp .env.example .env

# Редактируем конфигурацию
nano .env
```

#### Основные переменные (.env):
```bash
# Telegram Bot
TELEGRAM_TOKEN=your_bot_token_here
TELEGRAM_MODE=polling
WEBHOOK_URL=https://your-domain.com

# Database
DATABASE_URL=postgres://user:password@localhost:5432/language_exchange?sslmode=disable
DATABASE_MAX_OPEN_CONNS=25
DATABASE_MAX_IDLE_CONNS=10

# Redis
REDIS_URL=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0

# Server
PORT=8080
DEBUG=true

# Admin
ADMIN_CHAT_IDS=123456789,987654321
ADMIN_USERNAMES=admin1,admin2
```

### 4. Настройка базы данных

#### Вариант A: Docker Compose (рекомендуется)
```bash
# Запускаем только базу данных
docker-compose up -d postgres redis

# Проверяем статус
docker-compose ps
```

#### Вариант B: Локальная установка
```bash
# Ubuntu/Debian
sudo apt-get install postgresql postgresql-contrib redis-server

# macOS
brew install postgresql redis

# Windows
# Скачайте с официальных сайтов
```

### 5. Инициализация базы данных

```bash
# Запускаем миграции
cd services/deploy
./db-init/bootstrap.sh

# Или вручную
psql -h localhost -U postgres -d language_exchange -f db-init/01-init-schemas.sql
```

### 6. Установка зависимостей

```bash
# Переходим в директорию бота
cd services/bot

# Скачиваем зависимости
go mod download

# Проверяем зависимости
go mod verify
```

### 7. Запуск в development режиме

```bash
# Запускаем бота
go run cmd/bot/main.go

# Или с hot reload (если установлен air)
air
```

### 8. Проверка работоспособности

```bash
# Health check
curl http://localhost:8080/healthz

# API документация
open http://localhost:8080/swagger/

# Статус базы данных
curl http://localhost:8080/api/v1/stats
```

## 🔧 Настройка IDE

### VS Code

#### Установка расширений:
```bash
code --install-extension golang.go
code --install-extension ms-vscode.vscode-json
code --install-extension bradlc.vscode-tailwindcss
code --install-extension ms-vscode.vscode-docker
```

#### Настройка Go:
```json
// .vscode/settings.json
{
    "go.toolsManagement.checkForUpdates": "local",
    "go.useLanguageServer": true,
    "go.formatTool": "goimports",
    "go.lintTool": "golangci-lint",
    "go.testFlags": ["-v"],
    "go.buildTags": "debug"
}
```

#### Конфигурация отладки:
```json
// .vscode/launch.json
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Debug Bot",
            "type": "go",
            "request": "launch",
            "mode": "auto",
            "program": "${workspaceFolder}/services/bot/cmd/bot",
            "env": {
                "DEBUG": "true",
                "TELEGRAM_MODE": "polling"
            }
        }
    ]
}
```

### GoLand/IntelliJ

1. Откройте проект как Go module
2. Настройте Go SDK (1.25+)
3. Включите Go modules
4. Настройте run configuration для `cmd/bot/main.go`

## 🐳 Docker Development

### Docker Compose для разработки

```yaml
# docker-compose.dev.yml
version: '3.8'
services:
  postgres:
    image: postgres:14
    environment:
      POSTGRES_DB: language_exchange
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

  redis:
    image: redis:6-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data

  bot:
    build:
      context: ./services/bot
      dockerfile: Dockerfile.dev
    volumes:
      - ./services/bot:/app
    environment:
      - DEBUG=true
      - TELEGRAM_MODE=polling
    depends_on:
      - postgres
      - redis

volumes:
  postgres_data:
  redis_data:
```

### Запуск с Docker:
```bash
# Development режим
docker-compose -f docker-compose.dev.yml up

# С hot reload
docker-compose -f docker-compose.dev.yml up --build
```

## 🧪 Настройка тестирования

### Unit тесты
```bash
# Запуск всех тестов
go test ./...

# С покрытием
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Конкретный пакет
go test ./internal/cache/... -v
```

### Integration тесты
```bash
# Запуск integration тестов
go test ./tests/integration/... -v

# С Docker
docker-compose -f docker-compose.test.yml up --abort-on-container-exit
```

### Настройка тестовой базы данных
```bash
# Создаем тестовую базу
createdb language_exchange_test

# Запускаем миграции
psql -d language_exchange_test -f services/deploy/db-init/01-init-schemas.sql
```

## 🔍 Отладка и профилирование

### Логирование
```bash
# Включить debug логи
export DEBUG=true

# Уровень логирования
export LOG_LEVEL=debug
```

### Профилирование
```bash
# CPU профилирование
go run cmd/bot/main.go &
go tool pprof http://localhost:8080/debug/pprof/profile

# Memory профилирование
go tool pprof http://localhost:8080/debug/pprof/heap

# Goroutine профилирование
go tool pprof http://localhost:8080/debug/pprof/goroutine
```

### Отладка с Delve
```bash
# Установка Delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Запуск с отладчиком
dlv debug ./cmd/bot/main.go
```

## 🌐 Настройка Telegram Bot

### 1. Создание бота
1. Напишите [@BotFather](https://t.me/botfather) в Telegram
2. Используйте команду `/newbot`
3. Следуйте инструкциям
4. Сохраните полученный токен

### 2. Настройка webhook (опционально)
```bash
# Установите ngrok
ngrok http 8080

# Скопируйте HTTPS URL
# Установите в .env:
WEBHOOK_URL=https://your-ngrok-url.ngrok.io
TELEGRAM_MODE=webhook
```

### 3. Тестирование бота
```bash
# Запустите бота
go run cmd/bot/main.go

# Найдите бота в Telegram
# Отправьте /start
```

## 📊 Мониторинг и метрики

### Prometheus метрики
```bash
# Просмотр метрик
curl http://localhost:8080/metrics

# Grafana dashboard
open http://localhost:3000
```

### Health checks
```bash
# Readiness probe
curl http://localhost:8080/readyz

# Liveness probe
curl http://localhost:8080/healthz

# Детальная информация
curl http://localhost:8080/api/v1/stats
```

## 🚨 Troubleshooting

### Частые проблемы

#### 1. Ошибка подключения к базе данных
```bash
# Проверьте статус PostgreSQL
docker-compose ps postgres

# Проверьте логи
docker-compose logs postgres

# Проверьте подключение
psql -h localhost -U postgres -d language_exchange
```

#### 2. Ошибка подключения к Redis
```bash
# Проверьте статус Redis
docker-compose ps redis

# Проверьте подключение
redis-cli ping
```

#### 3. Проблемы с Telegram API
```bash
# Проверьте токен
curl "https://api.telegram.org/bot<YOUR_TOKEN>/getMe"

# Проверьте webhook
curl "https://api.telegram.org/bot<YOUR_TOKEN>/getWebhookInfo"
```

#### 4. Проблемы с портами
```bash
# Проверьте занятые порты
netstat -tulpn | grep :8080
netstat -tulpn | grep :5432
netstat -tulpn | grep :6379

# Освободите порты
sudo fuser -k 8080/tcp
```

### Логи и отладка
```bash
# Просмотр логов
docker-compose logs -f bot

# Логи с фильтрацией
docker-compose logs bot | grep ERROR

# Отладка Go
go run -race cmd/bot/main.go
```

## 📚 Дополнительные ресурсы

- [Go Documentation](https://golang.org/doc/)
- [Docker Documentation](https://docs.docker.com/)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [Redis Documentation](https://redis.io/documentation)
- [Telegram Bot API](https://core.telegram.org/bots/api)

---

**Готово! 🎉** Теперь вы можете начать разработку. Если возникли проблемы, обратитесь к команде разработки или создайте issue в GitHub.

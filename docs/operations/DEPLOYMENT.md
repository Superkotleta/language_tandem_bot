# Deployment Guide

Руководство по развертыванию Language Exchange Bot в production среде.

## Предварительные требования

### Системные требования

- **OS**: Linux (Ubuntu 20.04+ рекомендуется)
- **CPU**: 2+ cores
- **RAM**: 4GB+ (8GB+ рекомендуется)
- **Storage**: 20GB+ SSD
- **Network**: Стабильное интернет-соединение

### Зависимости

- **Go**: 1.21+
- **PostgreSQL**: 13+
- **Redis**: 6.0+
- **Docker**: 20.10+ (опционально)
- **Docker Compose**: 2.0+ (опционально)

## Конфигурация окружения

### 1. Переменные окружения

Создайте файл `.env` в корне проекта:

```bash
# Database
DATABASE_URL=postgres://username:password@localhost:5432/language_exchange_bot
DATABASE_MAX_OPEN_CONNS=25
DATABASE_MAX_IDLE_CONNS=10

# Redis
REDIS_URL=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0

# Telegram Bot
TELEGRAM_BOT_TOKEN=your_bot_token_here
TELEGRAM_WEBHOOK_URL=https://yourdomain.com/webhook/telegram
TELEGRAM_WEBHOOK_SECRET=your_webhook_secret

# Server
SERVER_PORT=8080
SERVER_HOST=0.0.0.0

# Logging
LOG_LEVEL=info
LOG_FORMAT=json

# Security
API_SECRET_KEY=your_secret_key_here
JWT_SECRET=your_jwt_secret_here

# Monitoring
PROMETHEUS_ENABLED=true
PROMETHEUS_PORT=9090
```

### 2. База данных

#### Создание базы данных

```sql
-- Подключитесь к PostgreSQL как суперпользователь
CREATE DATABASE language_exchange_bot;
CREATE USER bot_user WITH PASSWORD 'secure_password';
GRANT ALL PRIVILEGES ON DATABASE language_exchange_bot TO bot_user;
```

#### Применение миграций

```bash
# Установите migrate CLI
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Примените миграции
migrate -path services/deploy/migrations -database "postgres://bot_user:secure_password@localhost:5432/language_exchange_bot?sslmode=disable" up
```

### 3. Redis

#### Установка Redis

```bash
# Ubuntu/Debian
sudo apt update
sudo apt install redis-server

# CentOS/RHEL
sudo yum install redis
sudo systemctl enable redis
sudo systemctl start redis
```

#### Конфигурация Redis

```bash
# /etc/redis/redis.conf
maxmemory 2gb
maxmemory-policy allkeys-lru
save 900 1
save 300 10
save 60 10000
```

## Развертывание

### Вариант 1: Docker Compose (Рекомендуется)

#### 1. Создайте docker-compose.yml

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: language_exchange_bot
      POSTGRES_USER: bot_user
      POSTGRES_PASSWORD: secure_password
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./services/deploy/migrations:/docker-entrypoint-initdb.d
    ports:
      - "5432:5432"
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U bot_user -d language_exchange_bot"]
      interval: 10s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7-alpine
    command: redis-server --appendonly yes --maxmemory 2gb --maxmemory-policy allkeys-lru
    volumes:
      - redis_data:/data
    ports:
      - "6379:6379"
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5

  bot:
    build: .
    environment:
      - DATABASE_URL=postgres://bot_user:secure_password@postgres:5432/language_exchange_bot?sslmode=disable
      - REDIS_URL=redis:6379
      - TELEGRAM_BOT_TOKEN=${TELEGRAM_BOT_TOKEN}
      - TELEGRAM_WEBHOOK_URL=${TELEGRAM_WEBHOOK_URL}
      - SERVER_PORT=8080
    ports:
      - "8080:8080"
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3

volumes:
  postgres_data:
  redis_data:
```

#### 2. Создайте Dockerfile

```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/bot

FROM alpine:latest
RUN apk --no-cache add ca-certificates curl
WORKDIR /root/

COPY --from=builder /app/main .
COPY --from=builder /app/services/deploy/migrations ./migrations

EXPOSE 8080
CMD ["./main"]
```

#### 3. Запуск

```bash
# Создайте .env файл с переменными
cp .env.example .env
# Отредактируйте .env файл

# Запустите сервисы
docker-compose up -d

# Проверьте статус
docker-compose ps
docker-compose logs -f bot
```

### Вариант 2: Ручное развертывание

#### 1. Сборка приложения

```bash
# Клонируйте репозиторий
git clone https://github.com/your-org/language_tandem_bot.git
cd language_tandem_bot

# Установите зависимости
go mod download

# Соберите приложение
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bot ./cmd/bot
```

#### 2. Настройка systemd

Создайте файл `/etc/systemd/system/language-exchange-bot.service`:

```ini
[Unit]
Description=Language Exchange Bot
After=network.target postgresql.service redis.service

[Service]
Type=simple
User=bot
Group=bot
WorkingDirectory=/opt/language-exchange-bot
ExecStart=/opt/language-exchange-bot/bot
Restart=always
RestartSec=5
EnvironmentFile=/opt/language-exchange-bot/.env

[Install]
WantedBy=multi-user.target
```

#### 3. Запуск сервиса

```bash
# Создайте пользователя
sudo useradd -r -s /bin/false bot

# Создайте директорию
sudo mkdir -p /opt/language-exchange-bot
sudo cp bot /opt/language-exchange-bot/
sudo cp .env /opt/language-exchange-bot/
sudo chown -R bot:bot /opt/language-exchange-bot

# Запустите сервис
sudo systemctl daemon-reload
sudo systemctl enable language-exchange-bot
sudo systemctl start language-exchange-bot

# Проверьте статус
sudo systemctl status language-exchange-bot
```

## Настройка веб-хуков

### 1. Настройка Telegram Webhook

```bash
# Установите webhook
curl -X POST "https://api.telegram.org/bot${TELEGRAM_BOT_TOKEN}/setWebhook" \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://yourdomain.com/webhook/telegram",
    "secret_token": "your_webhook_secret"
  }'

# Проверьте статус webhook
curl "https://api.telegram.org/bot${TELEGRAM_BOT_TOKEN}/getWebhookInfo"
```

### 2. Настройка SSL/TLS

#### Используя Let's Encrypt

```bash
# Установите certbot
sudo apt install certbot

# Получите сертификат
sudo certbot certonly --standalone -d yourdomain.com

# Настройте автоматическое обновление
sudo crontab -e
# Добавьте: 0 12 * * * /usr/bin/certbot renew --quiet
```

#### Используя nginx

```nginx
server {
    listen 80;
    server_name yourdomain.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl;
    server_name yourdomain.com;

    ssl_certificate /etc/letsencrypt/live/yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/yourdomain.com/privkey.pem;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

## Health Checks

### 1. Endpoint проверки

```bash
# Health check
curl http://localhost:8080/health

# Metrics
curl http://localhost:8080/metrics

# API status
curl -H "Authorization: Bearer YOUR_TOKEN" http://localhost:8080/api/v1/stats
```

### 2. Мониторинг скрипт

Создайте `/opt/language-exchange-bot/health-check.sh`:

```bash
#!/bin/bash

HEALTH_URL="http://localhost:8080/health"
MAX_RETRIES=3
RETRY_INTERVAL=5

for i in $(seq 1 $MAX_RETRIES); do
    if curl -f -s $HEALTH_URL > /dev/null; then
        echo "Health check passed"
        exit 0
    fi
    echo "Health check failed, attempt $i/$MAX_RETRIES"
    sleep $RETRY_INTERVAL
done

echo "Health check failed after $MAX_RETRIES attempts"
exit 1
```

## Rollback процедуры

### 1. Откат к предыдущей версии

```bash
# Остановите сервис
sudo systemctl stop language-exchange-bot

# Восстановите предыдущую версию
sudo cp /opt/language-exchange-bot/bot.backup /opt/language-exchange-bot/bot

# Запустите сервис
sudo systemctl start language-exchange-bot
```

### 2. Откат базы данных

```bash
# Создайте бэкап перед развертыванием
pg_dump language_exchange_bot > backup_$(date +%Y%m%d_%H%M%S).sql

# Откат миграций
migrate -path services/deploy/migrations -database "postgres://bot_user:secure_password@localhost:5432/language_exchange_bot?sslmode=disable" down 1
```

## Troubleshooting

### 1. Проверка логов

```bash
# Systemd логи
sudo journalctl -u language-exchange-bot -f

# Docker логи
docker-compose logs -f bot

# Приложение логи
tail -f /opt/language-exchange-bot/logs/app.log
```

### 2. Проверка ресурсов

```bash
# CPU и память
htop

# Дисковое пространство
df -h

# Сетевые соединения
netstat -tulpn | grep :8080
```

### 3. Проверка зависимостей

```bash
# PostgreSQL
pg_isready -h localhost -p 5432

# Redis
redis-cli ping

# Telegram API
curl "https://api.telegram.org/bot${TELEGRAM_BOT_TOKEN}/getMe"
```

## Production Checklist

- [ ] Настроены переменные окружения
- [ ] Применены миграции базы данных
- [ ] Настроен Redis
- [ ] Настроен SSL/TLS
- [ ] Настроены webhook'и
- [ ] Настроен мониторинг
- [ ] Настроены логи
- [ ] Настроены бэкапы
- [ ] Проведено тестирование
- [ ] Документированы процедуры

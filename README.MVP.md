# Language Exchange Bot - MVP

Минимально рабочая версия Language Exchange Bot с микросервисной архитектурой.

## Архитектура MVP

```shell
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   API Gateway   │    │ Profile Service │    │   Bot Service   │
│   (Port 8080)   │◄──►│   (Port 8081)   │    │   (Port 8082)   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
                    ┌─────────────────┐
                    │   PostgreSQL    │
                    │   (Port 5432)   │
                    └─────────────────┘
```

## Сервисы

### 1. API Gateway (Port 8080)

- Единая точка входа для всех запросов
- Маршрутизация к микросервисам
- Rate limiting и логирование
- Health checks

### 2. Profile Service (Port 8081)

- Управление профилями пользователей
- Языки и интересы
- Предпочтения и характеристики
- Доступность по времени

### 3. Bot Service (Port 8082)

- HTTP API для Telegram webhook
- Обработка сообщений и callback queries
- Интеграция с Profile Service

### 4. PostgreSQL (Port 5432)

- Единая база данных для всех сервисов
- Схемы для каждого сервиса

## Быстрый старт

### Предварительные требования

- Docker и Docker Compose
- Go 1.22+ (для разработки)
- Telegram Bot Token (для работы с ботом)
- PowerShell (для Windows скриптов)

### Запуск

1. **Клонируйте репозиторий:**

   ```bash
   git clone <repository-url>
   cd language_exchange_bot
   ```

2. **Создайте .env файл:**

   ```bash
   cp env.example .env
   # Отредактируйте .env файл с вашими настройками
   ```

   **Важно:** Заполните следующие переменные в .env файле:
   - `TELEGRAM_TOKEN` - токен вашего Telegram бота
   - `ADMIN_CHAT_IDS` - ID чатов администраторов (через запятую)
   - `ADMIN_USERNAMES` - usernames администраторов (через запятую)

3. **Запустите все сервисы:**

   ```bash
   make -f Makefile.mvp run
   ```

4. **Проверьте статус сервисов:**

   ```bash
   make -f Makefile.mvp health
   ```

### API Endpoints

#### API Gateway (<http://localhost:8080>)

- `GET /healthz` - Health check
- `GET /readyz` - Readiness check
- `GET /api/v1/users/*` - Profile Service routes
- `POST /api/v1/bot/telegram/webhook` - Telegram webhook

#### Profile Service (<http://localhost:8081>)

- `GET /healthz` - Health check
- `GET /api/v1/users` - List users
- `POST /api/v1/users` - Create user
- `GET /api/v1/users/{id}` - Get user

#### Bot Service (<http://localhost:8082>)

- `GET /healthz` - Health check
- `POST /api/v1/bot/telegram/webhook` - Telegram webhook
- `POST /api/v1/bot/telegram/send` - Send message

## Разработка

### Локальная разработка

1. **Запустите только базу данных:**

   ```bash
   docker-compose -f docker-compose.mvp.yml up postgres -d
   ```

2. **Запустите сервисы локально:**

   ```bash
   # Profile Service
   cd services/profile
   make run

   # Bot Service
   cd services/bot
   make run

   # API Gateway
   cd services/api-gateway
   make run
   ```

### Тестирование

```bash
# Запустить все тесты
make -f Makefile.mvp test

# Тесты отдельного сервиса
cd services/profile && make test
cd services/api-gateway && make test
cd services/bot && make test
```

### Логи

```bash
# Все сервисы
make -f Makefile.mvp logs

# Конкретный сервис
docker-compose -f docker-compose.mvp.yml logs -f profile-service
docker-compose -f docker-compose.mvp.yml logs -f bot-service
docker-compose -f docker-compose.mvp.yml logs -f api-gateway
```

## Конфигурация

### Переменные окружения

#### API Gateway

- `HTTP_PORT` - Порт сервера (по умолчанию: 8080)
- `PROFILE_SERVICE_URL` - URL Profile Service
- `BOT_SERVICE_URL` - URL Bot Service
- `RATE_LIMIT_ENABLED` - Включить rate limiting
- `RATE_LIMIT_RPM` - Запросов в минуту

#### Profile Service

- `DATABASE_URL` - Строка подключения к PostgreSQL
- `DB_SCHEMA` - Схема базы данных
- `HTTP_PORT` - Порт сервера (по умолчанию: 8081)

#### Bot Service

- `DATABASE_URL` - Строка подключения к PostgreSQL
- `HTTP_PORT` - Порт сервера (по умолчанию: 8082)
- `TELEGRAM_TOKEN` - Токен Telegram бота
- `ADMIN_CHAT_IDS` - ID чатов администраторов
- `ADMIN_USERNAMES` - Usernames администраторов

## Мониторинг

### Health Checks

```bash
# API Gateway
curl http://localhost:8080/healthz
curl http://localhost:8080/readyz

# Profile Service
curl http://localhost:8081/healthz

# Bot Service
curl http://localhost:8082/healthz
```

### Логи мониторинга

Все сервисы используют структурированное логирование с Zap.

## Остановка

```bash
# Остановить все сервисы
make -f Makefile.mvp stop

# Очистить все данные
make -f Makefile.mvp clean
```

## Следующие шаги

1. **Matcher Service** - Алгоритм поиска партнеров
2. **Notification Service** - Уведомления пользователей
3. **Analytics Service** - Аналитика и метрики
4. **Аутентификация** - JWT токены
5. **Мониторинг** - Prometheus + Grafana
6. **Трассировка** - Jaeger
7. **Кэширование** - Redis
8. **Message Queue** - RabbitMQ

## Работа с Telegram ботом

### Настройка бота

1. **Создайте бота через @BotFather в Telegram**
2. **Получите токен и добавьте в .env файл:**

   ```env
   TELEGRAM_TOKEN=1234567890:ABCdefGHIjklMNOpqrsTUVwxyz
   ```

3. **Получите ваш Chat ID:**
   - Напишите боту в Telegram
   - Запустите: `.\scripts\get-bot-info.ps1`
   - Найдите ваш Chat ID в выводе

4. **Добавьте Chat ID в .env файл:**

   ```env
   ADMIN_CHAT_IDS=123456789,987654321
   ADMIN_USERNAMES=your_username,admin_username
   ```

### Тестирование бота

```powershell
# Получить информацию о боте
.\scripts\get-bot-info.ps1

# Отправить тестовое сообщение
.\scripts\test-telegram-bot.ps1

# Настроить webhook (для продакшена)
.\scripts\setup-telegram-webhook.ps1 -ApiGatewayUrl "https://your-domain.com"
```

### Просмотр логов

```bash
# Все сервисы
make -f Makefile.mvp logs

# Конкретный сервис
make -f Makefile.mvp logs-bot
make -f Makefile.mvp logs-profile
make -f Makefile.mvp logs-gateway

# PowerShell скрипт
.\scripts\view-logs.ps1 -Service "bot-service" -Follow
```

## Поддержка

При возникновении проблем:

1. Проверьте логи: `make -f Makefile.mvp logs`
2. Проверьте health checks: `make -f Makefile.mvp health`
3. Проверьте конфигурацию бота: `.\scripts\get-bot-info.ps1`
4. Перезапустите сервисы: `make -f Makefile.mvp stop && make -f Makefile.mvp run`

### Частые проблемы

**Бот не отвечает:**

- Проверьте токен в .env файле
- Убедитесь, что `ENABLE_TELEGRAM=true`
- Проверьте логи: `make -f Makefile.mvp logs-bot`

**Ошибки webhook:**

- Убедитесь, что у вас есть публичный HTTPS домен
- Проверьте настройки webhook: `.\scripts\get-bot-info.ps1`

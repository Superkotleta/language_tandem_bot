# API Gateway

API Gateway является единой точкой входа для всех клиентских запросов в микросервисной архитектуре Language Exchange Bot. Он обеспечивает маршрутизацию, аутентификацию, ограничение скорости и балансировку нагрузки.

## Возможности

- **Маршрутизация запросов**: Перенаправление запросов к соответствующим микросервисам
- **Ограничение скорости**: Защита от злоупотреблений с помощью rate limiting
- **Логирование**: Структурированное логирование всех запросов
- **Проверки состояния**: Health check и readiness check эндпоинты
- **Обработка ошибок**: Централизованная обработка ошибок
- **Мониторинг**: Отслеживание состояния backend сервисов

## API Эндпоинты

### Проверки состояния

- `GET /healthz` - Проверка состояния API Gateway
- `GET /readyz` - Проверка готовности (включая backend сервисы)

### Маршрутизация

- `/api/v1/users/*` - Перенаправление к Profile Service
- `/api/v1/languages/*` - Перенаправление к Profile Service
- `/api/v1/interests/*` - Перенаправление к Profile Service
- `/api/v1/preferences/*` - Перенаправление к Profile Service
- `/api/v1/traits/*` - Перенаправление к Profile Service
- `/api/v1/availability/*` - Перенаправление к Profile Service
- `/api/v1/bot/*` - Перенаправление к Bot Service
- `/api/v1/telegram/*` - Перенаправление к Bot Service
- `/api/v1/discord/*` - Перенаправление к Bot Service

## Конфигурация

Сервис может быть настроен с помощью переменных окружения:

- `HTTP_PORT` - Порт HTTP сервера (по умолчанию: 8080)
- `DEBUG` - Режим отладки (по умолчанию: false)
- `PROFILE_SERVICE_URL` - URL Profile Service (по умолчанию: <http://localhost:8081>)
- `PROFILE_SERVICE_TIMEOUT` - Таймаут для Profile Service в секундах (по умолчанию: 30)
- `BOT_SERVICE_URL` - URL Bot Service (по умолчанию: <http://localhost:8082>)
- `BOT_SERVICE_TIMEOUT` - Таймаут для Bot Service в секундах (по умолчанию: 30)
- `RATE_LIMIT_ENABLED` - Включить rate limiting (по умолчанию: true)
- `RATE_LIMIT_RPM` - Запросов в минуту (по умолчанию: 100)
- `RATE_LIMIT_BURST` - Размер burst (по умолчанию: 20)

## Разработка

### Предварительные требования

- Go 1.22+
- Docker (опционально)

### Настройка

1. Клонируйте репозиторий
2. Установите зависимости:

   ```bash
   make deps
   ```

3. Запустите сервис:

   ```bash
   make run
   ```

### Тестирование

Запустите тесты:

```bash
make test
```

### Сборка

Соберите сервис:

```bash
make build
```

### Docker

Соберите и запустите с Docker:

```bash
make docker-build
make docker-run
```

## Архитектура

API Gateway следует паттерну reverse proxy:

- **Handlers**: Обработка HTTP запросов и маршрутизация
- **Proxy**: Перенаправление запросов к backend сервисам
- **Middleware**: Логирование, rate limiting, recovery
- **Server**: HTTP сервер с graceful shutdown

## Зависимости

- **Gin**: HTTP веб-фреймворк
- **Zap**: Структурированное логирование
- **Validator**: Валидация запросов

## Мониторинг

Сервис включает:

- Health check эндпоинты
- Структурированное логирование
- Отслеживание состояния backend сервисов
- Метрики производительности (запланировано)

## Безопасность

- Rate limiting для защиты от DDoS
- Валидация входных данных
- Централизованная обработка ошибок
- Аутентификация/авторизация (запланировано)

## Производительность

- Эффективная маршрутизация запросов
- Connection pooling к backend сервисам
- Graceful shutdown
- Load balancing (запланировано)

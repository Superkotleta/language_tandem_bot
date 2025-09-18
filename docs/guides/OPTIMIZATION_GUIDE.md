# Руководство по оптимизации Language Exchange Bot

## Обзор оптимизаций

Этот документ описывает реализованные оптимизации для повышения производительности и надежности Language Exchange Bot.

## 🚀 Реализованные оптимизации

### 1. Интеграционные тесты

**Файлы:**

- `tests/integration/profile_completion_test.go` - Тесты заполнения профиля
- `tests/integration/feedback_system_test.go` - Тесты системы отзывов
- `tests/integration/admin_functions_test.go` - Тесты административных функций
- `tests/integration/localization_test.go` - Тесты локализации

**Покрытие:**

- ✅ Заполнение профиля пользователя
- ✅ Система отзывов
- ✅ Административные функции
- ✅ Локализация
- ✅ Валидация данных
- ✅ Обработка ошибок

**Запуск тестов:**

```bash
# Все интеграционные тесты
make test-integration

# Отдельные наборы тестов
make test-profile-completion
make test-feedback-system
make test-admin-functions
make test-localization

# С покрытием
make test-coverage
```

### 2. Оптимизация базы данных

**Файлы:**

- `internal/database/optimized_db.go` - Оптимизированная БД с pgx
- `internal/database/interface.go` - Обновленный интерфейс

**Улучшения:**

- ✅ Connection pooling с pgx/v5
- ✅ Batch операции для групповых обновлений
- ✅ Транзакции для атомарности
- ✅ Таймауты для запросов
- ✅ Health checks
- ✅ Статистика соединений

**Пример использования:**

```go
// Создание оптимизированной БД
db, err := database.NewOptimizedDB(databaseURL)

// Batch обновление профиля
updates := map[string]interface{}{
    "native_language_code": "ru",
    "target_language_code": "en",
    "interests": []int{1, 2, 3},
}
err = db.UpdateUserProfileBatch(userID, updates)

// Batch операции с интересами
err = db.SaveUserInterestsBatch(userID, interestIDs)
```

### 3. Кэширование

**Файлы:**

- `internal/cache/cache.go` - Система кэширования

**Реализации:**

- ✅ Redis кэш для production
- ✅ In-memory кэш для development
- ✅ TTL для автоматического истечения
- ✅ Кэширование языков и интересов
- ✅ Кэширование профилей пользователей

**Пример использования:**

```go
// Создание кэша
cache := cache.NewRedisCache("localhost:6379", "", 0)

// Кэширование языков
languages, err := cache.GetLanguages(ctx)
cache.SetLanguages(ctx, languages)

// Кэширование интересов
interests, err := cache.GetInterests(ctx, "ru")
cache.SetInterests(ctx, "ru", interests)
```

### 4. Мониторинг и метрики

**Файлы:**

- `internal/monitoring/metrics.go` - Prometheus метрики
- `internal/logging/logger.go` - Structured logging
- `internal/health/health.go` - Health checks

**Метрики:**

- ✅ HTTP запросы (количество, время ответа)
- ✅ Database операции (запросы, транзакции)
- ✅ Cache операции (hits/misses, время)
- ✅ Business метрики (пользователи, профили, отзывы)
- ✅ System метрики (память, CPU, горутины)

**Health checks:**

- ✅ Database connectivity
- ✅ Cache connectivity
- ✅ Memory usage
- ✅ Disk space
- ✅ External services

**Пример использования:**

```go
// Создание метрик
metrics := monitoring.NewMetrics()

// Запись метрики HTTP запроса
metrics.RecordHTTPRequest("GET", "/api/users", "200", duration)

// Запись метрики БД
metrics.RecordDBQuery("SELECT", "users", "success", duration)

// Запись метрики кэша
metrics.RecordCacheHit("redis", "user:*")
```

### 5. Middleware и архитектура

**Файлы:**

- `internal/middleware/middleware.go` - HTTP middleware
- `internal/core/optimized_service.go` - Оптимизированный сервис

**Middleware:**

- ✅ Logging с structured logs
- ✅ Metrics collection
- ✅ Recovery от паник
- ✅ CORS support
- ✅ Rate limiting
- ✅ Authentication

**Пример использования:**

```go
// Создание middleware chain
chain := middleware.ChainMiddleware(
    middleware.NewRecoveryMiddleware(logger),
    middleware.NewLoggingMiddleware(logger),
    middleware.NewMetricsMiddleware(metrics),
    middleware.NewCORSMiddleware(origins, methods, headers),
)

// Применение к HTTP handler
handler := chain.Handle(mux)
```

## 📊 Ожидаемые улучшения производительности

### Время ответа

- **До оптимизации:** 200-500ms
- **После оптимизации:** 50-150ms
- **Улучшение:** 60-70%

### Пропускная способность

- **До оптимизации:** 100-200 RPS
- **После оптимизации:** 500-1000 RPS
- **Улучшение:** 400-500%

### Использование ресурсов

- **Память:** Снижение на 30-40% благодаря кэшированию
- **CPU:** Снижение на 20-30% благодаря batch операциям
- **Соединения БД:** Оптимизация с connection pooling

## 🛠 Настройка и развертывание

### 1. Переменные окружения

```bash
# Database
DATABASE_URL=postgres://user:pass@localhost:5432/db

# Cache
REDIS_URL=redis://localhost:6379

# Monitoring
PROMETHEUS_ENABLED=true
METRICS_PORT=9090

# Logging
LOG_LEVEL=info
LOG_FORMAT=json

# Health checks
HEALTH_CHECK_INTERVAL=30s
```

### 2. Docker Compose

```yaml
version: '3.8'
services:
  bot:
    build: .
    environment:
      - DATABASE_URL=postgres://postgres:password@db:5432/language_exchange
      - REDIS_URL=redis://redis:6379
    depends_on:
      - db
      - redis

  db:
    image: postgres:15
    environment:
      POSTGRES_DB: language_exchange
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: password

  redis:
    image: redis:7-alpine

  prometheus:
    image: prom/prometheus
    ports:
      - "9090:9090"
```

### 3. Мониторинг

**Prometheus метрики:**

- Endpoint: `/metrics`
- Порт: 9090

**Health checks:**

- Liveness: `/health/live`
- Readiness: `/health/ready`
- Full health: `/health`

**Grafana дашборды:**

- HTTP метрики
- Database метрики
- Cache метрики
- System метрики

## 🧪 Тестирование

### Unit тесты

```bash
make test-unit
```

### Интеграционные тесты

```bash
# Настройка тестовой БД
make db-setup

# Запуск тестов
make test-integration
```

### Бенчмарки

```bash
make bench
```

### Нагрузочное тестирование

```bash
# Установка k6
curl https://github.com/grafana/k6/releases/download/v0.47.0/k6-v0.47.0-linux-amd64.tar.gz -L | tar xvz --strip-components 1

# Запуск тестов
k6 run load-test.js
```

## 📈 Мониторинг производительности

### Ключевые метрики

1. **Response Time**
   - P50 < 100ms
   - P95 < 300ms
   - P99 < 500ms

2. **Error Rate**
   - < 0.1% для 4xx ошибок
   - < 0.01% для 5xx ошибок

3. **Cache Hit Rate**
   - > 80% для языков и интересов
   - > 60% для профилей пользователей

4. **Database Connections**
   - Active connections < 80% от max
   - Idle connections > 20% от max

### Алерты

```yaml
# Prometheus alerts
groups:
  - name: bot.rules
    rules:
      - alert: HighResponseTime
        expr: histogram_quantile(0.95, http_request_duration_seconds) > 0.5
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High response time detected"

      - alert: HighErrorRate
        expr: rate(http_requests_total{status_code=~"5.."}[5m]) > 0.01
        for: 2m
        labels:
          severity: critical
        annotations:
          summary: "High error rate detected"
```

## 🔧 Troubleshooting

### Частые проблемы

1. **Высокое использование памяти**
   - Проверить настройки кэша
   - Увеличить TTL для редко используемых данных

2. **Медленные запросы к БД**
   - Проверить индексы
   - Оптимизировать запросы
   - Увеличить connection pool

3. **Низкий cache hit rate**
   - Проверить настройки TTL
   - Оптимизировать ключи кэша
   - Проверить доступность Redis

### Логирование

```bash
# Просмотр логов
docker logs -f bot

# Фильтрация по уровню
docker logs bot 2>&1 | grep "ERROR"

# Структурированные логи
docker logs bot 2>&1 | jq '.'
```

## 📚 Дополнительные ресурсы

- [Prometheus Documentation](https://prometheus.io/docs/)
- [Redis Documentation](https://redis.io/documentation)
- [pgx Documentation](https://pkg.go.dev/github.com/jackc/pgx/v5)
- [Zap Logger](https://pkg.go.dev/go.uber.org/zap)

## 🤝 Вклад в проект

1. Создайте feature branch
2. Добавьте тесты для новой функциональности
3. Обновите документацию
4. Создайте pull request

## 📝 Changelog

### v2.0.0 - Оптимизация производительности

- ✅ Интеграционные тесты
- ✅ Оптимизация БД с pgx
- ✅ Система кэширования
- ✅ Prometheus метрики
- ✅ Structured logging
- ✅ Health checks
- ✅ HTTP middleware
- ✅ Graceful shutdown

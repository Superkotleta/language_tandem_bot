# Troubleshooting Guide

Руководство по диагностике и решению проблем Language Exchange Bot.

## Общие проблемы

### 1. Приложение не запускается

#### Симптомы
- Сервис не стартует
- Ошибки в логах при запуске
- Health check не проходит

#### Диагностика

```bash
# Проверьте статус сервиса
sudo systemctl status language-exchange-bot

# Проверьте логи
sudo journalctl -u language-exchange-bot -f

# Проверьте конфигурацию
./bot --help
```

#### Возможные причины и решения

**1.1. Неверные переменные окружения**

```bash
# Проверьте .env файл
cat .env

# Проверьте переменные
env | grep -E "(DATABASE|REDIS|TELEGRAM)"

# Решение: Исправьте .env файл
```

**1.2. Проблемы с базой данных**

```bash
# Проверьте подключение к PostgreSQL
pg_isready -h localhost -p 5432

# Проверьте миграции
migrate -path services/deploy/migrations -database "$DATABASE_URL" version

# Решение: Примените миграции
migrate -path services/deploy/migrations -database "$DATABASE_URL" up
```

**1.3. Проблемы с Redis**

```bash
# Проверьте Redis
redis-cli ping

# Проверьте подключение
redis-cli -h localhost -p 6379 ping

# Решение: Перезапустите Redis
sudo systemctl restart redis
```

### 2. Высокое потребление памяти

#### Симптомы
- Медленная работа приложения
- OOM (Out of Memory) ошибки
- Высокое использование swap

#### Диагностика

```bash
# Проверьте использование памяти
free -h
htop

# Проверьте процессы
ps aux --sort=-%mem | head -10

# Проверьте логи OOM
dmesg | grep -i "killed process"
```

#### Решения

**2.1. Оптимизация кэша**

```bash
# Проверьте размер кэша
curl -H "Authorization: Bearer $API_TOKEN" http://localhost:8080/api/v1/cache/stats

# Очистите кэш
curl -X POST -H "Authorization: Bearer $API_TOKEN" http://localhost:8080/api/v1/cache/clear
```

**2.2. Настройка лимитов памяти**

```bash
# Увеличьте лимиты в systemd
sudo systemctl edit language-exchange-bot

# Добавьте:
[Service]
MemoryLimit=2G
MemoryHigh=1.5G
```

### 3. Медленные запросы к базе данных

#### Симптомы
- Высокое время ответа API
- Таймауты запросов
- Медленная работа бота

#### Диагностика

```bash
# Проверьте активные соединения
sudo -u postgres psql -c "SELECT count(*) FROM pg_stat_activity;"

# Проверьте медленные запросы
sudo -u postgres psql -c "SELECT query, mean_time, calls FROM pg_stat_statements ORDER BY mean_time DESC LIMIT 10;"

# Проверьте индексы
sudo -u postgres psql -c "SELECT schemaname, tablename, indexname, idx_scan, idx_tup_read FROM pg_stat_user_indexes ORDER BY idx_scan DESC;"
```

#### Решения

**3.1. Оптимизация запросов**

```sql
-- Добавьте индексы для часто используемых полей
CREATE INDEX CONCURRENTLY idx_users_telegram_id ON users (telegram_id);
CREATE INDEX CONCURRENTLY idx_users_state ON users (state);
CREATE INDEX CONCURRENTLY idx_user_interests_user_id ON user_interests (user_id);
```

**3.2. Настройка connection pool**

```bash
# Увеличьте лимиты соединений в .env
DATABASE_MAX_OPEN_CONNS=50
DATABASE_MAX_IDLE_CONNS=20
```

### 4. Проблемы с Redis

#### Симптомы
- Ошибки подключения к Redis
- Низкий hit ratio кэша
- Медленная работа кэша

#### Диагностика

```bash
# Проверьте статус Redis
redis-cli info

# Проверьте использование памяти
redis-cli info memory

# Проверьте hit ratio
redis-cli info stats | grep keyspace
```

#### Решения

**4.1. Настройка памяти Redis**

```bash
# В /etc/redis/redis.conf
maxmemory 2gb
maxmemory-policy allkeys-lru
```

**4.2. Очистка кэша**

```bash
# Очистите весь кэш
redis-cli FLUSHDB

# Очистите конкретные ключи
redis-cli --scan --pattern "user:*" | xargs redis-cli DEL
```

### 5. Проблемы с Telegram API

#### Симптомы
- Бот не отвечает на сообщения
- Ошибки webhook
- Таймауты при отправке сообщений

#### Диагностика

```bash
# Проверьте токен бота
curl "https://api.telegram.org/bot$TELEGRAM_BOT_TOKEN/getMe"

# Проверьте webhook
curl "https://api.telegram.org/bot$TELEGRAM_BOT_TOKEN/getWebhookInfo"

# Проверьте логи бота
sudo journalctl -u language-exchange-bot | grep -i telegram
```

#### Решения

**5.1. Настройка webhook**

```bash
# Установите webhook
curl -X POST "https://api.telegram.org/bot$TELEGRAM_BOT_TOKEN/setWebhook" \
  -H "Content-Type: application/json" \
  -d '{"url": "https://yourdomain.com/webhook/telegram", "secret_token": "your_secret"}'
```

**5.2. Обработка rate limiting**

```bash
# Добавьте задержки в код
# Или используйте circuit breaker
```

## Специфичные проблемы

### 1. Проблемы с миграциями

#### Симптомы
- Ошибки при применении миграций
- Несоответствие схемы базы данных
- Ошибки при запуске приложения

#### Диагностика

```bash
# Проверьте текущую версию миграций
migrate -path services/deploy/migrations -database "$DATABASE_URL" version

# Проверьте статус миграций
migrate -path services/deploy/migrations -database "$DATABASE_URL" status
```

#### Решения

**1.1. Откат миграций**

```bash
# Откатите на одну версию назад
migrate -path services/deploy/migrations -database "$DATABASE_URL" down 1

# Откатите все миграции
migrate -path services/deploy/migrations -database "$DATABASE_URL" down
```

**1.2. Принудительное применение**

```bash
# Принудительно установите версию
migrate -path services/deploy/migrations -database "$DATABASE_URL" force 1
```

### 2. Проблемы с кэшем

#### Симптомы
- Низкий hit ratio
- Медленная работа кэша
- Ошибки сериализации

#### Диагностика

```bash
# Проверьте статистику кэша
curl -H "Authorization: Bearer $API_TOKEN" http://localhost:8080/api/v1/cache/stats

# Проверьте Redis
redis-cli info stats
redis-cli monitor
```

#### Решения

**2.1. Настройка TTL**

```go
// Увеличьте TTL для статичных данных
config := cache.DefaultConfig()
config.LanguagesTTL = 24 * time.Hour
config.InterestsTTL = 24 * time.Hour
```

**2.2. Очистка проблемных ключей**

```bash
# Найдите проблемные ключи
redis-cli --scan --pattern "*" | head -100

# Удалите их
redis-cli --scan --pattern "problematic:*" | xargs redis-cli DEL
```

### 3. Проблемы с производительностью

#### Симптомы
- Медленные API запросы
- Высокое использование CPU
- Таймауты

#### Диагностика

```bash
# Профилирование CPU
go tool pprof http://localhost:8080/debug/pprof/profile

# Профилирование памяти
go tool pprof http://localhost:8080/debug/pprof/heap

# Проверьте горутины
go tool pprof http://localhost:8080/debug/pprof/goroutine
```

#### Решения

**3.1. Оптимизация запросов**

```go
// Используйте batch operations
users, err := db.BatchGetUsers(ctx, userIDs)

// Используйте connection pooling
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(10)
```

**3.2. Настройка кэша**

```go
// Включите Redis для production
service, err := core.NewBotServiceWithRedis(db, redisURL, "", 0, nil)
```

## Мониторинг и алерты

### 1. Настройка алертов

```yaml
# prometheus/alert_rules.yml
groups:
  - name: language_exchange_bot
    rules:
      - alert: BotDown
        expr: up{job="language-exchange-bot"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Bot is down"
          description: "Language Exchange Bot has been down for more than 1 minute"
```

### 2. Health checks

```bash
#!/bin/bash
# health-check.sh

# Проверка health endpoint
if ! curl -f -s http://localhost:8080/health > /dev/null; then
    echo "Health check failed"
    exit 1
fi

# Проверка метрик
if ! curl -f -s http://localhost:8080/metrics > /dev/null; then
    echo "Metrics endpoint failed"
    exit 1
fi

echo "All checks passed"
exit 0
```

## Логирование и отладка

### 1. Настройка логов

```go
// Структурированные логи
log.WithFields(log.Fields{
    "user_id": userID,
    "action": "user_registration",
    "duration": time.Since(start),
}).Info("User registered")
```

### 2. Отладочные эндпоинты

```bash
# Профилирование
curl http://localhost:8080/debug/pprof/profile > profile.out
go tool pprof profile.out

# Трассировка
curl http://localhost:8080/debug/pprof/trace?seconds=30 > trace.out
go tool trace trace.out
```

### 3. Анализ логов

```bash
# Фильтрация логов по уровню
sudo journalctl -u language-exchange-bot --since "1 hour ago" | grep ERROR

# Поиск по паттерну
sudo journalctl -u language-exchange-bot | grep -i "database"

# Статистика ошибок
sudo journalctl -u language-exchange-bot --since "1 day ago" | grep ERROR | wc -l
```

## Восстановление после сбоев

### 1. Автоматическое восстановление

```bash
# Systemd автоматический перезапуск
sudo systemctl edit language-exchange-bot

# Добавьте:
[Service]
Restart=always
RestartSec=5
StartLimitBurst=3
StartLimitInterval=60s
```

### 2. Ручное восстановление

```bash
# Остановите сервис
sudo systemctl stop language-exchange-bot

# Проверьте конфигурацию
./bot --config-check

# Очистите кэш
redis-cli FLUSHDB

# Перезапустите сервис
sudo systemctl start language-exchange-bot
```

### 3. Восстановление базы данных

```bash
# Создайте бэкап
pg_dump language_exchange_bot > backup_$(date +%Y%m%d_%H%M%S).sql

# Восстановите из бэкапа
psql language_exchange_bot < backup_20240120_143000.sql
```

## Профилактика

### 1. Регулярные проверки

```bash
#!/bin/bash
# daily-health-check.sh

# Проверка дискового пространства
df -h | awk '$5 > 80 {print "Disk space warning: " $0}'

# Проверка памяти
free -h | awk 'NR==2{if($3/$2 > 0.8) print "Memory usage high: " $3/$2*100 "%"}'

# Проверка логов на ошибки
sudo journalctl -u language-exchange-bot --since "1 day ago" | grep ERROR | wc -l
```

### 2. Мониторинг ресурсов

```bash
# Настройте мониторинг в cron
0 */6 * * * /opt/language-exchange-bot/health-check.sh
0 2 * * * /opt/language-exchange-bot/backup-database.sh
```

### 3. Обновления

```bash
# Регулярные обновления системы
sudo apt update && sudo apt upgrade -y

# Обновления приложения
git pull origin main
go build -o bot ./cmd/bot
sudo systemctl restart language-exchange-bot
```

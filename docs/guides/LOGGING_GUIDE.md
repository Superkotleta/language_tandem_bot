# 📊 Руководство по просмотру логов Language Exchange Bot

## 🔍 Где смотреть логи

### 1. Docker Compose (рекомендуется)

#### Все сервисы

```bash
# Логи всех сервисов в реальном времени
docker-compose -f ../deploy/docker-compose.yml logs -f

# Логи за последние 100 строк
docker-compose -f ../deploy/docker-compose.yml logs --tail=100

# Логи за определенный период
docker-compose -f ../deploy/docker-compose.yml logs --since="2025-09-18T10:00:00"
```

#### Конкретные сервисы

```bash
# Логи бота
docker-compose -f ../deploy/docker-compose.yml logs -f bot

# Логи базы данных
docker-compose -f ../deploy/docker-compose.yml logs -f postgres

# Логи Redis
docker-compose -f ../deploy/docker-compose.yml logs -f redis

# Логи PgAdmin
docker-compose -f ../deploy/docker-compose.yml logs -f pgadmin

# Логи Profile сервиса
docker-compose -f ../deploy/docker-compose.yml logs -f profile

# Логи Matcher сервиса
docker-compose -f ../deploy/docker-compose.yml logs -f matcher
```

### 2. Makefile команды

```bash
# Все сервисы
make logs

# Конкретные сервисы
make logs-bot
make logs-db
make logs-redis
```

### 3. Прямые Docker команды

```bash
# Логи контейнера бота
docker logs -f bot

# Логи контейнера PostgreSQL
docker logs -f pg

# Логи контейнера Redis
docker logs -f redis

# Логи с временными метками
docker logs -f -t bot

# Логи за последние 50 строк
docker logs --tail=50 bot
```

## 📋 Типы логов

### 1. Структурированные логи (JSON)

#### Формат лога

```json
{
  "timestamp": "2025-09-18T12:00:00.123Z",
  "level": "info",
  "service": "bot",
  "component": "telegram_handler",
  "message": "User started conversation",
  "user_id": 12345,
  "telegram_id": 12345,
  "username": "testuser",
  "action": "start_command",
  "duration_ms": 150,
  "request_id": "req_abc123"
}
```

#### Уровни логирования

- **DEBUG**: Детальная отладочная информация
- **INFO**: Общая информация о работе
- **WARN**: Предупреждения о потенциальных проблемах
- **ERROR**: Ошибки, которые не останавливают работу
- **FATAL**: Критические ошибки, останавливающие сервис

### 2. Логи по компонентам

#### Telegram Handler

```bash
# Фильтрация логов Telegram
docker-compose logs -f bot | grep "telegram_handler"

# Логи команд пользователей
docker-compose logs -f bot | grep "command"
```

#### Database Operations

```bash
# Логи операций с БД
docker-compose logs -f bot | grep "database"

# Логи медленных запросов
docker-compose logs -f bot | grep "slow_query"
```

#### Cache Operations

```bash
# Логи кэширования
docker-compose logs -f bot | grep "cache"

# Логи Redis
docker-compose logs -f redis
```

#### Localization

```bash
# Логи локализации
docker-compose logs -f bot | grep "localization"

# Ошибки переводов
docker-compose logs -f bot | grep "translation_error"
```

## 🔧 Фильтрация и поиск

### 1. По уровню логирования

```bash
# Только ошибки
docker-compose logs -f bot | grep '"level":"error"'

# Только предупреждения и ошибки
docker-compose logs -f bot | grep -E '"level":"(warn|error)"'

# Исключить debug логи
docker-compose logs -f bot | grep -v '"level":"debug"'
```

### 2. По пользователю

```bash
# Логи конкретного пользователя
docker-compose logs -f bot | grep '"user_id":12345'

# Логи по Telegram ID
docker-compose logs -f bot | grep '"telegram_id":12345'
```

### 3. По действию

```bash
# Логи команды /start
docker-compose logs -f bot | grep '"action":"start_command"'

# Логи заполнения профиля
docker-compose logs -f bot | grep '"action":"profile_update"'

# Логи отправки отзывов
docker-compose logs -f bot | grep '"action":"feedback_submit"'
```

### 4. По времени

```bash
# Логи за последний час
docker-compose logs -f bot --since="1h"

# Логи за определенную дату
docker-compose logs -f bot --since="2025-09-18T00:00:00" --until="2025-09-18T23:59:59"
```

## 📊 Мониторинг в реальном времени

### 1. Множественные логи

```bash
# Логи бота и БД одновременно
docker-compose logs -f bot postgres

# Логи всех сервисов с цветовой кодировкой
docker-compose logs -f --no-log-prefix
```

### 2. Агрегированные метрики

```bash
# Подсчет ошибок за последний час
docker-compose logs --since="1h" bot | grep '"level":"error"' | wc -l

# Топ пользователей по активности
docker-compose logs --since="1h" bot | grep '"action"' | jq -r '.user_id' | sort | uniq -c | sort -nr
```

### 3. Мониторинг производительности

```bash
# Медленные операции (>1 секунды)
docker-compose logs -f bot | grep '"duration_ms":[1-9][0-9][0-9][0-9]'

# Ошибки базы данных
docker-compose logs -f bot | grep "database_error"
```

## 🚨 Алерты и уведомления

### 1. Критические ошибки

```bash
# Мониторинг критических ошибок
docker-compose logs -f bot | grep '"level":"fatal"'

# Ошибки подключения к БД
docker-compose logs -f bot | grep "database_connection_error"
```

### 2. Производительность

```bash
# Медленные запросы
docker-compose logs -f bot | grep '"duration_ms":[5-9][0-9][0-9][0-9]'

# Высокое использование памяти
docker-compose logs -f bot | grep "memory_usage_high"
```

### 3. Безопасность

```bash
# Подозрительная активность
docker-compose logs -f bot | grep "suspicious_activity"

# Множественные неудачные попытки
docker-compose logs -f bot | grep "failed_attempts"
```

## 📁 Сохранение логов

### 1. Экспорт логов

```bash
# Сохранение логов в файл
docker-compose logs bot > bot_logs_$(date +%Y%m%d_%H%M%S).log

# Сохранение логов за определенный период
docker-compose logs --since="2025-09-18T00:00:00" bot > daily_logs.log
```

### 2. Ротация логов

```bash
# Настройка logrotate для Docker
sudo nano /etc/logrotate.d/docker-logs

# Содержимое:
/var/lib/docker/containers/*/*.log {
    daily
    missingok
    rotate 7
    compress
    delaycompress
    notifempty
    create 644 root root
}
```

## 🔍 Анализ логов

### 1. Использование jq для JSON логов

```bash
# Установка jq
sudo apt install jq

# Красивый вывод JSON
docker-compose logs -f bot | jq '.'

# Фильтрация по полям
docker-compose logs -f bot | jq 'select(.level == "error")'

# Группировка по действиям
docker-compose logs -f bot | jq -r '.action' | sort | uniq -c
```

### 2. Статистика

```bash
# Статистика по уровням логирования
docker-compose logs --since="1h" bot | jq -r '.level' | sort | uniq -c

# Статистика по пользователям
docker-compose logs --since="1h" bot | jq -r '.user_id' | sort | uniq -c | sort -nr | head -10

# Статистика по действиям
docker-compose logs --since="1h" bot | jq -r '.action' | sort | uniq -c | sort -nr
```

### 3. Поиск паттернов

```bash
# Поиск повторяющихся ошибок
docker-compose logs --since="1h" bot | jq -r 'select(.level == "error") | .message' | sort | uniq -c | sort -nr

# Анализ времени отклика
docker-compose logs --since="1h" bot | jq -r 'select(.duration_ms) | .duration_ms' | sort -n | tail -10
```

## 🛠️ Инструменты для анализа

### 1. ELK Stack (опционально)

```yaml
# docker-compose.override.yml
version: '3.8'
services:
  elasticsearch:
    image: docker.elastic.co/elasticsearch/elasticsearch:7.15.0
    environment:
      - discovery.type=single-node
    ports:
      - "9200:9200"
  
  logstash:
    image: docker.elastic.co/logstash/logstash:7.15.0
    volumes:
      - ./logstash.conf:/usr/share/logstash/pipeline/logstash.conf
  
  kibana:
    image: docker.elastic.co/kibana/kibana:7.15.0
    ports:
      - "5601:5601"
```

### 2. Grafana Loki (альтернатива)

```yaml
# docker-compose.override.yml
version: '3.8'
services:
  loki:
    image: grafana/loki:2.6.1
    ports:
      - "3100:3100"
  
  promtail:
    image: grafana/promtail:2.6.1
    volumes:
      - /var/lib/docker/containers:/var/lib/docker/containers:ro
      - /var/run/docker.sock:/var/run/docker.sock
```

## 📞 Поддержка

### Полезные команды для диагностики

```bash
# Проверка статуса всех сервисов
docker-compose ps

# Проверка использования ресурсов
docker stats

# Проверка сетевых соединений
docker network ls
docker network inspect deploy_app-network

# Проверка томов
docker volume ls
```

### Контакты для поддержки

- **Логи бота**: `make logs-bot`
- **Логи БД**: `make logs-db`
- **Мониторинг**: `make monitor`
- **Health checks**: `curl http://localhost:8080/health`

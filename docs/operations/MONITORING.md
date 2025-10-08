# Monitoring Guide

Руководство по настройке мониторинга для Language Exchange Bot.

## Обзор мониторинга

### Ключевые метрики

- **Application Metrics**: Response time, request rate, error rate
- **Infrastructure Metrics**: CPU, memory, disk, network
- **Database Metrics**: Connection pool, query performance, slow queries
- **Cache Metrics**: Hit ratio, memory usage, evictions
- **Business Metrics**: User registrations, matches, feedback

## Prometheus + Grafana

### 1. Установка Prometheus

```bash
# Создайте пользователя
sudo useradd --no-create-home --shell /bin/false prometheus

# Создайте директории
sudo mkdir /etc/prometheus
sudo mkdir /var/lib/prometheus
sudo chown prometheus:prometheus /etc/prometheus
sudo chown prometheus:prometheus /var/lib/prometheus

# Скачайте и установите Prometheus
cd /tmp
wget https://github.com/prometheus/prometheus/releases/download/v2.45.0/prometheus-2.45.0.linux-amd64.tar.gz
tar xvf prometheus-2.45.0.linux-amd64.tar.gz
sudo cp prometheus-2.45.0.linux-amd64/prometheus /usr/local/bin/
sudo cp prometheus-2.45.0.linux-amd64/promtool /usr/local/bin/
sudo chown prometheus:prometheus /usr/local/bin/prometheus
sudo chown prometheus:prometheus /usr/local/bin/promtool
```

### 2. Конфигурация Prometheus

Создайте `/etc/prometheus/prometheus.yml`:

```yaml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

rule_files:
  - "alert_rules.yml"

alerting:
  alertmanagers:
    - static_configs:
        - targets:
          - localhost:9093

scrape_configs:
  - job_name: 'prometheus'
    static_configs:
      - targets: ['localhost:9090']

  - job_name: 'language-exchange-bot'
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: '/metrics'
    scrape_interval: 10s

  - job_name: 'postgres'
    static_configs:
      - targets: ['localhost:9187']

  - job_name: 'redis'
    static_configs:
      - targets: ['localhost:9121']

  - job_name: 'node-exporter'
    static_configs:
      - targets: ['localhost:9100']
```

### 3. Systemd сервис для Prometheus

Создайте `/etc/systemd/system/prometheus.service`:

```ini
[Unit]
Description=Prometheus
Wants=network-online.target
After=network-online.target

[Service]
User=prometheus
Group=prometheus
Type=simple
ExecStart=/usr/local/bin/prometheus \
    --config.file /etc/prometheus/prometheus.yml \
    --storage.tsdb.path /var/lib/prometheus/ \
    --web.console.templates=/etc/prometheus/consoles \
    --web.console.libraries=/etc/prometheus/console_libraries \
    --web.listen-address=0.0.0.0:9090 \
    --web.enable-lifecycle

[Install]
WantedBy=multi-user.target
```

### 4. Установка Grafana

```bash
# Добавьте репозиторий Grafana
sudo apt-get install -y software-properties-common
sudo add-apt-repository "deb https://packages.grafana.com/oss/deb stable main"
wget -q -O - https://packages.grafana.com/gpg.key | sudo apt-key add -

# Установите Grafana
sudo apt-get update
sudo apt-get install grafana

# Запустите сервис
sudo systemctl daemon-reload
sudo systemctl start grafana-server
sudo systemctl enable grafana-server
```

## Экспортеры метрик

### 1. Node Exporter (системные метрики)

```bash
# Установите node_exporter
wget https://github.com/prometheus/node_exporter/releases/download/v1.6.1/node_exporter-1.6.1.linux-amd64.tar.gz
tar xvf node_exporter-1.6.1.linux-amd64.tar.gz
sudo cp node_exporter-1.6.1.linux-amd64/node_exporter /usr/local/bin/
sudo chown node_exporter:node_exporter /usr/local/bin/node_exporter

# Создайте пользователя
sudo useradd --no-create-home --shell /bin/false node_exporter

# Systemd сервис
sudo tee /etc/systemd/system/node_exporter.service > /dev/null <<EOF
[Unit]
Description=Node Exporter
Wants=network-online.target
After=network-online.target

[Service]
User=node_exporter
Group=node_exporter
Type=simple
ExecStart=/usr/local/bin/node_exporter

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl start node_exporter
sudo systemctl enable node_exporter
```

### 2. PostgreSQL Exporter

```bash
# Установите postgres_exporter
wget https://github.com/prometheus-community/postgres_exporter/releases/download/v0.13.2/postgres_exporter-0.13.2.linux-amd64.tar.gz
tar xvf postgres_exporter-0.13.2.linux-amd64.tar.gz
sudo cp postgres_exporter-0.13.2.linux-amd64/postgres_exporter /usr/local/bin/

# Создайте пользователя
sudo useradd --no-create-home --shell /bin/false postgres_exporter

# Systemd сервис
sudo tee /etc/systemd/system/postgres_exporter.service > /dev/null <<EOF
[Unit]
Description=PostgreSQL Exporter
Wants=network-online.target
After=network-online.target

[Service]
User=postgres_exporter
Group=postgres_exporter
Type=simple
ExecStart=/usr/local/bin/postgres_exporter
Environment=DATA_SOURCE_NAME="postgresql://bot_user:secure_password@localhost:5432/language_exchange_bot?sslmode=disable"

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl start postgres_exporter
sudo systemctl enable postgres_exporter
```

### 3. Redis Exporter

```bash
# Установите redis_exporter
wget https://github.com/oliver006/redis_exporter/releases/download/v1.52.0/redis_exporter-v1.52.0.linux-amd64.tar.gz
tar xvf redis_exporter-v1.52.0.linux-amd64.tar.gz
sudo cp redis_exporter-v1.52.0.linux-amd64/redis_exporter /usr/local/bin/

# Создайте пользователя
sudo useradd --no-create-home --shell /bin/false redis_exporter

# Systemd сервис
sudo tee /etc/systemd/system/redis_exporter.service > /dev/null <<EOF
[Unit]
Description=Redis Exporter
Wants=network-online.target
After=network-online.target

[Service]
User=redis_exporter
Group=redis_exporter
Type=simple
ExecStart=/usr/local/bin/redis_exporter --redis.addr=redis://localhost:6379

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl start redis_exporter
sudo systemctl enable redis_exporter
```

## Алерты

### 1. Правила алертов

Создайте `/etc/prometheus/alert_rules.yml`:

```yaml
groups:
  - name: language_exchange_bot
    rules:
      - alert: BotDown
        expr: up{job="language-exchange-bot"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Language Exchange Bot is down"
          description: "Bot has been down for more than 1 minute"

      - alert: HighErrorRate
        expr: rate(http_requests_total{status=~"5.."}[5m]) > 0.1
        for: 2m
        labels:
          severity: warning
        annotations:
          summary: "High error rate detected"
          description: "Error rate is {{ $value }} errors per second"

      - alert: HighResponseTime
        expr: histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m])) > 1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High response time"
          description: "95th percentile response time is {{ $value }} seconds"

      - alert: DatabaseConnectionsHigh
        expr: pg_stat_database_numbackends / pg_settings_max_connections > 0.8
        for: 2m
        labels:
          severity: warning
        annotations:
          summary: "High database connection usage"
          description: "Database connections are at {{ $value }}% of maximum"

      - alert: RedisMemoryHigh
        expr: redis_memory_used_bytes / redis_memory_max_bytes > 0.9
        for: 2m
        labels:
          severity: warning
        annotations:
          summary: "Redis memory usage high"
          description: "Redis memory usage is {{ $value }}% of maximum"

      - alert: CacheHitRatioLow
        expr: cache_hit_ratio < 0.8
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Low cache hit ratio"
          description: "Cache hit ratio is {{ $value }}%"

      - alert: DiskSpaceLow
        expr: (node_filesystem_avail_bytes / node_filesystem_size_bytes) < 0.1
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "Low disk space"
          description: "Disk space is {{ $value }}% full"
```

### 2. Alertmanager

```bash
# Установите Alertmanager
wget https://github.com/prometheus/alertmanager/releases/download/v0.25.0/alertmanager-0.25.0.linux-amd64.tar.gz
tar xvf alertmanager-0.25.0.linux-amd64.tar.gz
sudo cp alertmanager-0.25.0.linux-amd64/alertmanager /usr/local/bin/
sudo cp alertmanager-0.25.0.linux-amd64/amtool /usr/local/bin/

# Создайте пользователя
sudo useradd --no-create-home --shell /bin/false alertmanager

# Создайте директории
sudo mkdir /etc/alertmanager
sudo mkdir /var/lib/alertmanager
sudo chown alertmanager:alertmanager /etc/alertmanager
sudo chown alertmanager:alertmanager /var/lib/alertmanager
```

Создайте `/etc/alertmanager/alertmanager.yml`:

```yaml
global:
  smtp_smarthost: 'localhost:587'
  smtp_from: 'alerts@yourdomain.com'

route:
  group_by: ['alertname']
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 1h
  receiver: 'web.hook'

receivers:
  - name: 'web.hook'
    webhook_configs:
      - url: 'http://localhost:5001/'

  - name: 'email'
    email_configs:
      - to: 'admin@yourdomain.com'
        subject: 'Language Exchange Bot Alert'
        body: |
          {{ range .Alerts }}
          Alert: {{ .Annotations.summary }}
          Description: {{ .Annotations.description }}
          {{ end }}
```

## Grafana Dashboards

### 1. Application Dashboard

Создайте дашборд с панелями:

- **Request Rate**: `rate(http_requests_total[5m])`
- **Response Time**: `histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))`
- **Error Rate**: `rate(http_requests_total{status=~"5.."}[5m])`
- **Active Users**: `cache_users_count`
- **Cache Hit Ratio**: `cache_hit_ratio`

### 2. Infrastructure Dashboard

- **CPU Usage**: `100 - (avg by (instance) (rate(node_cpu_seconds_total{mode="idle"}[5m])) * 100)`
- **Memory Usage**: `(1 - (node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes)) * 100`
- **Disk Usage**: `100 - ((node_filesystem_avail_bytes * 100) / node_filesystem_size_bytes)`
- **Network I/O**: `rate(node_network_receive_bytes_total[5m])`

### 3. Database Dashboard

- **Active Connections**: `pg_stat_database_numbackends`
- **Slow Queries**: `pg_stat_database_tup_returned`
- **Cache Hit Ratio**: `pg_stat_database_blks_hit / (pg_stat_database_blks_hit + pg_stat_database_blks_read)`

### 4. Redis Dashboard

- **Memory Usage**: `redis_memory_used_bytes`
- **Hit Ratio**: `redis_keyspace_hits_total / (redis_keyspace_hits_total + redis_keyspace_misses_total)`
- **Connected Clients**: `redis_connected_clients`

## Логирование

### 1. Структурированные логи

```go
// Пример структурированного лога
log.WithFields(log.Fields{
    "user_id": userID,
    "action": "user_registration",
    "duration": time.Since(start),
    "success": true,
}).Info("User registered successfully")
```

### 2. ELK Stack (опционально)

```yaml
# docker-compose.yml для ELK
version: '3.8'
services:
  elasticsearch:
    image: elasticsearch:8.8.0
    environment:
      - discovery.type=single-node
      - xpack.security.enabled=false
    ports:
      - "9200:9200"

  logstash:
    image: logstash:8.8.0
    volumes:
      - ./logstash.conf:/usr/share/logstash/pipeline/logstash.conf
    ports:
      - "5044:5044"

  kibana:
    image: kibana:8.8.0
    ports:
      - "5601:5601"
```

## Health Checks

### 1. Application Health

```bash
#!/bin/bash
# /opt/language-exchange-bot/health-check.sh

HEALTH_URL="http://localhost:8080/health"
METRICS_URL="http://localhost:8080/metrics"

# Проверка health endpoint
if ! curl -f -s $HEALTH_URL > /dev/null; then
    echo "Health check failed"
    exit 1
fi

# Проверка метрик
if ! curl -f -s $METRICS_URL > /dev/null; then
    echo "Metrics endpoint failed"
    exit 1
fi

echo "All checks passed"
exit 0
```

### 2. Cron job для мониторинга

```bash
# Добавьте в crontab
*/5 * * * * /opt/language-exchange-bot/health-check.sh || systemctl restart language-exchange-bot
```

## Уведомления

### 1. Slack интеграция

```yaml
# alertmanager.yml
receivers:
  - name: 'slack'
    slack_configs:
      - api_url: 'https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK'
        channel: '#alerts'
        title: 'Language Exchange Bot Alert'
        text: |
          {{ range .Alerts }}
          *{{ .Annotations.summary }}*
          {{ .Annotations.description }}
          {{ end }}
```

### 2. Email уведомления

```yaml
receivers:
  - name: 'email'
    email_configs:
      - to: 'admin@yourdomain.com'
        subject: 'Language Exchange Bot Alert'
        body: |
          Alert: {{ .GroupLabels.alertname }}
          {{ range .Alerts }}
          Summary: {{ .Annotations.summary }}
          Description: {{ .Annotations.description }}
          {{ end }}
```

## Troubleshooting

### 1. Проверка метрик

```bash
# Проверьте Prometheus targets
curl http://localhost:9090/api/v1/targets

# Проверьте конкретную метрику
curl "http://localhost:9090/api/v1/query?query=up"

# Проверьте алерты
curl http://localhost:9090/api/v1/alerts
```

### 2. Отладка алертов

```bash
# Проверьте конфигурацию алертов
promtool check rules /etc/prometheus/alert_rules.yml

# Проверьте конфигурацию Alertmanager
amtool config validate /etc/alertmanager/alertmanager.yml
```

### 3. Логи мониторинга

```bash
# Prometheus логи
sudo journalctl -u prometheus -f

# Alertmanager логи
sudo journalctl -u alertmanager -f

# Grafana логи
sudo journalctl -u grafana-server -f
```

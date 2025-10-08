# Scaling Guide

Руководство по масштабированию Language Exchange Bot для высоких нагрузок.

## Стратегии масштабирования

### 1. Горизонтальное масштабирование

- **Load Balancer**: Распределение нагрузки между инстансами
- **Database Sharding**: Разделение данных по шардам
- **Cache Clustering**: Кластеризация Redis
- **Microservices**: Разделение на микросервисы

### 2. Вертикальное масштабирование

- **Увеличение ресурсов**: CPU, RAM, Storage
- **Оптимизация кода**: Профилирование и оптимизация
- **Database Tuning**: Настройка PostgreSQL
- **Cache Optimization**: Оптимизация Redis

## Load Balancer Setup

### 1. Nginx Load Balancer

```nginx
# /etc/nginx/sites-available/language-exchange-bot
upstream bot_backend {
    least_conn;
    server 127.0.0.1:8080 weight=3 max_fails=3 fail_timeout=30s;
    server 127.0.0.1:8081 weight=3 max_fails=3 fail_timeout=30s;
    server 127.0.0.1:8082 weight=2 max_fails=3 fail_timeout=30s;
    keepalive 32;
}

server {
    listen 80;
    server_name yourdomain.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name yourdomain.com;

    ssl_certificate /etc/letsencrypt/live/yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/yourdomain.com/privkey.pem;

    # Health check endpoint
    location /health {
        access_log off;
        proxy_pass http://bot_backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # API endpoints
    location /api/ {
        proxy_pass http://bot_backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # Timeouts
        proxy_connect_timeout 5s;
        proxy_send_timeout 10s;
        proxy_read_timeout 10s;
        
        # Buffer settings
        proxy_buffering on;
        proxy_buffer_size 4k;
        proxy_buffers 8 4k;
    }

    # Webhook endpoints
    location /webhook/ {
        proxy_pass http://bot_backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # Webhook specific settings
        proxy_connect_timeout 10s;
        proxy_send_timeout 30s;
        proxy_read_timeout 30s;
    }
}
```

### 2. HAProxy Configuration

```bash
# /etc/haproxy/haproxy.cfg
global
    daemon
    user haproxy
    group haproxy
    log 127.0.0.1:514 local0
    chroot /var/lib/haproxy
    stats socket /run/haproxy/admin.sock mode 660 level admin
    stats timeout 30s

defaults
    mode http
    timeout connect 5000ms
    timeout client 50000ms
    timeout server 50000ms
    option httplog
    option dontlognull
    option redispatch
    retries 3
    maxconn 2000

frontend bot_frontend
    bind *:80
    bind *:443 ssl crt /etc/ssl/certs/yourdomain.com.pem
    redirect scheme https if !{ ssl_fc }
    
    # Health check
    acl is_health_check path_beg /health
    use_backend health_backend if is_health_check
    
    # API endpoints
    acl is_api path_beg /api/
    use_backend api_backend if is_api
    
    # Webhook endpoints
    acl is_webhook path_beg /webhook/
    use_backend webhook_backend if is_webhook
    
    default_backend bot_backend

backend health_backend
    balance roundrobin
    option httpchk GET /health
    server bot1 127.0.0.1:8080 check
    server bot2 127.0.0.1:8081 check
    server bot3 127.0.0.1:8082 check

backend api_backend
    balance leastconn
    option httpchk GET /health
    server bot1 127.0.0.1:8080 check
    server bot2 127.0.0.1:8081 check
    server bot3 127.0.0.1:8082 check

backend webhook_backend
    balance roundrobin
    option httpchk GET /health
    server bot1 127.0.0.1:8080 check
    server bot2 127.0.0.1:8081 check
    server bot3 127.0.0.1:8082 check

backend bot_backend
    balance roundrobin
    option httpchk GET /health
    server bot1 127.0.0.1:8080 check
    server bot2 127.0.0.1:8081 check
    server bot3 127.0.0.1:8082 check

listen stats
    bind *:8404
    stats enable
    stats uri /stats
    stats refresh 30s
    stats admin if TRUE
```

## Database Scaling

### 1. Read Replicas

```bash
# Настройка read replica в PostgreSQL
# На master сервере
sudo -u postgres psql -c "CREATE USER replica_user WITH REPLICATION ENCRYPTED PASSWORD 'replica_password';"
sudo -u postgres psql -c "GRANT CONNECTION ON DATABASE language_exchange_bot TO replica_user;"

# В postgresql.conf на master
wal_level = replica
max_wal_senders = 3
max_replication_slots = 3

# В pg_hba.conf на master
host replication replica_user 0.0.0.0/0 md5

# На replica сервере
sudo -u postgres pg_basebackup -h master_host -D /var/lib/postgresql/15/main -U replica_user -v -P -W
```

### 2. Connection Pooling

```bash
# Установите PgBouncer
sudo apt install pgbouncer

# Конфигурация /etc/pgbouncer/pgbouncer.ini
[databases]
language_exchange_bot = host=localhost port=5432 dbname=language_exchange_bot

[pgbouncer]
listen_port = 6432
listen_addr = 127.0.0.1
auth_type = md5
auth_file = /etc/pgbouncer/userlist.txt
pool_mode = transaction
max_client_conn = 1000
default_pool_size = 25
```

### 3. Database Sharding

```go
// Пример шардинга по user_id
func getShard(userID int64) string {
    shardNumber := userID % 4 // 4 шарда
    return fmt.Sprintf("shard_%d", shardNumber)
}

func getDatabaseConnection(userID int64) *sql.DB {
    shard := getShard(userID)
    return databaseConnections[shard]
}
```

## Redis Clustering

### 1. Redis Cluster Setup

```bash
# Создайте 6 Redis инстансов (3 master + 3 replica)
mkdir -p /opt/redis/cluster/{7000,7001,7002,7003,7004,7005}

# Конфигурация для каждого инстанса
# /opt/redis/cluster/7000/redis.conf
port 7000
cluster-enabled yes
cluster-config-file nodes-7000.conf
cluster-node-timeout 5000
appendonly yes
```

### 2. Redis Sentinel

```bash
# Конфигурация Sentinel
# /etc/redis/sentinel.conf
port 26379
sentinel monitor mymaster 127.0.0.1 6379 2
sentinel down-after-milliseconds mymaster 30000
sentinel parallel-syncs mymaster 1
sentinel failover-timeout mymaster 180000
```

### 3. Redis Cluster в коде

```go
// Настройка Redis Cluster
func NewRedisClusterClient() *redis.ClusterClient {
    return redis.NewClusterClient(&redis.ClusterOptions{
        Addrs: []string{
            "localhost:7000",
            "localhost:7001", 
            "localhost:7002",
            "localhost:7003",
            "localhost:7004",
            "localhost:7005",
        },
        PoolSize: 100,
        MinIdleConns: 10,
        MaxRetries: 3,
        DialTimeout: 5 * time.Second,
        ReadTimeout: 3 * time.Second,
        WriteTimeout: 3 * time.Second,
    })
}
```

## Microservices Architecture

### 1. Разделение на микросервисы

```
language-exchange-bot/
├── services/
│   ├── bot-service/          # Основной бот сервис
│   ├── user-service/         # Управление пользователями
│   ├── matching-service/     # Алгоритм матчинга
│   ├── notification-service/ # Уведомления
│   └── analytics-service/    # Аналитика
├── shared/
│   ├── database/            # Общие DB операции
│   ├── cache/              # Общий кэш
│   └── messaging/          # Обмен сообщениями
└── infrastructure/
    ├── load-balancer/       # Load balancer
    ├── monitoring/         # Мониторинг
    └── logging/            # Централизованное логирование
```

### 2. Service Discovery

```yaml
# docker-compose.yml для микросервисов
version: '3.8'

services:
  consul:
    image: consul:latest
    ports:
      - "8500:8500"
    command: agent -server -bootstrap-expect=1 -ui -client=0.0.0.0

  bot-service:
    build: ./services/bot-service
    environment:
      - CONSUL_HOST=consul:8500
      - SERVICE_NAME=bot-service
      - SERVICE_PORT=8080
    depends_on:
      - consul

  user-service:
    build: ./services/user-service
    environment:
      - CONSUL_HOST=consul:8500
      - SERVICE_NAME=user-service
      - SERVICE_PORT=8081
    depends_on:
      - consul
```

### 3. Message Queue

```yaml
# RabbitMQ для асинхронной обработки
services:
  rabbitmq:
    image: rabbitmq:3-management
    environment:
      RABBITMQ_DEFAULT_USER: admin
      RABBITMQ_DEFAULT_PASS: password
    ports:
      - "5672:5672"
      - "15672:15672"

  bot-service:
    environment:
      - RABBITMQ_URL=amqp://admin:password@rabbitmq:5672/
```

## Container Orchestration

### 1. Docker Swarm

```bash
# Инициализация Swarm
docker swarm init

# Создайте overlay network
docker network create --driver overlay bot-network

# Разверните сервисы
docker service create \
  --name bot-service \
  --replicas 3 \
  --network bot-network \
  --publish 8080:8080 \
  your-registry/language-exchange-bot:latest
```

### 2. Kubernetes

```yaml
# bot-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: bot-service
spec:
  replicas: 3
  selector:
    matchLabels:
      app: bot-service
  template:
    metadata:
      labels:
        app: bot-service
    spec:
      containers:
      - name: bot-service
        image: your-registry/language-exchange-bot:latest
        ports:
        - containerPort: 8080
        env:
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: database-secret
              key: url
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
---
apiVersion: v1
kind: Service
metadata:
  name: bot-service
spec:
  selector:
    app: bot-service
  ports:
  - port: 80
    targetPort: 8080
  type: LoadBalancer
```

## Performance Optimization

### 1. Application Level

```go
// Connection pooling
func NewDBWithPool(databaseURL string) (*DB, error) {
    conn, err := sql.Open("postgres", databaseURL)
    if err != nil {
        return nil, err
    }
    
    // Оптимизированные настройки
    conn.SetMaxOpenConns(100)        // Увеличено для высокой нагрузки
    conn.SetMaxIdleConns(50)         // Больше idle соединений
    conn.SetConnMaxLifetime(30 * time.Minute)
    conn.SetConnMaxIdleTime(15 * time.Minute)
    
    return &DB{conn: conn}, nil
}

// Batch operations
func (db *DB) BatchInsertUsers(users []*models.User) error {
    tx, err := db.conn.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()
    
    stmt, err := tx.Prepare("INSERT INTO users (...) VALUES (...)")
    if err != nil {
        return err
    }
    defer stmt.Close()
    
    for _, user := range users {
        _, err = stmt.Exec(user.TelegramID, user.Username, ...)
        if err != nil {
            return err
        }
    }
    
    return tx.Commit()
}
```

### 2. Database Optimization

```sql
-- Партиционирование больших таблиц
CREATE TABLE user_interests_partitioned (
    LIKE user_interests INCLUDING ALL
) PARTITION BY HASH (user_id);

CREATE TABLE user_interests_0 PARTITION OF user_interests_partitioned
    FOR VALUES WITH (modulus 4, remainder 0);

-- Материализованные представления для аналитики
CREATE MATERIALIZED VIEW user_stats AS
SELECT 
    DATE_TRUNC('day', created_at) as date,
    COUNT(*) as new_users,
    COUNT(*) FILTER (WHERE profile_completion_level = 100) as completed_profiles
FROM users 
GROUP BY DATE_TRUNC('day', created_at);

-- Индексы для оптимизации
CREATE INDEX CONCURRENTLY idx_users_created_at ON users (created_at);
CREATE INDEX CONCURRENTLY idx_users_status_created ON users (status, created_at);
```

### 3. Cache Optimization

```go
// Redis Cluster с оптимизированными настройками
func NewOptimizedRedisCluster() *redis.ClusterClient {
    return redis.NewClusterClient(&redis.ClusterOptions{
        Addrs: []string{"localhost:7000", "localhost:7001", "localhost:7002"},
        PoolSize: 200,                    // Увеличено для высокой нагрузки
        MinIdleConns: 50,                 // Больше idle соединений
        MaxRetries: 3,
        DialTimeout: 5 * time.Second,
        ReadTimeout: 3 * time.Second,
        WriteTimeout: 3 * time.Second,
        IdleTimeout: 5 * time.Minute,
        IdleCheckFrequency: 1 * time.Minute,
    })
}

// Pipeline operations для массовых операций
func (r *RedisCluster) BatchSet(keys []string, values []interface{}) error {
    pipe := r.Pipeline()
    
    for i, key := range keys {
        pipe.Set(context.Background(), key, values[i], time.Hour)
    }
    
    _, err := pipe.Exec(context.Background())
    return err
}
```

## Monitoring at Scale

### 1. Distributed Tracing

```go
// Jaeger tracing
import "github.com/opentracing/opentracing-go"

func (s *BotService) GetUserWithTracing(ctx context.Context, userID int64) (*models.User, error) {
    span, ctx := opentracing.StartSpanFromContext(ctx, "GetUser")
    defer span.Finish()
    
    span.SetTag("user_id", userID)
    
    user, err := s.DB.GetUser(userID)
    if err != nil {
        span.SetTag("error", true)
        span.LogFields(log.Error(err))
        return nil, err
    }
    
    span.SetTag("user_found", user != nil)
    return user, nil
}
```

### 2. Metrics Collection

```go
// Prometheus metrics
import "github.com/prometheus/client_golang/prometheus"

var (
    requestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "http_request_duration_seconds",
            Help: "HTTP request duration in seconds",
        },
        []string{"method", "endpoint", "status"},
    )
    
    activeUsers = prometheus.NewGauge(
        prometheus.GaugeOpts{
            Name: "active_users_total",
            Help: "Number of active users",
        },
    )
)

func init() {
    prometheus.MustRegister(requestDuration)
    prometheus.MustRegister(activeUsers)
}
```

## Auto-scaling

### 1. Horizontal Pod Autoscaler (Kubernetes)

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: bot-service-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: bot-service
  minReplicas: 3
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
```

### 2. Custom Auto-scaling

```bash
#!/bin/bash
# auto-scale.sh

# Получите метрики
CPU_USAGE=$(curl -s http://localhost:9090/api/v1/query?query=rate\(container_cpu_usage_seconds_total\[5m\]\) | jq -r '.data.result[0].value[1]')
MEMORY_USAGE=$(curl -s http://localhost:9090/api/v1/query?query=container_memory_usage_bytes | jq -r '.data.result[0].value[1]')

# Масштабирование на основе метрик
if (( $(echo "$CPU_USAGE > 0.8" | bc -l) )); then
    echo "High CPU usage detected, scaling up..."
    kubectl scale deployment bot-service --replicas=5
elif (( $(echo "$CPU_USAGE < 0.3" | bc -l) )); then
    echo "Low CPU usage detected, scaling down..."
    kubectl scale deployment bot-service --replicas=2
fi
```

## Disaster Recovery

### 1. Multi-region Deployment

```yaml
# Multi-region Kubernetes deployment
apiVersion: v1
kind: ConfigMap
metadata:
  name: bot-config
data:
  primary_region: "us-east-1"
  backup_region: "eu-west-1"
  failover_threshold: "5m"
```

### 2. Database Replication

```bash
# Настройка streaming replication
# На master
sudo -u postgres psql -c "CREATE USER replica_user WITH REPLICATION ENCRYPTED PASSWORD 'password';"
sudo -u postgres psql -c "GRANT CONNECTION ON DATABASE language_exchange_bot TO replica_user;"

# На standby
sudo -u postgres pg_basebackup -h master_host -D /var/lib/postgresql/15/main -U replica_user -v -P -W -R
```

## Best Practices

### 1. Performance Testing

```bash
# Load testing с Apache Bench
ab -n 10000 -c 100 http://localhost:8080/api/v1/stats

# Load testing с wrk
wrk -t12 -c400 -d30s http://localhost:8080/api/v1/stats

# Stress testing
stress-ng --cpu 4 --io 2 --vm 1 --vm-bytes 1G --timeout 60s
```

### 2. Capacity Planning

```bash
# Мониторинг ресурсов
# CPU: < 70% utilization
# Memory: < 80% utilization  
# Disk I/O: < 80% utilization
# Network: < 80% bandwidth

# Формула для расчета capacity
# Required Instances = (Peak Load / Instance Capacity) * Safety Factor
# Safety Factor = 1.2-1.5
```

### 3. Monitoring Checklist

- [ ] CPU utilization < 70%
- [ ] Memory usage < 80%
- [ ] Disk I/O < 80%
- [ ] Network bandwidth < 80%
- [ ] Database connections < 80% of max
- [ ] Cache hit ratio > 80%
- [ ] Response time P95 < 500ms
- [ ] Error rate < 1%
- [ ] Queue depth < 1000
- [ ] Replication lag < 1s

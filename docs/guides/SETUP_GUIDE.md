# 🔧 Руководство по настройке Language Exchange Bot

## 📋 Предварительные требования

- **Docker Desktop** для Windows/Mac или Docker Engine для Linux
- **Docker Compose** v2.0+
- **Git** для клонирования репозитория
- **Telegram Bot Token** от @BotFather

## 🚀 Пошаговая установка

### 1. Подготовка окружения

```bash
# Клонирование репозитория
git clone <repository-url>
cd language_exchange_bot

# Проверка Docker
docker --version
docker-compose --version
```

### 2. Создание Telegram бота

1. Откройте Telegram и найдите @BotFather
2. Создайте нового бота:

   ```shell
   /newbot
   Language Exchange Bot
   your_bot_username
   ```

3. Сохраните токен для использования в конфигурации

### 3. Настройка переменных окружения

```bash
# Копирование примера конфигурации
cp services/deploy/.env.example services/deploy/.env

# Редактирование конфигурации
nano services/deploy/.env
```

#### Основные настройки

```bash
# ⚠️ ОБЯЗАТЕЛЬНО ЗАМЕНИТЕ НА ВАШИ ЗНАЧЕНИЯ!
TELEGRAM_TOKEN=your_bot_token_here
DATABASE_URL=postgres://postgres:password@localhost:5432/language_exchange
REDIS_URL=redis://localhost:6379

# Администраторы (замените на реальные)
ADMIN_CHAT_IDS=123456789,987654321
ADMIN_USERNAMES=admin1,admin2

# Настройки сервера
DEBUG=false
PORT=8080
BOT_PORT=8080

# База данных
POSTGRES_DB=language_exchange
POSTGRES_USER=postgres
POSTGRES_PASSWORD=your_secure_password

# PgAdmin
PGADMIN_EMAIL=admin@local.host
PGADMIN_PASSWORD=admin_password
```

### 4. Запуск сервисов

#### Windows

```batch
cd services\deploy
start.bat
```

#### Linux/Mac

```bash
cd services/deploy
docker-compose up --build
```

### 5. Проверка запуска

```bash
# Проверка статуса контейнеров
docker-compose ps

# Проверка логов
docker-compose logs -f bot

# Проверка health checks
curl http://localhost:8080/health
```

## 🔧 Дополнительная настройка

### Настройка команд бота

В Telegram найдите @BotFather и настройте команды:

```shell
/setcommands
start - Начать работу с ботом
profile - Управление профилем
feedback - Оставить отзыв
help - Помощь
```

### Настройка описания бота

```shell
/setdescription
Бот для языкового обмена. Найдите партнеров для изучения языков!
```

### Получение Chat ID администратора

1. Напишите боту @userinfobot
2. Отправьте `/start`
3. Скопируйте ваш Chat ID
4. Добавьте в `ADMIN_CHAT_IDS`

## 🌐 Production настройка

### SSL/TLS сертификаты

```bash
# Установка Certbot
sudo apt install certbot

# Получение сертификата
sudo certbot certonly --standalone -d yourdomain.com

# Настройка автообновления
sudo crontab -e
# Добавить: 0 12 * * * /usr/bin/certbot renew --quiet
```

### Nginx конфигурация

```nginx
server {
    listen 443 ssl;
    server_name yourdomain.com;
    
    ssl_certificate /etc/letsencrypt/live/yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/yourdomain.com/privkey.pem;
    
    location /webhook {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
    
    location /health {
        proxy_pass http://localhost:8080;
    }
    
    location /metrics {
        proxy_pass http://localhost:8080;
        # Ограничить доступ
        allow 127.0.0.1;
        deny all;
    }
}
```

### Системный сервис

```ini
# /etc/systemd/system/language-exchange-bot.service
[Unit]
Description=Language Exchange Bot
After=network.target postgresql.service redis.service

[Service]
Type=simple
User=bot
WorkingDirectory=/opt/language-exchange-bot
ExecStart=/opt/language-exchange-bot/bot
Restart=always
RestartSec=5
Environment=DEBUG=false
Environment=PORT=8080

[Install]
WantedBy=multi-user.target
```

```bash
# Активация сервиса
sudo systemctl daemon-reload
sudo systemctl enable language-exchange-bot
sudo systemctl start language-exchange-bot
```

## 🔍 Мониторинг

### Настройка Prometheus

```yaml
# prometheus.yml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'language-exchange-bot'
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: '/metrics'
  
  - job_name: 'profile-service'
    static_configs:
      - targets: ['localhost:8081']
  
  - job_name: 'matcher-service'
    static_configs:
      - targets: ['localhost:8082']
```

### Настройка Grafana

1. Импортируйте дашборд для Go приложений
2. Настройте источник данных Prometheus
3. Создайте алерты для критических метрик

### Настройка логирования

```bash
# Ротация логов
sudo nano /etc/logrotate.d/language-exchange-bot

# Содержимое:
/var/log/language-exchange-bot/*.log {
    daily
    missingok
    rotate 30
    compress
    delaycompress
    notifempty
    create 644 bot bot
}
```

## 🚨 Устранение неполадок

### Проблемы с запуском

```bash
# Проверка портов
netstat -tlnp | grep :8080
netstat -tlnp | grep :5432
netstat -tlnp | grep :6379

# Проверка Docker
docker ps -a
docker logs <container_id>

# Очистка и перезапуск
docker-compose down -v
docker-compose up --build --force-recreate
```

### Проблемы с базой данных

```bash
# Проверка подключения к PostgreSQL
pg_isready -h localhost -p 5432

# Подключение к базе
psql -h localhost -U postgres -d language_exchange

# Проверка миграций
docker-compose exec postgres psql -U postgres -d language_exchange -c "\dt"
```

### Проблемы с Redis

```bash
# Проверка Redis
redis-cli ping

# Очистка кэша
redis-cli FLUSHALL

# Мониторинг Redis
redis-cli monitor
```

### Проблемы с ботом

```bash
# Проверка токена
curl "https://api.telegram.org/bot$TELEGRAM_TOKEN/getMe"

# Проверка webhook (если используется)
curl "https://api.telegram.org/bot$TELEGRAM_TOKEN/getWebhookInfo"

# Тестирование API
curl http://localhost:8080/health
```

## 📞 Поддержка

При возникновении проблем:

1. **Проверьте логи**: `docker-compose logs -f`
2. **Проверьте конфигурацию**: `cat .env`
3. **Запустите тесты**: `make test`
4. **Проверьте статус**: `docker-compose ps`

**Документация**: [README.md](../README.md)  
**Безопасность**: [SECURITY.md](../reports/SECURITY.md)  
**Производительность**: [PERFORMANCE.md](../reports/PERFORMANCE.md)

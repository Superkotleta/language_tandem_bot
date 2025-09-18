# 🔧 Ручная настройка Language Exchange Bot

## 📋 Что нужно настроить вручную

### 1. Переменные окружения (.env)

Создайте файл `.env` в папке `deploy/`:

```bash
# === ОСНОВНЫЕ НАСТРОЙКИ ===
# ⚠️ ВНИМАНИЕ: Замените на ваши реальные значения!
TELEGRAM_TOKEN=your_bot_token_here
DATABASE_URL=postgres://postgres:password@localhost:5432/language_exchange
REDIS_URL=redis://localhost:6379

# === НАСТРОЙКИ СЕРВЕРА ===
DEBUG=false
PORT=8080
BOT_PORT=8080

# === АДМИНИСТРАТОРЫ ===
# ⚠️ ВНИМАНИЕ: Замените на реальные ID и username администраторов!
# ID чатов администраторов (через запятую)
ADMIN_CHAT_IDS=123456789,987654321
# Username администраторов (через запятую, без @)
ADMIN_USERNAMES=admin1,admin2

# === БАЗА ДАННЫХ ===
POSTGRES_DB=language_exchange
POSTGRES_USER=postgres
POSTGRES_PASSWORD=your_secure_password

# === PgAdmin ===
PGADMIN_EMAIL=admin@local.host
PGADMIN_PASSWORD=admin_password

# === ЛОКАЛИЗАЦИЯ ===
LOCALES_DIR=./locales

# === WEBHOOK (опционально) ===
WEBHOOK_URL=https://yourdomain.com/webhook

# === ВКЛЮЧЕНИЕ СЕРВИСОВ ===
ENABLE_TELEGRAM=true
ENABLE_DISCORD=false
```

### 2. Настройка Telegram Bot

1. **Создайте бота через @BotFather:**

   ```shell
   /newbot
   Language Exchange Bot
   your_bot_username
   ```

2. **Получите токен** и добавьте в `.env`

3. **Настройте команды бота:**

   ```shell
   /setcommands
   start - Начать работу с ботом
   profile - Управление профилем
   feedback - Оставить отзыв
   help - Помощь
   ```

4. **Настройте описание:**

   ```shell
   /setdescription
   Бот для языкового обмена. Найдите партнеров для изучения языков!
   ```

### 3. Настройка базы данных

#### PostgreSQL

```sql
-- Создание базы данных
CREATE DATABASE language_exchange;

-- Создание пользователя (опционально)
CREATE USER bot_user WITH PASSWORD 'secure_password';
GRANT ALL PRIVILEGES ON DATABASE language_exchange TO bot_user;
```

#### Применение миграций

```bash
# Автоматически через Docker Compose
docker-compose up postgres

# Или вручную
psql -h localhost -U postgres -d language_exchange -f deploy/db-init/01-init-schemas.sql
psql -h localhost -U postgres -d language_exchange -f deploy/db-init/02-init-languages.sql
# ... остальные файлы
```

### 4. Настройка Redis

```bash
# Установка Redis (Ubuntu/Debian)
sudo apt update
sudo apt install redis-server

# Запуск Redis
sudo systemctl start redis-server
sudo systemctl enable redis-server

# Проверка
redis-cli ping
```

### 5. Настройка мониторинга

#### Prometheus (опционально)

```yaml
# prometheus.yml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'language-exchange-bot'
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: '/metrics'
```

#### Grafana (опционально)

1. Установите Grafana
2. Импортируйте дашборд для мониторинга
3. Настройте источник данных Prometheus

### 6. Настройка логирования

#### Структурированные логи

Логи автоматически сохраняются в JSON формате:

```json
{
  "timestamp": "2025-09-18T12:00:00Z",
  "level": "info",
  "service": "bot",
  "message": "Bot started successfully",
  "user_id": 12345,
  "action": "start_command"
}
```

#### Ротация логов

```bash
# Настройка logrotate
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

### 7. Настройка SSL/TLS (для production)

#### Let's Encrypt

```bash
# Установка certbot
sudo apt install certbot

# Получение сертификата
sudo certbot certonly --standalone -d yourdomain.com

# Автоматическое обновление
sudo crontab -e
# Добавить: 0 12 * * * /usr/bin/certbot renew --quiet
```

#### Nginx конфигурация

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
}
```

### 8. Настройка бэкапов

#### Автоматический бэкап БД

```bash
#!/bin/bash
# backup.sh
DATE=$(date +%Y%m%d_%H%M%S)
pg_dump language_exchange > /backups/db_$DATE.sql
find /backups -name "db_*.sql" -mtime +7 -delete
```

#### Cron задача

```bash
# Добавить в crontab
0 2 * * * /path/to/backup.sh
```

### 9. Настройка системы

#### Systemd сервис

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

#### Запуск сервиса

```bash
sudo systemctl daemon-reload
sudo systemctl enable language-exchange-bot
sudo systemctl start language-exchange-bot
```

### 10. Настройка файрвола

```bash
# UFW (Ubuntu)
sudo ufw allow 22/tcp    # SSH
sudo ufw allow 80/tcp    # HTTP
sudo ufw allow 443/tcp   # HTTPS
sudo ufw allow 8080/tcp  # Bot API
sudo ufw enable
```

## 🔍 Проверка настройки

### 1. Проверка сервисов

```bash
# Проверка статуса
sudo systemctl status language-exchange-bot
sudo systemctl status postgresql
sudo systemctl status redis-server

# Проверка портов
netstat -tlnp | grep :8080
netstat -tlnp | grep :5432
netstat -tlnp | grep :6379
```

### 2. Проверка логов

```bash
# Логи бота
journalctl -u language-exchange-bot -f

# Логи PostgreSQL
sudo tail -f /var/log/postgresql/postgresql-*.log

# Логи Redis
sudo tail -f /var/log/redis/redis-server.log
```

### 3. Проверка метрик

```bash
# Health check
curl http://localhost:8080/health

# Метрики
curl http://localhost:8080/metrics
```

### 4. Тестирование бота

1. Найдите бота в Telegram
2. Отправьте `/start`
3. Проверьте логи на наличие ошибок
4. Протестируйте основные функции

## 🚨 Устранение неполадок

### Частые проблемы

#### 1. Бот не отвечает

```bash
# Проверьте токен
echo $TELEGRAM_TOKEN

# Проверьте логи
make logs-bot

# Проверьте подключение к БД
psql $DATABASE_URL -c "SELECT 1;"
```

#### 2. Ошибки базы данных

```bash
# Проверьте подключение
pg_isready -h localhost -p 5432

# Проверьте права пользователя
psql -U postgres -c "\du"

# Примените миграции заново
make db-setup
```

#### 3. Проблемы с Redis

```bash
# Проверьте Redis
redis-cli ping

# Очистите кэш
redis-cli FLUSHALL
```

#### 4. Высокое использование памяти

```bash
# Мониторинг ресурсов
htop
free -h
df -h

# Очистка логов
sudo journalctl --vacuum-time=7d
```

## 📞 Поддержка

При возникновении проблем:

1. Проверьте логи: `make logs`
2. Проверьте конфигурацию: `cat .env`
3. Запустите тесты: `make test`
4. Проверьте статус сервисов: `make monitor`

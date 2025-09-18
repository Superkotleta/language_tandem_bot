# 🚀 Production Deployment Guide

## VPS с доменом (Вариант 3)

Этот гайд описывает развертывание Language Exchange Bot на VPS с собственным доменом и SSL сертификатом.

### 📋 Требования

- **VPS** (DigitalOcean, Vultr, AWS EC2, Hetzner)
- **Домен** (Namecheap, GoDaddy, Cloudflare)
- **Ubuntu 20.04+** или **Debian 11+**
- **Root доступ** к серверу

### 🛠️ Пошаговая настройка

#### Шаг 1: Подготовка сервера

```bash
# Обновление системы
sudo apt update && sudo apt upgrade -y

# Установка необходимых пакетов
sudo apt install -y curl wget git nginx certbot python3-certbot-nginx ufw

# Настройка firewall
sudo ufw allow ssh
sudo ufw allow 80
sudo ufw allow 443
sudo ufw --force enable
```

#### Шаг 2: Установка Docker

```bash
# Установка Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Установка Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# Добавление пользователя в группу docker
sudo usermod -aG docker $USER
newgrp docker
```

#### Шаг 3: Настройка домена

```bash
# Создание DNS записей (в панели управления доменом)
# A запись: yourdomain.com -> IP_СЕРВЕРА
# A запись: www.yourdomain.com -> IP_СЕРВЕРА
```

#### Шаг 4: Клонирование проекта

```bash
# Клонирование репозитория
git clone <your-repository-url> /opt/language-exchange-bot
cd /opt/language-exchange-bot/services/deploy

# Создание .env файла
cp env.optimized.example .env
nano .env
```

#### Шаг 5: Настройка .env файла

```bash
# Основные настройки
TELEGRAM_TOKEN=your_actual_bot_token
DEBUG=false
PORT=8080

# База данных
DATABASE_URL=postgres://postgres:your_secure_password@db:5432/language_exchange?sslmode=disable
POSTGRES_PASSWORD=your_secure_password

# Webhook
WEBHOOK_URL=https://yourdomain.com/webhook/telegram

# Администраторы
ADMIN_CHAT_IDS=123456789,987654321
ADMIN_USERNAMES=admin1,admin2

# Безопасность
SSL_ENABLED=true
```

#### Шаг 6: Настройка Nginx

```bash
# Создание конфигурации Nginx
sudo nano /etc/nginx/sites-available/yourdomain.com
```

**Содержимое файла:**

```nginx
server {
    listen 80;
    server_name yourdomain.com www.yourdomain.com;

    # Логи
    access_log /var/log/nginx/yourdomain.com.access.log;
    error_log /var/log/nginx/yourdomain.com.error.log;

    # Основное приложение
    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # Таймауты
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }

    # Webhook endpoint (специальная обработка)
    location /webhook/telegram {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # Увеличенные таймауты для webhook
        proxy_connect_timeout 30s;
        proxy_send_timeout 30s;
        proxy_read_timeout 30s;
    }

    # Статические файлы (если есть)
    location /static/ {
        alias /opt/language-exchange-bot/static/;
        expires 1y;
        add_header Cache-Control "public, immutable";
    }
}
```

```bash
# Активация сайта
sudo ln -s /etc/nginx/sites-available/yourdomain.com /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

#### Шаг 7: Получение SSL сертификата

```bash
# Получение SSL сертификата от Let's Encrypt
sudo certbot --nginx -d yourdomain.com -d www.yourdomain.com

# Автоматическое обновление сертификата
sudo crontab -e
# Добавить строку:
# 0 12 * * * /usr/bin/certbot renew --quiet
```

#### Шаг 8: Запуск приложения

```bash
# Запуск Docker сервисов
make -f Makefile.optimized up

# Проверка статуса
docker-compose -f docker-compose.optimized.yml ps
```

#### Шаг 9: Настройка webhook

```bash
# Настройка webhook в Telegram
curl -X POST "https://api.telegram.org/botYOUR_BOT_TOKEN/setWebhook" \
     -H "Content-Type: application/json" \
     -d '{"url": "https://yourdomain.com/webhook/telegram"}'

# Проверка webhook
curl "https://api.telegram.org/botYOUR_BOT_TOKEN/getWebhookInfo"
```

#### Шаг 10: Настройка мониторинга

```bash
# Создание systemd сервиса для автозапуска
sudo nano /etc/systemd/system/language-exchange-bot.service
```

**Содержимое файла:**

```ini
[Unit]
Description=Language Exchange Bot
Requires=docker.service
After=docker.service

[Service]
Type=oneshot
RemainAfterExit=yes
WorkingDirectory=/opt/language-exchange-bot/services/deploy
ExecStart=/usr/local/bin/docker-compose -f docker-compose.optimized.yml up -d
ExecStop=/usr/local/bin/docker-compose -f docker-compose.optimized.yml down
TimeoutStartSec=0

[Install]
WantedBy=multi-user.target
```

```bash
# Активация сервиса
sudo systemctl enable language-exchange-bot.service
sudo systemctl start language-exchange-bot.service
```

### 🔧 Полезные команды

```bash
# Проверка статуса
sudo systemctl status language-exchange-bot.service

# Просмотр логов
make -f Makefile.optimized logs

# Перезапуск
sudo systemctl restart language-exchange-bot.service

# Обновление приложения
git pull
make -f Makefile.optimized restart

# Бэкап базы данных
docker-compose -f docker-compose.optimized.yml exec db pg_dump -U postgres language_exchange > backup.sql

# Восстановление базы данных
docker-compose -f docker-compose.optimized.yml exec -T db psql -U postgres language_exchange < backup.sql
```

### 📊 Мониторинг

- **Grafana**: <https://yourdomain.com:3000> (admin/admin)
- **Prometheus**: <https://yourdomain.com:9090>
- **Health Check**: <https://yourdomain.com/health>
- **Metrics**: <https://yourdomain.com/metrics>

### 🔒 Безопасность

```bash
# Настройка fail2ban
sudo apt install fail2ban
sudo systemctl enable fail2ban
sudo systemctl start fail2ban

# Настройка автоматических обновлений
sudo apt install unattended-upgrades
sudo dpkg-reconfigure unattended-upgrades

# Регулярные бэкапы
sudo crontab -e
# Добавить:
# 0 2 * * * /opt/language-exchange-bot/scripts/backup.sh
```

### 🚨 Troubleshooting

#### Проблема: Webhook не работает

```bash
# Проверка SSL сертификата
curl -I https://yourdomain.com/webhook/telegram

# Проверка логов Nginx
sudo tail -f /var/log/nginx/yourdomain.com.error.log

# Проверка логов приложения
make -f Makefile.optimized logs
```

#### Проблема: Высокая нагрузка

```bash
# Мониторинг ресурсов
htop
docker stats

# Оптимизация Nginx
sudo nano /etc/nginx/nginx.conf
# Увеличить worker_processes и worker_connections
```

#### Проблема: База данных

```bash
# Проверка подключения
docker-compose -f docker-compose.optimized.yml exec db psql -U postgres -c "SELECT 1;"

# Восстановление из бэкапа
docker-compose -f docker-compose.optimized.yml exec -T db psql -U postgres language_exchange < backup.sql
```

### 📈 Масштабирование

Для высоких нагрузок рассмотрите:

1. **Load Balancer** (HAProxy, Nginx)
2. **Database Clustering** (PostgreSQL streaming replication)
3. **Redis Cluster** для кэширования
4. **CDN** для статических файлов
5. **Kubernetes** для оркестрации

### 💰 Примерные затраты

- **VPS**: $5-20/месяц (DigitalOcean, Vultr)
- **Домен**: $10-15/год
- **SSL**: Бесплатно (Let's Encrypt)
- **Мониторинг**: Бесплатно (Grafana, Prometheus)

**Итого**: ~$10-25/месяц для production развертывания

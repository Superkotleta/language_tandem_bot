# Настройка переменных окружения

## Создание файла .env

Создайте файл `.env` в папке `services/deploy/` со следующим содержимым:

```env
# ===========================================
# Language Exchange Bot - Configuration
# ===========================================

# ===========================================
# Database Configuration
# ===========================================
POSTGRES_DB=language_exchange
POSTGRES_USER=postgres
POSTGRES_PASSWORD=password123
DATABASE_URL=postgres://postgres:password123@postgres:5432/language_exchange?sslmode=disable

# Database Schema Configuration
DB_SCHEMA=public

# Database Passwords for Services (для будущих микросервисов)
PROFILE_DB_PASS=profile_pwd
MATCHING_DB_PASS=matching_pwd
MATCHING_RO_DB_PASS=matching_ro_pwd

# ===========================================
# Telegram Bot Configuration
# ===========================================
# Получите токен у @BotFather в Telegram
TELEGRAM_TOKEN=your_telegram_bot_token_here

# Alternative: Load token from file
# TELEGRAM_TOKEN_FILE=/path/to/token/file

# ===========================================
# Admin Configuration
# ===========================================
# Укажите Chat ID или username администраторов через запятую
# Примеры:
# ADMIN_CHAT_IDS=123456789,@admin_username,987654321
# ADMIN_CHAT_IDS=@your_username
# ADMIN_CHAT_IDS=123456789
# 
# Как узнать свой username:
# 1. Откройте Telegram → Settings → Username
# 2. Скопируйте username (например: myusername)
# 3. Укажите: ADMIN_CHAT_IDS=@myusername
#
# Как узнать Chat ID:
# 1. Напишите боту @userinfobot
# 2. Скопируйте Chat ID (число)
# 3. Укажите: ADMIN_CHAT_IDS=123456789
ADMIN_CHAT_IDS=

# ===========================================
# Redis Configuration
# ===========================================
REDIS_URL=redis:6379
REDIS_PASSWORD=
REDIS_DB=0

# ===========================================
# Bot Settings
# ===========================================
DEBUG=false
ENABLE_TELEGRAM=true
ENABLE_DISCORD=false

# ===========================================
# Server Configuration
# ===========================================
PORT=8080
WEBHOOK_URL=

# Alternative: Load database URL from file
# DATABASE_URL_FILE=/path/to/db/url/file

# ===========================================
# Localization Configuration
# ===========================================
# Path to locales directory (optional)
LOCALES_DIR=./locales

# ===========================================
# PgAdmin Configuration
# ===========================================
PGADMIN_EMAIL=admin@local.host
PGADMIN_PASSWORD=password

# ===========================================
# Service Ports (для будущих микросервисов)
# ===========================================
# Profile Service (временно отключен)
HTTP_PORT=8081

# Matcher Service (временно отключен)
MATCHER_HTTP_PORT=8082

# ===========================================
# Migration Configuration
# ===========================================
# Migration directories for services (для будущих микросервисов)
MIGRATIONS_DIR=/migrations
```

## Важные настройки

### 1. Telegram Token

- Получите токен у [@BotFather](https://t.me/BotFather) в Telegram
- Замените `your_telegram_bot_token_here` на ваш реальный токен

### 2. Администраторы

- **Обязательно** настройте `ADMIN_CHAT_IDS` для доступа к команде `/feedbacks`
- Можно использовать username (с @) или Chat ID (число)
- Несколько администраторов указывайте через запятую

### 3. База данных

- Настройки базы данных уже готовы для Docker
- При необходимости измените пароли

### 4. Redis кэширование

- **REDIS_URL**: Адрес Redis сервера (по умолчанию: `redis:6379`)
- **REDIS_PASSWORD**: Пароль Redis (опционально)
- **REDIS_DB**: Номер базы данных Redis (по умолчанию: 0)
- Система автоматически переключается на in-memory кэш при недоступности Redis

## Проверка настройки

После создания файла `.env`:

1. Перезапустите бота
2. В логах должно появиться: `Загружен .env файл из: ../../deploy/.env`
3. Проверьте команду `/feedbacks` - она должна работать для администраторов

## Безопасность

- **НЕ** коммитьте файл `.env` в git
- Файл `.env` уже добавлен в `.gitignore`
- Используйте разные настройки для dev/prod окружений

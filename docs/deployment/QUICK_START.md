# 🚀 Quick Start Guide

## Вариант 1: Разработка с ngrok (Рекомендуется)

### 🎯 Быстрый старт за 5 минут

```bash
# 1. Клонируйте репозиторий
git clone <your-repo> language_exchange_bot
cd language_exchange_bot/services/deploy

# 2. Автоматическая настройка
make -f Makefile.optimized dev-setup
```

**Что происходит:**

- ✅ Установка ngrok
- ✅ Создание .env файла
- ✅ Настройка webhook URL
- ✅ Запуск Docker сервисов
- ✅ Настройка webhook в Telegram

### 📱 Настройка бота

1. **Создайте бота в Telegram:**
   - Найдите @BotFather
   - Отправьте `/newbot`
   - Следуйте инструкциям
   - Скопируйте токен

2. **Добавьте токен в .env:**

   ```bash
   nano .env
   # Измените: TELEGRAM_TOKEN=your_actual_token
   ```

3. **Перезапустите:**

   ```bash
   make -f Makefile.optimized restart
   ```

### 🧪 Тестирование

```bash
# Отправьте /start боту в Telegram
# Проверьте логи
make -f Makefile.optimized logs

# Мониторинг
make -f Makefile.optimized monitor
```

### 🛑 Остановка

```bash
# Остановить сервисы
make -f Makefile.optimized down

# Остановить ngrok
make -f Makefile.optimized ngrok-stop
```

---

## Вариант 2: Production на VPS

### 📋 Требования

- VPS с Ubuntu 20.04+
- Домен с SSL сертификатом
- Docker и Docker Compose

### 🚀 Развертывание

```bash
# 1. Подготовка сервера
sudo apt update && sudo apt upgrade -y
sudo apt install -y docker.io docker-compose nginx certbot

# 2. Клонирование проекта
git clone <your-repo> /opt/language-exchange-bot
cd /opt/language-exchange-bot/services/deploy

# 3. Настройка
cp env.production.example .env
nano .env  # Настройте переменные

# 4. Запуск
make -f Makefile.optimized prod
```

### 📖 Подробная инструкция

Следуйте [PRODUCTION_DEPLOYMENT.md](PRODUCTION_DEPLOYMENT.md) для полной настройки.

---

## 🔧 Полезные команды

### Основные

```bash
make -f Makefile.optimized help          # Справка
make -f Makefile.optimized up            # Запуск
make -f Makefile.optimized down          # Остановка
make -f Makefile.optimized logs          # Логи
make -f Makefile.optimized monitor       # Мониторинг
```

### Разработка

```bash
make -f Makefile.optimized dev-setup     # Полная настройка
make -f Makefile.optimized ngrok-setup   # Только ngrok
make -f Makefile.optimized webhook-setup # Только webhook
```

### Production

```bash
make -f Makefile.optimized prod          # Production режим
make -f Makefile.optimized production-setup # Инструкции
```

### База данных

```bash
make -f Makefile.optimized db-backup     # Бэкап
make -f Makefile.optimized db-restore    # Восстановление
```

---

## 📊 Мониторинг

После запуска доступны:

- **Bot API**: <http://localhost:8080>
- **Health Check**: <http://localhost:8080/health>
- **Metrics**: <http://localhost:8080/metrics>
- **Grafana**: <http://localhost:3000> (admin/admin)
- **Prometheus**: <http://localhost:9090>
- **PgAdmin**: <http://localhost:5050> (<admin@admin.com>/admin)
- **ngrok UI**: <http://localhost:4040> (только для разработки)

---

## 🆘 Troubleshooting

### Проблема: ngrok не запускается

```bash
# Проверьте авторизацию
ngrok config check

# Авторизуйтесь
ngrok config add-authtoken YOUR_TOKEN
```

### Проблема: Webhook не работает

```bash
# Проверьте URL
curl -I https://your-ngrok-url.ngrok.io/webhook/telegram

# Проверьте настройку
curl "https://api.telegram.org/botYOUR_TOKEN/getWebhookInfo"
```

### Проблема: Бот не отвечает

```bash
# Проверьте логи
make -f Makefile.optimized logs

# Проверьте здоровье
make -f Makefile.optimized health
```

---

## 📚 Дополнительные ресурсы

- [Полная документация](../README.md)
- [Production развертывание](PRODUCTION_DEPLOYMENT.md)
- [Архитектура проекта](../guides/ARCHITECTURE.md)
- [API документация](../api/README.md)

---

## 🎉 Готово

Теперь у вас есть полнофункциональный Language Exchange Bot с:

- ✅ Telegram интеграцией
- ✅ PostgreSQL базой данных
- ✅ Redis кэшированием
- ✅ Prometheus метриками
- ✅ Grafana мониторингом
- ✅ Автоматическими бэкапами
- ✅ Health checks
- ✅ Graceful shutdown

**Удачной разработки!** 🚀

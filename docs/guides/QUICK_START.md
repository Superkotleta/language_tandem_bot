# 🚀 Quick Start - Language Exchange Bot

## 🎯 Быстрый старт за 5 минут

### Вариант 1: Оптимизированная версия (Рекомендуется)

```bash
# Автоматическая настройка для разработки
make optimized-dev
```

**Что происходит:**

- ✅ Установка ngrok
- ✅ Создание .env файла
- ✅ Настройка webhook URL
- ✅ Запуск Docker сервисов
- ✅ Настройка webhook в Telegram

### Вариант 2: Классическая версия

```bash
# Запуск классической версии
make up
```

---

## 📱 Настройка бота

1. **Создайте бота в Telegram:**
   - Найдите @BotFather
   - Отправьте `/newbot`
   - Следуйте инструкциям
   - Скопируйте токен

2. **Добавьте токен в .env:**

   ```bash
   # Для оптимизированной версии
   cd deploy
   nano .env
   # Измените: TELEGRAM_TOKEN=your_actual_token
   
   # Для классической версии
   nano .env
   # Добавьте: TELEGRAM_TOKEN=your_actual_token
   ```

3. **Перезапустите:**

   ```bash
   # Оптимизированная версия
   make optimized-restart
   
   # Классическая версия
   make restart
   ```

---

## 🧪 Тестирование

```bash
# Отправьте /start боту в Telegram
# Проверьте логи
make optimized-logs    # Оптимизированная версия
make logs             # Классическая версия

# Мониторинг
make optimized-monitor # Оптимизированная версия
make check-health     # Классическая версия
```

---

## 🛑 Остановка

```bash
# Оптимизированная версия
make optimized-down
make optimized-stop-ngrok

# Классическая версия
make down
```

---

## 📊 Мониторинг

### Оптимизированная версия мониторинга

- **Bot API**: <http://localhost:8080>
- **Health Check**: <http://localhost:8080/health>
- **Metrics**: <http://localhost:8080/metrics>
- **Grafana**: <http://localhost:3000> (admin/admin)
- **Prometheus**: <http://localhost:9090>
- **PgAdmin**: <http://localhost:5050> (<admin@admin.com>/admin)
- **ngrok UI**: <http://localhost:4040>

### Классическая версия мониторинга

- **Bot API**: <http://localhost:8080>
- **Health Check**: <http://localhost:8080/health>
- **Metrics**: <http://localhost:8080/metrics>

---

## 🔧 Полезные команды

### Оптимизированная версия

```bash
make optimized-dev           # 🚀 Запустить для разработки
make optimized-prod          # 🏭 Запустить для production
make optimized-setup         # ⚙️ Полная настройка
make optimized-ngrok         # 🌐 Настроить ngrok
make optimized-webhook       # 🔗 Настроить webhook
make optimized-monitor       # 📊 Мониторинг
make optimized-logs          # 📝 Логи
make optimized-health        # 🏥 Health check
make optimized-backup        # 💾 Создать бэкап
make optimized-help          # ❓ Справка
```

### Классическая версия

```bash
make up              # Запуск
make down            # Остановка
make restart         # Перезапуск
make logs            # Логи
make check-health    # Health check
make psql-root       # Подключение к БД
make help            # Справка
```

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
make optimized-logs

# Проверьте здоровье
make optimized-health
```

---

## 📚 Дополнительные ресурсы

- [Полная документация](../README.md)
- [Production развертывание](../deployment/PRODUCTION_DEPLOYMENT.md)
- [Архитектура проекта](ARCHITECTURE.md)

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

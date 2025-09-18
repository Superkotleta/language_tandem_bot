# 📋 Makefile Guide - Language Exchange Bot

## 🎯 Основные команды

### 🚀 Быстрый старт

```bash
# Автоматическая настройка для разработки
make optimized-dev

# Показать все доступные команды
make help
```

---

## 📊 Полный список команд

### 🔧 Основные команды (v1.0)

| Команда | Описание |
|---------|----------|
| `make up` | Запустить контейнеры |
| `make down` | Остановить контейнеры и удалить тома |
| `make restart` | Перезапустить контейнеры |
| `make rebuild` | Пересобрать образы без кэша |
| `make clean-db` | Очистить тома БД и перезапустить |
| `make full-restart` | Полный перезапуск с очисткой БД |
| `make logs` | Показать логи всех контейнеров |
| `make logs-bot` | Показать только логи бота |
| `make logs-db` | Показать только логи БД |
| `make check-health` | Проверить статус контейнеров |

### 🚀 Оптимизированная версия (v2.0)

| Команда | Описание |
|---------|----------|
| `make optimized-dev` | 🚀 Запустить для разработки (с ngrok) |
| `make optimized-prod` | 🏭 Запустить для production |
| `make optimized-setup` | ⚙️ Полная настройка для разработки |
| `make optimized-ngrok` | 🌐 Настроить ngrok |
| `make optimized-webhook` | 🔗 Настроить webhook |
| `make optimized-stop-ngrok` | 🛑 Остановить ngrok |
| `make optimized-monitor` | 📊 Мониторинг |
| `make optimized-logs` | 📝 Логи |
| `make optimized-health` | 🏥 Health check |
| `make optimized-down` | ⏹️ Остановить |
| `make optimized-restart` | 🔄 Перезапустить |
| `make optimized-backup` | 💾 Создать бэкап |
| `make optimized-production-guide` | 📖 Инструкции для production |
| `make optimized-help` | ❓ Справка по оптимизированной версии |

### 🪟 Windows команды

| Команда | Описание |
|---------|----------|
| `make win-logs-bot` | Показать логи бота (PowerShell) |
| `make win-logs-db` | Показать логи БД (PowerShell) |
| `make win-check-emojis` | Проверить загрузку эмодзи |
| `make win-clean-and-restart` | Очистить и перезапустить (PowerShell) |
| `make win-clean-all` | Полная очистка (PowerShell) |
| `make win-diagnose` | Диагностика сети |

---

## 🎯 Сценарии использования

### 🚀 Разработка

```bash
# Полная настройка для разработки
make optimized-dev

# Проверка работы
make optimized-monitor
make optimized-logs

# Остановка
make optimized-down
make optimized-stop-ngrok
```

### 🏭 Production

```bash
# Показать инструкции
make optimized-production-guide

# Запуск в production режиме
make optimized-prod

# Мониторинг
make optimized-monitor
make optimized-health
```

### 🔧 Отладка

```bash
# Проверка здоровья
make optimized-health

# Логи
make optimized-logs

# Бэкап
make optimized-backup
```

### 🪟 Windows разработка

```bash
# Логи через PowerShell
make win-logs-bot
make win-logs-db

# Диагностика
make win-diagnose

# Очистка и перезапуск
make win-clean-and-restart
```

---

## 📊 Мониторинг

### Оптимизированная версия

- **Bot API**: <http://localhost:8080>
- **Health Check**: <http://localhost:8080/health>
- **Metrics**: <http://localhost:8080/metrics>
- **Grafana**: <http://localhost:3000> (admin/admin)
- **Prometheus**: <http://localhost:9090>
- **PgAdmin**: <http://localhost:5050> (<admin@admin.com>/admin)
- **ngrok UI**: <http://localhost:4040>

### Классическая версия

- **Bot API**: <http://localhost:8080>
- **Health Check**: <http://localhost:8080/health>
- **Metrics**: <http://localhost:8080/metrics>

---

## 🆘 Troubleshooting

### Проблема: Команда не найдена

```bash
# Проверьте, что вы в правильной директории
pwd
# Должно быть: .../language_exchange_bot/services

# Проверьте Makefile
ls -la Makefile
```

### Проблема: Docker не запускается

```bash
# Проверьте Docker
docker --version
docker-compose --version

# Проверьте статус
make check-health
```

### Проблема: ngrok не работает

```bash
# Проверьте авторизацию
ngrok config check

# Авторизуйтесь
ngrok config add-authtoken YOUR_TOKEN
```

### Проблема: Webhook не работает

```bash
# Проверьте настройку
make optimized-webhook

# Проверьте URL
curl -I https://your-ngrok-url.ngrok.io/webhook/telegram
```

---

## 📚 Дополнительные ресурсы

- [Quick Start Guide](QUICK_START.md) - Быстрый старт
- [Production Deployment](../deployment/PRODUCTION_DEPLOYMENT.md) - Production развертывание
- [Архитектура проекта](ARCHITECTURE.md)

---

## 🎉 Готово

Теперь у вас есть полный контроль над Language Exchange Bot через Makefile! 🚀

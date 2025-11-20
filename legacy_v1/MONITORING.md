# 📊 Мониторинг и производительность Language Exchange Bot

## 🚀 Быстрый старт

### 1. Запуск дашборда мониторинга

```go
// В main.go добавьте:
import "language-exchange-bot/internal/monitoring"

func main() {
    // ... инициализация бота ...
    
    // Запуск мониторинга
    monitoringService := monitoring.NewMonitoringService()
    go monitoringService.Start(context.Background(), 8080)
    
    // ... остальной код ...
}
```

### 2. Доступ к дашборду

- **🌐 Главная страница**: <http://localhost:8080>
- **📊 Метрики**: <http://localhost:8080/metrics>  
- **🚨 Ошибки**: <http://localhost:8080/errors>
- **⚠️ Алерты**: <http://localhost:8080/alerts>

## 📊 Компоненты мониторинга

### 🗄️ Расширенное кеширование

```go
// Кеш категорий интересов
categories, found := cache.GetInterestCategories(ctx, "ru")
if !found {
    categories = loadFromDB()
    cache.SetInterestCategories(ctx, "ru", categories)
}

// Кеш статистики пользователей
stats, found := cache.GetUserStats(ctx, userID)
if !found {
    stats = calculateStats(userID)
    cache.SetUserStats(ctx, userID, stats)
}

// Кеш конфигурации
config, found := cache.GetConfig(ctx, "max_interests")
if !found {
    config = loadConfig("max_interests")
    cache.SetConfig(ctx, "max_interests", config)
}
```

### ⚡ Батчевые операции

```go
// Массовое обновление интересов
err := batchLoader.BatchUpdateUserInterests(ctx, userID, interests, primaryInterests)

// Батчевая загрузка статистики
userStats, err := batchLoader.BatchLoadUserStatistics(ctx, userIDs)

// Популярные интересы
popular, err := batchLoader.BatchLoadPopularInterests(ctx, 10)
```

### 📝 Трейсинг запросов

```go
// Начало трейса
trace := monitoring.RecordOperation(requestID, userID, chatID, "edit_interests", "handler")

// Запись операций
monitoring.RecordDatabaseOperation(requestID, "update_user", duration, err)
monitoring.RecordCacheOperation(requestID, "get_user", hit, duration, err)
monitoring.RecordTelegramOperation(requestID, "send_message", duration, err)

// Завершение трейса
monitoring.EndOperation(requestID, "success", nil)
```

### 🚨 Обработка ошибок

```go
// Обработка обычных ошибок
monitoring.HandleError(ctx, err, requestID, userID, chatID, "operation")

// Обработка кастомных ошибок
customErr := errors.NewDatabaseError("connection failed", "DB error", ctx)
monitoring.HandleCustomError(ctx, customErr, requestID, userID, chatID, "operation")
```

## 📈 Метрики производительности

### 🎯 Основные метрики

- **⚡ Время ответа** - среднее время выполнения операций
- **📊 Кеш-хиты** - процент успешных обращений к кешу
- **🗄️ Обращения к БД** - количество запросов к базе данных
- **🚨 Ошибки** - частота и типы ошибок
- **👥 Активные пользователи** - количество одновременных пользователей

### 📊 Дашборд в реальном времени

- **🔄 Автообновление** - каждые 5 секунд
- **📈 Графики** - тренды производительности
- **🚨 Алерты** - критические проблемы
- **📝 Логи** - детальная информация об ошибках

## 🚨 Система алертов

### 📊 Уровни критичности

- **ℹ️ INFO** - информационные сообщения
- **⚠️ WARNING** - предупреждения
- **🚨 CRITICAL** - критические ошибки
- **🆘 EMERGENCY** - аварийные ситуации

### 🔔 Автоматические уведомления

```go
// Регистрация уведомителя
notifier := &TelegramNotifier{bot: bot, adminChatID: adminChatID}
errorHandler.RegisterNotifier(notifier)

// Алерты отправляются автоматически при:
// - Ошибках подключения к БД
// - Превышении лимитов Telegram API
// - Критических ошибках системы
```

## 🔧 Конфигурация

### ⚙️ Настройки кеширования

```go
config := &cache.Config{
    LanguagesTTL:    time.Hour,     // 1 час
    InterestsTTL:    time.Hour,     // 1 час  
    UsersTTL:        15 * time.Minute, // 15 минут
    StatsTTL:        5 * time.Minute,  // 5 минут
    TranslationsTTL: 30 * time.Minute, // 30 минут
}
```

### 📊 Настройки мониторинга

```go
// Порт дашборда
dashboardPort := 8080

// Интервалы обновления
metricsUpdateInterval := 5 * time.Second
alertsCheckInterval := 10 * time.Second
```

## 🛠️ Разработка

### 📝 Добавление новых метрик

```go
// Создание метрики
metricCollector.IncrementCounter("user_registrations", map[string]string{
    "source": "telegram",
    "language": "ru",
})

// Запись времени выполнения
metricCollector.RecordTimer("database_query", duration, map[string]string{
    "operation": "get_user",
    "table": "users",
})
```

### 🚨 Добавление новых типов ошибок

```go
// В errors/types.go
var ErrNewError = NewCustomError(
    ErrorTypeValidation, 
    "new error message", 
    "User-friendly message", 
    "",
)
```

### 📊 Расширение дашборда

```go
// Добавление нового endpoint
mux.HandleFunc("/api/custom", d.handleCustom)

func (d *Dashboard) handleCustom(w http.ResponseWriter, r *http.Request) {
    // Ваша логика
}
```

## 🔍 Отладка

### 📝 Логи трейсинга

```bash
# Просмотр трейсов
grep "TRACE_START" logs/bot.log
grep "TRACE_END" logs/bot.log
```

### 📊 Метрики производительности

```bash
# Экспорт метрик
curl http://localhost:8080/api/metrics > metrics.json

# Проверка здоровья
curl http://localhost:8080/api/health
```

### 🚨 Мониторинг алертов

```bash
# Активные алерты
curl http://localhost:8080/api/alerts

# Разрешение алерта
curl -X POST http://localhost:8080/api/alerts/{alert_id}/resolve
```

## 📚 Дополнительные ресурсы

- **📖 Документация Go**: <https://golang.org/doc/>
- **🗄️ Redis**: <https://redis.io/documentation>
- **📊 Prometheus**: <https://prometheus.io/docs/>
- **🚨 Grafana**: <https://grafana.com/docs/>

---

**🎉 Мониторинг готов к использованию! Ваш бот теперь работает на enterprise-уровне!**

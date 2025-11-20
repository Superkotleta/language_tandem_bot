# Circuit Breaker

Реализация паттерна Circuit Breaker для защиты от каскадных сбоев в системе.

## Описание

Circuit Breaker предотвращает каскадные сбои, ограничивая количество запросов к нестабильным или недоступным сервисам. Он автоматически переключается между тремя состояниями:

- **CLOSED** - нормальная работа, запросы проходят
- **OPEN** - сервис недоступен, запросы блокируются
- **HALF_OPEN** - ограниченное количество запросов для проверки восстановления

## Использование

### Базовое использование

```go
// Создание Circuit Breaker
cb := circuit_breaker.NewCircuitBreaker(circuit_breaker.Config{
    Name:        "my-service",
    MaxRequests: 3,
    Interval:    60 * time.Second,
    Timeout:     30 * time.Second,
    ReadyToTrip: func(counts circuit_breaker.Counts) bool {
        return counts.ConsecutiveFailures > 5
    },
})

// Выполнение операции
result, err := cb.Execute(func() (interface{}, error) {
    return myService.DoSomething()
})
```

### С контекстом

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

result, err := cb.ExecuteWithContext(ctx, func() (interface{}, error) {
    return myService.DoSomething()
})
```

### Готовые конфигурации

```go
// Для Telegram API
telegramCB := circuit_breaker.NewCircuitBreaker(circuit_breaker.TelegramConfig())

// Для базы данных
databaseCB := circuit_breaker.NewCircuitBreaker(circuit_breaker.DatabaseConfig())

// Для Redis
redisCB := circuit_breaker.NewCircuitBreaker(circuit_breaker.RedisConfig())
```

## Конфигурация

| Параметр | Описание | По умолчанию |
|----------|----------|--------------|
| `Name` | Имя для логирования | "default" |
| `MaxRequests` | Максимум запросов в HALF_OPEN | 3 |
| `Interval` | Интервал сброса счетчиков | 60s |
| `Timeout` | Время в OPEN перед HALF_OPEN | 60s |
| `ReadyToTrip` | Функция перехода в OPEN | >5 ошибок |
| `OnStateChange` | Callback при смене состояния | nil |

## Состояния

### CLOSED (Закрыто)

- Нормальная работа
- Все запросы проходят
- Счетчики обновляются

### OPEN (Открыто)

- Сервис недоступен
- Запросы блокируются
- Автоматический переход в HALF_OPEN через Timeout

### HALF_OPEN (Полуоткрыто)

- Ограниченное количество запросов (MaxRequests)
- При успехе → CLOSED
- При ошибке → OPEN

## Мониторинг

```go
// Получение состояния
state := cb.State()

// Получение счетчиков
counts := cb.Counts()
fmt.Printf("Requests: %d, Successes: %d, Failures: %d\n", 
    counts.Requests, counts.TotalSuccesses, counts.TotalFailures)
```

## Интеграция в BotService

```go
// Выполнение с защитой Telegram
result, err := service.ExecuteWithTelegramCircuitBreaker(func() (interface{}, error) {
    return telegramAPI.SendMessage(chatID, message)
})

// Выполнение с защитой базы данных
result, err := service.ExecuteWithDatabaseCircuitBreaker(func() (interface{}, error) {
    return db.Query("SELECT * FROM users")
})

// Получение состояний всех Circuit Breakers
states := service.GetCircuitBreakerStates()
fmt.Printf("Telegram: %s, Database: %s, Redis: %s\n", 
    states["telegram"], states["database"], states["redis"])
```

## Тестирование

```bash
# Запуск тестов
go test ./internal/circuit_breaker/

# С покрытием
go test -cover ./internal/circuit_breaker/
```

## Примеры

### Обработка ошибок Telegram API

```go
result, err := service.ExecuteWithTelegramCircuitBreaker(func() (interface{}, error) {
    return bot.SendMessage(chatID, message)
})

if err != nil {
    if strings.Contains(err.Error(), "circuit breaker is OPEN") {
        // Circuit Breaker заблокировал запрос
        log.Println("Telegram API недоступен, запрос заблокирован")
    } else {
        // Обычная ошибка API
        log.Printf("Ошибка Telegram API: %v", err)
    }
}
```

### Мониторинг состояния

```go
// Проверка состояния всех сервисов
states := service.GetCircuitBreakerStates()
for service, state := range states {
    if state == "OPEN" {
        log.Printf("⚠️ Сервис %s недоступен", service)
    } else if state == "HALF_OPEN" {
        log.Printf("🔄 Сервис %s восстанавливается", service)
    } else {
        log.Printf("✅ Сервис %s работает нормально", service)
    }
}
```

## Лучшие практики

1. **Настройка порогов** - адаптируйте `ReadyToTrip` под специфику сервиса
2. **Мониторинг** - отслеживайте состояния и счетчики
3. **Логирование** - используйте `OnStateChange` для уведомлений
4. **Fallback** - предусмотрите альтернативные пути при блокировке
5. **Тестирование** - проверяйте поведение при сбоях

## Производительность

- **Накладные расходы**: < 1мкс на запрос
- **Память**: ~100 байт на экземпляр
- **Thread-safe**: Да, использует мьютексы
- **Горутины**: Нет, синхронная работа

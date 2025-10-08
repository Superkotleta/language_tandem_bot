# API Documentation

Этот раздел содержит документацию для REST API Language Exchange Bot.

## Структура документации

- **[examples.md](./examples.md)** - Примеры использования API с curl командами
- **[postman_collection.json](./postman_collection.json)** - Postman коллекция для тестирования API
- **[swagger.yaml](./swagger.yaml)** - OpenAPI спецификация (генерируется автоматически)

## Быстрый старт

### 1. Установка Postman

1. Скачайте и установите [Postman](https://www.postman.com/downloads/)
2. Импортируйте коллекцию `postman_collection.json`
3. Установите переменные окружения:
   - `base_url`: `http://localhost:8080/api/v1`
   - `api_token`: `YOUR_API_TOKEN_HERE`

### 2. Настройка аутентификации

Замените `YOUR_API_TOKEN_HERE` на реальный API токен в переменных Postman.

### 3. Тестирование API

1. Запустите сервер: `go run cmd/bot/main.go`
2. Откройте Postman
3. Выберите коллекцию "Language Exchange Bot API"
4. Начните с "Health Check" для проверки доступности

## Endpoints

### Административные

| Method | Endpoint | Описание |
|--------|----------|----------|
| GET | `/stats` | Получение статистики системы |
| GET | `/users` | Список пользователей |
| GET | `/users/{id}` | Информация о пользователе |
| PUT | `/users/{id}` | Обновление пользователя |
| GET | `/languages` | Список языков |
| GET | `/interests` | Список интересов |
| GET | `/feedback` | Список отзывов |
| PUT | `/feedback/{id}/process` | Обработка отзыва |

### Кэш управление

| Method | Endpoint | Описание |
|--------|----------|----------|
| GET | `/cache/stats` | Статистика кэша |
| POST | `/cache/clear` | Очистка кэша |

### Webhooks

| Method | Endpoint | Описание |
|--------|----------|----------|
| POST | `/webhook/telegram` | Telegram webhook |

### Системные

| Method | Endpoint | Описание |
|--------|----------|----------|
| GET | `/health` | Health check |
| GET | `/metrics` | Prometheus метрики |
| GET | `/swagger/` | Swagger UI |

## Аутентификация

API использует Bearer token аутентификацию. Добавьте заголовок:

```
Authorization: Bearer YOUR_API_TOKEN
```

## Rate Limiting

- **Общие endpoints**: 100 запросов/минуту
- **Статистика**: 10 запросов/минуту  
- **Кэш операции**: 5 запросов/минуту

## Коды ответов

| Код | Описание |
|-----|----------|
| 200 | OK |
| 201 | Created |
| 400 | Bad Request |
| 401 | Unauthorized |
| 404 | Not Found |
| 429 | Too Many Requests |
| 500 | Internal Server Error |

## Примеры ошибок

### 400 Bad Request
```json
{
  "error": "Invalid request parameters",
  "details": "Missing required field: userId"
}
```

### 401 Unauthorized
```json
{
  "error": "Unauthorized",
  "details": "Invalid or missing API token"
}
```

### 429 Rate Limit Exceeded
```json
{
  "error": "Rate limit exceeded",
  "details": "Too many requests. Try again in 60 seconds.",
  "retry_after": 60
}
```

## Тестирование

### Автоматические тесты

Postman коллекция включает автоматические тесты:

- Проверка времени ответа (< 5 секунд)
- Проверка наличия Content-Type заголовка
- Проверка структуры успешных ответов
- Проверка структуры ошибок

### Ручное тестирование

1. **Health Check**: Проверьте доступность сервера
2. **Authentication**: Проверьте аутентификацию
3. **Statistics**: Получите статистику системы
4. **Users**: Протестируйте CRUD операции с пользователями
5. **Cache**: Проверьте управление кэшем

## Мониторинг

### Health Check
```bash
curl http://localhost:8080/health
```

### Metrics (Prometheus)
```bash
curl http://localhost:8080/metrics
```

### Swagger UI
Откройте в браузере: `http://localhost:8080/swagger/`

## Разработка

### Добавление нового endpoint

1. Добавьте метод в `server.go`
2. Обновите Swagger аннотации
3. Добавьте примеры в `examples.md`
4. Добавьте запросы в Postman коллекцию
5. Обновите этот README

### Генерация Swagger документации

```bash
# Установите swag
go install github.com/swaggo/swag/cmd/swag@latest

# Генерируйте документацию
swag init -g cmd/bot/main.go -o docs/api/
```

## Безопасность

- Все административные endpoints требуют аутентификации
- API токены должны быть защищены
- Используйте HTTPS в production
- Настройте rate limiting
- Логируйте все API запросы

## Поддержка

При возникновении проблем:

1. Проверьте логи сервера
2. Убедитесь в правильности API токена
3. Проверьте rate limiting
4. Обратитесь к разработчикам

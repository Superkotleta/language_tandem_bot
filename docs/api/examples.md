# API Examples

Этот документ содержит примеры использования REST API для Language Exchange Bot.

## Базовый URL

```
http://localhost:8080/api/v1
```

## Аутентификация

API использует Bearer token аутентификацию:

```bash
Authorization: Bearer YOUR_API_TOKEN
```

## Endpoints

### 1. Получение статистики

#### Запрос
```bash
curl -X GET "http://localhost:8080/api/v1/stats" \
  -H "Authorization: Bearer YOUR_API_TOKEN" \
  -H "Content-Type: application/json"
```

#### Ответ
```json
{
  "total_users": 1250,
  "active_users": 890,
  "new_users_today": 15,
  "matches_today": 23,
  "successful_matches": 156,
  "cache_stats": {
    "hits": 12500,
    "misses": 250,
    "hit_ratio": 0.98,
    "size": 500,
    "evictions": 5,
    "memory_usage": 1024000
  }
}
```

### 2. Получение списка пользователей

#### Запрос
```bash
curl -X GET "http://localhost:8080/api/v1/users?page=1&limit=10&status=active" \
  -H "Authorization: Bearer YOUR_API_TOKEN" \
  -H "Content-Type: application/json"
```

#### Ответ
```json
{
  "users": [
    {
      "id": 1,
      "telegramId": 123456789,
      "username": "john_doe",
      "firstName": "John",
      "nativeLanguageCode": "en",
      "targetLanguageCode": "ru",
      "targetLanguageLevel": "intermediate",
      "interfaceLanguageCode": "en",
      "state": "active",
      "status": "active",
      "profileCompletionLevel": 100,
      "createdAt": "2024-01-15T10:30:00Z",
      "updatedAt": "2024-01-20T14:45:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 10,
    "total": 1250,
    "totalPages": 125
  }
}
```

### 3. Получение пользователя по ID

#### Запрос
```bash
curl -X GET "http://localhost:8080/api/v1/users/123" \
  -H "Authorization: Bearer YOUR_API_TOKEN" \
  -H "Content-Type: application/json"
```

#### Ответ
```json
{
  "id": 123,
  "telegramId": 987654321,
  "username": "jane_smith",
  "firstName": "Jane",
  "nativeLanguageCode": "ru",
  "targetLanguageCode": "en",
  "targetLanguageLevel": "beginner",
  "interfaceLanguageCode": "ru",
  "state": "active",
  "status": "active",
  "profileCompletionLevel": 85,
  "interests": [1, 3, 5, 7],
  "timeAvailability": {
    "dayType": "weekdays",
    "specificDays": [],
    "timeSlot": "evening"
  },
  "friendshipPreferences": {
    "activityType": "educational",
    "communicationStyle": "text",
    "communicationFrequency": "weekly"
  },
  "createdAt": "2024-01-10T09:15:00Z",
  "updatedAt": "2024-01-22T16:20:00Z"
}
```

### 4. Обновление пользователя

#### Запрос
```bash
curl -X PUT "http://localhost:8080/api/v1/users/123" \
  -H "Authorization: Bearer YOUR_API_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "targetLanguageLevel": "intermediate",
    "profileCompletionLevel": 90,
    "interests": [1, 3, 5, 7, 9]
  }'
```

#### Ответ
```json
{
  "success": true,
  "message": "User updated successfully",
  "user": {
    "id": 123,
    "telegramId": 987654321,
    "username": "jane_smith",
    "firstName": "Jane",
    "nativeLanguageCode": "ru",
    "targetLanguageCode": "en",
    "targetLanguageLevel": "intermediate",
    "interfaceLanguageCode": "ru",
    "state": "active",
    "status": "active",
    "profileCompletionLevel": 90,
    "interests": [1, 3, 5, 7, 9],
    "updatedAt": "2024-01-22T16:25:00Z"
  }
}
```

### 5. Получение языков

#### Запрос
```bash
curl -X GET "http://localhost:8080/api/v1/languages" \
  -H "Authorization: Bearer YOUR_API_TOKEN" \
  -H "Content-Type: application/json"
```

#### Ответ
```json
{
  "languages": [
    {
      "id": 1,
      "code": "en",
      "nameEn": "English",
      "nameNative": "English",
      "isInterfaceLanguage": true,
      "createdAt": "2024-01-01T00:00:00Z"
    },
    {
      "id": 2,
      "code": "ru",
      "nameEn": "Russian",
      "nameNative": "Русский",
      "isInterfaceLanguage": true,
      "createdAt": "2024-01-01T00:00:00Z"
    }
  ]
}
```

### 6. Получение интересов

#### Запрос
```bash
curl -X GET "http://localhost:8080/api/v1/interests" \
  -H "Authorization: Bearer YOUR_API_TOKEN" \
  -H "Content-Type: application/json"
```

#### Ответ
```json
{
  "interests": [
    {
      "id": 1,
      "keyName": "movies",
      "categoryId": 1,
      "displayOrder": 1,
      "type": "entertainment",
      "createdAt": "2024-01-01T00:00:00Z"
    },
    {
      "id": 2,
      "keyName": "music",
      "categoryId": 1,
      "displayOrder": 2,
      "type": "entertainment",
      "createdAt": "2024-01-01T00:00:00Z"
    }
  ]
}
```

### 7. Получение отзывов

#### Запрос
```bash
curl -X GET "http://localhost:8080/api/v1/feedback?page=1&limit=10&is_processed=false" \
  -H "Authorization: Bearer YOUR_API_TOKEN" \
  -H "Content-Type: application/json"
```

#### Ответ
```json
{
  "feedback": [
    {
      "id": 1,
      "userId": 123,
      "message": "Great experience with the language exchange!",
      "rating": 5,
      "isProcessed": false,
      "createdAt": "2024-01-20T14:30:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 10,
    "total": 45,
    "totalPages": 5
  }
}
```

### 8. Обработка отзыва

#### Запрос
```bash
curl -X PUT "http://localhost:8080/api/v1/feedback/1/process" \
  -H "Authorization: Bearer YOUR_API_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "processed": true,
    "adminNotes": "Positive feedback processed"
  }'
```

#### Ответ
```json
{
  "success": true,
  "message": "Feedback processed successfully"
}
```

### 9. Получение кэш статистики

#### Запрос
```bash
curl -X GET "http://localhost:8080/api/v1/cache/stats" \
  -H "Authorization: Bearer YOUR_API_TOKEN" \
  -H "Content-Type: application/json"
```

#### Ответ
```json
{
  "cache_stats": {
    "hits": 12500,
    "misses": 250,
    "hit_ratio": 0.98,
    "size": 500,
    "evictions": 5,
    "memory_usage": 1024000
  },
  "redis_stats": {
    "connected_clients": 5,
    "used_memory": 2048000,
    "total_commands_processed": 50000,
    "keyspace_hits": 12500,
    "keyspace_misses": 250
  }
}
```

### 10. Очистка кэша

#### Запрос
```bash
curl -X POST "http://localhost:8080/api/v1/cache/clear" \
  -H "Authorization: Bearer YOUR_API_TOKEN" \
  -H "Content-Type: application/json"
```

#### Ответ
```json
{
  "success": true,
  "message": "Cache cleared successfully"
}
```

## Коды ошибок

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

### 404 Not Found
```json
{
  "error": "Not Found",
  "details": "User with ID 999 not found"
}
```

### 500 Internal Server Error
```json
{
  "error": "Internal Server Error",
  "details": "Database connection failed"
}
```

## Rate Limiting

API имеет ограничения по количеству запросов:

- **Общие endpoints**: 100 запросов в минуту
- **Статистика**: 10 запросов в минуту
- **Кэш операции**: 5 запросов в минуту

При превышении лимита возвращается ошибка:

```json
{
  "error": "Rate limit exceeded",
  "details": "Too many requests. Try again in 60 seconds.",
  "retry_after": 60
}
```

## Webhook Events

### Telegram Webhook

#### Запрос
```bash
curl -X POST "http://localhost:8080/webhook/telegram" \
  -H "Content-Type: application/json" \
  -d '{
    "update_id": 123456789,
    "message": {
      "message_id": 1,
      "from": {
        "id": 987654321,
        "is_bot": false,
        "first_name": "John",
        "username": "john_doe"
      },
      "chat": {
        "id": 987654321,
        "type": "private"
      },
      "date": 1640995200,
      "text": "/start"
    }
  }'
```

#### Ответ
```json
{
  "success": true,
  "message": "Webhook processed successfully"
}
```

## Тестирование API

### Health Check
```bash
curl -X GET "http://localhost:8080/health"
```

### Metrics
```bash
curl -X GET "http://localhost:8080/metrics"
```

### Swagger UI
```
http://localhost:8080/swagger/
```

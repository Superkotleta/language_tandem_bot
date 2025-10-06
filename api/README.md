# Language Exchange Bot - API Documentation

Этот документ описывает API (gRPC + REST) для микросервисной архитектуры Language Exchange Bot.

## 🎯 API Overview

### REST API (Admin)
- **Base URL:** `http://localhost:8080`
- **Version:** v1/v2
- **Authentication:** `X-Admin-Key` header
- **Documentation:** [Swagger UI](http://localhost:8080/swagger/)

### gRPC API (Services)
- **Protocol:** gRPC with Protocol Buffers
- **Services:** User, Interest, Matcher
- **Authentication:** mTLS for internal communication

## REST API Endpoints

### 🔧 Admin API v1

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| `GET` | `/healthz` | Health check | No |
| `GET` | `/readyz` | Readiness check | No |
| `GET` | `/api/v1/stats` | System statistics | Admin Key |
| `GET` | `/api/v1/users/{id}` | Get user by ID | Admin Key |
| `GET` | `/api/v1/users` | List users (paginated) | Admin Key |
| `GET` | `/api/v1/feedback/unprocessed` | Unprocessed feedback | Admin Key |
| `POST` | `/api/v1/feedback/{id}/process` | Process feedback | Admin Key |
| `GET` | `/api/v1/rate-limits/stats` | Rate limiting stats | Admin Key |
| `GET` | `/api/v1/cache/stats` | Cache statistics | Admin Key |
| `GET` | `/api/v1/webhook/status` | Webhook status | Admin Key |
| `POST` | `/api/v1/webhook/setup` | Setup webhook | Admin Key |
| `POST` | `/api/v1/webhook/remove` | Remove webhook | Admin Key |

### 🚀 Admin API v2 (Enhanced)

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| `GET` | `/api/v2/stats` | Enhanced system statistics | Admin Key |
| `GET` | `/api/v2/system/health` | Detailed system health | Admin Key |
| `GET` | `/api/v2/metrics/performance` | Performance metrics | Admin Key |

### 📊 Authentication

```bash
# All admin API requests require X-Admin-Key header
curl -H "X-Admin-Key: your-admin-secret-key" \
     http://localhost:8080/api/v1/stats
```

## gRPC Services Overview

Система состоит из следующих микросервисов:

1. **User Service** - управление профилями пользователей
2. **Interest Service** - управление интересами и категориями
3. **Matcher Service** - подбор партнеров для языкового обмена

## User Service API

Сервис для управления пользовательскими профилями, регистрацией и базовыми операциями.

### Основные методы

#### GetUser

```protobuf
rpc GetUser(GetUserRequest) returns (GetUserResponse);
```

Получить информацию о пользователе по Telegram ID.

**Параметры:**

- `telegram_id` (int64): Telegram ID пользователя

**Ответ:**

- `user` (User): полная информация о пользователе

#### CreateOrUpdateUser

```protobuf
rpc CreateOrUpdateUser(CreateOrUpdateUserRequest) returns (CreateOrUpdateUserResponse);
```

Создать нового пользователя или обновить существующего.

**Параметры:**

- `user` (User): данные пользователя

#### FindPartners

```protobuf
rpc FindPartners(FindPartnersRequest) returns (FindPartnersResponse);
```

Найти подходящих партнеров для языкового обмена.

**Параметры:**

- `user_id` (int64): ID пользователя
- `limit` (int32): максимальное количество результатов
- `offset` (int32): смещение для пагинации

**Ответ:**

- `partners` ([]User): список подходящих партнеров
- `total_count` (int32): общее количество найденных партнеров

#### UpdateUserLanguages

```protobuf
rpc UpdateUserInterests(UpdateUserInterestsRequest) returns (UpdateUserInterestsResponse);
```

Обновить интересы пользователя.

#### GetUserStats

```protobuf
rpc GetUserStats(GetUserStatsRequest) returns (GetUserStatsResponse);
```

Получить статистику по пользователям системы.

## Interest Service API

Сервис для управления системой интересов, категориями и алгоритмами совместимости.

### Методы Interest Service

#### GetInterests

```protobuf
rpc GetInterests(GetInterestsRequest) returns (GetInterestsResponse);
```

Получить все доступные интересы.

**Параметры:**

- `language_code` (string): код языка для локализации (опционально)

#### GetInterestsByCategories

```protobuf
rpc GetInterestsByCategories(GetInterestsByCategoriesRequest) returns (GetInterestsByCategoriesResponse);
```

Получить интересы, сгруппированные по категориям.

#### UpdateUserInterests

```protobuf
rpc UpdateUserInterests(UpdateUserInterestsRequest) returns (UpdateUserInterestsResponse);
```

Обновить выбор интересов пользователя с указанием приоритета (primary/additional).

#### FindCompatibleInterests

```protobuf
rpc FindCompatibleInterests(FindCompatibleInterestsRequest) returns (FindCompatibleInterestsResponse);
```

Найти совместимые интересы между двумя пользователями и рассчитать балл совместимости.

**Параметры:**

- `user_id` (int64): ID первого пользователя
- `partner_interest_ids` ([]int32): интересы второго пользователя

**Ответ:**

- `matches` ([]InterestMatch): детали совпадений интересов
- `compatibility_score` (int32): общий балл совместимости

## Matcher Service API

Сервис для подбора партнеров с учетом языков, интересов, расписания и предпочтений.

### Методы Matcher Service

#### MatchFindPartners

```protobuf
rpc FindPartners(FindPartnersRequest) returns (FindPartnersResponse);
```

Найти подходящих партнеров с учетом всех критериев совместимости.

**Параметры:**

- `criteria` (MatchCriteria): критерии поиска
- `limit` (int32): максимальное количество результатов
- `include_details` (bool): включать детали совместимости

#### CreateMatch

```protobuf
rpc CreateMatch(CreateMatchRequest) returns (CreateMatchResponse);
```

Создать предложение о матче между двумя пользователями.

#### UpdateMatchStatus

```protobuf
rpc UpdateMatchStatus(UpdateMatchStatusRequest) returns (UpdateMatchStatusResponse);
```

Обновить статус матча (принять/отклонить/завершить).

#### CalculateCompatibility

```protobuf
rpc CalculateCompatibility(CalculateCompatibilityRequest) returns (CalculateCompatibilityResponse);
```

Рассчитать балл совместимости между двумя пользователями.

**Параметры:**

- `user1_id` (int64): ID первого пользователя
- `user2_id` (int64): ID второго пользователя
- `detailed` (bool): возвращать детальную информацию

**Ответ:**

- `score` (int32): балл совместимости (0-100)
- `details` (MatchDetails): детальная информация о совместимости

## Алгоритм совместимости

Система использует многофакторный алгоритм для оценки совместимости партнеров:

### Факторы совместимости

1. **Языковая совместимость (30% веса)**
   - Совпадение родного языка одного с изучаемым языком другого
   - Совпадение уровней владения языком

2. **Совместимость интересов (40% веса)**
   - Совпадение основных интересов (высокий балл)
   - Совпадение дополнительных интересов (средний балл)
   - Количество общих интересов

3. **Совместимость расписания (15% веса)**
   - Совпадение предпочтительного времени общения
   - Совпадение типа дней (будни/выходные)

4. **Совместимость стиля общения (15% веса)**
   - Совпадение предпочтений по типу общения (текст/голос/видео)
   - Совпадение частоты общения

### Расчет итогового балла

```shell
compatibility_score = (language_score * 0.3) +
                     (interest_score * 0.4) +
                     (availability_score * 0.15) +
                     (communication_score * 0.15)
```

Балл совместимости варьируется от 0 до 100, где:

- 90-100: Отличная совместимость
- 70-89: Хорошая совместимость
- 50-69: Приемлемая совместимость
- 0-49: Низкая совместимость

## Установка и использование

### Генерация кода

Для генерации Go-кода из proto файлов используйте protoc:

```bash
# Установка protoc-gen-go и protoc-gen-go-grpc
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Генерация кода для всех сервисов
protoc --go_out=. --go-grpc_out=. api/proto/*.proto
```

### Запуск сервисов

```bash
# User Service
go run services/user/cmd/main.go

# Interest Service
go run services/interest/cmd/main.go

# Matcher Service
go run services/matcher/cmd/main.go
```

## Мониторинг и отладка

### Health Checks

Каждый сервис предоставляет health check эндпоинты:

- `GET /healthz` - проверка здоровья
- `GET /readyz` - проверка готовности

### Метрики

Сервисы экспортируют метрики в формате Prometheus:

- `GET /metrics` - метрики производительности

### Логирование

Все сервисы используют структурированное логирование с уровнями:

- DEBUG, INFO, WARN, ERROR

## Безопасность

### Аутентификация

- Внутреннее общение между сервисами использует mTLS
- API для внешних клиентов требуют JWT токены

### Авторизация

- Role-Based Access Control (RBAC)
- Проверка прав доступа на уровне методов

## Версионирование

API использует семантическое версионирование:

- `v1` - текущая стабильная версия
- Изменения совместимые назад не ломают существующую функциональность
- Критические изменения вводятся в новых major версиях

## Будущие улучшения

1. **Event-Driven Architecture** - переход на асинхронное общение через events
2. **GraphQL API** - для более гибких запросов от клиентов
3. **API Gateway** - единая точка входа для всех сервисов
4. **Service Mesh** - Istio для управления трафиком и observability
5. **Circuit Breaker** - защита от cascade failures

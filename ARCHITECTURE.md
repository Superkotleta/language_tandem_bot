# Архитектура Language Exchange Bot

## 🏗️ Общая архитектура системы

```mermaid
graph TB
    subgraph "External Services"
        TG[Telegram Bot API]
        USER[👤 Users]
    end
    
    subgraph "Language Exchange Bot System"
        BOT[🤖 Bot Service<br/>Go + Docker<br/>• Handlers<br/>• Controllers<br/>• Services<br/>• Validation<br/>• Logging]
        
        subgraph "Data Layer"
            PG[(🗄️ PostgreSQL<br/>Database<br/>• Users<br/>• Profiles<br/>• Interests<br/>• Languages)]
            REDIS[(⚡ Redis<br/>Cache<br/>• Languages<br/>• Interests<br/>• Translations<br/>• User Data)]
        end
        
        PGADMIN[🌐 PgAdmin<br/>Web Interface<br/>Port: 8080]
    end
    
    USER --> TG
    TG --> BOT
    BOT --> PG
    BOT --> REDIS
    PGADMIN --> PG
    
    classDef active fill:#90EE90,stroke:#333,stroke-width:2px
    classDef database fill:#87CEEB,stroke:#333,stroke-width:2px
    classDef external fill:#FFB6C1,stroke:#333,stroke-width:2px
    
    class BOT,PG,REDIS,PGADMIN active
    class PG,REDIS database
    class TG,USER external
```

## 🔧 Текущая архитектура (Упрощенная)

### Активные компоненты

#### 🤖 **Bot Service** - Основной сервис
- **Статус**: ✅ Полностью функционален
- **Технологии**: Go, Telegram Bot API, PostgreSQL, Redis
- **Функции**:
  - Обработка сообщений и команд
  - Управление профилями пользователей
  - Система интересов и языков
  - Административные функции
  - Кэширование и оптимизация

#### 🗄️ **PostgreSQL** - База данных
- **Статус**: ✅ Активна
- **Функции**:
  - Хранение пользовательских данных
  - Профили и настройки
  - Интересы и языки
  - Система отзывов

#### ⚡ **Redis** - Кэширование
- **Статус**: ✅ Активен
- **Функции**:
  - Высокопроизводительное кэширование
  - TTL управление
  - Fallback на in-memory кэш
  - Batch Loading оптимизация

#### 🌐 **PgAdmin** - Администрирование БД
- **Статус**: ✅ Активен
- **Порт**: 8080
- **Функции**: Веб-интерфейс для управления базой данных

### Отключенные компоненты (Временно)

#### 🎯 **Matcher Service** - Подбор партнеров
- **Статус**: ⏸️ Временно отключен
- **Причина**: Проблемы с миграциями
- **Планы**: Восстановление в будущих версиях

#### 👤 **Profile Service** - Управление профилями
- **Статус**: ⏸️ Временно отключен
- **Причина**: Проблемы с миграциями
- **Функциональность**: Перенесена в основной Bot Service

## 🚀 Архитектура кэширования и производительности

```mermaid
graph TD
    subgraph "Bot Service"
        BOT[🤖 Bot Service<br/>• Languages<br/>• Interests<br/>• Users<br/>• Batch Loading]
    end
    
    subgraph "Cache Layer"
        CACHE[🔄 Cache Interface<br/>• Get/Set<br/>• Invalidate<br/>• Stats<br/>• Batch Ops]
        
        subgraph "Cache Storage"
            REDIS[(⚡ Redis Cache<br/>Primary<br/>• Persistent<br/>• TTL Support<br/>• JSON Serial<br/>• Batch Support)]
            MEMORY[(💾 In-Memory Cache<br/>Fallback<br/>• Fast Access<br/>• No Network<br/>• Batch Support)]
        end
    end
    
    subgraph "Optimization Layer"
        BATCH[📊 Batch Loader<br/>• N+1 Fix<br/>• JOIN Queries<br/>• 75% Reduction]
        DB[(🗄️ PostgreSQL<br/>Database)]
    end
    
    BOT --> CACHE
    CACHE --> REDIS
    CACHE --> MEMORY
    CACHE --> BATCH
    BATCH --> DB
    
    classDef service fill:#90EE90,stroke:#333,stroke-width:2px
    classDef cache fill:#FFD700,stroke:#333,stroke-width:2px
    classDef database fill:#87CEEB,stroke:#333,stroke-width:2px
    classDef optimization fill:#DDA0DD,stroke:#333,stroke-width:2px
    
    class BOT service
    class CACHE,REDIS,MEMORY cache
    class DB database
    class BATCH optimization
```

## 📊 Потоки данных

### 1. Пользовательский поток

```mermaid
sequenceDiagram
    participant U as 👤 User
    participant TG as 📱 Telegram
    participant BOT as 🤖 Bot Service
    participant CACHE as ⚡ Cache
    participant DB as 🗄️ Database
    
    U->>TG: Send Message
    TG->>BOT: Process Message
    BOT->>CACHE: Check Cache
    alt Cache Hit
        CACHE-->>BOT: Return Cached Data
    else Cache Miss
        BOT->>DB: Query Database
        DB-->>BOT: Return Data
        BOT->>CACHE: Store in Cache
    end
    BOT->>TG: Send Response
    TG->>U: Display Message
```

### 2. Административный поток

```mermaid
sequenceDiagram
    participant A as 👨‍💼 Admin
    participant TG as 📱 Telegram
    participant BOT as 🤖 Bot Service
    participant AUTH as 🔐 Auth Check
    participant DB as 🗄️ Database
    
    A->>TG: /admin command
    TG->>BOT: Process Command
    BOT->>AUTH: Check Admin Rights
    alt Authorized
        AUTH-->>BOT: Access Granted
        BOT->>DB: Query Statistics
        DB-->>BOT: Return Data
        BOT->>TG: Send Admin Report
        TG->>A: Display Statistics
    else Unauthorized
        AUTH-->>BOT: Access Denied
        BOT->>TG: Send Error Message
        TG->>A: Display Error
    end
```

### 3. Система кэширования

```mermaid
flowchart TD
    REQ[📥 Request] --> CHECK{🔍 Cache Check}
    CHECK -->|Hit| HIT[✅ Cache Hit<br/>Return Data]
    CHECK -->|Miss| MISS[❌ Cache Miss]
    MISS --> DB_QUERY[🗄️ Database Query]
    DB_QUERY --> STORE[💾 Store in Cache]
    STORE --> RETURN[📤 Return Data]
    HIT --> RETURN
    
    subgraph "Cache Layers"
        REDIS_CHECK[⚡ Redis Check]
        MEMORY_CHECK[💾 Memory Check]
    end
    
    MISS --> REDIS_CHECK
    REDIS_CHECK -->|Available| MEMORY_CHECK
    REDIS_CHECK -->|Unavailable| MEMORY_CHECK
    
    classDef process fill:#E6F3FF,stroke:#333,stroke-width:2px
    classDef decision fill:#FFF2CC,stroke:#333,stroke-width:2px
    classDef storage fill:#E1F5FE,stroke:#333,stroke-width:2px
    
    class REQ,RETURN process
    class CHECK decision
    class REDIS_CHECK,MEMORY_CHECK,DB_QUERY,STORE storage
```

## 🛡️ Система обработки ошибок

### Архитектура обработки ошибок

```mermaid
graph TD
    subgraph "Error Types"
        TG_ERR[📱 Telegram API<br/>ErrorTypeTelegramAPI]
        DB_ERR[🗄️ Database<br/>ErrorTypeDatabase]
        VAL_ERR[✅ Validation<br/>ErrorTypeValidation]
        CACHE_ERR[⚡ Cache<br/>ErrorTypeCache]
        NET_ERR[🌐 Network<br/>ErrorTypeNetwork]
        INT_ERR[🔧 Internal<br/>ErrorTypeInternal]
    end
    
    subgraph "Error Processing"
        HANDLER[🛡️ Error Handler<br/>Centralized Processing]
        TRACE[🔍 RequestID Tracing<br/>req_1759152914113401600_2914]
        LOG[📝 Structured Logging<br/>JSON Format]
        ALERT[🚨 Admin Alerts<br/>Critical Errors]
    end
    
    subgraph "Error Context"
        CTX[📋 Request Context<br/>userID, chatID, operation]
        SEVERITY[⚠️ Severity Levels<br/>DEBUG, INFO, WARN, ERROR]
    end
    
    TG_ERR --> HANDLER
    DB_ERR --> HANDLER
    VAL_ERR --> HANDLER
    CACHE_ERR --> HANDLER
    NET_ERR --> HANDLER
    INT_ERR --> HANDLER
    
    HANDLER --> TRACE
    HANDLER --> LOG
    HANDLER --> ALERT
    
    CTX --> HANDLER
    SEVERITY --> LOG
    
    classDef error fill:#FFB6C1,stroke:#333,stroke-width:2px
    classDef process fill:#90EE90,stroke:#333,stroke-width:2px
    classDef context fill:#87CEEB,stroke:#333,stroke-width:2px
    
    class TG_ERR,DB_ERR,VAL_ERR,CACHE_ERR,NET_ERR,INT_ERR error
    class HANDLER,TRACE,LOG,ALERT process
    class CTX,SEVERITY context
```

### Типизированные ошибки
```go
ErrorTypeTelegramAPI  // Ошибки Telegram API
ErrorTypeDatabase     // Ошибки базы данных
ErrorTypeValidation   // Ошибки валидации
ErrorTypeCache        // Ошибки кэша
ErrorTypeNetwork      // Сетевые ошибки
ErrorTypeInternal     // Внутренние ошибки
```

### RequestID трейсинг
```go
ctx := errors.NewRequestContext(userID, chatID, "SendMessage")
// RequestID: req_1759152914113401600_2914
```

### Централизованная обработка
```go
return errorHandler.HandleTelegramError(
    err,
    message.Chat.ID,
    int64(user.ID),
    "SendMessage",
)
```

## 📝 Структурированное логирование

### Уровни логирования
- **DEBUG**: Детальная отладочная информация
- **INFO**: Общая информация о работе
- **WARN**: Предупреждения
- **ERROR**: Ошибки

### Специализированные логгеры
- **TelegramLogger**: Сообщения, команды, callback'и
- **DatabaseLogger**: Запросы, транзакции, соединения
- **CacheLogger**: Попадания/промахи кэша, инвалидация
- **ValidationLogger**: Валидация данных

### JSON формат логов
```json
{
  "timestamp": "2025-09-29T20:45:21.903065157+07:00",
  "level": 1,
  "message": "Message received",
  "request_id": "req_123",
  "user_id": 67890,
  "chat_id": 12345,
  "operation": "HandleMessage",
  "component": "telegram",
  "fields": {
    "text_length": 11,
    "has_text": true
  }
}
```

## ✅ Система валидации

### Базовые валидаторы
```go
// Валидация строк
validator.ValidateString("text", []string{"required", "max:50"})

// Валидация Telegram ID
validator.ValidateTelegramID(123456789)

// Валидация кода языка
validator.ValidateLanguageCode("en")

// Валидация состояния пользователя
validator.ValidateUserState("idle")
```

### Специализированные валидаторы
- **UserValidator**: Валидация пользователей и регистрации
- **MessageValidator**: Валидация сообщений и callback'ов
- **ValidationService**: Интеграция с системой ошибок

## 🚀 Развертывание

### Docker Compose архитектура

```mermaid
graph TB
    subgraph "Docker Network"
        subgraph "Application Layer"
            BOT[🤖 Bot Service<br/>Port: 8080<br/>Go + Docker]
        end
        
        subgraph "Data Layer"
            PG[(🗄️ PostgreSQL<br/>Port: 5432<br/>Database)]
            REDIS[(⚡ Redis<br/>Port: 6379<br/>Cache)]
        end
        
        subgraph "Management Layer"
            PGADMIN[🌐 PgAdmin<br/>Port: 8080<br/>Web Interface]
        end
    end
    
    subgraph "External"
        TG[📱 Telegram API]
        USER[👤 Users]
        ADMIN[👨‍💼 Admins]
    end
    
    USER --> TG
    TG --> BOT
    ADMIN --> PGADMIN
    BOT --> PG
    BOT --> REDIS
    PGADMIN --> PG
    
    classDef app fill:#90EE90,stroke:#333,stroke-width:2px
    classDef data fill:#87CEEB,stroke:#333,stroke-width:2px
    classDef mgmt fill:#DDA0DD,stroke:#333,stroke-width:2px
    classDef external fill:#FFB6C1,stroke:#333,stroke-width:2px
    
    class BOT app
    class PG,REDIS data
    class PGADMIN mgmt
    class TG,USER,ADMIN external
```

### Docker Compose сервисы
```yaml
services:
  bot:          # Основной Telegram бот
  postgres:     # База данных PostgreSQL
  redis:        # Кэш-сервер Redis
  pgadmin:      # Веб-интерфейс для БД
```

### Порты
- **Bot Service**: 8080 (HTTP API)
- **PostgreSQL**: 5432
- **Redis**: 6379
- **PgAdmin**: 8080 (веб-интерфейс)

### Переменные окружения
- **TELEGRAM_TOKEN**: Токен бота от @BotFather
- **ADMIN_CHAT_IDS**: Chat ID администраторов
- **ADMIN_USERNAMES**: Username администраторов
- **REDIS_URL**: Адрес Redis сервера
- **DATABASE_URL**: Строка подключения к БД

## 🔮 Планы развития

### Roadmap развития системы

```mermaid
gantt
    title Language Exchange Bot Development Roadmap
    dateFormat  YYYY-MM-DD
    section Phase 1 - Current
    Core Bot Functionality   :crit, core, 2025-09-01, 2025-09-29
    Redis Caching            :crit, cache, 2025-09-15, 2025-09-29
    Batch Loading            :crit, batch, 2025-09-20, 2025-09-29
    Error Handling           :crit, error, 2025-09-25, 2025-09-29
    
    section Phase 2 - Microservices
    Matcher Service          :active, matcher, 2025-10-01, 2025-10-15
    Profile Service          :profile, 2025-10-10, 2025-10-25
    API Gateway              :gateway, 2025-10-20, 2025-11-05
    
    section Phase 3 - Scaling
    Webhook Support          :webhook, 2025-11-01, 2025-11-15
    Redis Clustering         :redis-cluster, 2025-11-10, 2025-11-25
    Monitoring & Metrics     :monitoring, 2025-11-20, 2025-12-05
    
    section Phase 4 - DevOps
    CI/CD Pipeline           :cicd, 2025-12-01, 2025-12-15
    Auto Deployment          :deploy, 2025-12-10, 2025-12-25
```

### Архитектура будущего развития

```mermaid
graph TB
    subgraph "Current Architecture"
        BOT[🤖 Bot Service<br/>Monolithic]
        PG[(🗄️ PostgreSQL)]
        REDIS[(⚡ Redis)]
    end
    
    subgraph "Future Microservices"
        GATEWAY[🌐 API Gateway<br/>Load Balancer]
        
        subgraph "Core Services"
            BOT_MS[🤖 Bot Service<br/>Microservice]
            MATCHER[🎯 Matcher Service<br/>Partner Matching]
            PROFILE[👤 Profile Service<br/>User Management]
        end
        
        subgraph "Infrastructure"
            PG_CLUSTER[(🗄️ PostgreSQL<br/>Cluster)]
            REDIS_CLUSTER[(⚡ Redis<br/>Cluster)]
            MONITOR[📊 Monitoring<br/>Prometheus + Grafana]
        end
    end
    
    BOT -.->|Migration| GATEWAY
    PG -.->|Scaling| PG_CLUSTER
    REDIS -.->|Clustering| REDIS_CLUSTER
    
    GATEWAY --> BOT_MS
    GATEWAY --> MATCHER
    GATEWAY --> PROFILE
    
    BOT_MS --> PG_CLUSTER
    MATCHER --> PG_CLUSTER
    PROFILE --> PG_CLUSTER
    
    BOT_MS --> REDIS_CLUSTER
    MATCHER --> REDIS_CLUSTER
    PROFILE --> REDIS_CLUSTER
    
    MONITOR --> BOT_MS
    MONITOR --> MATCHER
    MONITOR --> PROFILE
    
    classDef current fill:#90EE90,stroke:#333,stroke-width:2px
    classDef future fill:#FFD700,stroke:#333,stroke-width:2px
    classDef infrastructure fill:#87CEEB,stroke:#333,stroke-width:2px
    
    class BOT,PG,REDIS current
    class GATEWAY,BOT_MS,MATCHER,PROFILE future
    class PG_CLUSTER,REDIS_CLUSTER,MONITOR infrastructure
```

### Восстановление микросервисов
1. **Matcher Service** - алгоритмы подбора партнеров
2. **Profile Service** - выделенное управление профилями
3. **API Gateway** - единая точка входа для микросервисов

### Дополнительные возможности
1. **Webhook поддержка** - для высоконагруженных систем
2. **Кластеризация Redis** - для масштабирования
3. **Мониторинг и метрики** - Prometheus + Grafana
4. **CI/CD пайплайн** - автоматическое развертывание

---

**Статус**: Система готова к продакшену с упрощенной архитектурой. Все критические ошибки исправлены, производительность оптимизирована.

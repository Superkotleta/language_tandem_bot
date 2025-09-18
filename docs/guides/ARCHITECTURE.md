# 🏗️ Архитектура Language Exchange Bot

## 📋 Обзор системы

Language Exchange Bot построен на микросервисной архитектуре с четким разделением ответственности, высокой производительностью и отказоустойчивостью.

## 🏛️ Общая архитектура

### Системная диаграмма

```mermaid
graph TB
    subgraph "👥 Пользователи"
        User[👤 Пользователь]
        Admin[👨‍💼 Администратор]
    end
    
    subgraph "📱 Интерфейсы"
        TG[📱 Telegram Bot API]
        Web[🌐 Web Interface]
    end
    
    subgraph "🤖 Микросервисы"
        Bot[🤖 Bot Service<br/>:8080]
        Profile[👤 Profile Service<br/>:8081]
        Matcher[🎯 Matcher Service<br/>:8082]
    end
    
    subgraph "💾 Хранилище данных"
        DB[(🗄️ PostgreSQL<br/>:5432)]
        Cache[🔴 Redis Cache<br/>:6379]
    end
    
    subgraph "🔧 Администрирование"
        PgAdmin[🔧 PgAdmin<br/>:8080]
        Monitor[📊 Мониторинг]
    end
    
    subgraph "📈 Наблюдаемость"
        Prometheus[📈 Prometheus<br/>:9090]
        Grafana[📊 Grafana<br/>:3000]
        Logs[📝 Structured Logs]
    end
    
    %% Пользовательские потоки
    User --> TG
    TG --> Bot
    Bot --> Profile
    Bot --> Matcher
    Bot --> Cache
    Bot --> DB
    
    %% Административные потоки
    Admin --> PgAdmin
    Admin --> Web
    PgAdmin --> DB
    
    %% Внутренние связи
    Profile --> DB
    Matcher --> DB
    Profile --> Cache
    Matcher --> Cache
    
    %% Мониторинг
    Bot --> Prometheus
    Profile --> Prometheus
    Matcher --> Prometheus
    Bot --> Logs
    Profile --> Logs
    Matcher --> Logs
    Prometheus --> Grafana
    Monitor --> Grafana
    
    %% Стили
    classDef userClass fill:#e1f5fe,stroke:#01579b,stroke-width:2px
    classDef serviceClass fill:#f3e5f5,stroke:#4a148c,stroke-width:2px
    classDef storageClass fill:#e8f5e8,stroke:#1b5e20,stroke-width:2px
    classDef adminClass fill:#fff3e0,stroke:#e65100,stroke-width:2px
    classDef monitorClass fill:#fce4ec,stroke:#880e4f,stroke-width:2px
    
    class User,Admin userClass
    class Bot,Profile,Matcher serviceClass
    class DB,Cache storageClass
    class PgAdmin,Web adminClass
    class Prometheus,Grafana,Logs,Monitor monitorClass
```

### Детальная архитектура сервисов

```mermaid
graph LR
    subgraph "🤖 Bot Service"
        BotAPI[📱 Telegram API]
        BotCore[🧠 Core Logic]
        BotHandlers[⚡ Message Handlers]
        BotLocalization[🌐 Localization]
    end
    
    subgraph "👤 Profile Service"
        ProfileAPI[🔌 REST API]
        ProfileCore[🧠 Profile Logic]
        ProfileValidation[✅ Validation]
        ProfileCache[💾 Profile Cache]
    end
    
    subgraph "🎯 Matcher Service"
        MatcherAPI[🔌 REST API]
        MatcherCore[🧠 Matching Logic]
        CompatibilityEngine[⚖️ Compatibility Engine]
        MatchQueue[📋 Match Queue]
    end
    
    subgraph "💾 Data Layer"
        PostgreSQL[(🗄️ PostgreSQL)]
        Redis[(🔴 Redis)]
        Migrations[📋 Database Migrations]
    end
    
    %% Связи между компонентами
    BotAPI --> BotCore
    BotCore --> BotHandlers
    BotCore --> BotLocalization
    BotCore --> ProfileAPI
    BotCore --> MatcherAPI
    
    ProfileAPI --> ProfileCore
    ProfileCore --> ProfileValidation
    ProfileCore --> ProfileCache
    ProfileCore --> PostgreSQL
    
    MatcherAPI --> MatcherCore
    MatcherCore --> CompatibilityEngine
    MatcherCore --> MatchQueue
    MatcherCore --> PostgreSQL
    
    %% Кэширование
    BotCore --> Redis
    ProfileCache --> Redis
    MatcherCore --> Redis
    
    %% Миграции
    Migrations --> PostgreSQL
```

## 🎯 Сервисы и их роли

### 🤖 Bot Service (Основной бот)

**Порт**: 8080  
**Ответственность**:

- Обработка Telegram сообщений
- Пользовательский интерфейс
- Локализация
- Административные функции
- Обратная связь

**Технологии**:

- Go 1.21 + Telegram Bot API
- Redis для кэширования
- Zap для логирования
- Prometheus для метрик

**Endpoints**:

- `GET /health` - Health check
- `GET /metrics` - Prometheus метрики
- `POST /webhook` - Telegram webhook

### 👤 Profile Service

**Порт**: 8081  
**Ответственность**:

- CRUD операции с профилями
- Управление языковыми настройками
- Интересы и предпочтения
- Статистика пользователей

**API Endpoints**:

```http
GET /profiles/{user_id}      # Получение профиля
PUT /profiles/{user_id}      # Обновление профиля
DELETE /profiles/{user_id}   # Удаление профиля
GET /profiles/{user_id}/stats # Статистика
```

### 🎯 Matcher Service

**Порт**: 8082  
**Ответственность**:

- Алгоритмы подбора партнеров
- Совместимость по языкам
- Фильтрация по интересам
- Очередь матчинга

**API Endpoints**:

```http
POST /matches/find           # Поиск партнеров
GET /matches/{user_id}       # Текущие матчи
POST /matches/feedback       # Обратная связь по матчу
```

## 🗄️ База данных

### PostgreSQL 15

**Структура схем**:

- `public` - Основные таблицы
- `profile` - Данные профилей
- `matching` - Алгоритмы подбора
- `feedback` - Система отзывов

### Основные таблицы

```sql
-- Пользователи
users (
    id SERIAL PRIMARY KEY,
    telegram_id BIGINT UNIQUE,
    username VARCHAR(255),
    first_name VARCHAR(255),
    interface_language_code VARCHAR(10),
    native_language_code VARCHAR(10),
    target_language_code VARCHAR(10),
    target_language_level VARCHAR(10),
    status VARCHAR(50),
    profile_completion_level INTEGER,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);

-- Языки
languages (
    id SERIAL PRIMARY KEY,
    code VARCHAR(10) UNIQUE,
    name VARCHAR(100),
    english_name VARCHAR(100)
);

-- Интересы
interests (
    id SERIAL PRIMARY KEY,
    name_key VARCHAR(100),
    category VARCHAR(50)
);

-- Интересы пользователей
user_interests (
    user_id INTEGER REFERENCES users(id),
    interest_id INTEGER REFERENCES interests(id),
    is_primary BOOLEAN DEFAULT false,
    PRIMARY KEY (user_id, interest_id)
);

-- Обратная связь
feedback (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    feedback_text TEXT,
    contact_info TEXT,
    is_processed BOOLEAN DEFAULT false,
    admin_response TEXT,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);
```

## 🔴 Redis Cache

### Структура кэширования

```redis
# Языки (TTL: 24 часа)
languages: [{"id":1,"code":"en","name":"English"}...]

# Интересы по языкам (TTL: 12 часов)
interests:ru: {"1":"Фильмы","2":"Музыка"}
interests:en: {"1":"Movies","2":"Music"}

# Профили пользователей (TTL: 30 минут)
user:12345: {"id":12345,"name":"Ivan","native":"ru"}

# Результаты поиска (TTL: 15 минут)
matches:12345: [{"user_id":67890,"compatibility":95}]

# Статистика (TTL: 1 час)
stats:daily: {"active_users":150,"new_profiles":23}
```

### Стратегии кэширования

- **Cache-Aside**: Профили пользователей
- **Write-Through**: Статистика
- **Write-Behind**: Метрики производительности

## 🔄 Взаимодействие сервисов

### 1. Регистрация пользователя

```mermaid
sequenceDiagram
    participant U as User
    participant B as Bot Service
    participant P as Profile Service
    participant DB as PostgreSQL
    participant R as Redis

    U->>B: /start
    B->>P: POST /profiles
    P->>DB: INSERT user
    P->>B: 201 Created
    B->>R: CACHE user profile
    B->>U: Welcome message
```

### 2. Поиск партнера

```mermaid
sequenceDiagram
    participant U as User
    participant B as Bot Service
    participant M as Matcher Service
    participant R as Redis
    participant DB as PostgreSQL

    U->>B: Find partner
    B->>R: GET cached matches
    alt Cache miss
        B->>M: POST /matches/find
        M->>DB: SELECT compatible users
        M->>B: Matches list
        B->>R: CACHE matches
    end
    B->>U: Partner suggestions
```

## 📊 Мониторинг и наблюдаемость

### Метрики Prometheus

```yaml
# Bot Service
telegram_messages_total         # Счетчик сообщений
telegram_commands_duration      # Время выполнения команд
database_queries_total          # Счетчик запросов к БД
cache_hits_total               # Попадания в кэш
cache_misses_total            # Промахи кэша

# Profile Service
profiles_created_total         # Созданные профили
profiles_updated_total        # Обновленные профили
api_requests_duration_seconds # Время отклика API

# Matcher Service
matches_found_total           # Найденные матчи
matching_algorithm_duration  # Время выполнения алгоритма
compatibility_score_histogram # Распределение совместимости
```

### Структурированное логирование

```json
{
  "timestamp": "2025-09-18T12:00:00Z",
  "level": "info",
  "service": "bot",
  "component": "telegram_handler",
  "message": "User profile updated",
  "user_id": 12345,
  "telegram_id": 123456789,
  "action": "profile_update",
  "duration_ms": 150,
  "request_id": "req_abc123",
  "metadata": {
    "language": "ru",
    "completion_level": 85
  }
}
```

## 🛡️ Безопасность

### Защита на уровне сети

- **Rate Limiting**: 100 запросов/минуту на пользователя
- **IP Whitelisting**: Ограничение доступа к админ API
- **DDoS Protection**: Circuit Breaker паттерн

### Защита данных

- **Шифрование**: TLS 1.3 для всех соединений
- **Валидация**: Строгая проверка всех входных данных
- **Санитизация**: Очистка пользовательского ввода

### Аутентификация и авторизация

```go
// Проверка администратора
func (h *AdminHandler) IsAdmin(chatID int64, username string) bool {
    // Проверка по Chat ID
    for _, adminID := range h.adminChatIDs {
        if chatID == adminID {
            return true
        }
    }
    
    // Проверка по Username
    for _, adminUsername := range h.adminUsernames {
        if username == adminUsername {
            return true
        }
    }
    
    return false
}
```

## 🚀 Масштабирование

### Горизонтальное масштабирование

- **Stateless сервисы**: Все состояние в Redis/PostgreSQL
- **Load Balancer**: Nginx для распределения нагрузки
- **Database Replication**: Master-Slave для чтения

### Вертикальное масштабирование

- **Connection Pooling**: Оптимальное использование БД
- **Batch Operations**: Массовые операции
- **Async Processing**: Неблокирующие операции

### Автомасштабирование

```yaml
# docker-compose.yml
services:
  bot:
    deploy:
      replicas: 3
      resources:
        limits:
          cpus: '0.50'
          memory: 512M
        reservations:
          cpus: '0.25'
          memory: 256M
```

## 🔄 Паттерны и принципы

### Clean Architecture

```shell
┌─────────────────────────────────────┐
│           Frameworks & Drivers      │
│  (Telegram API, PostgreSQL, Redis) │
├─────────────────────────────────────┤
│        Interface Adapters           │
│     (Controllers, Gateways)         │
├─────────────────────────────────────┤
│         Application Business        │
│         Rules (Use Cases)           │
├─────────────────────────────────────┤
│        Enterprise Business          │
│           Rules (Entities)          │
└─────────────────────────────────────┘
```

### SOLID принципы

- **S**ingle Responsibility: Один сервис = одна ответственность
- **O**pen/Closed: Расширяемость через интерфейсы
- **L**iskov Substitution: Заменяемость реализаций
- **I**nterface Segregation: Минимальные интерфейсы
- **D**ependency Inversion: Зависимость от абстракций

### Паттерны проектирования

- **Repository**: Абстракция доступа к данным
- **Factory**: Создание сервисов
- **Strategy**: Алгоритмы подбора партнеров
- **Observer**: Уведомления администраторов
- **Circuit Breaker**: Отказоустойчивость

## 📈 Производительность

### Оптимизации

- **Redis кэширование**: 50x ускорение доступа к данным
- **Connection pooling**: Эффективное использование БД
- **Batch операции**: Массовые обновления профилей
- **Lazy loading**: Загрузка данных по требованию

### Benchmarks

```shell
Операция                    | Без кэша  | С кэшем   | Ускорение
----------------------------|-----------|-----------|----------
Загрузка языков            | 50ms      | 1ms       | 50x
Получение профиля          | 25ms      | 2ms       | 12.5x
Поиск интересов           | 30ms      | 1.5ms     | 20x
Локализация               | 15ms      | 0.5ms     | 30x
```

## 🔮 Будущее развитие

### Планируемые улучшения

- **GraphQL API**: Более гибкие запросы
- **Event Sourcing**: Аудит всех изменений
- **CQRS**: Разделение команд и запросов
- **WebSocket**: Реальное время общения
- **ML Matching**: ИИ для лучшего подбора

### Новые сервисы

- **Notification Service**: Централизованные уведомления
- **Analytics Service**: Продвинутая аналитика
- **Chat Service**: Встроенный чат
- **Recommendation Service**: Рекомендации контента

---

**Документация**: [README.md](../README.md)  
**Настройка**: [SETUP_GUIDE.md](SETUP_GUIDE.md)  
**Безопасность**: [SECURITY.md](../reports/SECURITY.md)

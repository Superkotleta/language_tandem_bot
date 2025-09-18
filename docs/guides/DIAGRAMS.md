# 📊 Диаграммы Language Exchange Bot

## 🎯 Обзор

Этот документ содержит все Mermaid диаграммы проекта для лучшего понимания архитектуры и процессов.

## 🏗️ Архитектурные диаграммы

### Общая системная архитектура

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

### Детальная архитектура компонентов

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

### Схема базы данных

```mermaid
erDiagram
    USERS {
        int id PK
        bigint telegram_id UK
        string username
        string first_name
        string interface_language_code
        string native_language_code
        string target_language_code
        string target_language_level
        string status
        int profile_completion_level
        timestamp created_at
        timestamp updated_at
    }
    
    LANGUAGES {
        int id PK
        string code UK
        string name
        string english_name
    }
    
    INTERESTS {
        int id PK
        string name_key UK
        string category
    }
    
    USER_INTERESTS {
        int user_id FK
        int interest_id FK
        boolean is_primary
        timestamp created_at
    }
    
    FEEDBACK {
        int id PK
        int user_id FK
        text feedback_text
        text contact_info
        boolean is_processed
        text admin_response
        timestamp created_at
        timestamp updated_at
    }
    
    MATCHES {
        int id PK
        int user1_id FK
        int user2_id FK
        decimal compatibility_score
        text match_reason
        string status
        timestamp created_at
        timestamp updated_at
    }
    
    MATCH_QUEUE {
        int id PK
        int user_id FK
        decimal priority_score
        string status
        timestamp created_at
        timestamp updated_at
    }
    
    %% Связи
    USERS ||--o{ USER_INTERESTS : has
    INTERESTS ||--o{ USER_INTERESTS : belongs_to
    USERS ||--o{ FEEDBACK : writes
    USERS ||--o{ MATCHES : participates_in
    USERS ||--o{ MATCH_QUEUE : queued_in
```

## 🔄 Sequence диаграммы процессов

### Процесс поиска языкового партнера

```mermaid
sequenceDiagram
    participant U as 👤 Пользователь
    participant B as 🤖 Bot Service
    participant M as 🎯 Matcher Service
    participant P as 👤 Profile Service
    participant R as 🔴 Redis Cache
    participant DB as 🗄️ PostgreSQL

    Note over U,DB: Процесс поиска языкового партнера

    U->>B: /find_partner
    B->>P: GET /profiles/{user_id}
    P->>DB: SELECT user profile
    DB-->>P: User data
    P-->>B: Profile info

    B->>M: POST /matches/find
    Note over M: Алгоритм подбора партнеров
    
    M->>DB: SELECT compatible candidates
    DB-->>M: Candidate list
    
    loop Для каждого кандидата
        M->>R: GET cached compatibility
        alt Cache hit
            R-->>M: Cached score
        else Cache miss
            M->>M: Calculate compatibility
            M->>R: SET compatibility score
        end
    end
    
    M->>M: Find best match
    M->>DB: INSERT match result
    M-->>B: Best partner + score
    
    B->>U: Partner suggestion
    U->>B: Accept/Decline
    
    alt Accept
        B->>M: POST /matches/confirm
        M->>DB: UPDATE match status
        M-->>B: Match confirmed
        B->>U: Success message
    else Decline
        B->>M: POST /matches/decline
        M->>DB: UPDATE match status
        M-->>B: Match declined
        B->>U: Try again message
    end
```

### Процесс регистрации пользователя

```mermaid
sequenceDiagram
    participant U as 👤 Пользователь
    participant B as 🤖 Bot Service
    participant P as 👤 Profile Service
    participant R as 🔴 Redis Cache
    participant DB as 🗄️ PostgreSQL
    participant L as 🌐 Localization

    Note over U,L: Процесс регистрации и настройки профиля

    U->>B: /start
    B->>L: Get welcome message
    L-->>B: Localized message
    B->>U: Welcome + main menu

    U->>B: Setup profile
    B->>U: Language selection
    
    U->>B: Select native language
    B->>P: PUT /profiles/{user_id}/native_lang
    P->>DB: UPDATE user profile
    DB-->>P: Success
    P-->>B: Profile updated

    U->>B: Select target language
    B->>P: PUT /profiles/{user_id}/target_lang
    P->>DB: UPDATE user profile
    DB-->>P: Success
    P-->>B: Profile updated

    U->>B: Select interests
    B->>P: PUT /profiles/{user_id}/interests
    P->>DB: INSERT user interests
    DB-->>P: Success
    P-->>B: Interests saved

    B->>P: GET /profiles/{user_id}/completion
    P->>DB: Calculate completion level
    DB-->>P: Completion percentage
    P-->>B: 100% complete

    B->>R: CACHE user profile
    R-->>B: Cached
    B->>U: Profile completed! 🎉
```

### Система обратной связи

```mermaid
sequenceDiagram
    participant U as 👤 Пользователь
    participant B as 🤖 Bot Service
    participant P as 👤 Profile Service
    participant A as 👨‍💼 Администратор
    participant DB as 🗄️ PostgreSQL
    participant N as 🔔 Notifications

    Note over U,N: Процесс отправки и обработки обратной связи

    U->>B: /feedback
    B->>U: Request feedback text
    U->>B: Feedback message
    B->>U: Contact info? (optional)
    U->>B: Contact details

    B->>P: POST /feedback
    P->>DB: INSERT feedback
    DB-->>P: Feedback ID
    P-->>B: Feedback saved

    B->>N: Send admin notification
    N->>A: New feedback alert
    A->>B: /admin feedbacks

    B->>P: GET /feedback/unprocessed
    P->>DB: SELECT unprocessed feedback
    DB-->>P: Feedback list
    P-->>B: Feedback data

    B->>A: Show feedback with actions
    
    alt Process feedback
        A->>B: Process feedback
        B->>P: PUT /feedback/{id}/process
        P->>DB: UPDATE feedback status
        DB-->>P: Success
        P-->>B: Feedback processed
        B->>A: Success message
    else Delete feedback
        A->>B: Delete feedback
        B->>P: DELETE /feedback/{id}
        P->>DB: DELETE feedback
        DB-->>P: Success
        P-->>B: Feedback deleted
        B->>A: Success message
    end
```

### Мониторинг и логирование

```mermaid
sequenceDiagram
    participant App as 🤖 Application
    participant Logger as 📝 Logger
    participant Metrics as 📊 Metrics
    participant Prometheus as 📈 Prometheus
    participant Grafana as 📊 Grafana
    participant AlertManager as 🚨 AlertManager

    Note over App,AlertManager: Процесс мониторинга и алертинга

    App->>Logger: Log event
    Logger->>Logger: Format structured log
    Logger->>Logger: Write to file/stream

    App->>Metrics: Record metric
    Metrics->>Prometheus: Expose metrics
    Prometheus->>Prometheus: Scrape metrics
    Prometheus->>Grafana: Query metrics
    Grafana->>Grafana: Render dashboard

    alt Metric threshold exceeded
        Prometheus->>AlertManager: Trigger alert
        AlertManager->>AlertManager: Evaluate rules
        AlertManager->>AlertManager: Send notification
    end
```

## 🎯 Диаграммы алгоритмов

### Алгоритм расчета совместимости

```mermaid
flowchart TD
    Start([Начало расчета совместимости]) --> CheckReciprocal{Проверить<br/>взаимность языков}
    
    CheckReciprocal -->|Нет взаимности| NoMatch([Совместимость = 0%])
    CheckReciprocal -->|Есть взаимность| CalcLanguage[Рассчитать языковую<br/>совместимость 40%]
    
    CalcLanguage --> CalcInterests[Рассчитать совместимость<br/>интересов 25%]
    CalcInterests --> CalcTime[Рассчитать временную<br/>совместимость 20%]
    CalcTime --> CalcPersonality[Рассчитать личностную<br/>совместимость 15%]
    
    CalcPersonality --> WeightedSum[Взвешенная сумма<br/>всех компонентов]
    WeightedSum --> CheckThreshold{Совместимость >= 60%?}
    
    CheckThreshold -->|Да| GoodMatch([Хороший матч])
    CheckThreshold -->|Нет| PoorMatch([Плохой матч])
    
    NoMatch --> End([Конец])
    GoodMatch --> End
    PoorMatch --> End
    
    %% Стили
    classDef startEnd fill:#e8f5e8,stroke:#2e7d32,stroke-width:2px
    classDef process fill:#e3f2fd,stroke:#1565c0,stroke-width:2px
    classDef decision fill:#fff3e0,stroke:#ef6c00,stroke-width:2px
    classDef result fill:#fce4ec,stroke:#c2185b,stroke-width:2px
    
    class Start,End startEnd
    class CalcLanguage,CalcInterests,CalcTime,CalcPersonality,WeightedSum process
    class CheckReciprocal,CheckThreshold decision
    class NoMatch,GoodMatch,PoorMatch result
```

### Процесс кэширования

```mermaid
flowchart TD
    Request([Запрос данных]) --> CheckCache{Данные в кэше?}
    
    CheckCache -->|Да| CheckTTL{TTL истек?}
    CheckCache -->|Нет| QueryDB[Запрос к БД]
    
    CheckTTL -->|Нет| ReturnCache[Вернуть из кэша]
    CheckTTL -->|Да| QueryDB
    
    QueryDB --> GetData[Получить данные]
    GetData --> UpdateCache[Обновить кэш]
    UpdateCache --> ReturnData[Вернуть данные]
    
    ReturnCache --> End([Конец])
    ReturnData --> End
    
    %% Стили
    classDef startEnd fill:#e8f5e8,stroke:#2e7d32,stroke-width:2px
    classDef process fill:#e3f2fd,stroke:#1565c0,stroke-width:2px
    classDef decision fill:#fff3e0,stroke:#ef6c00,stroke-width:2px
    classDef cache fill:#f3e5f5,stroke:#7b1fa2,stroke-width:2px
    
    class Request,End startEnd
    class QueryDB,GetData,UpdateCache,ReturnData process
    class CheckCache,CheckTTL decision
    class ReturnCache cache
```

## 🚀 Диаграммы развертывания

### Docker Compose архитектура

```mermaid
graph TB
    subgraph "🐳 Docker Environment"
        subgraph "📦 Services"
            Bot[🤖 Bot Service<br/>Port: 8080]
            Profile[👤 Profile Service<br/>Port: 8081]
            Matcher[🎯 Matcher Service<br/>Port: 8082]
        end
        
        subgraph "💾 Data Services"
            PostgreSQL[(🗄️ PostgreSQL<br/>Port: 5432)]
            Redis[(🔴 Redis<br/>Port: 6379)]
        end
        
        subgraph "🔧 Admin Tools"
            PgAdmin[🔧 PgAdmin<br/>Port: 8080]
        end
        
        subgraph "📊 Monitoring"
            Prometheus[📈 Prometheus<br/>Port: 9090]
            Grafana[📊 Grafana<br/>Port: 3000]
        end
    end
    
    subgraph "🌐 External"
        Telegram[📱 Telegram API]
        Internet[🌍 Internet]
    end
    
    %% Связи
    Bot --> PostgreSQL
    Bot --> Redis
    Profile --> PostgreSQL
    Profile --> Redis
    Matcher --> PostgreSQL
    Matcher --> Redis
    
    PgAdmin --> PostgreSQL
    Prometheus --> Bot
    Prometheus --> Profile
    Prometheus --> Matcher
    Grafana --> Prometheus
    
    Bot --> Telegram
    Internet --> PgAdmin
    Internet --> Grafana
```

### CI/CD Pipeline

```mermaid
flowchart LR
    subgraph "🔄 CI/CD Pipeline"
        Commit[📝 Git Commit] --> Build[🔨 Build]
        Build --> Test[🧪 Run Tests]
        Test --> Security[🔒 Security Scan]
        Security --> Quality[📊 Code Quality]
        Quality --> Package[📦 Package]
        Package --> Deploy[🚀 Deploy]
    end
    
    subgraph "🏗️ Build Stage"
        Build --> DockerBuild[🐳 Docker Build]
        DockerBuild --> PushRegistry[📤 Push to Registry]
    end
    
    subgraph "🧪 Test Stage"
        Test --> UnitTests[Unit Tests]
        Test --> IntegrationTests[Integration Tests]
        Test --> E2ETests[E2E Tests]
    end
    
    subgraph "🚀 Deploy Stage"
        Deploy --> Staging[🧪 Staging]
        Staging --> Production[🏭 Production]
    end
    
    %% Стили
    classDef stage fill:#e3f2fd,stroke:#1565c0,stroke-width:2px
    classDef process fill:#f3e5f5,stroke:#7b1fa2,stroke-width:2px
    classDef environment fill:#e8f5e8,stroke:#2e7d32,stroke-width:2px
    
    class Build,Test,Security,Quality,Package,Deploy stage
    class DockerBuild,PushRegistry,UnitTests,IntegrationTests,E2ETests process
    class Staging,Production environment
```

## 📈 Диаграммы производительности

### Архитектура кэширования

```mermaid
graph TB
    subgraph "📱 Application Layer"
        Bot[🤖 Bot Service]
        Profile[👤 Profile Service]
        Matcher[🎯 Matcher Service]
    end
    
    subgraph "💾 Cache Layer"
        L1Cache[⚡ L1: In-Memory Cache<br/>TTL: 5 minutes]
        L2Cache[🔴 L2: Redis Cache<br/>TTL: 30 minutes]
        L3Cache[🗄️ L3: Database<br/>Persistent]
    end
    
    subgraph "📊 Cache Strategy"
        WriteThrough[📝 Write-Through]
        WriteBehind[⏰ Write-Behind]
        CacheAside[🔄 Cache-Aside]
    end
    
    Bot --> L1Cache
    Profile --> L1Cache
    Matcher --> L1Cache
    
    L1Cache -->|Cache Miss| L2Cache
    L2Cache -->|Cache Miss| L3Cache
    
    L3Cache -->|Update| L2Cache
    L2Cache -->|Update| L1Cache
    
    WriteThrough --> L1Cache
    WriteBehind --> L2Cache
    CacheAside --> L3Cache
```

### Мониторинг производительности

```mermaid
graph LR
    subgraph "📊 Metrics Collection"
        AppMetrics[📈 Application Metrics]
        SystemMetrics[💻 System Metrics]
        BusinessMetrics[📊 Business Metrics]
    end
    
    subgraph "📈 Processing"
        Prometheus[📈 Prometheus]
        AlertManager[🚨 Alert Manager]
    end
    
    subgraph "📊 Visualization"
        Grafana[📊 Grafana Dashboards]
        Alerts[🔔 Alerts & Notifications]
    end
    
    AppMetrics --> Prometheus
    SystemMetrics --> Prometheus
    BusinessMetrics --> Prometheus
    
    Prometheus --> AlertManager
    Prometheus --> Grafana
    
    AlertManager --> Alerts
    
    %% Метрики
    AppMetrics -.->|Response Time| Prometheus
    AppMetrics -.->|Error Rate| Prometheus
    AppMetrics -.->|Throughput| Prometheus
    
    SystemMetrics -.->|CPU Usage| Prometheus
    SystemMetrics -.->|Memory Usage| Prometheus
    SystemMetrics -.->|Disk I/O| Prometheus
    
    BusinessMetrics -.->|Active Users| Prometheus
    BusinessMetrics -.->|Matches Created| Prometheus
    BusinessMetrics -.->|Profile Completions| Prometheus
```

---

**Примечание**: Все диаграммы созданы с использованием Mermaid и могут быть отображены в GitHub, GitLab и других поддерживающих Mermaid платформах.

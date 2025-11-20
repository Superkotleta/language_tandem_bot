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
- **Новые возможности**: Расширенный профиль с детальной информацией о пользователе, временной доступности и предпочтениях общения

## 🎯 Архитектура новой системы интересов

### Структура данных

```mermaid
erDiagram
    INTEREST_CATEGORIES {
        int id PK
        string key_name UK
        int display_order
        timestamp created_at
    }
    
    INTERESTS {
        int id PK
        string key_name UK
        int category_id FK
        int display_order
        string type
        timestamp created_at
    }
    
    USER_INTEREST_SELECTIONS {
        int id PK
        int user_id FK
        int interest_id FK
        boolean is_primary
        int selection_order
        timestamp created_at
    }
    
    INTEREST_LIMITS_CONFIG {
        int id PK
        int min_primary_interests
        int max_primary_interests
        decimal primary_percentage
        timestamp created_at
        timestamp updated_at
    }
    
    MATCHING_CONFIG {
        int id PK
        string config_key UK
        string config_value
        timestamp created_at
        timestamp updated_at
    }
    
    USERS {
        int id PK
        bigint telegram_id UK
        string first_name
        string last_name
        string username
        string interface_language_code
        string state
        string status
        timestamp created_at
        timestamp updated_at
    }
    
    INTEREST_CATEGORIES ||--o{ INTERESTS : "contains"
    USERS ||--o{ USER_INTEREST_SELECTIONS : "selects"
    INTERESTS ||--o{ USER_INTEREST_SELECTIONS : "selected_in"
```

### Поток обработки интересов

```mermaid
sequenceDiagram
    participant U as 👤 User
    participant BOT as 🤖 Bot
    participant HANDLER as 🎯 Interest Handler
    participant SERVICE as ⚙️ Interest Service
    participant DB as 🗄️ Database
    participant CACHE as ⚡ Cache
    
    U->>BOT: Select Interest Category
    BOT->>HANDLER: HandleInterestCategorySelection
    HANDLER->>SERVICE: GetInterestCategories
    SERVICE->>CACHE: Check Cache
    alt Cache Hit
        CACHE-->>SERVICE: Return Cached Categories
    else Cache Miss
        SERVICE->>DB: Query Categories
        DB-->>SERVICE: Return Categories
        SERVICE->>CACHE: Store in Cache
    end
    SERVICE-->>HANDLER: Return Categories
    HANDLER->>BOT: Send Category Interests
    BOT->>U: Display Interests
    
    U->>BOT: Select Interest
    BOT->>HANDLER: HandleInterestSelection
    HANDLER->>SERVICE: ToggleInterestSelection
    SERVICE->>DB: Update Selection
    DB-->>SERVICE: Confirm Update
    SERVICE-->>HANDLER: Return Success
    HANDLER->>BOT: Update Keyboard
    BOT->>U: Show Updated Selection
```

### Компоненты системы

#### 🎯 InterestService

```go
type InterestService struct {
    db     *sql.DB
    config *InterestsConfig
}

// Основные методы
func (s *InterestService) GetInterestCategories() ([]InterestCategory, error)
func (s *InterestService) GetInterestsByCategory(categoryID int) ([]Interest, error)
func (s *InterestService) GetUserInterestSelections(userID int) ([]InterestSelection, error)
func (s *InterestService) AddUserInterestSelection(userID, interestID int, isPrimary bool) error
func (s *InterestService) RemoveUserInterestSelection(userID, interestID int) error
func (s *InterestService) SetPrimaryInterest(userID, interestID int, isPrimary bool) error
func (s *InterestService) GetUserInterestSummary(userID int) (*UserInterestSummary, error)
func (s *InterestService) CalculateCompatibilityScore(user1ID, user2ID int) (int, error)
```

#### 🔧 ProfileInterestHandler

```go
type ProfileInterestHandler struct {
    service         *BotService
    interestService *InterestService
    bot             *BotAPI
    keyboardBuilder *KeyboardBuilder
    errorHandler    *ErrorHandler
}

// Методы для редактирования из профиля
func (h *ProfileInterestHandler) HandleEditInterestsFromProfile(callback *CallbackQuery, user *User) error
func (h *ProfileInterestHandler) HandleEditInterestCategoryFromProfile(callback *CallbackQuery, user *User, categoryKey string) error
func (h *ProfileInterestHandler) HandleEditInterestSelectionFromProfile(callback *CallbackQuery, user *User, interestIDStr string) error
func (h *ProfileInterestHandler) HandleEditPrimaryInterestsFromProfile(callback *CallbackQuery, user *User) error
func (h *ProfileInterestHandler) HandleSaveInterestEditsFromProfile(callback *CallbackQuery, user *User) error
```

#### 💾 TemporaryInterestStorage

```go
type TemporaryInterestStorage struct {
    mu      sync.RWMutex
    storage map[int][]TemporaryInterestSelection
}

// Thread-safe операции
func (s *TemporaryInterestStorage) AddInterest(userID, interestID int, isPrimary bool)
func (s *TemporaryInterestStorage) RemoveInterest(userID, interestID int)
func (s *TemporaryInterestStorage) ToggleInterest(userID, interestID int) bool
func (s *TemporaryInterestStorage) TogglePrimary(userID, interestID int) bool
func (s *TemporaryInterestStorage) SaveToDatabase(userID int, interestService *InterestService) error
```

### Конфигурация системы

#### ⚙️ interests.json

```json
{
  "matching": {
    "primary_interest_score": 3,
    "additional_interest_score": 1,
    "min_compatibility_score": 5,
    "max_matches_per_user": 10
  },
  "interest_limits": {
    "min_primary_interests": 1,
    "max_primary_interests": 5,
    "primary_percentage": 0.3
  },
  "categories": {
    "entertainment": { "display_order": 1, "max_primary_per_category": 2 },
    "education": { "display_order": 2, "max_primary_per_category": 2 },
    "active": { "display_order": 3, "max_primary_per_category": 2 },
    "creative": { "display_order": 4, "max_primary_per_category": 2 },
    "social": { "display_order": 5, "max_primary_per_category": 2 }
  }
}
```

### Алгоритм совместимости

```mermaid
flowchart TD
    START[🎯 Start Matching] --> GET_USER1[👤 Get User 1 Interests]
    GET_USER1 --> GET_USER2[👤 Get User 2 Interests]
    GET_USER2 --> CALC_PRIMARY[⭐ Calculate Primary Score]
    CALC_PRIMARY --> CALC_ADDITIONAL[➕ Calculate Additional Score]
    CALC_ADDITIONAL --> TOTAL_SCORE[📊 Total Compatibility Score]
    TOTAL_SCORE --> CHECK_MIN{🔍 Score >= Min?}
    CHECK_MIN -->|Yes| MATCH[✅ Compatible Match]
    CHECK_MIN -->|No| NO_MATCH[❌ No Match]
    
    subgraph "Scoring Algorithm"
        PRIMARY[⭐ Primary Interests<br/>Score: 3 points each]
        ADDITIONAL[➕ Additional Interests<br/>Score: 1 point each]
        CONFIG[⚙️ Configurable Weights<br/>From interests.json]
    end
    
    CALC_PRIMARY --> PRIMARY
    CALC_ADDITIONAL --> ADDITIONAL
    PRIMARY --> CONFIG
    ADDITIONAL --> CONFIG
```

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

## 🎯 Подробная схема работы системы интересов

### Архитектура компонентов системы интересов

```mermaid
graph TB
    subgraph "👤 Пользовательский интерфейс"
        USER[👤 User]
        TG[📱 Telegram Bot]
    end
    
    subgraph "🎯 Система интересов"
        subgraph "Обработчики"
            PROFILE_H[🔧 ProfileInterestHandler<br/>Редактирование из профиля]
            NEW_H[🆕 NewInterestHandler<br/>Новая система выбора]
            IMPROVED_H[⚡ ImprovedInterestHandler<br/>Улучшенный UX с временным хранением]
        end
        
        subgraph "Сервисы"
            INTEREST_S[⚙️ InterestService<br/>Бизнес-логика интересов]
            TEMP_STORAGE[💾 TemporaryInterestStorage<br/>Временное хранение]
        end
        
        subgraph "Клавиатуры"
            CATEGORY_KB[📂 CreateInterestCategoriesKeyboard<br/>Выбор категорий]
            INTEREST_KB[🎯 CreateCategoryInterestsKeyboard<br/>Выбор интересов в категории]
            PRIMARY_KB[⭐ CreatePrimaryInterestsKeyboard<br/>Выбор основных интересов]
        end
    end
    
    subgraph "🗄️ База данных"
        CATEGORIES_TBL[(📂 interest_categories<br/>Категории интересов)]
        INTERESTS_TBL[(🎯 interests<br/>Интересы с категориями)]
        SELECTIONS_TBL[(⭐ user_interest_selections<br/>Выборы пользователей)]
        LIMITS_TBL[(⚙️ interest_limits_config<br/>Конфигурация лимитов)]
        MATCHING_TBL[(📊 matching_config<br/>Настройки алгоритма)]
    end
    
    subgraph "⚡ Кэширование"
        CACHE[🔄 Cache Interface]
        REDIS[(⚡ Redis Cache)]
        MEMORY[(💾 In-Memory Cache)]
    end
    
    subgraph "📊 Алгоритм совместимости"
        COMPAT[📊 CalculateCompatibilityScore<br/>Расчет баллов совместимости]
        MATCHING[🎯 Matching Algorithm<br/>Подбор партнеров]
    end
    
    USER --> TG
    TG --> PROFILE_H
    TG --> NEW_H
    TG --> IMPROVED_H
    
    PROFILE_H --> INTEREST_S
    NEW_H --> INTEREST_S
    IMPROVED_H --> INTEREST_S
    IMPROVED_H --> TEMP_STORAGE
    
    INTEREST_S --> CATEGORY_KB
    INTEREST_S --> INTEREST_KB
    INTEREST_S --> PRIMARY_KB
    
    INTEREST_S --> CACHE
    CACHE --> REDIS
    CACHE --> MEMORY
    
    INTEREST_S --> CATEGORIES_TBL
    INTEREST_S --> INTERESTS_TBL
    INTEREST_S --> SELECTIONS_TBL
    INTEREST_S --> LIMITS_TBL
    INTEREST_S --> MATCHING_TBL
    
    INTEREST_S --> COMPAT
    COMPAT --> MATCHING
    
    classDef user fill:#FFB6C1,stroke:#333,stroke-width:2px
    classDef handler fill:#90EE90,stroke:#333,stroke-width:2px
    classDef service fill:#87CEEB,stroke:#333,stroke-width:2px
    classDef keyboard fill:#DDA0DD,stroke:#333,stroke-width:2px
    classDef database fill:#F0E68C,stroke:#333,stroke-width:2px
    classDef cache fill:#FFD700,stroke:#333,stroke-width:2px
    classDef algorithm fill:#FFA07A,stroke:#333,stroke-width:2px
    
    class USER,TG user
    class PROFILE_H,NEW_H,IMPROVED_H handler
    class INTEREST_S,TEMP_STORAGE service
    class CATEGORY_KB,INTEREST_KB,PRIMARY_KB keyboard
    class CATEGORIES_TBL,INTERESTS_TBL,SELECTIONS_TBL,LIMITS_TBL,MATCHING_TBL database
    class CACHE,REDIS,MEMORY cache
    class COMPAT,MATCHING algorithm
```

### Детальный поток выбора интересов

```mermaid
sequenceDiagram
    participant U as 👤 User
    participant TG as 📱 Telegram
    participant BOT as 🤖 Bot Service
    participant PROFILE_H as 🔧 ProfileInterestHandler
    participant INTEREST_S as ⚙️ InterestService
    participant CACHE as ⚡ Cache
    participant DB as 🗄️ Database
    participant TEMP as 💾 TemporaryStorage
    
    Note over U,TEMP: 🎯 Процесс выбора интересов из профиля
    
    U->>TG: Нажимает "Редактировать интересы"
    TG->>BOT: Callback: edit_interests_new
    BOT->>PROFILE_H: HandleEditInterestsFromProfile()
    
    PROFILE_H->>INTEREST_S: GetInterestCategories()
    INTEREST_S->>CACHE: Check Cache
    alt Cache Hit
        CACHE-->>INTEREST_S: Return Cached Categories
    else Cache Miss
        INTEREST_S->>DB: SELECT * FROM interest_categories
        DB-->>INTEREST_S: Return Categories
        INTEREST_S->>CACHE: Store in Cache
    end
    INTEREST_S-->>PROFILE_H: Return Categories
    
    PROFILE_H->>BOT: CreateInterestCategoriesKeyboard()
    BOT->>TG: Send Categories Keyboard
    TG->>U: Показать категории интересов
    
    U->>TG: Выбирает категорию "Развлечения"
    TG->>BOT: Callback: edit_interest_category_entertainment
    BOT->>PROFILE_H: HandleEditInterestCategoryFromProfile()
    
    PROFILE_H->>INTEREST_S: GetInterestsByCategoryKey("entertainment")
    INTEREST_S->>CACHE: Check Cache
    alt Cache Hit
        CACHE-->>INTEREST_S: Return Cached Interests
    else Cache Miss
        INTEREST_S->>DB: SELECT * FROM interests WHERE category_id = ?
        DB-->>INTEREST_S: Return Interests
        INTEREST_S->>CACHE: Store in Cache
    end
    INTEREST_S-->>PROFILE_H: Return Interests
    
    PROFILE_H->>INTEREST_S: GetUserInterestSelections(userID)
    INTEREST_S->>DB: SELECT * FROM user_interest_selections WHERE user_id = ?
    DB-->>INTEREST_S: Return Selections
    INTEREST_S-->>PROFILE_H: Return Selections
    
    PROFILE_H->>BOT: CreateCategoryInterestsKeyboard()
    BOT->>TG: Send Interests Keyboard
    TG->>U: Показать интересы в категории
    
    U->>TG: Выбирает интерес "Фильмы"
    TG->>BOT: Callback: edit_interest_select_entertainment_1
    BOT->>PROFILE_H: HandleEditInterestSelectionFromProfile()
    
    PROFILE_H->>INTEREST_S: ToggleInterestSelection(userID, interestID)
    INTEREST_S->>DB: INSERT/UPDATE/DELETE user_interest_selections
    DB-->>INTEREST_S: Confirm Update
    INTEREST_S-->>PROFILE_H: Return Success
    
    PROFILE_H->>BOT: Update Keyboard
    BOT->>TG: Update Interests Keyboard
    TG->>U: Показать обновленный выбор
    
    Note over U,TEMP: 🔄 Процесс продолжается для других категорий
    
    U->>TG: Нажимает "Выбрать основные интересы"
    TG->>BOT: Callback: edit_primary_interests
    BOT->>PROFILE_H: HandleEditPrimaryInterestsFromProfile()
    
    PROFILE_H->>INTEREST_S: GetUserInterestSelections(userID)
    INTEREST_S->>DB: SELECT * FROM user_interest_selections WHERE user_id = ?
    DB-->>INTEREST_S: Return Selections
    INTEREST_S-->>PROFILE_H: Return Selections
    
    PROFILE_H->>BOT: CreatePrimaryInterestsKeyboard()
    BOT->>TG: Send Primary Interests Keyboard
    TG->>U: Показать выбор основных интересов
    
    U->>TG: Выбирает основной интерес "Фильмы"
    TG->>BOT: Callback: edit_primary_interest_1
    BOT->>PROFILE_H: HandleEditPrimaryInterestSelectionFromProfile()
    
    PROFILE_H->>INTEREST_S: UpdateUserInterestPrimaryStatus(userID, interestID, true)
    INTEREST_S->>DB: UPDATE user_interest_selections SET is_primary = true
    DB-->>INTEREST_S: Confirm Update
    INTEREST_S-->>PROFILE_H: Return Success
    
    PROFILE_H->>BOT: Update Keyboard
    BOT->>TG: Update Primary Interests Keyboard
    TG->>U: Показать обновленный выбор основных интересов
    
    U->>TG: Нажимает "Сохранить изменения"
    TG->>BOT: Callback: save_interest_edits
    BOT->>PROFILE_H: HandleSaveInterestEditsFromProfile()
    
    PROFILE_H->>INTEREST_S: GetUserInterestSummary(userID)
    INTEREST_S->>DB: SELECT с JOIN для получения сводки
    DB-->>INTEREST_S: Return Summary
    INTEREST_S-->>PROFILE_H: Return Summary
    
    PROFILE_H->>BOT: CreateProfileMenuKeyboard()
    BOT->>TG: Send Profile Menu
    TG->>U: Показать обновленный профиль
```

### Алгоритм совместимости и подбора партнеров

```mermaid
flowchart TD
    START[🎯 Начало подбора партнеров] --> GET_USER1[👤 Получить интересы пользователя 1]
    GET_USER1 --> GET_USER2[👤 Получить интересы пользователя 2]
    GET_USER2 --> GET_CONFIG[⚙️ Загрузить конфигурацию алгоритма]
    GET_CONFIG --> CALC_PRIMARY[⭐ Расчет баллов основных интересов]
    CALC_PRIMARY --> CALC_ADDITIONAL[➕ Расчет баллов дополнительных интересов]
    CALC_ADDITIONAL --> TOTAL_SCORE[📊 Общий балл совместимости]
    TOTAL_SCORE --> CHECK_MIN{🔍 Балл >= Минимального порога?}
    CHECK_MIN -->|Да| CHECK_MAX{🔍 Количество совпадений < Максимального?}
    CHECK_MAX -->|Да| MATCH[✅ Совместимые партнеры]
    CHECK_MAX -->|Нет| NO_MATCH[❌ Превышен лимит совпадений]
    CHECK_MIN -->|Нет| NO_MATCH
    
    subgraph "📊 Детальный расчет баллов"
        PRIMARY_SCORE[⭐ Основные интересы<br/>Балл: 3 за совпадение<br/>Настраивается в config]
        ADDITIONAL_SCORE[➕ Дополнительные интересы<br/>Балл: 1 за совпадение<br/>Настраивается в config]
        MIN_THRESHOLD[🔍 Минимальный порог<br/>По умолчанию: 5 баллов<br/>Настраивается в config]
        MAX_MATCHES[🔢 Максимум совпадений<br/>По умолчанию: 10<br/>Настраивается в config]
    end
    
    CALC_PRIMARY --> PRIMARY_SCORE
    CALC_ADDITIONAL --> ADDITIONAL_SCORE
    CHECK_MIN --> MIN_THRESHOLD
    CHECK_MAX --> MAX_MATCHES
    
    classDef start fill:#90EE90,stroke:#333,stroke-width:2px
    classDef process fill:#87CEEB,stroke:#333,stroke-width:2px
    classDef decision fill:#FFD700,stroke:#333,stroke-width:2px
    classDef result fill:#FFA07A,stroke:#333,stroke-width:2px
    classDef config fill:#DDA0DD,stroke:#333,stroke-width:2px
    
    class START start
    class GET_USER1,GET_USER2,GET_CONFIG,CALC_PRIMARY,CALC_ADDITIONAL,TOTAL_SCORE process
    class CHECK_MIN,CHECK_MAX decision
    class MATCH,NO_MATCH result
    class PRIMARY_SCORE,ADDITIONAL_SCORE,MIN_THRESHOLD,MAX_MATCHES config
```

### Временное хранение и улучшенный UX

```mermaid
graph TB
    subgraph "💾 TemporaryInterestStorage"
        TEMP_STORAGE[💾 TemporaryInterestStorage<br/>Thread-safe операции]
        
        subgraph "Операции"
            ADD[➕ AddInterest<br/>Добавить интерес]
            REMOVE[➖ RemoveInterest<br/>Удалить интерес]
            TOGGLE[🔄 ToggleInterest<br/>Переключить выбор]
            TOGGLE_PRIMARY[⭐ TogglePrimary<br/>Переключить основной статус]
            GET_SELECTIONS[📋 GetSelections<br/>Получить выборы]
            SAVE_DB[💾 SaveToDatabase<br/>Сохранить в БД]
        end
        
        subgraph "Thread Safety"
            MUTEX[🔒 sync.RWMutex<br/>Безопасность потоков]
            STORAGE[🗄️ map[int][]TemporaryInterestSelection<br/>Временное хранилище]
        end
    end
    
    subgraph "🔄 Поток данных"
        USER_ACTION[👤 Действие пользователя] --> TEMP_OP[💾 Операция с временным хранилищем]
        TEMP_OP --> UPDATE_UI[🖥️ Обновление интерфейса]
        UPDATE_UI --> USER_CONFIRM{👤 Подтверждение?}
        USER_CONFIRM -->|Да| SAVE_DB
        USER_CONFIRM -->|Нет| CANCEL[❌ Отмена изменений]
        SAVE_DB --> CLEAR_TEMP[🧹 Очистка временного хранилища]
        CANCEL --> CLEAR_TEMP
    end
    
    TEMP_STORAGE --> ADD
    TEMP_STORAGE --> REMOVE
    TEMP_STORAGE --> TOGGLE
    TEMP_STORAGE --> TOGGLE_PRIMARY
    TEMP_STORAGE --> GET_SELECTIONS
    TEMP_STORAGE --> SAVE_DB
    
    ADD --> MUTEX
    REMOVE --> MUTEX
    TOGGLE --> MUTEX
    TOGGLE_PRIMARY --> MUTEX
    GET_SELECTIONS --> MUTEX
    SAVE_DB --> MUTEX
    
    MUTEX --> STORAGE
    
    classDef storage fill:#90EE90,stroke:#333,stroke-width:2px
    classDef operation fill:#87CEEB,stroke:#333,stroke-width:2px
    classDef safety fill:#FFD700,stroke:#333,stroke-width:2px
    classDef flow fill:#DDA0DD,stroke:#333,stroke-width:2px
    classDef decision fill:#FFA07A,stroke:#333,stroke-width:2px
    
    class TEMP_STORAGE storage
    class ADD,REMOVE,TOGGLE,TOGGLE_PRIMARY,GET_SELECTIONS,SAVE_DB operation
    class MUTEX,STORAGE safety
    class USER_ACTION,TEMP_OP,UPDATE_UI,CLEAR_TEMP flow
    class USER_CONFIRM decision
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

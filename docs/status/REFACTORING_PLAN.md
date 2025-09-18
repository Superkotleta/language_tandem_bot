# План рефакторинга Language Exchange Bot

## Описание проекта

Language Exchange Bot - Telegram бот для языкового обмена, написанный на Go. Бот позволяет пользователям создавать профили с языками, интересами и обмениваться отзывами.

## Анализ текущего состояния кода

### Выявленные проблемы

#### 1. Структура файлов

- **handlers.go** (4000+ строк) - слишком большой файл, нарушает принцип единственной ответственности
- Смешение бизнес-логики и UI-кода в одном месте
- Отсутствие четкого разделения по ответственности

#### 2. Дублирование кода

- Многократное повторение логики создания клавиатур
- Дублирующиеся методы в database слое (Save/Update методы)
- Копипаста проверок админ-доступа

#### 3. Архитектурные проблемы

- Хендлеры напрямую работают с базой данных, нарушая инверсию зависимостей
- Нет абстракций для UI-компонентов (клавиатуры, сообщения)
- Отсутствие интерфейсов для основных компонентов

#### 4. Код-кворум проблемы

- Пустые case блоки без обработки
- Неиспользуемые переменные и импорты
- Магические числа и строки

#### 5. Производительность

- Множество индивидуальных запросов к БД без транзакций
- Неэффективная загрузка локализации в память

## Детальный план рефакторинга

### Фазы выполнения

#### Фаза 0: Подготовительная фаза - Интеграционные тесты (КРИТИЧЕСКИ ВАЖНО)

**Цель:** Зафиксировать текущее поведение системы тестами перед началом рефакторинга.

##### 0.1 Структура тестов

Создать отдельную папку для тестов:

```shell
services/bot/
├── tests/
│   ├── integration/
│   │   ├── telegram_bot_test.go      # Основные команды и сценарии
│   │   ├── profile_flow_test.go      # Поток создания профиля
│   │   ├── feedback_flow_test.go     # Система отзывов
│   │   ├── admin_commands_test.go    # Административные команды
│   │   └── localization_test.go      # Проверка локализации
│   ├── fixtures/
│   │   ├── test_users.json           # Тестовые пользователи
│   │   ├── test_messages.json        # Тестовые сообщения
│   │   └── test_callbacks.json       # Тестовые callback'и
│   ├── mocks/
│   │   ├── telegram_api_mock.go      # Мок Telegram API
│   │   └── database_mock.go          # Мок базы данных
│   └── helpers/
│       ├── test_setup.go             # Настройка тестового окружения
│       └── assertions.go             # Кастомные проверки
├── internal/
└── cmd/
```

##### 0.2 Интеграционные тесты основных сценариев

```go
// tests/integration/telegram_bot_test.go
package integration

import (
    "testing"
    "context"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/suite"
)

type TelegramBotSuite struct {
    suite.Suite
    bot     *telegram.TelegramHandler
    mockDB  *mocks.DatabaseMock
    ctx     context.Context
}

func (s *TelegramBotSuite) SetupSuite() {
    // Настройка тестового окружения
    s.ctx = context.Background()
    s.mockDB = mocks.NewDatabaseMock()
    s.bot = setupTestBot(s.mockDB)
}

func (s *TelegramBotSuite) TestStartCommand() {
    // Тест команды /start для нового пользователя
    message := createTestMessage("/start", 12345, "testuser")
    
    err := s.bot.HandleUpdate(createUpdateWithMessage(message))
    
    assert.NoError(s.T(), err)
    // Проверяем, что пользователь создан в БД
    // Проверяем, что отправлено приветственное сообщение
    // Проверяем, что показано главное меню
}

func (s *TelegramBotSuite) TestProfileCreationFlow() {
    // Тест полного потока создания профиля
    user := createTestUser()
    
    // 1. Выбор родного языка
    callback := createTestCallback("native_lang_ru", user.TelegramID)
    err := s.bot.HandleUpdate(createUpdateWithCallback(callback))
    assert.NoError(s.T(), err)
    
    // 2. Выбор изучаемого языка
    callback = createTestCallback("target_lang_en", user.TelegramID)
    err = s.bot.HandleUpdate(createUpdateWithCallback(callback))
    assert.NoError(s.T(), err)
    
    // 3. Выбор интересов
    callback = createTestCallback("interest_movies", user.TelegramID)
    err = s.bot.HandleUpdate(createUpdateWithCallback(callback))
    assert.NoError(s.T(), err)
    
    // Проверяем финальное состояние профиля
    profile := s.mockDB.GetUser(user.TelegramID)
    assert.Equal(s.T(), "ru", profile.NativeLanguageCode)
    assert.Equal(s.T(), "en", profile.TargetLanguageCode)
    assert.Contains(s.T(), profile.Interests, "movies")
}

func TestTelegramBotSuite(t *testing.T) {
    suite.Run(t, new(TelegramBotSuite))
}
```

##### 0.3 Тесты локализации

```go
// tests/integration/localization_test.go
func TestLocalizationConsistency(t *testing.T) {
    languages := []string{"ru", "en", "es", "zh"}
    
    for _, lang := range languages {
        t.Run(fmt.Sprintf("language_%s", lang), func(t *testing.T) {
            localizer := setupLocalizer()
            
            // Проверяем наличие всех ключевых переводов
            requiredKeys := []string{
                "welcome_message",
                "choose_native_language", 
                "choose_target_language",
                "main_menu_title",
                "profile_completed",
            }
            
            for _, key := range requiredKeys {
                translation := localizer.Get(lang, key)
                assert.NotEmpty(t, translation, "Missing translation for key %s in language %s", key, lang)
                assert.NotEqual(t, key, translation, "Translation not found for key %s in language %s", key, lang)
            }
        })
    }
}
```

##### 0.4 Тесты административных функций

```go
// tests/integration/admin_commands_test.go
func TestAdminFeedbackFlow(t *testing.T) {
    adminUser := createAdminUser()
    regularUser := createRegularUser()
    
    // 1. Обычный пользователь отправляет отзыв
    feedbackMessage := createTestMessage("Отличный бот!", regularUser.TelegramID, "user")
    // ... отправляем через feedback flow
    
    // 2. Админ просматривает отзывы
    adminMessage := createTestMessage("/feedback", adminUser.TelegramID, "admin")
    err := bot.HandleUpdate(createUpdateWithMessage(adminMessage))
    assert.NoError(t, err)
    
    // 3. Проверяем, что отзыв отображается
    // 4. Админ обрабатывает отзыв
    callback := createTestCallback("feedback_processed_1", adminUser.TelegramID)
    err = bot.HandleUpdate(createUpdateWithCallback(callback))
    assert.NoError(t, err)
}
```

##### 0.5 Настройка тестового окружения

```go
// tests/helpers/test_setup.go
package helpers

func SetupTestBot(mockDB database.Database) *telegram.TelegramHandler {
    // Создаем мок Telegram API
    mockBot := &mocks.TelegramAPIMock{}
    
    // Создаем тестовый сервис
    service := core.NewBotService(mockDB, mockLocalizer())
    
    // Создаем хендлер с моками
    handler := telegram.NewTelegramHandler(mockBot, service, []int64{123456789})
    
    return handler
}

func LoadTestFixtures() *TestFixtures {
    return &TestFixtures{
        Users:     loadUsersFromJSON("tests/fixtures/test_users.json"),
        Messages:  loadMessagesFromJSON("tests/fixtures/test_messages.json"),
        Callbacks: loadCallbacksFromJSON("tests/fixtures/test_callbacks.json"),
    }
}
```

##### 0.6 Зависимости для тестирования

Добавить в `go.mod`:

```go
require (
    github.com/stretchr/testify v1.8.4
    github.com/golang/mock v1.6.0
    github.com/DATA-DOG/go-sqlmock v1.5.0
)
```

Команды для установки:

```bash
cd services/bot
go get github.com/stretchr/testify/suite
go get github.com/stretchr/testify/assert
go get github.com/golang/mock/gomock
go get github.com/DATA-DOG/go-sqlmock
go install github.com/golang/mock/mockgen@latest
```

Генерация моков:

```bash
//go:generate mockgen -source=internal/database/db.go -destination=tests/mocks/database_mock.go
//go:generate mockgen -source=internal/core/service.go -destination=tests/mocks/service_mock.go
```

##### 0.7 Конфигурация для тестов

```go
// tests/helpers/config.go
func GetTestConfig() *config.Config {
    return &config.Config{
        TelegramToken: "test_token",
        DatabaseURL:   "postgres://test:test@localhost/test_db",
        Debug:         true,
        AdminChatIDs:  []int64{123456789},
        AdminUsernames: []string{"testadmin"},
    }
}
```

**Время выполнения:** 6-8 часов
**Критичность:** ОБЯЗАТЕЛЬНО - без этого рефакторинг опасен

#### Фаза 1: Архитектурные изменения (Высокий приоритет)

**Цель:** Разделить большой handlers.go и установить правильные границы ответственности.

##### 1.1 Разделение handlers.go

Создать отдельные файлы:

```shell
internal/adapters/telegram/
├── handlers/
│   ├── profile_handlers.go      # Редактирование профиля, языки, интересы
│   ├── feedback_handlers.go     # Отзывы: отправка, администрирование
│   ├── menu_handlers.go         # Главное меню, навигация
│   ├── keyboard_helpers.go      # Абстракции для создания клавиатур
│   └── admin_handlers.go        # Админ-функциональность (если потребуется)
├── TelegramHandler.go           # Основной хендлер (интерфейс + оркестрация)
├── keyboards.go                 # Удалить, переместить в keyboard_helpers.go
└── handlers.go                  # Удалить после миграции
```

##### 1.2 Создание интерфейсов

```go
type Handler interface {
    HandleCommand(command string, user *models.User) error
    HandleCallback(callback *tgbotapi.CallbackQuery, user *models.User) error
}

type KeyboardBuilder interface {
    BuildMainMenu(lang string) tgbotapi.InlineKeyboardMarkup
    BuildLanguageSelection(lang string, context string) tgbotapi.InlineKeyboardMarkup
    // и т.д.
}
```

**Время выполнения:** 4-6 часов
**Риск:** Переименование структур может сломать код

#### Фаза 2: Оптимизация базы данных (Средний приоритет)

##### 2.1 Удаление дублированных методов

Объединить:

- `SaveNativeLanguage` → `UpdateUserNativeLanguage`
- `SaveTargetLanguage` → `UpdateUserTargetLanguage`
- Аналогично для других дубликатов

##### 2.2 Оптимизация запросов

- Использовать транзакции для комплексных операций
- Заменить множественные запросы на один с JOIN или CTE
- Кэшировать часто запрашиваемые данные

##### 2.3 Добавление индексов (если требуется)

Анализ запросов и добавление индексов для часто фильтруемых полей.

**Время выполнения:** 2-3 часа
**Риск:** Влияние на производительность (нужно тестировать)

#### Фаза 3: Улучшение сервисов (Средний приоритет)

##### 3.1 Выделение бизнес-логики из хендлеров

Переместить в сервисы:

- Валидацию данных профиля
- Логику обработки состояний пользователей
- Расчеты уровня завершения профиля

##### 3.2 Улучшение локализации

- Загружать локали в память при старте
- Добавить кэширование часто используемых строк
- Убрать хардкод текстов из кода

##### 3.3 Validator service

Создать отдельный сервис для валидации:

```go
type Validator struct {
    localizer *localization.Localizer
}

func (v *Validator) ValidateProfile(user *models.User) []ValidationError
```

**Время выполнения:** 3-4 часа
**Риск:** Изменения API сервисов

#### Фаза 4: Очистка и оптимизация (Низкий приоритет)

##### 4.1 Удаление dead code

- Пустые switch case'ы
- Неиспользуемые переменные и функции
- Старые комментарии и TODO

##### 4.2 Оптимизация импортов

- Убрать неиспользуемые импорты
- Группировать импорты по типам (standard, third-party, internal)

##### 4.3 Добавление документации

- Комментарии к функциям и структурам
- Пояснения сложной бизнес-логики

**Время выполнения:** 1-2 часа
**Риск:** Минимальный

## Порядок выполнения

### День 0: Фаза 0 (Подготовка - ОБЯЗАТЕЛЬНО)

1. Создать структуру папок для тестов (1 час)
2. Написать интеграционные тесты для текущего поведения (4-5 часов)
3. Настроить моки и тестовое окружение (2 часа)
4. Запустить базовые тесты и убедиться, что все работает (1 час)

**Критерий готовности:** Все тесты проходят и покрывают основные сценарии

### День 1: Фаза 1.1 (Разделение handlers.go)

1. Создать базовые файлы хендлеров
2. Разделить обработчики команд
3. Разделить обработчики колбэков  
4. Вынести создание клавиатур
5. **После каждого изменения:** запускать интеграционные тесты

### День 2: Фаза 1.2 + Фаза 2 (Интерфейсы + БД)

1. Создать интерфейсы и абстракции
2. Очистить database слой
3. Провести рефакторинг main.go
4. **После каждого изменения:** запускать интеграционные тесты

### День 3: Фаза 3 (Сервисы)

1. Улучшить сервисный слой
2. Добавить валидацию  
3. Оптимизировать локализацию
4. **После каждого изменения:** запускать интеграционные тесты

### День 4: Фаза 4 (Зачистка)

1. Удалить лишний код
2. Профилирование и тестирование
3. Финальные правки
4. **Финальная проверка:** все интеграционные тесты должны проходить

### День 5: Регрессионное тестирование

1. Полное тестирование всех сценариев
2. Тестирование производительности
3. Проверка совместимости с текущими данными

## Критерии успеха

### Технические метрики

- Сокращение handlers.go с 4000+ до ~100-200 строк
- Уменьшение количества строк кода на 20-30%
- Увеличение покрытия тестами на 15%
- Улучшение средней сложности цикликоматичной сложности

### Качественные показатели

- Легкость добавления новых функций
- Улучшение читаемости кода
- Снижение количества багов при изменениях
- Повышение скорости разработки

## Риски и митингации

### Высокий риск: Перелом архитектуры

- **Митигация:** Создавать коммиты после каждого файла
- **Тестирование:** Запуск интеграционных тестов между этапами

### Средний риск: Регрессионные баги

- **Митигация:** Юнит-тесты для критических функций
- **План В:** Возможность отката к предыдущему состоянию

### Низкий риск: Временные простои

- **Митигация:** Рефакторинг по вечерам/выходным
- **Тестирование:** Deployment на staging окружении

## Преимущества после рефакторинга

1. **Удобство сопровождения** - код станет модульным и понятным
2. **Расширяемость** - легко добавлять новые платформы и функции
3. **Производительность** - оптимизированные запросы и кэширование
4. **Качество** - меньше багов, лучшее тестирование
5. **Командная разработка** - параллельная работа над разными компонентами

## Современные практики оптимизации Go кода

### 1. Миграция на Go 1.22

**Текущая версия:** Bot использует Go 1.21
**Рекомендация:** Обновить до Go 1.22 для:

- Улучшенной производительности датаборов
- Лучшей работы с пулом соединений к БД
- Оптимизации сборки мусора

### 2. Использование Generics

Текущий код использует много повторов для обработки коллекций. Generics решают:

```go
// Вместо копипасты этих функций
func filterUserIDs(users []int64, predicate func(int64)) []int64 { ... }
func filterUsernames(users []string, predicate func(string)) []string { ... }

// Использование generics
func Filter[T any](slice []T, predicate func(T) bool) []T {
    result := make([]T, 0, len(slice))
    for _, item := range slice {
        if predicate(item) {
            result = append(result, item)
        }
    }
    return result
}
```

### 3. Context с таймаутами и отменой

Заменить бесконечные контекстсты на таймаут:

```go
// Вместо ctx := context.Background()
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
```

### 4. Использование sync.Pool для объектов

Для часто создаваемых объектов (буферы, слайсы):

```go
var bufPool = sync.Pool{
    New: func() interface{} { return new(bytes.Buffer) },
}

func processData() {
    buf := bufPool.Get().(*bytes.Buffer)
    defer bufPool.Put(buf.Reset())
    // использовать buf
}
```

### 5. Atomic операции вместо мьютексов

Для простых счетчиков и флагов использовать atomic:

```go
import "sync/atomic"

var counter int64

// Вместо mutex.Lock()
atomic.AddInt64(&counter, 1)
current := atomic.LoadInt64(&counter)
```

### 6. Структурированное логирование

Заменить fmt.Printf на структурированные логи:

```go
import "github.com/rs/zerolog"

logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

// Вместо log.Printf("Error: %v", err)
logger.Error().Err(err).Str("component", "handler").Msg("Failed to process request")
```

### 7. Использование горутин-пулов

Для обработки множества запросов:

```go
// Вместо бесконечных горутин
workerPool := &WorkerPool{
    workers: make(chan func(), 100),
}

for i := 0; i < numWorkers; i++ {
    go workerPool.worker()
}
```

### 8. Кэширование с TTL

```go
// Использовать go-cache или ristretto вместо простых map
cache := ttlcache.New[string, *models.User]().
    WithTTL(5 * time.Minute)

func getCachedUser(id int) (*models.User, error) {
    user, found := cache.Get(strconv.Itoa(id))
    if found {
        return user, nil
    }
    // load from DB
    user, err := loadUserFromDB(id)
    if err == nil {
        cache.Set(strconv.Itoa(id), user)
    }
    return user, err
}
```

### 9. Миграция на Go Fiber для HTTP сервисов

Заменить стандартный HTTP на fibre:

```go
app := fiber.New(fiber.Config{
    ErrorHandler: globalErrorHandler,
})

app.Use(middleware.Logger())
app.Use(middleware.Recover())
app.Use(cors.New())
```

### 10. Unit тестирование с testify и gomock

```go
// Улучшить тесты
func TestBotService_GetWelcomeMessage(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    mockDB := mocks.NewMockDatabase(ctrl)
    service := NewBotService(mockDB)

    // Тесты с моками вместо реальной БД
}
```

### 11. Добавление Prometheus метрик

```go
// Мониторинг производительности
var (
    dbQueriesTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "db_queries_total",
            Help: "Total number of DB queries",
        },
        []string{"method", "status"},
    )
)
```

### 12. Graceful shutdown с помощью errgroup

```go
func runServer() error {
    g, ctx := errgroup.WithContext(context.Background())

    g.Go(func() error { return runHTTP(ctx) })
    g.Go(func() error { return runTelegram(ctx) })

    return g.Wait()
}
```

### 13. Использование slices и maps пакетов

Go 1.21+ имеет новые пакеты:

```go
import (
    "maps"
    "slices"
)

users := slices.Clone(originalSlice)
settings := maps.Clone(originalMap)
```

### 14. Структурные теги для валидации

```go
type User struct {
    ID     int    `json:"id" validate:"required,min=1"`
    Name   string `json:"name" validate:"required,min=2,max=100"`
    Age    int    `json:"age" validate:"min=0,max=150"`
}
```

### 15. CI/CD pipeline оптимизации

- Использовать GitHub Actions с кэшированием
- Мультимодульные сборки
- Security scanning с gosec

### 16. Database оптимизации

- Миграция на pgx/v5 (matcher уже использует)
- Prepared statements pool
- Connection pooling с максимальным количеством

### 17. Dependency injection

```go
type ServiceContainer struct {
    Config     *config.Config
    DB         database.Database
    BotService *BotService
    Localizer  *Localization
}

container := wire.New()
```

### 18. Middleware паттерн для хендлеров

```go
type Middleware func(Handler) Handler

func LoggingMiddleware(logger *zerolog.Logger) Middleware {
    return func(next Handler) Handler {
        return HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            logger.Info().Msg("Request started")
            next.ServeHTTP(w, r)
            logger.Info().Msg("Request finished")
        })
    }
}
```

## Метрики прогресса

### Фаза 0: Подготовительная

- [] Создать структуру папок tests/ с подпапками
- [] Написать интеграционные тесты для /start команды
- [] Написать тесты для полного потока создания профиля
- [] Написать тесты для системы отзывов
- [] Написать тесты для административных команд
- [] Написать тесты локализации (4 языка)
- [] Создать моки для Telegram API и БД
- [] Настроить тестовые фикстуры
- [] Все интеграционные тесты проходят

### Фаза 1: Архитектурные изменения

- [] Разбить handlers.go на 4+ файла
- [] Реализовать интерфейсы хендлеров
- [] Создать Command Pattern для команд
- [] Вынести создание клавиатур в отдельный модуль
- [] Интеграционные тесты проходят после каждого изменения

### Фаза 2: Оптимизация БД

- [] Удалить дублированные методы БД
- [] Добавить батчевые операции с транзакциями
- [] Оптимизировать запросы с JOIN
- [] Интеграционные тесты проходят после изменений БД

### Фаза 3: Улучшение сервисов

- [] Создать Validator сервис
- [] Добавить кэширование локализации
- [] Выделить бизнес-логику из хендлеров
- [] Добавить Request/Response паттерн
- [] Интеграционные тесты проходят после рефакторинга сервисов

### Фаза 4: Очистка и современизация

- [] Удалить dead code (минимум 500 строк)
- [] Обновить Go до 1.22
- [] Добавить структурированное логирование
- [] Реализовать кэширование с TTL
- [] Добавить Prometheus метрики
- [] Миграция на pgx/v5
- [] Добавить graceful shutdown с errgroup

### Финальная проверка

- [] Увеличить покрытие тестами на 15%
- [] Все интеграционные тесты проходят
- [] Производительность не ухудшилась
- [] Совместимость с текущими данными сохранена

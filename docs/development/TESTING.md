# 🧪 Testing Guide

Подробное руководство по тестированию Language Exchange Bot.

## 📋 Обзор тестирования

### Типы тестов

- **Unit Tests** - тестирование отдельных компонентов
- **Integration Tests** - тестирование взаимодействия компонентов
- **E2E Tests** - тестирование полных пользовательских сценариев
- **Performance Tests** - тестирование производительности
- **Load Tests** - тестирование под нагрузкой

### Покрытие тестами

- **Цель**: >70% покрытия кода
- **Критичные компоненты**: >90% покрытия
- **Минимум**: >50% покрытия для новых компонентов

## 🚀 Быстрый старт

### Запуск всех тестов

```bash
# Все тесты
go test ./...

# С покрытием
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Verbose режим
go test ./... -v
```

### Запуск конкретных тестов

```bash
# Конкретный пакет
go test ./internal/cache/... -v

# Конкретный тест
go test ./internal/cache/... -run TestCacheService_Get

# С таймаутом
go test ./... -timeout 30s
```

## 🔧 Unit Testing

### Структура unit тестов

```go
func TestFunctionName(t *testing.T) {
    // Arrange - подготовка данных
    input := "test input"
    expected := "expected output"
    
    // Act - выполнение действия
    result := FunctionName(input)
    
    // Assert - проверка результата
    assert.Equal(t, expected, result)
}
```

### Примеры unit тестов

#### Тест кэша

```go
func TestCacheService_Get(t *testing.T) {
    // Arrange
    cache := cache.NewService(cache.DefaultConfig())
    ctx := context.Background()
    key := "test-key"
    value := "test-value"
    
    // Act
    cache.Set(ctx, key, value, time.Hour)
    result, found := cache.Get(ctx, key)
    
    // Assert
    assert.True(t, found)
    assert.Equal(t, value, result)
}
```

#### Тест с моками

```go
func TestUserService_CreateUser(t *testing.T) {
    // Arrange
    mockDB := &mocks.Database{}
    mockDB.On("CreateUser", mock.Anything).Return(&models.User{ID: 1}, nil)
    
    service := NewUserService(mockDB)
    
    // Act
    user, err := service.CreateUser("test@example.com")
    
    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, user)
    mockDB.AssertExpectations(t)
}
```

### Лучшие практики

#### 1. Именование тестов

```go
// Хорошо
func TestCacheService_Get_ReturnsValue_WhenKeyExists(t *testing.T) {}

// Плохо
func TestCache(t *testing.T) {}
```

#### 2. Структура тестов

```go
func TestFunctionName(t *testing.T) {
    t.Run("should return error when input is invalid", func(t *testing.T) {
        // Тест случая
    })
    
    t.Run("should return success when input is valid", func(t *testing.T) {
        // Тест случая
    })
}
```

#### 3. Тестовые данные

```go
func TestUserValidation(t *testing.T) {
    tests := []struct {
        name     string
        user     *models.User
        expected bool
    }{
        {
            name:     "valid user",
            user:     &models.User{Email: "test@example.com"},
            expected: true,
        },
        {
            name:     "invalid email",
            user:     &models.User{Email: "invalid"},
            expected: false,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := ValidateUser(tt.user)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

## 🔗 Integration Testing

### Настройка тестовой среды

```go
func setupTestDB(t *testing.T) *database.DB {
    // Создаем тестовую базу данных
    db, err := database.NewDB("postgres://test:test@localhost:5432/test_db?sslmode=disable")
    require.NoError(t, err)
    
    // Запускаем миграции
    err = runMigrations(db)
    require.NoError(t, err)
    
    return db
}

func cleanupTestDB(t *testing.T, db *database.DB) {
    // Очищаем тестовые данные
    db.Exec("DELETE FROM users")
    db.Exec("DELETE FROM interests")
    db.Close()
}
```

### Пример integration теста

```go
func TestUserRegistrationFlow(t *testing.T) {
    // Arrange
    db := setupTestDB(t)
    defer cleanupTestDB(t, db)
    
    cache := cache.NewService(cache.DefaultConfig())
    service := NewBotService(db, cache)
    
    // Act
    user, err := service.HandleUserRegistration(
        12345,           // telegramID
        "testuser",      // username
        "Test User",     // firstName
        "en",            // language
    )
    
    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, user)
    assert.Equal(t, int64(12345), user.TelegramID)
    assert.Equal(t, "testuser", user.Username)
}
```

### Тестирование с Docker

```yaml
# docker-compose.test.yml
version: '3.8'
services:
  test-postgres:
    image: postgres:14
    environment:
      POSTGRES_DB: test_db
      POSTGRES_USER: test
      POSTGRES_PASSWORD: test
    ports:
      - "5433:5432"
    
  test-redis:
    image: redis:6-alpine
    ports:
      - "6380:6379"
```

```bash
# Запуск integration тестов с Docker
docker-compose -f docker-compose.test.yml up -d
go test ./tests/integration/... -v
docker-compose -f docker-compose.test.yml down
```

## 🌐 E2E Testing

### Настройка E2E тестов

```go
func TestUserRegistrationE2E(t *testing.T) {
    // Arrange
    bot := setupTestBot(t)
    defer cleanupTestBot(t, bot)
    
    // Act - симулируем полный flow регистрации
    update := createTestUpdate("/start")
    err := bot.HandleUpdate(update)
    require.NoError(t, err)
    
    // Проверяем, что пользователь создан
    user, err := bot.GetUserByTelegramID(12345)
    assert.NoError(t, err)
    assert.NotNil(t, user)
    
    // Проверяем, что отправлено приветственное сообщение
    messages := bot.GetSentMessages()
    assert.Len(t, messages, 1)
    assert.Contains(t, messages[0].Text, "Welcome")
}
```

### Тестирование Telegram Bot

```go
func TestTelegramBot_HandleMessage(t *testing.T) {
    // Arrange
    mockAPI := &mocks.TelegramAPI{}
    bot := NewTelegramBot(mockAPI)
    
    update := tgbotapi.Update{
        Message: &tgbotapi.Message{
            Text: "/start",
            From: &tgbotapi.User{ID: 12345},
        },
    }
    
    // Act
    err := bot.HandleUpdate(update)
    
    // Assert
    assert.NoError(t, err)
    mockAPI.AssertExpectations(t)
}
```

## ⚡ Performance Testing

### Бенчмарки

```go
func BenchmarkCacheGet(b *testing.B) {
    cache := cache.NewService(cache.DefaultConfig())
    ctx := context.Background()
    key := "test-key"
    value := "test-value"
    
    cache.Set(ctx, key, value, time.Hour)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        cache.Get(ctx, key)
    }
}
```

### Нагрузочное тестирование

```go
func TestDatabasePerformance(t *testing.T) {
    db := setupTestDB(t)
    defer cleanupTestDB(t, db)
    
    // Тест производительности
    start := time.Now()
    
    for i := 0; i < 1000; i++ {
        _, err := db.CreateUser(&models.User{
            TelegramID: int64(i),
            Username:   fmt.Sprintf("user%d", i),
        })
        require.NoError(t, err)
    }
    
    duration := time.Since(start)
    assert.Less(t, duration, 5*time.Second)
}
```

## 🛠️ Test Utilities

### Test Helpers

```go
// tests/helpers/test_setup.go
func CreateTestUser(t *testing.T, db *database.DB) *models.User {
    user := &models.User{
        TelegramID: 12345,
        Username:  "testuser",
        FirstName: "Test User",
    }
    
    err := db.CreateUser(user)
    require.NoError(t, err)
    
    return user
}

func CreateTestUpdate(text string) tgbotapi.Update {
    return tgbotapi.Update{
        Message: &tgbotapi.Message{
            Text: text,
            From: &tgbotapi.User{ID: 12345},
        },
    }
}
```

### Test Mocks

```go
// tests/mocks/database_mock.go
type MockDatabase struct {
    mock.Mock
}

func (m *MockDatabase) CreateUser(user *models.User) error {
    args := m.Called(user)
    return args.Error(0)
}

func (m *MockDatabase) GetUserByTelegramID(id int64) (*models.User, error) {
    args := m.Called(id)
    return args.Get(0).(*models.User), args.Error(1)
}
```

## 📊 Test Coverage

### Анализ покрытия

```bash
# Генерация отчета покрытия
go test ./... -coverprofile=coverage.out

# HTML отчет
go tool cover -html=coverage.out -o coverage.html

# Функциональное покрытие
go test ./... -coverprofile=coverage.out -covermode=count
go tool cover -func=coverage.out
```

### Цели покрытия по компонентам

| Компонент | Минимум | Цель |
|-----------|---------|------|
| Models | 90% | 95% |
| Cache | 80% | 90% |
| Database | 70% | 80% |
| Handlers | 60% | 70% |
| Core | 80% | 90% |

## 🚨 Test Troubleshooting

### Частые проблемы

#### 1. Тесты падают из-за race conditions

```bash
# Запуск с race detector
go test ./... -race
```

#### 2. Тесты зависают

```bash
# Запуск с таймаутом
go test ./... -timeout 30s
```

#### 3. Проблемы с базой данных

```bash
# Очистка тестовой базы
go test ./... -cleanup
```

### Отладка тестов

```go
func TestWithDebug(t *testing.T) {
    // Включить verbose режим
    t.Log("Starting test...")
    
    // Проверить состояние
    if testing.Verbose() {
        t.Logf("Debug info: %+v", debugInfo)
    }
    
    // Условное выполнение
    if !testing.Short() {
        // Долгие тесты только в полном режиме
        t.Run("long test", func(t *testing.T) {
            // Тест
        })
    }
}
```

## 📚 Best Practices

### 1. Структура тестов

- Один тест = одна проверка
- Используйте table-driven tests для множественных случаев
- Группируйте связанные тесты с `t.Run()`

### 2. Тестовые данные

- Используйте фикстуры для сложных данных
- Очищайте данные после тестов
- Используйте уникальные данные для каждого теста

### 3. Моки и стабы

- Мокайте только внешние зависимости
- Проверяйте вызовы моков
- Используйте реалистичные данные в моках

### 4. Производительность

- Запускайте быстрые тесты часто
- Отделяйте медленные тесты
- Используйте параллельное выполнение где возможно

---

**Готово! 🎉** Теперь вы знаете, как эффективно тестировать Language Exchange Bot. Для дополнительной информации обратитесь к [Go Testing Documentation](https://golang.org/pkg/testing/).

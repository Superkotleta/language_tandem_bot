# 🚀 Developer Guide - Language Exchange Bot

Добро пожаловать в команду разработки Language Exchange Bot! Этот гайд поможет вам быстро начать работу с проектом.

## 📋 Содержание

- [Требования](#требования)
- [Быстрый старт](#быстрый-старт)
- [Структура проекта](#структура-проекта)
- [Архитектура](#архитектура)
- [Разработка](#разработка)
- [Тестирование](#тестирование)
- [Отладка](#отладка)
- [Code Review](#code-review)

## 🔧 Требования

### Обязательные
- **Go 1.25+** - основной язык разработки
- **Docker & Docker Compose** - для контейнеризации
- **PostgreSQL 14+** - основная база данных
- **Redis 6+** - кэширование
- **Git** - система контроля версий

### Рекомендуемые
- **VS Code** с Go extension
- **Postman** - для тестирования API
- **pgAdmin** - для работы с PostgreSQL
- **Redis Commander** - для работы с Redis

### Опциональные
- **Telegram Bot Token** - для тестирования бота
- **ngrok** - для webhook тестирования

## 🚀 Быстрый старт

### 1. Клонирование репозитория
```bash
git clone https://github.com/your-org/language-tandem-bot.git
cd language-tandem-bot
```

### 2. Настройка окружения
```bash
# Копируем пример конфигурации
cp .env.example .env

# Редактируем конфигурацию
nano .env
```

### 3. Запуск с Docker Compose
```bash
# Запускаем все сервисы
docker-compose up -d

# Проверяем статус
docker-compose ps
```

### 4. Проверка работоспособности
```bash
# Health check
curl http://localhost:8080/healthz

# API документация
open http://localhost:8080/swagger/
```

### 5. Development режим
```bash
# Переходим в директорию бота
cd services/bot

# Устанавливаем зависимости
go mod download

# Запускаем в development режиме
go run cmd/bot/main.go
```

## 🏗️ Структура проекта

```
language-tandem-bot/
├── docs/                          # 📚 Документация
│   ├── adr/                       # Architecture Decision Records
│   ├── api/                       # API документация
│   ├── development/               # Development guides
│   └── operations/                # Operations guides
├── services/                      # 🔧 Микросервисы
│   ├── bot/                       # 🤖 Основной бот (АКТИВЕН)
│   │   ├── cmd/bot/               # Точка входа
│   │   ├── internal/              # Внутренняя логика
│   │   │   ├── adapters/          # Внешние интерфейсы
│   │   │   ├── core/              # Бизнес-логика
│   │   │   ├── database/          # Работа с БД
│   │   │   ├── models/            # Модели данных
│   │   │   ├── cache/             # Кэширование
│   │   │   ├── errors/            # Обработка ошибок
│   │   │   └── server/            # HTTP сервер
│   │   ├── tests/                 # Тесты
│   │   └── locales/               # Переводы
│   ├── matcher/                   # 🎯 Сервис matching (ОТКЛЮЧЕН)
│   └── profile/                   # 👤 Сервис профилей (ОТКЛЮЧЕН)
├── services/deploy/               # 🚀 Инфраструктура
│   ├── docker-compose.yml         # Конфигурация сервисов
│   ├── db-init/                   # SQL скрипты
│   └── migrations/                # Миграции БД
└── .github/                       # GitHub workflows
```

## 🏛️ Архитектура

### Clean Architecture
Проект следует принципам Clean Architecture с четким разделением на слои:

```
┌─────────────────────────────────────────────────┐
│                 Delivery Layer                   │
│  ┌─────────────────────────────────────────────┐ │
│  │           Adapters Layer                     │ │
│  │  ┌─────────────────────────────────────────┐ │ │
│  │  │        Core Business Logic               │ │ │
│  │  │  ┌─────────────────────────────────────┐ │ │ │
│  │  │  │      Database/External APIs          │ │ │ │
│  │  │  └─────────────────────────────────────┘ │ │ │
│  │  └─────────────────────────────────────────┘ │ │
│  └─────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────┘
```

### Ключевые принципы:
- **Dependency Inversion**: Внешние слои зависят от внутренних через интерфейсы
- **Single Responsibility**: Каждый компонент имеет одну ответственность
- **Interface Segregation**: Интерфейсы разделены по функциональности

## 💻 Разработка

### Добавление нового endpoint

1. **Создайте handler в `internal/server/`**:
```go
func (s *Server) handleNewEndpoint(w http.ResponseWriter, r *http.Request) {
    // Логика обработки
}
```

2. **Добавьте роут в `server.go`**:
```go
router.HandleFunc("/api/v1/new-endpoint", s.handleNewEndpoint).Methods("GET")
```

3. **Добавьте Swagger документацию**:
```go
// @Summary New endpoint
// @Description Description of the endpoint
// @Tags admin
// @Accept json
// @Produce json
// @Success 200 {object} ResponseModel
// @Router /api/v1/new-endpoint [get]
```

### Добавление новой команды бота

1. **Создайте handler в `internal/adapters/telegram/handlers/`**:
```go
func (h *TelegramHandler) HandleNewCommand(update tgbotapi.Update) error {
    // Логика команды
}
```

2. **Зарегистрируйте команду в `handlers.go`**:
```go
case "/newcommand":
    return h.HandleNewCommand(update)
```

3. **Добавьте локализацию в `locales/`**:
```json
{
  "commands": {
    "newcommand": "New Command Description"
  }
}
```

### Работа с базой данных

1. **Добавьте метод в `internal/database/db.go`**:
```go
func (db *DB) NewMethod() error {
    // SQL запрос
}
```

2. **Добавьте в интерфейс `internal/database/interface.go`**:
```go
type Database interface {
    // ... существующие методы
    NewMethod() error
}
```

3. **Создайте миграцию в `services/deploy/migrations/`**:
```sql
-- Migration: 003_add_new_feature.sql
ALTER TABLE table_name ADD COLUMN new_column VARCHAR(255);
```

## 🧪 Тестирование

### Unit тесты
```bash
# Запуск всех unit тестов
go test ./...

# Запуск с покрытием
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Запуск конкретного пакета
go test ./internal/cache/...
```

### Integration тесты
```bash
# Запуск integration тестов
go test ./tests/integration/... -v

# Запуск с Docker
docker-compose -f docker-compose.test.yml up --abort-on-container-exit
```

### E2E тесты
```bash
# Запуск E2E тестов
go test ./tests/e2e/... -v
```

### Написание тестов

#### Unit тест:
```go
func TestCacheService_Get(t *testing.T) {
    // Arrange
    cache := cache.NewService(cache.DefaultConfig())
    
    // Act
    result, found := cache.Get(context.Background(), "key")
    
    // Assert
    assert.False(t, found)
    assert.Nil(t, result)
}
```

#### Integration тест:
```go
func TestUserRegistration(t *testing.T) {
    // Setup test database
    db := setupTestDB(t)
    defer cleanupTestDB(t, db)
    
    // Test user registration flow
    user, err := db.CreateUser(testUser)
    assert.NoError(t, err)
    assert.NotNil(t, user)
}
```

## 🐛 Отладка

### Логирование
```go
// Используйте структурированное логирование
log.WithFields(log.Fields{
    "user_id": userID,
    "operation": "create_user",
}).Info("User created successfully")
```

### Профилирование
```bash
# CPU профилирование
go run cmd/bot/main.go &
go tool pprof http://localhost:8080/debug/pprof/profile

# Memory профилирование
go tool pprof http://localhost:8080/debug/pprof/heap
```

### Отладка в VS Code
1. Установите Go extension
2. Создайте `.vscode/launch.json`:
```json
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Debug Bot",
            "type": "go",
            "request": "launch",
            "mode": "auto",
            "program": "${workspaceFolder}/services/bot/cmd/bot"
        }
    ]
}
```

### Отладка с Docker
```bash
# Запуск с отладкой
docker-compose -f docker-compose.debug.yml up

# Подключение к контейнеру
docker exec -it language-tandem-bot-bot-1 /bin/bash
```

## 📝 Code Review

### Процесс
1. **Создайте feature branch**:
```bash
git checkout -b feature/new-feature
```

2. **Сделайте изменения и коммиты**:
```bash
git add .
git commit -m "feat: add new feature"
```

3. **Запустите тесты**:
```bash
go test ./...
go vet ./...
golangci-lint run
```

4. **Создайте Pull Request** с описанием изменений

### Стандарты кода

#### Commit Messages
```
type(scope): description

feat(api): add new endpoint for user stats
fix(cache): resolve memory leak in Redis client
docs(readme): update installation instructions
```

#### Code Style
- Используйте `gofmt` для форматирования
- Следуйте Go naming conventions
- Добавляйте комментарии к экспортируемым функциям
- Используйте `golangci-lint` для проверки качества

#### Тестирование
- Покрытие unit тестами > 70%
- Все новые функции должны иметь тесты
- Integration тесты для критических путей

### Checklist для PR
- [ ] Код соответствует стандартам проекта
- [ ] Все тесты проходят
- [ ] Покрытие тестами не снизилось
- [ ] Документация обновлена
- [ ] Логирование добавлено где необходимо
- [ ] Обработка ошибок реализована
- [ ] Performance impact оценен

## 🔗 Полезные ссылки

- [Go Documentation](https://golang.org/doc/)
- [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [Go Testing](https://golang.org/pkg/testing/)
- [Docker Documentation](https://docs.docker.com/)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)

## ❓ FAQ

### Q: Как добавить новый язык интерфейса?
A: Добавьте файл в `services/bot/locales/` и обновите конфигурацию в `internal/localization/`.

### Q: Как настроить webhook для Telegram?
A: Установите `TELEGRAM_MODE=webhook` и `WEBHOOK_URL=https://your-domain.com` в `.env`.

### Q: Как добавить новую команду бота?
A: Создайте handler в `internal/adapters/telegram/handlers/` и зарегистрируйте в `handlers.go`.

### Q: Как отладить проблемы с кэшем?
A: Проверьте логи Redis и используйте `curl http://localhost:8080/api/v1/cache/stats`.

---

**Удачи в разработке! 🚀**

Если у вас есть вопросы, создайте issue в GitHub или обратитесь к команде разработки.

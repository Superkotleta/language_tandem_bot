# Makefile для Language Exchange Bot

.PHONY: help test test-unit test-integration test-coverage build run clean deps

# Переменные
BINARY_NAME=bot
MAIN_PATH=./cmd/bot
TEST_DATABASE_URL=postgres://postgres:password@localhost:5432/language_exchange_test?sslmode=disable

# Помощь
help: ## Показать справку
	@echo "Доступные команды:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Зависимости
deps: ## Установить зависимости
	go mod download
	go mod tidy

# Тестирование
test: test-unit test-integration ## Запустить все тесты

test-unit: ## Запустить unit тесты
	go test -v ./internal/... -race

test-integration: ## Запустить интеграционные тесты
	@echo "Запуск интеграционных тестов..."
	@echo "Убедитесь, что тестовая БД запущена: $(TEST_DATABASE_URL)"
	TEST_DATABASE_URL=$(TEST_DATABASE_URL) go test -v ./tests/integration/... -race

test-coverage: ## Запустить тесты с покрытием
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Отчет о покрытии сохранен в coverage.html"

test-profile-completion: ## Запустить тесты заполнения профиля
	TEST_DATABASE_URL=$(TEST_DATABASE_URL) go test -v ./tests/integration/profile_completion_test.go -run TestProfileCompletionSuite

test-feedback-system: ## Запустить тесты системы отзывов
	TEST_DATABASE_URL=$(TEST_DATABASE_URL) go test -v ./tests/integration/feedback_system_test.go -run TestFeedbackSystemSuite

test-admin-functions: ## Запустить тесты административных функций
	TEST_DATABASE_URL=$(TEST_DATABASE_URL) go test -v ./tests/integration/admin_functions_test.go -run TestAdminFunctionsSuite

test-localization: ## Запустить тесты локализации
	TEST_DATABASE_URL=$(TEST_DATABASE_URL) go test -v ./tests/integration/localization_test.go -run TestLocalizationSuite

# Сборка
build: ## Собрать приложение
	go build -o $(BINARY_NAME) $(MAIN_PATH)

build-optimized: ## Собрать оптимизированную версию
	go build -o $(BINARY_NAME)-optimized ./cmd/optimized

build-linux: ## Собрать для Linux
	GOOS=linux GOARCH=amd64 go build -o $(BINARY_NAME)-linux $(MAIN_PATH)

build-windows: ## Собрать для Windows
	GOOS=windows GOARCH=amd64 go build -o $(BINARY_NAME).exe $(MAIN_PATH)

# Запуск
run: ## Запустить приложение
	go run $(MAIN_PATH)

run-optimized: ## Запустить оптимизированную версию
	go run ./cmd/optimized

run-dev: ## Запустить в режиме разработки
	DEBUG=true go run $(MAIN_PATH)

# Очистка
clean: ## Очистить артефакты сборки
	go clean
	rm -f $(BINARY_NAME) $(BINARY_NAME)-linux $(BINARY_NAME).exe
	rm -f coverage.out coverage.html

# Линтинг
lint: ## Запустить линтер
	golangci-lint run

lint-fix: ## Исправить ошибки линтера
	golangci-lint run --fix

# Форматирование
fmt: ## Форматировать код
	go fmt ./...

# Проверка
check: fmt lint test ## Выполнить все проверки

# Docker
docker-build: ## Собрать Docker образ
	docker build -t language-exchange-bot .

docker-run: ## Запустить в Docker
	docker run --rm -p 8080:8080 language-exchange-bot

# База данных
db-setup: ## Настроить тестовую базу данных
	@echo "Создание тестовой базы данных..."
	createdb language_exchange_test || true
	@echo "Применение миграций..."
	psql $(TEST_DATABASE_URL) -f ../../deploy/db-init/01-init-schemas.sql
	psql $(TEST_DATABASE_URL) -f ../../deploy/db-init/02-init-languages.sql
	psql $(TEST_DATABASE_URL) -f ../../deploy/db-init/03-init-users.sql
	psql $(TEST_DATABASE_URL) -f ../../deploy/db-init/04-init-user-language-pairs.sql
	psql $(TEST_DATABASE_URL) -f ../../deploy/db-init/05-init-friendship-preferences.sql
	psql $(TEST_DATABASE_URL) -f ../../deploy/db-init/06-init-user-time-availability.sql
	psql $(TEST_DATABASE_URL) -f ../../deploy/db-init/07-init-interests.sql
	psql $(TEST_DATABASE_URL) -f ../../deploy/db-init/08-init-user-interests.sql
	psql $(TEST_DATABASE_URL) -f ../../deploy/db-init/09-init-user-traits.sql
	psql $(TEST_DATABASE_URL) -f ../../deploy/db-init/10-init-match-queue.sql
	psql $(TEST_DATABASE_URL) -f ../../deploy/db-init/11-init-tasks.sql
	psql $(TEST_DATABASE_URL) -f ../../deploy/db-init/12-init-localization.sql
	psql $(TEST_DATABASE_URL) -f ../../deploy/db-init/13-add-language-level.sql
	psql $(TEST_DATABASE_URL) -f ../../deploy/db-init/14-init-feedback.sql
	psql $(TEST_DATABASE_URL) -f ../../deploy/db-init/99-seed-data.sql

db-clean: ## Очистить тестовую базу данных
	@echo "Очистка тестовой базы данных..."
	dropdb language_exchange_test || true

# Бенчмарки
bench: ## Запустить бенчмарки
	go test -bench=. -benchmem ./...

# Профилирование
profile: ## Запустить с профилированием
	go run $(MAIN_PATH) -cpuprofile=cpu.prof -memprofile=mem.prof

# Генерация моков
generate-mocks: ## Генерировать моки
	go generate ./...

# Установка инструментов
install-tools: ## Установить инструменты разработки
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/golang/mock/mockgen@latest

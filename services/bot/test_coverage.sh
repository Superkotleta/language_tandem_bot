#!/bin/bash

echo "🧪 Запуск тестов покрытия обработки ошибок и логирования"

# Тестирование системы ошибок
echo "📊 Тестирование системы ошибок..."
go test -coverprofile=error_coverage.out ./internal/errors/... -v
error_coverage=$(go tool cover -func=error_coverage.out | grep total | awk '{print $3}')
echo "Покрытие системы ошибок: $error_coverage"

# Тестирование системы логирования
echo "📊 Тестирование системы логирования..."
go test -coverprofile=logging_coverage.out ./internal/logging/... -v
logging_coverage=$(go tool cover -func=logging_coverage.out | grep total | awk '{print $3}')
echo "Покрытие системы логирования: $logging_coverage"

# Тестирование системы валидации
echo "📊 Тестирование системы валидации..."
go test -coverprofile=validation_coverage.out ./internal/validation/... -v
validation_coverage=$(go tool cover -func=validation_coverage.out | grep total | awk '{print $3}')
echo "Покрытие системы валидации: $validation_coverage"

# Тестирование обработчиков
echo "📊 Тестирование обработчиков..."
go test -coverprofile=handlers_coverage.out ./internal/adapters/telegram/handlers/... -v
handlers_coverage=$(go tool cover -func=handlers_coverage.out | grep total | awk '{print $3}')
echo "Покрытие обработчиков: $handlers_coverage"

# Общее покрытие
echo "📊 Общее покрытие:"
go test -coverprofile=total_coverage.out ./...
total_coverage=$(go tool cover -func=total_coverage.out | grep total | awk '{print $3}')
echo "Общее покрытие: $total_coverage"

# Генерация HTML отчетов
echo "📊 Генерация HTML отчетов..."
go tool cover -html=error_coverage.out -o error_coverage.html
go tool cover -html=logging_coverage.out -o logging_coverage.html
go tool cover -html=validation_coverage.out -o validation_coverage.html
go tool cover -html=handlers_coverage.out -o handlers_coverage.html
go tool cover -html=total_coverage.out -o total_coverage.html

echo "✅ Тестирование завершено. HTML отчеты созданы."

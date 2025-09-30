#!/bin/bash

# Скрипт для запуска линтера
# Использование: ./lint.sh [сервис]
# Примеры:
#   ./lint.sh          - проверить все сервисы
#   ./lint.sh bot      - проверить только bot сервис
#   ./lint.sh matcher  - проверить только matcher сервис
#   ./lint.sh profile  - проверить только profile сервис

set -e

# Цвета для вывода
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Путь к golangci-lint
GOLANGCI_LINT="./golangci-lint"
LINT_CONFIG=".golangci-compatible.yml"

# Функция для вывода сообщений
print_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Проверка наличия golangci-lint
if [ ! -f "$GOLANGCI_LINT" ]; then
    print_error "golangci-lint не найден по пути: $GOLANGCI_LINT"
    print_info "Установите golangci-lint: go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.61.0"
    exit 1
fi

# Проверка наличия конфигурации
if [ ! -f "$LINT_CONFIG" ]; then
    print_error "Конфигурация линтера не найдена: $LINT_CONFIG"
    exit 1
fi

# Функция для запуска линтера на сервисе
lint_service() {
    local service=$1
    local service_path="services/$service"
    
    if [ ! -d "$service_path" ]; then
        print_error "Сервис $service не найден в $service_path"
        return 1
    fi
    
    print_info "Запуск линтера на сервисе: $service"
    
    # Находим все Go файлы в сервисе
    local go_files=$(find "$service_path" -name "*.go" -not -path "*/vendor/*" -not -path "*/tests/*" | head -20)
    
    if [ -z "$go_files" ]; then
        print_warning "Go файлы не найдены в $service_path"
        return 0
    fi
    
    # Запускаем линтер на найденных файлах
    echo "$go_files" | xargs $GOLANGCI_LINT run --config="$LINT_CONFIG"
}

# Основная логика
case "${1:-all}" in
    "bot")
        lint_service "bot"
        ;;
    "matcher")
        lint_service "matcher"
        ;;
    "profile")
        lint_service "profile"
        ;;
    "all"|"")
        print_info "Запуск линтера на всех сервисах..."
        lint_service "bot"
        lint_service "matcher"
        lint_service "profile"
        print_info "Линтер завершен для всех сервисов"
        ;;
    *)
        print_error "Неизвестный сервис: $1"
        print_info "Доступные сервисы: bot, matcher, profile"
        print_info "Использование: $0 [сервис]"
        exit 1
        ;;
esac

print_info "Линтер завершен успешно!"

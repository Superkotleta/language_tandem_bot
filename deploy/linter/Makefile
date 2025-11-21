# Makefile для линтера Language Exchange Bot V2

# Переменные
PROJECT_ROOT = ../..
LINT_CONFIG_GENTLE = .golangci-gentle.yml
LINT_CONFIG_FAST = .golangci-fast.yml
LINT_CONFIG_ENHANCED = .golangci-enhanced.yml
LINT_CONFIG_STRICT = .golangci-strict.yml
GOLANGCI_LINT = $(shell which golangci-lint 2>/dev/null || echo "$(shell go env GOPATH)/bin/golangci-lint")

# Цвета для вывода
GREEN = \033[0;32m
YELLOW = \033[1;33m
RED = \033[0;31m
NC = \033[0m # No Color

.PHONY: help lint lint-gentle lint-fast lint-enhanced lint-strict fmt vet fix clean install

# Показать справку
help:
	@echo "$(GREEN)Доступные команды линтера:$(NC)"
	@echo ""
	@echo "  $(YELLOW)install$(NC)       - Установить golangci-lint (если не установлен)"
	@echo "  $(YELLOW)lint-gentle$(NC)   - Мягкая конфигурация (рекомендуется для начала)"
	@echo "  $(YELLOW)lint-fast$(NC)     - Быстрая конфигурация для разработки"
	@echo "  $(YELLOW)lint-enhanced$(NC) - Улучшенная конфигурация для ежедневной работы"
	@echo "  $(YELLOW)lint-strict$(NC)   - Строгая конфигурация для production"
	@echo "  $(YELLOW)lint$(NC)          - Алиас для lint-gentle"
	@echo ""
	@echo "  $(YELLOW)fmt$(NC)           - Форматирование кода (go fmt)"
	@echo "  $(YELLOW)vet$(NC)           - Запустить go vet"
	@echo "  $(YELLOW)fix$(NC)           - Автоисправление (go fmt + go mod tidy)"
	@echo "  $(YELLOW)clean$(NC)         - Очистить временные файлы"
	@echo ""
	@echo "$(GREEN)Конфигурации:$(NC)"
	@echo "  $(YELLOW)gentle$(NC)   - Мягкая (20 complexity, 150 lines)"
	@echo "  $(YELLOW)fast$(NC)     - Быстрая (timeout 3m)"
	@echo "  $(YELLOW)enhanced$(NC) - Улучшенная (15 complexity, 100 lines)"
	@echo "  $(YELLOW)strict$(NC)   - Строгая (10 complexity, 50 lines)"

# Установка golangci-lint
install:
	@echo "$(GREEN)Проверка установки golangci-lint...$(NC)"
	@if ! command -v golangci-lint >/dev/null 2>&1 && ! [ -x "$(shell go env GOPATH)/bin/golangci-lint" ]; then \
		echo "$(YELLOW)golangci-lint не найден. Устанавливаем...$(NC)"; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
		echo "$(GREEN)✓ golangci-lint установлен в $(shell go env GOPATH)/bin/golangci-lint$(NC)"; \
	else \
		echo "$(GREEN)✓ golangci-lint уже установлен: $(GOLANGCI_LINT)$(NC)"; \
	fi

# Базовая проверка (алиас для gentle)
lint: lint-gentle

# Мягкая конфигурация (рекомендуется)
lint-gentle:
	@echo "$(GREEN)Запуск мягкого линтера...$(NC)"
	@cd $(PROJECT_ROOT) && $(GOLANGCI_LINT) run --config=deploy/linter/$(LINT_CONFIG_GENTLE) ./...
	@echo "$(GREEN)✓ Мягкий линтер завершен!$(NC)"

# Быстрая конфигурация
lint-fast:
	@echo "$(GREEN)Запуск быстрого линтера...$(NC)"
	@cd $(PROJECT_ROOT) && $(GOLANGCI_LINT) run --config=deploy/linter/$(LINT_CONFIG_FAST) ./...
	@echo "$(GREEN)✓ Быстрый линтер завершен!$(NC)"

# Улучшенная конфигурация
lint-enhanced:
	@echo "$(GREEN)Запуск улучшенного линтера...$(NC)"
	@cd $(PROJECT_ROOT) && $(GOLANGCI_LINT) run --config=deploy/linter/$(LINT_CONFIG_ENHANCED) ./...
	@echo "$(GREEN)✓ Улучшенный линтер завершен!$(NC)"

# Строгая конфигурация
lint-strict:
	@echo "$(GREEN)Запуск строгого линтера...$(NC)"
	@cd $(PROJECT_ROOT) && $(GOLANGCI_LINT) run --config=deploy/linter/$(LINT_CONFIG_STRICT) ./...
	@echo "$(GREEN)✓ Строгий линтер завершен!$(NC)"

# Форматирование кода
fmt:
	@echo "$(GREEN)Форматирование кода...$(NC)"
	@cd $(PROJECT_ROOT) && go fmt ./...
	@echo "$(GREEN)✓ Форматирование завершено!$(NC)"

# Go vet
vet:
	@echo "$(GREEN)Запуск go vet...$(NC)"
	@cd $(PROJECT_ROOT) && go vet ./...
	@echo "$(GREEN)✓ go vet завершен!$(NC)"

# Автоисправление
fix:
	@echo "$(GREEN)Автоисправление...$(NC)"
	@cd $(PROJECT_ROOT) && go fmt ./...
	@cd $(PROJECT_ROOT) && go mod tidy
	@cd $(PROJECT_ROOT) && $(GOLANGCI_LINT) run --config=deploy/linter/$(LINT_CONFIG_GENTLE) --fix ./...
	@echo "$(GREEN)✓ Автоисправление завершено!$(NC)"

# Очистка
clean:
	@echo "$(GREEN)Очистка временных файлов...$(NC)"
	@cd $(PROJECT_ROOT) && go clean ./...
	@cd $(PROJECT_ROOT) && rm -rf bin/
	@echo "$(GREEN)✓ Очистка завершена!$(NC)"

# Проверка перед коммитом
pre-commit: fmt vet lint-gentle
	@echo "$(GREEN)✓ Pre-commit проверки пройдены!$(NC)"


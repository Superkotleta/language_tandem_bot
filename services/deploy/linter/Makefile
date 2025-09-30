# Makefile для проекта Language Tandem Bot

# Переменные
GOLANGCI_LINT = ./golangci-lint
LINT_CONFIG = ./.golangci-compatible.yml
ENHANCED_CONFIG = ./.golangci-enhanced.yml
STRICT_CONFIG = ./.golangci-strict.yml
FAST_CONFIG = ./.golangci-fast.yml
GENTLE_CONFIG = ./.golangci-gentle.yml

# Цвета для вывода
GREEN = \033[0;32m
YELLOW = \033[1;33m
RED = \033[0;31m
NC = \033[0m # No Color

.PHONY: help lint lint-bot lint-matcher lint-profile lint-all fmt vet test clean

# Показать справку
help:
	@echo "$(GREEN)Доступные команды:$(NC)"
	@echo "  $(YELLOW)lint$(NC)          - Запустить линтер на всех сервисах (базовая конфигурация)"
	@echo "  $(YELLOW)lint-enhanced$(NC) - Запустить улучшенный линтер на всех сервисах"
	@echo "  $(YELLOW)lint-strict$(NC)  - Запустить строгий линтер на всех сервисах"
	@echo "  $(YELLOW)lint-fast$(NC)    - Запустить быстрый линтер на всех сервисах"
	@echo "  $(YELLOW)lint-gentle$(NC)  - Запустить мягкий линтер на всех сервисах"
	@echo "  $(YELLOW)lint-bot$(NC)      - Запустить линтер только на bot сервисе"
	@echo "  $(YELLOW)lint-matcher$(NC) - Запустить линтер только на matcher сервисе"
	@echo "  $(YELLOW)lint-profile$(NC) - Запустить линтер только на profile сервисе"
	@echo "  $(YELLOW)lint-all$(NC)     - Запустить линтер на всех сервисах с подробным выводом"
	@echo "  $(YELLOW)fmt$(NC)           - Форматировать код (go fmt)"
	@echo "  $(YELLOW)vet$(NC)          - Запустить go vet на всех сервисах"
	@echo "  $(YELLOW)test$(NC)         - Запустить тесты"
	@echo "  $(YELLOW)clean$(NC)        - Очистить временные файлы"
	@echo ""
	@echo "$(GREEN)Конфигурации:$(NC)"
	@echo "  $(YELLOW)compatible$(NC)   - Базовая конфигурация (по умолчанию)"
	@echo "  $(YELLOW)enhanced$(NC)     - Улучшенная конфигурация с современными практиками"
	@echo "  $(YELLOW)strict$(NC)        - Строгая конфигурация для production"
	@echo "  $(YELLOW)fast$(NC)          - Быстрая конфигурация для разработки"
	@echo "  $(YELLOW)gentle$(NC)        - Мягкая конфигурация для постепенного внедрения"

# Линтер для всех сервисов (базовая конфигурация)
lint:
	@echo "$(GREEN)Запуск линтера на всех сервисах (базовая конфигурация)...$(NC)"
	@$(MAKE) lint-bot
	@$(MAKE) lint-matcher  
	@$(MAKE) lint-profile

# Улучшенный линтер для всех сервисов
lint-enhanced:
	@echo "$(GREEN)Запуск улучшенного линтера на всех сервисах...$(NC)"
	@$(GOLANGCI_LINT) run --config=$(ENHANCED_CONFIG) ../../bot/internal/errors/ ../../bot/internal/localization/ ../../bot/internal/adapters/ ../../bot/internal/cache/ ../../bot/internal/config/ ../../bot/internal/core/ ../../bot/internal/database/ ../../bot/internal/models/ ../../bot/internal/validation/
	@$(GOLANGCI_LINT) run --config=$(ENHANCED_CONFIG) ../../matcher/internal/
	@$(GOLANGCI_LINT) run --config=$(ENHANCED_CONFIG) ../../profile/internal/

# Строгий линтер для всех сервисов
lint-strict:
	@echo "$(GREEN)Запуск строгого линтера на всех сервисах...$(NC)"
	@$(GOLANGCI_LINT) run --config=$(STRICT_CONFIG) ../../bot/internal/errors/ ../../bot/internal/localization/ ../../bot/internal/adapters/ ../../bot/internal/cache/ ../../bot/internal/config/ ../../bot/internal/core/ ../../bot/internal/database/ ../../bot/internal/models/ ../../bot/internal/validation/
	@$(GOLANGCI_LINT) run --config=$(STRICT_CONFIG) ../../matcher/internal/
	@$(GOLANGCI_LINT) run --config=$(STRICT_CONFIG) ../../profile/internal/

# Быстрый линтер для всех сервисов
lint-fast:
	@echo "$(GREEN)Запуск быстрого линтера на всех сервисах...$(NC)"
	@$(GOLANGCI_LINT) run --config=$(FAST_CONFIG) ../../bot/internal/errors/ ../../bot/internal/localization/ ../../bot/internal/adapters/ ../../bot/internal/cache/ ../../bot/internal/config/ ../../bot/internal/core/ ../../bot/internal/database/ ../../bot/internal/models/ ../../bot/internal/validation/
	@$(GOLANGCI_LINT) run --config=$(FAST_CONFIG) ../../matcher/internal/
	@$(GOLANGCI_LINT) run --config=$(FAST_CONFIG) ../../profile/internal/

# Мягкий линтер для всех сервисов
lint-gentle:
	@echo "$(GREEN)Запуск мягкого линтера на всех сервисах...$(NC)"
	@$(GOLANGCI_LINT) run --config=$(GENTLE_CONFIG) ../../bot/internal/errors/ ../../bot/internal/localization/ ../../bot/internal/adapters/ ../../bot/internal/cache/ ../../bot/internal/config/ ../../bot/internal/core/ ../../bot/internal/database/ ../../bot/internal/models/ ../../bot/internal/validation/
	@$(GOLANGCI_LINT) run --config=$(GENTLE_CONFIG) ../../matcher/internal/
	@$(GOLANGCI_LINT) run --config=$(GENTLE_CONFIG) ../../profile/internal/

# Линтер для bot сервиса
lint-bot:
	@echo "$(GREEN)Запуск линтера на bot сервисе...$(NC)"
	@$(GOLANGCI_LINT) run --config=$(LINT_CONFIG) ../../bot/internal/errors/ ../../bot/internal/localization/ ../../bot/internal/adapters/ ../../bot/internal/cache/ ../../bot/internal/config/ ../../bot/internal/core/ ../../bot/internal/database/ ../../bot/internal/models/ ../../bot/internal/validation/

# Линтер для matcher сервиса
lint-matcher:
	@echo "$(GREEN)Запуск линтера на matcher сервисе...$(NC)"
	@$(GOLANGCI_LINT) run --config=$(LINT_CONFIG) ../../matcher/internal/

# Линтер для profile сервиса
lint-profile:
	@echo "$(GREEN)Запуск линтера на profile сервисе...$(NC)"
	@$(GOLANGCI_LINT) run --config=$(LINT_CONFIG) ../../profile/internal/

# Линтер с подробным выводом
lint-all:
	@echo "$(GREEN)Запуск линтера с подробным выводом...$(NC)"
	@$(GOLANGCI_LINT) run --config=$(LINT_CONFIG) --verbose ../../bot/internal/errors/ ../../bot/internal/localization/ ../../bot/internal/adapters/ ../../bot/internal/cache/ ../../bot/internal/config/ ../../bot/internal/core/ ../../bot/internal/database/ ../../bot/internal/models/ ../../bot/internal/validation/
	@$(GOLANGCI_LINT) run --config=$(LINT_CONFIG) --verbose ../../matcher/internal/
	@$(GOLANGCI_LINT) run --config=$(LINT_CONFIG) --verbose ../../profile/internal/

# Форматирование кода
fmt:
	@echo "$(GREEN)Форматирование кода...$(NC)"
	@go fmt ../../bot/...
	@go fmt ../../matcher/...
	@go fmt ../../profile/...

# Go vet
vet:
	@echo "$(GREEN)Запуск go vet...$(NC)"
	@go vet ../../bot/...
	@go vet ../../matcher/...
	@go vet ../../profile/...

# Тесты
test:
	@echo "$(GREEN)Запуск тестов...$(NC)"
	@go test ../../bot/...
	@go test ../../matcher/...
	@go test ../../profile/...

# Очистка
clean:
	@echo "$(GREEN)Очистка временных файлов...$(NC)"
	@go clean ../../bot/...
	@go clean ../../matcher/...
	@go clean ../../profile/...

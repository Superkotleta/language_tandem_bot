# Руководство по использованию линтера

## 🚀 Быстрый старт

### 1. Запуск линтера на всех сервисах

```bash
make lint
# или
./lint.sh
```

### 2. Запуск линтера на конкретном сервисе

```bash
make lint-bot      # только bot сервис
make lint-matcher   # только matcher сервис  
make lint-profile   # только profile сервис

# или через скрипт
./lint.sh bot
./lint.sh matcher
./lint.sh profile
```

## 📋 Доступные команды

### Makefile команды

```bash
make help          # Показать справку
make lint          # Линтер на всех сервисах
make lint-bot      # Линтер только на bot
make lint-matcher  # Линтер только на matcher
make lint-profile  # Линтер только на profile
make lint-all      # Линтер с подробным выводом
make fmt           # Форматирование кода
make vet           # Go vet
make test          # Запуск тестов
make clean         # Очистка временных файлов
```

### Прямые команды golangci-lint

```bash
# Линтер на всех сервисах
./services/deploy/linter/golangci-lint run --config=.golangci-compatible.yml services/bot/internal/ services/matcher/internal/ services/profile/internal/

# Линтер на конкретном сервисе
./services/deploy/linter/golangci-lint run --config=.golangci-compatible.yml services/bot/internal/

# Линтер на конкретной директории
./services/deploy/linter/golangci-lint run --config=.golangci-compatible.yml services/bot/internal/errors/
```

## ⚙️ Конфигурация

### Файл конфигурации: `.golangci-compatible.yml`

```yaml
# Конфигурация golangci-lint
run:
  timeout: 5m
  tests: false

linters:
  enable:
    - unused        # неиспользуемый код
    - gofmt         # форматирование
    - goimports     # импорты
    - govet         # go vet проверки
    - ineffassign   # неэффективные присваивания
    - gosimple      # упрощения кода
    - staticcheck   # статический анализ

linters-settings:
  govet:
    enable:
      - assign
      - atomic
      - bools
      - buildtag
      - errorsas
      - httpresponse
      - loopclosure
      - lostcancel
      - nilfunc
      - printf
      - shadow
      - shift
      - sortslice
      - tests
      - timeformat
      - unusedwrite

issues:
  exclude-rules:
    - linters: [golint]
      text: "should have comment"
```

## 🔧 Установка и настройка

### 1. Установка golangci-lint

Линтер уже установлен в папке `services/deploy/linter/`:

```bash
# Проверка установки
./services/deploy/linter/golangci-lint --version
```

### 2. Структура линтера

```shell
services/deploy/linter/
├── golangci-lint                    # Исполняемый файл линтера
├── .golangci-compatible.yml        # Конфигурация линтера
├── LINTING.md                     # Документация по линтеру
└── lint.sh                        # Скрипт для запуска линтера
```

### 3. Версии Go

- **Текущая версия в проекте**: Go 1.25
- **Версия golangci-lint**: v1.61.0 (совместима с Go 1.22)

## 🐛 Решение проблем

### Проблема: "Go language version is lower than targeted"

**Решение**: Убедитесь, что версия Go в go.mod файлах совместима с версией golangci-lint.

### Проблема: "command not found: golangci-lint"

**Решение**: Установите golangci-lint или используйте полный путь: `/home/konstantin/go/bin/golangci-lint`

### Проблема: "can't load config"

**Решение**: Проверьте синтаксис файла `.golangci-compatible.yml`

## 📊 Примеры вывода

### Успешный запуск

```shell
[INFO] Запуск линтера на сервисе: bot
[INFO] Линтер завершен успешно!
```

### Найденные проблемы

```shell
services/bot/internal/localization/localization.go:40:18: Error return value of `filepath.WalkDir` is not checked (errcheck)
 filepath.WalkDir(localesPath, func(path string, d os.DirEntry, err error) error {
                 ^
```

## 🎯 Рекомендации

1. **Регулярно запускайте линтер** перед коммитом
2. **Исправляйте найденные проблемы** для поддержания качества кода
3. **Используйте `make fmt`** для автоматического форматирования
4. **Запускайте `make vet`** для дополнительных проверок

## 🔗 Полезные ссылки

- [golangci-lint документация](https://golangci-lint.run/)
- [Список линтеров](https://golangci-lint.run/usage/linters/)
- [Конфигурация](https://golangci-lint.run/usage/configuration/)

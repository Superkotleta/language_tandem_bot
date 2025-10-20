# Language Editor Tests Documentation

## Обзор

Тесты для изолированного редактора языков (`IsolatedLanguageEditor`) проверяют функциональность редактирования языковых настроек профиля пользователя.

**Файл тестов:** `services/bot/internal/adapters/telegram/handlers/language_editor_test.go`

---

## Запуск тестов

### Все тесты языкового редактора
```bash
go test -v ./internal/adapters/telegram/handlers -run TestLanguage
```

### С покрытием кода
```bash
go test -cover ./internal/adapters/telegram/handlers -run TestLanguage
```

### Все тесты handlers
```bash
go test ./internal/adapters/telegram/handlers
```

---

## Категории тестов

### 1. **Session Structure Tests** (`TestLanguageEditSessionStructure`)

Проверяют корректность структуры сессии редактирования языков.

**Сценарии:**
- ✅ Создание новой сессии с изменениями
- ✅ Создание пустой сессии без изменений
- ✅ Сессия с множественными изменениями

**Что проверяется:**
- Корректность полей `UserID`, `OriginalNativeLang`, `CurrentNativeLang`
- Массив изменений `Changes`
- Временные метки `SessionStart`, `LastActivity`
- Текущий шаг `CurrentStep`

---

### 2. **Language Change Logic Tests** (`TestLanguageChangeLogic`)

Проверяют бизнес-логику изменения языков.

**Сценарии:**
- ✅ Изменение родного языка с не-русского на русский
- ✅ Изменение родного языка с русского на не-русский
- ✅ Ограничение: изучаемый язык можно менять только если родной = русский
- ✅ Изменение уровня владения языком

**Специфическая логика:**
```
Если native != "ru" → target автоматически = "ru"
Если native = "ru"  → target можно выбрать любой (кроме "ru")
```

---

### 3. **Undo Functionality Tests** (`TestUndoFunctionality`)

Проверяют функциональность отмены изменений.

**Сценарии:**
- ✅ Отмена последнего изменения
- ✅ Отмена при множественных изменениях
- ✅ Невозможность отмены при отсутствии изменений

**Логика:**
- Удаление последнего элемента из `Changes`
- Восстановление предыдущего значения поля
- Хронологический порядок отмены (LIFO)

---

### 4. **Validation Tests** (`TestLanguageValidation`)

Проверяют валидацию языковых данных.

**Что проверяется:**
- ✅ Валидные коды языков (2 символа: `ru`, `en`, `es`, `de`)
- ✅ Валидные уровни владения (`A1`, `A2`, `B1`, `B2`, `C1`, `C2`)
- ✅ Родной и изучаемый языки должны различаться
- ✅ Специфическая логика для русского языка

---

### 5. **Session Lifecycle Tests** (`TestSessionLifecycle`)

Проверяют жизненный цикл сессии.

**Сценарии:**
- ✅ Новая сессия начинается без изменений
- ✅ Отслеживание последней активности
- ✅ Прогрессия между шагами (`main_menu` → `native` → `target` → `level` → `preview`)

---

### 6. **Changes Tracking Tests** (`TestChangesTracking`)

Проверяют отслеживание изменений.

**Что проверяется:**
- ✅ Запись одиночного изменения
- ✅ Запись множественных изменений в хронологическом порядке
- ✅ Подсчет количества изменений

**Структура изменения:**
```go
type LanguageChange struct {
    Field     string      // "native_language", "target_language", "target_level"
    OldValue  interface{} // Предыдущее значение
    NewValue  interface{} // Новое значение
    Timestamp time.Time   // Время изменения
}
```

---

## Результаты тестов

### ✅ Все тесты успешно пройдены

```
=== RUN   TestLanguageEditSessionStructure
--- PASS: TestLanguageEditSessionStructure (0.00s)

=== RUN   TestLanguageChangeLogic
--- PASS: TestLanguageChangeLogic (0.00s)

=== RUN   TestUndoFunctionality
--- PASS: TestUndoFunctionality (0.00s)

=== RUN   TestLanguageValidation
--- PASS: TestLanguageValidation (0.00s)

=== RUN   TestSessionLifecycle
--- PASS: TestSessionLifecycle (0.00s)

=== RUN   TestChangesTracking
--- PASS: TestChangesTracking (0.00s)

PASS
ok  	language-exchange-bot/internal/adapters/telegram/handlers	0.003s
```

---

## Покрытие кода

Текущие тесты покрывают:
- ✅ **Структуры данных**: `LanguageEditSession`, `LanguageChange`
- ✅ **Бизнес-логику**: Правила изменения языков
- ✅ **Валидацию**: Коды языков и уровней
- ✅ **Функциональность**: Undo, отслеживание изменений

**Не покрыто (требует интеграционных тестов):**
- ⚠️ Взаимодействие с Telegram API
- ⚠️ Работа с Redis/Cache для сессий
- ⚠️ Callback handlers и роутинг
- ⚠️ Сохранение в БД

---

## Дальнейшие улучшения

### Рекомендуемые дополнительные тесты:

1. **Integration Tests** - тесты с реальным Redis и БД
2. **Handler Tests** - тесты методов `HandleEditNativeLanguage`, `HandleSaveChanges` и т.д.
3. **Keyboard Tests** - проверка корректности callback data в клавиатурах
4. **Error Handling Tests** - тесты обработки ошибок

### Пример интеграционного теста:
```go
func TestLanguageEditorIntegration(t *testing.T) {
    // Setup: создать mock Redis, БД, Telegram bot
    // Test: полный flow редактирования от начала до конца
    // Verify: проверить что изменения сохранились в БД
}
```

---

## Ручное тестирование

Для полной проверки функциональности рекомендуется провести ручное тестирование:

### Чек-лист:
- [ ] Открыть редактор языков из профиля
- [ ] Изменить родной язык
- [ ] Проверить автоматическое изменение изучаемого языка
- [ ] Изменить уровень владения
- [ ] Проверить предпросмотр изменений
- [ ] Отменить последнее изменение (Undo)
- [ ] Сохранить изменения
- [ ] Отменить редактирование (Cancel)
- [ ] Проверить, что изменения применились в профиле

---

## Заключение

✅ **Все unit тесты успешно проходят**
✅ **Бизнес-логика покрыта тестами**
✅ **Валидация работает корректно**
⚠️ **Требуется ручное тестирование для проверки UI/UX**
⚠️ **Рекомендуется добавить интеграционные тесты**

**Следующий шаг:** Ручное тестирование в реальном боте для проверки взаимодействия с пользователем.

# 🎯 Система доступности пользователя

## 📋 Обзор

Система доступности позволяет пользователям указывать предпочтения по времени общения и стилю коммуникации для более точного подбора партнеров по языковому обмену.

## 🏗️ Архитектура

### 📁 Структура файлов

```shell
services/bot/internal/adapters/telegram/handlers/
├── availability_handlers.go     # Основная логика обработчиков
├── availability_keyboards.go   # Создание клавиатур для UI
└── handlers.go                 # Интеграция в основной роутер

services/bot/internal/core/
└── service.go                  # Форматирование данных в профиле

services/bot/internal/database/
├── interface.go                # Интерфейсы для работы с БД
└── db.go                       # Реализация методов БД

services/bot/internal/models/
└── user.go                     # Модели данных доступности
```

### 🗄️ База данных

#### Таблица `user_time_availability`

```sql
CREATE TABLE user_time_availability (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id) ON DELETE CASCADE,
    day_type TEXT CHECK (day_type IN ('weekdays', 'weekends', 'any', 'specific')),
    specific_days TEXT[] DEFAULT NULL,
    time_slot TEXT CHECK (time_slot IN ('morning', 'day', 'evening', 'late')),
    created_at TIMESTAMP DEFAULT NOW()
);
```

#### Таблица `friendship_preferences`

```sql
CREATE TABLE friendship_preferences (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id) ON DELETE CASCADE,
    activity_type TEXT CHECK (activity_type IN ('movies', 'games', 'casual_chat', 'creative', 'active', 'educational')),
    communication_style TEXT CHECK (communication_style IN ('text', 'voice_msg', 'audio_call', 'video_call', 'meet_person')),
    communication_frequency TEXT CHECK (communication_frequency IN ('spontaneous', 'weekly', 'daily', 'intensive')),
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(user_id)
);
```

## 🔄 Процесс настройки доступности

### Phase 1: Временная доступность

1. **Выбор типа дней**
   - `weekdays` - будние дни
   - `weekends` - выходные дни
   - `any` - любое время
   - `specific` - конкретные дни

2. **Выбор конкретных дней** (если выбран `specific`)
   - Мультивыбор дней недели с чекбоксами
   - Минимум 1 день, максимум 7 дней

3. **Выбор времени дня**
   - `morning` - утро (6:00-12:00)
   - `day` - день (12:00-18:00)
   - `evening` - вечер (18:00-22:00)
   - `late` - поздно (22:00-6:00)

### Phase 2: Предпочтения общения

1. **Тип активности**
   - `movies` - фильмы и сериалы
   - `games` - компьютерные игры
   - `casual_chat` - легкий разговор
   - `creative` - творчество и искусство
   - `active` - активный отдых
   - `educational` - образование и саморазвитие

2. **Стиль общения**
   - `text` - текстовые сообщения
   - `voice_msg` - голосовые сообщения
   - `audio_call` - аудиозвонки
   - `video_call` - видеозвонки
   - `meet_person` - личная встреча

3. **Частота общения**
   - `spontaneous` - спонтанно
   - `weekly` - раз в неделю
   - `daily` - ежедневно
   - `intensive` - интенсивно

## 🎮 Пользовательский интерфейс

### Основной процесс настройки

```shell
🎯 Выбери интересы → ✅ Готово
    ↓
⏰ Настроить доступность → Выбор типа дней
    ↓
📅 Выбор дней → Выбор времени
    ↓
🤝 Предпочтения общения → Выбор активности
    ↓
💬 Стиль общения → Частота общения
    ↓
✅ Завершение настройки
```

### Кнопки в профиле

- **⏰ Редактировать доступность** - переход к редактированию настроек
- **📱 Просмотр профиля** - отображение настроек в читаемом виде

## 🌐 Локализация

### Русский язык (`ru.json`)

```json
{
  "time_availability_intro": "⏰ Настройка доступности\n\nДавайте настроим, когда вы обычно свободны для общения на иностранном языке.",
  "select_specific_days": "📅 Выберите дни недели, когда вы обычно свободны:",
  "select_time_slot": "🕐 Выберите удобное время дня:",
  "friendship_preferences_intro": "🤝 Предпочтения общения\n\nРасскажите о том, как вы предпочитаете общаться с партнерами по языковому обмену.",
  "select_communication_style": "💬 Выберите предпочитаемый способ общения:",
  "select_communication_frequency": "📊 Как часто вы хотите общаться:"
}
```

### Английский язык (`en.json`)

```json
{
  "time_availability_intro": "⏰ Time Availability Setup\n\nLet's set up when you're usually available for language exchange conversations.",
  "select_specific_days": "📅 Select the days of the week when you're usually available:",
  "select_time_slot": "🕐 Choose your preferred time of day:",
  "friendship_preferences_intro": "🤝 Communication Preferences\n\nTell us about how you prefer to communicate with your language exchange partners.",
  "select_communication_style": "💬 Choose your preferred communication method:",
  "select_communication_frequency": "📊 How often do you want to communicate:"
}
```

## 🧪 Тестирование

### Интеграционные тесты

```bash
# Запуск всех тестов доступности
go test ./tests/integration -run TestAvailabilitySystemIntegration -v

# Запуск конкретного теста
go test ./tests/integration -run TestAvailabilitySystemIntegration/TestSaveAndGetTimeAvailability -v
```

### Покрытые сценарии

- ✅ Сохранение и получение временной доступности
- ✅ Сохранение и получение предпочтений общения
- ✅ Выбор конкретных дней недели
- ✅ Значения по умолчанию для новых пользователей
- ✅ Обновление существующих настроек

## 🔧 API методы

### Database Interface

```go
// Временная доступность
SaveTimeAvailability(userID int, availability *TimeAvailability) error
GetTimeAvailability(userID int) (*TimeAvailability, error)

// Предпочтения общения
SaveFriendshipPreferences(userID int, preferences *FriendshipPreferences) error
GetFriendshipPreferences(userID int) (*FriendshipPreferences, error)
```

### Handler Methods

```go
// Запуск настройки
HandleTimeAvailabilityStart(callback *CallbackQuery, user *User) error
HandleFriendshipPreferencesStart(callback *CallbackQuery, user *User) error

// Обработка выбора
HandleDayTypeSelection(callback *CallbackQuery, user *User, dayType string) error
HandleTimeSlotSelection(callback *CallbackQuery, user *User, timeSlot string) error
HandleActivityTypeSelection(callback *CallbackQuery, user *User, activityType string) error
HandleCommunicationStyleSelection(callback *CallbackQuery, user *User, style string) error
HandleCommunicationFrequencySelection(callback *CallbackQuery, user *User, frequency string) error
```

## 📊 Метрики успеха

- **Завершаемость настройки**: >90% пользователей проходят полную настройку
- **Время настройки**: <3 минуты в среднем
- **Удовлетворенность**: >80% пользователей находят подходящих партнеров
- **Точность matching**: Улучшение на 40% по сравнению с базовым алгоритмом

## 🚀 Следующие шаги

1. **Интеграция в алгоритм matching** - использовать данные доступности для подбора партнеров
2. **Уведомления** - отправка уведомлений в подходящее время
3. **Статистика** - аналитика предпочтений пользователей
4. **A/B тестирование** - сравнение эффективности разных подходов к настройке

---

*Система доступности реализована в рамках Phase 2 проекта Language Exchange Bot* 🎯

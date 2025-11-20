-- Инициализация таблиц доступности пользователей
-- Создание таблиц: user_time_availability и friendship_preferences
-- Дата создания: 2025-10-14

-- =============================================================================
-- ТАБЛИЦА ВРЕМЕННОЙ ДОСТУПНОСТИ ПОЛЬЗОВАТЕЛЕЙ
-- =============================================================================

CREATE TABLE IF NOT EXISTS user_time_availability (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id) ON DELETE CASCADE,
    day_type TEXT NOT NULL CHECK (day_type IN ('weekdays', 'weekends', 'any', 'specific')),
    specific_days TEXT[] DEFAULT '{}',
    time_slots TEXT[] DEFAULT '{}',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    -- Уникальный constraint на пользователя
    CONSTRAINT unique_user_availability UNIQUE (user_id),

    -- Проверки на массивы
    CONSTRAINT user_time_availability_specific_days_check
        CHECK (day_type != 'specific' OR array_length(specific_days, 1) > 0),
    CONSTRAINT user_time_availability_time_slots_check
        CHECK (array_length(time_slots, 1) > 0)
);

-- Индексы для производительности
CREATE INDEX IF NOT EXISTS idx_user_time_availability_user_id ON user_time_availability(user_id);
CREATE INDEX IF NOT EXISTS idx_user_time_availability_day_type ON user_time_availability(day_type);
CREATE INDEX IF NOT EXISTS idx_user_time_availability_time_slots ON user_time_availability USING GIN (time_slots);

-- Комментарии к полям
COMMENT ON TABLE user_time_availability IS 'Временная доступность пользователей для языкового обмена';
COMMENT ON COLUMN user_time_availability.day_type IS 'Тип дней: weekdays, weekends, any, specific';
COMMENT ON COLUMN user_time_availability.specific_days IS 'Массив конкретных дней недели при day_type=specific';
COMMENT ON COLUMN user_time_availability.time_slots IS 'Массив временных слотов: morning, day, evening, late';

-- =============================================================================
-- ТАБЛИЦА ПРЕДПОЧТЕНИЙ ПО ОБЩЕНИЮ
-- =============================================================================

-- Удаляем старую таблицу, если существует
DROP TABLE IF EXISTS friendship_preferences;

CREATE TABLE friendship_preferences (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id) ON DELETE CASCADE,
    activity_type TEXT DEFAULT 'casual_chat' CHECK (activity_type IN ('movies', 'games', 'educational', 'casual_chat')),
    communication_styles TEXT[] DEFAULT '{}',
    communication_frequency TEXT DEFAULT 'weekly' CHECK (communication_frequency IN ('multiple_weekly', 'weekly', 'multiple_monthly', 'flexible')),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    -- Уникальный constraint на пользователя
    CONSTRAINT unique_user_friendship_preferences UNIQUE (user_id),

    -- Проверка на массив способов общения
    CONSTRAINT friendship_preferences_communication_styles_check
        CHECK (array_length(communication_styles, 1) > 0)
);

-- Индексы для производительности
CREATE INDEX IF NOT EXISTS idx_friendship_preferences_user_id ON friendship_preferences(user_id);
CREATE INDEX IF NOT EXISTS idx_friendship_preferences_activity_type ON friendship_preferences(activity_type);
CREATE INDEX IF NOT EXISTS idx_friendship_preferences_communication_styles ON friendship_preferences USING GIN (communication_styles);
CREATE INDEX IF NOT EXISTS idx_friendship_preferences_communication_frequency ON friendship_preferences(communication_frequency);

-- Комментарии к полям
COMMENT ON TABLE friendship_preferences IS 'Предпочтения пользователей по общению в языковом обмене';
COMMENT ON COLUMN friendship_preferences.activity_type IS 'Тип активности: movies, games, educational, casual_chat';
COMMENT ON COLUMN friendship_preferences.communication_styles IS 'Массив способов общения: text, voice_msg, audio_call, video_call, meet_person';
COMMENT ON COLUMN friendship_preferences.communication_frequency IS 'Частота общения: multiple_weekly, weekly, multiple_monthly, flexible';

-- =============================================================================
-- ДОБАВЛЕНИЕ ТРИГГЕРОВ ДЛЯ updated_at
-- =============================================================================

-- Триггер для user_time_availability
CREATE OR REPLACE FUNCTION update_user_time_availability_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_update_user_time_availability_updated_at ON user_time_availability;
CREATE TRIGGER trigger_update_user_time_availability_updated_at
    BEFORE UPDATE ON user_time_availability
    FOR EACH ROW EXECUTE FUNCTION update_user_time_availability_updated_at();

-- Триггер для friendship_preferences
CREATE OR REPLACE FUNCTION update_friendship_preferences_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_update_friendship_preferences_updated_at ON friendship_preferences;
CREATE TRIGGER trigger_update_friendship_preferences_updated_at
    BEFORE UPDATE ON friendship_preferences
    FOR EACH ROW EXECUTE FUNCTION update_friendship_preferences_updated_at();

-- =============================================================================
-- НАЧАЛЬНЫЕ ДАННЫЕ (опционально)
-- =============================================================================

-- Примеры начальных данных для тестирования (можно удалить в production)
-- INSERT INTO user_time_availability (user_id, day_type, time_slots)
-- VALUES (1, 'weekdays', ARRAY['morning', 'evening'])
-- ON CONFLICT (user_id) DO NOTHING;

-- INSERT INTO friendship_preferences (user_id, communication_styles, communication_frequency)
-- VALUES (1, ARRAY['text', 'video_call'], 'weekly')
-- ON CONFLICT (user_id) DO NOTHING;

-- Миграция: Обновление таблиц доступности для поддержки мультивыбора
-- Дата создания: 2025-10-14
-- Описание: Изменение структур таблиц user_time_availability и friendship_preferences
-- для поддержки массивов вместо одиночных значений

-- =============================================================================
-- ОБНОВЛЕНИЕ ТАБЛИЦЫ user_time_availability
-- =============================================================================

-- Добавляем новую колонку time_slots как массив
ALTER TABLE user_time_availability
ADD COLUMN IF NOT EXISTS time_slots TEXT[] DEFAULT '{}';

-- Переносим данные из старой колонки в новую (если time_slot существует)
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.columns
               WHERE table_name = 'user_time_availability'
               AND column_name = 'time_slot') THEN

        -- Обновляем существующие записи, конвертируя одиночные значения в массивы
        UPDATE user_time_availability
        SET time_slots = CASE
            WHEN time_slot IS NOT NULL AND time_slot != '' THEN ARRAY[time_slot]
            ELSE ARRAY['any']
        END
        WHERE time_slots = '{}';

        -- Удаляем старую колонку
        ALTER TABLE user_time_availability DROP COLUMN IF EXISTS time_slot;
    END IF;
END $$;

-- Обновляем индексы
DROP INDEX IF EXISTS idx_user_time_availability_time_slot;
CREATE INDEX IF NOT EXISTS idx_user_time_availability_time_slots ON user_time_availability USING GIN (time_slots);

-- =============================================================================
-- ОБНОВЛЕНИЕ ТАБЛИЦЫ friendship_preferences
-- =============================================================================

-- Добавляем новую колонку communication_styles как массив
ALTER TABLE friendship_preferences
ADD COLUMN IF NOT EXISTS communication_styles TEXT[] DEFAULT '{}';

-- Переносим данные из старой колонки в новую (если communication_style существует)
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.columns
               WHERE table_name = 'friendship_preferences'
               AND column_name = 'communication_style') THEN

        -- Обновляем существующие записи, конвертируя одиночные значения в массивы
        UPDATE friendship_preferences
        SET communication_styles = CASE
            WHEN communication_style IS NOT NULL AND communication_style != '' THEN ARRAY[communication_style]
            ELSE ARRAY['text']
        END
        WHERE communication_styles = '{}';

        -- Удаляем старую колонку
        ALTER TABLE friendship_preferences DROP COLUMN IF EXISTS communication_style;
    END IF;
END $$;

-- Обновляем индексы
DROP INDEX IF EXISTS idx_friendship_preferences_communication_style;
CREATE INDEX IF NOT EXISTS idx_friendship_preferences_communication_styles ON friendship_preferences USING GIN (communication_styles);

-- =============================================================================
-- ДОБАВЛЕНИЕ КОНСТРЕЙНТОВ И ПРОВЕРОК
-- =============================================================================

-- Добавляем проверки для массивов (минимум 1 элемент)
ALTER TABLE user_time_availability
DROP CONSTRAINT IF EXISTS user_time_availability_time_slots_check,
ADD CONSTRAINT user_time_availability_time_slots_check
CHECK (array_length(time_slots, 1) > 0);

ALTER TABLE friendship_preferences
DROP CONSTRAINT IF EXISTS friendship_preferences_communication_styles_check,
ADD CONSTRAINT friendship_preferences_communication_styles_check
CHECK (array_length(communication_styles, 1) > 0);

-- =============================================================================
-- ДОБАВЛЕНИЕ КОММЕНТАРИЕВ К ТАБЛИЦАМ
-- =============================================================================

COMMENT ON COLUMN user_time_availability.day_type IS 'Тип дней: weekdays, weekends, any, specific';
COMMENT ON COLUMN user_time_availability.specific_days IS 'Массив конкретных дней недели при day_type=specific';
COMMENT ON COLUMN user_time_availability.time_slots IS 'Массив временных слотов: morning, day, evening, late';

COMMENT ON COLUMN friendship_preferences.activity_type IS 'Тип активности: movies, games, educational, casual_chat';
COMMENT ON COLUMN friendship_preferences.communication_styles IS 'Массив способов общения: text, voice_msg, audio_call, video_call, meet_person';
COMMENT ON COLUMN friendship_preferences.communication_frequency IS 'Частота общения: multiple_weekly, weekly, multiple_monthly, flexible';

-- =============================================================================
-- ПРОВЕРКА МИГРАЦИИ
-- =============================================================================

-- Выводим информацию о структуре таблиц после миграции
DO $$
BEGIN
    RAISE NOTICE 'Миграция завершена успешно!';
    RAISE NOTICE 'Проверьте структуру таблиц:';
    RAISE NOTICE '- user_time_availability.time_slots: массив временных слотов';
    RAISE NOTICE '- friendship_preferences.communication_styles: массив способов общения';
END $$;

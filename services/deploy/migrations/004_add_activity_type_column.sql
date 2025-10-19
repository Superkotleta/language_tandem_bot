-- Миграция: Добавление колонки activity_type в таблицу friendship_preferences
-- Дата создания: 2025-10-19
-- Описание: Добавление недостающей колонки activity_type для типов активности

-- Добавляем колонку activity_type, если она не существует
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                   WHERE table_name = 'friendship_preferences'
                   AND column_name = 'activity_type') THEN

        ALTER TABLE friendship_preferences
        ADD COLUMN activity_type TEXT;

        -- Добавляем constraint для проверки допустимых значений
        ALTER TABLE friendship_preferences
        ADD CONSTRAINT friendship_preferences_activity_type_check
        CHECK (activity_type IN (
            'movies', 'games', 'casual_chat', 'creative', 'active', 'educational'
        ));

        -- Устанавливаем значение по умолчанию для существующих записей
        UPDATE friendship_preferences
        SET activity_type = 'casual_chat'
        WHERE activity_type IS NULL;

        -- Делаем колонку NOT NULL
        ALTER TABLE friendship_preferences
        ALTER COLUMN activity_type SET NOT NULL;

        RAISE NOTICE 'Колонка activity_type успешно добавлена в таблицу friendship_preferences';
    ELSE
        RAISE NOTICE 'Колонка activity_type уже существует в таблице friendship_preferences';
    END IF;
END $$;

-- Добавляем индекс для колонки activity_type
CREATE INDEX IF NOT EXISTS idx_friendship_preferences_activity_type
ON friendship_preferences (activity_type);

-- Добавляем комментарий к колонке
COMMENT ON COLUMN friendship_preferences.activity_type IS 'Тип активности: movies, games, educational, casual_chat';

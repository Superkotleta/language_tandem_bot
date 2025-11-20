-- Миграция: добавление колонки updated_at к существующей таблице user_feedback
-- Выполните этот скрипт один раз в PostgreSQL-контейнере

-- Добавляем колонку если её нет (работаем с текущим пользователем)
DO $$
BEGIN
    -- Проверяем и добавляем колонку
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_column_usage = current_database()
        AND table_name = 'user_feedback'
        AND column_name = 'updated_at'
    ) THEN
        EXECUTE 'ALTER TABLE user_feedback ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP DEFAULT NOW()';
        RAISE NOTICE 'Колонка updated_at добавлена в таблицу user_feedback';
    ELSE
        RAISE NOTICE 'Колонка updated_at уже существует в таблице user_feedback';
    END IF;

    -- Обновляем все записи с NULL значением
    EXECUTE 'UPDATE user_feedback SET updated_at = created_at WHERE updated_at IS NULL';

END $$;

-- Обновляем существуюшие записи (если есть)
UPDATE user_feedback
SET updated_at = created_at
WHERE updated_at IS NULL;

-- Создаем функцию триггера если её нет
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Создаем триггер если его нет
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.triggers
        WHERE event_object_table = 'user_feedback'
        AND trigger_name = 'update_user_feedback_updated_at'
    ) THEN
        CREATE TRIGGER update_user_feedback_updated_at
            BEFORE UPDATE ON user_feedback
            FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
        RAISE NOTICE 'Триггер update_user_feedback_updated_at создан';
    ELSE
        RAISE NOTICE 'Триггер update_user_feedback_updated_at уже существует';
    END IF;
END $$;

-- Выводим информацию о таблице для проверки
SELECT
    c.column_name,
    c.data_type,
    c.is_nullable,
    c.column_default
FROM information_schema.columns c
WHERE c.table_name = 'user_feedback'
ORDER BY c.ordinal_position;

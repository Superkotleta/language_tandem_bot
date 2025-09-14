-- Обновление для исправления ошибки сохранения обратной связи
-- Запустите этот скрипт в вашей базе данных

-- Добавляем индексы для производительности
CREATE INDEX IF NOT EXISTS idx_user_feedback_user_id ON user_feedback(user_id);
CREATE INDEX IF NOT EXISTS idx_user_feedback_created_at ON user_feedback(created_at);
CREATE INDEX IF NOT EXISTS idx_user_feedback_processed ON user_feedback(is_processed);

-- Проверяем структуру таблицы
SELECT
    column_name,
    data_type,
    is_nullable,
    column_default
FROM information_schema.columns
WHERE table_name = 'user_feedback'
    AND table_schema = 'public'
ORDER BY ordinal_position;

-- Добавляем колонку admin_response если её нет
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                   WHERE table_name = 'user_feedback'
                   AND column_name = 'admin_response') THEN
        ALTER TABLE user_feedback ADD COLUMN admin_response TEXT;
    END IF;
END $$;

-- Проверяем данные
SELECT COUNT(*) as total_feedbacks FROM user_feedback;
SELECT * FROM user_feedback ORDER BY created_at DESC LIMIT 5;

-- Проверяем права пользователя
SELECT
    grantee,
    privilege_type
FROM information_schema.role_table_grants
WHERE table_name = 'user_feedback'
ORDER BY grantee;

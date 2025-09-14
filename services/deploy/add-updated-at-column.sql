-- Добавление колонки updated_at к существующей таблице user_feedback
-- Выполните этот скрипт в PostgreSQL-контейнере

-- Добавляем колонку если её нет
ALTER TABLE user_feedback
ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP DEFAULT NOW();

-- Обновляем updated_at для существующих записей
UPDATE user_feedback
SET updated_at = created_at
WHERE updated_at IS NULL;

-- Создаем функцию триггера
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Добавляем триггер (будет выполнять обновление updated_at автоматически)
DROP TRIGGER IF EXISTS update_user_feedback_updated_at ON user_feedback;
CREATE TRIGGER update_user_feedback_updated_at
    BEFORE UPDATE ON user_feedback
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Проверяем, что колонка добавлена
SELECT column_name, data_type, is_nullable, column_default
FROM information_schema.columns
WHERE table_name = 'user_feedback' AND column_name = 'updated_at';

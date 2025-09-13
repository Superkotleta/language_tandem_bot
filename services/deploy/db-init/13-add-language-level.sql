-- Добавляем поле для уровня владения изучаемым языком
ALTER TABLE users 
ADD COLUMN IF NOT EXISTS target_language_level TEXT DEFAULT '';

-- Добавляем индекс для поля уровня языка
CREATE INDEX IF NOT EXISTS idx_users_target_language_level ON users(target_language_level);
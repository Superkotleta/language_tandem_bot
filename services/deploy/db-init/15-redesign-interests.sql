-- Переработка системы интересов с категоризацией и приоритетами
-- Создаем новые таблицы для категоризированных интересов

-- 1. Категории интересов
CREATE TABLE IF NOT EXISTS interest_categories (
    id SERIAL PRIMARY KEY,
    key_name TEXT UNIQUE NOT NULL,
    display_order INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW()
);

-- 2. Обновляем таблицу интересов
ALTER TABLE interests 
ADD COLUMN IF NOT EXISTS category_id INT REFERENCES interest_categories(id),
ADD COLUMN IF NOT EXISTS display_order INT DEFAULT 0;

-- 3. Новая таблица для выборов пользователя с приоритетами
CREATE TABLE IF NOT EXISTS user_interest_selections (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id) ON DELETE CASCADE,
    interest_id INT REFERENCES interests(id) ON DELETE CASCADE,
    is_primary BOOLEAN DEFAULT FALSE,
    selection_order INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(user_id, interest_id)
);

-- 4. Конфигурация для баллов совместимости
CREATE TABLE IF NOT EXISTS matching_config (
    id SERIAL PRIMARY KEY,
    config_key TEXT UNIQUE NOT NULL,
    config_value TEXT NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 5. Настройки лимитов основных интересов
CREATE TABLE IF NOT EXISTS interest_limits_config (
    id SERIAL PRIMARY KEY,
    min_primary_interests INT DEFAULT 1,
    max_primary_interests INT DEFAULT 3,
    primary_percentage DECIMAL(5,2) DEFAULT 0.3, -- 30% от общего количества
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Индексы для производительности
CREATE INDEX IF NOT EXISTS idx_interest_categories_display_order ON interest_categories(display_order);
CREATE INDEX IF NOT EXISTS idx_interests_category_id ON interests(category_id);
CREATE INDEX IF NOT EXISTS idx_interests_display_order ON interests(display_order);
CREATE INDEX IF NOT EXISTS idx_user_interest_selections_user_id ON user_interest_selections(user_id);
CREATE INDEX IF NOT EXISTS idx_user_interest_selections_is_primary ON user_interest_selections(is_primary);
CREATE INDEX IF NOT EXISTS idx_user_interest_selections_selection_order ON user_interest_selections(selection_order);

-- Триггер для обновления updated_at в matching_config
CREATE OR REPLACE FUNCTION update_matching_config_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

DROP TRIGGER IF EXISTS update_matching_config_updated_at ON matching_config;
CREATE TRIGGER update_matching_config_updated_at
    BEFORE UPDATE ON matching_config
    FOR EACH ROW EXECUTE FUNCTION update_matching_config_updated_at();

-- Триггер для обновления updated_at в interest_limits_config
CREATE OR REPLACE FUNCTION update_interest_limits_config_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

DROP TRIGGER IF EXISTS update_interest_limits_config_updated_at ON interest_limits_config;
CREATE TRIGGER update_interest_limits_config_updated_at
    BEFORE UPDATE ON interest_limits_config
    FOR EACH ROW EXECUTE FUNCTION update_interest_limits_config_updated_at();

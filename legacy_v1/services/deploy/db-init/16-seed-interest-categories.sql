-- Заполнение категорий интересов
INSERT INTO interest_categories (key_name, display_order) VALUES 
    ('entertainment', 1),    -- Развлечения
    ('education', 2),        -- Образование
    ('active', 3),           -- Активный образ жизни
    ('creative', 4),         -- Творчество
    ('social', 5)            -- Социальные
ON CONFLICT (key_name) DO UPDATE SET
    display_order = EXCLUDED.display_order;

-- Обновляем существующие интересы, привязывая их к категориям
UPDATE interests SET category_id = (
    SELECT id FROM interest_categories WHERE key_name = 'entertainment'
) WHERE key_name IN ('movies_tv', 'music', 'games');

UPDATE interests SET category_id = (
    SELECT id FROM interest_categories WHERE key_name = 'education'
) WHERE key_name IN ('books', 'technology', 'science');

UPDATE interests SET category_id = (
    SELECT id FROM interest_categories WHERE key_name = 'active'
) WHERE key_name IN ('sports', 'travel');

UPDATE interests SET category_id = (
    SELECT id FROM interest_categories WHERE key_name = 'creative'
) WHERE key_name IN ('cooking', 'art');

-- Добавляем новые интересы для полноты категорий
INSERT INTO interests (key_name, category_id, type, display_order) VALUES 
    -- Развлечения
    ('tv_shows', (SELECT id FROM interest_categories WHERE key_name = 'entertainment'), 'entertainment', 1),
    ('comedy', (SELECT id FROM interest_categories WHERE key_name = 'entertainment'), 'entertainment', 2),
    ('anime', (SELECT id FROM interest_categories WHERE key_name = 'entertainment'), 'entertainment', 3),
    
    -- Образование
    ('languages', (SELECT id FROM interest_categories WHERE key_name = 'education'), 'education', 1),
    ('history', (SELECT id FROM interest_categories WHERE key_name = 'education'), 'education', 2),
    ('philosophy', (SELECT id FROM interest_categories WHERE key_name = 'education'), 'education', 3),
    
    -- Активный образ жизни
    ('fitness', (SELECT id FROM interest_categories WHERE key_name = 'active'), 'active', 1),
    ('outdoor', (SELECT id FROM interest_categories WHERE key_name = 'active'), 'active', 2),
    ('dancing', (SELECT id FROM interest_categories WHERE key_name = 'active'), 'active', 3),
    
    -- Творчество
    ('photography', (SELECT id FROM interest_categories WHERE key_name = 'creative'), 'creative', 1),
    ('writing', (SELECT id FROM interest_categories WHERE key_name = 'creative'), 'creative', 2),
    ('design', (SELECT id FROM interest_categories WHERE key_name = 'creative'), 'creative', 3),
    
    -- Социальные
    ('volunteering', (SELECT id FROM interest_categories WHERE key_name = 'social'), 'social', 1),
    ('politics', (SELECT id FROM interest_categories WHERE key_name = 'social'), 'social', 2),
    ('psychology', (SELECT id FROM interest_categories WHERE key_name = 'social'), 'social', 3)
ON CONFLICT (key_name) DO NOTHING;

-- Настройка конфигурации баллов совместимости
INSERT INTO matching_config (config_key, config_value, description) VALUES 
    ('primary_interest_score', '3', 'Баллы за совпадение основных интересов'),
    ('additional_interest_score', '1', 'Баллы за совпадение дополнительных интересов'),
    ('min_compatibility_score', '5', 'Минимальный балл совместимости для показа'),
    ('max_matches_per_user', '10', 'Максимальное количество совпадений на пользователя')
ON CONFLICT (config_key) DO UPDATE SET
    config_value = EXCLUDED.config_value,
    description = EXCLUDED.description,
    updated_at = NOW();

-- Настройка лимитов основных интересов
INSERT INTO interest_limits_config (min_primary_interests, max_primary_interests, primary_percentage) VALUES 
    (1, 5, 0.3)  -- Минимум 1, максимум 5, 30% от общего количества
ON CONFLICT (id) DO UPDATE SET
    min_primary_interests = EXCLUDED.min_primary_interests,
    max_primary_interests = EXCLUDED.max_primary_interests,
    primary_percentage = EXCLUDED.primary_percentage,
    updated_at = NOW();

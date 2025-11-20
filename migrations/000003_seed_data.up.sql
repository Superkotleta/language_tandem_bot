-- Seed Languages
INSERT INTO languages (code, names, flag) VALUES
('en', '{"en": "English", "ru": "Английский", "es": "Inglés", "zh": "英语", "native": "English"}', '🇬🇧'),
('ru', '{"en": "Russian", "ru": "Русский", "es": "Ruso", "zh": "俄语", "native": "Русский"}', '🇷🇺'),
('es', '{"en": "Spanish", "ru": "Испанский", "es": "Español", "zh": "西班牙语", "native": "Español"}', '🇪🇸'),
('zh', '{"en": "Chinese", "ru": "Китайский", "es": "Chino", "zh": "中文", "native": "中文"}', '🇨🇳')
ON CONFLICT (code) DO UPDATE SET names = EXCLUDED.names, flag = EXCLUDED.flag;

-- Seed Categories and Interests using a DO block
DO $$
DECLARE
    cat_id UUID;
BEGIN
    -- Category: Entertainment
    INSERT INTO interest_categories (slug, names, display_order)
    VALUES ('entertainment', '{"en": "Entertainment", "ru": "Развлечения"}', 10)
    ON CONFLICT (slug) DO UPDATE SET names = EXCLUDED.names
    RETURNING id INTO cat_id;

    INSERT INTO interests (category_id, slug, names) VALUES
    (cat_id, 'movies', '{"en": "Movies & TV", "ru": "Кино и сериалы"}'),
    (cat_id, 'music', '{"en": "Music", "ru": "Музыка"}'),
    (cat_id, 'games', '{"en": "Games", "ru": "Игры"}'),
    (cat_id, 'anime', '{"en": "Anime", "ru": "Аниме"}')
    ON CONFLICT (slug) DO UPDATE SET names = EXCLUDED.names, category_id = cat_id;

    -- Category: Education
    INSERT INTO interest_categories (slug, names, display_order)
    VALUES ('education', '{"en": "Education", "ru": "Образование"}', 20)
    ON CONFLICT (slug) DO UPDATE SET names = EXCLUDED.names
    RETURNING id INTO cat_id;

    INSERT INTO interests (category_id, slug, names) VALUES
    (cat_id, 'books', '{"en": "Books", "ru": "Книги"}'),
    (cat_id, 'languages', '{"en": "Languages", "ru": "Языки"}'),
    (cat_id, 'science', '{"en": "Science", "ru": "Наука"}'),
    (cat_id, 'technology', '{"en": "Technology", "ru": "Технологии"}')
    ON CONFLICT (slug) DO UPDATE SET names = EXCLUDED.names, category_id = cat_id;

    -- Category: Active Lifestyle
    INSERT INTO interest_categories (slug, names, display_order)
    VALUES ('active', '{"en": "Active Lifestyle", "ru": "Активный отдых"}', 30)
    ON CONFLICT (slug) DO UPDATE SET names = EXCLUDED.names
    RETURNING id INTO cat_id;

    INSERT INTO interests (category_id, slug, names) VALUES
    (cat_id, 'sports', '{"en": "Sports", "ru": "Спорт"}'),
    (cat_id, 'travel', '{"en": "Travel", "ru": "Путешествия"}'),
    (cat_id, 'fitness', '{"en": "Fitness", "ru": "Фитнес"}'),
    (cat_id, 'outdoor', '{"en": "Outdoor", "ru": "Природа"}')
    ON CONFLICT (slug) DO UPDATE SET names = EXCLUDED.names, category_id = cat_id;

    -- Category: Creative
    INSERT INTO interest_categories (slug, names, display_order)
    VALUES ('creative', '{"en": "Creative", "ru": "Творчество"}', 40)
    ON CONFLICT (slug) DO UPDATE SET names = EXCLUDED.names
    RETURNING id INTO cat_id;

    INSERT INTO interests (category_id, slug, names) VALUES
    (cat_id, 'art', '{"en": "Art", "ru": "Искусство"}'),
    (cat_id, 'photography', '{"en": "Photography", "ru": "Фотография"}'),
    (cat_id, 'writing', '{"en": "Writing", "ru": "Писательство"}'),
    (cat_id, 'cooking', '{"en": "Cooking", "ru": "Кулинария"}')
    ON CONFLICT (slug) DO UPDATE SET names = EXCLUDED.names, category_id = cat_id;

    -- Category: Social
    INSERT INTO interest_categories (slug, names, display_order)
    VALUES ('social', '{"en": "Social", "ru": "Общество"}', 50)
    ON CONFLICT (slug) DO UPDATE SET names = EXCLUDED.names
    RETURNING id INTO cat_id;

    INSERT INTO interests (category_id, slug, names) VALUES
    (cat_id, 'psychology', '{"en": "Psychology", "ru": "Психология"}'),
    (cat_id, 'politics', '{"en": "Politics", "ru": "Политика"}'),
    (cat_id, 'volunteering', '{"en": "Volunteering", "ru": "Волонтерство"}')
    ON CONFLICT (slug) DO UPDATE SET names = EXCLUDED.names, category_id = cat_id;

END $$;


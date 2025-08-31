-- Поддерживаемые языки
INSERT INTO languages (code, name_native, name_en, is_interface_language) VALUES 
    ('ru', 'Русский', 'Russian', true),
    ('en', 'English', 'English', true),
    ('es', 'Español', 'Spanish', true),
    ('zh', '中文', 'Chinese', true)
ON CONFLICT (code) DO UPDATE SET
    name_native = EXCLUDED.name_native,
    name_en = EXCLUDED.name_en,
    is_interface_language = EXCLUDED.is_interface_language;

-- Интересы (базовые ключи)
INSERT INTO interests (key_name, type) VALUES 
    ('movies_tv', 'entertainment'),
    ('music', 'entertainment'),
    ('sports', 'active'),
    ('travel', 'active'),
    ('books', 'education'),
    ('technology', 'education'),
    ('cooking', 'creative'),
    ('art', 'creative'),
    ('games', 'entertainment'),
    ('science', 'education')
ON CONFLICT (key_name) DO NOTHING;

-- Переводы интересов
INSERT INTO interest_translations (interest_id, language_code, name) 
SELECT 
    i.id,
    l.lang_code,
    t.translation
FROM interests i
CROSS JOIN (VALUES 
    ('ru'), ('en'), ('es'), ('zh')
) l(lang_code)
JOIN (VALUES 
    ('movies_tv', 'ru', 'Фильмы и сериалы'),
    ('movies_tv', 'en', 'Movies & TV'),
    ('movies_tv', 'es', 'Películas y series'),
    ('movies_tv', 'zh', '电影和电视'),
    
    ('music', 'ru', 'Музыка'),
    ('music', 'en', 'Music'),
    ('music', 'es', 'Música'),
    ('music', 'zh', '音乐'),
    
    ('sports', 'ru', 'Спорт'),
    ('sports', 'en', 'Sports'),
    ('sports', 'es', 'Deportes'),
    ('sports', 'zh', '体育'),
    
    ('travel', 'ru', 'Путешествия'),
    ('travel', 'en', 'Travel'),
    ('travel', 'es', 'Viajes'),
    ('travel', 'zh', '旅行'),
    
    ('books', 'ru', 'Книги'),
    ('books', 'en', 'Books'),
    ('books', 'es', 'Libros'),
    ('books', 'zh', '书籍'),
    
    ('technology', 'ru', 'Технологии'),
    ('technology', 'en', 'Technology'),
    ('technology', 'es', 'Tecnología'),
    ('technology', 'zh', '技术'),
    
    ('cooking', 'ru', 'Кулинария'),
    ('cooking', 'en', 'Cooking'),
    ('cooking', 'es', 'Cocina'),
    ('cooking', 'zh', '烹饪'),
    
    ('art', 'ru', 'Искусство'),
    ('art', 'en', 'Art'),
    ('art', 'es', 'Arte'),
    ('art', 'zh', '艺术'),
    
    ('games', 'ru', 'Игры'),
    ('games', 'en', 'Games'),
    ('games', 'es', 'Juegos'),
    ('games', 'zh', '游戏'),
    
    ('science', 'ru', 'Наука'),
    ('science', 'en', 'Science'),
    ('science', 'es', 'Ciencia'),
    ('science', 'zh', '科学')
) t(interest_key, lang_code, translation) 
    ON i.key_name = t.interest_key AND l.lang_code = t.lang_code
ON CONFLICT (interest_id, language_code) DO UPDATE SET
    name = EXCLUDED.name;

-- Локализация интерфейса бота
INSERT INTO localizations (key_name, language_code, translation, context) VALUES 
    -- Приветствие
    ('welcome_message', 'ru', '🎉 Привет, {name}! Добро пожаловать в Language Exchange Bot!

Я помогу найти тебе идеального языкового партнера для практики.

Давай заполним твой профиль! 📝', 'start_command'),
    
    ('welcome_message', 'en', '🎉 Hello, {name}! Welcome to Language Exchange Bot!

I''ll help you find the perfect language partner for practice.

Let''s fill out your profile! 📝', 'start_command'),
    
    ('welcome_message', 'es', '🎉 ¡Hola, {name}! ¡Bienvenido al Bot de Intercambio de Idiomas!

Te ayudaré a encontrar el compañero de idioma perfecto para practicar.

¡Vamos a completar tu perfil! 📝', 'start_command'),
    
    ('welcome_message', 'zh', '🎉 你好，{name}！欢迎使用语言交换机器人！

我会帮你找到完美的语言练习伙伴。

让我们来完善你的资料吧！📝', 'start_command'),

    -- Выбор родного языка
    ('choose_native_language', 'ru', 'Шаг 1: Выбери свой родной язык:', 'language_selection'),
    ('choose_native_language', 'en', 'Step 1: Choose your native language:', 'language_selection'),
    ('choose_native_language', 'es', 'Paso 1: Elige tu idioma nativo:', 'language_selection'),
    ('choose_native_language', 'zh', '第1步：选择你的母语：', 'language_selection'),

    -- Выбор изучаемого языка
    ('choose_target_language', 'ru', '✅ Родной язык: {language}

Теперь выбери язык, который хочешь изучать:', 'language_selection'),
    ('choose_target_language', 'en', '✅ Native language: {language}

Now choose the language you want to learn:', 'language_selection'),
    ('choose_target_language', 'es', '✅ Idioma nativo: {language}

Ahora elige el idioma que quieres aprender:', 'language_selection'),
    ('choose_target_language', 'zh', '✅ 母语：{language}

现在选择你想学习的语言：', 'language_selection'),

    -- Языки для кнопок
    ('language_ru', 'ru', '🇷🇺 Русский'),
    ('language_ru', 'en', '🇷🇺 Russian'),
    ('language_ru', 'es', '🇷🇺 Ruso'),
    ('language_ru', 'zh', '🇷🇺 俄语'),
    
    ('language_en', 'ru', '🇺🇸 Английский'),
    ('language_en', 'en', '🇺🇸 English'),
    ('language_en', 'es', '🇺🇸 Inglés'),
    ('language_en', 'zh', '🇺🇸 英语'),
    
    ('language_es', 'ru', '🇪🇸 Испанский'),
    ('language_es', 'en', '🇪🇸 Spanish'),
    ('language_es', 'es', '🇪🇸 Español'),
    ('language_es', 'zh', '🇪🇸 西班牙语'),
    
    ('language_zh', 'ru', '🇨🇳 Китайский'),
    ('language_zh', 'en', '🇨🇳 Chinese'),
    ('language_zh', 'es', '🇨🇳 Chino'),
    ('language_zh', 'zh', '🇨🇳 中文')

ON CONFLICT (key_name, language_code) DO UPDATE SET
    translation = EXCLUDED.translation,
    updated_at = NOW();

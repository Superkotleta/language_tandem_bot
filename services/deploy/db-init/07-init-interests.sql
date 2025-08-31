-- Справочник интересов с поддержкой локализации
CREATE TABLE IF NOT EXISTS interests (
    id SERIAL PRIMARY KEY,
    key_name TEXT NOT NULL UNIQUE,   -- Уникальный ключ интереса
    type TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Переводы интересов
CREATE TABLE IF NOT EXISTS interest_translations (
    id SERIAL PRIMARY KEY,
    interest_id INT REFERENCES interests(id) ON DELETE CASCADE,
    language_code VARCHAR(10) NOT NULL,
    name TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(interest_id, language_code),
    FOREIGN KEY (language_code) REFERENCES languages(code)
);

CREATE INDEX IF NOT EXISTS idx_interests_type ON interests(type);
CREATE INDEX IF NOT EXISTS idx_interest_translations_lang ON interest_translations(language_code);

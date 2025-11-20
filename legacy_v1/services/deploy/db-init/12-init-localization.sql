-- Таблица локализации для интерфейса бота
CREATE TABLE IF NOT EXISTS localizations (
    id SERIAL PRIMARY KEY,
    key_name TEXT NOT NULL,           -- Ключ для перевода, например 'welcome_message'
    language_code VARCHAR(10) NOT NULL,
    translation TEXT NOT NULL,        -- Перевод
    context TEXT DEFAULT NULL,        -- Контекст использования
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(key_name, language_code),
    FOREIGN KEY (language_code) REFERENCES languages(code)
);

CREATE INDEX IF NOT EXISTS idx_localizations_key_lang ON localizations(key_name, language_code);
CREATE INDEX IF NOT EXISTS idx_localizations_language ON localizations(language_code);

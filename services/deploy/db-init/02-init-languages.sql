-- Справочник языков (только 4 поддерживаемых)
CREATE TABLE IF NOT EXISTS languages (
    id SERIAL PRIMARY KEY,
    code VARCHAR(10) UNIQUE NOT NULL,
    name_native TEXT NOT NULL,  -- Название на родном языке
    name_en TEXT NOT NULL,      -- Название на английском
    is_interface_language BOOLEAN DEFAULT TRUE,  -- Доступен ли как язык интерфейса
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_languages_code ON languages(code);

-- Основная таблица пользователей с языком интерфейса
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    telegram_id BIGINT UNIQUE NOT NULL,
    username TEXT,
    first_name TEXT,
    native_language_code VARCHAR(10), -- ✅ ДОБАВЛЯЕМ
    target_language_code VARCHAR(10), -- ✅ ДОБАВЛЯЕМ
    interface_language_code VARCHAR(10) DEFAULT 'en',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    state TEXT DEFAULT 'new', -- ✅ ИСПРАВЛЯЕМ
    profile_completion_level INT DEFAULT 0,
    status TEXT DEFAULT 'new' -- ✅ ИСПРАВЛЯЕМ
        CHECK (status IN ('new', 'filling_profile', 'active', 'paused')), -- ✅ ДОБАВЛЯЕМ нужные статусы
    FOREIGN KEY (interface_language_code) REFERENCES languages(code),
    FOREIGN KEY (native_language_code) REFERENCES languages(code),
    FOREIGN KEY (target_language_code) REFERENCES languages(code)
);

CREATE INDEX IF NOT EXISTS idx_users_telegram_id ON users(telegram_id);
CREATE INDEX IF NOT EXISTS idx_users_status ON users(status);
CREATE INDEX IF NOT EXISTS idx_users_interface_language ON users(interface_language_code);
CREATE INDEX IF NOT EXISTS idx_users_native_language ON users(native_language_code);
CREATE INDEX IF NOT EXISTS idx_users_target_language ON users(target_language_code);

-- Языковые пары пользователей
CREATE TABLE IF NOT EXISTS user_language_pairs (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id) ON DELETE CASCADE,
    native_language_id INT REFERENCES languages(id),
    target_language_id INT REFERENCES languages(id),
    target_level TEXT CHECK (target_level IN ('beginner', 'intermediate', 'advanced')),
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(user_id, native_language_id, target_language_id)
);

CREATE INDEX IF NOT EXISTS idx_user_language_pairs_user_id ON user_language_pairs(user_id);
CREATE INDEX IF NOT EXISTS idx_user_language_pairs_native ON user_language_pairs(native_language_id);
CREATE INDEX IF NOT EXISTS idx_user_language_pairs_target ON user_language_pairs(target_language_id);

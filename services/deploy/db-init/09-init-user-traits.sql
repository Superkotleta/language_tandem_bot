-- Черты характера пользователей
CREATE TABLE IF NOT EXISTS user_traits (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id) ON DELETE CASCADE,
    trait TEXT NOT NULL,
    value TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_user_traits_user_id ON user_traits(user_id);
CREATE INDEX IF NOT EXISTS idx_user_traits_trait ON user_traits(trait);

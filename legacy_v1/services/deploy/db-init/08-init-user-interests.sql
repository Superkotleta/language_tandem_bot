-- Связь пользователей с интересами
CREATE TABLE IF NOT EXISTS user_interests (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id) ON DELETE CASCADE,
    interest_id INT REFERENCES interests(id) ON DELETE CASCADE,
    is_primary BOOLEAN DEFAULT FALSE, -- ✅ ИСПРАВЛЯЕМ is_main -> is_primary
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(user_id, interest_id)
);

CREATE INDEX IF NOT EXISTS idx_user_interests_user_id ON user_interests(user_id);
CREATE INDEX IF NOT EXISTS idx_user_interests_interest_id ON user_interests(interest_id);
CREATE INDEX IF NOT EXISTS idx_user_interests_is_primary ON user_interests(is_primary); -- ✅ ИСПРАВЛЯЕМ

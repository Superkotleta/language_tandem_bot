-- Очередь найденных совпадений
CREATE TABLE IF NOT EXISTS match_queue (
    id SERIAL PRIMARY KEY,
    user1_id INT REFERENCES users(id) ON DELETE CASCADE,
    user2_id INT REFERENCES users(id) ON DELETE CASCADE,
    compatibility_score INT DEFAULT 0,
    found_at TIMESTAMP DEFAULT NOW(),
    sent_at TIMESTAMP NULL,
    status TEXT DEFAULT 'pending' 
        CHECK (status IN ('pending', 'sent', 'cancelled')),
    UNIQUE(user1_id, user2_id)
);

CREATE INDEX IF NOT EXISTS idx_match_queue_user1 ON match_queue(user1_id);
CREATE INDEX IF NOT EXISTS idx_match_queue_user2 ON match_queue(user2_id);
CREATE INDEX IF NOT EXISTS idx_match_queue_status ON match_queue(status);
CREATE INDEX IF NOT EXISTS idx_match_queue_found_at ON match_queue(found_at);

-- Индекс для предотвращения дублирующих пар (A-B = B-A)
CREATE UNIQUE INDEX IF NOT EXISTS idx_match_queue_users_unique 
    ON match_queue(LEAST(user1_id, user2_id), GREATEST(user1_id, user2_id));

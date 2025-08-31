-- Временная доступность пользователей
CREATE TABLE IF NOT EXISTS user_time_availability (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id) ON DELETE CASCADE,
    day_type TEXT CHECK (day_type IN ('weekdays', 'weekends', 'any', 'specific')),
    specific_days TEXT[] DEFAULT NULL,
    time_slot TEXT CHECK (time_slot IN ('morning', 'day', 'evening', 'late')),
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_user_time_availability_user_id ON user_time_availability(user_id);
CREATE INDEX IF NOT EXISTS idx_user_time_availability_day_type ON user_time_availability(day_type);
CREATE INDEX IF NOT EXISTS idx_user_time_availability_time_slot ON user_time_availability(time_slot);

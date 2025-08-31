-- Предпочтения по общению
CREATE TABLE IF NOT EXISTS friendship_preferences (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id) ON DELETE CASCADE,
    activity_type TEXT CHECK (activity_type IN (
        'movies', 'games', 'casual_chat', 'creative', 'active', 'educational'
    )),
    communication_style TEXT CHECK (communication_style IN (
        'text', 'voice_msg', 'audio_call', 'video_call', 'meet_person'
    )),
    communication_frequency TEXT CHECK (communication_frequency IN (
        'spontaneous', 'weekly', 'daily', 'intensive'
    )),
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(user_id)
);

CREATE INDEX IF NOT EXISTS idx_friendship_preferences_user_id ON friendship_preferences(user_id);

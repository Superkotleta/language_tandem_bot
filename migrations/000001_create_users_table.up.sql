CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    social_id VARCHAR(255) NOT NULL,
    platform VARCHAR(50) NOT NULL,
    first_name VARCHAR(255),
    username VARCHAR(255),
    interface_lang VARCHAR(10) DEFAULT 'en',
    native_lang VARCHAR(10),
    target_lang VARCHAR(10),
    target_level VARCHAR(20),
    status VARCHAR(50) DEFAULT 'filling_profile',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(social_id, platform)
);

CREATE INDEX idx_users_social_platform ON users(social_id, platform);

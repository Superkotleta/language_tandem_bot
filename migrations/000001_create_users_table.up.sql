CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    social_id VARCHAR(255) NOT NULL,
    platform VARCHAR(50) NOT NULL,
    first_name VARCHAR(255),
    username VARCHAR(255),
    language VARCHAR(10) DEFAULT 'en',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(social_id, platform)
);

CREATE INDEX idx_users_social_platform ON users(social_id, platform);



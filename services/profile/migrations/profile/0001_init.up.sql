-- Create profile schema
CREATE SCHEMA IF NOT EXISTS profile;

-- Set search path to profile schema
SET search_path TO profile, public;

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    telegram_id BIGINT UNIQUE,
    discord_id BIGINT UNIQUE,
    username VARCHAR(255),
    first_name VARCHAR(255),
    last_name VARCHAR(255),
    email VARCHAR(255),
    phone VARCHAR(20),
    bio TEXT,
    age INTEGER CHECK (age >= 13 AND age <= 120),
    gender VARCHAR(20) CHECK (gender IN ('male', 'female', 'other', 'prefer_not_to_say')),
    country VARCHAR(100),
    city VARCHAR(100),
    timezone VARCHAR(50),
    profile_picture_url TEXT,
    is_active BOOLEAN DEFAULT true,
    is_verified BOOLEAN DEFAULT false,
    last_seen TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Languages table
CREATE TABLE IF NOT EXISTS languages (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(10) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    native_name VARCHAR(100) NOT NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- User languages table (languages user speaks)
CREATE TABLE IF NOT EXISTS user_languages (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    language_id BIGINT NOT NULL REFERENCES languages(id) ON DELETE CASCADE,
    level VARCHAR(20) NOT NULL CHECK (level IN ('beginner', 'elementary', 'intermediate', 'upper_intermediate', 'advanced', 'native')),
    is_learning BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(user_id, language_id)
);

-- Interests table
CREATE TABLE IF NOT EXISTS interests (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    category VARCHAR(50) NOT NULL,
    description TEXT,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- User interests table
CREATE TABLE IF NOT EXISTS user_interests (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    interest_id BIGINT NOT NULL REFERENCES interests(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(user_id, interest_id)
);

-- User preferences table
CREATE TABLE IF NOT EXISTS user_preferences (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    min_age INTEGER CHECK (min_age >= 13 AND min_age <= 120),
    max_age INTEGER CHECK (max_age >= 13 AND max_age <= 120),
    preferred_gender VARCHAR(20) CHECK (preferred_gender IN ('male', 'female', 'other', 'any')),
    preferred_countries TEXT[], -- Array of country codes
    preferred_languages TEXT[], -- Array of language codes
    max_distance INTEGER, -- in kilometers
    timezone_offset INTEGER, -- in minutes from UTC
    availability_start TIME,
    availability_end TIME,
    is_online_only BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(user_id)
);

-- User traits table (personality, learning style, etc.)
CREATE TABLE IF NOT EXISTS user_traits (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    trait_type VARCHAR(50) NOT NULL,
    trait_value VARCHAR(100) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(user_id, trait_type)
);

-- User time availability table
CREATE TABLE IF NOT EXISTS user_time_availability (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    day_of_week INTEGER NOT NULL CHECK (day_of_week >= 0 AND day_of_week <= 6), -- 0 = Sunday, 6 = Saturday
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    timezone VARCHAR(50) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_users_telegram_id ON users(telegram_id);
CREATE INDEX IF NOT EXISTS idx_users_discord_id ON users(discord_id);
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_is_active ON users(is_active);
CREATE INDEX IF NOT EXISTS idx_users_country ON users(country);
CREATE INDEX IF NOT EXISTS idx_users_city ON users(city);
CREATE INDEX IF NOT EXISTS idx_users_age ON users(age);
CREATE INDEX IF NOT EXISTS idx_users_gender ON users(gender);
CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at);

CREATE INDEX IF NOT EXISTS idx_user_languages_user_id ON user_languages(user_id);
CREATE INDEX IF NOT EXISTS idx_user_languages_language_id ON user_languages(language_id);
CREATE INDEX IF NOT EXISTS idx_user_languages_level ON user_languages(level);
CREATE INDEX IF NOT EXISTS idx_user_languages_is_learning ON user_languages(is_learning);

CREATE INDEX IF NOT EXISTS idx_user_interests_user_id ON user_interests(user_id);
CREATE INDEX IF NOT EXISTS idx_user_interests_interest_id ON user_interests(interest_id);

CREATE INDEX IF NOT EXISTS idx_user_preferences_user_id ON user_preferences(user_id);
CREATE INDEX IF NOT EXISTS idx_user_preferences_min_age ON user_preferences(min_age);
CREATE INDEX IF NOT EXISTS idx_user_preferences_max_age ON user_preferences(max_age);
CREATE INDEX IF NOT EXISTS idx_user_preferences_preferred_gender ON user_preferences(preferred_gender);

CREATE INDEX IF NOT EXISTS idx_user_traits_user_id ON user_traits(user_id);
CREATE INDEX IF NOT EXISTS idx_user_traits_trait_type ON user_traits(trait_type);

CREATE INDEX IF NOT EXISTS idx_user_time_availability_user_id ON user_time_availability(user_id);
CREATE INDEX IF NOT EXISTS idx_user_time_availability_day_of_week ON user_time_availability(day_of_week);

-- Create updated_at trigger function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers for updated_at
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_user_languages_updated_at BEFORE UPDATE ON user_languages
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_user_preferences_updated_at BEFORE UPDATE ON user_preferences
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_user_traits_updated_at BEFORE UPDATE ON user_traits
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_user_time_availability_updated_at BEFORE UPDATE ON user_time_availability
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

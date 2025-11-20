-- Languages dictionary
CREATE TABLE IF NOT EXISTS languages (
    code VARCHAR(10) PRIMARY KEY,
    names JSONB NOT NULL,
    flag VARCHAR(10) NOT NULL
);

-- Interest Categories
CREATE TABLE IF NOT EXISTS interest_categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug VARCHAR(50) UNIQUE NOT NULL,
    names JSONB NOT NULL,
    display_order INT DEFAULT 0
);

-- Interests
CREATE TABLE IF NOT EXISTS interests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    category_id UUID REFERENCES interest_categories(id) ON DELETE CASCADE,
    slug VARCHAR(50) UNIQUE NOT NULL,
    names JSONB NOT NULL
);

CREATE INDEX idx_interests_category_id ON interests(category_id);

-- User Interests
CREATE TABLE IF NOT EXISTS user_interests (
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    interest_id UUID REFERENCES interests(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, interest_id)
);


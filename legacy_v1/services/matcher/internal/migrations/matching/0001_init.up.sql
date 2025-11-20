-- Ensure schema
CREATE SCHEMA IF NOT EXISTS matching;

-- match_queue
CREATE TABLE IF NOT EXISTS matching.match_queue (
  id SERIAL PRIMARY KEY,
  user1_id INT NOT NULL REFERENCES profile.users(id) ON DELETE CASCADE,
  user2_id INT NOT NULL REFERENCES profile.users(id) ON DELETE CASCADE,
  compatibility_score INT,
  found_at TIMESTAMP DEFAULT NOW(),
  sent_at TIMESTAMP NULL,
  status TEXT DEFAULT 'pending' CHECK (status IN ('pending','sent','cancelled'))
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_match_queue_users ON matching.match_queue(user1_id, user2_id);

-- tasks
CREATE TABLE IF NOT EXISTS matching.tasks (
  id SERIAL PRIMARY KEY,
  description TEXT NOT NULL,
  due_date TIMESTAMP NULL,
  target_users TEXT[] DEFAULT NULL,
  created_at TIMESTAMP DEFAULT NOW(),
  status TEXT DEFAULT 'pending' CHECK (status IN ('pending','sent','completed'))
);

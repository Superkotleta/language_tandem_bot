CREATE SCHEMA IF NOT EXISTS profile;

CREATE TABLE IF NOT EXISTS profile.users (...);
CREATE INDEX IF NOT EXISTS idx_users_telegram_id ON profile.users(telegram_id);

CREATE TABLE IF NOT EXISTS profile.languages (...);

CREATE TABLE IF NOT EXISTS profile.user_language_pairs (...);
CREATE INDEX IF NOT EXISTS idx_user_language_pairs_user_id ON profile.user_language_pairs(user_id);

CREATE TABLE IF NOT EXISTS profile.friendship_preferences (...);
CREATE TABLE IF NOT EXISTS profile.user_time_availability (...);
CREATE TABLE IF NOT EXISTS profile.interests (...);
CREATE TABLE IF NOT EXISTS profile.user_interests (...);
CREATE INDEX IF NOT EXISTS idx_user_interests_user_id ON profile.user_interests(user_id);
CREATE TABLE IF NOT EXISTS profile.user_traits (...);
CREATE INDEX IF NOT EXISTS idx_user_traits_user_id ON profile.user_traits(user_id);

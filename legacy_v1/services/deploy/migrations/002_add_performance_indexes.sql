-- Migration: 002_add_performance_indexes.sql
-- Description: Добавление индексов для оптимизации производительности запросов
-- Author: Performance Optimization Team
-- Date: 2025-01-XX

-- =============================================
-- Индексы для таблицы users
-- =============================================

-- Основной lookup по telegram_id (самый частый запрос)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_telegram_id 
ON users(telegram_id);

-- Фильтрация по языку интерфейса
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_interface_language 
ON users(interface_language_code);

-- Фильтрация по статусу пользователя
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_status 
ON users(status);

-- Фильтрация по состоянию пользователя
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_state 
ON users(state);

-- Композитный индекс для поиска активных пользователей по языкам
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_active_languages 
ON users(status, native_language_code, target_language_code) 
WHERE status = 'active';

-- Индекс для сортировки по дате создания
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_created_at 
ON users(created_at DESC);

-- =============================================
-- Индексы для таблицы user_interests
-- =============================================

-- Основной lookup по user_id (для получения интересов пользователя)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_user_interests_user_id 
ON user_interests(user_id);

-- Reverse lookup по interest_id (для поиска пользователей с определенным интересом)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_user_interests_interest_id 
ON user_interests(interest_id);

-- Композитный индекс для поиска пользователей с определенными интересами
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_user_interests_user_interest 
ON user_interests(user_id, interest_id);

-- Индекс для primary интересов
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_user_interests_primary 
ON user_interests(user_id, is_primary) 
WHERE is_primary = true;

-- =============================================
-- Индексы для таблицы feedback
-- =============================================

-- Основной индекс для admin queries (необработанные отзывы)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_feedback_unprocessed 
ON feedback(is_processed, created_at DESC) 
WHERE is_processed = false;

-- Индекс для поиска отзывов по пользователю
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_feedback_user_id 
ON feedback(user_id, created_at DESC);

-- Индекс для сортировки по дате
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_feedback_created_at 
ON feedback(created_at DESC);

-- =============================================
-- Индексы для таблицы interests
-- =============================================

-- Индекс для поиска по категории
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_interests_category_id 
ON interests(category_id);

-- Индекс для сортировки по порядку отображения
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_interests_display_order 
ON interests(display_order);

-- =============================================
-- Индексы для таблицы languages
-- =============================================

-- Индекс для поиска по коду языка
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_languages_code 
ON languages(code);

-- Индекс для интерфейсных языков
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_languages_interface 
ON languages(is_interface_language) 
WHERE is_interface_language = true;

-- =============================================
-- Индексы для таблицы user_time_availability
-- =============================================

-- Индекс для поиска доступности пользователя
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_user_time_availability_user_id 
ON user_time_availability(user_id);

-- =============================================
-- Индексы для таблицы user_friendship_preferences
-- =============================================

-- Индекс для поиска предпочтений пользователя
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_user_friendship_preferences_user_id 
ON user_friendship_preferences(user_id);

-- =============================================
-- Композитные индексы для matching алгоритма
-- =============================================

-- Индекс для поиска совместимых пользователей по языкам
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_matching_languages 
ON users(native_language_code, target_language_code, status) 
WHERE status = 'active';

-- Индекс для поиска пользователей с определенным уровнем языка
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_target_level 
ON users(target_language_level, status) 
WHERE status = 'active';

-- =============================================
-- Статистика по индексам
-- =============================================

-- Обновляем статистику для оптимизатора запросов
ANALYZE users;
ANALYZE user_interests;
ANALYZE feedback;
ANALYZE interests;
ANALYZE languages;
ANALYZE user_time_availability;
ANALYZE user_friendship_preferences;

-- =============================================
-- Комментарии к индексам
-- =============================================

COMMENT ON INDEX idx_users_telegram_id IS 'Основной lookup пользователей по Telegram ID';
COMMENT ON INDEX idx_users_interface_language IS 'Фильтрация пользователей по языку интерфейса';
COMMENT ON INDEX idx_users_active_languages IS 'Поиск активных пользователей по языковым парам';
COMMENT ON INDEX idx_user_interests_user_id IS 'Получение интересов пользователя';
COMMENT ON INDEX idx_user_interests_interest_id IS 'Поиск пользователей с определенным интересом';
COMMENT ON INDEX idx_feedback_unprocessed IS 'Административные запросы необработанных отзывов';
COMMENT ON INDEX idx_users_matching_languages IS 'Matching алгоритм по языковым парам';

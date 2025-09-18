-- Создание схем
CREATE SCHEMA IF NOT EXISTS profile;
CREATE SCHEMA IF NOT EXISTS matching;

-- Выдача прав пользователю postgres на все схемы
GRANT ALL PRIVILEGES ON SCHEMA profile TO postgres;
GRANT ALL PRIVILEGES ON SCHEMA matching TO postgres;
GRANT ALL PRIVILEGES ON SCHEMA public TO postgres;

-- Создание пользователей для сервисов (если нужно)
DO $$
BEGIN
   IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'profile_rw') THEN
      CREATE ROLE profile_rw LOGIN PASSWORD 'profile_pass';
   END IF;
   
   IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'matching_rw') THEN
      CREATE ROLE matching_rw LOGIN PASSWORD 'matching_pass';
   END IF;
END$$;

-- Права для сервисных пользователей
GRANT CONNECT ON DATABASE languagebot TO profile_rw;
GRANT CONNECT ON DATABASE languagebot TO matching_rw;

GRANT USAGE, CREATE ON SCHEMA profile TO profile_rw;
GRANT USAGE, CREATE ON SCHEMA matching TO matching_rw;

GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA profile TO profile_rw;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA matching TO matching_rw;

ALTER DEFAULT PRIVILEGES IN SCHEMA profile GRANT ALL ON TABLES TO profile_rw;
ALTER DEFAULT PRIVILEGES IN SCHEMA matching GRANT ALL ON TABLES TO matching_rw;

GRANT USAGE ON ALL SEQUENCES IN SCHEMA profile TO profile_rw;
GRANT USAGE ON ALL SEQUENCES IN SCHEMA matching TO matching_rw;

ALTER DEFAULT PRIVILEGES IN SCHEMA profile GRANT USAGE ON SEQUENCES TO profile_rw;
ALTER DEFAULT PRIVILEGES IN SCHEMA matching GRANT USAGE ON SEQUENCES TO matching_rw;

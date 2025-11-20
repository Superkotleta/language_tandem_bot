#!/bin/sh
set -euo pipefail

# Wait for postgres to be ready
until pg_isready -h postgres -U "${POSTGRES_USER}" -d "${POSTGRES_DB}" -q; do
  echo "Waiting for postgres..."
  sleep 2
done

echo "Bootstrap starting..."

export PGPASSWORD="${PGPASSWORD}"

# Create schemas, roles, grants (idempotent), set passwords from env
psql -h postgres -U "${POSTGRES_USER}" -d "${POSTGRES_DB}" <<'SQL'
-- Schemas
CREATE SCHEMA IF NOT EXISTS profile;
CREATE SCHEMA IF NOT EXISTS matching;
SQL

# Roles with passwords from env require shell variable expansion; use cat <<EOF
psql -h postgres -U "${POSTGRES_USER}" -d "${POSTGRES_DB}" <<EOF
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'profile_rw') THEN
    CREATE ROLE profile_rw LOGIN PASSWORD '${PROFILE_DB_PASS}';
  ELSE
    ALTER ROLE profile_rw WITH LOGIN PASSWORD '${PROFILE_DB_PASS}';
  END IF;
END $$;

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'matching_rw') THEN
    CREATE ROLE matching_rw LOGIN PASSWORD '${MATCHING_DB_PASS}';
  ELSE
    ALTER ROLE matching_rw WITH LOGIN PASSWORD '${MATCHING_DB_PASS}';
  END IF;
END $$;

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'matching_ro') THEN
    CREATE ROLE matching_ro LOGIN PASSWORD '${MATCHING_RO_DB_PASS}';
  ELSE
    ALTER ROLE matching_ro WITH LOGIN PASSWORD '${MATCHING_RO_DB_PASS}';
  END IF;
END $$;

-- Grants on schemas (allow CREATE where needed)
GRANT USAGE, CREATE ON SCHEMA profile TO profile_rw;
GRANT USAGE ON SCHEMA profile TO matching_ro;

GRANT USAGE, CREATE ON SCHEMA matching TO matching_rw;

-- Table and sequence privileges for existing objects
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA profile TO profile_rw;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA profile TO profile_rw;

GRANT SELECT ON ALL TABLES IN SCHEMA profile TO matching_ro;

GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA matching TO matching_rw;
GRANT USAGE, SELECT, UPDATE ON ALL SEQUENCES IN SCHEMA matching TO matching_rw;

-- Default privileges for future objects
ALTER DEFAULT PRIVILEGES IN SCHEMA profile GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO profile_rw;
ALTER DEFAULT PRIVILEGES IN SCHEMA profile GRANT USAGE, SELECT ON SEQUENCES TO profile_rw;
ALTER DEFAULT PRIVILEGES IN SCHEMA profile GRANT SELECT ON TABLES TO matching_ro;

ALTER DEFAULT PRIVILEGES IN SCHEMA matching GRANT ALL ON TABLES TO matching_rw;
ALTER DEFAULT PRIVILEGES IN SCHEMA matching GRANT USAGE, SELECT, UPDATE ON SEQUENCES TO matching_rw;
EOF

echo "Bootstrap completed."

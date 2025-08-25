SHELL := /bin/sh

.PHONY: up down restart logs psql-root psql-profile psql-matcher demo-rbac

up:
	cd deploy && docker compose up -d --build

down:
	cd deploy && docker compose down -v

restart: down up

logs:
	cd deploy && docker compose logs -f --tail=200

psql-root:
	cd deploy && docker compose exec -e PGPASSWORD=$$POSTGRES_PASSWORD postgres psql -U $$POSTGRES_USER $$POSTGRES_DB

psql-profile:
	cd deploy && docker compose exec -e PGPASSWORD=$$PROFILE_DB_PASS postgres psql "postgresql://profile_rw:$$PROFILE_DB_PASS@postgres:5432/$$POSTGRES_DB?sslmode=disable"

psql-matcher:
	cd deploy && docker compose exec -e PGPASSWORD=$$MATCHING_DB_PASS postgres psql "postgresql://matching_rw:$$MATCHING_DB_PASS@postgres:5432/$$POSTGRES_DB?sslmode=disable"

demo-rbac:
	@echo "== matching_ro can read profile.users, but cannot write =="
	cd deploy && docker compose exec -e PGPASSWORD=$$MATCHING_RO_DB_PASS postgres psql "postgresql://matching_ro:$$MATCHING_RO_DB_PASS@postgres:5432/$$POSTGRES_DB?sslmode=disable" -c "SET search_path=profile,public; SELECT count(*) FROM users;"
	cd deploy && docker compose exec -e PGPASSWORD=$$MATCHING_RO_DB_PASS postgres psql "postgresql://matching_ro:$$MATCHING_RO_DB_PASS@postgres:5432/$$POSTGRES_DB?sslmode=disable" -c "SET search_path=profile,public; INSERT INTO users(telegram_id) VALUES (9999); " || true

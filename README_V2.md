# Language Exchange Bot - Version 2.0 (Rewrite)

## What's New

This is a complete rewrite of the bot with a focus on:
- **Clean Architecture**: Clear separation between Domain, Service, Repository, and Delivery layers
- **Multi-platform Ready**: User profiles are platform-agnostic (telegram_id, vk_id, etc.)
- **Modern Stack**: pgx/v5, UUID primary keys, JSONB for flexible data
- **Internationalization**: Full i18n support with JSON locale files
- **Minimal Tech Debt**: No legacy dependencies, clean codebase

## Architecture

```
cmd/bot/                 # Entry point
internal/
  ├── domain/            # Business entities (User, Interest, Language)
  ├── repository/        # Database access (pgx)
  ├── service/           # Business logic
  ├── delivery/telegram/ # Telegram bot adapter
  ├── pkg/i18n/          # Localization
  └── ui/                # Message & Keyboard builders
```

## Database Schema

- `users`: Single profile with UUID, supports multiple platforms
- `languages`: Language dictionary with JSONB translations and flags
- `interest_categories`: JSONB-based categories with ordering
- `interests`: JSONB-based interests linked to categories
- `user_interests`: Many-to-many relationship

## Setup

1. Set environment variables:
```bash
export DATABASE_URL="postgres://user:pass@localhost:5432/dbname"
export TELEGRAM_TOKEN="your_bot_token"
export LOCALES_PATH="./locales"
```

2. Run migrations:
```bash
psql $DATABASE_URL < migrations/000001_create_users_table.up.sql
psql $DATABASE_URL < migrations/000002_create_reference_tables.up.sql
psql $DATABASE_URL < migrations/000003_seed_data.up.sql
```

3. Run the bot:
```bash
go run cmd/bot/main.go
```

## Current Status

✅ Database schema created & migrated to UUID/JSONB
✅ Domain models defined (User, Language, Interest, Category)
✅ Repositories implemented (UserRepo, ReferenceRepo)
✅ Service layer ready
✅ Telegram bot with menu logic
✅ Localization system
✅ Seed data for languages and interests

🚧 Profile wizard (onboarding flow) - coming next
🚧 Interest selection UI
🚧 Partner matching algorithm

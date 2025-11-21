# Changelog

## [2.0.1] - 2025-11-21

### Fixed
- Fixed connection pool leak in PostgreSQL repository on ping failure
- Added proper error checking for database row iterations
- Implemented graceful shutdown for Telegram bot updates channel
- Added error handling for Telegram message sending
- Normalized language codes in user registration (e.g., "es-ES" -> "es")
- Fixed template variable substitution in welcome messages
- Fixed potential panic by handling nil message.From in Telegram handler

## [2.0.0] - 2025-11-21

### Added
- Complete rewrite of the bot architecture
- Clean Architecture implementation (Domain, Service, Repository, Delivery)
- PostgreSQL database with pgx/v5
- JSONB-based reference tables for languages and interests
- Multi-language support (i18n)
- User profile management

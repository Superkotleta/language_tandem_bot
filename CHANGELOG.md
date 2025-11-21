# Changelog

## [2.0.3] - 2025-11-21

### Added
- **Linting System** in `deploy/linter/` folder
  - 4 golangci-lint configurations (gentle, fast, enhanced, strict)
  - Comprehensive Makefile with linter commands
  - Detailed linter documentation (README.md)
  - Integration with main deploy Makefile
  - Pre-commit hook support
- 100+ enabled linters for code quality
- Security, performance, and style checks
- Configurable complexity thresholds

### Changed
- Updated README.md with linter information
- Updated deploy/README.md with linter commands
- Enhanced Makefile with help command

## [2.0.2] - 2025-11-21

### Added
- Docker deployment configuration in `deploy/` folder
- Dockerfile with multi-stage build (Go 1.25 + Alpine)
- docker-compose.yml with PostgreSQL 17, pgAdmin, and bot services
- Makefile for convenient Docker management
- Automated log rotation for all containers (10MB max, 3 files)
- Healthcheck for PostgreSQL service
- Network isolation with bridge network
- Volume persistence for PostgreSQL data

### Changed
- Updated README_V2.md with Docker setup instructions
- Updated RUN_INSTRUCTIONS.md with detailed Docker usage guide

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

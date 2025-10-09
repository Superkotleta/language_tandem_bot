// Package localization provides localization constants for the bot.
// This file contains both localization keys and common application constants.
package localization

import "time"

// =============================================================================
// APPLICATION CONSTANTS (shared across multiple files)
// =============================================================================

// Telegram Handler Constants
// Used in: services/bot/internal/adapters/telegram/handlers.go.
const (
	MinPartsForFeedbackNav = 2 // Минимальное количество частей для навигации по отзывам
	MinPartsForNav         = 4 // Минимальное количество частей для навигации
)

// Note: User states and statuses are kept in models/user.go to avoid circular dependencies

// Validation Constants
// Used in: services/bot/internal/validation/validators.go.
const (
	MinTelegramID        = 100000000
	MaxUsernameLength    = 50
	MaxBioLength         = 500
	MaxContactInfoLength = 64
	MaxCommandLength     = 32
	MaxStringLength      = 100
	MaxInterestCount     = 10
	MinTextLength        = 10
	MaxTextLength        = 1000
	MinStringLength      = 3
	LanguageCodeLength   = 2
)

// Bot Service Constants
// Used in: services/bot/internal/core/service.go.
const (
	MinFeedbackLength = 10
	MaxFeedbackLength = 1000
)

// =============================================================================
// TIME CONSTANTS (centralized time configuration)
// =============================================================================
// Future Enhancement: Consider moving these to a configurable file (YAML/JSON)
// to allow runtime configuration without recompilation for different environments.
// This would require a config loader and env var support.
//
// Current approach: Compile-time constants for maximum performance (zero allocation)
// =============================================================================

// Cache TTL Constants (in minutes)
// Used in: services/bot/internal/cache/types.go, services/bot/internal/cache/cache.go.
const (
	TranslationsTTLMinutes = 30 // How long translations are cached (30 minutes)
	UsersTTLMinutes        = 15 // How long user data is cached (15 minutes)
	StatsTTLMinutes        = 5  // How long statistics are cached (5 minutes)
	CacheCleanupMinutes    = 5  // Interval between cache cleanup operations (5 minutes)
)

// Rate Limiter Constants (in minutes and seconds)
// Used in: services/bot/internal/adapters/telegram/rate_limiter.go.
const (
	RateLimitWindowMinutes  = 1  // Time window for rate limiting (1 minute)
	RateLimitBlockMinutes   = 2  // How long to block after exceeding limits (2 minutes)
	RateLimitCleanupMinutes = 10 // Interval for cleaning up expired rate limit entries (10 minutes)
)

// Redis Connection Constants (in seconds)
// Used in: services/bot/internal/cache/redis_cache.go.
const (
	RedisDialTimeoutSeconds  = 5   // Timeout for establishing Redis connection (5 seconds)
	RedisReadTimeoutSeconds  = 3   // Timeout for Redis read operations (3 seconds)
	RedisWriteTimeoutSeconds = 3   // Timeout for Redis write operations (3 seconds)
	RedisPoolSize            = 10  // Number of Redis connections in pool (10 connections)
	RedisMaxRetries          = 3   // Maximum number of retry attempts (3 retries)
	RedisMinRetryBackoffMs   = 8   // Minimum backoff between retries (8 milliseconds)
	RedisMaxRetryBackoffMs   = 512 // Maximum backoff between retries (512 milliseconds)
)

// Circuit Breaker Constants
// Used in: services/bot/internal/circuit_breaker/circuit_breaker.go.
const (
	DefaultMaxRequests         = 3  // Максимальное количество запросов в полуоткрытом состоянии
	DefaultIntervalSeconds     = 60 // Интервал в секундах между проверками
	DefaultTimeoutSeconds      = 60 // Таймаут в секундах для возврата в закрытое состояние
	DefaultConsecutiveFailures = 5  // Количество последовательных неудач для открытия

	// Service-specific Circuit Breaker configurations.
	TelegramMaxRequests      = 5  // Максимум запросов для Telegram
	TelegramIntervalSeconds  = 30 // Интервал для Telegram
	TelegramTimeoutSeconds   = 30 // Таймаут для Telegram
	TelegramFailureThreshold = 3  // Порог неудач для Telegram
	DatabaseMaxRequests      = 10 // Максимум запросов для БД
	DatabaseIntervalSeconds  = 60 // Интервал для БД
	DatabaseTimeoutSeconds   = 30 // Таймаут для БД
	DatabaseFailureThreshold = 5  // Порог неудач для БД
	MatcherMaxRequests       = 5  // Максимум запросов для Matcher
	MatcherIntervalSeconds   = 30 // Интервал для Matcher
	MatcherTimeoutSeconds    = 20 // Таймаут для Matcher
	MatcherFailureThreshold  = 3  // Порог неудач для Matcher
)

// Database Fallback Constants
// Used in: services/bot/internal/database/db.go.
const (
	FallbackLanguageID1 = 1
	FallbackLanguageID2 = 2
	FallbackLanguageID3 = 3
	FallbackLanguageID4 = 4

	FallbackInterestID1 = 1
	FallbackInterestID2 = 2
	FallbackInterestID3 = 3
	FallbackInterestID4 = 4
)

// Keyboard Symbols
// Used in: services/bot/internal/adapters/telegram/keyboard_helpers.go.
const (
	SymbolUnchecked = "☐ "
)

// Interest Profile Completion
// Used in: services/bot/internal/adapters/telegram/handlers/improved_interest_handlers.go.
const (
	ImprovedInterestProfileCompletionLevelComplete = 100 // Профиль полностью завершен
)

// Callback Data Constants
// Used in: services/bot/internal/adapters/telegram/keyboards.go.
const (
	CallbackBackToMainMenu     = "back_to_main_menu"
	CallbackBackToPreviousStep = "back_to_previous_step"
)

// Handler Limits
// Used in: services/bot/internal/adapters/telegram/handlers/profile_interest_handlers.go.
const (
	MinPartsForInterestCallback = 4 // Минимальное количество частей в callback data для интересов
)

// Profile Completion Levels
// Used in: services/bot/internal/adapters/telegram/handlers/profile_handlers.go,
//
//	services/bot/internal/adapters/telegram/handlers/new_interest_handlers.go,
//	services/bot/internal/adapters/telegram/handlers/improved_interest_handlers.go
const (
	ProfileCompletionLevelComplete = 100 // Профиль полностью завершен
)

// Language Fallback IDs
// Used in: services/bot/internal/adapters/telegram/handlers/keyboard_helpers.go.
const (
	LanguageIDEnglish = 1
	LanguageIDRussian = 2
	LanguageIDSpanish = 3
	LanguageIDChinese = 4
)

// Batch Loader Performance Constants
// Used in: services/bot/internal/database/batch_loader.go.
const (
	DefaultQueryTimeout = 30 * time.Second // Default timeout for SQL queries
	MaxBatchSize        = 1000             // Maximum batch size for data loading
)

// Interests Configuration Constants
// Used in: services/bot/internal/config/interests_config.go.
const (
	DefaultPrimaryPercentage    = 0.3  // 30% основных интересов от общего количества
	DefaultDirectoryPermissions = 0755 // Права доступа для директорий
	DefaultFilePermissions      = 0600 // Права доступа для файлов конфигурации
	DefaultMaxMatchesPerUser    = 10   // Максимальное количество совпадений на пользователя

	// Matching algorithm scores.
	DefaultPrimaryInterestScore    = 3 // Балл за основной интерес
	DefaultAdditionalInterestScore = 1 // Балл за дополнительный интерес
	DefaultMinCompatibilityScore   = 5 // Минимальный балл совместимости

	// Interest limits.
	DefaultMinPrimaryInterests   = 1 // Минимальное количество основных интересов
	DefaultMaxPrimaryInterests   = 5 // Максимальное количество основных интересов
	DefaultMaxPrimaryPerCategory = 2 // Максимум основных интересов на категорию

	// Category display orders.
	EntertainmentDisplayOrder = 1
	EducationDisplayOrder     = 2
	ActiveDisplayOrder        = 3
	CreativeDisplayOrder      = 4
	SocialDisplayOrder        = 5
)

// Main Application Constants
// Used in: services/bot/cmd/bot/main.go.
const (
	ForceShutdownTimeoutSeconds = 10 // in seconds
)

// =============================================================================
// LOCALIZATION KEYS (text message identifiers)
// =============================================================================

// Locale keys for menu and navigation.
const (
	LocaleMainMenuTitle           = "main_menu_title"
	LocaleEmptyProfileMessage     = "empty_profile_message"
	LocaleSetupProfileButton      = "setup_profile_button"
	LocaleBackToMain              = "back_to_main_menu"
	LocaleYourStatus              = "your_status"
	LocaleStatus                  = "status"
	LocaleState                   = "state"
	LocaleProfileCompletion       = "profile_completion"
	LocaleInterfaceLanguage       = "interface_language"
	LocaleProfileReset            = "profile_reset"
	LocaleChooseInterfaceLanguage = "choose_interface_language"
)

// Locale keys for profile management.
const (
	LocaleProfileFieldName          = "profile_field_name"
	LocaleProfileFieldUsername      = "profile_field_username"
	LocaleProfileFieldNative        = "profile_field_native"
	LocaleProfileFieldTarget        = "profile_field_target"
	LocaleProfileFieldInterests     = "profile_field_interests"
	LocaleProfileFieldStatus        = "profile_field_status"
	LocaleProfileFieldMemberSince   = "profile_field_member_since"
	LocaleProfileFieldAvailability  = "profile_field_availability"
	LocaleProfileFieldCommunication = "profile_field_communication"
	LocaleProfileShow               = "profile_show"
	LocaleProfileResetAsk           = "profile_reset_ask"
	LocaleProfileResetYes           = "profile_reset_yes"
	LocaleProfileResetNo            = "profile_reset_no"
	LocaleEditLanguages             = "edit_languages"
	LocaleEditNativeLang            = "edit_native_lang"
	LocaleEditTargetLang            = "edit_target_lang"
	LocaleEditLevel                 = "edit_level"
)

// Locale keys for interests management.
const (
	LocaleEditInterestsFromProfile        = "edit_interests_from_profile"
	LocaleChooseInterestCategory          = "choose_interest_category"
	LocaleEditInterestsInCategory         = "edit_interests_in_category"
	LocaleChooseInterests                 = "choose_interests"
	LocaleSelectAllInCategory             = "select_all_in_category"
	LocaleClearAllInCategory              = "clear_all_in_category"
	LocaleEditInterestsByCategory         = "edit_interests_by_category"
	LocaleEditPrimaryInterests            = "edit_primary_interests"
	LocaleChoosePrimaryInterests          = "choose_primary_interests"
	LocaleChoosePrimaryInterestsDynamic   = "choose_primary_interests_dynamic"
	LocaleChoosePrimaryInterestsRemaining = "choose_primary_interests_remaining"
	LocaleMaxPrimaryInterestsReached      = "max_primary_interests_reached"
	LocaleInterestsUpdatedSuccessfully    = "interests_updated_successfully"
	LocaleTotalInterests                  = "total_interests"
	LocalePrimaryInterestsLabel           = "primary_interests_label"
	LocaleAdditionalInterestsLabel        = "additional_interests_label"
	LocaleInterestsSelectionComplete      = "interests_selection_complete"
	LocaleBackToCategories                = "back_to_categories"
	LocaleBackToInterests                 = "back_to_interests"
	LocaleBackToEditMenu                  = "back_to_edit_menu"
	LocaleSaveChanges                     = "save_changes"
	LocaleCancelEdit                      = "cancel_edit"
	LocaleUndoLastChange                  = "undo_last_change"
)

// Locale keys for language management.
const (
	LocaleChooseNativeLanguage     = "choose_native_language"
	LocaleChooseTargetLanguage     = "choose_target_language"
	LocaleLanguagesContinueFilling = "languages_continue_filling"
	LocaleLanguagesReselect        = "languages_reselect"
	LocaleBackToLanguageLevel      = "back_to_language_level"
)

// Locale keys for feedback management.
const (
	LocaleFeedbackText        = "feedback_text"
	LocaleFeedbackHelpTitle   = "feedback_help_title"
	LocaleFeedbackHelpContent = "feedback_help_content"
	LocaleFeedbackBackToMain  = "feedback_back_to_main"
	LocaleFeedbackHelp        = "feedback_help"
)

// Locale keys for welcome and general messages.
const (
	LocaleWelcomeMessage = "welcome_message"
	LocaleUnknownCommand = "unknown_command"
	LocaleUseMenuAbove   = "use_menu_above"
)

// Locale keys for time and communication preferences.
const (
	LocaleTimeWeekdays    = "time_weekdays"
	LocaleTimeWeekends    = "time_weekends"
	LocaleTimeAny         = "time_any"
	LocaleTimeMorning     = "time_morning"
	LocaleTimeDay         = "time_day"
	LocaleTimeEvening     = "time_evening"
	LocaleTimeLate        = "time_late"
	LocaleCommText        = "comm_text"
	LocaleCommVoice       = "comm_voice"
	LocaleCommAudio       = "comm_audio"
	LocaleCommVideo       = "comm_video"
	LocaleCommMeet        = "comm_meet"
	LocaleFreqSpontaneous = "freq_spontaneous"
	LocaleFreqWeekly      = "freq_weekly"
	LocaleFreqDaily       = "freq_daily"
	LocaleFreqIntensive   = "freq_intensive"
)

// Locale keys for user status.
const (
	LocaleStatusNew     = "status_new"
	LocaleStatusFilling = "status_filling"
	LocaleStatusActive  = "status_active"
	LocaleStatusPaused  = "status_paused"
)

// Locale keys for interest categories and interests
// Categories.
const (
	LocaleCategoryEntertainment = "category_entertainment"
	LocaleCategoryEducation     = "category_education"
	LocaleCategoryActive        = "category_active"
	LocaleCategoryCreative      = "category_creative"
	LocaleCategorySocial        = "category_social"
)

// Interests (examples - can be extended).
const (
	LocaleInterestMovies = "interest_movies"
	LocaleInterestMusic  = "interest_music"
	LocaleInterestSports = "interest_sports"
	LocaleInterestTravel = "interest_travel"
)

// Locale keys for admin functionality.
const (
	LocaleShowActiveFeedbacks    = "show_active_feedbacks"
	LocaleShowArchiveFeedbacks   = "show_archive_feedbacks"
	LocaleShowAllFeedbacks       = "show_all_feedbacks"
	LocaleBackToFeedbackStats    = "back_to_feedback_stats"
	LocaleBackToActiveFeedbacks  = "back_to_active_feedbacks"
	LocaleBackToArchiveFeedbacks = "back_to_archive_feedbacks"
	LocaleBackToAllFeedbacks     = "back_to_all_feedbacks"
)

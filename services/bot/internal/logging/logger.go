package logging

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger интерфейс для логирования
type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	Fatal(msg string, fields ...Field)
	With(fields ...Field) Logger
	WithContext(ctx context.Context) Logger
}

// Field представляет поле лога
type Field struct {
	Key   string
	Value interface{}
}

// ZapLogger реализация логгера на основе zap
type ZapLogger struct {
	logger *zap.Logger
}

// NewLogger создает новый логгер
func NewLogger(level string, format string) (Logger, error) {
	var config zap.Config

	if format == "json" {
		config = zap.NewProductionConfig()
	} else {
		config = zap.NewDevelopmentConfig()
	}

	// Настройка уровня логирования
	switch strings.ToLower(level) {
	case "debug":
		config.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	case "info":
		config.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	case "warn":
		config.Level = zap.NewAtomicLevelAt(zapcore.WarnLevel)
	case "error":
		config.Level = zap.NewAtomicLevelAt(zapcore.ErrorLevel)
	default:
		config.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	}

	// Настройка кодирования времени
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	// Настройка кодирования уровня
	config.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	// Настройка кодирования caller
	config.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	logger, err := config.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	return &ZapLogger{logger: logger}, nil
}

// NewProductionLogger создает production логгер
func NewProductionLogger() (Logger, error) {
	config := zap.NewProductionConfig()
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	logger, err := config.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to create production logger: %w", err)
	}

	return &ZapLogger{logger: logger}, nil
}

// NewDevelopmentLogger создает development логгер
func NewDevelopmentLogger() (Logger, error) {
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	logger, err := config.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to create development logger: %w", err)
	}

	return &ZapLogger{logger: logger}, nil
}

// Debug логирует debug сообщение
func (l *ZapLogger) Debug(msg string, fields ...Field) {
	l.logger.Debug(msg, l.fieldsToZap(fields)...)
}

// Info логирует info сообщение
func (l *ZapLogger) Info(msg string, fields ...Field) {
	l.logger.Info(msg, l.fieldsToZap(fields)...)
}

// Warn логирует warning сообщение
func (l *ZapLogger) Warn(msg string, fields ...Field) {
	l.logger.Warn(msg, l.fieldsToZap(fields)...)
}

// Error логирует error сообщение
func (l *ZapLogger) Error(msg string, fields ...Field) {
	l.logger.Error(msg, l.fieldsToZap(fields)...)
}

// Fatal логирует fatal сообщение и завершает программу
func (l *ZapLogger) Fatal(msg string, fields ...Field) {
	l.logger.Fatal(msg, l.fieldsToZap(fields)...)
}

// With создает новый логгер с дополнительными полями
func (l *ZapLogger) With(fields ...Field) Logger {
	return &ZapLogger{
		logger: l.logger.With(l.fieldsToZap(fields)...),
	}
}

// WithContext создает новый логгер с контекстом
func (l *ZapLogger) WithContext(ctx context.Context) Logger {
	fields := l.extractContextFields(ctx)
	return &ZapLogger{
		logger: l.logger.With(l.fieldsToZap(fields)...),
	}
}

// fieldsToZap конвертирует поля в zap поля
func (l *ZapLogger) fieldsToZap(fields []Field) []zap.Field {
	zapFields := make([]zap.Field, len(fields))
	for i, field := range fields {
		zapFields[i] = zap.Any(field.Key, field.Value)
	}
	return zapFields
}

// extractContextFields извлекает поля из контекста
func (l *ZapLogger) extractContextFields(ctx context.Context) []Field {
	fields := make([]Field, 0)

	// Извлекаем request ID если есть
	if requestID := ctx.Value("request_id"); requestID != nil {
		fields = append(fields, Field{Key: "request_id", Value: requestID})
	}

	// Извлекаем user ID если есть
	if userID := ctx.Value("user_id"); userID != nil {
		fields = append(fields, Field{Key: "user_id", Value: userID})
	}

	// Извлекаем trace ID если есть
	if traceID := ctx.Value("trace_id"); traceID != nil {
		fields = append(fields, Field{Key: "trace_id", Value: traceID})
	}

	return fields
}

// Sync синхронизирует логгер
func (l *ZapLogger) Sync() error {
	return l.logger.Sync()
}

// Helper функции для создания полей

// String создает строковое поле
func String(key, value string) Field {
	return Field{Key: key, Value: value}
}

// Int создает целочисленное поле
func Int(key string, value int) Field {
	return Field{Key: key, Value: value}
}

// Int64 создает 64-битное целочисленное поле
func Int64(key string, value int64) Field {
	return Field{Key: key, Value: value}
}

// Float64 создает поле с плавающей точкой
func Float64(key string, value float64) Field {
	return Field{Key: key, Value: value}
}

// Bool создает булево поле
func Bool(key string, value bool) Field {
	return Field{Key: key, Value: value}
}

// Duration создает поле с продолжительностью
func Duration(key string, value time.Duration) Field {
	return Field{Key: key, Value: value}
}

// Time создает поле с временем
func Time(key string, value time.Time) Field {
	return Field{Key: key, Value: value}
}

// ErrorField создает поле с ошибкой
func ErrorField(err error) Field {
	return Field{Key: "error", Value: err.Error()}
}

// Any создает поле с любым значением
func Any(key string, value interface{}) Field {
	return Field{Key: key, Value: value}
}

// RequestID создает поле с ID запроса
func RequestID(id string) Field {
	return Field{Key: "request_id", Value: id}
}

// UserID создает поле с ID пользователя
func UserID(id int64) Field {
	return Field{Key: "user_id", Value: id}
}

// TraceID создает поле с ID трассировки
func TraceID(id string) Field {
	return Field{Key: "trace_id", Value: id}
}

// Component создает поле с компонентом
func Component(name string) Field {
	return Field{Key: "component", Value: name}
}

// Operation создает поле с операцией
func Operation(name string) Field {
	return Field{Key: "operation", Value: name}
}

// DatabaseLogger специализированный логгер для БД операций
type DatabaseLogger struct {
	Logger
}

// NewDatabaseLogger создает новый логгер для БД
func NewDatabaseLogger(logger Logger) *DatabaseLogger {
	return &DatabaseLogger{
		Logger: logger.With(Component("database")),
	}
}

// LogQuery логирует SQL запрос
func (l *DatabaseLogger) LogQuery(operation, table string, duration time.Duration, err error) {
	fields := []Field{
		Operation(operation),
		String("table", table),
		Duration("duration", duration),
	}

	if err != nil {
		fields = append(fields, ErrorField(err))
		l.Error("Database query failed", fields...)
	} else {
		l.Debug("Database query executed", fields...)
	}
}

// LogTransaction логирует транзакцию
func (l *DatabaseLogger) LogTransaction(operation string, duration time.Duration, err error) {
	fields := []Field{
		Operation(operation),
		Duration("duration", duration),
	}

	if err != nil {
		fields = append(fields, ErrorField(err))
		l.Error("Database transaction failed", fields...)
	} else {
		l.Info("Database transaction completed", fields...)
	}
}

// CacheLogger специализированный логгер для кэша
type CacheLogger struct {
	Logger
}

// NewCacheLogger создает новый логгер для кэша
func NewCacheLogger(logger Logger) *CacheLogger {
	return &CacheLogger{
		Logger: logger.With(Component("cache")),
	}
}

// LogCacheOperation логирует операцию с кэшем
func (l *CacheLogger) LogCacheOperation(operation, cacheType, key string, hit bool, duration time.Duration, err error) {
	fields := []Field{
		Operation(operation),
		String("cache_type", cacheType),
		String("key", key),
		Bool("hit", hit),
		Duration("duration", duration),
	}

	if err != nil {
		fields = append(fields, ErrorField(err))
		l.Error("Cache operation failed", fields...)
	} else {
		l.Debug("Cache operation completed", fields...)
	}
}

// BusinessLogger специализированный логгер для бизнес-логики
type BusinessLogger struct {
	Logger
}

// NewBusinessLogger создает новый логгер для бизнес-логики
func NewBusinessLogger(logger Logger) *BusinessLogger {
	return &BusinessLogger{
		Logger: logger.With(Component("business")),
	}
}

// LogUserAction логирует действие пользователя
func (l *BusinessLogger) LogUserAction(userID int64, action string, details map[string]interface{}) {
	fields := []Field{
		UserID(userID),
		Operation(action),
	}

	for key, value := range details {
		fields = append(fields, Any(key, value))
	}

	l.Info("User action", fields...)
}

// LogProfileUpdate логирует обновление профиля
func (l *BusinessLogger) LogProfileUpdate(userID int64, fields []string) {
	l.Info("Profile updated",
		UserID(userID),
		Any("updated_fields", fields),
	)
}

// LogFeedbackSubmission логирует отправку отзыва
func (l *BusinessLogger) LogFeedbackSubmission(userID int64, hasContactInfo bool) {
	l.Info("Feedback submitted",
		UserID(userID),
		Bool("has_contact_info", hasContactInfo),
	)
}

// Global logger instance
var globalLogger Logger

// InitGlobalLogger инициализирует глобальный логгер
func InitGlobalLogger(level, format string) error {
	logger, err := NewLogger(level, format)
	if err != nil {
		return err
	}
	globalLogger = logger
	return nil
}

// GetGlobalLogger возвращает глобальный логгер
func GetGlobalLogger() Logger {
	if globalLogger == nil {
		// Fallback на простой логгер
		logger, _ := NewDevelopmentLogger()
		return logger
	}
	return globalLogger
}

// SetGlobalLogger устанавливает глобальный логгер
func SetGlobalLogger(logger Logger) {
	globalLogger = logger
}

// Convenience functions для глобального логгера

// Debug логирует debug сообщение через глобальный логгер
func Debug(msg string, fields ...Field) {
	GetGlobalLogger().Debug(msg, fields...)
}

// Info логирует info сообщение через глобальный логгер
func Info(msg string, fields ...Field) {
	GetGlobalLogger().Info(msg, fields...)
}

// Warn логирует warning сообщение через глобальный логгер
func Warn(msg string, fields ...Field) {
	GetGlobalLogger().Warn(msg, fields...)
}

// Error логирует error сообщение через глобальный логгер
func Error(msg string, fields ...Field) {
	GetGlobalLogger().Error(msg, fields...)
}

// Fatal логирует fatal сообщение через глобальный логгер
func Fatal(msg string, fields ...Field) {
	GetGlobalLogger().Fatal(msg, fields...)
}

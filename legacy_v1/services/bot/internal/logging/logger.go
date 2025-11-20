// Package logging provides structured logging functionality.
package logging

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"
)

// LogLevel определяет уровень логирования.
type LogLevel int

const (
	// DEBUG represents debug log level for detailed information.
	DEBUG LogLevel = iota
	// INFO represents info log level for general information.
	INFO
	// WARN represents warning log level for non-critical issues.
	WARN
	// ERROR represents error log level for critical issues.
	ERROR
)

// String возвращает строковое представление уровня логирования.
func (ll LogLevel) String() string {
	switch ll {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// ParseLogLevel парсит строку в LogLevel.
func ParseLogLevel(level string) LogLevel {
	switch level {
	case "DEBUG":
		return DEBUG
	case "INFO":
		return INFO
	case "WARN":
		return WARN
	case "ERROR":
		return ERROR
	default:
		return INFO
	}
}

// LogEntry представляет структурированную запись лога.
type LogEntry struct {
	Timestamp time.Time              `json:"timestamp"`
	Level     LogLevel               `json:"level"`
	Message   string                 `json:"message"`
	RequestID string                 `json:"requestId,omitempty"`
	UserID    int64                  `json:"userId,omitempty"`
	ChatID    int64                  `json:"chatId,omitempty"`
	Operation string                 `json:"operation,omitempty"`
	Component string                 `json:"component,omitempty"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
	Error     string                 `json:"error,omitempty"`
}

// Logger предоставляет структурированное логирование.
type Logger struct {
	level     LogLevel
	component string
}

// NewLogger создает новый логгер.
func NewLogger(level LogLevel, component string) *Logger {
	return &Logger{
		level:     level,
		component: component,
	}
}

// NewLoggerFromEnv создает логгер из переменных окружения.
func NewLoggerFromEnv(component string) *Logger {
	level := ParseLogLevel(os.Getenv("LOG_LEVEL"))

	return NewLogger(level, component)
}

// Debug логирует сообщение уровня DEBUG.
func (l *Logger) Debug(message string, fields ...map[string]interface{}) {
	l.log(DEBUG, message, fields...)
}

// Info логирует сообщение уровня INFO.
func (l *Logger) Info(message string, fields ...map[string]interface{}) {
	l.log(INFO, message, fields...)
}

// Warn логирует сообщение уровня WARN.
func (l *Logger) Warn(message string, fields ...map[string]interface{}) {
	l.log(WARN, message, fields...)
}

// Error логирует сообщение уровня ERROR.
func (l *Logger) Error(message string, fields ...map[string]interface{}) {
	l.log(ERROR, message, fields...)
}

// DebugWithContext логирует сообщение уровня DEBUG с контекстом.
func (l *Logger) DebugWithContext(
	message string,
	requestID string,
	userID,
	chatID int64,
	operation string,
	fields ...map[string]interface{},
) {
	entry := l.createLogEntry(DEBUG, message, requestID, userID, chatID, operation, fields...)
	l.writeLog(entry)
}

// InfoWithContext логирует сообщение уровня INFO с контекстом.
func (l *Logger) InfoWithContext(
	message string,
	requestID string,
	userID,
	chatID int64,
	operation string,
	fields ...map[string]interface{},
) {
	entry := l.createLogEntry(INFO, message, requestID, userID, chatID, operation, fields...)
	l.writeLog(entry)
}

// WarnWithContext логирует сообщение уровня WARN с контекстом.
func (l *Logger) WarnWithContext(
	message string,
	requestID string,
	userID,
	chatID int64,
	operation string,
	fields ...map[string]interface{},
) {
	entry := l.createLogEntry(WARN, message, requestID, userID, chatID, operation, fields...)
	l.writeLog(entry)
}

// ErrorWithContext логирует сообщение уровня ERROR с контекстом.
func (l *Logger) ErrorWithContext(
	message string,
	requestID string,
	userID,
	chatID int64,
	operation string,
	fields ...map[string]interface{},
) {
	entry := l.createLogEntry(ERROR, message, requestID, userID, chatID, operation, fields...)
	l.writeLog(entry)
}

// log внутренний метод для логирования.
func (l *Logger) log(level LogLevel, message string, fields ...map[string]interface{}) {
	if level < l.level {
		return
	}

	entry := l.createLogEntry(level, message, "", 0, 0, "", fields...)
	l.writeLog(entry)
}

// createLogEntry создает структурированную запись лога.
func (l *Logger) createLogEntry(level LogLevel, message string, requestID string, userID, chatID int64, operation string, fields ...map[string]interface{}) *LogEntry {
	entry := &LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   message,
		Component: l.component,
	}

	if requestID != "" {
		entry.RequestID = requestID
	}

	if userID > 0 {
		entry.UserID = userID
	}

	if chatID > 0 {
		entry.ChatID = chatID
	}

	if operation != "" {
		entry.Operation = operation
	}

	// Объединяем все поля
	if len(fields) > 0 {
		entry.Fields = make(map[string]interface{})

		for _, fieldMap := range fields {
			for key, value := range fieldMap {
				entry.Fields[key] = value
			}
		}
	}

	return entry
}

// writeLog записывает лог в формате JSON.
func (l *Logger) writeLog(entry *LogEntry) {
	if entry.Level < l.level {
		return
	}

	jsonData, err := json.Marshal(entry)
	if err != nil {
		// Fallback на простое логирование
		log.Printf("[%s] %s: %s", entry.Level.String(), entry.Component, entry.Message)

		return
	}

	// Выводим в stdout для структурированного логирования
	fmt.Println(string(jsonData))
}

// SetLevel устанавливает уровень логирования.
func (l *Logger) SetLevel(level LogLevel) {
	l.level = level
}

// GetLevel возвращает текущий уровень логирования.
func (l *Logger) GetLevel() LogLevel {
	return l.level
}

// WithFields creates a new logger with additional fields.
func (l *Logger) WithFields(_ map[string]interface{}) *Logger {
	return &Logger{
		level:     l.level,
		component: l.component,
	}
}

// WithComponent создает новый логгер с указанным компонентом.
func (l *Logger) WithComponent(component string) *Logger {
	return &Logger{
		level:     l.level,
		component: component,
	}
}

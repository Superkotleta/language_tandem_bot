# 📝 План реализации Audit Logging для Compliance

## 📋 Обзор

Реализация системы аудита для соответствия требованиям compliance и отслеживания всех действий пользователей и администраторов в системе.

## 🎯 Цели

- **Compliance**: Соответствие требованиям GDPR, SOX, HIPAA
- **Security**: Отслеживание подозрительной активности
- **Accountability**: Ответственность за действия пользователей
- **Forensics**: Расследование инцидентов безопасности

## 🏗️ Архитектура решения

### 1. **Audit Event Model**

```go
// internal/audit/event.go
type AuditEvent struct {
    ID          string                 `json:"id"`
    Timestamp   time.Time              `json:"timestamp"`
    UserID      int64                  `json:"user_id"`
    UserType    string                 `json:"user_type"` // "user", "admin", "system"
    Action      string                 `json:"action"`
    Resource    string                 `json:"resource"`
    ResourceID  string                 `json:"resource_id"`
    IPAddress   string                 `json:"ip_address"`
    UserAgent   string                 `json:"user_agent"`
    SessionID   string                 `json:"session_id"`
    Result      string                 `json:"result"` // "success", "failure", "error"
    Details     map[string]interface{} `json:"details"`
    Metadata    map[string]interface{} `json:"metadata"`
}
```

### 2. **Audit Logger Interface**

```go
// internal/audit/logger.go
type AuditLogger interface {
    LogEvent(event *AuditEvent) error
    LogUserAction(userID int64, action, resource string, details map[string]interface{}) error
    LogAdminAction(adminID int64, action, resource string, details map[string]interface{}) error
    LogSystemEvent(action, resource string, details map[string]interface{}) error
    LogSecurityEvent(eventType, description string, details map[string]interface{}) error
}
```

### 3. **Event Categories**

```go
const (
    // User Actions
    EventUserLogin           = "user.login"
    EventUserLogout          = "user.logout"
    EventUserRegistration    = "user.registration"
    EventUserProfileUpdate   = "user.profile.update"
    EventUserInterestUpdate  = "user.interest.update"
    EventUserLanguageUpdate  = "user.language.update"
    
    // Admin Actions
    EventAdminLogin          = "admin.login"
    EventAdminLogout         = "admin.logout"
    EventAdminUserView       = "admin.user.view"
    EventAdminUserEdit       = "admin.user.edit"
    EventAdminUserDelete     = "admin.user.delete"
    EventAdminFeedbackView   = "admin.feedback.view"
    EventAdminFeedbackProcess = "admin.feedback.process"
    
    // System Events
    EventSystemStartup       = "system.startup"
    EventSystemShutdown      = "system.shutdown"
    EventSystemError         = "system.error"
    EventSystemMaintenance   = "system.maintenance"
    
    // Security Events
    EventSecurityFailedLogin = "security.failed_login"
    EventSecuritySuspiciousActivity = "security.suspicious_activity"
    EventSecurityRateLimitExceeded = "security.rate_limit_exceeded"
    EventSecurityUnauthorizedAccess = "security.unauthorized_access"
)
```

## 📁 Структура файлов

```
services/bot/internal/audit/
├── event.go              # Audit event model
├── logger.go             # Audit logger interface
├── storage.go            # Audit storage
├── middleware.go         # Audit middleware
├── compliance.go         # Compliance helpers
├── retention.go          # Data retention policies
└── audit_test.go         # Tests
```

## 🔧 Реализация

### Phase 1: Core Audit Infrastructure (1 неделя)

#### 1.1 Event Model
- [ ] Определение структуры AuditEvent
- [ ] Валидация событий
- [ ] Сериализация в JSON
- [ ] Уникальные ID для событий

#### 1.2 Audit Logger
- [ ] Базовый AuditLogger interface
- [ ] Синхронное логирование
- [ ] Асинхронное логирование (queue)
- [ ] Batch processing для производительности

#### 1.3 Storage Backend
- [ ] PostgreSQL storage для audit events
- [ ] Индексы для быстрого поиска
- [ ] Партиционирование по датам
- [ ] Compression для старых данных

### Phase 2: Event Collection (1 неделя)

#### 2.1 Middleware Integration
```go
// internal/middleware/audit_middleware.go
func AuditMiddleware(auditLogger AuditLogger) gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        
        // Extract user info
        userID := getUserIDFromContext(c)
        userType := getUserTypeFromContext(c)
        
        // Process request
        c.Next()
        
        // Log event
        event := &AuditEvent{
            UserID:     userID,
            UserType:   userType,
            Action:     c.Request.Method + " " + c.Request.URL.Path,
            Resource:   extractResource(c),
            IPAddress:  c.ClientIP(),
            UserAgent:  c.Request.UserAgent(),
            Result:     getResultFromStatus(c.Writer.Status()),
            Duration:   time.Since(start),
        }
        
        auditLogger.LogEvent(event)
    }
}
```

#### 2.2 Business Logic Integration
- [ ] User registration/login events
- [ ] Profile update events
- [ ] Interest selection events
- [ ] Admin action events
- [ ] System events

#### 2.3 Security Events
- [ ] Failed login attempts
- [ ] Rate limiting violations
- [ ] Unauthorized access attempts
- [ ] Suspicious activity patterns

### Phase 3: Compliance Features (1 неделя)

#### 3.1 Data Retention Policies
```go
// internal/audit/retention.go
type RetentionPolicy struct {
    EventType    string        `json:"event_type"`
    RetentionPeriod time.Duration `json:"retention_period"`
    ArchiveAfter time.Duration `json:"archive_after"`
    DeleteAfter  time.Duration `json:"delete_after"`
}

// Default policies
var DefaultRetentionPolicies = []RetentionPolicy{
    {EventType: "user.*", RetentionPeriod: 2 * 365 * 24 * time.Hour}, // 2 years
    {EventType: "admin.*", RetentionPeriod: 7 * 365 * 24 * time.Hour}, // 7 years
    {EventType: "security.*", RetentionPeriod: 7 * 365 * 24 * time.Hour}, // 7 years
    {EventType: "system.*", RetentionPeriod: 1 * 365 * 24 * time.Hour}, // 1 year
}
```

#### 3.2 Data Anonymization
- [ ] PII detection в audit events
- [ ] Автоматическая анонимизация
- [ ] Конфигурируемые правила
- [ ] GDPR compliance

#### 3.3 Export and Reporting
- [ ] Export audit logs в различных форматах
- [ ] Compliance reports
- [ ] Security incident reports
- [ ] User activity reports

### Phase 4: Advanced Features (1 неделя)

#### 4.1 Real-time Monitoring
- [ ] WebSocket для real-time audit events
- [ ] Dashboard для мониторинга
- [ ] Alerts для подозрительной активности
- [ ] Integration с SIEM системами

#### 4.2 Analytics and Insights
- [ ] User behavior analytics
- [ ] Security threat detection
- [ ] Performance impact analysis
- [ ] Compliance metrics

## 🗄️ Database Schema

### Audit Events Table
```sql
CREATE TABLE audit_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    user_id BIGINT,
    user_type VARCHAR(20) NOT NULL,
    action VARCHAR(100) NOT NULL,
    resource VARCHAR(100),
    resource_id VARCHAR(100),
    ip_address INET,
    user_agent TEXT,
    session_id VARCHAR(100),
    result VARCHAR(20) NOT NULL,
    details JSONB,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Индексы для производительности
CREATE INDEX idx_audit_events_timestamp ON audit_events(timestamp);
CREATE INDEX idx_audit_events_user_id ON audit_events(user_id);
CREATE INDEX idx_audit_events_action ON audit_events(action);
CREATE INDEX idx_audit_events_resource ON audit_events(resource);
CREATE INDEX idx_audit_events_result ON audit_events(result);

-- Партиционирование по месяцам
CREATE TABLE audit_events_y2025m01 PARTITION OF audit_events
    FOR VALUES FROM ('2025-01-01') TO ('2025-02-01');
```

### Audit Configuration Table
```sql
CREATE TABLE audit_config (
    id SERIAL PRIMARY KEY,
    event_type VARCHAR(100) NOT NULL,
    retention_days INTEGER NOT NULL,
    archive_days INTEGER,
    delete_days INTEGER,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
```

## 🔧 Конфигурация

### Environment Variables
```bash
# Audit Configuration
AUDIT_ENABLED=true
AUDIT_LOG_LEVEL=info
AUDIT_RETENTION_DAYS=2555  # 7 years
AUDIT_ARCHIVE_DAYS=365     # 1 year
AUDIT_DELETE_DAYS=2555     # 7 years

# Storage
AUDIT_STORAGE_TYPE=postgresql
AUDIT_STORAGE_URL=postgres://user:pass@localhost/audit_db
AUDIT_BATCH_SIZE=100
AUDIT_FLUSH_INTERVAL=5s

# Security
AUDIT_ENCRYPT_DATA=true
AUDIT_ENCRYPTION_KEY=your-encryption-key
AUDIT_ANONYMIZE_PII=true
```

### Config Structure
```go
type AuditConfig struct {
    Enabled        bool          `json:"enabled"`
    LogLevel       string        `json:"log_level"`
    RetentionDays  int           `json:"retention_days"`
    ArchiveDays    int           `json:"archive_days"`
    DeleteDays     int           `json:"delete_days"`
    Storage        StorageConfig `json:"storage"`
    Security       SecurityConfig `json:"security"`
}

type StorageConfig struct {
    Type           string        `json:"type"`
    URL            string        `json:"url"`
    BatchSize      int           `json:"batch_size"`
    FlushInterval  time.Duration `json:"flush_interval"`
}

type SecurityConfig struct {
    EncryptData    bool   `json:"encrypt_data"`
    EncryptionKey  string `json:"encryption_key"`
    AnonymizePII   bool   `json:"anonymize_pii"`
}
```

## 🧪 Тестирование

### Unit Tests
- [ ] Event creation и validation
- [ ] Logger functionality
- [ ] Storage operations
- [ ] Retention policies

### Integration Tests
- [ ] End-to-end audit flow
- [ ] Middleware integration
- [ ] Database operations
- [ ] Performance testing

### Compliance Tests
- [ ] GDPR compliance
- [ ] Data retention
- [ ] PII anonymization
- [ ] Export functionality

## 📊 Мониторинг и метрики

### Audit Metrics
- [ ] Количество audit events в секунду
- [ ] Размер audit database
- [ ] Время обработки events
- [ ] Ошибки в audit logging

### Compliance Metrics
- [ ] Retention policy compliance
- [ ] PII detection rate
- [ ] Export success rate
- [ ] Security event frequency

### Performance Metrics
- [ ] Audit logging overhead
- [ ] Database query performance
- [ ] Storage utilization
- [ ] Memory usage

## 🚀 Deployment

### Production Considerations
- [ ] Dedicated audit database
- [ ] Backup и recovery procedures
- [ ] Monitoring и alerting
- [ ] Compliance reporting

### Security Considerations
- [ ] Encryption в transit и at rest
- [ ] Access control для audit data
- [ ] Tamper-proof storage
- [ ] Regular security audits

## 📈 Success Metrics

- **Compliance**: 100% соответствие требованиям
- **Performance**: <5ms overhead для audit logging
- **Reliability**: 99.99% audit event capture rate
- **Security**: 0 unauthorized access к audit data

## 🔄 Timeline

| Phase | Duration | Deliverables |
|-------|----------|--------------|
| **Phase 1** | 1 неделя | Core audit infrastructure |
| **Phase 2** | 1 неделя | Event collection |
| **Phase 3** | 1 неделя | Compliance features |
| **Phase 4** | 1 неделя | Advanced features |

**Total: 4 недели**

## 💰 Ресурсы

- **Backend Developer**: 1 FTE
- **Compliance Expert**: 0.3 FTE
- **Security Review**: 0.2 FTE
- **Testing**: 0.3 FTE
- **Documentation**: 0.2 FTE

**Total: 2 FTE (4 недели)**

## 📋 Compliance Requirements

### GDPR Compliance
- [ ] Right to be forgotten (data deletion)
- [ ] Data portability (export user data)
- [ ] Consent tracking
- [ ] Data minimization

### SOX Compliance
- [ ] Financial data access logging
- [ ] Change tracking
- [ ] Segregation of duties
- [ ] Management oversight

### HIPAA Compliance (если применимо)
- [ ] PHI access logging
- [ ] Encryption requirements
- [ ] Access controls
- [ ] Breach notification

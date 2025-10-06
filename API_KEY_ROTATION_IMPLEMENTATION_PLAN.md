# 🔑 План реализации ротации API ключей

## 📋 Обзор

Реализация системы автоматической ротации API ключей для повышения безопасности и соответствия best practices.

## 🎯 Цели

- **Безопасность**: Регулярная смена ключей для минимизации рисков
- **Compliance**: Соответствие требованиям безопасности
- **Automation**: Автоматизация процесса ротации
- **Monitoring**: Отслеживание использования ключей

## 🏗️ Архитектура решения

### 1. **API Key Model**

```go
// internal/security/api_key.go
type APIKey struct {
    ID          string    `json:"id"`
    Name        string    `json:"name"`
    KeyHash     string    `json:"key_hash"`
    KeyPrefix   string    `json:"key_prefix"` // Первые 8 символов для идентификации
    UserID      int64     `json:"user_id"`
    UserType    string    `json:"user_type"` // "admin", "service", "integration"
    Permissions []string  `json:"permissions"`
    Scopes      []string  `json:"scopes"`
    IsActive    bool      `json:"is_active"`
    CreatedAt   time.Time `json:"created_at"`
    ExpiresAt   time.Time `json:"expires_at"`
    LastUsedAt  time.Time `json:"last_used_at"`
    RotatedAt   time.Time `json:"rotated_at"`
    Metadata    map[string]interface{} `json:"metadata"`
}
```

### 2. **Key Rotation Manager**

```go
// internal/security/key_rotation.go
type KeyRotationManager struct {
    keyStore    KeyStore
    scheduler   Scheduler
    notifier    Notifier
    config      RotationConfig
}

type RotationConfig struct {
    DefaultRotationDays int           `json:"default_rotation_days"`
    WarningDays         int           `json:"warning_days"`
    GracePeriodDays     int           `json:"grace_period_days"`
    AutoRotation        bool          `json:"auto_rotation"`
    NotificationChannels []string     `json:"notification_channels"`
}
```

### 3. **Key Store Interface**

```go
// internal/security/key_store.go
type KeyStore interface {
    CreateKey(key *APIKey) error
    GetKeyByID(id string) (*APIKey, error)
    GetKeyByHash(hash string) (*APIKey, error)
    UpdateKey(key *APIKey) error
    DeleteKey(id string) error
    ListKeys(userID int64) ([]*APIKey, error)
    ListExpiringKeys(days int) ([]*APIKey, error)
    RotateKey(id string) (*APIKey, error)
    DeactivateKey(id string) error
}
```

## 📁 Структура файлов

```
services/bot/internal/security/
├── api_key.go           # API key model
├── key_store.go         # Key storage interface
├── key_rotation.go      # Rotation logic
├── key_generator.go     # Key generation
├── key_validator.go     # Key validation
├── key_middleware.go    # Middleware for key validation
├── notifications.go     # Rotation notifications
└── security_test.go     # Tests
```

## 🔧 Реализация

### Phase 1: Key Management Infrastructure (1 неделя)

#### 1.1 API Key Model
- [ ] Определение структуры APIKey
- [ ] Генерация безопасных ключей
- [ ] Хеширование ключей (bcrypt/scrypt)
- [ ] Валидация ключей

#### 1.2 Key Store Implementation
```go
// internal/security/key_store_impl.go
type PostgreSQLKeyStore struct {
    db *sql.DB
}

func (s *PostgreSQLKeyStore) CreateKey(key *APIKey) error {
    // Генерируем уникальный ключ
    keyValue := generateSecureKey()
    keyHash := hashKey(keyValue)
    keyPrefix := keyValue[:8]
    
    key.ID = generateKeyID()
    key.KeyHash = keyHash
    key.KeyPrefix = keyPrefix
    key.CreatedAt = time.Now()
    key.ExpiresAt = time.Now().AddDate(0, 0, s.config.DefaultRotationDays)
    
    return s.insertKey(key)
}
```

#### 1.3 Key Generation
- [ ] Криптографически стойкая генерация
- [ ] Уникальность ключей
- [ ] Форматирование ключей (prefix-suffix)
- [ ] Валидация сложности

### Phase 2: Rotation Logic (1 неделя)

#### 2.1 Rotation Scheduler
```go
// internal/security/rotation_scheduler.go
type RotationScheduler struct {
    manager *KeyRotationManager
    ticker  *time.Ticker
}

func (s *RotationScheduler) Start() {
    s.ticker = time.NewTicker(24 * time.Hour)
    go func() {
        for range s.ticker.C {
            s.checkAndRotateKeys()
        }
    }()
}

func (s *RotationScheduler) checkAndRotateKeys() {
    // Получаем ключи, которые скоро истекают
    expiringKeys := s.manager.GetExpiringKeys(s.config.WarningDays)
    
    for _, key := range expiringKeys {
        if s.config.AutoRotation {
            s.manager.RotateKey(key.ID)
        } else {
            s.manager.NotifyExpiration(key)
        }
    }
}
```

#### 2.2 Rotation Process
- [ ] Создание нового ключа
- [ ] Grace period для старого ключа
- [ ] Уведомления пользователей
- [ ] Деактивация старого ключа

#### 2.3 Notification System
```go
// internal/security/notifications.go
type NotificationService struct {
    emailNotifier EmailNotifier
    slackNotifier SlackNotifier
    webhookNotifier WebhookNotifier
}

func (n *NotificationService) NotifyKeyExpiration(key *APIKey, daysLeft int) error {
    message := fmt.Sprintf(
        "API Key '%s' expires in %d days. Please rotate it soon.",
        key.Name, daysLeft
    )
    
    // Отправляем уведомления через все каналы
    n.emailNotifier.Send(key.UserID, "API Key Expiration", message)
    n.slackNotifier.Send(key.UserID, message)
    n.webhookNotifier.Send(key.UserID, message)
    
    return nil
}
```

### Phase 3: Security Enhancements (1 неделя)

#### 3.1 Key Validation Middleware
```go
// internal/middleware/api_key_middleware.go
func APIKeyMiddleware(keyStore KeyStore) gin.HandlerFunc {
    return func(c *gin.Context) {
        apiKey := extractAPIKey(c)
        if apiKey == "" {
            c.JSON(401, gin.H{"error": "API key required"})
            c.Abort()
            return
        }
        
        key, err := keyStore.GetKeyByHash(hashKey(apiKey))
        if err != nil || !key.IsActive {
            c.JSON(401, gin.H{"error": "Invalid API key"})
            c.Abort()
            return
        }
        
        if time.Now().After(key.ExpiresAt) {
            c.JSON(401, gin.H{"error": "API key expired"})
            c.Abort()
            return
        }
        
        // Обновляем время последнего использования
        key.LastUsedAt = time.Now()
        keyStore.UpdateKey(key)
        
        c.Set("api_key", key)
        c.Next()
    }
}
```

#### 3.2 Permission System
- [ ] Scope-based permissions
- [ ] Resource-based access control
- [ ] Rate limiting per key
- [ ] Audit logging для использования ключей

#### 3.3 Security Monitoring
- [ ] Отслеживание использования ключей
- [ ] Detection подозрительной активности
- [ ] Automatic key revocation
- [ ] Security alerts

### Phase 4: Advanced Features (1 неделя)

#### 4.1 Key Lifecycle Management
```go
// internal/security/lifecycle.go
type KeyLifecycle struct {
    Created    time.Time `json:"created"`
    Activated  time.Time `json:"activated"`
    Rotated    time.Time `json:"rotated"`
    Expired    time.Time `json:"expired"`
    Revoked    time.Time `json:"revoked"`
    Deleted    time.Time `json:"deleted"`
}

func (l *KeyLifecycle) GetStatus() string {
    if !l.Deleted.IsZero() {
        return "deleted"
    }
    if !l.Revoked.IsZero() {
        return "revoked"
    }
    if time.Now().After(l.Expired) {
        return "expired"
    }
    if time.Now().After(l.Rotated) {
        return "rotated"
    }
    return "active"
}
```

#### 4.2 Bulk Operations
- [ ] Массовая ротация ключей
- [ ] Bulk key generation
- [ ] Batch notifications
- [ ] Bulk key revocation

#### 4.3 Integration Features
- [ ] REST API для управления ключами
- [ ] CLI tools для администраторов
- [ ] Webhook notifications
- [ ] Integration с external systems

## 🗄️ Database Schema

### API Keys Table
```sql
CREATE TABLE api_keys (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    key_hash VARCHAR(255) NOT NULL UNIQUE,
    key_prefix VARCHAR(8) NOT NULL,
    user_id BIGINT NOT NULL,
    user_type VARCHAR(20) NOT NULL,
    permissions JSONB,
    scopes JSONB,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    last_used_at TIMESTAMP WITH TIME ZONE,
    rotated_at TIMESTAMP WITH TIME ZONE,
    metadata JSONB,
    
    FOREIGN KEY (user_id) REFERENCES admin_users(id)
);

-- Индексы
CREATE INDEX idx_api_keys_user_id ON api_keys(user_id);
CREATE INDEX idx_api_keys_key_hash ON api_keys(key_hash);
CREATE INDEX idx_api_keys_expires_at ON api_keys(expires_at);
CREATE INDEX idx_api_keys_is_active ON api_keys(is_active);
```

### Key Rotation History
```sql
CREATE TABLE api_key_rotations (
    id SERIAL PRIMARY KEY,
    key_id VARCHAR(36) NOT NULL,
    old_key_hash VARCHAR(255),
    new_key_hash VARCHAR(255),
    rotated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    rotated_by BIGINT,
    reason VARCHAR(100),
    
    FOREIGN KEY (key_id) REFERENCES api_keys(id),
    FOREIGN KEY (rotated_by) REFERENCES admin_users(id)
);
```

### Key Usage Logs
```sql
CREATE TABLE api_key_usage (
    id SERIAL PRIMARY KEY,
    key_id VARCHAR(36) NOT NULL,
    endpoint VARCHAR(255) NOT NULL,
    method VARCHAR(10) NOT NULL,
    ip_address INET,
    user_agent TEXT,
    response_code INTEGER,
    duration_ms INTEGER,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    FOREIGN KEY (key_id) REFERENCES api_keys(id)
);

-- Партиционирование по месяцам
CREATE TABLE api_key_usage_y2025m01 PARTITION OF api_key_usage
    FOR VALUES FROM ('2025-01-01') TO ('2025-02-01');
```

## 🔧 Конфигурация

### Environment Variables
```bash
# Key Rotation Configuration
API_KEY_ROTATION_ENABLED=true
API_KEY_DEFAULT_LIFETIME_DAYS=90
API_KEY_WARNING_DAYS=7
API_KEY_GRACE_PERIOD_DAYS=3
API_KEY_AUTO_ROTATION=true

# Key Generation
API_KEY_LENGTH=32
API_KEY_ALPHABET=ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789
API_KEY_PREFIX_LENGTH=8

# Notifications
API_KEY_NOTIFICATION_EMAIL=true
API_KEY_NOTIFICATION_SLACK=true
API_KEY_NOTIFICATION_WEBHOOK=true
API_KEY_NOTIFICATION_WEBHOOK_URL=https://hooks.slack.com/your-webhook
```

### Config Structure
```go
type KeyRotationConfig struct {
    Enabled           bool          `json:"enabled"`
    DefaultLifetime   int           `json:"default_lifetime_days"`
    WarningDays       int           `json:"warning_days"`
    GracePeriodDays   int           `json:"grace_period_days"`
    AutoRotation      bool          `json:"auto_rotation"`
    Notifications     NotificationConfig `json:"notifications"`
    Generation        GenerationConfig `json:"generation"`
}

type NotificationConfig struct {
    Email   bool   `json:"email"`
    Slack   bool   `json:"slack"`
    Webhook bool   `json:"webhook"`
    WebhookURL string `json:"webhook_url"`
}

type GenerationConfig struct {
    Length   int    `json:"length"`
    Alphabet string `json:"alphabet"`
    PrefixLength int `json:"prefix_length"`
}
```

## 🧪 Тестирование

### Unit Tests
- [ ] Key generation и validation
- [ ] Rotation logic
- [ ] Notification system
- [ ] Permission checking

### Integration Tests
- [ ] End-to-end rotation flow
- [ ] API key validation
- [ ] Notification delivery
- [ ] Database operations

### Security Tests
- [ ] Key uniqueness
- [ ] Hash security
- [ ] Permission enforcement
- [ ] Rate limiting

## 📊 Мониторинг и метрики

### Key Metrics
- [ ] Количество активных ключей
- [ ] Количество ротаций в день
- [ ] Время жизни ключей
- [ ] Использование ключей

### Security Metrics
- [ ] Неудачные попытки аутентификации
- [ ] Подозрительная активность
- [ ] Expired key usage attempts
- [ ] Unauthorized access attempts

### Performance Metrics
- [ ] Время генерации ключей
- [ ] Время валидации ключей
- [ ] Database query performance
- [ ] Notification delivery time

## 🚀 Deployment

### Production Considerations
- [ ] Secure key storage
- [ ] Backup и recovery
- [ ] Monitoring и alerting
- [ ] Performance optimization

### Security Considerations
- [ ] Encryption в transit и at rest
- [ ] Access control для key management
- [ ] Audit logging
- [ ] Regular security reviews

## 📈 Success Metrics

- **Security**: 0 compromised keys
- **Performance**: <10ms key validation time
- **Reliability**: 99.99% key rotation success rate
- **Compliance**: 100% key lifecycle tracking

## 🔄 Timeline

| Phase | Duration | Deliverables |
|-------|----------|--------------|
| **Phase 1** | 1 неделя | Key management infrastructure |
| **Phase 2** | 1 неделя | Rotation logic |
| **Phase 3** | 1 неделя | Security enhancements |
| **Phase 4** | 1 неделя | Advanced features |

**Total: 4 недели**

## 💰 Ресурсы

- **Backend Developer**: 1 FTE
- **Security Expert**: 0.3 FTE
- **DevOps Engineer**: 0.2 FTE
- **Testing**: 0.3 FTE
- **Documentation**: 0.2 FTE

**Total: 2 FTE (4 недели)**

## 📋 Best Practices

### Key Generation
- [ ] Использование криптографически стойких генераторов
- [ ] Достаточная длина ключей (минимум 32 символа)
- [ ] Уникальность ключей
- [ ] Случайность генерации

### Key Storage
- [ ] Хеширование ключей (bcrypt/scrypt)
- [ ] Шифрование в базе данных
- [ ] Secure key transmission
- [ ] Access control

### Key Rotation
- [ ] Регулярная ротация (90 дней)
- [ ] Grace period для перехода
- [ ] Уведомления пользователей
- [ ] Audit trail

### Key Monitoring
- [ ] Отслеживание использования
- [ ] Detection аномалий
- [ ] Automatic revocation
- [ ] Security alerts

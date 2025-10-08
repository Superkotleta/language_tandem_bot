# 🔐 План реализации OAuth2/JWT аутентификации для Admin API

## 📋 Обзор

Добавление OAuth2/JWT аутентификации для защиты Admin API endpoints и обеспечения безопасного доступа к административным функциям.

## 🎯 Цели

- **Безопасность**: Защита Admin API от несанкционированного доступа
- **Аутентификация**: Проверка личности администраторов
- **Авторизация**: Контроль доступа к различным функциям
- **Аудит**: Отслеживание действий администраторов

## 🏗️ Архитектура решения

### 1. **JWT Token Management**

```go
// internal/auth/jwt_manager.go
type JWTManager struct {
    secretKey     []byte
    tokenDuration time.Duration
    issuer        string
}

type Claims struct {
    UserID    int64    `json:"user_id"`
    Username  string   `json:"username"`
    Roles     []string `json:"roles"`
    Permissions []string `json:"permissions"`
    jwt.StandardClaims
}
```

### 2. **OAuth2 Provider Integration**

```go
// internal/auth/oauth_provider.go
type OAuth2Provider interface {
    GetAuthURL(state string) string
    ExchangeCodeForToken(code string) (*TokenResponse, error)
    GetUserInfo(token string) (*UserInfo, error)
}

// Поддержка провайдеров:
// - Google OAuth2
// - GitHub OAuth2  
// - Custom OAuth2 (для внутренних пользователей)
```

### 3. **Middleware для защиты endpoints**

```go
// internal/middleware/auth_middleware.go
func JWTAuthMiddleware(jwtManager *JWTManager) gin.HandlerFunc {
    return func(c *gin.Context) {
        token := extractToken(c)
        claims, err := jwtManager.ValidateToken(token)
        if err != nil {
            c.JSON(401, gin.H{"error": "Unauthorized"})
            c.Abort()
            return
        }
        
        c.Set("user_claims", claims)
        c.Next()
    }
}
```

## 📁 Структура файлов

```
services/bot/internal/auth/
├── jwt_manager.go          # JWT токены
├── oauth_provider.go       # OAuth2 провайдеры
├── middleware.go           # Auth middleware
├── permissions.go          # Система разрешений
├── user_store.go           # Хранение пользователей
└── auth_test.go            # Тесты аутентификации
```

## 🔧 Реализация

### Phase 1: JWT Infrastructure (1 неделя)

#### 1.1 JWT Manager

- [ ] Создание и валидация JWT токенов
- [ ] Настройка секретного ключа
- [ ] Управление временем жизни токенов
- [ ] Refresh token механизм

#### 1.2 User Store

- [ ] Модель администратора
- [ ] Хранение в PostgreSQL
- [ ] CRUD операции
- [ ] Хеширование паролей

#### 1.3 Basic Auth Endpoints

- [ ] `POST /auth/login` - логин с username/password
- [ ] `POST /auth/refresh` - обновление токена
- [ ] `POST /auth/logout` - выход из системы

### Phase 2: OAuth2 Integration (1 неделя)

#### 2.1 Google OAuth2

- [ ] Настройка Google OAuth2 credentials
- [ ] Реализация Google OAuth2 provider
- [ ] Автоматическое создание пользователей
- [ ] Связывание с существующими аккаунтами

#### 2.2 GitHub OAuth2

- [ ] Настройка GitHub OAuth2 app
- [ ] Реализация GitHub OAuth2 provider
- [ ] Получение информации о пользователе
- [ ] Авторизация по GitHub организации

#### 2.3 OAuth2 Endpoints

- [ ] `GET /auth/oauth/{provider}` - начало OAuth2 flow
- [ ] `GET /auth/oauth/{provider}/callback` - обработка callback
- [ ] `POST /auth/oauth/link` - связывание аккаунтов

### Phase 3: Authorization System (1 неделя)

#### 3.1 Role-Based Access Control (RBAC)

```go
type Role struct {
    ID          int      `json:"id"`
    Name        string   `json:"name"`
    Permissions []string `json:"permissions"`
}

type Permission struct {
    ID          int    `json:"id"`
    Name        string `json:"name"`
    Resource    string `json:"resource"`
    Action      string `json:"action"`
    Description string `json:"description"`
}
```

#### 3.2 Permission System

- [ ] Определение ролей (admin, moderator, viewer)
- [ ] Система разрешений по ресурсам
- [ ] Middleware для проверки разрешений
- [ ] Динамическое управление правами

#### 3.3 Protected Endpoints

- [ ] Защита всех Admin API endpoints
- [ ] Различные уровни доступа
- [ ] Аудит доступа к endpoints

### Phase 4: Advanced Features (1 неделя)

#### 4.1 Session Management

- [ ] Управление активными сессиями
- [ ] Принудительный logout
- [ ] Отзыв токенов
- [ ] Мониторинг сессий

#### 4.2 Security Enhancements

- [ ] Rate limiting для auth endpoints
- [ ] IP whitelist для администраторов
- [ ] 2FA поддержка (TOTP)
- [ ] Audit logging всех auth событий

## 🗄️ Database Schema

### Таблица администраторов

```sql
CREATE TABLE admin_users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255),
    oauth_provider VARCHAR(20),
    oauth_id VARCHAR(100),
    is_active BOOLEAN DEFAULT true,
    last_login TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
```

### Таблица ролей

```sql
CREATE TABLE admin_roles (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);
```

### Таблица разрешений

```sql
CREATE TABLE admin_permissions (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    resource VARCHAR(50) NOT NULL,
    action VARCHAR(50) NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);
```

### Связующие таблицы

```sql
CREATE TABLE admin_user_roles (
    user_id INTEGER REFERENCES admin_users(id),
    role_id INTEGER REFERENCES admin_roles(id),
    PRIMARY KEY (user_id, role_id)
);

CREATE TABLE admin_role_permissions (
    role_id INTEGER REFERENCES admin_roles(id),
    permission_id INTEGER REFERENCES admin_permissions(id),
    PRIMARY KEY (role_id, permission_id)
);
```

## 🔧 Конфигурация

### Environment Variables

```bash
# JWT Configuration
JWT_SECRET_KEY=your-secret-key-here
JWT_TOKEN_DURATION=24h
JWT_REFRESH_DURATION=168h

# OAuth2 Providers
GOOGLE_CLIENT_ID=your-google-client-id
GOOGLE_CLIENT_SECRET=your-google-client-secret
GITHUB_CLIENT_ID=your-github-client-id
GITHUB_CLIENT_SECRET=your-github-client-secret

# Security
ADMIN_IP_WHITELIST=192.168.1.0/24,10.0.0.0/8
ENABLE_2FA=true
SESSION_TIMEOUT=30m
```

### Config Structure

```go
type AuthConfig struct {
    JWT JWTConfig `json:"jwt"`
    OAuth2 OAuth2Config `json:"oauth2"`
    Security SecurityConfig `json:"security"`
}

type JWTConfig struct {
    SecretKey     string        `json:"secret_key"`
    TokenDuration time.Duration `json:"token_duration"`
    RefreshDuration time.Duration `json:"refresh_duration"`
}

type OAuth2Config struct {
    Google GoogleOAuth2 `json:"google"`
    GitHub GitHubOAuth2 `json:"github"`
}

type SecurityConfig struct {
    IPWhitelist   []string      `json:"ip_whitelist"`
    Enable2FA     bool          `json:"enable_2fa"`
    SessionTimeout time.Duration `json:"session_timeout"`
}
```

## 🧪 Тестирование

### Unit Tests

- [ ] JWT token creation/validation
- [ ] OAuth2 provider integration
- [ ] Permission checking
- [ ] User authentication

### Integration Tests

- [ ] Full OAuth2 flow
- [ ] API endpoint protection
- [ ] Role-based access control
- [ ] Session management

### Security Tests

- [ ] Token tampering protection
- [ ] CSRF protection
- [ ] Rate limiting
- [ ] Input validation

## 📊 Мониторинг и метрики

### Auth Metrics

- [ ] Количество успешных/неуспешных логинов
- [ ] Время жизни сессий
- [ ] Использование OAuth2 провайдеров
- [ ] Попытки несанкционированного доступа

### Security Alerts

- [ ] Множественные неудачные попытки входа
- [ ] Подозрительная активность
- [ ] Попытки доступа с неавторизованных IP
- [ ] Использование истекших токенов

## 🚀 Deployment

### Production Considerations

- [ ] Использование внешнего Key Management Service
- [ ] Настройка HTTPS для всех auth endpoints
- [ ] Конфигурация CORS для OAuth2 callbacks
- [ ] Backup и восстановление auth данных

### Migration Strategy

- [ ] Постепенное внедрение без нарушения работы
- [ ] Fallback на старую систему аутентификации
- [ ] Миграция существующих администраторов
- [ ] Обновление документации API

## 📈 Success Metrics

- **Security**: 0 критических уязвимостей в auth системе
- **Performance**: <100ms время аутентификации
- **Availability**: 99.9% uptime для auth endpoints
- **User Experience**: <3 клика для входа через OAuth2

## 🔄 Timeline

| Phase | Duration | Deliverables |
|-------|----------|--------------|
| **Phase 1** | 1 неделя | JWT infrastructure, basic auth |
| **Phase 2** | 1 неделя | OAuth2 integration |
| **Phase 3** | 1 неделя | RBAC system |
| **Phase 4** | 1 неделя | Advanced features |

**Total: 4 недели**

## 💰 Ресурсы

- **Backend Developer**: 1 FTE
- **Security Review**: 0.2 FTE
- **Testing**: 0.3 FTE
- **Documentation**: 0.1 FTE

**Total: 1.6 FTE (4 недели)**

# Design Document: Auth Session Lifecycle

## Overview

This design document outlines the implementation of JWT refresh token functionality and explicit logout capabilities for the Spexus Product Requirements Management System. The feature enables frontend clients to maintain user sessions without repeated authentication while providing secure session termination.

### Goals

- Implement POST /auth/refresh endpoint for token renewal
- Implement POST /auth/logout endpoint for session termination
- Extend POST /auth/login to issue refresh tokens
- Secure refresh token storage with bcrypt hashing
- Rate limiting to prevent abuse
- Automatic cleanup of expired tokens
- Comprehensive API documentation

### Non-Goals

- OAuth2/OIDC integration (future consideration)
- Multi-device session management UI
- Session analytics and monitoring
- Refresh token families (single-use refresh tokens)

## Architecture

### High-Level Flow

```
┌─────────────┐         ┌──────────────┐         ┌─────────────┐
│   Frontend  │         │  Auth API    │         │  Database   │
│   Client    │         │  (Gin)       │         │ (PostgreSQL)│
└──────┬──────┘         └──────┬───────┘         └──────┬──────┘
       │                       │                        │
       │  POST /auth/login     │                        │
       │──────────────────────>│                        │
       │                       │  Create refresh token  │
       │                       │───────────────────────>│
       │  {access, refresh}    │                        │
       │<──────────────────────│                        │
       │                       │                        │
       │  POST /auth/refresh   │                        │
       │──────────────────────>│                        │
       │                       │  Validate & rotate     │
       │                       │───────────────────────>│
       │  {new access, refresh}│                        │
       │<──────────────────────│                        │
       │                       │                        │
       │  POST /auth/logout    │                        │
       │──────────────────────>│                        │
       │                       │  Invalidate token      │
       │                       │───────────────────────>│
       │  204 No Content       │                        │
       │<──────────────────────│                        │
```


### Component Architecture

```
┌────────────────────────────────────────────────────────────┐
│                     HTTP Layer (Gin)                       │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐    │
│  │ POST /login  │  │ POST /refresh│  │ POST /logout │    │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘    │
└─────────┼──────────────────┼──────────────────┼───────────┘
          │                  │                  │
┌─────────▼──────────────────▼──────────────────▼───────────┐
│              Auth Handlers (internal/auth)                 │
│  ┌──────────────────────────────────────────────────────┐ │
│  │  - Login()                                           │ │
│  │  - RefreshToken()                                    │ │
│  │  - Logout()                                          │ │
│  └──────────────────────────────────────────────────────┘ │
└─────────┬──────────────────┬──────────────────┬───────────┘
          │                  │                  │
┌─────────▼──────────────────▼──────────────────▼───────────┐
│              Auth Service (internal/auth)                  │
│  ┌──────────────────────────────────────────────────────┐ │
│  │  - GenerateToken()                                   │ │
│  │  - GenerateRefreshToken()                            │ │
│  │  - ValidateRefreshToken()                            │ │
│  │  - RevokeRefreshToken()                              │ │
│  │  - CleanupExpiredTokens()                            │ │
│  └──────────────────────────────────────────────────────┘ │
└─────────┬──────────────────┬──────────────────┬───────────┘
          │                  │                  │
┌─────────▼──────────────────▼──────────────────▼───────────┐
│         Refresh Token Repository (internal/repository)     │
│  ┌──────────────────────────────────────────────────────┐ │
│  │  - Create()                                          │ │
│  │  - FindByTokenHash()                                 │ │
│  │  │  - Update()                                          │ │
│  │  - Delete()                                          │ │
│  │  - DeleteExpired()                                   │ │
│  └──────────────────────────────────────────────────────┘ │
└─────────┬──────────────────────────────────────────────────┘
          │
┌─────────▼──────────────────────────────────────────────────┐
│                  Database (PostgreSQL)                      │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  refresh_tokens table                                │  │
│  │  - id (UUID)                                         │  │
│  │  - user_id (UUID FK)                                 │  │
│  │  - token_hash (TEXT)                                 │  │
│  │  - created_at (TIMESTAMP)                            │  │
│  │  - expires_at (TIMESTAMP)                            │  │
│  │  - last_used_at (TIMESTAMP)                          │  │
│  └──────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```


## Components and Interfaces

### 1. Database Schema

#### Refresh Tokens Table

```sql
CREATE TABLE refresh_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    last_used_at TIMESTAMP WITH TIME ZONE,
    CONSTRAINT fk_refresh_tokens_user FOREIGN KEY (user_id) 
        REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);
CREATE INDEX idx_refresh_tokens_token_hash ON refresh_tokens(token_hash);
```

**Migration File**: `migrations/000009_add_refresh_tokens.up.sql`

### 2. Data Models

#### RefreshToken Model (internal/models/refresh_token.go)

```go
package models

import (
    "time"
    "github.com/google/uuid"
    "gorm.io/gorm"
)

type RefreshToken struct {
    ID          uuid.UUID  `gorm:"type:uuid;primary_key" json:"id"`
    UserID      uuid.UUID  `gorm:"type:uuid;not null" json:"user_id"`
    TokenHash   string     `gorm:"not null" json:"-"` // Never expose in JSON
    CreatedAt   time.Time  `json:"created_at"`
    ExpiresAt   time.Time  `gorm:"not null" json:"expires_at"`
    LastUsedAt  *time.Time `json:"last_used_at,omitempty"`
    
    // Relationships
    User User `gorm:"foreignKey:UserID" json:"-"`
}

func (rt *RefreshToken) BeforeCreate(tx *gorm.DB) error {
    if rt.ID == uuid.Nil {
        rt.ID = uuid.New()
    }
    return nil
}

func (RefreshToken) TableName() string {
    return "refresh_tokens"
}

func (rt *RefreshToken) IsExpired() bool {
    return time.Now().After(rt.ExpiresAt)
}
```


### 3. Repository Layer

#### RefreshTokenRepository Interface (internal/repository/interfaces.go)

```go
type RefreshTokenRepository interface {
    Create(ctx context.Context, token *models.RefreshToken) error
    FindByTokenHash(ctx context.Context, tokenHash string) (*models.RefreshToken, error)
    FindByUserID(ctx context.Context, userID uuid.UUID) ([]*models.RefreshToken, error)
    Update(ctx context.Context, token *models.RefreshToken) error
    Delete(ctx context.Context, id uuid.UUID) error
    DeleteByUserID(ctx context.Context, userID uuid.UUID) error
    DeleteExpired(ctx context.Context) (int64, error)
}
```

#### RefreshTokenRepository Implementation (internal/repository/refresh_token_repository.go)

```go
package repository

import (
    "context"
    "time"
    "product-requirements-management/internal/models"
    "github.com/google/uuid"
    "gorm.io/gorm"
)

type refreshTokenRepository struct {
    db *gorm.DB
}

func NewRefreshTokenRepository(db *gorm.DB) RefreshTokenRepository {
    return &refreshTokenRepository{db: db}
}

func (r *refreshTokenRepository) Create(ctx context.Context, token *models.RefreshToken) error {
    return r.db.WithContext(ctx).Create(token).Error
}

func (r *refreshTokenRepository) FindByTokenHash(ctx context.Context, tokenHash string) (*models.RefreshToken, error) {
    var token models.RefreshToken
    err := r.db.WithContext(ctx).
        Where("token_hash = ?", tokenHash).
        First(&token).Error
    if err != nil {
        return nil, err
    }
    return &token, nil
}

func (r *refreshTokenRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*models.RefreshToken, error) {
    var tokens []*models.RefreshToken
    err := r.db.WithContext(ctx).
        Where("user_id = ?", userID).
        Order("created_at DESC").
        Find(&tokens).Error
    return tokens, err
}

func (r *refreshTokenRepository) Update(ctx context.Context, token *models.RefreshToken) error {
    return r.db.WithContext(ctx).Save(token).Error
}

func (r *refreshTokenRepository) Delete(ctx context.Context, id uuid.UUID) error {
    return r.db.WithContext(ctx).Delete(&models.RefreshToken{}, "id = ?", id).Error
}

func (r *refreshTokenRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
    return r.db.WithContext(ctx).Delete(&models.RefreshToken{}, "user_id = ?", userID).Error
}

func (r *refreshTokenRepository) DeleteExpired(ctx context.Context) (int64, error) {
    result := r.db.WithContext(ctx).
        Delete(&models.RefreshToken{}, "expires_at < ?", time.Now())
    return result.RowsAffected, result.Error
}
```


### 4. Service Layer

#### Auth Service Extensions (internal/auth/service.go)

```go
// Add to existing Service struct
type Service struct {
    jwtSecret           []byte
    tokenDuration       time.Duration
    refreshTokenRepo    repository.RefreshTokenRepository
    refreshTokenExpiry  time.Duration // 30 days
}

// Update NewService constructor
func NewService(jwtSecret string, tokenDuration time.Duration, refreshTokenRepo repository.RefreshTokenRepository) *Service {
    return &Service{
        jwtSecret:          []byte(jwtSecret),
        tokenDuration:      tokenDuration,
        refreshTokenRepo:   refreshTokenRepo,
        refreshTokenExpiry: 30 * 24 * time.Hour, // 30 days
    }
}

// GenerateRefreshToken creates a new refresh token
func (s *Service) GenerateRefreshToken(ctx context.Context, user *models.User) (string, error) {
    // Generate secure random token (32 bytes = 256 bits)
    tokenBytes := make([]byte, 32)
    if _, err := rand.Read(tokenBytes); err != nil {
        return "", fmt.Errorf("failed to generate random token: %w", err)
    }
    
    // Encode to base64 URL-safe string
    token := base64.URLEncoding.EncodeToString(tokenBytes)
    
    // Hash the token for storage
    tokenHash, err := bcrypt.GenerateFromPassword([]byte(token), bcrypt.DefaultCost)
    if err != nil {
        return "", fmt.Errorf("failed to hash token: %w", err)
    }
    
    // Create refresh token record
    refreshToken := &models.RefreshToken{
        UserID:    user.ID,
        TokenHash: string(tokenHash),
        ExpiresAt: time.Now().Add(s.refreshTokenExpiry),
    }
    
    if err := s.refreshTokenRepo.Create(ctx, refreshToken); err != nil {
        return "", fmt.Errorf("failed to store refresh token: %w", err)
    }
    
    return token, nil
}

// ValidateRefreshToken validates and returns user for refresh token
func (s *Service) ValidateRefreshToken(ctx context.Context, token string) (*models.User, string, error) {
    // Find all refresh tokens (we need to check hashes)
    // In production, consider adding a token prefix/identifier to optimize this
    var allTokens []*models.RefreshToken
    if err := s.refreshTokenRepo.db.WithContext(ctx).Find(&allTokens).Error; err != nil {
        return nil, "", ErrInvalidToken
    }
    
    // Find matching token by comparing hashes
    var matchedToken *models.RefreshToken
    for _, rt := range allTokens {
        if err := bcrypt.CompareHashAndPassword([]byte(rt.TokenHash), []byte(token)); err == nil {
            matchedToken = rt
            break
        }
    }
    
    if matchedToken == nil {
        return nil, "", ErrInvalidToken
    }
    
    // Check expiration
    if matchedToken.IsExpired() {
        // Clean up expired token
        s.refreshTokenRepo.Delete(ctx, matchedToken.ID)
        return nil, "", ErrTokenExpired
    }
    
    // Update last used timestamp
    now := time.Now()
    matchedToken.LastUsedAt = &now
    s.refreshTokenRepo.Update(ctx, matchedToken)
    
    // Get user
    var user models.User
    if err := s.refreshTokenRepo.db.WithContext(ctx).First(&user, "id = ?", matchedToken.UserID).Error; err != nil {
        return nil, "", ErrInvalidToken
    }
    
    // Generate new refresh token (token rotation)
    newRefreshToken, err := s.GenerateRefreshToken(ctx, &user)
    if err != nil {
        return nil, "", err
    }
    
    // Revoke old token
    s.refreshTokenRepo.Delete(ctx, matchedToken.ID)
    
    return &user, newRefreshToken, nil
}

// RevokeRefreshToken invalidates a refresh token
func (s *Service) RevokeRefreshToken(ctx context.Context, token string) error {
    // Find and delete the token
    var allTokens []*models.RefreshToken
    if err := s.refreshTokenRepo.db.WithContext(ctx).Find(&allTokens).Error; err != nil {
        return ErrInvalidToken
    }
    
    for _, rt := range allTokens {
        if err := bcrypt.CompareHashAndPassword([]byte(rt.TokenHash), []byte(token)); err == nil {
            return s.refreshTokenRepo.Delete(ctx, rt.ID)
        }
    }
    
    return ErrInvalidToken
}

// CleanupExpiredTokens removes expired refresh tokens
func (s *Service) CleanupExpiredTokens(ctx context.Context) (int64, error) {
    return s.refreshTokenRepo.DeleteExpired(ctx)
}
```


### 5. Handler Layer

#### Request/Response Types (internal/auth/handlers.go)

```go
// Update LoginResponse to include refresh token
type LoginResponse struct {
    Token        string       `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
    RefreshToken string       `json:"refresh_token" example:"dGhpc19pc19hX3JlZnJlc2hfdG9rZW4="`
    User         UserResponse `json:"user"`
    ExpiresAt    time.Time    `json:"expires_at" example:"2023-01-02T12:30:00Z"`
}

// RefreshRequest represents a token refresh request
type RefreshRequest struct {
    RefreshToken string `json:"refresh_token" binding:"required" example:"dGhpc19pc19hX3JlZnJlc2hfdG9rZW4="`
}

// RefreshResponse represents a token refresh response
type RefreshResponse struct {
    Token        string    `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
    RefreshToken string    `json:"refresh_token" example:"bmV3X3JlZnJlc2hfdG9rZW4="`
    ExpiresAt    time.Time `json:"expires_at" example:"2023-01-02T12:30:00Z"`
}

// LogoutRequest represents a logout request
type LogoutRequest struct {
    RefreshToken string `json:"refresh_token" binding:"required" example:"dGhpc19pc19hX3JlZnJlc2hfdG9rZW4="`
}
```

#### Handler Methods (internal/auth/handlers.go)

```go
// Update Login handler to include refresh token
func (h *Handlers) Login(c *gin.Context) {
    var req LoginRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    var user models.User
    if err := h.db.Where("username = ?", req.Username).First(&user).Error; err != nil {
        if err == gorm.ErrRecordNotFound {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
        return
    }

    if err := h.service.VerifyPassword(req.Password, user.PasswordHash); err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
        return
    }

    // Generate access token
    token, err := h.service.GenerateToken(&user)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
        return
    }

    // Generate refresh token
    refreshToken, err := h.service.GenerateRefreshToken(c.Request.Context(), &user)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate refresh token"})
        return
    }

    response := LoginResponse{
        Token:        token,
        RefreshToken: refreshToken,
        User: UserResponse{
            ID:        user.ID.String(),
            Username:  user.Username,
            Email:     user.Email,
            Role:      user.Role,
            CreatedAt: user.CreatedAt,
            UpdatedAt: user.UpdatedAt,
        },
        ExpiresAt: time.Now().Add(h.service.tokenDuration),
    }

    c.JSON(http.StatusOK, response)
}

// RefreshToken handles token refresh
func (h *Handlers) RefreshToken(c *gin.Context) {
    var req RefreshRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": map[string]string{
                "code":    "VALIDATION_ERROR",
                "message": err.Error(),
            },
        })
        return
    }

    // Validate refresh token and get user
    user, newRefreshToken, err := h.service.ValidateRefreshToken(c.Request.Context(), req.RefreshToken)
    if err != nil {
        if err == auth.ErrTokenExpired {
            c.JSON(http.StatusUnauthorized, gin.H{
                "error": map[string]string{
                    "code":    "REFRESH_TOKEN_EXPIRED",
                    "message": "Refresh token has expired",
                },
            })
            return
        }
        c.JSON(http.StatusUnauthorized, gin.H{
            "error": map[string]string{
                "code":    "INVALID_REFRESH_TOKEN",
                "message": "Invalid or revoked refresh token",
            },
        })
        return
    }

    // Generate new access token
    accessToken, err := h.service.GenerateToken(user)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": map[string]string{
                "code":    "INTERNAL_ERROR",
                "message": "Failed to generate access token",
            },
        })
        return
    }

    response := RefreshResponse{
        Token:        accessToken,
        RefreshToken: newRefreshToken,
        ExpiresAt:    time.Now().Add(h.service.tokenDuration),
    }

    c.JSON(http.StatusOK, response)
}

// Logout handles user logout
func (h *Handlers) Logout(c *gin.Context) {
    var req LogoutRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": map[string]string{
                "code":    "VALIDATION_ERROR",
                "message": err.Error(),
            },
        })
        return
    }

    // Revoke refresh token
    if err := h.service.RevokeRefreshToken(c.Request.Context(), req.RefreshToken); err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{
            "error": map[string]string{
                "code":    "INVALID_REFRESH_TOKEN",
                "message": "Session already logged out",
            },
        })
        return
    }

    c.Status(http.StatusNoContent)
}
```


### 6. Rate Limiting

#### Rate Limiter Middleware (internal/auth/rate_limiter.go)

```go
package auth

import (
    "fmt"
    "net/http"
    "sync"
    "time"
    "github.com/gin-gonic/gin"
)

type rateLimiter struct {
    requests map[string][]time.Time
    mu       sync.RWMutex
    limit    int
    window   time.Duration
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
    rl := &rateLimiter{
        requests: make(map[string][]time.Time),
        limit:    limit,
        window:   window,
    }
    
    // Start cleanup goroutine
    go rl.cleanup()
    
    return rl
}

func (rl *rateLimiter) cleanup() {
    ticker := time.NewTicker(rl.window)
    defer ticker.Stop()
    
    for range ticker.C {
        rl.mu.Lock()
        now := time.Now()
        for key, timestamps := range rl.requests {
            // Remove expired timestamps
            valid := []time.Time{}
            for _, ts := range timestamps {
                if now.Sub(ts) < rl.window {
                    valid = append(valid, ts)
                }
            }
            if len(valid) == 0 {
                delete(rl.requests, key)
            } else {
                rl.requests[key] = valid
            }
        }
        rl.mu.Unlock()
    }
}

func (rl *rateLimiter) allow(key string) (bool, time.Duration) {
    rl.mu.Lock()
    defer rl.mu.Unlock()
    
    now := time.Now()
    timestamps := rl.requests[key]
    
    // Remove expired timestamps
    valid := []time.Time{}
    for _, ts := range timestamps {
        if now.Sub(ts) < rl.window {
            valid = append(valid, ts)
        }
    }
    
    if len(valid) >= rl.limit {
        // Calculate retry after
        oldest := valid[0]
        retryAfter := rl.window - now.Sub(oldest)
        return false, retryAfter
    }
    
    // Add current request
    valid = append(valid, now)
    rl.requests[key] = valid
    
    return true, 0
}

// RefreshRateLimitMiddleware limits refresh token requests
func RefreshRateLimitMiddleware() gin.HandlerFunc {
    limiter := newRateLimiter(10, time.Minute) // 10 requests per minute
    
    return func(c *gin.Context) {
        // Use IP address as key (in production, consider using user ID)
        key := c.ClientIP()
        
        allowed, retryAfter := limiter.allow(key)
        if !allowed {
            c.Header("Retry-After", fmt.Sprintf("%d", int(retryAfter.Seconds())))
            c.JSON(http.StatusTooManyRequests, gin.H{
                "error": map[string]string{
                    "code":    "RATE_LIMIT_EXCEEDED",
                    "message": "Too many refresh attempts",
                },
            })
            c.Abort()
            return
        }
        
        c.Next()
    }
}
```


### 7. Background Cleanup Job

#### Token Cleanup Service (internal/auth/cleanup.go)

```go
package auth

import (
    "context"
    "time"
    "github.com/sirupsen/logrus"
)

type CleanupService struct {
    authService *Service
    logger      *logrus.Logger
    ticker      *time.Ticker
    done        chan bool
}

func NewCleanupService(authService *Service, logger *logrus.Logger) *CleanupService {
    return &CleanupService{
        authService: authService,
        logger:      logger,
        done:        make(chan bool),
    }
}

// Start begins the cleanup job
func (cs *CleanupService) Start() {
    cs.ticker = time.NewTicker(24 * time.Hour)
    
    go func() {
        // Run immediately on start
        cs.runCleanup()
        
        // Then run on schedule
        for {
            select {
            case <-cs.ticker.C:
                cs.runCleanup()
            case <-cs.done:
                return
            }
        }
    }()
    
    cs.logger.Info("Refresh token cleanup service started")
}

// Stop gracefully stops the cleanup job
func (cs *CleanupService) Stop() {
    if cs.ticker != nil {
        cs.ticker.Stop()
    }
    cs.done <- true
    cs.logger.Info("Refresh token cleanup service stopped")
}

func (cs *CleanupService) runCleanup() {
    ctx := context.Background()
    
    cs.logger.Info("Starting refresh token cleanup")
    
    count, err := cs.authService.CleanupExpiredTokens(ctx)
    if err != nil {
        cs.logger.WithError(err).Error("Failed to cleanup expired refresh tokens")
        return
    }
    
    cs.logger.WithField("count", count).Info("Cleaned up expired refresh tokens")
}
```

#### Integration in Server (cmd/server/main.go)

```go
// Add to main function
cleanupService := auth.NewCleanupService(authService, logger.Logger)
cleanupService.Start()
defer cleanupService.Stop()
```


### 8. Routing Configuration

#### Route Setup (internal/server/routes/routes.go)

```go
// Add to authGroup in Setup function
authGroup := router.Group("/auth")
{
    authGroup.POST("/login", authHandler.Login)
    authGroup.POST("/refresh", auth.RefreshRateLimitMiddleware(), authHandler.RefreshToken)
    authGroup.POST("/logout", authHandler.Logout)
    authGroup.GET("/profile", authService.Middleware(), authHandler.GetProfile)
    authGroup.POST("/change-password", authService.Middleware(), authHandler.ChangePassword)
    
    // Admin-only user management routes
    authGroup.POST("/users", authService.Middleware(), authService.RequireAdministrator(), authHandler.CreateUser)
    authGroup.GET("/users", authService.Middleware(), authService.RequireAdministrator(), authHandler.GetUsers)
    authGroup.GET("/users/:id", authService.Middleware(), authService.RequireAdministrator(), authHandler.GetUser)
    authGroup.PUT("/users/:id", authService.Middleware(), authService.RequireAdministrator(), authHandler.UpdateUser)
    authGroup.DELETE("/users/:id", authService.Middleware(), authService.RequireAdministrator(), authHandler.DeleteUser)
}
```

## Data Models

### RefreshToken

| Field | Type | Description | Constraints |
|-------|------|-------------|-------------|
| id | UUID | Primary key | NOT NULL, PK |
| user_id | UUID | Foreign key to users | NOT NULL, FK |
| token_hash | TEXT | Bcrypt hash of token | NOT NULL |
| created_at | TIMESTAMP | Creation timestamp | NOT NULL, DEFAULT NOW() |
| expires_at | TIMESTAMP | Expiration timestamp | NOT NULL |
| last_used_at | TIMESTAMP | Last usage timestamp | NULL |

### Updated LoginResponse

| Field | Type | Description |
|-------|------|-------------|
| token | string | JWT access token |
| refresh_token | string | Refresh token |
| user | UserResponse | User information |
| expires_at | time.Time | Access token expiration |

### RefreshResponse

| Field | Type | Description |
|-------|------|-------------|
| token | string | New JWT access token |
| refresh_token | string | New refresh token |
| expires_at | time.Time | Access token expiration |


## Error Handling

### Error Codes and HTTP Status Codes

| Error Code | HTTP Status | Description | When to Use |
|------------|-------------|-------------|-------------|
| VALIDATION_ERROR | 400 | Request validation failed | Invalid JSON or missing required fields |
| INVALID_REFRESH_TOKEN | 401 | Refresh token is invalid or revoked | Token not found or already used |
| REFRESH_TOKEN_EXPIRED | 401 | Refresh token has expired | Token past expiration date |
| SESSION_INVALIDATED | 401 | Session has been logged out | Attempting to use token after logout |
| RATE_LIMIT_EXCEEDED | 429 | Too many refresh attempts | Exceeded 10 requests per minute |
| INTERNAL_ERROR | 500 | Server-side error | Database errors, token generation failures |

### Error Response Format

All errors follow the `internal/handlers.ErrorResponse` format:

```go
type ErrorResponse struct {
    Error struct {
        Code    string `json:"code"`
        Message string `json:"message"`
    } `json:"error"`
}
```

### Example Error Responses

```json
// 401 - Expired Token
{
    "error": {
        "code": "REFRESH_TOKEN_EXPIRED",
        "message": "Refresh token has expired"
    }
}

// 429 - Rate Limit
{
    "error": {
        "code": "RATE_LIMIT_EXCEEDED",
        "message": "Too many refresh attempts"
    }
}
```


## Testing Strategy

### Unit Tests

#### Service Layer Tests (internal/auth/service_test.go)

```go
func TestGenerateRefreshToken(t *testing.T)
func TestValidateRefreshToken(t *testing.T)
func TestValidateRefreshToken_Expired(t *testing.T)
func TestValidateRefreshToken_Invalid(t *testing.T)
func TestRevokeRefreshToken(t *testing.T)
func TestCleanupExpiredTokens(t *testing.T)
func TestTokenRotation(t *testing.T)
```

#### Repository Layer Tests (internal/repository/refresh_token_repository_test.go)

```go
func TestCreate(t *testing.T)
func TestFindByTokenHash(t *testing.T)
func TestFindByUserID(t *testing.T)
func TestUpdate(t *testing.T)
func TestDelete(t *testing.T)
func TestDeleteExpired(t *testing.T)
```

#### Handler Layer Tests (internal/auth/handlers_test.go)

```go
func TestLogin_WithRefreshToken(t *testing.T)
func TestRefreshToken_Success(t *testing.T)
func TestRefreshToken_Expired(t *testing.T)
func TestRefreshToken_Invalid(t *testing.T)
func TestRefreshToken_RateLimit(t *testing.T)
func TestLogout_Success(t *testing.T)
func TestLogout_AlreadyLoggedOut(t *testing.T)
```

### Integration Tests

#### End-to-End Flow Tests (internal/integration/auth_session_test.go)

```go
func TestAuthSessionFlow(t *testing.T) {
    // 1. Login and receive tokens
    // 2. Use access token for API calls
    // 3. Refresh tokens before expiration
    // 4. Logout and verify token invalidation
}

func TestTokenRotation(t *testing.T) {
    // 1. Login
    // 2. Refresh token
    // 3. Verify old refresh token is invalid
    // 4. Verify new refresh token works
}

func TestConcurrentRefresh(t *testing.T) {
    // Test concurrent refresh requests
}
```

### Test Data

- Valid user credentials
- Expired refresh tokens
- Invalid refresh tokens
- Multiple refresh tokens per user


## API Documentation (Swagger)

### POST /auth/login

```go
// @Summary User login
// @Description Authenticate user and receive access token and refresh token
// @Tags authentication
// @Accept json
// @Produce json
// @Param login body LoginRequest true "Login credentials"
// @Success 200 {object} LoginResponse "Successful authentication with tokens"
// @Failure 400 {object} ErrorResponse "Invalid request format"
// @Failure 401 {object} ErrorResponse "Invalid credentials"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /auth/login [post]
```

### POST /auth/refresh

```go
// @Summary Refresh access token
// @Description Exchange refresh token for new access and refresh tokens
// @Tags authentication
// @Accept json
// @Produce json
// @Param refresh body RefreshRequest true "Refresh token"
// @Success 200 {object} RefreshResponse "New tokens issued"
// @Failure 400 {object} ErrorResponse "Invalid request format"
// @Failure 401 {object} ErrorResponse "Invalid or expired refresh token"
// @Failure 429 {object} ErrorResponse "Rate limit exceeded"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /auth/refresh [post]
```

### POST /auth/logout

```go
// @Summary Logout user
// @Description Invalidate refresh token and end session
// @Tags authentication
// @Accept json
// @Produce json
// @Param logout body LogoutRequest true "Refresh token to revoke"
// @Success 204 "Successfully logged out"
// @Failure 400 {object} ErrorResponse "Invalid request format"
// @Failure 401 {object} ErrorResponse "Invalid refresh token"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /auth/logout [post]
```

### Swagger Schema Definitions

```json
{
  "LoginResponse": {
    "type": "object",
    "properties": {
      "token": {
        "type": "string",
        "example": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
      },
      "refresh_token": {
        "type": "string",
        "example": "dGhpc19pc19hX3JlZnJlc2hfdG9rZW4="
      },
      "user": {
        "$ref": "#/definitions/UserResponse"
      },
      "expires_at": {
        "type": "string",
        "format": "date-time",
        "example": "2023-01-02T12:30:00Z"
      }
    }
  },
  "RefreshRequest": {
    "type": "object",
    "required": ["refresh_token"],
    "properties": {
      "refresh_token": {
        "type": "string",
        "example": "dGhpc19pc19hX3JlZnJlc2hfdG9rZW4="
      }
    }
  },
  "RefreshResponse": {
    "type": "object",
    "properties": {
      "token": {
        "type": "string",
        "example": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
      },
      "refresh_token": {
        "type": "string",
        "example": "bmV3X3JlZnJlc2hfdG9rZW4="
      },
      "expires_at": {
        "type": "string",
        "format": "date-time",
        "example": "2023-01-02T12:30:00Z"
      }
    }
  },
  "LogoutRequest": {
    "type": "object",
    "required": ["refresh_token"],
    "properties": {
      "refresh_token": {
        "type": "string",
        "example": "dGhpc19pc19hX3JlZnJlc2hfdG9rZW4="
      }
    }
  },
  "ErrorResponse": {
    "type": "object",
    "properties": {
      "error": {
        "type": "object",
        "properties": {
          "code": {
            "type": "string",
            "example": "INVALID_REFRESH_TOKEN"
          },
          "message": {
            "type": "string",
            "example": "Invalid or revoked refresh token"
          }
        }
      }
    }
  }
}
```


## Security Considerations

### Token Security

1. **Refresh Token Storage**
   - Never store plain-text refresh tokens in database
   - Use bcrypt hashing (cost 10) for token storage
   - Tokens are 32 bytes (256 bits) of cryptographically secure random data

2. **Token Rotation**
   - Implement automatic token rotation on each refresh
   - Invalidate old refresh token immediately after issuing new one
   - Prevents token replay attacks

3. **Token Expiration**
   - Access tokens: 24 hours (configurable)
   - Refresh tokens: 30 days (configurable)
   - Automatic cleanup of expired tokens

### Rate Limiting

1. **Refresh Endpoint**
   - Limit: 10 requests per minute per IP
   - Returns 429 with Retry-After header
   - Prevents brute force attacks

2. **Implementation**
   - In-memory rate limiter with automatic cleanup
   - Consider Redis for distributed systems
   - Track by IP address (consider user ID in production)

### Best Practices

1. **HTTPS Only**
   - All authentication endpoints must use HTTPS in production
   - Tokens transmitted in request body, not URL parameters

2. **Token Transmission**
   - Access tokens in Authorization header: `Bearer <token>`
   - Refresh tokens in request body only
   - Never log tokens

3. **Error Messages**
   - Generic error messages to prevent information leakage
   - Detailed logging for debugging (server-side only)

4. **Database Security**
   - Foreign key constraints with CASCADE delete
   - Indexes on frequently queried fields
   - Regular cleanup of expired tokens


## Implementation Sequence

### Phase 1: Database and Models
1. Create migration file `000009_add_refresh_tokens.up.sql`
2. Create migration file `000009_add_refresh_tokens.down.sql`
3. Create `internal/models/refresh_token.go`
4. Run migrations and verify schema

### Phase 2: Repository Layer
1. Add `RefreshTokenRepository` interface to `internal/repository/interfaces.go`
2. Implement `internal/repository/refresh_token_repository.go`
3. Update `internal/repository/repository.go` to include new repository
4. Write unit tests for repository

### Phase 3: Service Layer
1. Update `internal/auth/service.go` with refresh token methods
2. Implement token generation, validation, and revocation
3. Implement token rotation logic
4. Write unit tests for service methods

### Phase 4: Handler Layer
1. Update `internal/auth/handlers.go` with new request/response types
2. Update `Login` handler to include refresh token
3. Implement `RefreshToken` handler
4. Implement `Logout` handler
5. Write unit tests for handlers

### Phase 5: Rate Limiting
1. Create `internal/auth/rate_limiter.go`
2. Implement rate limiting middleware
3. Write unit tests for rate limiter

### Phase 6: Background Cleanup
1. Create `internal/auth/cleanup.go`
2. Implement cleanup service
3. Integrate cleanup service in `cmd/server/main.go`
4. Write unit tests for cleanup service

### Phase 7: Routing
1. Update `internal/server/routes/routes.go`
2. Add new routes with appropriate middleware
3. Verify route registration

### Phase 8: API Documentation
1. Add swagger annotations to handlers
2. Generate swagger documentation
3. Verify swagger UI displays new endpoints
4. Update `docs/api-client-export.md` with new endpoints

### Phase 9: Integration Testing
1. Write end-to-end auth flow tests
2. Write token rotation tests
3. Write rate limiting tests
4. Write cleanup job tests

### Phase 10: Documentation
1. Update README with new authentication flow
2. Update API documentation
3. Create migration guide for existing clients


## Migration Strategy

### Database Migration

The migration will be backward compatible:
- New `refresh_tokens` table added
- Existing `users` table unchanged
- No data migration required

### API Compatibility

The changes are backward compatible:
- Existing `/auth/login` endpoint enhanced with additional field
- Existing clients can ignore `refresh_token` field
- New endpoints (`/auth/refresh`, `/auth/logout`) are additive

### Deployment Steps

1. **Pre-deployment**
   - Review and test all changes in staging
   - Backup production database
   - Prepare rollback plan

2. **Deployment**
   - Deploy new application version
   - Run database migrations
   - Verify health checks pass
   - Monitor error logs

3. **Post-deployment**
   - Verify new endpoints are accessible
   - Test token refresh flow
   - Monitor rate limiting effectiveness
   - Verify cleanup job is running

4. **Rollback Plan**
   - Revert to previous application version
   - Run down migration if needed
   - Restore database from backup if necessary

### Client Migration Guide

For frontend clients to adopt the new authentication flow:

1. **Update Login Flow**
   ```typescript
   // Old
   const { token, user, expires_at } = await login(username, password);
   
   // New
   const { token, refresh_token, user, expires_at } = await login(username, password);
   localStorage.setItem('refresh_token', refresh_token);
   ```

2. **Implement Token Refresh**
   ```typescript
   async function refreshAccessToken() {
     const refresh_token = localStorage.getItem('refresh_token');
     const { token, refresh_token: new_refresh_token, expires_at } = 
       await fetch('/auth/refresh', {
         method: 'POST',
         body: JSON.stringify({ refresh_token })
       });
     
     localStorage.setItem('refresh_token', new_refresh_token);
     return token;
   }
   ```

3. **Implement Logout**
   ```typescript
   async function logout() {
     const refresh_token = localStorage.getItem('refresh_token');
     await fetch('/auth/logout', {
       method: 'POST',
       body: JSON.stringify({ refresh_token })
     });
     
     localStorage.removeItem('refresh_token');
   }
   ```

4. **Handle Token Expiration**
   ```typescript
   // Intercept 401 responses and attempt refresh
   axios.interceptors.response.use(
     response => response,
     async error => {
       if (error.response?.status === 401) {
         try {
           const newToken = await refreshAccessToken();
           error.config.headers.Authorization = `Bearer ${newToken}`;
           return axios.request(error.config);
         } catch (refreshError) {
           // Refresh failed, redirect to login
           window.location.href = '/login';
         }
       }
       return Promise.reject(error);
     }
   );
   ```

## Performance Considerations

### Database Performance

1. **Indexes**
   - `idx_refresh_tokens_user_id`: Fast user session lookups
   - `idx_refresh_tokens_expires_at`: Efficient cleanup queries
   - `idx_refresh_tokens_token_hash`: Fast token validation

2. **Query Optimization**
   - Use prepared statements
   - Limit result sets
   - Use connection pooling

### Memory Usage

1. **Rate Limiter**
   - In-memory storage with automatic cleanup
   - Bounded memory usage (max entries × window duration)
   - Consider Redis for distributed systems

2. **Token Storage**
   - Tokens stored in database, not memory
   - Cleanup job prevents unbounded growth

### Scalability

1. **Horizontal Scaling**
   - Stateless authentication (JWT)
   - Shared database for refresh tokens
   - Consider Redis for rate limiting in multi-instance deployments

2. **Load Considerations**
   - Refresh endpoint may see high traffic
   - Rate limiting prevents abuse
   - Database indexes optimize queries

## Monitoring and Observability

### Metrics to Track

1. **Authentication Metrics**
   - Login success/failure rate
   - Token refresh success/failure rate
   - Logout rate
   - Active sessions count

2. **Performance Metrics**
   - Token generation time
   - Token validation time
   - Database query latency
   - Rate limit hit rate

3. **Security Metrics**
   - Failed refresh attempts
   - Expired token usage attempts
   - Rate limit violations

### Logging

1. **Info Level**
   - Successful logins
   - Successful token refreshes
   - Successful logouts
   - Cleanup job execution

2. **Warning Level**
   - Rate limit exceeded
   - Expired token usage attempts

3. **Error Level**
   - Token generation failures
   - Database errors
   - Cleanup job failures

### Alerts

1. **Critical Alerts**
   - Database connection failures
   - High error rate (>5%)
   - Cleanup job failures

2. **Warning Alerts**
   - High rate limit violations
   - Unusual token refresh patterns
   - Database query latency spikes

## Future Enhancements

### Potential Improvements

1. **Multi-Device Session Management**
   - Track device information
   - Allow users to view/revoke sessions
   - Device-specific refresh tokens

2. **Refresh Token Families**
   - Implement single-use refresh tokens
   - Detect token reuse attacks
   - Automatic session revocation on suspicious activity

3. **OAuth2/OIDC Integration**
   - Support external identity providers
   - Implement authorization code flow
   - Add scope-based permissions

4. **Session Analytics**
   - Track session duration
   - Monitor user activity patterns
   - Generate security reports

5. **Advanced Rate Limiting**
   - User-specific rate limits
   - Adaptive rate limiting based on behavior
   - Distributed rate limiting with Redis

6. **Token Introspection Endpoint**
   - Allow clients to validate tokens
   - Provide token metadata
   - Support token revocation lists

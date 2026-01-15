package auth

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ============ JWT Auth ============

// Auth 认证服务
type Auth struct {
	mu         sync.RWMutex
	jwtSecret  []byte
	apiKeys    map[string]*APIKeyInfo
	rateLimiter *RateLimiter
}

// APIKeyInfo API Key 信息
type APIKeyInfo struct {
	ID        string
	Name      string
	Key       string
	CreatedAt time.Time
	ExpiresAt *time.Time
	Scopes    []string
}

// Claims JWT Claims
type Claims struct {
	UserID string   `json:"user_id"`
	Scopes []string `json:"scopes"`
	jwt.RegisteredClaims
}

// NewAuth 创建认证服务
func NewAuth(jwtSecret string) *Auth {
	return &Auth{
		jwtSecret:  []byte(jwtSecret),
		apiKeys:    make(map[string]*APIKeyInfo),
		rateLimiter: NewRateLimiter(60, time.Minute), // 60 requests per minute
	}
}

// GenerateToken 生成 JWT Token
func (a *Auth) GenerateToken(userID string, scopes []string, expiresIn time.Duration) (string, error) {
	claims := &Claims{
		UserID: userID,
		Scopes: scopes,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "tortoise",
			Subject:   userID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(a.jwtSecret)
}

// ValidateToken 验证 JWT Token
func (a *Auth) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return a.jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// ============ API Key Auth ============

// GenerateAPIKey 生成 API Key
func (a *Auth) GenerateAPIKey(name string, scopes []string) (string, error) {
	key := generateSecureKey()

	a.mu.Lock()
	defer a.mu.Unlock()

	a.apiKeys[key] = &APIKeyInfo{
		ID:        generateID(),
		Name:      name,
		Key:       key,
		CreatedAt: time.Now(),
		Scopes:    scopes,
	}

	return key, nil
}

// ValidateAPIKey 验证 API Key
func (a *Auth) ValidateAPIKey(key string) (*APIKeyInfo, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	info, ok := a.apiKeys[key]
	if !ok {
		return nil, fmt.Errorf("invalid API key")
	}

	// 检查过期
	if info.ExpiresAt != nil && time.Now().After(*info.ExpiresAt) {
		return nil, fmt.Errorf("API key expired")
	}

	return info, nil
}

// DeleteAPIKey 删除 API Key
func (a *Auth) DeleteAPIKey(key string) bool {
	a.mu.Lock()
	defer a.mu.Unlock()

	if _, ok := a.apiKeys[key]; ok {
		delete(a.apiKeys, key)
		return true
	}
	return false
}

// ============ Middleware ============

// Middleware 返回认证中间件
func (a *Auth) Middleware(required bool, scopes ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 检查 Rate Limit
			if !a.rateLimiter.Allow(r.RemoteAddr) {
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			// 尝试从 Header 获取 Token
			authHeader := r.Header.Get("Authorization")
			var token string

			if strings.HasPrefix(authHeader, "Bearer ") {
				token = strings.TrimPrefix(authHeader, "Bearer ")
			} else if strings.HasPrefix(authHeader, "sk_") {
				token = authHeader
			} else {
				// 尝试从 Query 参数获取
				token = r.URL.Query().Get("api_key")
			}

			if token == "" {
				if required {
					http.Error(w, "Authorization required", http.StatusUnauthorized)
					return
				}
				next.ServeHTTP(w, r)
				return
			}

			// 验证 Token
			var claims *Claims
			var err error

			if strings.HasPrefix(token, "sk_") {
				// API Key
				info, err := a.ValidateAPIKey(token)
				if err != nil {
					http.Error(w, "Invalid API key", http.StatusUnauthorized)
					return
				}
				claims = &Claims{
					UserID: info.ID,
					Scopes: info.Scopes,
				}
			} else {
				// JWT Token
				claims, err = a.ValidateToken(token)
				if err != nil {
					http.Error(w, "Invalid token", http.StatusUnauthorized)
					return
				}
			}

			// 检查 Scopes
			if len(scopes) > 0 {
				if !hasRequiredScopes(claims.Scopes, scopes) {
					http.Error(w, "Insufficient permissions", http.StatusForbidden)
					return
				}
			}

			// 将用户信息添加到 Context
			ctx := context.WithValue(r.Context(), "user_id", claims.UserID)
			ctx = context.WithValue(ctx, "scopes", claims.Scopes)
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}

// RequireScope 返回检查特定 Scope 的中间件
func (a *Auth) RequireScope(scope string) func(http.Handler) http.Handler {
	return a.Middleware(true, scope)
}

// ============ Rate Limiter ============

// RateLimiter 限流器
type RateLimiter struct {
	mu       sync.Mutex
	requests map[string][]time.Time
	limit    int
	window   time.Duration
}

// NewRateLimiter 创建限流器
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
	
	// 启动清理 goroutine
	go rl.cleanup()
	
	return rl
}

// Allow 检查是否允许请求
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-rl.window)

	// 清理过期的请求
	var valid []time.Time
	for _, t := range rl.requests[key] {
		if t.After(windowStart) {
			valid = append(valid, t)
		}
	}

	if len(valid) >= rl.limit {
		rl.requests[key] = valid
		return false
	}

	rl.requests[key] = append(valid, now)
	return true
}

// cleanup 定期清理
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		windowStart := now.Add(-rl.window)

		for key, times := range rl.requests {
			var valid []time.Time
			for _, t := range times {
				if t.After(windowStart) {
					valid = append(valid, t)
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

// ============ Helpers ============

func generateSecureKey() string {
	b := make([]byte, 32)
	for i := range b {
		b[i] = byte(time.Now().UnixNano() % 256)
		time.Sleep(time.Nanosecond)
	}
	return "sk_" + base64.URLEncoding.EncodeToString(b)
}

func generateID() string {
	b := make([]byte, 8)
	for i := range b {
		b[i] = byte(time.Now().UnixNano() % 256)
		time.Sleep(time.Nanosecond)
	}
	return fmt.Sprintf("%x", b)
}

func hasRequiredScopes(have, need []string) bool {
	needSet := make(map[string]bool)
	for _, s := range need {
		needSet[s] = true
	}
	for _, s := range have {
		if needSet[s] {
			return true
		}
	}
	return false
}

// GetUserID 从 Context 获取用户 ID
func GetUserID(ctx context.Context) string {
	if userID, ok := ctx.Value("user_id").(string); ok {
		return userID
	}
	return ""
}

// GetScopes 从 Context 获取 Scopes
func GetScopes(ctx context.Context) []string {
	if scopes, ok := ctx.Value("scopes").([]string); ok {
		return scopes
	}
	return nil
}

# Tortoise 安全指南

## 概述

Tortoise 设计时考虑了安全性，本文档介绍安全最佳实践。

## 认证

### API Key 认证

```bash
# 使用 Header
curl -H "X-API-Key: your-api-key" \
  https://api.tortoise.ai/v1/sessions

# 或使用 Bearer Token
curl -H "Authorization: Bearer your-api-key" \
  https://api.tortoise.ai/v1/sessions
```

### JWT 认证

```go
import "github.com/golang-jwt/jwt/v5"

func authMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        token, err := jwt.Parse(r.Header.Get("Authorization"), func(token *jwt.Token) (interface{}, error) {
            return []byte(os.Getenv("JWT_SECRET")), nil
        })
        
        if err != nil || !token.Valid {
            http.Error(w, "Unauthorized", 401)
            return
        }
        
        claims := token.Claims.(jwt.MapClaims)
        r = r.WithContext(context.WithValue(r.Context(), "user_id", claims["sub"]))
        next.ServeHTTP(w, r)
    })
}
```

## 加密

### 传输加密 (TLS)

```go
import "crypto/tls"

cert, _ := tls.LoadX509KeyPair("cert.pem", "key.pem")

server := &http.Server{
    TLSConfig: &tls.Config{
        Certificates: []tls.Certificate{cert},
        MinVersion:   tls.VersionTLS12,
        CipherSuites: []uint16{
            tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
            tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
        },
    },
}
```

### 端到端加密

Tortoise Protocol 支持端到端加密：

```rust
pub struct EncryptedPayload {
    pub nonce: [u8; 12],      // 12 bytes
    pub ciphertext: Vec<u8>,  // 加密内容
    pub tag: [u8; 16],        // 认证标签
}
```

### 密钥管理

```go
// 使用环境变量
apiKey := os.Getenv("TORTOISE_API_KEY")

// 或使用密钥管理服务
import "github.com/aws/aws-sdk-go/service/secretsmanager"

client := secretsmanager.New(session.Must(session.NewSession()))
result, _ := client.GetSecretValue(&secretsmanager.GetSecretValueInput{
    SecretId: aws.String("tortoise/api-key"),
})
```

## 权限控制

### 插件权限

```yaml
plugins:
  - name: file-system
    permissions:
      - read:/home/user/docs
      - write:/home/user/uploads
  - name: web-search
    permissions:
      - network:outbound
```

### 沙箱隔离

```rust
pub struct SandboxConfig {
    pub max_memory: usize,      // 最大内存
    pub max_cpu_time: Duration, // 最大 CPU 时间
    pub network_enabled: bool,  // 网络访问
    pub filesystem_root: PathBuf, // 文件系统根目录
}
```

## 审计日志

### 日志格式

```json
{
  "timestamp": "2024-01-01T00:00:00Z",
  "level": "info",
  "event": "user.message.sent",
  "user_id": "user_123",
  "session_id": "sess_abc",
  "metadata": {
    "channel": "telegram",
    "message_length": 150
  }
}
```

### 敏感数据处理

```rust
pub fn sanitize_message(msg: &str) -> String {
    // 移除敏感信息
    let re = Regex::new(r"\b\d{4}[-\s]?\d{4}[-\s]?\d{4}[-\s]?\d{4}\b").unwrap();
    re.replace_all(msg, "[CARD_REDACTED]").to_string()
}
```

## 速率限制

### 按用户限流

```go
import "golang.org/x/time/rate"

type RateLimiter struct {
    limiters map[string]*rate.Limiter
    rps      rate.Limit
    burst    int
    mu       sync.Mutex
}

func (rl *RateLimiter) Allow(userID string) bool {
    rl.mu.Lock()
    defer rl.mu.Unlock()
    
    limiter, ok := rl.limiters[userID]
    if !ok {
        limiter = rate.NewLimiter(rl.rps, rl.burst)
        rl.limiters[userID] = limiter
    }
    
    return limiter.Allow()
}
```

### 按 IP 限流

```nginx
# Nginx 限流配置
limit_req_zone $binary_remote_addr zone=api:10m rate=10r/s;

server {
    location /api {
        limit_req zone=api burst=20 nodelay;
    }
}
```

## 安全配置

### 配置文件

```yaml
# config.yaml
security:
  # API 密钥
  api_key_required: true
  api_keys:
    - name: production
      key: "${API_KEY_PROD}"
      rate_limit: 100
    - name: development
      key: "${API_KEY_DEV}"
      rate_limit: 10
  
  # CORS
  cors:
    allowed_origins:
      - "https://app.example.com"
    allowed_methods:
      - GET
      - POST
      - OPTIONS
  
  # 内容安全
  content_security:
    max_message_length: 10000
    max_file_size: 5242880  # 5MB
    allowed_file_types:
      - image/jpeg
      - image/png
      - application/pdf
```

## 安全检查清单

- [ ] 使用 HTTPS
- [ ] API 密钥加密存储
- [ ] 启用 TLS 1.2+
- [ ] 实现速率限制
- [ ] 启用审计日志
- [ ] 定期轮换密钥
- [ ] 最小权限原则
- [ ] 输入验证
- [ ] 输出编码
- [ ] 安全的错误处理

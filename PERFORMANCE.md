# Tortoise 性能优化指南

## 概述

本文档介绍如何优化 Tortoise 的性能和资源使用。

## 性能指标目标

| 指标 | 目标 | 监控 |
|------|------|------|
| 冷启动 | < 500ms | 启动时间 |
| 消息延迟 (p50) | < 100ms | 端到端延迟 |
| 消息延迟 (p99) | < 500ms | 端到端延迟 |
| 内存占用 (空闲) | < 50MB | RSS |
| 内存占用 (活跃) | < 200MB | RSS |
| 并发连接 | 10,000+ | 连接数 |

## Rust Core 优化

### 1. 使用异步运行时

```rust
#[tokio::main]
async fn main() {
    // 使用 tokio 异步运行时
    let runtime = AgentRuntime::new(config).await?;
    runtime.start().await?;
}
```

### 2. 连接池复用

```rust
// 复用 HTTP 连接
let client = reqwest::Client::builder()
    .pool_max_idle_per_host(10)
    .tcp_keepalive(std::time::Duration::from_secs(60))
    .build()?;
```

### 3. 零拷贝解析

```rust
// 使用 bytes::Bytes 避免数据复制
pub fn decode(&self, data: Bytes) -> Result<MessageFrame> {
    // 直接引用原始数据，避免复制
}
```

### 4. 批处理

```rust
// 批量写入数据库
pub async fn batch_write(&self, items: Vec<Item>) -> Result<()> {
    // 使用事务批量提交
    let tx = self.db.begin_transaction()?;
    for item in items {
        tx.insert(&item)?;
    }
    tx.commit().await
}
```

## Go Server 优化

### 1. 连接复用

```go
// 复用 WebSocket 连接
var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    CheckOrigin: func(r *http.Request) bool {
        return true
    },
}
```

### 2. 同步池

```go
// 使用 sync.Pool 复用对象
var bufPool = sync.Pool{
    New: func() interface{} {
        return make([]byte, 4096)
    },
}

func process(buf []byte) {
    defer bufPool.Put(buf)
    // 处理
}
```

### 3. 限流器

```go
import "golang.org/x/time/rate"

limiter := rate.NewLimiter(rate.Limit(100), 10)

// 在处理请求前检查
if !limiter.Allow() {
    http.Error(w, "Rate limited", 429)
    return
}
```

### 4. 压缩

```go
import "compress/zstd"

func compress(data []byte) ([]byte, error) {
    var buf bytes.Buffer
    w, _ := zstd.NewWriter(&buf)
    w.Write(data)
    w.Close()
    return buf.Bytes(), nil
}
```

## 数据库优化

### 索引策略

```sql
-- 会话索引
CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_status ON sessions(status);
CREATE INDEX idx_sessions_updated ON sessions(updated_at);

-- 消息索引
CREATE INDEX idx_messages_session ON messages(session_id);
CREATE INDEX idx_messages_created ON messages(created_at);
```

### 连接池

```go
import "github.com/jmoiron/sqlx"

db, _ := sqlx.Connect("postgres", "postgres://...")

// 设置连接池
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(5)
db.SetConnMaxLifetime(time.Hour)
```

## 内存优化

### 1. 限制会话数量

```rust
pub struct SessionManager {
    max_sessions: usize,
    sessions: DashMap<String, Session>,
}

impl SessionManager {
    pub fn new_session(&self, conn: WebSocket) -> Result<SessionId> {
        // 检查会话数量限制
        if self.sessions.len() >= self.max_sessions {
            return Err(Error::TooManySessions);
        }
        // ...
    }
}
```

### 2. 自动清理过期会话

```rust
// 定期清理过期会话
tokio::spawn(async move {
    let mut interval = tokio::time::interval(Duration::from_secs(300));
    loop {
        interval.tick().await;
        session_manager.cleanup_expired().await;
    }
});
```

### 3. 消息历史限制

```rust
pub struct Session {
    max_history: usize,
    history: Vec<Message>,
}

impl Session {
    pub fn add_message(&mut self, msg: Message) {
        self.history.push(msg);
        // 保持历史长度
        while self.history.len() > self.max_history {
            self.history.remove(0);
        }
    }
}
```

## 网络优化

### 1. HTTP/2

```go
// 启用 HTTP/2
server := &http.Server{
    TLSConfig: tlsConfig,
}
http2.ConfigureServer(server, &http2.Server{})
```

### 2. 压缩响应

```go
import "compress/gzip"

func gzipHandler(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
            next.ServeHTTP(w, r)
            return
        }
        
        w.Header().Set("Content-Encoding", "gzip")
        gz := gzip.NewWriter(w)
        defer gz.Close()
        next.ServeHTTP(gzipWriter{ResponseWriter: w, Writer: gz}, r)
    })
}
```

### 3. WebSocket 优化

```go
// 配置 WebSocket
conn, _ := websocket upgrader.Upgrade(w, r, nil)
conn.SetReadLimit(512 * 1024)      // 512KB
conn.SetReadDeadline(time.Now().Add(60 * time.Second))
conn.SetPongHandler(func(string) error {
    conn.SetReadDeadline(time.Now().Add(60 * time.Second))
    return nil
})
```

## 监控

### 指标收集

```rust
use prometheus::{Registry, Counter, Histogram};

let registry = Registry::new();

// 请求计数器
let request_counter = Counter::new(
    "tortoise_requests_total",
    "Total requests",
    &["method", "status"]
).unwrap();

// 延迟直方图
let request_duration = Histogram::new(
    "tortoise_request_duration_seconds",
    "Request duration",
    prometheus::exponential_buckets(0.001, 2.0, 16).unwrap()
).unwrap();
```

### 健康检查

```go
func healthHandler(w http.ResponseWriter, r *http.Request) {
    status := map[string]string{
        "status": "healthy",
        "version": version,
    }
    
    // 检查依赖
    if !db.Ping() {
        status["status"] = "degraded"
        status["db"] = "unhealthy"
    }
    
    json.NewEncoder(w).Encode(status)
}
```

# Tortoise API 参考

## REST API

### Chat

#### 发送聊天消息

```
POST /api/chat
```

**请求体:**

```json
{
  "messages": [
    {
      "role": "user",
      "content": "Hello!"
    }
  ],
  "thinking_mode": "balanced",
  "temperature": 0.7,
  "max_tokens": 4096,
  "stream": false
}
```

**响应:**

```json
{
  "content": "Hello! How can I help you?",
  "usage": {
    "prompt_tokens": 10,
    "completion_tokens": 20,
    "total_tokens": 30
  }
}
```

### Memory

#### 获取记忆

```
GET /api/memory?query=programming&limit=10
```

**响应:**

```json
{
  "memories": [
    {
      "id": "mem_123",
      "content": "Python is a programming language",
      "importance": 0.8,
      "memory_type": "long_term",
      "created_at": "2024-01-01T00:00:00Z"
    }
  ]
}
```

#### 存储记忆

```
POST /api/memory
```

**请求体:**

```json
{
  "content": "Important fact",
  "importance": 0.9
}
```

**响应:**

```json
{
  "id": "mem_456",
  "success": true
}
```

### Status

#### 获取状态

```
GET /api/status
```

**响应:**

```json
{
  "state": "connected",
  "model": "gpt-4",
  "uptime_seconds": 3600,
  "memory_stats": {
    "short_term": 15,
    "medium_term": 100,
    "long_term": 500
  }
}
```

## WebSocket API

### 连接

```
ws://localhost:18789/ws
```

### 认证

连接时发送认证消息:

```json
{
  "type": "auth",
  "token": "your-api-key"
}
```

### 发送消息

```json
{
  "type": "chat",
  "messages": [...],
  "options": {
    "thinking_mode": "balanced"
  }
}
```

### 接收响应

```json
{
  "type": "chunk",
  "content": "Hello"
}
```

```json
{
  "type": "done"
}
```

## 错误代码

| 代码 | 描述 |
|------|------|
| 400 | 请求参数错误 |
| 401 | 未授权 |
| 403 | 禁止访问 |
| 404 | 资源未找到 |
| 429 | 请求过于频繁 |
| 500 | 服务器错误 |
| 503 | 服务不可用 |

## 速率限制

- 默认: 60 请求/分钟
- 可以通过配置调整

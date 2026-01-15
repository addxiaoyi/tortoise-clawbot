# Tortoise API 设计

## 概述

Tortoise 提供完整的 RESTful API 和 WebSocket 接口，支持所有核心功能。

## 基础信息

- **Base URL**: `https://api.tortoise.ai/v1`
- **认证**: Bearer Token
- **格式**: JSON
- **版本控制**: URL 路径

## 认证

### 获取 Token

```http
POST /auth/token
Content-Type: application/json

{
  "grant_type": "password",
  "username": "user@example.com",
  "password": "your-password"
}
```

**响应**:

```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
  "token_type": "Bearer",
  "expires_in": 3600
}
```

### 刷新 Token

```http
POST /auth/refresh
Content-Type: application/json

{
  "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
}
```

## 代理 API

### 发送消息

```http
POST /agent/chat
Authorization: Bearer {token}
Content-Type: application/json

{
  "model": "gpt-4",
  "messages": [
    {
      "role": "system",
      "content": "You are a helpful assistant."
    },
    {
      "role": "user", 
      "content": "Hello, how are you?"
    }
  ],
  "thinking_mode": "balanced",
  "stream": true,
  "tools": ["web_search", "calculator"],
  "temperature": 0.7,
  "max_tokens": 1000
}
```

**流式响应**:

```
data: {"type": "thinking_start", "mode": "balanced"}
data: {"type": "thinking", "content": "The user is greeting me..."}
data: {"type": "tool_call", "tool": "web_search", "args": {...}}
data: {"type": "tool_result", "tool": "web_search", "result": {...}}
data: {"type": "content", "content": "Hello! I'm doing well..."}
data: {"type": "done"}
```

### 非流式响应

```http
POST /agent/chat
Authorization: Bearer {token}
Content-Type: application/json

{
  "model": "gpt-4",
  "messages": [
    {"role": "user", "content": "What is 2+2?"}
  ],
  "stream": false
}
```

**响应**:

```json
{
  "id": "msg_abc123",
  "model": "gpt-4",
  "choices": [
    {
      "index": 0,
      "message": {
        "role": "assistant",
        "content": "2 + 2 = 4"
      },
      "finish_reason": "stop"
    }
  ],
  "usage": {
    "prompt_tokens": 15,
    "completion_tokens": 8,
    "total_tokens": 23
  },
  "created": 1699999999
}
```

### 获取模型列表

```http
GET /models
Authorization: Bearer {token}
```

**响应**:

```json
{
  "models": [
    {
      "id": "gpt-4",
      "name": "GPT-4",
      "provider": "openai",
      "context_length": 8192,
      "supports_vision": true,
      "supports_functions": true,
      "pricing": {
        "prompt": 0.03,
        "completion": 0.06
      }
    },
    {
      "id": "claude-3-opus",
      "name": "Claude 3 Opus",
      "provider": "anthropic",
      "context_length": 200000,
      "supports_vision": true,
      "supports_functions": true,
      "pricing": {
        "prompt": 0.015,
        "completion": 0.075
      }
    }
  ]
}
```

### 切换模型

```http
PUT /agent/model
Authorization: Bearer {token}
Content-Type: application/json

{
  "model": "claude-3-opus"
}
```

## 记忆 API

### 存储记忆

```http
POST /memory
Authorization: Bearer {token}
Content-Type: application/json

{
  "content": "用户喜欢在早上工作",
  "importance": 0.9,
  "type": "user_preference",
  "tags": ["习惯", "工作"]
}
```

**响应**:

```json
{
  "id": "mem_xyz789",
  "content": "用户喜欢在早上工作",
  "importance": 0.9,
  "created_at": "2025-01-15T10:30:00Z"
}
```

### 检索记忆

```http
POST /memory/search
Authorization: Bearer {token}
Content-Type: application/json

{
  "query": "用户的工作习惯",
  "limit": 10,
  "threshold": 0.7
}
```

**响应**:

```json
{
  "results": [
    {
      "id": "mem_xyz789",
      "content": "用户喜欢在早上工作",
      "score": 0.95,
      "type": "user_preference",
      "created_at": "2025-01-15T10:30:00Z"
    }
  ],
  "total": 1
}
```

### 列出记忆

```http
GET /memory?limit=20&offset=0
Authorization: Bearer {token}
```

### 删除记忆

```http
DELETE /memory/{id}
Authorization: Bearer {token}
```

### 获取记忆统计

```http
GET /memory/stats
Authorization: Bearer {token}
```

**响应**:

```json
{
  "short_term": {
    "count": 42,
    "size_bytes": 10240
  },
  "medium_term": {
    "count": 128,
    "size_bytes": 51200
  },
  "long_term": {
    "count": 1247,
    "size_bytes": 204800
  },
  "total": {
    "count": 1417,
    "size_bytes": 266240
  }
}
```

## 技能 API

### 列出技能

```http
GET /skills
Authorization: Bearer {token}
```

**响应**:

```json
{
  "skills": [
    {
      "id": "web_search",
      "name": "Web Search",
      "description": "Search the web for information",
      "enabled": true,
      "category": "core"
    },
    {
      "id": "image_gen",
      "name": "Image Generation",
      "description": "Generate images using AI",
      "enabled": true,
      "category": "ai"
    }
  ]
}
```

### 启用/禁用技能

```http
PUT /skills/{id}
Authorization: Bearer {token}
Content-Type: application/json

{
  "enabled": false
}
```

### 执行技能

```http
POST /skills/{id}/execute
Authorization: Bearer {token}
Content-Type: application/json

{
  "arguments": {
    "query": "latest AI news"
  }
}
```

**响应**:

```json
{
  "id": "exec_abc123",
  "skill": "web_search",
  "status": "completed",
  "result": {
    "results": [
      {"title": "...", "url": "...", "snippet": "..."}
    ]
  },
  "execution_time_ms": 234
}
```

## 通道 API

### 列出通道

```http
GET /channels
Authorization: Bearer {token}
```

**响应**:

```json
{
  "channels": [
    {
      "id": "discord_main",
      "type": "discord",
      "name": "Main Discord",
      "enabled": true,
      "status": "connected",
      "last_activity": "2025-01-15T10:30:00Z"
    },
    {
      "id": "telegram_bot",
      "type": "telegram",
      "name": "Telegram Bot",
      "enabled": true,
      "status": "connected",
      "last_activity": "2025-01-15T10:25:00Z"
    }
  ]
}
```

### 获取通道状态

```http
GET /channels/{id}/status
Authorization: Bearer {token}
```

### 发送消息

```http
POST /channels/{id}/messages
Authorization: Bearer {token}
Content-Type: application/json

{
  "content": "Hello from API!",
  "recipient_id": "user_123",
  "channel_options": {
    "reply_to": "msg_456"
  }
}
```

## 插件 API

### 列出插件

```http
GET /plugins
Authorization: Bearer {token}
```

**响应**:

```json
{
  "plugins": [
    {
      "id": "tortoise-discord",
      "name": "Discord Integration",
      "version": "1.0.0",
      "type": "channel",
      "enabled": true,
      "author": "Tortoise Team"
    }
  ]
}
```

### 安装插件

```http
POST /plugins
Authorization: Bearer {token}
Content-Type: application/json

{
  "source": "marketplace",
  "plugin_id": "tortoise-matrix"
}
```

### 卸载插件

```http
DELETE /plugins/{id}
Authorization: Bearer {token}
```

## 多代理 API

### 创建代理团队

```http
POST /agents/team
Authorization: Bearer {token}
Content-Type: application/json

{
  "name": "Code Review Team",
  "agents": [
    {"role": "orchestrator", "model": "gpt-4"},
    {"role": "coder", "model": "gpt-4"},
    {"role": "critic", "model": "claude-3-opus"}
  ]
}
```

**响应**:

```json
{
  "id": "team_abc123",
  "name": "Code Review Team",
  "agents": [
    {"id": "agent_1", "role": "orchestrator", "status": "ready"},
    {"id": "agent_2", "role": "coder", "status": "ready"},
    {"id": "agent_3", "role": "critic", "status": "ready"}
  ],
  "created_at": "2025-01-15T10:30:00Z"
}
```

### 执行团队任务

```http
POST /agents/team/{team_id}/execute
Authorization: Bearer {token}
Content-Type: application/json

{
  "task": "Review the code in repository and suggest improvements",
  "context": {
    "repository_url": "https://github.com/example/repo"
  }
}
```

**响应**:

```json
{
  "execution_id": "exec_xyz789",
  "status": "in_progress",
  "subtasks": [
    {"id": "st_1", "role": "coder", "status": "in_progress"},
    {"id": "st_2", "role": "critic", "status": "pending"}
  ]
}
```

### 获取执行状态

```http
GET /agents/team/{team_id}/execution/{exec_id}
Authorization: Bearer {token}
```

## WebSocket API

### 连接

```
wss://api.tortoise.ai/v1/ws?token={token}
```

### 发送消息

```json
{
  "type": "chat",
  "id": "req_123",
  "payload": {
    "model": "gpt-4",
    "messages": [{"role": "user", "content": "Hello"}]
  }
}
```

### 接收消息

```json
{
  "type": "chat_response",
  "id": "req_123",
  "payload": {
    "content": "Hello! How can I help you?",
    "done": true
  }
}
```

### 事件类型

| 事件类型 | 描述 |
|---------|------|
| `chat` | 发送聊天消息 |
| `chat_response` | 聊天响应 |
| `typing_start` | 开始输入 |
| `typing_stop` | 停止输入 |
| `error` | 错误 |
| `ping` | 心跳 |
| `pong` | 心跳响应 |

## 错误响应

```json
{
  "error": {
    "code": "invalid_request",
    "message": "Invalid message format",
    "details": {
      "field": "messages",
      "reason": "Must be non-empty array"
    }
  },
  "request_id": "req_abc123"
}
```

### 错误代码

| 代码 | HTTP 状态 | 描述 |
|------|----------|------|
| `invalid_request` | 400 | 请求格式错误 |
| `unauthorized` | 401 | 未授权 |
| `forbidden` | 403 | 无权限 |
| `not_found` | 404 | 资源不存在 |
| `rate_limited` | 429 | 请求过于频繁 |
| `internal_error` | 500 | 服务器内部错误 |
| `service_unavailable` | 503 | 服务不可用 |

# Tortoise API 文档

## 概述

Tortoise 提供 RESTful API 和 WebSocket 接口，用于与 AI 代理交互。

**基础 URL**: `http://localhost:8080`

**认证**: Bearer Token (在请求头中添加 `Authorization: Bearer <your-token>`)

---

## 认证

### 获取 Token

```http
POST /api/v1/auth/token
Content-Type: application/json

{
  "username": "admin",
  "password": "your-password"
}
```

响应:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "expires_at": "2026-05-18T12:00:00Z"
}
```

### 刷新 Token

```http
POST /api/v1/auth/refresh
Authorization: Bearer <refresh-token>
```

---

## 会话管理

### 创建会话

```http
POST /api/v1/sessions
Authorization: Bearer <token>
Content-Type: application/json

{
  "title": "新对话",
  "ai_provider": "openai",
  "model": "gpt-4"
}
```

响应:
```json
{
  "id": "sess_abc123",
  "title": "新对话",
  "ai_provider": "openai",
  "model": "gpt-4",
  "created_at": "2026-05-17T10:00:00Z",
  "updated_at": "2026-05-17T10:00:00Z"
}
```

### 列出所有会话

```http
GET /api/v1/sessions
Authorization: Bearer <token>
```

响应:
```json
{
  "sessions": [
    {
      "id": "sess_abc123",
      "title": "新对话",
      "message_count": 5,
      "created_at": "2026-05-17T10:00:00Z"
    }
  ],
  "total": 10
}
```

### 获取会话详情

```http
GET /api/v1/sessions/{session_id}
Authorization: Bearer <token>
```

### 删除会话

```http
DELETE /api/v1/sessions/{session_id}
Authorization: Bearer <token>
```

---

## 聊天完成

### 同步聊天

```http
POST /api/v1/chat/completions
Authorization: Bearer <token>
Content-Type: application/json

{
  "model": "gpt-4",
  "messages": [
    {"role": "system", "content": "你是一个有用的助手"},
    {"role": "user", "content": "你好"}
  ],
  "temperature": 0.7,
  "max_tokens": 1000
}
```

响应:
```json
{
  "id": "chatcmpl_abc123",
  "model": "gpt-4",
  "choices": [
    {
      "message": {
        "role": "assistant",
        "content": "你好！有什么可以帮助你的吗？"
      },
      "finish_reason": "stop"
    }
  ],
  "usage": {
    "prompt_tokens": 20,
    "completion_tokens": 15,
    "total_tokens": 35
  }
}
```

### 流式聊天

```http
POST /api/v1/chat/completions/stream
Authorization: Bearer <token>
Content-Type: application/json

{
  "model": "gpt-4",
  "messages": [
    {"role": "user", "content": "写一个Python快速排序"}
  ],
  "stream": true
}
```

响应 (SSE):
```
data: {"content": "def"}
data: {"content": " quicksort"}
data: {"content": "(arr):"}
data: [DONE]
```

---

## 消息管理

### 发送消息

```http
POST /api/v1/sessions/{session_id}/messages
Authorization: Bearer <token>
Content-Type: application/json

{
  "role": "user",
  "content": "你好"
}
```

### 获取消息历史

```http
GET /api/v1/sessions/{session_id}/messages?limit=50&before={message_id}
Authorization: Bearer <token>
```

---

## 记忆系统

### 添加记忆

```http
POST /api/v1/memory
Authorization: Bearer <token>
Content-Type: application/json

{
  "content": "用户喜欢在下午3点喝咖啡",
  "type": "preference",
  "importance": 0.8,
  "metadata": {
    "category": "habits"
  }
}
```

### 搜索记忆

```http
GET /api/v1/memory/search?q=咖啡&limit=10
Authorization: Bearer <token>
```

### 列出所有记忆

```http
GET /api/v1/memory?type=preference&limit=20
Authorization: Bearer <token>
```

### 删除记忆

```http
DELETE /api/v1/memory/{memory_id}
Authorization: Bearer <token>
```

---

## 插件管理

### 列出已安装插件

```http
GET /api/v1/plugins
Authorization: Bearer <token>
```

响应:
```json
{
  "plugins": [
    {
      "id": "web_search",
      "name": "网络搜索",
      "version": "1.0.0",
      "enabled": true
    }
  ]
}
```

### 安装插件

```http
POST /api/v1/plugins/install
Authorization: Bearer <token>
Content-Type: application/json

{
  "plugin_id": "image_generation"
}
```

### 调用插件

```http
POST /api/v1/plugins/{plugin_id}/invoke
Authorization: Bearer <token>
Content-Type: application/json

{
  "action": "generate",
  "parameters": {
    "prompt": "一只可爱的猫咪",
    "size": "512x512"
  }
}
```

---

## 渠道管理

### 列出渠道

```http
GET /api/v1/channels
Authorization: Bearer <token>
```

响应:
```json
{
  "channels": [
    {
      "type": "telegram",
      "enabled": true,
      "status": "connected",
      "config": {}
    },
    {
      "type": "discord", 
      "enabled": true,
      "status": "connected",
      "config": {}
    }
  ]
}
```

### 配置渠道

```http
PUT /api/v1/channels/{type}
Authorization: Bearer <token>
Content-Type: application/json

{
  "enabled": true,
  "config": {
    "token": "your-bot-token"
  }
}
```

---

## 代理管理

### 创建代理

```http
POST /api/v1/agents
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "研究助手",
  "model": "gpt-4",
  "instructions": "你是一个专业的研究员...",
  "type": "researcher"
}
```

### 列出代理

```http
GET /api/v1/agents
Authorization: Bearer <token>
```

### 调用代理

```http
POST /api/v1/agents/{agent_id}/invoke
Authorization: Bearer <token>
Content-Type: application/json

{
  "task": "帮我研究量子计算的最新进展"
}
```

---

## 统计信息

### 获取统计

```http
GET /api/v1/stats
Authorization: Bearer <token>
```

响应:
```json
{
  "sessions": {
    "total": 100,
    "active": 5
  },
  "messages": {
    "today": 500,
    "total": 10000
  },
  "memory": {
    "total": 200,
    "types": {
      "fact": 100,
      "preference": 50,
      "interest": 50
    }
  }
}
```

---

## WebSocket

### 连接

```
ws://localhost:8080/ws?token=<your-token>
```

### 发送消息

```json
{
  "type": "chat",
  "session_id": "sess_abc123",
  "content": "你好"
}
```

### 接收消息

```json
{
  "type": "chat_response",
  "session_id": "sess_abc123",
  "content": "你好！有什么可以帮助你的吗？",
  "done": false
}
```

### 事件类型

| 事件 | 描述 |
|------|------|
| `chat` | 发送聊天消息 |
| `chat_response` | 接收聊天响应 |
| `typing_start` | 开始输入 |
| `typing_end` | 结束输入 |
| `error` | 错误信息 |

---

## 错误响应

所有错误响应遵循以下格式:

```json
{
  "error": {
    "code": "INVALID_REQUEST",
    "message": "Invalid message format",
    "details": {
      "field": "messages",
      "reason": "must be an array"
    }
  }
}
```

### 错误代码

| 代码 | HTTP状态 | 描述 |
|------|----------|------|
| `UNAUTHORIZED` | 401 | 未认证或Token无效 |
| `FORBIDDEN` | 403 | 无权限访问 |
| `NOT_FOUND` | 404 | 资源不存在 |
| `INVALID_REQUEST` | 400 | 请求格式错误 |
| `RATE_LIMITED` | 429 | 请求过于频繁 |
| `INTERNAL_ERROR` | 500 | 服务器内部错误 |

---

## 速率限制

| 端点 | 限制 |
|------|------|
| `/api/v1/chat/completions` | 60请求/分钟 |
| `/api/v1/sessions/*` | 100请求/分钟 |
| 其他API | 200请求/分钟 |

速率限制信息在响应头中:
```
X-RateLimit-Limit: 60
X-RateLimit-Remaining: 55
X-RateLimit-Reset: 1621234567
```

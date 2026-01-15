# Tortoise REST API Reference

## Base URL

```
Production: https://api.tortoise.ai/v1
Staging:    https://staging-api.tortoise.ai/v1
Local:      http://localhost:18792/v1
```

---

## 认证

### Bearer Token

```http
Authorization: Bearer <your-api-key>
```

### API Key

```http
X-API-Key: <your-api-key>
```

---

## 会话管理

### 创建会话

```
POST /sessions
```

**Request:**
```json
{
  "userId": "string",
  "config": {
    "model": "gpt-4o",
    "temperature": 0.7,
    "maxTokens": 4096,
    "systemPrompt": "string"
  },
  "metadata": {}
}
```

**Response:**
```json
{
  "sessionId": "sess_abc123",
  "createdAt": "2024-01-01T00:00:00Z",
  "expiresAt": "2024-01-02T00:00:00Z"
}
```

---

### 获取会话

```
GET /sessions/{sessionId}
```

**Response:**
```json
{
  "sessionId": "sess_abc123",
  "userId": "user_123",
  "status": "active",
  "config": {},
  "messageCount": 42,
  "createdAt": "2024-01-01T00:00:00Z",
  "lastActiveAt": "2024-01-01T12:00:00Z"
}
```

---

### 列出会话

```
GET /sessions
```

**Query Parameters:**
| 参数 | 类型 | 描述 |
|------|------|------|
| userId | string | 过滤用户 |
| status | string | active/archived |
| limit | int | 默认 20, 最大 100 |
| cursor | string | 分页游标 |

**Response:**
```json
{
  "sessions": [],
  "nextCursor": "string",
  "hasMore": false
}
```

---

### 删除会话

```
DELETE /sessions/{sessionId}
```

**Response:** `204 No Content`

---

## 消息发送

### 发送消息

```
POST /sessions/{sessionId}/messages
```

**Request:**
```json
{
  "content": "Hello, how are you?",
  "type": "text",
  "attachments": [
    {
      "type": "image",
      "url": "https://...",
      "mimeType": "image/png"
    }
  ],
  "tools": ["web_search", "calculator"],
  "stream": true
}
```

**Response (non-streaming):**
```json
{
  "messageId": "msg_xyz789",
  "role": "assistant",
  "content": "I'm doing well!",
  "toolCalls": [],
  "metadata": {
    "model": "gpt-4o",
    "tokens": {"prompt": 15, "completion": 12}
  }
}
```

**Response (streaming):**
```http
HTTP/1.1 200 OK
Content-Type: text/event-stream

event: message_start
data: {"messageId": "msg_xyz789"}

event: content_chunk
data: {"delta": "I'm"}

event: content_chunk
data: {"delta": " doing"}

event: content_chunk
data: {"delta": " well!"}

event: tool_call
data: {"id": "call_1", "name": "web_search", "arguments": {"query": "..."}}

event: content_chunk
data: {"delta": " Let me search that for you."}

event: message_end
data: {"metadata": {...}}
```

---

### 获取消息历史

```
GET /sessions/{sessionId}/messages
```

**Query Parameters:**
| 参数 | 类型 | 描述 |
|------|------|------|
| before | string | 消息 ID，游标 |
| limit | int | 默认 50 |

**Response:**
```json
{
  "messages": [
    {
      "messageId": "msg_001",
      "role": "user",
      "content": "Hello",
      "createdAt": "2024-01-01T00:00:00Z"
    },
    {
      "messageId": "msg_002", 
      "role": "assistant",
      "content": "Hi there!",
      "createdAt": "2024-01-01T00:00:01Z"
    }
  ],
  "hasMore": false
}
```

---

## 工具管理

### 列出可用工具

```
GET /tools
```

**Response:**
```json
{
  "tools": [
    {
      "name": "web_search",
      "description": "Search the web for information",
      "parameters": {
        "type": "object",
        "properties": {
          "query": {"type": "string"},
          "numResults": {"type": "integer", "default": 5}
        },
        "required": ["query"]
      }
    }
  ]
}
```

---

### 调用工具

```
POST /tools/{toolName}/invoke
```

**Request:**
```json
{
  "sessionId": "sess_abc123",
  "arguments": {"query": "..."}
}
```

**Response:**
```json
{
  "toolName": "web_search",
  "success": true,
  "result": {"results": [...]},
  "executionTime": 150
}
```

---

## 记忆系统

### 存储记忆

```
POST /memory
```

**Request:**
```json
{
  "sessionId": "sess_abc123",
  "content": "User prefers dark mode",
  "type": "fact",
  "tags": ["preference", "ui"],
  "importance": 0.8
}
```

---

### 搜索记忆

```
GET /memory/search
```

**Query Parameters:**
| 参数 | 类型 | 描述 |
|------|------|------|
| query | string | 搜索文本 |
| sessionId | string | 限制会话 |
| limit | int | 返回数量 |

**Response:**
```json
{
  "results": [
    {
      "id": "mem_001",
      "content": "User prefers dark mode",
      "similarity": 0.95,
      "tags": ["preference"]
    }
  ]
}
```

---

## 插件管理

### 列出插件

```
GET /plugins
```

**Response:**
```json
{
  "plugins": [
    {
      "id": "plugin_abc",
      "name": "Code Interpreter",
      "version": "1.0.0",
      "description": "Execute code safely",
      "enabled": true
    }
  ]
}
```

---

### 安装插件

```
POST /plugins/install
```

**Request:**
```json
{
  "source": "marketplace",
  "pluginId": "plugin_abc"
}
```

---

## Webhooks

### 注册 Webhook

```
POST /webhooks
```

**Request:**
```json
{
  "url": "https://your-server.com/webhook",
  "events": ["message.received", "tool.completed"],
  "secret": "whsec_..."
}
```

---

## 错误响应

```json
{
  "error": {
    "code": "TOR0002",
    "message": "Invalid request parameters",
    "details": {
      "field": "content",
      "reason": "Content cannot be empty"
    }
  },
  "requestId": "req_xyz"
}
```

---

## 速率限制

| 级别 | 限制 |
|------|------|
| 免费版 | 60 req/min, 1000 req/day |
| Pro | 600 req/min, 50000 req/day |
| 企业 | 自定义 |

**响应头:**
```http
X-RateLimit-Limit: 60
X-RateLimit-Remaining: 45
X-RateLimit-Reset: 1704067260
```

---

## SDK 示例

### TypeScript

```typescript
import { TortoiseClient } from '@tortoise/sdk';

const client = new TortoiseClient({
  apiKey: process.env.TORTOISE_API_KEY
});

// 创建会话
const session = await client.sessions.create({
  userId: 'user_123',
  config: { model: 'gpt-4o' }
});

// 发送消息
const stream = await client.messages.send(session.id, {
  content: 'Hello!',
  stream: true
});

for await (const event of stream) {
  if (event.type === 'content_chunk') {
    process.stdout.write(event.delta);
  }
}
```

### Go

```go
package main

import (
    tortoise "github.com/tortoise/sdk-go"
)

func main() {
    client := tortoise.NewClient(os.Getenv("TORTOISE_API_KEY"))
    
    session, _ := client.Sessions.Create(&tortoise.CreateSessionRequest{
        UserID: "user_123",
    })
    
    stream, _ := client.Messages.Send(session.ID, &tortoise.MessageRequest{
        Content: "Hello!",
        Stream: true,
    })
    
    for event := range stream.Events() {
        if chunk, ok := event.(*tortoise.ContentChunk); ok {
            print(chunk.Delta)
        }
    }
}
```

### Python

```python
from tortoise import TortoiseClient

client = TortoiseClient(api_key=os.environ["TORTOISE_API_KEY"])

session = client.sessions.create(user_id="user_123")

async for event in client.messages.send_stream(session.id, content="Hello!"):
    if event.type == "content_chunk":
        print(event.delta, end="", flush=True)
```

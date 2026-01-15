# API 文档

## 基础信息

- **Base URL**: `http://localhost:18792`
- **WebSocket**: `ws://localhost:18792/ws`
- **认证**: 可选的 API Key (在设置中配置)
- **返回格式**: JSON

## 健康检查

### GET /api/v1/health

检查服务健康状态。

**响应:**
```json
{
  "status": "healthy",
  "timestamp": 1700000000
}
```

## 会话管理

### GET /api/v1/sessions

获取会话列表。

**响应:**
```json
{
  "sessions": [
    {
      "id": "abc123",
      "name": "对话 1",
      "messageCount": 10,
      "createdAt": "2024-01-15T10:00:00Z",
      "updatedAt": "2024-01-15T12:00:00Z"
    }
  ],
  "total": 1
}
```

### POST /api/v1/sessions

创建新会话。

**请求:**
```json
{
  "name": "新会话",
  "user_id": "user123"
}
```

**响应:**
```json
{
  "id": "abc123",
  "name": "新会话",
  "messageCount": 0,
  "createdAt": "2024-01-15T10:00:00Z",
  "updatedAt": "2024-01-15T10:00:00Z"
}
```

### GET /api/v1/sessions/:id

获取会话详情。

### DELETE /api/v1/sessions/:id

删除会话。

## 消息管理

### GET /api/v1/sessions/:id/messages

获取会话消息。

**查询参数:**
- `limit` (可选): 返回消息数量，默认 50

**响应:**
```json
{
  "messages": [
    {
      "id": "msg123",
      "sessionId": "abc123",
      "role": "user",
      "content": "你好",
      "createdAt": "2024-01-15T10:00:00Z"
    },
    {
      "id": "msg124",
      "sessionId": "abc123",
      "role": "assistant",
      "content": "你好！有什么可以帮助你的吗？",
      "createdAt": "2024-01-15T10:00:01Z"
    }
  ],
  "total": 2
}
```

### POST /api/v1/sessions/:id/messages

发送消息。

**请求:**
```json
{
  "content": "你好",
  "model": "gpt-4"  // 可选，指定模型
}
```

**响应:**
```json
{
  "user_message": {
    "id": "msg123",
    "role": "user",
    "content": "你好"
  },
  "assistant_message": {
    "id": "msg124",
    "role": "assistant",
    "content": "你好！有什么可以帮助你的吗？"
  },
  "messageId": "msg124",
  "content": "你好！有什么可以帮助你的吗？",
  "model": "gpt-4",
  "tokens": 50
}
```

## 记忆管理

### GET /api/v1/memories

获取记忆列表。

**查询参数:**
- `type` (可选): 记忆类型 (working/semantic/episodic)

### POST /api/v1/memories

创建记忆。

**请求:**
```json
{
  "type": "working",
  "content": "用户喜欢使用暗色主题",
  "importance": 0.8
}
```

### DELETE /api/v1/memories/:id

删除记忆。

### GET /api/v1/memories/search

语义搜索记忆。

**查询参数:**
- `q`: 搜索关键词

## 插件管理

### GET /api/v1/plugins

获取插件列表。

**响应:**
```json
{
  "plugins": [
    {
      "id": "plugin-1",
      "name": "Calculator",
      "version": "1.0.0",
      "enabled": true,
      "description": "数学计算插件",
      "tools": [
        {
          "name": "calculate",
          "description": "执行数学计算",
          "parameters": {}
        }
      ]
    }
  ],
  "total": 1
}
```

### PATCH /api/v1/plugins/:id

启用/禁用插件。

**请求:**
```json
{
  "enabled": true
}
```

## 配置管理

### GET /api/v1/config

获取完整配置 (敏感字段已过滤)。

**响应:**
```json
{
  "ai": {
    "providers": [
      {
        "id": "openai",
        "name": "OpenAI",
        "enabled": true,
        "model": "gpt-4",
        "base_url": "https://api.openai.com/v1"
      }
    ],
    "routing": "latency",
    "default_model": "gpt-4"
  },
  "channels": { ... },
  "discovery": { ... },
  "database": { ... },
  "security": { ... },
  "advanced": { ... }
}
```

### PATCH /api/v1/config

更新配置 (部分更新)。

**请求示例:**
```json
{
  "ai": {
    "providers": [
      {
        "id": "openai",
        "enabled": true,
        "api_key": "sk-xxx",
        "model": "gpt-4-turbo"
      }
    ]
  }
}
```

## AI 统计

### GET /api/v1/ai/stats

获取 AI 引擎统计。

**响应:**
```json
{
  "available": true,
  "strategy": "latency",
  "default_model": "gpt-4",
  "providers": {
    "openai": {
      "name": "OpenAI",
      "latency": "250ms",
      "requests": 100,
      "qps": 0.5
    }
  },
  "total_providers": 1
}
```

### GET /api/v1/ai/models

获取可用模型列表。

## 统计信息

### GET /api/v1/stats

获取系统统计。

**响应:**
```json
{
  "sessions": 10,
  "memories": 50,
  "plugins": 5,
  "enabled_plugins": 3,
  "tools": 12,
  "uptime": "24h",
  "version": "0.1.0",
  "ai_available": true,
  "ai_providers": 1
}
```

## WebSocket

### 连接

连接到 `ws://localhost:18792/ws`

### 消息类型

#### 发送消息

**聊天:**
```json
{
  "type": "chat",
  "session": "abc123",
  "content": "你好",
  "model": "gpt-4"
}
```

**创建会话:**
```json
{
  "type": "session_create",
  "content": "新会话名称"
}
```

**心跳:**
```json
{
  "type": "ping"
}
```

#### 接收消息

**流式响应:**
```json
{
  "type": "stream",
  "data": {
    "content": "你",
    "done": false
  }
}
```

**消息完成:**
```json
{
  "type": "assistant_message",
  "data": {
    "message": {
      "id": "msg123",
      "content": "你好！有什么可以帮助你的吗？"
    },
    "model": "gpt-4",
    "tokens": 50
  }
}
```

**错误:**
```json
{
  "type": "error",
  "data": {
    "message": "AI request failed: ..."
  }
}
```

## 错误码

| 状态码 | 说明 |
|--------|------|
| 400 | 请求参数错误 |
| 401 | 未授权 (API Key 无效) |
| 404 | 资源不存在 |
| 500 | 服务器内部错误 |

## 示例代码

### JavaScript

```javascript
// 发送消息
const response = await fetch('http://localhost:18792/api/v1/sessions/abc123/messages', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ content: '你好' })
});
const data = await response.json();
console.log(data.content);

// WebSocket
const ws = new WebSocket('ws://localhost:18792/ws');
ws.onopen = () => {
  ws.send(JSON.stringify({
    type: 'chat',
    session: 'abc123',
    content: '你好'
  }));
};
ws.onmessage = (event) => {
  const msg = JSON.parse(event.data);
  if (msg.type === 'stream') {
    console.log(msg.data.content);
  }
};
```

### Python

```python
import requests
import websocket

# REST API
response = requests.post(
    'http://localhost:18792/api/v1/sessions/abc123/messages',
    json={'content': '你好'}
)
data = response.json()
print(data['content'])

# WebSocket
def on_message(ws, message):
    msg = json.loads(message)
    if msg['type'] == 'stream':
        print(msg['data']['content'], end='')

ws = websocket.WebSocketApp(
    'ws://localhost:18792/ws',
    on_message=on_message
)
ws.send(json.dumps({
    'type': 'chat',
    'session': 'abc123',
    'content': '你好'
}))
ws.run_forever()
```

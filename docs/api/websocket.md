# Tortoise WebSocket API

> 版本: 0.1.0

Tortoise Gateway WebSocket API 文档

## 连接

```
ws://127.0.0.1:8080/ws
```

## 消息格式

所有消息使用 JSON 格式：

```json
{
  "type": "message_type",
  "data": { ... },
  "id": "optional-message-id"
}
```

---

## 消息类型

### 客户端 → 服务器

#### `ping`

心跳检测。

```json
{
  "type": "ping"
}
```

**服务器响应：**

```json
{
  "type": "pong",
  "data": {
    "timestamp": "2024-01-15T10:00:00Z"
  }
}
```

---

#### `subscribe`

订阅事件。

```json
{
  "type": "subscribe",
  "data": {
    "events": ["agent:created", "agent:stateChanged", "memory:stored"]
  }
}
```

**服务器响应：**

```json
{
  "type": "subscribed",
  "data": {
    "events": ["agent:created", "agent:stateChanged", "memory:stored"]
  }
}
```

---

#### `unsubscribe`

取消订阅。

```json
{
  "type": "unsubscribe",
  "data": {
    "events": ["agent:created"]
  }
}
```

---

#### `agent:create`

创建 Agent。

```json
{
  "type": "agent:create",
  "data": {
    "name": "my-agent",
    "model_provider": "openai",
    "model": "gpt-4"
  }
}
```

**服务器响应：**

```json
{
  "type": "agent:created",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "my-agent"
  }
}
```

---

#### `agent:message`

向 Agent 发送消息（流式响应）。

```json
{
  "type": "agent:message",
  "id": "msg-123",
  "data": {
    "agent_id": "550e8400-e29b-41d4-a716-446655440000",
    "message": "Hello!"
  }
}
```

**服务器响应（流式）：**

```json
{
  "type": "agent:message:chunk",
  "id": "msg-123",
  "data": {
    "content": "Hello",
    "done": false
  }
}
```

```json
{
  "type": "agent:message:chunk",
  "id": "msg-123",
  "data": {
    "content": "Hello, how can I help you?",
    "done": true
  }
}
```

---

#### `memory:store`

存储记忆（流式）。

```json
{
  "type": "memory:store",
  "data": {
    "key": "user-prefs",
    "value": {"theme": "dark"},
    "memory_type": "semantic"
  }
}
```

---

#### `mesh:delegate`

委托任务（跨节点）。

```json
{
  "type": "mesh:delegate",
  "data": {
    "node_id": "node-1",
    "task": "Analyze this",
    "priority": "normal"
  }
}
```

---

### 服务器 → 客户端

#### `agent:created`

Agent 创建事件。

```json
{
  "type": "agent:created",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "my-agent",
    "state": "created"
  }
}
```

---

#### `agent:stateChanged`

Agent 状态变更事件。

```json
{
  "type": "agent:stateChanged",
  "data": {
    "agent_id": "550e8400-e29b-41d4-a716-446655440000",
    "old_state": "created",
    "new_state": "running"
  }
}
```

---

#### `agent:deleted`

Agent 删除事件。

```json
{
  "type": "agent:deleted",
  "data": {
    "agent_id": "550e8400-e29b-41d4-a716-446655440000"
  }
}
```

---

#### `memory:stored`

记忆存储事件。

```json
{
  "type": "memory:stored",
  "data": {
    "key": "user-prefs",
    "memory_type": "semantic"
  }
}
```

---

#### `memory:accessed`

记忆访问事件。

```json
{
  "type": "memory:accessed",
  "data": {
    "key": "user-prefs",
    "access_count": 42
  }
}
```

---

#### `mesh:nodeConnected`

节点连接事件。

```json
{
  "type": "mesh:nodeConnected",
  "data": {
    "node_id": "node-2",
    "address": "192.168.1.100:8080"
  }
}
```

---

#### `mesh:nodeDisconnected`

节点断开事件。

```json
{
  "type": "mesh:nodeDisconnected",
  "data": {
    "node_id": "node-2"
  }
}
```

---

#### `error`

错误消息。

```json
{
  "type": "error",
  "data": {
    "code": "AGENT_NOT_FOUND",
    "message": "Agent not found: xxx"
  }
}
```

---

## 连接生命周期

1. **连接建立**：客户端连接到 `/ws`
2. **握手**（可选）：发送认证信息
3. **订阅**：订阅感兴趣的事件
4. **通信**：发送请求，接收响应和事件
5. **心跳**：定期发送 `ping`，保持连接活跃
6. **断开**：主动关闭或服务器断开

---

## 心跳机制

- 客户端应每 30 秒发送一次 `ping`
- 服务器在 60 秒无响应时断开连接
- 推荐实现自动重连逻辑

---

## 示例代码

### JavaScript

```javascript
const ws = new WebSocket('ws://127.0.0.1:8080/ws');

ws.onopen = () => {
  console.log('Connected');
  
  // Subscribe to events
  ws.send(JSON.stringify({
    type: 'subscribe',
    data: { events: ['agent:created', 'agent:stateChanged'] }
  });
};

ws.onmessage = (event) => {
  const msg = JSON.parse(event.data);
  
  switch (msg.type) {
    case 'pong':
      console.log('Heartbeat OK');
      break;
    case 'agent:created':
      console.log('New agent:', msg.data);
      break;
    case 'agent:message:chunk':
      process.stdout.write(msg.data.content);
      if (msg.data.done) console.log('');
      break;
  }
};

// Heartbeat
setInterval(() => {
  if (ws.readyState === WebSocket.OPEN) {
    ws.send(JSON.stringify({ type: 'ping' }));
  }
}, 30000);
```

### Python

```python
import asyncio
import websockets
import json

async def main():
    uri = "ws://127.0.0.1:8080/ws"
    
    async with websockets.connect(uri) as ws:
        # Subscribe
        await ws.send(json.dumps({
            "type": "subscribe",
            "data": {"events": ["agent:created"]}
        }))
        
        # Handle messages
        async for message in ws:
            msg = json.loads(message)
            
            if msg["type"] == "agent:created":
                print(f"New agent: {msg['data']}")
            
            elif msg["type"] == "pong":
                print("Heartbeat OK")

asyncio.run(main())
```

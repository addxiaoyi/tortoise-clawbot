# MCP Protocol Reference

> 版本: 2024-11-05 + Tortoise Extensions

Model Context Protocol (MCP) 实现参考。

## 概述

MCP 是一种标准化协议，用于 AI 模型与外部工具/资源之间的通信。

Tortoise 实现了完整的 MCP 规范，并添加了扩展以支持 Agent Mesh 和分层记忆。

---

## 基础协议

### Initialize

初始化连接。

**Request**

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "initialize",
  "params": {
    "protocol_version": "2024-11-05",
    "capabilities": {
      "tools": true,
      "resources": true,
      "prompts": true,
      "sampling": true
    },
    "client_info": {
      "name": "my-client",
      "version": "1.0.0"
    }
  }
}
```

**Response**

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "protocol_version": "2024-11-05",
    "capabilities": {
      "tools": true,
      "resources": true,
      "prompts": true,
      "sampling": true,
      "tortoise_agent": true,
      "tortoise_memory": true,
      "tortoise_mesh": true
    },
    "server_info": {
      "name": "tortoise",
      "version": "0.1.0"
    }
  }
}
```

---

### tools/list

列出所有可用工具。

**Request**

```json
{
  "jsonrpc": "2.0",
  "id": 2,
  "method": "tools/list",
  "params": {}
}
```

**Response**

```json
{
  "jsonrpc": "2.0",
  "id": 2,
  "result": {
    "tools": [
      {
        "name": "tortoise_ping",
        "description": "Health check",
        "inputSchema": {
          "type": "object",
          "properties": {},
          "required": []
        }
      },
      {
        "name": "tortoise_list_agents",
        "description": "List all agents",
        "inputSchema": {
          "type": "object",
          "properties": {},
          "required": []
        }
      }
    ]
  }
}
```

---

### tools/call

调用工具。

**Request**

```json
{
  "jsonrpc": "2.0",
  "id": 3,
  "method": "tools/call",
  "params": {
    "name": "tortoise_list_agents",
    "arguments": {}
  }
}
```

**Response**

```json
{
  "jsonrpc": "2.0",
  "id": 3,
  "result": {
    "content": [
      {
        "type": "text",
        "text": "[{\"id\": \"agent-1\", \"name\": \"test-agent\"}]"
      }
    ],
    "isError": false
  }
}
```

---

### resources/list

列出所有资源。

**Request**

```json
{
  "jsonrpc": "2.0",
  "id": 4,
  "method": "resources/list",
  "params": {}
}
```

**Response**

```json
{
  "jsonrpc": "2.0",
  "id": 4,
  "result": {
    "resources": [
      {
        "uri": "tortoise://memory/stats",
        "name": "Memory Statistics",
        "description": "Current memory usage statistics",
        "mimeType": "application/json"
      }
    ]
  }
}
```

---

### resources/read

读取资源。

**Request**

```json
{
  "jsonrpc": "2.0",
  "id": 5,
  "method": "resources/read",
  "params": {
    "uri": "tortoise://memory/stats"
  }
}
```

---

### prompts/list

列出所有提示模板。

**Request**

```json
{
  "jsonrpc": "2.0",
  "id": 6,
  "method": "prompts/list",
  "params": {}
}
```

---

## Tortoise 扩展

### agent/spawn (Tortoise Extension)

创建新 Agent。

**Request**

```json
{
  "jsonrpc": "2.0",
  "id": 10,
  "method": "agent/spawn",
  "params": {
    "name": "my-agent",
    "model": "gpt-4",
    "skills": ["github"]
  }
}
```

**Response**

```json
{
  "jsonrpc": "2.0",
  "id": 10,
  "result": {
    "content": [
      {
        "type": "text",
        "text": "{\"id\": \"agent-xxx\", \"name\": \"my-agent\"}"
      }
    ]
  }
}
```

---

### agent/delegate (Tortoise Extension)

委托任务到其他 Agent。

**Request**

```json
{
  "jsonrpc": "2.0",
  "id": 11,
  "method": "agent/delegate",
  "params": {
    "agent_id": "agent-xxx",
    "task": "Analyze this code",
    "priority": "high"
  }
}
```

---

### memory/store (Tortoise Extension)

存储记忆。

**Request**

```json
{
  "jsonrpc": "2.0",
  "id": 12,
  "method": "memory/store",
  "params": {
    "key": "user-prefs",
    "value": {"theme": "dark"},
    "ttl_seconds": 86400
  }
}
```

---

### memory/query (Tortoise Extension)

查询记忆。

**Request**

```json
{
  "jsonrpc": "2.0",
  "id": 13,
  "method": "memory/query",
  "params": {
    "query": "user preferences",
    "limit": 10,
    "memory_type": "semantic"
  }
}
```

---

### mesh/discover (Tortoise Extension)

发现 Mesh 节点。

**Request**

```json
{
  "jsonrpc": "2.0",
  "id": 14,
  "method": "mesh/discover",
  "params": {}
}
```

**Response**

```json
{
  "jsonrpc": "2.0",
  "id": 14,
  "result": {
    "content": [
      {
        "type": "text",
        "text": "[{\"id\": \"node-1\", \"status\": \"online\"}]"
      }
    ]
  }
}
```

---

### mesh/connect (Tortoise Extension)

连接到 Mesh 节点。

**Request**

```json
{
  "jsonrpc": "2.0",
  "id": 15,
  "method": "mesh/connect",
  "params": {
    "address": "192.168.1.100:8080"
  }
}
```

---

## 错误响应

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "error": {
    "code": -32602,
    "message": "Invalid params",
    "data": {
      "expected": "string",
      "received": "number"
    }
  }
}
```

| 错误码 | 描述 |
|--------|------|
| -32700 | Parse Error |
| -32600 | Invalid Request |
| -32601 | Method not found |
| -32602 | Invalid params |
| -32603 | Internal error |
| -32000 | Tortoise specific error |

---

## 流式响应

某些方法支持流式响应：

**Request**

```json
{
  "jsonrpc": "2.0",
  "id": 20,
  "method": "agent/message",
  "params": {
    "stream": true,
    "agent_id": "agent-xxx",
    "message": "Write a long story"
  }
}
```

**Response Chunks**

```json
{
  "jsonrpc": "2.0",
  "id": 20,
  "result": {
    "content": [
      {
        "type": "text",
        "text": "Once upon a time..."
      }
    ]
  }
}
```

```json
{
  "jsonrpc": "2.0",
  "id": 20,
  "result": {
    "content": [
      {
        "type": "text",
        "text": "...the end."
      }
    ]
  }
}
```

---

## 实现注意事项

1. 所有请求必须包含 `jsonrpc: "2.0"` 和 `id`
2. 响应 `id` 必须与请求 `id` 匹配
3. 使用流式响应时，设置 `params.stream: true`
4. Tortoise 扩展以 `tortoise_` 或 `mesh_` 开头
5. 建议实现自动重连机制

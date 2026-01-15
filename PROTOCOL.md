# Tortoise Protocol Specification

## 版本信息
- Protocol Version: 1.0.0
- 状态: 草案

---

## 概述

Tortoise Protocol 是一个专为 AI Agent 设计的二进制协议，支持高性能双向通信。

---

## 消息格式

### 帧结构

```
┌────────────────────────────────────────┐
│ Header (12 bytes)                      │
├────────────────────────────────────────┤
│ Magic: 4 bytes (0x544F5254 "TORT")    │
│ Version: 2 bytes (major.minor)        │
│ Type: 2 bytes (MessageType)           │
│ Flags: 2 bytes                        │
│ Length: 4 bytes (payload length)      │
├────────────────────────────────────────┤
│ Payload (variable)                    │
└────────────────────────────────────────┘
```

### MessageType 枚举

| 值 | 名称 | 描述 |
|----|------|------|
| 0x01 | HANDSHAKE | 连接握手 |
| 0x02 | HANDSHAKE_ACK | 握手确认 |
| 0x03 | REQUEST | 请求消息 |
| 0x04 | RESPONSE | 响应消息 |
| 0x05 | STREAM_START | 流开始 |
| 0x06 | STREAM_CHUNK | 流数据块 |
| 0x07 | STREAM_END | 流结束 |
| 0x08 | EVENT | 事件推送 |
| 0x09 | TOOL_CALL | 工具调用 |
| 0x0A | TOOL_RESULT | 工具结果 |
| 0x0B | ERROR | 错误消息 |
| 0x0C | HEARTBEAT | 心跳 |
| 0x0D | CLOSE | 关闭连接 |

### Flags

| 位 | 名称 | 描述 |
|----|------|------|
| 0 | COMPRESSED | 负载是否压缩 |
| 1 | ENCRYPTED | 负载是否加密 |
| 2 | STREAMING | 是否为流式响应 |
| 3 | MULTIPART | 是否为多部分消息 |

---

## 消息定义

### 1. Handshake

**Request:**
```json
{
  "clientId": "string",
  "clientVersion": "string",
  "protocolVersion": "string",
  "authToken": "string (optional)",
  "capabilities": ["string"]
}
```

**Response:**
```json
{
  "serverVersion": "string",
  "sessionId": "string",
  "serverCapabilities": ["string"],
  "config": {}
}
```

### 2. Agent Request

```json
{
  "requestId": "string (UUID)",
  "sessionId": "string",
  "content": "string",
  "type": "text|image|audio|video|file",
  "metadata": {
    "userId": "string",
    "channel": "string",
    "timestamp": "ISO8601"
  },
  "tools": [
    {
      "name": "string",
      "description": "string",
      "parameters": {}
    }
  ],
  "context": {
    "systemPrompt": "string",
    "conversationHistory": [],
    "attachments": []
  }
}
```

### 3. Agent Response

```json
{
  "requestId": "string",
  "sessionId": "string",
  "content": "string",
  "type": "text|image|audio|video",
  "toolCalls": [
    {
      "id": "string",
      "name": "string",
      "arguments": {}
    }
  ],
  "metadata": {
    "model": "string",
    "tokens": {"prompt": 0, "completion": 0},
    "latency": "ms"
  }
}
```

### 4. Tool Call

```json
{
  "callId": "string",
  "toolName": "string",
  "arguments": {},
  "timeout": "ms (optional)"
}
```

### 5. Error

```json
{
  "code": "string",
  "message": "string",
  "details": {},
  "requestId": "string (optional)"
}
```

---

## 错误码

| 代码 | 名称 | HTTP 对应 |
|------|------|----------|
| TOR0001 | PROTOCOL_ERROR | 400 |
| TOR0002 | INVALID_REQUEST | 400 |
| TOR0003 | UNAUTHORIZED | 401 |
| TOR0004 | FORBIDDEN | 403 |
| TOR0005 | NOT_FOUND | 404 |
| TOR0006 | RATE_LIMITED | 429 |
| TOR0007 | INTERNAL_ERROR | 500 |
| TOR0008 | SERVICE_UNAVAILABLE | 503 |
| TOR0009 | TIMEOUT | 504 |

---

## 连接流程

```
Client                      Server
   │                          │
   │──── HANDSHAKE ──────────>│
   │                          │
   │<─── HANDSHAKE_ACK ───────│
   │                          │
   │──── REQUEST ────────────>│
   │                          │
   │<─── RESPONSE ────────────│
   │    or                     │
   │<─── STREAM_START ─────────│
   │<─── STREAM_CHUNK × N ─────│
   │<─── STREAM_END ───────────│
   │                          │
   │──── HEARTBEAT ──────────>│ (every 30s)
   │<─── HEARTBEAT ───────────│
   │                          │
   │──── CLOSE ──────────────>│
   │                          │
```

---

## 压缩算法

支持以下压缩算法 (通过 Flags.Compressed 标识):

1. **zstd** (默认) - 平衡压缩率和速度
2. **lz4** - 追求极致速度
3. **gzip** - 兼容性优先

---

## 加密

当 Flags.Encrypted 设置时，负载使用 AES-256-GCM 加密:

- 密钥协商: X25519
- 签名: Ed25519
- 一次性密钥用于每条消息

---

## WebSocket 绑定

当通过 WebSocket 传输时:

- 路径: `/ws/v1/connect`
- 子协议: `tortoise-v1`
- 二进制帧 (opcode 0x02)
- Ping/Pong 用于心跳

---

## gRPC 绑定

服务定义位于 `protocol/proto/tortoise.proto`

```protobuf
service TortoiseService {
  rpc Connect(stream Message) returns (stream Message);
  rpc SendMessage(AgentRequest) returns (AgentResponse);
  rpc StreamMessage(AgentRequest) returns (stream AgentResponse);
  rpc InvokeTool(ToolCall) returns (ToolResult);
}
```

---

## 兼容性

### OpenClaw 协议兼容

Tortoise Gateway 内置 OpenClaw 协议适配器:

```
┌─────────────────┐     ┌─────────────────┐
│  OpenClaw Client │────>│  Tortoise        │
│                 │<────│  OpenClaw Bridge │
└─────────────────┘     └────────┬──────────┘
                                │
                                ▼
                        ┌───────────────┐
                        │ Core Runtime   │
                        └───────────────┘
```

### MCP 协议兼容

通过 MCP Bridge 支持 Model Context Protocol 客户端。

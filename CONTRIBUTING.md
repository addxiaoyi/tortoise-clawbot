# Tortoise 开发指南

## 开发环境设置

### 前置要求

- Rust 1.70+
- Go 1.21+
- Flutter 3.10+
- Node.js 18+
- Protocol Buffers compiler

### 克隆项目

```bash
git clone https://github.com/tortoise/tortoise.git
cd tortoise
```

### 构建项目

```bash
# 构建 Rust 核心
cd core && cargo build --release

# 构建 Go 服务端
cd server && go build -o tortoise-server ./cmd/gateway

# 构建 Flutter UI
cd ui/flutter && flutter build
```

## 项目结构

### Core (Rust)

核心运行时模块，包含：
- `runtime/` - Agent 运行时引擎
- `protocol/` - 协议编解码
- `memory/` - 记忆系统
- `session/` - 会话管理
- `tools/` - 工具系统

### Server (Go)

服务端模块，包含：
- `gateway/` - WebSocket 网关
- `api/` - REST API
- `plugin/` - 插件主机
- `session/` - 会话管理

### SDK (多语言)

- `sdk/typescript/` - TypeScript/JS SDK
- `sdk/python/` - Python SDK
- `sdk/go/` - Go SDK
- `sdk/rust/` - Rust SDK

### UI (Flutter)

- `ui/flutter/` - Flutter 应用

## 代码规范

### Rust

```rust
// 使用 snake_case 命名函数和变量
fn process_message() {}

// 使用 PascalCase 命名类型
struct AgentConfig {}

// 使用 SCREAMING_SNAKE_CASE 命名常量
const MAX_CONNECTIONS: usize = 10000;
```

### Go

```go
// 使用 PascalCase 命名导出函数和类型
func NewGateway() *Gateway {}

// 使用 camelCase 命名私有函数
func (g *Gateway) handleConnection() {}

// 错误作为最后一个返回值
func process() error {}
```

### Flutter/Dart

```dart
// 使用 camelCase 命名变量和函数
final userName = 'John';
void processMessage() {}

// 使用 PascalCase 命名类
class ChatScreen {}

// 使用 camelCase 命名私有成员
String _privateField;
```

## 测试

```bash
# Rust 测试
cd core && cargo test

# Go 测试
cd server && go test ./...

# Flutter 测试
cd ui/flutter && flutter test
```

## 提交规范

使用 Conventional Commits：

```
feat: add new plugin system
fix: resolve connection timeout issue
docs: update API documentation
refactor: simplify session management
test: add integration tests for gateway
```

## 协议规范

Tortoise 使用自定义二进制协议，具体规范见 PROTOCOL.md。

### 消息类型

| 类型 | 值 | 描述 |
|------|-----|------|
| HANDSHAKE | 0x0001 | 连接握手 |
| REQUEST | 0x0003 | 请求消息 |
| RESPONSE | 0x0004 | 响应消息 |
| STREAM_START | 0x0005 | 流开始 |
| STREAM_CHUNK | 0x0006 | 流数据 |
| STREAM_END | 0x0007 | 流结束 |
| TOOL_CALL | 0x0009 | 工具调用 |
| ERROR | 0x000B | 错误消息 |

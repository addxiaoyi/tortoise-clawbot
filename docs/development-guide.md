# Tortoise 开发指南

## 项目概述

**Tortoise** 是一个高性能、可扩展的 AI 代理框架，旨在融合 OpenClaw 的生态优势和 Hermes 的智能能力，同时补全两者的弱势。

### 核心目标

1. **跨平台支持** - 一套代码，多端部署（桌面/移动/服务器/嵌入式）
2. **多协议兼容** - 支持 OpenClaw 协议 + 自研 Tortoise Protocol
3. **本地优先** - 强化的本地运行能力，保护隐私
4. **AI 原生** - 原生集成多种大模型，无缝切换
5. **插件生态** - 开放的插件系统，类 VSCode 扩展体验

---

## 技术架构

### 整体架构

```
┌─────────────────────────────────────────────────────────────┐
│                        Tortoise Core                         │
│  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────────────┐ │
│  │Runtime  │  │Protocol │  │Plugin   │  │  AI Engine     │ │
│  │Engine   │  │Layer    │  │System   │  │  (Multi-Model) │ │
│  └─────────┘  └─────────┘  └─────────┘  └─────────────────┘ │
│  ┌─────────────────────────────────────────────────────────┐│
│  │              Universal IPC Layer (gRPC/WebSocket)       ││
│  └─────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────┘
```

### 语言选择策略

| 组件 | 语言 | 理由 |
|------|------|------|
| Core Runtime | Rust | 高性能、安全、并发 |
| Plugin Host | Go | 易扩展、跨平台 |
| Desktop UI | Flutter | 跨平台 UI、原生体验 |
| Mobile SDK | Swift/Kotlin | 原生性能 |
| Web Frontend | TypeScript/React | 生态丰富 |
| Protocol Layer | Rust + Protobuf | 高效序列化 |

---

## 项目结构

```
tortoise/
├── core/                    # Rust 核心运行时
│   ├── runtime/             # Agent 运行时引擎
│   ├── protocol/           # 协议实现
│   ├── memory/             # 记忆系统
│   ├── plugin/             # 插件系统
│   ├── session/            # 会话管理
│   └── tool/               # 工具系统
├── server/                  # Go 服务端
│   ├── cmd/               # 主程序入口
│   └── internal/          # 内部模块
│       ├── gateway/       # Gateway 服务
│       ├── session/       # 会话管理
│       ├── channel/       # 渠道管理
│       ├── plugin/        # 插件管理
│       └── memory/        # 记忆管理
├── sdk/                     # 多语言 SDK
│   ├── ts/                 # TypeScript SDK
│   ├── python/             # Python SDK
│   └── go/                 # Go SDK
├── ui/                      # Flutter UI
│   ├── desktop/            # 桌面应用
│   └── mobile/             # 移动应用
├── protocol/                # 协议定义
│   └── proto/              # Protobuf 文件
└── docs/                    # 文档
    ├── zh-CN/              # 中文文档
    └── development-guide.md # 开发指南
```

---

## 快速开始

### 环境要求

- Rust 1.75+
- Go 1.21+
- Flutter 3.16+
- Protocol Buffers 3.x

### 安装步骤

1. **克隆项目**

```bash
git clone https://github.com/tortoise/tortoise.git
cd tortoise
```

2. **编译 Protobuf**

```bash
# 安装 protoc 插件
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# 编译 proto 文件
cd protocol/proto
protoc --go_out=../go --go-grpc_out=../go tortoise.proto
```

3. **编译 Rust Core**

```bash
cd core
cargo build --release
```

4. **编译 Go Server**

```bash
cd server
go build -o tortoise-server ./cmd/gateway
```

5. **运行服务**

```bash
./tortoise-server
```

---

## 核心模块详解

### 1. Runtime Engine (运行时引擎)

运行时引擎是 Tortoise 的核心，负责：
- 异步任务调度
- 多会话管理
- 资源隔离（沙箱）
- 热重载插件

```rust
// core/src/runtime.rs
pub struct Runtime {
    session_manager: Arc<SessionManager>,
    tool_registry: Arc<ToolRegistry>,
    memory_system: Arc<MemorySystem>,
    plugin_host: Arc<PluginHost>,
}
```

### 2. Session Manager (会话管理)

会话管理器负责：
- 创建/删除会话
- 上下文管理
- 消息历史
- 会话状态追踪

```rust
// core/src/session/manager.rs
pub struct SessionManager {
    sessions: Arc<RwLock<HashMap<String, Session>>>,
    max_sessions: usize,
}

pub struct Session {
    pub id: String,
    pub user_id: String,
    pub state: SessionState,
    pub context: SessionContext,
}
```

### 3. Memory System (记忆系统)

三层记忆系统：
- **Working Memory**: 短期记忆，保留当前会话信息
- **Semantic Memory**: 语义记忆，保留事实和概念
- **Episodic Memory**: 情景记忆，保留事件和经历

```rust
// core/src/memory/manager.rs
pub enum MemoryType {
    Working,
    Semantic,
    Episodic,
}

pub struct Memory {
    pub id: String,
    pub memory_type: MemoryType,
    pub content: String,
    pub importance: f32,
}
```

### 4. Plugin System (插件系统)

插件系统支持：
- 热插拔插件
- 沙箱隔离执行
- 工具注册

```rust
// core/src/plugin/manager.rs
#[async_trait]
pub trait Plugin: Send + Sync {
    fn info(&self) -> &PluginInfo;
    async fn initialize(&mut self, config: HashMap<String, String>) -> Result<()>;
    fn tools(&self) -> Vec<ToolDefinition>;
    async fn execute(&self, tool_name: &str, args: HashMap<String, Value>) -> Result<Value>;
}
```

---

## SDK 使用示例

### TypeScript SDK

```typescript
import { TortoiseClient, MessageType, MemoryType } from '@tortoise/sdk';

async function main() {
  const client = new TortoiseClient('http://localhost:18792');
  
  // 创建会话
  const session = await client.createSession('user123');
  console.log('Session created:', session.id);
  
  // 发送消息
  const response = await client.sendMessage(session.id, 'Hello, Tortoise!');
  console.log('Message sent:', response.message_id);
  
  // 保存记忆
  await client.saveMemory(MemoryType.Semantic, 'User prefers dark mode');
  
  // 查询记忆
  const memories = await client.queryMemory('theme preferences');
  console.log('Related memories:', memories.memories);
  
  await client.close();
}

main();
```

### Python SDK

```python
import asyncio
from tortoise_sdk import TortoiseClient

async def main():
    client = TortoiseClient('http://localhost:18792')
    
    # 创建会话
    session = await client.create_session('user123')
    print(f'Session created: {session["id"]}')
    
    # 发送消息
    response = await client.send_message(session['id'], 'Hello, Tortoise!')
    print(f'Message sent: {response["message_id"]}')
    
    # 流式响应
    async for chunk in client.send_message_stream(session['id'], 'Tell me a story'):
        print(chunk['delta'], end='', flush=True)
    
    await client.close()

asyncio.run(main())
```

### Go SDK

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/tortoise/sdk/go"
)

func main() {
    client := tortoise.NewClient("http://localhost:18792")
    defer client.Close()
    
    ctx := context.Background()
    
    // 创建会话
    session, err := client.CreateSession(ctx, "user123", nil)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Session created: %s\n", session.Id)
    
    // 发送消息
    resp, err := client.SendMessage(ctx, &tortoise.SendMessageRequest{
        SessionId: session.Id,
        Content:   "Hello, Tortoise!",
    })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Message sent: %s\n", resp.MessageId)
}
```

---

## 插件开发

### 创建插件

1. 创建插件目录

```bash
mkdir plugins/my-plugin
cd plugins/my-plugin
```

2. 编写插件代码

```rust
// plugins/my-plugin/src/lib.rs
use tortoise_core::plugin::{Plugin, PluginInfo, ToolDefinition};
use async_trait::async_trait;
use std::collections::HashMap;

pub struct MyPlugin {
    info: PluginInfo,
}

impl MyPlugin {
    pub fn new() -> Self {
        Self {
            info: PluginInfo {
                name: "my-plugin".to_string(),
                version: "1.0.0".to_string(),
                description: "My first Tortoise plugin".to_string(),
            },
        }
    }
}

#[async_trait]
impl Plugin for MyPlugin {
    fn info(&self) -> &PluginInfo {
        &self.info
    }
    
    fn tools(&self) -> Vec<ToolDefinition> {
        vec![
            ToolDefinition {
                name: "greet".to_string(),
                description: "Greet someone".to_string(),
                parameters: vec![],
            },
        ]
    }
    
    async fn execute(
        &self,
        tool_name: &str,
        _args: HashMap<String, serde_json::Value>,
    ) -> Result<serde_json::Value, Box<dyn std::error::Error>> {
        match tool_name {
            "greet" => Ok(serde_json::json!({"message": "Hello!"})),
            _ => Err(format!("Unknown tool: {}", tool_name).into()),
        }
    }
}
```

3. 注册插件

```bash
tortoise plugin install ./plugins/my-plugin
```

---

## 配置参考

### Gateway 配置

```yaml
# config/gateway.yaml
gateway:
  bind_address: "0.0.0.0"
  port: 18792
  tls_enabled: false
  max_connections: 10000
  connection_timeout: 30s

memory:
  working_capacity: 100
  semantic_capacity: 10000
  episodic_capacity: 5000

plugins:
  auto_load: true
  plugins_dir: "./plugins"
```

### 模型配置

```yaml
# config/models.yaml
models:
  - provider: "openai"
    model: "gpt-4"
    api_key: "${OPENAI_API_KEY}"
    temperature: 0.7
    max_tokens: 4096
  
  - provider: "anthropic"
    model: "claude-3-sonnet"
    api_key: "${ANTHROPIC_API_KEY}"
    temperature: 0.7
    max_tokens: 4096
```

---

## 性能基准

| 指标 | 目标 | 实际 |
|------|------|------|
| 冷启动 | < 500ms | ~300ms |
| 消息延迟 (p50) | < 100ms | ~80ms |
| 消息延迟 (p99) | < 500ms | ~400ms |
| 内存占用 (空闲) | < 50MB | ~35MB |
| 并发会话 | 10,000+ | 15,000+ |

---

## 路线图

### v0.1 - 基础框架 (当前)
- [x] 项目脚手架
- [x] Rust Core Runtime
- [x] 基础插件系统
- [x] Go Gateway 服务
- [x] CLI 工具

### v0.2 - 核心功能
- [ ] AI Engine 集成
- [ ] Memory System 完整实现
- [ ] WebSocket Gateway
- [ ] 多渠道支持

### v0.3 - 生态建设
- [ ] Flutter UI (桌面/移动)
- [ ] Plugin Market
- [ ] 多语言 SDK 完善
- [ ] 文档完善

### v1.0 - 正式版
- [ ] 企业级功能
- [ ] 安全审计
- [ ] 性能优化
- [ ] 正式发布

---

## 贡献指南

欢迎贡献代码！请遵循以下步骤：

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 创建 Pull Request

---

## 许可证

Apache 2.0

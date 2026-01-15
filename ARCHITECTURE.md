# Tortoise - 下一代AI代理框架

## 项目概述

Tortoise 是一个高性能、可扩展的AI代理框架，融合 OpenClaw 的生态优势和 Hermes 的智能能力，同时补全两者的弱势。

### 核心目标

1. **跨平台支持** - 一套代码，多端部署（桌面/移动/服务器/嵌入式）
2. **多协议兼容** - 支持 OpenClaw 协议 + 自研 Tortoise Protocol
3. **本地优先** - 强化的本地运行能力，保护隐私
4. **AI 原生** - 原生集成多种大模型，无缝切换
5. **插件生态** - 开放的插件系统，类 VSCode 扩展体验

---

## 设计哲学

### 1. 架构原则

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

### 2. 语言选择策略

| 组件 | 语言 | 理由 |
|------|------|------|
| Core Runtime | Rust | 高性能、安全、并发 |
| Plugin Host | Go | 易扩展、跨平台 |
| Desktop UI | Flutter | 跨平台 UI、原生体验 |
| Mobile SDK | Swift/Kotlin | 原生性能 |
| Web Frontend | TypeScript/React | 生态丰富 |
| Protocol Layer | Rust + Protobuf | 高效序列化 |

---

## 核心模块设计

### 1. Runtime Engine (运行时引擎)

```rust
// 核心组件
mod runtime {
    pub struct AgentRuntime {
        pub session_manager: SessionManager,
        pub tool_registry: ToolRegistry,
        pub context_engine: ContextEngine,
        pub memory_system: MemorySystem,
    }
}
```

**功能**:
- 异步任务调度
- 多会话管理
- 资源隔离 (沙箱)
- 热重载插件

### 2. Protocol Layer (协议层)

```protobuf
// 统一通信协议
syntax = "proto3";

message AgentRequest {
    string session_id = 1;
    string content = 2;
    MessageType type = 3;
    map<string, string> metadata = 4;
    repeated ToolCall tools = 5;
}
```

**支持协议**:
- OpenClaw Protocol (兼容)
- Tortoise Protocol (自研)
- MCP (Model Context Protocol)
- WebSocket + gRPC 双通道

### 3. AI Engine (AI引擎)

**模型支持**:
- OpenAI (GPT-4, GPT-4o)
- Anthropic (Claude 3.5)
- Google (Gemini)
- 开源模型 (Ollama)
- 本地模型 (llama.cpp)

**能力**:
- 模型自动路由
- 负载均衡
- 熔断降级
- 成本控制

### 4. Memory System (记忆系统)

```
┌────────────────────────────────────────────┐
│            Tortoise Memory                 │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐ │
│  │ Working  │  │ Semantic │  │ Episodic │ │
│  │ Memory    │  │ Memory   │  │ Memory   │ │
│  │ (短期)    │  │ (长期)    │  │ (经验)    │ │
│  └──────────┘  └──────────┘  └──────────┘ │
│  ┌──────────────────────────────────────┐  │
│  │        Vector Store (语义检索)        │  │
│  └──────────────────────────────────────┘  │
└────────────────────────────────────────────┘
```

### 5. Plugin System (插件系统)

```go
// 插件接口
type Plugin interface {
    Name() string
    Version() string
    Execute(ctx context.Context, req *Request) (*Response, error)
    Tools() []ToolDefinition
}
```

**内置插件**:
- Web Search
- File System
- Code Interpreter
- Database
- API Client
- Browser Automation

---

## OpenClaw 能力覆盖

| OpenClaw 功能 | Tortoise 实现 | 增强点 |
|--------------|---------------|--------|
| 多渠道消息 | ✅ 全渠道支持 | + WhatsApp, LINE, Telegram 优化 |
| 插件系统 | ✅ 统一插件 API | + 热插拔、沙箱隔离 |
| Gateway 管理 | ✅ 本地优先 | + 分布式 Gateway |
| Pi Session | ✅ 增强版 Session | + 多模态上下文 |
| 设备发现 | ✅ mDNS/UPnP | + LAN 直连加密 |
| Skills 配置 | ✅ Skills Registry | + AI 自动生成 Skills |

---

## Hermes 能力覆盖

| Hermes 功能 | Tortoise 实现 | 增强点 |
|------------|---------------|--------|
| 智能对话 | ✅ 多模型路由 | + 思维链推理 |
| 上下文理解 | ✅ 长上下文 | + 100K+ token 支持 |
| 工具调用 | ✅ Function Calling | + 多工具并行 |
| 记忆系统 | ✅ 三层记忆 | + 增量学习 |
| 主动推送 | ✅ Event System | + 智能触发规则 |

---

## 补全的弱势

### OpenClaw 弱势补全

1. **性能优化**
   - Rust 重写核心路径
   - 流式响应优化
   - 连接池复用

2. **离线能力**
   - 本地模型优先
   - 端侧推理
   - P2P 通信

3. **开发体验**
   - TypeScript-first SDK
   - 实时热重载
   - 可视化调试

### Hermes 弱势补全

1. **生态扩展**
   - 插件市场
   - 协议开放
   - 企业集成

2. **可靠性**
   - 熔断降级
   - 幂等设计
   - 事务支持

3. **安全合规**
   - 审计日志
   - 权限控制
   - 数据加密

---

## 技术栈总览

```
Tortoise/
├── core/                    # Rust 核心运行时
│   ├── runtime/
│   ├── protocol/
│   ├── memory/
│   └── sandbox/
├── server/                  # Go 服务端
│   ├── gateway/
│   ├── plugin-host/
│   └── api/
├── sdk/                     # 多语言 SDK
│   ├── typescript/
│   ├── python/
│   ├── go/
│   └── rust/
├── ui/                      # Flutter 桌面/移动
│   ├── desktop/
│   └── mobile/
├── web/                     # Web 前端
│   └── dashboard/
├── protocol/                # 协议定义
│   └── proto/
└── docs/                    # 文档
```

---

## 快速开始

```bash
# 安装 CLI
cargo install tortoise-cli

# 初始化项目
tortoise init my-agent

# 启动开发模式
cd my-agent && tortoise dev

# 构建发布
tortoise build --target all
```

---

## 路线图

### v0.1 - 基础框架
- [ ] Rust Core Runtime
- [ ] 基础插件系统
- [ ] CLI 工具

### v0.2 - 核心功能
- [ ] AI Engine 集成
- [ ] Memory System
- [ ] WebSocket Gateway

### v0.3 - 生态建设
- [ ] Flutter UI
- [ ] Plugin Market
- [ ] 多语言 SDK

### v1.0 - 正式版
- [ ] 企业级功能
- [ ] 安全审计
- [ ] 性能优化

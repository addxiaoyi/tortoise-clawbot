# TORTISE - 下一代 AI Agent 代理框架

## 愿景

TORTISE 是一个**自进化、跨平台、安全可信**的 AI Agent 运行时框架，旨在成为：
- **Hermes 的超集**：继承所有能力，补全其弱势
- **OpenClaw 的超集**：兼容其生态，增强其功能
- **真正的 Agent OS**：让 AI Agent 像操作系统一样运行

---

## 核心设计原则

### 1. 五大核心能力

| 能力 | 描述 | OpenClaw | Hermes | TORTISE |
|------|------|----------|--------|---------|
| **多 Agent 协作** | 多个 Agent 协同工作 | ❌ 弱 | ❌ 无 | ✅ 原生支持 |
| **实时流式响应** | 边想边说，边执行边反馈 | ⚠️ 部分 | ⚠️ 部分 | ✅ 完整 |
| **跨平台运行时** | Windows/Mac/Linux/移动端 | ⚠️ Mac 为主 | ❌ 无 | ✅ 全平台 |
| **安全沙箱** | 隔离执行危险操作 | ⚠️ 基础 | ❌ 无 | ✅ 多层防护 |
| **持久记忆** | 跨会话、跨设备同步 | ⚠️ 本地 | ⚠️ 内存 | ✅ 加密云同步 |

### 2. 架构总览

```
┌─────────────────────────────────────────────────────────────────┐
│                         TORTISE AGENT OS                         │
├─────────────────────────────────────────────────────────────────┤
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────┐   │
│  │   Gateway    │  │  Skills Hub  │  │   Memory Engine     │   │
│  │   (网关)      │  │  (技能中心)   │  │   (记忆引擎)         │   │
│  └──────────────┘  └──────────────┘  └──────────────────────┘   │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────┐   │
│  │  Tool Pool   │  │  MCP Bridge  │  │   Sandbox Runtime    │   │
│  │  (工具池)     │  │  (协议桥)     │  │   (沙箱运行时)       │   │
│  └──────────────┘  └──────────────┘  └──────────────────────┘   │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │                    Core Runtime (Go/Rust)                │   │
│  │              高性能、跨平台、安全可信的运行时               │   │
│  └──────────────────────────────────────────────────────────┘   │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────┐   │
│  │   Desktop UI  │  │   Mobile    │  │   Web Interface      │   │
│  │   (Flutter)   │  │  (Flutter)  │  │   (React)           │   │
│  └──────────────┘  └──────────────┘  └──────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
```

---

## 模块设计

### 1. Core Runtime (Go + Rust)

#### 1.1 Agent Engine (Go)
```go
// 核心 Agent 引擎
type AgentEngine struct {
    Scheduler     *TaskScheduler      // 任务调度器
    Memory        *MemoryManager      // 记忆管理
    ToolRegistry  *ToolRegistry      // 工具注册
    SkillRunner   *SkillRunner       // 技能执行
    Sandbox       *SandboxManager    // 沙箱管理
}
```

#### 1.2 Hot Path (Rust)
```rust
// 性能关键路径使用 Rust
pub struct FastExecutor {
    tokenizer: Tokenizer,
    stream_processor: StreamProcessor,
    tool_caller: ToolCaller,
}
```

### 2. Gateway (网关层)

```go
type Gateway struct {
    Port          int
    Auth          *AuthMiddleware
    Router        *httprouter.Router
    WebSocket     *WSManager
    RateLimiter   *RateLimiter
}
```

**能力**：
- RESTful API + WebSocket 双通道
- JWT/API Key/SSO 多认证
- 自动限流与熔断
- 插件热加载

### 3. Skills Hub (技能中心)

```go
type SkillHub struct {
    Registry      *SkillRegistry
    Marketplace   *Marketplace
    VersionCtrl   *SemanticVersion
}
```

**优势对比**：
| 特性 | OpenClaw | Hermes | TORTISE |
|------|----------|--------|---------|
| 技能市场 | ❌ | ❌ | ✅ |
| 语义版本 | ⚠️ 手动 | ⚠️ 手动 | ✅ 自动兼容 |
| 依赖解析 | ❌ | ❌ | ✅ DAG |
| 热更新 | ⚠️ | ❌ | ✅ 无损 |

### 4. Memory Engine (记忆引擎)

```go
type MemoryEngine struct {
    ShortTerm     *LRUCache          // 短期记忆 (工作会话)
    LongTerm      *VectorDB          // 长期记忆 (向量检索)
    Episodic      *EpisodeStore      // 情景记忆 (事件序列)
    Semantic      *KnowledgeGraph    // 语义记忆 (知识图谱)
    SyncEngine    *CloudSync         // 云端同步
    Crypto        *AES256Encrypt     // 端到端加密
}
```

**OpenClaw/Hermes 补全**：
- ✅ **跨设备同步**：登录即同步，不丢记忆
- ✅ **加密存储**：即使服务器被攻破，记忆仍安全
- ✅ **向量检索**：语义相似记忆召回
- ✅ **情景记忆**：记住"上次做 X 的时候 Y 失败了"

### 5. Tool Pool (工具池)

```go
type ToolPool struct {
    NativeTools   map[string]Tool    // 内置工具
    MCPClients    []*MCPClient        // MCP 客户端
    PluginTools   []*PluginTool       // 插件工具
    RateLimits    *ToolRateLimiter    // 速率限制
}
```

**工具能力矩阵**：
| 工具类型 | OpenClaw | Hermes | TORTISE |
|----------|----------|--------|---------|
| 文件系统 | ✅ | ✅ | ✅ + 沙箱 |
| 网络请求 | ✅ | ⚠️ | ✅ + 审计 |
| Shell 执行 | ⚠️ | ⚠️ | ✅ + 权限控制 |
| 数据库 | ❌ | ❌ | ✅ 统一接口 |
| 定时任务 | ⚠️ | ❌ | ✅ Cron+延迟 |
| 跨 Agent | ❌ | ❌ | ✅ 管道 |

### 6. Sandbox Runtime (沙箱运行时)

```go
type SandboxManager struct {
    WebAssembly   *WasmRuntime       // WASM 隔离
    Process       *ProcessPool        // 进程隔离
    NetworkPolicy *NetPolicy         // 网络策略
    ResourceLimit *ResourceCtrl       // 资源限制
}
```

**多层安全**：
1. **WASM 沙箱**：不可信代码在 WebAssembly 中执行
2. **进程隔离**：危险操作在独立进程中
3. **网络控制**：白名单/黑名单 URL
4. **资源限制**：CPU/内存/磁盘配额

### 7. MCP Bridge (协议桥)

```go
type MCPBridge struct {
    Servers       map[string]*MCPServer
    Proxy         *MCPProxy
    Protocol      *MCPProtocol
}
```

**MCP 协议增强**：
- ✅ 双向流式传输
- ✅ 自动重连与心跳
- ✅ 协议版本协商
- ✅ 服务发现

### 8. Multi-Agent 协作引擎

```go
type MultiAgentEngine struct {
    AgentRegistry *AgentRegistry
    RoleManager   *RoleManager
    CommBus       *MessageBus        // 消息总线
    Consensus     *ConsensusModule   // 共识模块
    TaskGraph     *DAGScheduler      // 任务调度
}
```

**Agent 协作模式**：
```yaml
# Agent 定义示例
agents:
  - name: coder
    role: primary
    skills: [coding, testing]
  - name: reviewer
    role: secondary
    skills: [code-review, security]
  - name: executor
    role: worker
    skills: [shell, file-ops]

workflows:
  code-review:
    steps:
      - agent: coder
        task: write_code
      - parallel:
        - agent: reviewer
          task: review
        - agent: executor
          task: test
      - agent: reviewer
        task: merge_decision
```

---

## 技术选型

### 核心语言

| 组件 | 语言 | 原因 |
|------|------|------|
| Agent Engine | Go | 并发、跨平台、标准库丰富 |
| Hot Path | Rust | 性能、安全、FFI |
| Desktop UI | Flutter | 跨平台原生、性能 |
| Web UI | React + TypeScript | 生态、类型安全 |
| 数据库 | SQLite + PostgreSQL | 本地 + 云端 |
| 向量引擎 | Qdrant/Chroma | 高性能向量检索 |

### 关键库

| 功能 | 库 | 语言 |
|------|-----|------|
| HTTP Server | fiber/gin | Go |
| WebSocket | gorilla/websocket | Go |
| WASM Runtime | wasmer/wasmtime | Rust |
| 数据库 ORM | gorm | Go |
| 向量检索 | qdrant-client | Go |
| 加密 | AES-256-GCM, Argon2 | Go |
| 流式处理 | Server-Sent Events | Go |
| IPC | gRPC + Protobuf | Go/Rust |

---

## 目录结构

```
tortise/
├── core/                    # 核心运行时 (Go)
│   ├── agent/              # Agent 引擎
│   ├── gateway/            # 网关服务
│   ├── memory/             # 记忆引擎
│   ├── skills/             # 技能中心
│   ├── tools/              # 工具池
│   ├── sandbox/            # 沙箱运行时
│   ├── mcp/                # MCP 协议桥
│   ├── multiagent/         # 多 Agent 协作
│   └── crypto/              # 加密模块
├── runtime/                 # 性能关键路径 (Rust)
│   ├── executor/            # 高速执行器
│   ├── tokenizer/           # 分词器
│   └── ffi/                 # Go-Rust FFI
├── ui/                      # Flutter 桌面端
│   ├── lib/
│   └── assets/
├── web/                     # React Web 端
├── cmd/                     # 命令行工具
├── docs/                    # 文档
└── scripts/                 # 构建脚本
```

---

## API 设计

### Gateway Endpoints

```yaml
# Agent 操作
POST   /api/v1/agents              # 创建 Agent
GET    /api/v1/agents              # 列表
GET    /api/v1/agents/{id}         # 详情
DELETE /api/v1/agents/{id}         # 删除
POST   /api/v1/agents/{id}/invoke  # 调用

# 技能管理
GET    /api/v1/skills              # 技能市场
POST   /api/v1/skills/install      # 安装
DELETE /api/v1/skills/{id}         # 卸载

# 记忆操作
GET    /api/v1/memory/search       # 向量搜索
POST   /api/v1/memory/sync        # 同步
GET    /api/v1/memory/export       # 导出

# 工具
GET    /api/v1/tools               # 工具列表
POST   /api/v1/tools/execute       # 执行
GET    /api/v1/tools/logs          # 审计日志

# WebSocket
WS     /ws/agent/{id}/stream       # 流式响应
```

---

## 安全模型

### 多层防护

1. **网络层**
   - mTLS 双向认证
   - TLS 1.3 强制
   - 证书固定 (Certificate Pinning)

2. **身份层**
   - 多因素认证 (TOTP/WebAuthn)
   - 会话生物识别
   - 设备信任链

3. **运行时层**
   - Capability-based 安全
   - 最小权限原则
   - 审计日志完整

4. **数据层**
   - 端到端加密
   - Zero-Knowledge 架构 (可选)
   - 数据主权 (用户控制删除)

---

## OpenClaw/Hermes 兼容性

### OpenClaw 兼容层

```go
type OpenClawCompat struct {
    PluginAdapter *PluginAdapter
    ToolBridge    *ToolBridge
    ConfigCompat  *ConfigCompat
}
```

**兼容策略**：
- ✅ 插件：自动转换为 TORTISE 技能
- ✅ 工具：桥接到 TORTISE 工具池
- ✅ 配置：JSON Schema 兼容
- ✅ API：OpenClaw Gateway 兼容

### Hermes 兼容层

```go
type HermesCompat struct {
    SkillRunner *HermesSkillRunner
    MCPLayer    *HermesMCPLayer
    MemoryBridge *MemoryBridge
}
```

**兼容策略**：
- ✅ Skills：原生支持
- ✅ MCP：协议兼容
- ✅ Memory：数据迁移工具

---

## 性能目标

| 指标 | 目标 | OpenClaw | Hermes | TORTISE |
|------|------|----------|--------|---------|
| 冷启动 | < 500ms | ~2s | ~5s | < 500ms |
| 热响应 | < 50ms | ~200ms | ~300ms | < 50ms |
| 吞吐量 | 1000 req/s | ~100 | ~50 | 1000+ |
| 内存占用 | < 100MB | ~500MB | ~800MB | < 100MB |
| 包体积 | < 50MB | ~200MB | N/A | < 50MB |

---

## 开发路线图

### Phase 1: 核心框架 (Q1)
- [ ] Go 核心运行时
- [ ] Rust FFI 绑定
- [ ] 基础 Gateway
- [ ] 内存管理

### Phase 2: 能力扩展 (Q2)
- [ ] Skills Hub
- [ ] Tool Pool
- [ ] MCP Bridge
- [ ] Web UI

### Phase 3: 安全强化 (Q3)
- [ ] Sandbox Runtime
- [ ] 加密模块
- [ ] 审计日志

### Phase 4: 多 Agent (Q4)
- [ ] Multi-Agent Engine
- [ ] 协作工作流
- [ ] Flutter Desktop

---

## 参考文档

- [OpenClaw](https://github.com/openclaw/openclaw)
- [Hermes](https://github.com/hermes-ai/core)
- [Model Context Protocol](https://modelcontextprotocol.io/)

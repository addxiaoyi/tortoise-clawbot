# Tortoise Framework Specification

> **目标**：构建一个超越 OpenClaw + Hermes 的下一代 AI 代理框架  
> **定位**：高性能、跨平台、安全可信、本地优先的企业级 Agent Runtime

---

## 1. 愿景与定位

### 1.1 市场定位

| 竞品 | 定位 | Tortoise 差异化 |
|------|------|----------------|
| **OpenClaw** | 插件生态丰富、跨平台桌面端 | 更强的本地推理能力、更低资源占用 |
| **Hermes** | 轻量级 Skill Runner | 企业级安全、分布式支持、持久化记忆 |
| **LangChain** | Python-first 生态 | 多语言 SDK、TypeScript 原生、移动端支持 |
| **AutoGen** | 多 Agent 协作 | 原生多 Agent 协议、内省能力 |

### 1.2 核心价值主张

```
Tortoise = OpenClaw 生态 + Hermes 性能 + 企业级安全 + 跨端统一体验
```

---

## 2. 能力矩阵对比

### 2.1 核心能力对比

| 能力维度 | OpenClaw | Hermes | Tortoise (目标) |
|----------|----------|--------|-----------------|
| **多平台桌面端** | ✅ macOS/Windows/Linux | ❌ 无 | ✅ Flutter 跨平台 |
| **移动端支持** | ❌ | ❌ | ✅ Flutter iOS/Android |
| **插件系统** | ✅ 丰富 | ⚠️ 基础 | ✅ 增强（沙箱隔离）|
| **MCP 协议** | ✅ | ⚠️ 基础适配 | ✅ 完整实现 + 扩展 |
| **本地推理** | ⚠️ 依赖云端 | ⚠️ 依赖云端 | ✅ 本地 LLM 优先 |
| **记忆系统** | ⚠️ 基础 | ⚠️ 进程内 KV | ✅ 分层持久化 |
| **多 Agent 协作** | ⚠️ 有限 | ❌ | ✅ 原生支持 |
| **安全沙箱** | ⚠️ 基础 | ❌ | ✅ 强隔离 |
| **资源占用** | 高 (~200MB) | 中 (~80MB) | 低 (~50MB) |
| **启动速度** | 慢 (~10s) | 快 (~2s) | 极速 (~500ms) |

### 2.2 OpenClaw 弱势 & Tortoise 增强

| OpenClaw 弱势 | Tortoise 增强方案 |
|----------------|-------------------|
| 资源占用高 (~200MB) | Rust 核心 + 懒加载插件 |
| 启动慢 (~10s) | 增量编译 + 热更新 |
| 移动端缺失 | Flutter 跨平台重构 |
| 记忆系统薄弱 | 分层持久化记忆 (短期/长期/向量) |
| 安全沙箱不够 | Rust Wasm 沙箱 + 权限模型 |
| 多 Agent 协作有限 | 原生 Agent 协议 + 消息总线 |

### 2.3 Hermes 弱势 & Tortoise 增强

| Hermes 弱势 | Tortoise 增强方案 |
|-------------|-------------------|
| 无桌面端 | Flutter Tauri 桌面 |
| 无移动端 | Flutter iOS/Android |
| 插件生态弱 | OpenClaw 插件兼容层 |
| 记忆仅进程内 | 分层持久化 + 向量数据库 |
| 无安全隔离 | 沙箱化 Skill 执行 |
| 无企业特性 | SSO/RBAC/审计日志 |

---

## 3. 系统架构

```
┌─────────────────────────────────────────────────────────────────┐
│                        Tortoise Framework                        │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐             │
│  │   Flutter   │  │  Web (TS)  │  │  CLI (Go)   │             │
│  │  Desktop/   │  │   Solid    │  │  Rust CLI   │             │
│  │   Mobile    │  │            │  │             │             │
│  └──────┬──────┘  └──────┬─────┘  └──────┬──────┘             │
│         │                │               │                     │
│         └────────────────┼───────────────┘                     │
│                          ▼                                      │
│  ┌─────────────────────────────────────────────────────────────┐│
│  │              Tortoise Gateway (Rust + Go)                   ││
│  │  MCP Server │ Skill Executor │ Memory Service │ Agent Mesh ││
│  └─────────────────────────────────────────────────────────────┘│
│                          │                                      │
│  ┌─────────────────────────────────────────────────────────────┐│
│  │           Plugin Runtime (Wasm/Rust)                         ││
│  │  Channel │ Tool │ Skill │ Storage Adapters                   ││
│  └─────────────────────────────────────────────────────────────┘│
│                          │                                      │
│  ┌─────────────────────────────────────────────────────────────┐│
│  │           Backend Services (Go)                              ││
│  │  Auth │ API BFF │ Vector DB │ Notification                  ││
│  └─────────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────────┘
```

---

## 4. 核心模块规格

### 4.1 Rust Core (tortoise-core/)

| 模块 | 职责 |
|------|------|
| `runtime/` | Agent 运行时核心、状态机、事件循环 |
| `mcp/` | MCP 协议实现、Transport 抽象 |
| `sandbox/` | Wasm 沙箱、权限管理 |
| `protocol/` | Agent 间通信协议 |
| `ffi/` | 跨语言 FFI 绑定 (C ABI) |

### 4.2 Go Backend (tortoise-cloud/)

| 模块 | 职责 |
|------|------|
| `cmd/server/` | BFF 服务入口 |
| `internal/auth/` | SuperTokens 集成 |
| `internal/api/` | REST API |
| `internal/vector/` | 向量数据库代理 |
| `internal/audit/` | 审计日志 |

### 4.3 Flutter (tortoise/)

| 模块 | 职责 |
|------|------|
| `lib/core/` | Rust 核心绑定 |
| `lib/channels/` | 消息渠道 UI |
| `lib/agents/` | Agent 管理 |
| `lib/skills/` | Skill 配置执行 |
| `lib/memory/` | 记忆查看器 |

---

## 5. Agent Runtime

### 5.1 状态机

```
Init → Created → Running → Paused → Stopped
                  ↓
               Error
```

### 5.2 生命周期接口

```rust
pub trait Agent: Send + Sync {
    fn init(&self) -> Result<()>;
    fn start(&self, ctx: AgentContext) -> Result<()>;
    fn pause(&self) -> Result<()>;
    fn resume(&self) -> Result<()>;
    fn stop(&self) -> Result<()>;
}
```

---

## 6. Multi-Agent Mesh

### 6.1 通信协议

```rust
pub struct AgentMessage {
    pub id: Uuid,
    pub from: AgentId,
    pub to: AgentId,
    pub message_type: MessageType,
    pub payload: Value,
}

pub enum MessageType {
    Request { method: String, params: Value },
    Response { result: Value },
    Event { event: String, data: Value },
    Delegate { task: Task },
    Collaborate { task: Task, participants: Vec<AgentId> },
}
```

---

## 7. Plugin System

### 7.1 插件类型

| 类型 | 描述 | 示例 |
|------|------|------|
| `channel` | 消息渠道 | Discord, Telegram |
| `tool` | 工具集成 | GitHub, Notion |
| `skill` | 技能包 | Code Review |
| `memory` | 存储适配 | SQLite, Postgres |

### 7.2 沙箱隔离

```rust
pub struct SandboxConfig {
    pub memory_limit: Bytes,
    pub cpu_limit: CpuQuota,
    pub network_enabled: bool,
    pub filesystem_scope: PathBuf,
}
```

---

## 8. Memory System

### 8.1 分层架构

```
Episodic (短期) → Semantic (向量) → Procedural (程序)
     ↓                ↓                 ↓
   24-48h TTL      向量检索          版本化存储
```

---

## 9. 安全模型

### 9.1 权限级别

```rust
pub enum PermissionLevel {
    None,
    Read,
    Write,
    Execute,
    Admin,
}
```

### 9.2 审计日志

```rust
pub struct AuditEvent {
    pub timestamp: DateTime<Utc>,
    pub actor: Actor,
    pub action: Action,
    pub resource: Resource,
    pub result: ActionResult,
}
```

---

## 10. 性能目标

| 指标 | OpenClaw | Hermes | Tortoise 目标 |
|------|----------|--------|---------------|
| 冷启动 | ~10s | ~2s | <500ms |
| 内存占用 | ~200MB | ~80MB | <50MB |
| 并发 Agent | 5 | 2 | 50+ |
| MCP 延迟 | ~50ms | ~20ms | <10ms |

---

## 11. 实现计划

| 阶段 | 目标 | 主要交付 |
|------|------|----------|
| Phase 1 | 核心基础设施 | Rust 核心 + Go BFF + CLI |
| Phase 2 | 插件系统 | Wasm 沙箱 + 插件管理 |
| Phase 3 | Agent Mesh | 通信协议 + 协作机制 |
| Phase 4 | 记忆系统 | 分层存储 + 向量检索 |
| Phase 5 | 企业特性 | RBAC + 审计 + SSO |

---

*最后更新: 2026-05-14*

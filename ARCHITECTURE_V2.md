# Tortoise Agent Platform — V2 架构设计

> 定位：100% 自研，超越 Hermes 的下一代智能代理平台
> 目标：多 Agent 协作 · 自动工作流搭建 · 自我升级

---

## 一、技术栈选型

| 层级 | 语言/框架 | 理由 |
|---|---|---|
| **核心运行时** | Rust (Tokio) | 零拷贝、内存安全、毫秒启动 |
| **插件系统** | TypeScript (napi-rs) | 复用现有 70+ 插件生态 |
| **Agent 编排** | Rust (async) | 自研多 Agent 调度引擎 |
| **通信协议** | MCP 2.0 + 自定义二进制 | 标准化 + 高性能 |
| **前端** | Tauri 2.0 + React | 原生性能 + TS 生态 |
| **持久化** | SQLite (rusqlite) | 单文件、零依赖、跨平台 |
| **配置** | TOML + JSON 双格式 | 人类可读 + 机器兼容 |

---

## 二、分层架构

```
┌─────────────────────────────────────────────────────┐
│                   前端层 (Tauri + React)              │
│  tortoise/src/ui/                                   │
│  聊天界面 · 工作流编辑器 · 插件市场 · 设置            │
└───────────────────────┬─────────────────────────────┘
                        │ Tauri IPC / 事件总线
┌───────────────────────▼─────────────────────────────┐
│              Agent 协调层 (Rust)                      │
│  agent-coordinator/                                 │
│  ├── gateway.rs      网关管理                         │
│  ├── session.rs      会话生命周期                      │
│  ├── config.rs       配置管理                         │
│  ├── self_upgrade.rs 自我升级检测与执行                │
│  └── health.rs       健康检查                         │
└───────────────────────┬─────────────────────────────┘
                        │ Rust 直接调用
┌───────────────────────▼─────────────────────────────┐
│          核心引擎 (Tortoise Core — Rust)               │
│  tortoise-core/                                     │
│  ├── engine.rs       技能/工具执行引擎                │
│  ├── tools.rs        工具注册与调用                   │
│  ├── context.rs      运行时上下文                     │
│  ├── memory.rs       会话记忆 (SQLite + Embedding)    │
│  ├── workspace.rs    工作空间操作                     │
│  ├── scheduler.rs    并发调度器                       │
│  └── embedding.rs    向量嵌入 (语义检索)              │
└───────────────────────┬─────────────────────────────┘
                        │ napi-rs 绑定
┌───────────────────────▼─────────────────────────────┐
│           插件系统 (TS/JS 运行时)                      │
│  plugin-runtime/                                    │
│  ├── registry.rs     插件注册表                       │
│  ├── sandbox.rs      安全沙箱                         │
│  ├── lifecycle.rs    插件生命周期                     │
│  └── hot-reload.rs   热重载                           │
└───────────────────────┬─────────────────────────────┘
                        │ 插件 API
┌───────────────────────▼─────────────────────────────┐
│           插件市场 (可热插拔)                         │
│  plugins/                                           │
│  ├── git-workflow/   TS 插件                         │
│  ├── github/         TS 插件                         │
│  ├── notion/         TS 插件                         │
│  ├── office/         TS 插件                         │
│  └── custom/         用户自定义插件                   │
└───────────────────────┬─────────────────────────────┘
                        │ MCP 2.0 协议
┌───────────────────────▼─────────────────────────────┐
│         MCP 适配器层 (Rust)                           │
│  mcp-adapter/                                       │
│  ├── server.rs       MCP 服务器                       │
│  ├── client.rs       MCP 客户端                       │
│  ├── streaming.rs    流式传输                         │
│  └── compat.rs       向后兼容层                       │
└─────────────────────────────────────────────────────┘

新增差异化能力层（独立 Rust 模块）：

┌─────────────────────────────────────────────────────┐
│        多 Agent 协作引擎 (Rust)                       │
│  agent-orchestrator/                                │
│  ├── planner.rs      任务分解与规划                   │
│  ├── coordinator.rs  Agent 间协调与通信               │
│  ├── resolver.rs     冲突解决与结果合并               │
│  └── memory.rs       Agent 间共享记忆                 │
└─────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────┐
│        工作流引擎 (Rust + TS)                         │
│  workflow-engine/                                   │
│  ├── parser.rs       工作流 DSL 解析                   │
│  ├── runner.rs       执行引擎                         │
│  ├── trigger.rs      触发器管理 (cron/事件/webhook)   │
│  ├── ui/             可视化工作流编辑器 (TS)           │
│  └── registry.rs     节点注册表                       │
└─────────────────────────────────────────────────────┘
```

### 3.5 多 Agent 协作引擎（V2 新增）

这是区别于 Hermes 的核心能力：Tortoise 不再是一个单一 Agent，而是一组可以自主协作的 Agent 网络。

```rust
// agent-orchestrator/src/planner.rs
pub struct TaskPlanner {
    // 任务分解：将复杂任务拆分为子任务 DAG
    decomposer: TaskDecomposer,
    // 任务分配：将子任务分配给合适的 Agent
    assigner: TaskAssigner,
    // 任务执行调度
    scheduler: AgentScheduler,
}

pub struct AgentPool {
    // Agent 实例注册
    agents: HashMap<AgentId, AgentInstance>,
    // Agent 能力描述（用于匹配任务）
    capabilities: HashMap<AgentId, AgentCapabilities>,
}

pub struct AgentInstance {
    pub id: AgentId,
    pub name: String,
    pub role: AgentRole,       // 如 "researcher", "coder", "reviewer"
    pub model: ModelConfig,     // 使用的 LLM 配置
    pub tools: Vec<ToolRef>,    // 可用的工具集合
    pub memory: AgentMemory,    // Agent 私有记忆
}

pub enum AgentRole {
    Researcher,   // 研究：搜索、分析、信息收集
    Coder,        // 编码：编写、重构、调试代码
    Reviewer,     // 审查：代码审查、质量把关
    Orchestrator, // 协调：任务拆分、进度跟踪、最终决策
}
```

**协作流程：**

1. **任务分解** — Orchestrator Agent 接收用户请求，将复杂任务拆分为子任务 DAG
2. **能力匹配** — 根据子任务类型，分发给对应角色的 Agent（研究、编码、审查等）
3. **并行执行** — 无依赖的子任务并行执行，有依赖的按 DAG 顺序执行
4. **结果合并** — Resolver 合并各 Agent 的结果，解决冲突
5. **自我迭代** — 审查 Agent 发现质量问题时，自动触发返工循环

```
用户请求 → Orchestrator (拆分任务) → [Agent A: 研究] [Agent B: 编码]
                                         ↓ (并行)
                                   Resolver (合并结果) → Reviewer (审查)
                                         ↓
                                   反馈循环 → 返工或交付
```

**Agent 间通信协议：**

```typescript
// agent-orchestrator/src/coordinator.ts
export interface AgentMessage {
  type: 'task' | 'result' | 'question' | 'feedback';
  from: AgentId;
  to: AgentId | 'broadcast';
  payload: Record<string, unknown>;
  timestamp: number;
}

export interface AgentMemory {
  // 私有记忆：该 Agent 独有的上下文
  private: ConversationHistory;
  // 共享记忆：与其他 Agent 协作的信息
  shared: SharedMemory;
}
```

### 3.6 工作流引擎（V2 新增）

用户可自动搭建工作流，让 Agent 按预定逻辑自动执行。

```rust
// workflow-engine/src/parser.rs
// 工作流 DSL 定义
pub enum WorkflowNode {
    Trigger(TriggerConfig),    // 触发器：cron / webhook / 事件
    Action(ActionConfig),       // 动作：调用工具/插件/Skill
    Condition(ConditionConfig), // 条件：分支逻辑
    Wait(WaitConfig),           // 等待：延迟 / 外部输入
    SubWorkflow(SubWorkflow),   // 子工作流
}

pub struct Workflow {
    pub id: WorkflowId,
    pub name: String,
    pub nodes: Vec<WorkflowNode>,
    pub triggers: Vec<Trigger>,
    pub state: WorkflowState,
}
```

**工作流节点类型：**

| 节点类型 | 说明 | 示例 |
|---|---|---|
| Trigger | 触发执行 | Cron 定时、文件变更、Webhook 回调 |
| Action | 执行操作 | 调用 Git 插件、发送消息、读取文件 |
| Condition | 条件分支 | 如果 PR 审查通过则合并，否则驳回 |
| Wait | 等待输入 | 等用户确认、等外部系统响应 |
| SubWorkflow | 嵌套工作流 | 复用已有工作流定义 |

**可视化编辑器：** 前端提供拖拽式工作流画布，自动生成工作流 DSL 配置。

### 3.7 自我升级机制（V2 新增）

Tortoise 能自主检测、评估并执行升级，无需人工干预。

```rust
// agent-coordinator/src/self_upgrade.rs
pub struct SelfUpgradeManager {
    // 升级策略配置
    strategy: UpgradeStrategy,
    // 升级状态跟踪
    state: UpgradeState,
    // 升级回滚机制
    rollback: RollbackManager,
}

pub enum UpgradeStrategy {
    // 保守：仅小版本 patch 升级
    Conservative,
    // 平衡：接受 patch 和 minor 版本
    Balanced,
    // 激进：接受所有版本（包括 beta）
    Aggressive,
}

pub struct UpgradeTask {
    pub target_version: String,
    pub changelog: String,
    pub risk_level: RiskLevel,    // low / medium / high
    pub auto_approve: bool,       // 低风险自动批准
    pub backup_before: bool,      // 升级前自动备份
}
```

**升级流程：**

1. 定期检查（cron）检查新版可用性
2. 获取新版本信息（语义版本 + changelog）
3. 评估风险等级（minor 为 low，beta 为 medium/高）
4. 低风险自动批准，高风险通知用户
5. 升级前自动备份配置和会话数据
6. 执行升级并验证
7. 升级失败自动回滚

---

## 四、文件结构规划

```
tortoise/
├── Cargo.toml                  # Rust 工作区
├── package.json                # TS 插件元数据
├── rust-toolchain.toml         # Rust 版本锁定
├── tortoise.toml               # 主配置文件
│
├── tortoise-core/              # 核心运行时 (Rust)
│   ├── Cargo.toml
│   ├── src/
│   │   ├── engine.rs           # 技能执行引擎
│   │   ├── tools.rs            # 工具注册/调用
│   │   ├── context.rs          # 运行时上下文
│   │   ├── memory.rs           # 会话记忆 (SQLite + Embedding)
│   │   ├── workspace.rs        # 工作空间操作
│   │   ├── scheduler.rs        # 并发调度器
│   │   ├── embedding.rs        # 向量嵌入引擎
│   │   └── lib.rs              # 库入口
│   └── tests/
│
├── agent-coordinator/          # Agent 协调层 (Rust)
│   ├── Cargo.toml
│   └── src/
│       ├── gateway.rs
│       ├── session.rs
│       ├── config.rs
│       ├── self_upgrade.rs     # 自我升级
│       └── health.rs
│
├── agent-orchestrator/         # 多 Agent 协作引擎 (Rust)
│   ├── Cargo.toml
│   └── src/
│       ├── planner.rs          # 任务规划
│       ├── coordinator.rs      # Agent 协调
│       ├── resolver.rs         # 结果合并
│       ├── pool.rs             # Agent 池管理
│       └── memory.rs           # Agent 共享记忆
│
├── workflow-engine/            # 工作流引擎 (Rust + TS)
│   ├── Cargo.toml
│   ├── src/
│   │   ├── parser.rs           # DSL 解析
│   │   ├── runner.rs           # 执行引擎
│   │   ├── trigger.rs          # 触发器
│   │   └── registry.rs         # 节点注册表
│   └── ui/                     # 可视化编辑器 (TS)
│
├── plugin-runtime/             # 插件运行时 (Rust)
│   ├── Cargo.toml
│   └── src/
│       ├── registry.rs
│       ├── sandbox.rs
│       ├── lifecycle.rs
│       └── hot-reload.rs
│
├── mcp-adapter/                # MCP 适配器 (Rust)
│   ├── Cargo.toml
│   └── src/
│       ├── server.rs
│       ├── client.rs
│       ├── streaming.rs
│       └── compat.rs
│
├── plugins/                    # TS 插件市场
│   ├── git-workflow/
│   ├── github/
│   ├── notion/
│   ├── office/
│   └── custom/
│
├── tortoise/                   # Tauri 前端（桌面端 UI 壳）
│   ├── package.json
│   ├── src-tauri/              # Tauri 后端 (Rust)
│   │   ├── src/
│   │   │   ├── tauri.rs        # Tauri IPC 桥接
│   │   │   └── lib.rs          # Tauri 入口
│   │   └── tauri.conf.json
│   └── src/                    # React 前端
│       ├── ui/                 # 组件库
│       ├── pages/              # 页面
│       ├── stores/             # Zustand 状态管理
│       └── hooks/              # 自定义 hooks
│
├── scripts/                    # 构建脚本
│   ├── build-all.sh
│   ├── test-all.sh
│   └── release.sh
│
└── docs/
    ├── getting-started.md
    ├── plugin-dev.md
    ├── architecture.md
    ├── migration.md
    └── agent-collab.md         # 多 Agent 协作指南（新增）
```

---

## 五、Rust 核心模块设计

### 5.1 技能执行引擎 (engine.rs)

```rust
use tokio::sync::mpsc;
use std::collections::HashMap;

pub struct SkillEngine {
    skills: HashMap<String, Skill>,
    scheduler: Scheduler,
    executors: ExecutorPool,
}

pub struct Skill {
    pub id: String,
    pub name: String,
    pub handler: SkillHandler,
    pub timeout: Duration,
    pub priority: u32,
}

pub type SkillHandler = Box<dyn Fn(SkillContext) -> Future<Output = Result<SkillResult>> + Send>;

impl SkillEngine {
    pub fn register(&mut self, skill: Skill) {
        self.skills.insert(skill.id.clone(), skill);
    }

    pub async fn execute(&self, skill_id: &str, ctx: SkillContext) -> Result<SkillResult> {
        let skill = self.skills.get(skill_id)
            .ok_or_else(|| Error::SkillNotFound(skill_id.into()))?;
        self.scheduler.schedule(|| {
            (skill.handler)(ctx)
        }).await
    }

    pub fn list_skills(&self) -> Vec<&Skill> {
        self.skills.values().collect()
    }
}
```

### 5.2 工具注册与调用 (tools.rs)

```rust
use serde::{Deserialize, Serialize};
use schemars::JsonSchema;

pub struct ToolRegistry {
    tools: HashMap<String, Tool>,
    schema_cache: HashMap<String, schemars::schema::RootSchema>,
}

pub struct Tool {
    pub name: String,
    pub description: String,
    pub input_schema: schemars::schema::RootSchema,
    pub handler: ToolHandler,
}

pub type ToolHandler = Box<dyn Fn(Value) -> Future<Output = Result<Value>> + Send + Sync>;

impl ToolRegistry {
    pub fn register<T: JsonSchema + for<'de> Deserialize<'de> + Send + 'static>(
        &mut self,
        name: &str,
        description: &str,
        handler: impl Fn(T) -> Future<Output = Result<Value>> + Send + 'static,
    ) {
        let schema = schemars::schema_for!(T);
        self.tools.insert(name.into(), Tool { /* ... */ });
        self.schema_cache.insert(name.into(), schema);
    }

    pub async fn call(&self, name: &str, args: Value) -> Result<Value> {
        let tool = self.tools.get(name)
            .ok_or_else(|| Error::ToolNotFound(name.into()))?;
        (tool.handler)(args).await
    }
}
```

### 5.3 运行时上下文 (context.rs)

```rust
use std::sync::Arc;
use parking_lot::RwLock;

pub struct RuntimeContext {
    workspace_root: Arc<str>,
    session_id: Option<String>,
    config: Arc<RwLock<Config>>,
    logger: Arc<dyn Logger>,
    memory: Arc<MemoryStore>,
}

impl RuntimeContext {
    pub fn new(workspace_root: &str, session_id: Option<String>) -> Self {
        Self {
            workspace_root: workspace_root.into(),
            session_id,
            config: Arc::new(RwLock::new(Config::default())),
            logger: Arc::new(ConsoleLogger),
            memory: Arc::new(MemoryStore::default()),
        }
    }

    pub fn resolve_path(&self, path: &str) -> Result<String> {
        let full_path = Path::new(&self.workspace_root).join(path);
        let canonical = full_path.canonicalize()?;
        if !canonical.starts_with(&self.workspace_root) {
            return Err(Error::PathEscape(path.into()));
        }
        Ok(canonical.to_string_lossy().into())
    }
}
```

### 5.4 会话记忆 (memory.rs) — 带语义检索

```rust
use rusqlite::{Connection, params};

pub struct MemoryStore {
    conn: Arc<Connection>,
    cache: Arc<Mutex<HashMap<String, Vec<MemoryEntry>>>>,
    // V2 新增：向量嵌入存储
    embedding_store: Arc<EmbeddingStore>,
}

pub struct MemoryEntry {
    pub id: String,
    pub session_id: String,
    pub role: String,
    pub content: String,
    pub timestamp: chrono::DateTime<chrono::Utc>,
    pub embedding: Option<Vec<f32>>,
}

impl MemoryStore {
    pub fn new(path: &str) -> Result<Self> {
        let conn = Connection::open(path)?;
        conn.execute_batch(
            "CREATE TABLE IF NOT EXISTS memories (
                id TEXT PRIMARY KEY,
                session_id TEXT NOT NULL,
                role TEXT NOT NULL,
                content TEXT NOT NULL,
                timestamp INTEGER NOT NULL,
                embedding BLOB
            );
            CREATE INDEX IF NOT EXISTS idx_session ON memories(session_id);"
        )?;
        Ok(Self {
            conn: Arc::new(conn),
            cache: Arc::new(Mutex::new(HashMap::new())),
            embedding_store: Arc::new(EmbeddingStore::new()),
        })
    }

    pub async fn search(&self, query: &str, limit: usize) -> Result<Vec<MemoryEntry>> {
        // V2 新增：语义检索
        let query_embedding = self.embedding_store.encode(query).await?;
        let results = self.embedding_store.similarity_search(&query_embedding, limit).await?;
        Ok(results)
    }
}
```

### 5.5 并发调度器 (scheduler.rs)

```rust
use tokio::sync::Semaphore;
use std::time::Duration;

pub struct Scheduler {
    semaphore: Arc<Semaphore>,
    timeout_mgr: TimeoutManager,
}

impl Scheduler {
    pub fn new(max_concurrency: usize) -> Self {
        Self {
            semaphore: Arc::new(Semaphore::new(max_concurrency)),
            timeout_mgr: TimeoutManager { timeouts: HashMap::new() },
        }
    }

    pub async fn schedule<F, Fut, T>(&self, f: F) -> Result<T>
    where
        F: FnOnce() -> Fut + Send + 'static,
        Fut: Future<Output = Result<T>> + Send,
        T: Send + 'static,
    {
        let _permit = self.semaphore.acquire().await?;
        Ok(f().await?)
    }
}
```

---

## 六、napi-rs 绑定层

```rust
// napi/src/lib.rs
use napi::bindgen_prelude::*;
use tortoise_core::engine::SkillEngine;

#[napi]
pub struct AgentRuntime {
    inner: Arc<Mutex<SkillEngine>>,
}

#[napi]
impl AgentRuntime {
    #[napi(factory)]
    pub fn new(workspace_root: String, session_id: Option<String>) -> Result<Self> {
        let ctx = RuntimeContext::new(&workspace_root, session_id);
        let engine = SkillEngine::new(ctx);
        Ok(Self {
            inner: Arc::new(Mutex::new(engine)),
        })
    }

    #[napi]
    pub async fn ping(&self, message: Option<String>) -> Result<JsUnknown> {
        let result = serde_json::json!({
            "status": "ok",
            "message": message.unwrap_or_default(),
            "node_version": env!("CARGO_PKG_VERSION"),
        });
        let env = get_env();
        env.to_js_value(&result).map_err(Into::into)
    }

    #[napi]
    pub async fn list_skills(&self) -> Result<JsUnknown> {
        let engine = self.inner.lock().await;
        let skills: Vec<_> = engine.list_skills();
        let result = serde_json::json!(skills);
        let env = get_env();
        env.to_js_value(&result).map_err(Into::into)
    }

    #[napi]
    pub async fn invoke_skill(
        &self,
        skill: String,
        tool: String,
        args: Option<JsUnknown>,
        timeout_ms: Option<u32>,
    ) -> Result<JsUnknown> {
        let engine = self.inner.lock().await;
        let result = engine.execute(&skill, SkillContext::default()).await;
        let result_json = serde_json::json!(result);
        let env = get_env();
        env.to_js_value(&result_json).map_err(Into::into)
    }
}
```

---

## 七、多 Agent 协作引擎（详细设计）

### 7.1 任务分解算法

```rust
// agent-orchestrator/src/planner.rs
pub struct TaskDecomposer {
    // 基于 LLM 的任务分解 prompt
    system_prompt: String,
}

impl TaskDecomposer {
    pub async fn decompose(&self, task: &str) -> Result<TaskDAG> {
        // 1. 调用 LLM 将自然语言任务拆分为结构化子任务
        // 2. 构建依赖关系图（DAG）
        // 3. 标注每个子任务的角色需求（researcher / coder / reviewer）
        // 4. 返回可执行的 DAG
        let response = self.call_llm(task).await?;
        parse_dag(response)
    }
}

pub struct TaskDAG {
    pub nodes: Vec<TaskNode>,
    pub edges: Vec<Edge>,
}

pub struct TaskNode {
    pub id: TaskId,
    pub description: String,
    pub role: AgentRole,
    pub required_tools: Vec<String>,
    pub dependencies: Vec<TaskId>,
}
```

### 7.2 结果合并与冲突解决

```rust
// agent-orchestrator/src/resolver.rs
pub struct ResultResolver {
    strategy: MergeStrategy,
}

pub enum MergeStrategy {
    // 直接合并：各 Agent 结果互不冲突
    DirectMerge,
    // 投票合并：多个 Agent 给出不同结果时取多数
    Voting,
    // 优先级合并：按 Agent 角色优先级决定
    PriorityBased,
    // 人工裁决：冲突时请求用户输入
    HumanInLoop,
}

impl ResultResolver {
    pub async fn resolve(
        &self,
        results: Vec<AgentResult>,
    ) -> Result<MergedResult> {
        match &self.strategy {
            MergeStrategy::DirectMerge => self.direct_merge(results),
            MergeStrategy::Voting => self.voting_merge(results),
            MergeStrategy::PriorityBased => self.priority_merge(results),
            MergeStrategy::HumanInLoop => self.human_in_loop(results),
        }
    }
}
```

### 7.3 Agent 间通信

```rust
// agent-orchestrator/src/coordinator.rs
pub struct AgentCoordinator {
    // Agent 消息总线
    bus: Arc<MessageBus<AgentMessage>>,
    // Agent 池
    pool: Arc<AgentPool>,
}

impl AgentCoordinator {
    // 发送消息给指定 Agent
    pub async fn send(&self, msg: AgentMessage) -> Result<()> {
        self.bus.send(msg).await
    }

    // 广播消息给所有 Agent
    pub async fn broadcast(&self, msg: AgentMessage) -> Result<()> {
        self.bus.broadcast(msg).await
    }

    // 获取指定角色的所有 Agent
    pub fn get_by_role(&self, role: AgentRole) -> Vec<AgentInstance> {
        self.pool.agents_by_role(role)
    }
}
```

---

## 八、工作流引擎（详细设计）

### 8.1 工作流 DSL

```yaml
# workflow-examples/deploy.yml
name: "自动部署"
triggers:
  - type: "event"
    on:
      event: "pull_request.closed"
      conditions:
        merged: true
        base_branch: "main"

nodes:
  - id: "lint"
    type: "action"
    action: "npm run lint"
    workspace: "."

  - id: "test"
    type: "action"
    action: "npm test"
    depends_on: ["lint"]

  - id: "build"
    type: "action"
    action: "npm run build"
    depends_on: ["test"]

  - id: "deploy"
    type: "action"
    action: "./scripts/deploy.sh"
    depends_on: ["build"]
    permissions: ["execute"]

  - id: "notify"
    type: "action"
    action: "notify_slack"
    depends_on: ["deploy"]
    input:
      channel: "#deployments"
      message: "✅ {{workflow.name}} 部署成功"
```

### 8.2 执行引擎

```rust
// workflow-engine/src/runner.rs
pub struct WorkflowRunner {
    registry: Arc<Registry>,
    state_store: Arc<WorkflowStateStore>,
}

impl WorkflowRunner {
    pub async fn run(&self, workflow: &Workflow) -> Result<WorkflowResult> {
        // 1. 验证工作流合法性
        self.validate(workflow)?;
        // 2. 初始化执行状态
        let state = self.initialize_state(workflow);
        // 3. 按拓扑顺序执行节点
        for node in self.topological_sort(workflow) {
            self.execute_node(&state, node).await?;
        }
        // 4. 返回执行结果
        Ok(state.into_result())
    }
}
```

---

## 九、自我升级机制（详细设计）

### 9.1 升级检测

```rust
// agent-coordinator/src/self_upgrade.rs
pub struct UpgradeChecker {
    current_version: Version,
    update_channel: UpdateChannel,
    registry: UpdateRegistry,
}

impl UpgradeChecker {
    pub async fn check(&self) -> Result<Option<UpgradeTask>> {
        // 1. 查询 npm registry 获取最新版本
        let latest = self.registry.query_latest().await?;
        // 2. 比较版本，检查是否符合当前升级通道
        if !self.update_channel.accepts(&latest) {
            return Ok(None);
        }
        // 3. 获取 changelog 评估风险
        let changelog = self.registry.get_changelog(&latest).await?;
        let risk = self.evaluate_risk(&self.current_version, &latest);
        // 4. 根据风险等级决定是否自动批准
        Ok(Some(UpgradeTask {
            target_version: latest,
            changelog,
            risk_level: risk,
            auto_approve: risk.is_low(),
            backup_before: true,
        }))
    }
}
```

### 9.2 升级执行与回滚

```rust
pub struct RollbackManager {
    // 备份目录
    backup_dir: PathBuf,
    // 版本快照
    snapshots: HashMap<Version, PathBuf>,
}

impl RollbackManager {
    pub async fn rollback(&self, target_version: &Version) -> Result<()> {
        let backup = self.snapshots.get(target_version)
            .ok_or_else(|| Error::SnapshotNotFound(target_version.clone()))?;
        // 1. 停掉当前服务
        // 2. 恢复备份
        // 3. 重启服务
        // 4. 验证健康状态
        Ok(())
    }
}
```

---

## 十、迁移路径

### 阶段 1: 核心 Rust 化（1-2 周）

- [ ] 重写 `engine.rs` — 技能执行引擎
- [ ] 重写 `tools.rs` — 工具注册/调用
- [ ] 重写 `context.rs` — 运行时上下文
- [ ] 重写 `memory.rs` — 会话记忆
- [ ] 重写 `workspace.rs` — 工作空间操作
- [ ] 添加 napi-rs 绑定

### 阶段 2: 插件系统升级（2-3 周）

- [ ] 迁移 70+ 插件到插件 API 2.0
- [ ] 实现热重载机制
- [ ] 兼容性层开发

### 阶段 3: 新增能力模块（3-4 周）

- [ ] 多 Agent 协作引擎（planner / coordinator / resolver）
- [ ] 工作流引擎（parser / runner / trigger）
- [ ] 自我升级机制（checker / executor / rollback）
- [ ] 向量嵌入引擎（embedding）

### 阶段 4: MCP 适配器重写（1 周）

- [ ] Rust 实现 MCP 2.0 协议
- [ ] 流式传输优化
- [ ] 向后兼容层

### 阶段 5: 前端升级（1-2 周）

- [ ] Tauri 2.0 迁移
- [ ] React 组件重构
- [ ] 工作流可视化编辑器
- [ ] 性能优化

### 阶段 6: 测试与发布（1 周）

- [ ] 全量测试
- [ ] 性能基准对比
- [ ] 文档更新
- [ ] 正式发布

---

## 十一、性能目标

| 指标 | 当前 | 目标 |
|---|---|---|
| 启动时间 | 2-5 秒 | < 100ms |
| 内存占用 | 100-200MB | < 30MB |
| 技能调用延迟 | 50-200ms | < 10ms |
| 插件加载 | 同步阻塞 | 异步热重载 |
| 并发技能数 | 3-5 | 50+ |
| 内存安全 | GC 管理 | 编译期保证 |
| Agent 协作数 | 1 (单 Agent) | 10+ 并行 |
| 工作流节点 | 无 | 支持 100+ 节点/流程 |
| 语义检索 | 无 | < 50ms (千级条目) |

---

## 十二、安全性设计

### 12.1 沙箱隔离

```rust
pub struct PluginSandbox {
    fs: SandboxFS,     // 文件系统访问限制
    net: SandboxNet,   // 网络访问控制
    env: SandboxEnv,   // 环境变量隔离
    process: SandboxProcess, // 进程执行限制
}
```

### 12.2 权限模型

```toml
[plugins.github]
permissions = ["read", "execute"]
allowed_network = ["api.github.com"]
allowed_fs = ["workspace"]

[plugins.notion]
permissions = ["read", "write"]
allowed_network = ["api.notion.com"]
allowed_fs = ["workspace", "cache"]
```

### 12.3 Agent 权限隔离

```rust
// 每个 Agent 只能使用分配的工具和权限
pub struct AgentPermissions {
    pub allowed_tools: Vec<String>,
    pub allowed_network: Vec<String>,
    pub allowed_fs: Vec<FileSystemScope>,
    pub max_concurrency: usize,
    pub budget: Option<BudgetLimit>, // 如 token / 执行次数限制
}
```

---

## 十三、总结

本架构实现以下目标：

1. **100% 自研** — 核心引擎完全自研，不依赖 Hermes 或 OpenClaw
2. **多 Agent 协作** — 支持任务自动拆分、角色分配、并行执行、结果合并
3. **自动工作流** — 可视化编排工作流，支持 cron/事件/webhook 触发
4. **自我升级** — 自主检测、评估、执行升级，失败自动回滚
5. **绝对兼容** — 保留所有现有 API 签名和 70+ 插件生态
6. **性能极致** — Rust 核心 + 零拷贝，启动速度提升 50x
7. **易于拓展** — 插件热插拔、权限模型、兼容矩阵

下一步：开始实现核心 Rust 模块？

## 三、关键设计决策

### 3.1 混合架构：Rust 核心 + TS 插件

```
Rust (核心) ←─ napi-rs 绑定 ─→ TypeScript (插件)
```

- **Rust 核心**：性能敏感、内存安全、并发调度 — 重写 Hermes 核心，约 20KB 代码
- **TS 插件**：70+ 现有插件无需重写，通过 napi-rs 调用 Rust 核心能力
- **热重载**：插件可独立升级，无需重启核心
- **差异化**：核心新增 Agent 编排、工作流引擎、自我升级模块

### 3.2 绝对兼容性策略

| 现有 API | 兼容方式 |
|---|---|
| `tohelp_ping` | 保留，Rust 实现相同响应 |
| `tohelp_invoke_skill` | 保留，调用 TS 插件 + Rust 引擎 |
| `tohelp_memory` | 保留，Rust SQLite 后端 |
| `SKILL_IDS` 枚举 | 保留，JSON 配置驱动 |
| MCP 工具名称 | 保留，自动注册 |

### 3.3 插件热插拔架构

```typescript
// plugins/types.ts
export interface Plugin {
  name: string;
  version: string;
  apiVersion: string; // 兼容矩阵
  tools: ToolDefinition[];
  onLoad?(ctx: PluginContext): Promise<void>;
  onUnload?(ctx: PluginContext): Promise<void>;
}

export interface PluginContext {
  core: CoreAPI;      // 调用 Tortoise 核心能力
  memory: MemoryAPI;  // 会话记忆
  workspace: WorkspaceAPI; // 工作空间操作
  config: ConfigAPI;  // 配置管理
  agent?: AgentAPI;   // 多 Agent 协作（V2 新增）
  workflow?: WorkflowAPI; // 工作流触发（V2 新增）
}
```

**热重载流程：**

1. 检测到新版本插件
2. 暂停旧插件调用（完成 pending 请求）
3. 卸载旧插件（调用 `onUnload`）
4. 加载新插件（调用 `onLoad`）
5. 更新注册表

### 3.4 版本管理

```toml
# tortoise.toml - 主配置文件
[agent]
version = "2.0.0"
core_version = "rust:1.75.0"
plugin_api_version = "2.0"

[agent.updates]
auto_update = true
channel = "stable"  # stable | beta | nightly
interval_hours = 24

[agent.plugins]
auto_install = true
hot_reload = true
max_plugins = 100
```

</anth Thinking>

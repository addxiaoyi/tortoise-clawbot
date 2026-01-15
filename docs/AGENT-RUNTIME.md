# Agent Runtime 架构设计

> 构建同时支持 **Hermes 核心能力** 和 **OpenClaw 生态** 的全功能 Agent Runtime

## 1. 设计目标

| 目标 | 说明 |
|------|------|
| **Hermes Core** | Skill Runner、Plugin Lifecycle、Memory KV、Gateway |
| **OpenClaw Compatible** | Channels（Telegram/Discord/Slack/WhatsApp）、Providers、Tools |
| **MCP Native** | MCP stdio 服务器、工具注册 |
| **Extensible** | 插件化架构，支持第三方扩展 |

## 2. 整体架构

```
┌─────────────────────────────────────────────────────────────────────────┐
│                              Client Layer                               │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  │
│  │  MCP Client  │  │   HTTP SDK  │  │   WebSocket  │  │    CLI      │  │
│  └──────────────┘  └──────────────┘  └──────────────┘  └──────────────┘  │
└───────────────────────────────┬─────────────────────────────────────────┘
                                │
┌───────────────────────────────▼─────────────────────────────────────────┐
│                           Gateway Layer                                  │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │                     GatewayServer                                 │   │
│  │  - HTTP/WebSocket Server                                         │   │
│  │  - Router: /health, /invoke, /tools, /memory, /session          │   │
│  │  - Auth: Token/Session middleware                                 │   │
│  │  - Rate Limiting                                                 │   │
│  │  - CORS support                                                  │   │
│  └─────────────────────────────────────────────────────────────────┘   │
└───────────────────────────────┬─────────────────────────────────────────┘
                                │
┌───────────────────────────────▼─────────────────────────────────────────┐
│                          Runtime Core                                    │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────────────┐ │
│  │   SkillRunner   │  │  PluginSystem    │  │     MemoryManager        │ │
│  │  - 插件加载      │  │  - Container     │  │  - KV Storage            │ │
│  │  - 工具调度      │  │  - Registry      │  │  - Session Isolation     │ │
│  │  - 超时控制      │  │  - Lifecycle     │  │  - Prefix隔离           │ │
│  │  - 错误处理      │  │  - DI            │  │  - Persistence (可选)    │ │
│  └─────────────────┘  └─────────────────┘  └─────────────────────────┘ │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────────────┐ │
│  │  SessionManager  │  │   ToolRegistry  │  │     EventBus            │ │
│  │  - 会话管理      │  │  - 工具注册      │  │  - 事件驱动              │ │
│  │  - 上下文        │  │  - 权限控制      │  │  - 发布/订阅             │ │
│  └─────────────────┘  └─────────────────┘  └─────────────────────────┘ │
└───────────────────────────────┬─────────────────────────────────────────┘
                                │
┌───────────────────────────────▼─────────────────────────────────────────┐
│                         Plugin Layer                                     │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────────────┐ │
│  │  Skill Plugins  │  │ Channel Plugins │  │   Provider Plugins       │ │
│  │  - github      │  │ - telegram     │  │  - anthropic            │ │
│  │  - slack       │  │ - discord      │  │  - openai               │ │
│  │  - notion      │  │ - slack        │  │  - ollama               │ │
│  │  - canvas      │  │ - whatsapp     │  │  - openrouter           │ │
│  │  - ...        │  │ - signal       │  │  - ...                  │ │
│  └─────────────────┘  └─────────────────┘  └─────────────────────────┘ │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │                     Custom Plugins (Third-party)                │   │
│  └─────────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────────┘
```

## 3. 核心模块

### 3.1 GatewayServer

```typescript
// src/runtime/gateway/server.ts
interface GatewayConfig {
  port: number;
  host: string;
  auth: {
    token?: string;
    sessionSecret?: string;
  };
  cors: {
    enabled: boolean;
    origins?: string[];
  };
  rateLimit: {
    windowMs: number;
    maxRequests: number;
  };
}
```

**路由：**

| Method | Path | 说明 |
|--------|------|------|
| GET | `/health` | 健康检查 |
| GET | `/tools` | 列出所有工具 |
| POST | `/invoke` | 调用工具 |
| GET/POST | `/memory` | 内存操作 |
| WS | `/ws` | WebSocket 实时连接 |

### 3.2 PluginSystem

```typescript
// src/runtime/plugin/container.ts
interface PluginContainer {
  // 注册插件
  register(plugin: PluginLifecycle, config?: PluginConfig): void;
  
  // 启动所有插件
  startAll(): Promise<void>;
  
  // 停止所有插件
  stopAll(): Promise<void>;
  
  // 获取插件
  get<T extends PluginLifecycle>(id: string): T | undefined;
  
  // 获取所有插件
  getAll(): PluginLifecycle[];
}
```

### 3.3 SkillRunner

```typescript
// src/runtime/skill/runner.ts
interface SkillRunner {
  // 调用技能工具
  invoke(
    skill: string,
    tool: string,
    args: Record<string, unknown>,
    options?: InvokeOptions
  ): Promise<unknown>;
  
  // 列出所有技能
  listSkills(): SkillDefinition[];
  
  // 验证技能/工具存在
  validate(skill: string, tool: string): boolean;
}
```

### 3.4 MemoryManager

```typescript
// src/runtime/memory/manager.ts
interface MemoryConfig {
  prefix: string;              // 键前缀隔离
  maxValueBytes: number;       // 值大小限制
  sessionScoped: boolean;       // 会话级隔离
  requireSessionForRead: boolean;
  requireSessionForWrite: boolean;
  persistence?: {
    type: 'file' | 'redis' | 'sqlite';
    path?: string;
  };
}
```

### 3.5 Channel System（OpenClaw 兼容）

```typescript
// src/channels/base.ts
interface ChannelAdapter {
  readonly name: string;
  readonly capabilities: ChannelCapability[];
  
  // 生命周期
  onInit(ctx: PluginContext): Promise<void>;
  onStart(): Promise<void>;
  onStop(): Promise<void>;
  
  // 消息处理
  send(message: OutboundMessage): Promise<void>;
  formatForChannel(content: Content): Promise<string>;
}

// 内置 Channel 实现
class TelegramChannel implements ChannelAdapter { ... }
class DiscordChannel implements ChannelAdapter { ... }
class SlackChannel implements ChannelAdapter { ... }
class WhatsAppChannel implements ChannelAdapter { ... }
class SignalChannel implements ChannelAdapter { ... }
class IMessageChannel implements ChannelAdapter { ... }
```

### 3.6 Provider System（OpenClaw 兼容）

```typescript
// src/providers/base.ts
interface ModelProvider {
  readonly name: string;
  
  // 模型调用
  complete(options: CompletionOptions): Promise<CompletionResult>;
  
  // 流式调用
  completeStream(options: CompletionOptions): AsyncGenerator<StreamChunk>;
  
  // Embeddings
  embed(options: EmbedOptions): Promise<EmbedResult>;
}

class AnthropicProvider implements ModelProvider { ... }
class OpenAIProvider implements ModelProvider { ... }
class OllamaProvider implements ModelProvider { ... }
class OpenRouterProvider implements ModelProvider { ... }
```

## 4. 插件类型层级

```
PluginLifecycle (基础接口)
    │
    ├── SkillPlugin (技能插件)
    │       └── 需实现 getSkillDefinition()
    │
    ├── ChannelPlugin (频道插件)
    │       └── 需实现消息收发
    │
    ├── ProviderPlugin (模型提供商插件)
    │       └── 需实现模型调用
    │
    └── SystemPlugin (系统插件)
            └── Gateway, Memory, Session 等
```

## 5. 配置注入

```typescript
// 多层配置合并
const configSources = [
  process.env,                           // 环境变量
  fileConfig['runtime'],                 // 配置文件
  fileConfig['plugins'][pluginId],       // 插件配置
  runtimeConfig['skills'],               // 运行时注入
];

const finalConfig = deepMerge(...configSources);
```

## 6. 目录结构

```
src/
├── runtime/
│   ├── gateway/         # HTTP/WebSocket 网关
│   │   ├── server.ts
│   │   ├── router.ts
│   │   ├── middleware/
│   │   └── ws-handler.ts
│   ├── plugin/          # 插件系统
│   │   ├── container.ts
│   │   ├── registry.ts
│   │   └── loader.ts
│   ├── skill/           # 技能执行器
│   │   ├── runner.ts
│   │   └── validator.ts
│   ├── memory/          # 内存管理
│   │   ├── manager.ts
│   │   └── persistence/
│   ├── session/         # 会话管理
│   │   └── manager.ts
│   ├── tool/            # 工具注册
│   │   └── registry.ts
│   └── types.ts         # 核心类型
│
├── channels/            # OpenClaw 兼容频道
│   ├── base.ts
│   ├── telegram/
│   ├── discord/
│   ├── slack/
│   ├── whatsapp/
│   ├── signal/
│   └── imessage/
│
├── providers/           # OpenClaw 兼容 Provider
│   ├── base.ts
│   ├── anthropic.ts
│   ├── openai.ts
│   ├── ollama.ts
│   └── openrouter.ts
│
├── skills/              # 技能插件（迁移自 new-core）
│   ├── github/
│   ├── slack/
│   ├── notion/
│   └── ...
│
├── mcp/                 # MCP 适配层
│   ├── server.ts
│   └── register-tools.ts
│
├── cli/                 # CLI 命令
│   ├── gateway.ts
│   ├── agent.ts
│   └── plugin.ts
│
└── index.ts             # 导出
```

## 7. 启动流程

```
1. Load Config (环境变量 + 配置文件)
         │
         ▼
2. Initialize Logger
         │
         ▼
3. Create PluginContainer
         │
         ▼
4. Load System Plugins (Memory, Session, ToolRegistry, EventBus)
         │
         ▼
5. Load Skill Plugins (from config)
         │
         ▼
6. Load Channel Plugins (from config)
         │
         ▼
7. Load Provider Plugins (from config)
         │
         ▼
8. Start PluginContainer (call onStart for all)
         │
         ▼
9. Start GatewayServer
         │
         ▼
10. Ready! 🎉
```

## 8. 相关文档

| 文档 | 说明 |
|------|------|
| `docs/AGENT-RUNTIME.md` | 本文档 — 架构设计 |
| `docs/CHANNELS.md` | OpenClaw 频道实现指南 |
| `docs/PROVIDERS.md` | 模型 Provider 实现指南 |
| `docs/PLUGIN-DEV.md` | 插件开发指南 |
| `docs/MIGRATION.md` | 从现有 Hermes/OpenClaw 迁移 |

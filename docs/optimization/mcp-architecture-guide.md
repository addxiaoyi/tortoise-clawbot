# MCP 与 Tohelp 架构说明

本文描述本仓库中 **OpenClaw 网关**、**OpenClaw 扩展（插件）** 与 **`src/plugins/new-core`** 三者的关系，便于后续把能力接到 MCP 或 Agent 工具层。

## 1. OpenClaw（`openclaw-main`）

负责网关、配置、官方扩展发现、运行时加载与 `openclaw/plugin-sdk` 契约。

## 2. 桥接扩展 `extensions/tohelp-openclaw`

通过配置项 `plugins.load.paths` 指向的目录被加载。启动时注册 **service** 日志，并调用 `registerTohelpTools(api)`，向网关登记可选工具：

- `tohelp_ping`：桥接存活检查  
- `tohelp_list_new_core_skills`：枚举 `src/plugins/new-core` 下各 `SkillPlugin` 的 `getSkillDefinition()`（不发起外网请求）  
- `tohelp_resolve_workspace_path`：封装 `api.resolvePath`  
- `tohelp_invoke_skill`：统一入口，按 `skill` / `tool` / `args` 调用各 `SkillPlugin` 的 `execute`（生命周期由 `src/bridge/skill-invoke.ts` 驱动）  
- `tohelp_memory`：进程内 `MemoryPlugin` 键值（get / set / list / delete / clear），与网关进程同生命周期。**非安全存储**：无多租户隔离，不要写入密钥。可选环境变量 **`TOHELP_MEMORY_KEY_PREFIX`**（逻辑 key 自动加前缀；`list` 仅列出该前缀下键名；`clear` 只删除带此前缀的条目；未设置时 `clear` 仍为整表清空）、**`TOHELP_MEMORY_MAX_VALUE_BYTES`**（每条 `set` 的 value 在 `JSON.stringify` 后 UTF-8 字节上限，抑制 DoS）、**`TOHELP_MEMORY_AUTO_SESSION_PREFIX=1`**（仅 OpenClaw：按 `sessionScopeId` 追加 `sess:<id>:`，其中 `sessionKey` 优先，`sessionId` 次之）、**`TOHELP_MEMORY_REQUIRE_SESSION_KEY_FOR_WRITE=1`**（`set/delete/clear` 必须有 `sessionKey`）、**`TOHELP_MEMORY_REQUIRE_SESSION_KEY_FOR_READ=1`**（`get/list` 也必须有 `sessionKey`）。独立 **`npm run mcp:tohelp`** 通常无会话上下文，开启 read/write 强制开关后会拒绝无 `sessionKey` 的调用。见仓库根 `.env.example`。  
- `tohelp_gateway_health_probe`：在随机本机端口启动 `GatewayPlugin`，请求 `GET /health` 后关闭，用于快速验证网关插件  

### `tohelp_memory` 推荐配置模板

```env
# 本地开发（兼容优先）
TOHELP_MEMORY_KEY_PREFIX=dev:
TOHELP_MEMORY_MAX_VALUE_BYTES=262144
TOHELP_MEMORY_AUTO_SESSION_PREFIX=0
TOHELP_MEMORY_REQUIRE_SESSION_KEY_FOR_WRITE=0
TOHELP_MEMORY_REQUIRE_SESSION_KEY_FOR_READ=0
```

```env
# 生产/多会话（隔离优先）
TOHELP_MEMORY_KEY_PREFIX=prod:
TOHELP_MEMORY_MAX_VALUE_BYTES=1048576
TOHELP_MEMORY_AUTO_SESSION_PREFIX=1
TOHELP_MEMORY_REQUIRE_SESSION_KEY_FOR_WRITE=1
TOHELP_MEMORY_REQUIRE_SESSION_KEY_FOR_READ=1
```

切换建议：
- 若客户端是 OpenClaw 网关且能稳定提供 `sessionKey`，优先使用「隔离优先」模板。
- 若是独立 MCP（无会话上下文），不要直接启用 read/write 强制开关，除非你在调用链自行注入 `sessionKey`。

实现集中在 **`src/bridge/tohelp-tools.ts`**（登记）、**`tohelp-executors.ts`**（与 OpenClaw / MCP 共用的执行逻辑）、**`skill-invoke.ts`**、**`skill-registry.ts`**、**`bridge-context.ts`**、**`bridge-memory.ts`**、**`with-deadline.ts`**；聚合导出 **`src/bridge/index.ts`**。扩展入口仅负责接入 OpenClaw。

### 独立 MCP Server（stdio）

不经过 OpenClaw 网关时，可用 **`npm run mcp:tohelp`** 启动基于 `@modelcontextprotocol/sdk` 的 stdio MCP，工具集与桥接侧对齐（ping、列技能、解析路径、invoke_skill、memory、gateway 探活）。

- 工作区根目录：环境变量 **`TOHELP_WORKSPACE_ROOT`**（未设置则用进程 `cwd`），用于 `tohelp_resolve_workspace_path` / 相对路径解析。  
- 与 `tohelp_invoke_skill` 相关的密钥与 `invokeTimeoutMs`：可将网关插件 config 的 JSON 放在 **`TOHELP_PLUGIN_CONFIG_JSON`**（单行 JSON 字符串），行为类似 `api.pluginConfig`。  

在 Cursor / Claude Desktop 等客户端中，将 command 配为 `npx tsx` 或 `node`，args 指向 `src/mcp/tohelp-mcp-server.ts`（或先 `npm run mcp:tohelp` 的等价命令），transport 选 **stdio** 即可。工具注册逻辑在 **`src/mcp/register-tohelp-tools.ts`**，便于在自定义传输层上复用。

### 插件配置

网关会把 `plugins.entries` 中对应条目的 `config` 注入为 `api.pluginConfig`。桥接层通过 `createBridgePluginContext` 将 `pluginConfig.skills.<skillId>` 合并进 `getConfig()`，便于为 `slack`、`notion` 等写入独立密钥字段。

开发示例见仓库 `.openclaw-dev/openclaw.json` 中的 `plugins.entries.tohelp-openclaw`。扩展清单 `extensions/tohelp-openclaw/openclaw.plugin.json` 的 `configSchema` 允许顶层任意键与 `skills`、`invokeTimeoutMs` 等字段。

### 超时、中止与日志

`tohelp_invoke_skill` 使用 `withDeadline`：默认 120s，可通过 `config.invokeTimeoutMs` 或工具参数 `timeoutMs` 调整；若 OpenClaw 传入 `AbortSignal`，会在中止时提前结束（无法打断已同步执行的第三方库内部逻辑，仅能在异步边界返回）。

网关上会打 **`[tohelp] invoke start|done|error`** 日志：包含 `skill`、`tool`、`timeoutMs`、**参数名列表 `argKeys`**（不含参数值，降低密钥泄露风险）。

## 3. `src/plugins/new-core`

使用自有的 `Container`、`PluginRegistry`、`PluginContext` 与 `PluginLifecycle`，用于单元测试与模块化组织；**不是** OpenClaw 的 `openclaw.plugin.json` 插件。与 OpenClaw 的集成方式应为：在 `tohelp-openclaw` 中编写适配层，而不是假定两者自动互通。

## 已知边界

- **`tohelp_invoke_skill` 仍只调度 `SkillPlugin`**（见 `src/bridge/skill-registry.ts`）。`MemoryPlugin` / `GatewayPlugin` 通过专用工具 **`tohelp_memory`**、**`tohelp_gateway_health_probe`**（及 MCP 同名工具）暴露，不经过 `invoke_skill`。  
- **`github` skill** 依赖本机 **`gh` CLI** 与登录状态；与仅写在 `pluginConfig` 里的键无必然对应，部署环境需预装并认证 `gh`。  
- **超时**只能作用于异步等待边界；CPU 同步长循环无法被可靠打断。

## 配置入口

- 开发配置：`.openclaw-dev/openclaw.json`  
- 通过环境变量 `OPENCLAW_CONFIG_PATH` 指向该文件（见 `start-dev.ps1` / `start-dev.sh`）。  
- 网关 Token：`gateway.auth.token` 使用 `${OPENCLAW_GATEWAY_TOKEN}`，由 `.env` 或进程环境注入。

## 推荐演进路径

1. 保持 `new-core` 核心逻辑在 `src/`，用 Vitest 覆盖。  
2. 在 `extensions/tohelp-openclaw` / `src/bridge` 中按需增加工具或收窄 `tohelp_invoke_skill` 暴露面。  
3. 需要跨进程 MCP 时，使用 **`npm run mcp:tohelp`**（`src/mcp/tohelp-mcp-server.ts`）与网关并行部署；客户端只连 stdio MCP，不占用 OpenClaw HTTP 路由。

## 参考

- OpenClaw 文档：<https://docs.openclaw.ai/>
- 将 `openclaw-main` 改为 git 子模块（可选）：见仓库内 **`docs/setup-openclaw-submodule.md`**

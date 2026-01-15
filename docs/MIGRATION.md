# 迁移指南

> 从 OpenClaw/Hermes 迁移到新 Agent Runtime

## 1. 迁移概述

本指南帮助你将现有的 OpenClaw 或 Hermes 配置迁移到新的 Agent Runtime。

### 迁移路径

```
OpenClaw → Hermes → Agent Runtime (新)
     ↓          ↓
     └──────────┴──→ 完全兼容
```

## 2. OpenClaw 配置迁移

### 2.1 openclaw.json → runtime.json

**OpenClaw 旧配置：**

```json
// openclaw.json
{
  "gateway": {
    "port": 19001,
    "host": "0.0.0.0",
    "auth": {
      "token": "${OPENCLAW_GATEWAY_TOKEN}"
    }
  },
  "plugins": {
    "entries": {
      "tohelp-openclaw": {
        "enabled": true,
        "config": {
          "skills": {
            "github": { "token": "${GITHUB_TOKEN}" },
            "slack": { "botToken": "${SLACK_BOT_TOKEN}" }
          }
        }
      }
    }
  },
  "channels": {
    "discord": { "botToken": "${DISCORD_BOT_TOKEN}" },
    "telegram": { "botToken": "${TELEGRAM_BOT_TOKEN}" }
  },
  "providers": {
    "anthropic": { "apiKey": "${ANTHROPIC_API_KEY}" },
    "openai": { "apiKey": "${OPENAI_API_KEY}" }
  }
}
```

**新 Agent Runtime 配置：**

```json
// runtime.json
{
  "runtime": {
    "name": "tohelp-agent",
    "invokeTimeoutMs": 120000
  },
  "gateway": {
    "port": 3000,
    "host": "127.0.0.1",
    "auth": { "token": "${GATEWAY_TOKEN}" }
  },
  "skills": {
    "github": { "token": "${GITHUB_TOKEN}" },
    "slack": { "botToken": "${SLACK_BOT_TOKEN}" }
  },
  "channels": {
    "discord": {
      "enabled": true,
      "config": { "botToken": "${DISCORD_BOT_TOKEN}" }
    },
    "telegram": {
      "enabled": true,
      "config": { "botToken": "${TELEGRAM_BOT_TOKEN}" }
    }
  },
  "providers": {
    "anthropic": {
      "enabled": true,
      "default": true,
      "config": { "apiKey": "${ANTHROPIC_API_KEY}" }
    },
    "openai": {
      "enabled": true,
      "config": { "apiKey": "${OPENAI_API_KEY}" }
    }
  }
}
```

### 2.2 字段映射表

| OpenClaw | Agent Runtime | 说明 |
|----------|---------------|------|
| `gateway.port` | `gateway.port` | 保持不变 |
| `gateway.host` | `gateway.host` | 保持不变 |
| `gateway.auth.token` | `gateway.auth.token` | 保持不变 |
| `plugins.entries.*.config` | `skills.*` | 技能配置直接放在 skills 下 |
| `channels.*` | `channels.*` | Channel 配置格式统一 |
| `providers.*` | `providers.*` | Provider 配置格式统一 |

### 2.3 环境变量迁移

**旧环境变量：**

```bash
# OpenClaw
OPENCLAW_GATEWAY_PORT=19001
OPENCLAW_GATEWAY_TOKEN=xxx
ANTHROPIC_API_KEY=sk-ant-xxx
OPENAI_API_KEY=sk-xxx
```

**新环境变量：**

```bash
# Agent Runtime
GATEWAY_PORT=3000
GATEWAY_TOKEN=xxx
ANTHROPIC_API_KEY=sk-ant-xxx
OPENAI_API_KEY=sk-xxx
```

**环境变量对照表：**

| 旧变量 | 新变量 |
|--------|--------|
| `OPENCLAW_GATEWAY_PORT` | `GATEWAY_PORT` |
| `OPENCLAW_GATEWAY_TOKEN` | `GATEWAY_TOKEN` |
| `OPENCLAW_CONFIG_PATH` | 不再需要 |
| `HERMES_CONFIG_JSON` | `RUNTIME_CONFIG_JSON` |
| `TOHELP_PLUGIN_CONFIG_JSON` | `SKILLS_CONFIG_JSON` |

## 3. Hermes 配置迁移

### 3.1 Hermes 配置 → 新配置

**Hermes 旧配置：**

```json
{
  "invokeTimeoutMs": 120000,
  "skills": {
    "github": { "token": "xxx" },
    "slack": { "botToken": "xxx" }
  }
}
```

**新配置：**

```json
{
  "runtime": {
    "invokeTimeoutMs": 120000
  },
  "skills": {
    "github": { "token": "xxx" },
    "slack": { "botToken": "xxx" }
  }
}
```

### 3.2 Memory 配置迁移

**Hermes Memory：**

```bash
TOHELP_MEMORY_KEY_PREFIX=dev:
TOHELP_MEMORY_MAX_VALUE_BYTES=1048576
```

**新配置：**

```bash
RUNTIME_MEMORY_PREFIX=dev:
RUNTIME_MEMORY_MAX_VALUE_BYTES=1048576
```

## 4. 插件迁移

### 4.1 OpenClaw 插件 → 新插件

**OpenClaw 插件结构：**

```
extensions/
  my-plugin/
    index.ts
    openclaw.plugin.json
```

**新插件结构：**

```
src/plugins/
  my-plugin/
    index.ts
    plugin.json
```

### 4.2 插件清单迁移

**OpenClaw 清单：**

```json
{
  "id": "my-plugin",
  "name": "My Plugin",
  "version": "1.0.0",
  "runtime": "openclaw"
}
```

**新清单：**

```json
{
  "id": "my-plugin",
  "name": "My Plugin",
  "version": "1.0.0",
  "runtime": "agent-runtime",
  "type": "skill" | "channel" | "provider"
}
```

### 4.3 插件接口适配

**旧 OpenClaw 接口：**

```typescript
// OpenClaw
class MyPlugin implements OpenClawPlugin {
  async init(ctx: OpenClawContext) { ... }
  async start() { ... }
  async stop() { ... }
}
```

**新接口：**

```typescript
// Agent Runtime
import type { SkillPlugin, PluginContext } from '../runtime/types.js';

class MyPlugin implements SkillPlugin {
  async onInit(ctx: PluginContext) { ... }
  async onStart() { ... }
  async onStop() { ... }
  getSkillDefinition() { ... }
}
```

## 5. MCP 工具迁移

### 5.1 工具名称对照

| 旧工具名 | 新工具名 |
|----------|----------|
| `tohelp_ping` | `agent_ping` |
| `tohelp_list_new_core_skills` | `agent_list_skills` |
| `tohelp_invoke_skill` | `agent_invoke` |
| `tohelp_memory` | `agent_memory` |
| `tohelp_resolve_workspace_path` | `agent_resolve_path` |

### 5.2 MCP 端点迁移

**旧端点：**

```
GET /health
POST /invoke
GET /tools
```

**新端点：**

```
GET  /health
POST /invoke
GET  /tools
GET  /memory?action=list
POST /memory (body: { action, key, value })
```

## 6. Channel 迁移

### 6.1 Discord 配置

**旧配置：**

```json
{
  "channels": {
    "discord": {
      "botToken": "xxx",
      "guildId": "123"
    }
  }
}
```

**新配置：**

```json
{
  "channels": {
    "discord": {
      "enabled": true,
      "config": {
        "botToken": "xxx",
        "guildId": "123",
        "allowedChannels": ["456"]
      }
    }
  }
}
```

### 6.2 Telegram 配置

**旧配置：**

```json
{
  "channels": {
    "telegram": {
      "botToken": "xxx"
    }
  }
}
```

**新配置：**

```json
{
  "channels": {
    "telegram": {
      "enabled": true,
      "config": {
        "botToken": "xxx",
        "allowedChats": [],
        "parseMode": "MarkdownV2"
      }
    }
  }
}
```

## 7. Provider 迁移

### 7.1 Anthropic 配置

**旧配置：**

```json
{
  "providers": {
    "anthropic": {
      "apiKey": "xxx",
      "model": "claude-3-5-sonnet"
    }
  }
}
```

**新配置：**

```json
{
  "providers": {
    "anthropic": {
      "enabled": true,
      "default": true,
      "config": {
        "apiKey": "xxx",
        "defaultModel": "claude-3-5-sonnet-20241022"
      }
    }
  }
}
```

## 8. 迁移脚本

### 8.1 自动迁移脚本

```typescript
// scripts/migrate-config.ts
import fs from 'node:fs';

function migrateConfig(inputPath: string, outputPath: string) {
  const oldConfig = JSON.parse(fs.readFileSync(inputPath, 'utf-8'));

  const newConfig = {
    runtime: {
      name: 'tohelp-agent',
      invokeTimeoutMs: 120000,
    },
    gateway: oldConfig.gateway ? {
      port: oldConfig.gateway.port || 3000,
      host: oldConfig.gateway.host || '127.0.0.1',
      auth: oldConfig.gateway.auth,
    } : undefined,
    skills: oldConfig.plugins?.entries?.['tohelp-openclaw']?.config?.skills,
    channels: oldConfig.channels,
    providers: oldConfig.providers,
  };

  // 清理 undefined 值
  const cleaned = JSON.parse(JSON.stringify(newConfig));

  fs.writeFileSync(outputPath, JSON.stringify(cleaned, null, 2));
  console.log(`Migrated config written to ${outputPath}`);
}

// 使用
migrateConfig('./openclaw.json', './runtime.json');
```

### 8.2 迁移验证

```bash
# 验证配置
npm run doctor:config -- --config runtime.json

# 启动测试
npm run gateway -- --config runtime.json
```

## 9. 回滚计划

### 9.1 保留旧配置

在迁移期间，保留旧配置文件：

```bash
cp openclaw.json openclaw.json.bak
```

### 9.2 并行运行

在新旧配置之间切换：

```bash
# 使用旧配置
npm run gateway -- --config openclaw.json

# 使用新配置
npm run gateway -- --config runtime.json
```

### 9.3 回滚步骤

1. 停止新运行时
2. 恢复旧配置文件
3. 重启旧运行时

## 10. 常见问题

### Q: MCP 工具调用失败？

检查工具名称是否已更新为新名称。

### Q: Channel 连接失败？

确认 `channels.*.config` 路径正确。

### Q: Provider 调用失败？

确认 `providers.*.config.apiKey` 路径正确。

### Q: Memory 数据丢失？

检查 `memory.prefix` 是否与旧配置一致。

## 11. 相关文档

| 文档 | 说明 |
|------|------|
| `docs/AGENT-RUNTIME.md` | 整体架构设计 |
| `docs/CONFIGURATION.md` | 配置参考 |
| `docs/CHANNELS.md` | Channel 实现指南 |
| `docs/PROVIDERS.md` | Provider 实现指南 |

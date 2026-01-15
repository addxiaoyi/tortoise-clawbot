# 插件开发指南

> 完整的插件生命周期、类型系统和最佳实践

## 1. 概述

插件系统是 Agent Runtime 的核心扩展机制，支持 Skill、Channel、Provider 和系统插件。

### 插件类型

| 类型 | 说明 | 实现接口 |
|------|------|----------|
| **SkillPlugin** | 技能插件，提供工具 | `SkillPlugin` |
| **ChannelPlugin** | 频道插件，消息收发 | `ChannelAdapter` |
| **ProviderPlugin** | 模型供应商插件 | `ModelProvider` |
| **SystemPlugin** | 系统插件 | `PluginLifecycle` |

## 2. 核心接口

### 2.1 PluginLifecycle

所有插件必须实现的基础接口：

```typescript
interface PluginLifecycle {
  /**
   * 插件加载时调用
   * 用于初始化静态资源、验证配置
   */
  onInit(ctx: PluginContext): Promise<void> | void;

  /**
   * 系统启动时调用
   * 连接外部服务、启动监听等
   */
  onStart?(): Promise<void> | void;

  /**
   * 系统停止时调用
   * 清理资源、关闭连接
   */
  onStop?(): Promise<void> | void;

  /**
   * 配置变更时调用（可选）
   * 支持热重载
   */
  onConfigChange?(newConfig: PluginConfig): Promise<void> | void;
}
```

### 2.2 PluginContext

插件运行时上下文：

```typescript
interface PluginContext {
  readonly meta: PluginMetadata;       // 插件元数据
  readonly logger: PluginLogger;        // 日志接口
  readonly storage: PluginStorage;       // 持久化存储
  readonly events: PluginEventBus;      // 事件总线
  readonly getConfig: <T>() => T;      // 获取配置
}
```

### 2.3 元数据类型

```typescript
interface PluginMetadata {
  id: string;           // 唯一标识符
  name: string;         // 显示名称
  version: string;      // 版本号
  description?: string; // 描述
  author?: string;      // 作者
  license?: string;     // 许可证
}
```

## 3. 开发 Skill 插件

### 3.1 完整示例

```typescript
// src/skills/my-skill/plugin.ts
import type { SkillPlugin, SkillDefinition, PluginContext } from '../runtime/types.js';

export class MySkillPlugin implements SkillPlugin {
  private ctx?: PluginContext;

  async onInit(ctx: PluginContext): Promise<void> {
    this.ctx = ctx;
    ctx.logger.info('[my-skill] Initialized');
  }

  async onStart(): Promise<void> {
    this.ctx?.logger.info('[my-skill] Started');
  }

  async onStop(): Promise<void> {
    this.ctx?.logger.info('[my-skill] Stopped');
  }

  getSkillDefinition(): SkillDefinition {
    return {
      name: 'my-skill',
      version: '1.0.0',
      description: 'A custom skill for doing something useful',
      tools: [
        {
          name: 'do_something',
          description: 'Do something useful with the provided input',
          parameters: [
            {
              name: 'input',
              type: 'string',
              description: 'The input to process',
              required: true,
            },
            {
              name: 'options',
              type: 'object',
              description: 'Optional settings',
              required: false,
              default: {},
            },
          ],
          execute: this.doSomething.bind(this),
        },
        {
          name: 'get_status',
          description: 'Get the current status of the skill',
          parameters: [],
          execute: this.getStatus.bind(this),
        },
      ],
    };
  }

  private async doSomething(
    args: Record<string, unknown>,
    ctx: PluginContext
  ): Promise<unknown> {
    const input = args.input as string;
    const options = (args.options || {}) as Record<string, unknown>;

    this.ctx?.logger.info(`[my-skill] Processing: ${input}`);

    // 业务逻辑
    const result = await this.processInput(input, options);

    return {
      success: true,
      input,
      result,
      timestamp: Date.now(),
    };
  }

  private async getStatus(
    _args: Record<string, unknown>,
    ctx: PluginContext
  ): Promise<unknown> {
    return {
      status: 'running',
      uptime: process.uptime(),
      memory: process.memoryUsage(),
    };
  }

  private async processInput(
    input: string,
    options: Record<string, unknown>
  ): Promise<string> {
    // 实现处理逻辑
    return `Processed: ${input} with options ${JSON.stringify(options)}`;
  }
}
```

### 3.2 注册 Skill

```typescript
// src/bridge/skill-registry.ts
import { MySkillPlugin } from '../skills/my-skill/plugin.js';

export const SKILL_IDS = [
  'github',
  'slack',
  'my-skill', // 添加新 skill
  // ...
] as const;

const factories = {
  // ...
  'my-skill': () => new MySkillPlugin(),
};
```

## 4. 工具参数验证

使用 Zod 进行参数验证：

```typescript
import { z } from 'zod';

const DoSomethingSchema = z.object({
  input: z.string().min(1).max(1000),
  options: z.object({
    mode: z.enum(['fast', 'accurate', 'balanced']).optional(),
    maxRetries: z.number().min(0).max(10).optional(),
  }).optional(),
});

async doSomething(args: Record<string, unknown>, ctx: PluginContext) {
  // 验证参数
  const parsed = DoSomethingSchema.safeParse(args);

  if (!parsed.success) {
    throw new Error(`Invalid arguments: ${parsed.error.message}`);
  }

  const { input, options } = parsed.data;
  // ...
}
```

## 5. 异步工具与进度

```typescript
import { ctx as agentContext } from '../runtime/agent-context.js';

const tool = {
  name: 'long_running_task',
  description: 'A task that takes a long time',
  parameters: [{ name: 'taskId', type: 'string', required: true }],

  async execute(args: Record<string, unknown>, ctx: PluginContext) {
    const taskId = args.taskId as string;
    const total = 100;

    for (let i = 0; i < total; i++) {
      // 更新进度
      await ctx.events.emit('tool:progress', {
        tool: 'long_running_task',
        taskId,
        progress: (i / total) * 100,
        current: i,
        total,
      });

      // 执行一步
      await this.executeStep(taskId, i);

      // 支持取消
      if (ctx.signal?.aborted) {
        return { cancelled: true, completedSteps: i };
      }
    }

    return { success: true, taskId };
  },
};
```

## 6. 错误处理

### 6.1 自定义错误类

```typescript
export class MySkillError extends Error {
  constructor(
    message: string,
    public code: string,
    public details?: Record<string, unknown>
  ) {
    super(message);
    this.name = 'MySkillError';
  }
}

// 使用
throw new MySkillError(
  'Task failed',
  'TASK_FAILED',
  { taskId, reason: 'timeout' }
);
```

### 6.2 错误边界

```typescript
async execute(args: Record<string, unknown>, ctx: PluginContext) {
  try {
    return await this.doWork(args);
  } catch (error) {
    if (error instanceof MySkillError) {
      ctx.logger.error('[my-skill]', error.message, error.details);
      return { error: error.code, message: error.message };
    }

    ctx.logger.error('[my-skill] Unexpected error', error);
    return { error: 'INTERNAL_ERROR', message: 'An unexpected error occurred' };
  }
}
```

## 7. 存储与持久化

### 7.1 使用 PluginStorage

```typescript
async onStart(): Promise<void> {
  const cache = await this.ctx!.storage.getItem<any>('cache');
  if (cache) {
    this.data = cache;
  }
}

async saveState(): Promise<void> {
  await this.ctx!.storage.setItem('cache', this.data);
}
```

### 7.2 存储接口

```typescript
interface PluginStorage {
  getItem<T>(key: string): Promise<T | null>;
  setItem<T>(key: string, value: T): Promise<void>;
  removeItem(key: string): Promise<void>;
  clear(): Promise<void>;
}
```

## 8. 事件总线

### 8.1 发布事件

```typescript
// 在工具中发布
await this.ctx!.events.emit('my-skill:task-completed', {
  taskId,
  result,
  duration: Date.now() - startTime,
});

// 发布错误
await this.ctx!.events.emit('my-skill:error', {
  taskId,
  error: error.message,
});
```

### 8.2 订阅事件

```typescript
async onStart(): Promise<void> {
  this.unsubscribe = this.ctx!.events.on(
    'other-skill:event',
    (payload) => this.handleEvent(payload)
  );
}

async onStop(): Promise<void> {
  this.unsubscribe?.();
}
```

## 9. 配置管理

### 9.1 类型化配置

```typescript
interface MySkillConfig {
  apiKey: string;
  baseUrl?: string;
  timeoutMs?: number;
  retries?: number;
}

async onInit(ctx: PluginContext): Promise<void> {
  const config = ctx.getConfig<MySkillConfig>();

  if (!config.apiKey) {
    throw new Error('apiKey is required');
  }

  this.config = {
    timeoutMs: 30000,
    retries: 3,
    ...config,
  };
}
```

### 9.2 热重载

```typescript
async onConfigChange(newConfig: PluginConfig): Promise<void> {
  this.ctx?.logger.info('[my-skill] Config changed, reloading...');

  await this.onStop();
  await this.onInit({
    ...this.ctx!,
    getConfig: () => newConfig as MySkillConfig,
  });
  await this.onStart();
}
```

## 10. 依赖注入

### 10.1 声明依赖

```typescript
interface MySkillDeps {
  httpClient: HttpClient;
  cache: CacheService;
  logger: Logger;
}
```

### 10.2 注入依赖

```typescript
class MySkillPlugin implements SkillPlugin {
  private deps?: MySkillDeps;

  setDependencies(deps: MySkillDeps): void {
    this.deps = deps;
  }

  async execute(args: Record<string, unknown>, ctx: PluginContext) {
    if (!this.deps) {
      throw new Error('Dependencies not set');
    }

    return this.doWork(args, this.deps);
  }
}
```

### 10.3 容器注册

```typescript
// src/runtime/plugin/container.ts
class PluginContainer {
  register(plugin: PluginLifecycle, deps?: unknown): void {
    if ('setDependencies' in plugin && typeof plugin.setDependencies === 'function') {
      (plugin as any).setDependencies(deps);
    }
    // ...
  }
}
```

## 11. 测试

### 11.1 单元测试

```typescript
import { describe, it, expect, vi, beforeEach } from 'vitest';

describe('MySkillPlugin', () => {
  let plugin: MySkillPlugin;
  let mockCtx: PluginContext;

  beforeEach(() => {
    mockCtx = {
      meta: { id: 'test', name: 'Test', version: '1.0.0' },
      logger: {
        debug: vi.fn(),
        info: vi.fn(),
        warn: vi.fn(),
        error: vi.fn(),
      },
      storage: {
        getItem: vi.fn().mockResolvedValue(null),
        setItem: vi.fn().mockResolvedValue(undefined),
        removeItem: vi.fn().mockResolvedValue(undefined),
        clear: vi.fn().mockResolvedValue(undefined),
      },
      events: {
        emit: vi.fn(),
        on: vi.fn().mockReturnValue(vi.fn()),
        off: vi.fn(),
      },
      getConfig: () => ({ apiKey: 'test-key' }),
    };
  });

  it('should initialize', async () => {
    plugin = new MySkillPlugin();
    await plugin.onInit(mockCtx);
    expect(mockCtx.logger.info).toHaveBeenCalledWith('[my-skill] Initialized');
  });

  it('should do something', async () => {
    plugin = new MySkillPlugin();
    await plugin.onInit(mockCtx);

    const def = plugin.getSkillDefinition();
    const tool = def.tools.find(t => t.name === 'do_something');

    const result = await tool!.execute({ input: 'test' }, mockCtx);

    expect(result).toMatchObject({
      success: true,
      input: 'test',
    });
  });
});
```

### 11.2 集成测试

```typescript
describe('MySkillPlugin Integration', () => {
  it('should work with real dependencies', async () => {
    const container = new PluginContainer();
    const plugin = new MySkillPlugin();

    container.register(plugin, {
      httpClient: new RealHttpClient(),
      cache: new RedisCache(),
      logger: console,
    });

    await container.start();

    // 测试工具调用
    const result = await container.invoke('my-skill', 'do_something', {
      input: 'integration test',
    });

    expect(result.success).toBe(true);

    await container.stop();
  });
});
```

## 12. 发布插件

### 12.1 插件清单

```json
// my-skill/openclaw.plugin.json
{
  "id": "tohelp-my-skill",
  "name": "My Skill",
  "version": "1.0.0",
  "description": "A custom skill for doing something useful",
  "author": "Your Name",
  "license": "MIT",
  "runtime": "hermes",
  "main": "./dist/plugin.js",
  "skill": {
    "name": "my-skill",
    "tools": ["do_something", "get_status"]
  }
}
```

### 12.2 package.json

```json
{
  "name": "@tohelp/my-skill",
  "version": "1.0.0",
  "type": "module",
  "main": "./dist/plugin.js",
  "exports": {
    ".": {
      "import": "./dist/plugin.js"
    }
  },
  "peerDependencies": {
    "openclaw": ">=0.1.0"
  }
}
```

## 13. 最佳实践

1. **始终验证输入** - 使用 Zod 或类似库
2. **优雅错误处理** - 返回结构化错误而非抛出
3. **记录日志** - 使用 ctx.logger 记录关键操作
4. **支持取消** - 检查 ctx.signal?.aborted
5. **进度报告** - 对于长时间操作发布进度事件
6. **配置验证** - 在 onInit 时验证必需配置
7. **资源清理** - 在 onStop 时释放所有资源
8. **测试覆盖** - 单元测试 + 集成测试

## 14. 相关文档

| 文档 | 说明 |
|------|------|
| `docs/AGENT-RUNTIME.md` | 整体架构设计 |
| `docs/CHANNELS.md` | Channel 实现指南 |
| `docs/PROVIDERS.md` | Provider 实现指南 |

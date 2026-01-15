# Channel 系统实现指南

> OpenClaw 兼容的 Channel 适配器开发指南

## 1. 概述

Channel 系统让你的 Agent 可以接入多个消息平台（Discord、Telegram、Slack 等），实现跨平台消息收发。

### 支持的 Channel

| Channel | 状态 | 能力 |
|---------|------|------|
| Telegram | ✅ 已实现 | text, markdown, html, images, audio, video, files, typing, reactions, reply, threads |
| Discord | ✅ 已实现 | text, markdown, images, audio, video, files, typing, reactions, reply, threads |
| Slack | ✅ 已实现 | text, markdown, images, files, typing, reactions, reply, threads |
| WhatsApp | ✅ 已实现 | text, images, audio, video, files |
| Signal | 🔲 待实现 | — |
| iMessage | 🔲 待实现 | — |

## 2. 架构设计

```
┌─────────────────────────────────────────────────────────────┐
│                      Agent Runtime                           │
│  ┌─────────────────────────────────────────────────────┐   │
│  │               ChannelRegistry                         │   │
│  │  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐  │   │
│  │  │Telegram │ │ Discord │ │  Slack  │ │WhatsApp│  │   │
│  │  └─────────┘ └─────────┘ └─────────┘ └─────────┘  │   │
│  └─────────────────────────────────────────────────────┘   │
│                            │                                │
│  ┌─────────────────────────┴─────────────────────────────┐ │
│  │                    EventBus                            │ │
│  │  channel:message → Agent 处理                          │ │
│  └─────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

## 3. 基础接口

### ChannelAdapter

```typescript
interface ChannelAdapter extends PluginLifecycle {
  // Channel 名称（唯一标识）
  readonly name: string;

  // 支持的能力列表
  readonly capabilities: ChannelCapability[];

  // 发送消息
  send(message: OutboundMessage): Promise<void>;

  // 格式化内容为 Channel 特定格式
  formatForChannel(content: string): Promise<string>;

  // 处理传入的更新/ webhook
  handleUpdate(update: unknown): Promise<void>;
}
```

### ChannelCapability

```typescript
type ChannelCapability =
  | 'text'           // 纯文本
  | 'markdown'       // Markdown 格式
  | 'html'          // HTML 格式
  | 'images'         // 图片支持
  | 'audio'         // 音频支持
  | 'video'         // 视频支持
  | 'files'         // 文件支持
  | 'typing'         // 打字状态
  | 'read-receipts'  // 已读回执
  | 'reactions'      // 表情反应
  | 'threads'        // 线程/回复
  | 'reply';         // 回复功能
```

### 消息格式

```typescript
// 发送给 Channel 的消息
interface OutboundMessage {
  to: string;           // 接收者 ID
  content: string;      // 消息内容
  channel: string;      // Channel 名称
  options?: {
    replyTo?: string;   // 回复的消息 ID
    threadId?: string;  // 线程 ID
    parseMode?: 'markdown' | 'html' | 'plain';
    typing?: boolean;
  };
}

// 从 Channel 接收的消息
interface ChannelMessage {
  id: string;
  channel: string;
  from: string;
  content: string;
  timestamp: number;
  metadata?: Record<string, unknown>;
}
```

## 4. 开发新的 Channel

### 4.1 创建 Channel 类

```typescript
// src/channels/mychannel.ts
import { BaseChannelAdapter } from './base.js';
import type {
  ChannelCapability,
  OutboundMessage,
  PluginContext,
} from '../runtime/types.js';

export interface MyChannelConfig {
  apiKey: string;
  allowedUsers?: string[];
}

export class MyChannelAdapter extends BaseChannelAdapter {
  readonly name = 'mychannel';
  readonly capabilities: ChannelCapability[] = [
    'text',
    'markdown',
    'images',
  ];

  private config?: MyChannelConfig;

  async onInit(ctx: PluginContext): Promise<void> {
    await super.onInit(ctx);
    this.config = ctx.getConfig<MyChannelConfig>();

    if (!this.config?.apiKey) {
      ctx.logger.warn('[mychannel] No API key configured');
    }
  }

  async onStart(): Promise<void> {
    await super.onStart();

    if (!this.config?.apiKey) {
      throw new Error('[mychannel] apiKey is required');
    }

    // 验证连接
    await this.verifyConnection();
    this.ctx?.logger.info('[mychannel] Connected successfully');
  }

  async send(message: OutboundMessage): Promise<void> {
    this.validateMessage(message);

    const formatted = await this.formatForChannel(message.content);
    const payload = this.buildPayload(message.to, formatted);

    await this.apiCall('/send', payload);
    this.ctx?.logger.debug(`[mychannel] Message sent to ${message.to}`);
  }

  async formatForChannel(content: string): Promise<string> {
    // 1. 将 Markdown 转换为 Channel 格式
    // 2. 处理特殊字符
    // 3. 返回格式化后的字符串
    return content
      .replace(/\*\*([^*]+)\*\*/g, '<b>$1</b>')
      .replace(/\*([^*]+)\*/g, '<i>$1</i>');
  }

  async handleUpdate(update: unknown): Promise<void> {
    // 解析 webhook 传入的更新
    const message = this.parseIncomingMessage(update);

    // 检查权限
    if (!this.isAllowedUser(message.from)) {
      return;
    }

    // 发送到事件总线
    this.ctx?.events.emit('channel:message', {
      channel: this.name,
      message,
    });
  }

  protected parseIncomingMessage(raw: unknown): ChannelMessage {
    // 解析为统一的 ChannelMessage 格式
    const data = raw as {
      id: string;
      sender: { id: string };
      text: string;
      timestamp: number;
    };

    return {
      id: data.id,
      channel: this.name,
      from: data.sender.id,
      content: data.text,
      timestamp: data.timestamp,
    };
  }

  private isAllowedUser(userId: string): boolean {
    if (!this.config?.allowedUsers?.length) {
      return true;
    }
    return this.config.allowedUsers.includes(userId);
  }

  private buildPayload(to: string, content: string): Record<string, unknown> {
    return { recipient: to, message: content };
  }

  private async apiCall(endpoint: string, payload: Record<string, unknown>): Promise<void> {
    const response = await fetch(`https://api.mychannel.com${endpoint}`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${this.config?.apiKey}`,
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(payload),
    });

    if (!response.ok) {
      throw new Error(`MyChannel API error: ${response.status}`);
    }
  }

  private async verifyConnection(): Promise<void> {
    // 实现连接验证
  }
}
```

### 4.2 注册 Channel

```typescript
// src/channels/index.ts
import { ChannelRegistry } from './base.js';
import { MyChannelAdapter } from './mychannel.js';

export const channelRegistry = new ChannelRegistry();

// 注册新 Channel
channelRegistry.register(new MyChannelAdapter());

// 或者批量注册
const channels = [
  new MyChannelAdapter(),
  // ...
];

for (const channel of channels) {
  channelRegistry.register(channel);
}
```

### 4.3 配置示例

```json
{
  "channels": {
    "mychannel": {
      "enabled": true,
      "config": {
        "apiKey": "${MYCHANNEL_API_KEY}",
        "allowedUsers": ["123456", "789012"]
      }
    }
  }
}
```

## 5. Webhook 处理

大多数 Channel 使用 Webhook 接收消息：

```typescript
// 在 Gateway 中添加 webhook 端点
async function setupWebhooks(gateway: GatewayServer) {
  // Telegram
  gateway.router.post('/webhook/telegram', async (req, res) => {
    const telegram = channelRegistry.get('telegram');
    await telegram?.handleUpdate(req.body);
    res.sendStatus(200);
  });

  // Discord (使用交互式组件)
  gateway.router.post('/webhook/discord', async (req, res) => {
    const discord = channelRegistry.get('discord');
    await discord?.handleUpdate(req.body);
    res.sendStatus(200);
  });
}
```

## 6. 事件处理

```typescript
// 监听所有 Channel 消息
eventBus.on('channel:message', async ({ channel, message }) => {
  console.log(`[${channel}] ${message.from}: ${message.content}`);

  // 调用 Agent 处理
  const response = await agent.process({
    userId: message.from,
    channel,
    message: message.content,
    context: message.metadata,
  });

  // 发送回复
  const channelAdapter = channelRegistry.get(channel);
  await channelAdapter.send({
    to: message.from,
    content: response,
    channel,
  });
});
```

## 7. 最佳实践

### 7.1 消息验证

```typescript
protected validateMessage(message: OutboundMessage): void {
  if (!message.to) {
    throw new Error('Message must have a recipient (to)');
  }
  if (!message.content || !message.content.trim()) {
    throw new Error('Message must have non-empty content');
  }
  // 添加长度限制
  if (message.content.length > 4096) {
    throw new Error('Message content exceeds maximum length');
  }
}
```

### 7.2 速率限制

```typescript
private rateLimiter = new Map<string, number[]>();

protected async checkRateLimit(userId: string): Promise<boolean> {
  const now = Date.now();
  const window = 60 * 1000; // 1 分钟窗口
  const maxMessages = 20;

  const timestamps = this.rateLimiter.get(userId) || [];
  const recent = timestamps.filter(t => now - t < window);

  if (recent.length >= maxMessages) {
    return false;
  }

  recent.push(now);
  this.rateLimiter.set(userId, recent);
  return true;
}
```

### 7.3 错误处理

```typescript
async send(message: OutboundMessage): Promise<void> {
  try {
    await this.sendInternal(message);
  } catch (error) {
    this.ctx?.logger.error('[mychannel] Send failed', error);

    // 重试逻辑
    if (this.isRetryable(error)) {
      await this.delay(1000);
      return this.sendInternal(message);
    }

    throw error;
  }
}
```

## 8. 测试

```typescript
import { describe, it, expect, vi } from 'vitest';

describe('MyChannelAdapter', () => {
  const mockCtx = {
    meta: { id: 'test', name: 'Test', version: '1.0.0' },
    logger: { info: vi.fn(), debug: vi.fn(), warn: vi.fn(), error: vi.fn() },
    storage: { getItem: vi.fn(), setItem: vi.fn() },
    events: { emit: vi.fn(), on: vi.fn(), off: vi.fn() },
    getConfig: () => ({ apiKey: 'test-key' }),
  };

  it('should send message', async () => {
    const channel = new MyChannelAdapter();
    await channel.onInit(mockCtx as any);
    await channel.onStart();

    global.fetch = vi.fn().mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({ success: true }),
    });

    await channel.send({
      to: 'user123',
      content: 'Hello!',
      channel: 'mychannel',
    });

    expect(fetch).toHaveBeenCalled();
  });
});
```

## 9. 相关文档

| 文档 | 说明 |
|------|------|
| `docs/AGENT-RUNTIME.md` | 整体架构设计 |
| `docs/PROVIDERS.md` | 模型 Provider 实现指南 |
| `docs/PLUGIN-DEV.md` | 插件开发指南 |

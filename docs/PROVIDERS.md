# Provider 系统实现指南

> OpenClaw 兼容的模型 Provider 开发指南

## 1. 概述

Provider 系统让你的 Agent 可以接入多种 AI 模型服务（Anthropic Claude、OpenAI GPT、Ollama 本地模型等）。

### 支持的 Provider

| Provider | 状态 | 模型 | 流式 | Embeddings |
|----------|------|------|------|------------|
| Anthropic | ✅ 已实现 | claude-3-5-sonnet, claude-3-opus, claude-3-haiku | ✅ | ❌ |
| OpenAI | ✅ 已实现 | gpt-4o, gpt-4-turbo, gpt-3.5-turbo | ✅ | ✅ |
| Ollama | ✅ 已实现 | 本地模型（llama3.2, mistral 等） | ✅ | ✅ |
| OpenRouter | 🔲 待实现 | 聚合多个模型 | ✅ | ✅ |
| Groq | 🔲 待实现 | llama, mistral | ✅ | ✅ |
| DeepSeek | 🔲 待实现 | deepseek-chat, deepseek-coder | ✅ | ✅ |

## 2. 架构设计

```
┌─────────────────────────────────────────────────────────────┐
│                      Agent Runtime                           │
│  ┌─────────────────────────────────────────────────────┐   │
│  │               ProviderRegistry                         │   │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐           │   │
│  │  │Anthropic│ │  OpenAI  │ │  Ollama  │           │   │
│  │  └──────────┘ └──────────┘ └──────────┘           │   │
│  └─────────────────────────────────────────────────────┘   │
│                            │                                │
│  ┌─────────────────────────┴─────────────────────────────┐ │
│  │                    Completion API                       │ │
│  │  complete() → 单次调用                                │ │
│  │  completeStream() → 流式调用                           │ │
│  │  embed() → 向量嵌入                                   │ │
│  └─────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

## 3. 基础接口

### ModelProvider

```typescript
interface ModelProvider extends PluginLifecycle {
  readonly name: string;           // Provider 名称
  readonly defaultModel: string;    // 默认模型
  readonly supportedModels: string[]; // 支持的模型列表

  // 单次完成调用
  complete(options: CompletionOptions): Promise<CompletionResult>;

  // 流式完成调用
  completeStream(options: CompletionOptions): AsyncGenerator<StreamChunk>;

  // 向量嵌入
  embed(options: EmbedOptions): Promise<EmbedResult>;
}
```

### 请求/响应类型

```typescript
interface CompletionOptions {
  model: string;
  messages: Array<{
    role: 'system' | 'user' | 'assistant' | 'developer';
    content: string;
  }>;
  temperature?: number;      // 0-2，默认 1
  maxTokens?: number;       // 最大 token 数
  topP?: number;           // top_p 采样
  stop?: string[];         // 停止序列
  stream?: boolean;        // 是否流式
  tools?: Tool[];          // 工具定义
  toolChoice?: ToolChoice;  // 工具选择策略
}

interface CompletionResult {
  content: string;
  usage?: {
    promptTokens: number;
    completionTokens: number;
    totalTokens: number;
  };
  model: string;
  stopReason?: string;
  toolCalls?: Array<{
    name: string;
    arguments: string;
  }>;
}

interface StreamChunk {
  type: 'content' | 'tool-call' | 'done' | 'error';
  content?: string;
  toolCall?: {
    name: string;
    arguments: string;
  };
  usage?: CompletionResult['usage'];
  stopReason?: string;
}

interface Tool {
  name: string;
  description?: string;
  inputSchema: Record<string, unknown>;
}

type ToolChoice = 'auto' | 'none' | { type: 'function'; function: { name: string } };
```

## 4. 开发新的 Provider

### 4.1 创建 Provider 类

```typescript
// src/providers/myprovider.ts
import { BaseModelProvider } from './base.js';
import type {
  CompletionOptions,
  CompletionResult,
  EmbedOptions,
  EmbedResult,
  PluginContext,
  StreamChunk,
} from '../runtime/types.js';

export interface MyProviderConfig {
  apiKey?: string;
  baseUrl?: string;
  defaultModel?: string;
  timeoutMs?: number;
}

export class MyProvider extends BaseModelProvider {
  readonly name = 'myprovider';
  readonly defaultModel = 'my-model';
  readonly supportedModels = [
    'my-model',
    'my-model-fast',
    'my-model-large',
  ];

  private baseUrl = 'https://api.myprovider.com/v1';
  private timeoutMs = 60000;

  async onInit(ctx: PluginContext): Promise<void> {
    await super.onInit(ctx);
    const config = ctx.getConfig<MyProviderConfig>();

    if (config?.baseUrl) this.baseUrl = config.baseUrl;
    if (config?.defaultModel) this.defaultModel = config.defaultModel;
    if (config?.timeoutMs) this.timeoutMs = config.timeoutMs;
  }

  async complete(options: CompletionOptions): Promise<CompletionResult> {
    const model = options.model || this.defaultModel;
    this.validateModel(model);

    // 构建请求体
    const body = this.buildRequestBody(options);

    // 发送请求
    const response = await this.request('/chat/completions', body);

    // 解析响应
    return this.parseResponse(response, model);
  }

  async *completeStream(options: CompletionOptions): AsyncGenerator<StreamChunk> {
    const body = {
      ...this.buildRequestBody(options),
      stream: true,
    };

    const response = await fetch(`${this.baseUrl}/chat/completions`, {
      method: 'POST',
      headers: this.getHeaders(),
      body: JSON.stringify(body),
    });

    if (!response.ok) {
      throw new Error(`MyProvider API error: ${response.status}`);
    }

    const reader = response.body?.getReader();
    if (!reader) throw new Error('No response body');

    const decoder = new TextDecoder();

    while (true) {
      const { done, value } = await reader.read();
      if (done) break;

      const text = decoder.decode(value, { stream: true });
      const lines = text.split('\n').filter(l => l.trim());

      for (const line of lines) {
        if (line.startsWith('data: ')) {
          try {
            const chunk = JSON.parse(line.slice(6));
            const parsed = this.parseStreamChunk(chunk);
            if (parsed) yield parsed;
          } catch {}
        }
      }
    }

    yield { type: 'done' };
  }

  async embed(options: EmbedOptions): Promise<EmbedResult> {
    const model = (options.model || this.defaultModel).split(':')[0];
    const input = typeof options.input === 'string'
      ? [options.input]
      : options.input;

    const response = await this.request('/embeddings', {
      model,
      input,
    }) as { data?: Array<{ embedding: number[] }> };

    return {
      embeddings: (response.data || []).map(d => d.embedding),
      model,
    };
  }

  // ==================== 辅助方法 ====================

  private buildRequestBody(options: CompletionOptions): Record<string, unknown> {
    return {
      model: options.model || this.defaultModel,
      messages: options.messages,
      temperature: options.temperature ?? 0.7,
      max_tokens: options.maxTokens,
      top_p: options.topP,
      stop: options.stop,
      tools: options.tools?.map(t => ({
        type: 'function',
        function: {
          name: t.name,
          description: t.description,
          parameters: t.inputSchema,
        },
      })),
      tool_choice: options.toolChoice,
    };
  }

  private getHeaders(): Record<string, string> {
    return {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${this.apiKey}`,
    };
  }

  private async request(endpoint: string, body: Record<string, unknown>): Promise<unknown> {
    const controller = new AbortController();
    const timeout = setTimeout(() => controller.abort(), this.timeoutMs);

    try {
      const response = await fetch(`${this.baseUrl}${endpoint}`, {
        method: 'POST',
        headers: this.getHeaders(),
        body: JSON.stringify(body),
        signal: controller.signal,
      });

      if (!response.ok) {
        const error = await response.text();
        throw new Error(`MyProvider API error: ${response.status} ${error}`);
      }

      return response.json();
    } finally {
      clearTimeout(timeout);
    }
  }

  private parseResponse(response: unknown, model: string): CompletionResult {
    // 根据 API 响应格式解析
    const data = response as {
      choices: Array<{
        message: { content?: string; tool_calls?: Array<{ function: { name: string; arguments: string } }> };
        finish_reason: string;
      }>;
      usage?: { prompt_tokens: number; completion_tokens: number; total_tokens: number };
    };

    const choice = data.choices[0];

    return {
      content: choice.message.content || '',
      model,
      usage: data.usage ? {
        promptTokens: data.usage.prompt_tokens,
        completionTokens: data.usage.completion_tokens,
        totalTokens: data.usage.total_tokens,
      } : undefined,
      stopReason: choice.finish_reason,
      toolCalls: choice.message.tool_calls?.map(tc => ({
        name: tc.function.name,
        arguments: tc.function.arguments,
      })),
    };
  }

  private parseStreamChunk(chunk: unknown): StreamChunk | null {
    // 解析 SSE 数据块
    const data = chunk as {
      choices?: Array<{
        delta?: { content?: string; tool_calls?: Array<{ function: { name: string; arguments: string } }> };
        finish_reason?: string;
      }>;
    };

    const delta = data.choices?.[0]?.delta;

    if (delta?.content) {
      return { type: 'content', content: delta.content };
    }

    if (delta?.tool_calls?.[0]) {
      const tc = delta.tool_calls[0].function;
      return {
        type: 'tool-call',
        toolCall: { name: tc.name, arguments: tc.arguments },
      };
    }

    return null;
  }
}
```

### 4.2 注册 Provider

```typescript
// src/providers/index.ts
import { ProviderRegistry } from './base.js';
import { MyProvider } from './myprovider.js';

export const providerRegistry = new ProviderRegistry();

// 注册并设为默认
providerRegistry.register(new MyProvider(), true);

// 或者只注册
providerRegistry.register(new MyProvider());
```

### 4.3 配置示例

```json
{
  "providers": {
    "myprovider": {
      "enabled": true,
      "default": true,
      "config": {
        "apiKey": "${MYPROVIDER_API_KEY}",
        "baseUrl": "https://api.myprovider.com/v1",
        "defaultModel": "my-model",
        "timeoutMs": 60000
      }
    }
  }
}
```

## 5. 使用 Provider

### 5.1 基本调用

```typescript
// 获取默认 Provider
const provider = providerRegistry.getDefault();

// 调用 complete
const result = await provider.complete({
  model: 'gpt-4o',
  messages: [
    { role: 'system', content: 'You are a helpful assistant.' },
    { role: 'user', content: 'Hello!' },
  ],
  temperature: 0.7,
  maxTokens: 1000,
});

console.log(result.content);
console.log(result.usage);
```

### 5.2 流式调用

```typescript
const provider = providerRegistry.get('openai');

for await (const chunk of provider.completeStream({
  model: 'gpt-4o',
  messages: [{ role: 'user', content: 'Tell me a story' }],
})) {
  switch (chunk.type) {
    case 'content':
      process.stdout.write(chunk.content);
      break;
    case 'tool-call':
      console.log('Tool call:', chunk.toolCall);
      break;
    case 'done':
      console.log('Done! Usage:', chunk.usage);
      break;
    case 'error':
      console.error('Error:', chunk);
      break;
  }
}
```

### 5.3 工具调用（Function Calling）

```typescript
const provider = providerRegistry.get('openai');

const result = await provider.complete({
  model: 'gpt-4o',
  messages: [{ role: 'user', content: 'What is the weather in Tokyo?' }],
  tools: [
    {
      name: 'get_weather',
      description: 'Get current weather for a location',
      inputSchema: {
        type: 'object',
        properties: {
          location: { type: 'string' },
          unit: { type: 'string', enum: ['celsius', 'fahrenheit'] },
        },
        required: ['location'],
      },
    },
  ],
  toolChoice: 'auto',
});

if (result.toolCalls?.length) {
  const toolCall = result.toolCalls[0];
  console.log(`Call ${toolCall.name} with:`, toolCall.arguments);
}
```

### 5.4 Embeddings

```typescript
const provider = providerRegistry.get('openai');

const result = await provider.embed({
  model: 'text-embedding-3-small',
  input: ['Hello world', 'How are you?'],
});

console.log('Embeddings:', result.embeddings);
console.log('Dimensions:', result.embeddings[0].length);
```

## 6. 模型选择策略

```typescript
class ModelSelector {
  private providers: ProviderRegistry;

  selectModel(task: string): { provider: string; model: string } {
    // 根据任务类型选择模型
    if (task.includes('code') || task.includes('debug')) {
      return { provider: 'anthropic', model: 'claude-3-5-sonnet-20241022' };
    }

    if (task.includes('fast') || task.includes('simple')) {
      return { provider: 'openai', model: 'gpt-4o-mini' };
    }

    // 默认
    return { provider: 'openai', model: 'gpt-4o' };
  }
}
```

## 7. 错误处理与重试

```typescript
class RetryProvider implements ModelProvider {
  constructor(
    private inner: ModelProvider,
    private maxRetries = 3,
    private baseDelay = 1000,
  ) {}

  async complete(options: CompletionOptions): Promise<CompletionResult> {
    let lastError: Error;

    for (let i = 0; i < this.maxRetries; i++) {
      try {
        return await this.inner.complete(options);
      } catch (err) {
        lastError = err as Error;

        if (!this.isRetryable(err)) {
          throw err;
        }

        // 指数退避
        const delay = this.baseDelay * Math.pow(2, i);
        await new Promise(r => setTimeout(r, delay));
      }
    }

    throw lastError!;
  }

  private isRetryable(err: unknown): boolean {
    if (err instanceof Error) {
      // 网络错误、超时、5xx 服务器错误可重试
      return err.message.includes('timeout') ||
             err.message.includes('ECONNRESET') ||
             err.message.includes('500');
    }
    return false;
  }
}
```

## 8. 测试

```typescript
import { describe, it, expect, vi } from 'vitest';

describe('MyProvider', () => {
  const mockCtx = {
    meta: { id: 'test', name: 'Test', version: '1.0.0' },
    logger: { info: vi.fn(), debug: vi.fn(), warn: vi.fn(), error: vi.fn() },
    getConfig: () => ({ apiKey: 'test-key', baseUrl: 'https://test.com' }),
  };

  it('should complete request', async () => {
    global.fetch = vi.fn().mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({
        choices: [{ message: { content: 'Hello!' }, finish_reason: 'stop' }],
        usage: { prompt_tokens: 10, completion_tokens: 5, total_tokens: 15 },
      }),
    });

    const provider = new MyProvider();
    await provider.onInit(mockCtx as any);

    const result = await provider.complete({
      model: 'test-model',
      messages: [{ role: 'user', content: 'Hi' }],
    });

    expect(result.content).toBe('Hello!');
    expect(result.usage?.totalTokens).toBe(15);
  });

  it('should stream response', async () => {
    const chunks = [
      { choices: [{ delta: { content: 'Hello' } }] },
      { choices: [{ delta: { content: ' World' } }] },
    ];

    const mockStream = new ReadableStream({
      start(controller) {
        for (const chunk of chunks) {
          controller.enqueue(`data: ${JSON.stringify(chunk)}\n\n`);
        }
        controller.close();
      },
    });

    global.fetch = vi.fn().mockResolvedValue({
      ok: true,
      body: mockStream,
    });

    const provider = new MyProvider();
    await provider.onInit(mockCtx as any);

    const received: string[] = [];
    for await (const chunk of provider.completeStream({
      model: 'test-model',
      messages: [{ role: 'user', content: 'Hi' }],
    })) {
      if (chunk.type === 'content' && chunk.content) {
        received.push(chunk.content);
      }
    }

    expect(received).toEqual(['Hello', ' World']);
  });
});
```

## 9. 相关文档

| 文档 | 说明 |
|------|------|
| `docs/AGENT-RUNTIME.md` | 整体架构设计 |
| `docs/CHANNELS.md` | Channel 实现指南 |
| `docs/PLUGIN-DEV.md` | 插件开发指南 |

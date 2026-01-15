/**
 * Anthropic Provider (OpenClaw Compatible)
 */

import { BaseModelProvider } from './base.js';
import type {
  CompletionOptions,
  CompletionResult,
  PluginContext,
  StreamChunk,
} from '../runtime/types.js';

export interface AnthropicConfig {
  apiKey?: string;
  baseUrl?: string;
  defaultModel?: string;
  maxTokens?: number;
}

export class AnthropicProvider extends BaseModelProvider {
  readonly name = 'anthropic';
  readonly defaultModel = 'claude-3-5-sonnet-20241022';
  readonly supportedModels = [
    'claude-3-5-sonnet-20241022',
    'claude-3-5-haiku-20241022',
    'claude-3-opus-20240229',
    'claude-3-haiku-20240307',
  ];

  private baseUrl = 'https://api.anthropic.com';
  private defaultMaxTokens = 4096;

  async onInit(ctx: PluginContext): Promise<void> {
    await super.onInit(ctx);
    const config = ctx.getConfig<AnthropicConfig>();
    if (config?.baseUrl) this.baseUrl = config.baseUrl;
    if (config?.apiKey) this.apiKey = config.apiKey;
    if (config?.defaultModel) this.defaultModel = config.defaultModel;
    if (config?.maxTokens) this.defaultMaxTokens = config.maxTokens;
  }

  async complete(options: CompletionOptions): Promise<CompletionResult> {
    const model = options.model || this.defaultModel;
    this.validateModel(model);

    const body: Record<string, unknown> = {
      model,
      messages: options.messages,
      max_tokens: options.maxTokens || this.defaultMaxTokens,
      temperature: options.temperature ?? 1,
    };

    if (options.topP !== undefined) body.top_p = options.topP;
    if (options.stop) body.stop_sequences = options.stop;
    if (options.tools?.length) {
      body.tools = options.tools;
      if (options.toolChoice) body.tool_choice = options.toolChoice;
    }

    const response = await this.request('/v1/messages', body) as {
      content: Array<{ type: string; text?: string; name?: string; input?: unknown }>;
      usage?: { input_tokens: number; output_tokens: number };
      stop_reason?: string;
    };

    let content = '';
    const toolCalls: CompletionResult['toolCalls'] = [];

    for (const block of response.content || []) {
      if (block.type === 'text') content += block.text || '';
      else if (block.type === 'tool_use') {
        toolCalls.push({ name: block.name || '', arguments: JSON.stringify(block.input || {}) });
      }
    }

    return {
      content,
      model,
      usage: this.parseUsage(response.usage),
      stopReason: response.stop_reason,
      toolCalls: toolCalls.length ? toolCalls : undefined,
    };
  }

  async *completeStream(options: CompletionOptions): AsyncGenerator<StreamChunk> {
    const body: Record<string, unknown> = {
      model: options.model || this.defaultModel,
      messages: options.messages,
      max_tokens: options.maxTokens || this.defaultMaxTokens,
      temperature: options.temperature ?? 1,
      stream: true,
    };
    if (options.tools?.length) body.tools = options.tools;

    const response = await fetch(`${this.baseUrl}/v1/messages`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'x-api-key': this.apiKey || '',
        'anthropic-version': '2023-06-01',
        'anthropic-dangerous-direct-browser-access': '',
      },
      body: JSON.stringify(body),
    });

    const reader = response.body?.getReader();
    if (!reader) throw new Error('No response body');
    const decoder = new TextDecoder();
    let buffer = '';

    while (true) {
      const { done, value } = await reader.read();
      if (done) break;
      buffer += decoder.decode(value, { stream: true });
      const lines = buffer.split('\n');
      buffer = lines.pop() || '';

      for (const line of lines) {
        if (line.startsWith('data: ')) {
          try {
            const chunk = JSON.parse(line.slice(6));
            if (chunk.type === 'content_block_delta') {
              if (chunk.delta.type === 'text_delta') yield { type: 'content', content: chunk.delta.text };
            } else if (chunk.type === 'message_delta') {
              yield { type: 'done', stopReason: chunk.delta.stop_reason, usage: this.parseUsage(chunk.usage) };
            }
          } catch {}
        }
      }
    }
  }

  private async request(endpoint: string, body: Record<string, unknown>): Promise<unknown> {
    const response = await fetch(`${this.baseUrl}${endpoint}`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'x-api-key': this.apiKey || '',
        'anthropic-version': '2023-06-01',
      },
      body: JSON.stringify(body),
    });
    if (!response.ok) throw new Error(`Anthropic API error: ${response.status}`);
    return response.json();
  }

  private parseUsage(usage: unknown): CompletionResult['usage'] {
    if (!usage || typeof usage !== 'object') return undefined;
    const u = usage as { input_tokens?: number; output_tokens?: number };
    return { promptTokens: u.input_tokens || 0, completionTokens: u.output_tokens || 0, totalTokens: (u.input_tokens || 0) + (u.output_tokens || 0) };
  }
}

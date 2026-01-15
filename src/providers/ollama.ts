/**
 * Ollama Provider (OpenClaw Compatible)
 */

import { BaseModelProvider } from './base.js';
import type {
  CompletionOptions,
  CompletionResult,
  EmbedOptions,
  EmbedResult,
  PluginContext,
  StreamChunk,
} from '../runtime/types.js';

export interface OllamaConfig {
  baseUrl?: string;
  defaultModel?: string;
}

export class OllamaProvider extends BaseModelProvider {
  readonly name = 'ollama';
  readonly defaultModel = 'llama3.2';
  readonly supportedModels: string[] = [];

  private baseUrl = 'http://localhost:11434';

  async onInit(ctx: PluginContext): Promise<void> {
    await super.onInit(ctx);
    const config = ctx.getConfig<OllamaConfig>();
    if (config?.baseUrl) this.baseUrl = config.baseUrl;
    if (config?.defaultModel) this.defaultModel = config.defaultModel;
  }

  async onStart(): Promise<void> {
    await super.onStart();
    try {
      const response = await fetch(`${this.baseUrl}/api/tags`);
      if (response.ok) {
        const data = await response.json() as { models?: Array<{ name: string }> };
        this.supportedModels = (data.models || []).map((m) => m.name);
      }
    } catch (err) {
      this.ctx?.logger.warn(`[ollama] Could not fetch models: ${err}`);
    }
  }

  async complete(options: CompletionOptions): Promise<CompletionResult> {
    const model = options.model || this.defaultModel;
    const response = await this.request('/api/chat', {
      model,
      messages: options.messages,
      stream: false,
      temperature: options.temperature,
      options: { num_predict: options.maxTokens, stop: options.stop },
    }) as { message?: { content: string }; eval_count?: number; prompt_eval_count?: number };

    return {
      content: response.message?.content || '',
      model,
      usage: { promptTokens: response.prompt_eval_count || 0, completionTokens: response.eval_count || 0, totalTokens: (response.prompt_eval_count || 0) + (response.eval_count || 0) },
    };
  }

  async *completeStream(options: CompletionOptions): AsyncGenerator<StreamChunk> {
    const response = await fetch(`${this.baseUrl}/api/chat`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ model: options.model || this.defaultModel, messages: options.messages, stream: true, temperature: options.temperature }),
    });

    const reader = response.body?.getReader();
    if (!reader) throw new Error('No response body');
    const decoder = new TextDecoder();

    while (true) {
      const { done, value } = await reader.read();
      if (done) break;
      const text = decoder.decode(value, { stream: true });
      for (const line of text.split('\n').filter((l) => l.trim())) {
        try {
          const data = JSON.parse(line);
          if (data.message?.content) yield { type: 'content', content: data.message.content };
          if (data.done) yield { type: 'done' };
        } catch {}
      }
    }
  }

  async embed(options: EmbedOptions): Promise<EmbedResult> {
    const model = (options.model || this.defaultModel).split(':')[0];
    const response = await this.request('/api/embeddings', { model, prompt: typeof options.input === 'string' ? options.input : options.input[0] }) as { embedding?: number[] };
    return { embeddings: [response.embedding || []], model };
  }

  private async request(endpoint: string, body: Record<string, unknown>): Promise<unknown> {
    const response = await fetch(`${this.baseUrl}${endpoint}`, { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body) });
    if (!response.ok) throw new Error(`Ollama API error: ${response.status}`);
    return response.json();
  }
}

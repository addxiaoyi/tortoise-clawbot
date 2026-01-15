/**
 * OpenAI Provider (OpenClaw Compatible)
 */

import { BaseModelProvider } from './base.js';
import type {
  CompletionOptions,
  CompletionResult,
  PluginContext,
  StreamChunk,
} from '../runtime/types.js';

export interface OpenAIConfig {
  apiKey?: string;
  baseUrl?: string;
  defaultModel?: string;
}

export class OpenAIProvider extends BaseModelProvider {
  readonly name = 'openai';
  readonly defaultModel = 'gpt-4o';
  readonly supportedModels = [
    'gpt-4o', 'gpt-4o-mini', 'gpt-4-turbo', 'gpt-4',
    'gpt-3.5-turbo', 'gpt-3.5-turbo-16k',
  ];

  private baseUrl = 'https://api.openai.com/v1';

  async onInit(ctx: PluginContext): Promise<void> {
    await super.onInit(ctx);
    const config = ctx.getConfig<OpenAIConfig>();
    if (config?.baseUrl) this.baseUrl = config.baseUrl;
    if (config?.apiKey) this.apiKey = config.apiKey;
    if (config?.defaultModel) this.defaultModel = config.defaultModel;
  }

  async complete(options: CompletionOptions): Promise<CompletionResult> {
    const model = options.model || this.defaultModel;

    const body: Record<string, unknown> = {
      model,
      messages: options.messages,
      temperature: options.temperature ?? 0.7,
    };

    if (options.maxTokens) body.max_tokens = options.maxTokens;
    if (options.topP !== undefined) body.top_p = options.topP;
    if (options.stop) body.stop = options.stop;
    if (options.tools?.length) {
      body.tools = options.tools;
      if (options.toolChoice) body.tool_choice = options.toolChoice;
    }

    const response = await this.request('/chat/completions', body) as {
      choices: Array<{ message: { content?: string; tool_calls?: Array<{ function: { name: string; arguments: string } }> }; finish_reason: string }>;
      usage: { prompt_tokens: number; completion_tokens: number; total_tokens: number };
      model: string;
    };

    const choice = response.choices[0];
    let content = choice.message.content || '';
    const toolCalls: CompletionResult['toolCalls'] = choice.message.tool_calls?.map((tc) => ({
      name: tc.function.name,
      arguments: tc.function.arguments,
    }));

    return {
      content,
      model: response.model,
      usage: this.parseUsage(response.usage),
      stopReason: choice.finish_reason,
      toolCalls,
    };
  }

  async *completeStream(options: CompletionOptions): AsyncGenerator<StreamChunk> {
    const body: Record<string, unknown> = {
      model: options.model || this.defaultModel,
      messages: options.messages,
      temperature: options.temperature ?? 0.7,
      stream: true,
    };
    if (options.tools?.length) body.tools = options.tools;

    const response = await fetch(`${this.baseUrl}/chat/completions`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', 'Authorization': `Bearer ${this.apiKey}` },
      body: JSON.stringify(body),
    });

    const reader = response.body?.getReader();
    if (!reader) throw new Error('No response body');
    const decoder = new TextDecoder();

    while (true) {
      const { done, value } = await reader.read();
      if (done) break;
      const text = decoder.decode(value, { stream: true });
      const lines = text.split('\n').filter((l) => l.trim() && l.startsWith('data: '));

      for (const line of lines) {
        if (line.includes('[DONE]')) { yield { type: 'done' }; continue; }
        try {
          const chunk = JSON.parse(line.slice(6));
          if (chunk.choices?.[0]?.delta) {
            const delta = chunk.choices[0].delta;
            if (delta.content) yield { type: 'content', content: delta.content };
            if (delta.tool_calls?.[0]) {
              const tc = delta.tool_calls[0];
              yield { type: 'tool-call', toolCall: { name: tc.function?.name || '', arguments: tc.function?.arguments || '' } };
            }
          }
        } catch {}
      }
    }
  }

  private async request(endpoint: string, body: Record<string, unknown>): Promise<unknown> {
    const response = await fetch(`${this.baseUrl}${endpoint}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', 'Authorization': `Bearer ${this.apiKey}` },
      body: JSON.stringify(body),
    });
    if (!response.ok) throw new Error(`OpenAI API error: ${response.status}`);
    return response.json();
  }
}

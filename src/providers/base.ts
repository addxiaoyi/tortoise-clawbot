/**
 * Provider System Base (OpenClaw Compatible)
 * Supports: Anthropic, OpenAI, Ollama, OpenRouter, etc.
 */

import type {
  CompletionOptions,
  CompletionResult,
  EmbedOptions,
  EmbedResult,
  ModelProvider,
  PluginContext,
  StreamChunk,
} from '../runtime/types.js';

/**
 * Base class for all model providers
 * Provides common functionality and lifecycle management
 */
export abstract class BaseModelProvider implements ModelProvider {
  abstract readonly name: string;
  abstract readonly defaultModel: string;
  abstract readonly supportedModels: string[];

  protected ctx?: PluginContext;
  protected config?: Record<string, unknown>;
  protected apiKey?: string;
  protected baseUrl?: string;

  async onInit(ctx: PluginContext): Promise<void> {
    this.ctx = ctx;
    this.config = ctx.getConfig();
    
    // Load common config
    this.apiKey = this.config?.apiKey as string || process.env[`${this.name.toUpperCase()}_API_KEY`] as string;
    this.baseUrl = this.config?.baseUrl as string || process.env[`${this.name.toUpperCase()}_BASE_URL`] as string;
    
    ctx.logger.info(`[${this.name}] Provider initialized`);
  }

  async onStart(): Promise<void> {
    if (!this.apiKey && !this.baseUrl) {
      this.ctx?.logger.warn(`[${this.name}] No API key or base URL configured`);
    }
    this.ctx?.logger.info(`[${this.name}] Provider started (default: ${this.defaultModel})`);
  }

  async onStop(): Promise<void> {
    this.ctx?.logger.info(`[${this.name}] Provider stopped`);
  }

  abstract complete(options: CompletionOptions): Promise<CompletionResult>;
  abstract completeStream(options: CompletionOptions): AsyncGenerator<StreamChunk>;
  abstract embed(options: EmbedOptions): Promise<EmbedResult>;

  /**
   * Validate model is supported
   */
  protected validateModel(model: string): void {
    if (!this.supportedModels.includes(model) && model !== this.defaultModel) {
      throw new Error(`Model "${model}" not supported by ${this.name}. Supported: ${this.supportedModels.join(', ')}`);
    }
  }

  /**
   * Parse usage from response
   */
  protected parseUsage(usage: unknown): CompletionResult['usage'] {
    if (!usage || typeof usage !== 'object') return undefined;
    const u = usage as { prompt_tokens?: number; completion_tokens?: number; total_tokens?: number };
    return {
      promptTokens: u.prompt_tokens || 0,
      completionTokens: u.completion_tokens || 0,
      totalTokens: u.total_tokens || 0,
    };
  }
}

/**
 * Provider registry for managing all providers
 */
export class ProviderRegistry {
  private providers = new Map<string, ModelProvider>();
  private defaultProvider?: string;

  register(provider: ModelProvider, setAsDefault = false): void {
    if (this.providers.has(provider.name)) {
      throw new Error(`Provider "${provider.name}" already registered`);
    }
    this.providers.set(provider.name, provider);
    
    if (setAsDefault || !this.defaultProvider) {
      this.defaultProvider = provider.name;
    }
  }

  unregister(name: string): void {
    this.providers.delete(name);
    if (this.defaultProvider === name) {
      this.defaultProvider = this.providers.keys().next().value;
    }
  }

  get(name: string): ModelProvider | undefined {
    return this.providers.get(name);
  }

  getDefault(): ModelProvider | undefined {
    return this.defaultProvider ? this.providers.get(this.defaultProvider) : undefined;
  }

  getAll(): ModelProvider[] {
    return Array.from(this.providers.values());
  }

  async startAll(): Promise<void> {
    await Promise.all(this.getAll().map((p) => p.onStart()));
  }

  async stopAll(): Promise<void> {
    await Promise.all(this.getAll().map((p) => p.onStop()));
  }
}

// ============================================
// Built-in Provider Exports
// ============================================

export { AnthropicProvider } from './anthropic.js';
export { OpenAIProvider } from './openai.js';
export { OllamaProvider } from './ollama.js';
// export { OpenRouterProvider } from './openrouter.js';
import { PluginContext, PluginLifecycle, PluginConfig } from '../types.js';

export interface MemoryConfig extends PluginConfig {
  /** 新 key 在容量满时按插入顺序驱逐最旧项；更新已有 key 不会误删其它项。 */
  maxItems?: number;
  /** 预留：存活时间（毫秒）。当前未在 get/list 中校验过期。 */
  ttl?: number;
}

/**
 * MCP-based Memory Plugin.
 * Demonstrates the new architecture.
 */
export class MemoryPlugin implements PluginLifecycle {
  private ctx!: PluginContext;
  private readonly memory = new Map<string, any>();
  private maxItems: number = 1000;
  private ttl: number = 3600 * 1000;

  public async onInit(ctx: PluginContext): Promise<void> {
    this.ctx = ctx;
    ctx.logger.info('MemoryPlugin initialized');
    
    // Initial config load
    const config = ctx.getConfig<MemoryConfig>();
    this.applyConfig(config);
  }

  public async onStart(): Promise<void> {
    this.ctx.logger.info('MemoryPlugin started');
    // In a real plugin, we might connect to Redis here
  }

  public async onStop(): Promise<void> {
    this.ctx.logger.info('MemoryPlugin stopped');
    this.memory.clear();
  }

  public async onConfigChange(newConfig: MemoryConfig): Promise<void> {
    this.ctx.logger.info('MemoryPlugin config updated');
    this.applyConfig(newConfig);
  }

  private applyConfig(config: MemoryConfig) {
    if (config.maxItems !== undefined) this.maxItems = config.maxItems;
    if (config.ttl !== undefined) this.ttl = config.ttl;
  }

  // Public API exposed via events or direct access if allowed
  public set(key: string, value: any) {
    const isUpdate = this.memory.has(key);
    // 已满时仅在为「新 key」腾位时驱逐；更新已有 key 不得删掉其它条目
    while (!isUpdate && this.memory.size >= this.maxItems) {
      const firstKey = this.memory.keys().next().value;
      if (firstKey !== undefined) this.memory.delete(firstKey);
      else break;
    }
    this.memory.set(key, value);
  }

  public get(key: string): any {
    return this.memory.get(key);
  }

  public listKeys(): string[] {
    return [...this.memory.keys()];
  }

  public removeKey(key: string): boolean {
    return this.memory.delete(key);
  }

  public clearAll(): void {
    this.memory.clear();
  }
}

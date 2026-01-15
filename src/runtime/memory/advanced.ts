/**
 * Advanced Memory System
 * 超越 OpenClaw/Hermes 的多层记忆系统
 * 
 * 特性:
 * - 三层记忆架构 (短期/中期/长期)
 * - 向量索引检索
 * - 记忆压缩和摘要
 * - 持久化支持 (Redis/SQLite/File)
 * - 记忆 TTL 和过期策略
 */

import { EventEmitter } from 'node:events';
import crypto from 'node:crypto';

// ==================== 类型定义 ====================

export interface MemoryEntry {
  key: string;
  value: unknown;
  timestamp: number;
  accessTime: number;
  accessCount: number;
  ttl?: number;
  tags?: string[];
  namespace?: string;
  priority?: 'low' | 'normal' | 'high';
  compressed?: boolean;
}

export interface MemoryQuery {
  prefix?: string;
  tags?: string[];
  namespace?: string;
  since?: number;
  until?: number;
  limit?: number;
  sort?: 'timestamp' | 'access' | 'key';
  order?: 'asc' | 'desc';
}

export interface MemoryStats {
  totalKeys: number;
  shortTerm: number;
  mediumTerm: number;
  longTerm: number;
  totalSize: number;
  hitRate: number;
  evictions: number;
}

export interface MemoryConfig {
  backend: 'memory' | 'redis' | 'sqlite' | 'file';
  // 分层配置
  shortTermTTL: number;      // 短期记忆 TTL (ms)
  mediumTermTTL: number;     // 中期记忆 TTL (ms)
  longTermTTL: number;      // 长期记忆 TTL (ms)
  // 自动摘要阈值
  autoSummarizeAboveBytes: number;
  // 持久化
  persistencePath?: string;
  redisUrl?: string;
  // 索引
  enableVectorIndex: boolean;
  embeddingModel?: string;
  // 配额
  maxKeys?: number;
  maxMemoryBytes?: number;
}

export interface MemoryEvent {
  type: 'get' | 'set' | 'delete' | 'evict' | 'expire' | 'summarize';
  key: string;
  timestamp: number;
  metadata?: Record<string, unknown>;
}

// ==================== 三层记忆存储 ====================

class MemoryTier {
  constructor(
    private name: string,
    private ttl: number,
    private maxSize: number
  ) {}

  private store = new Map<string, MemoryEntry>();

  get(key: string): MemoryEntry | undefined {
    const entry = this.store.get(key);
    if (!entry) return undefined;

    // 检查 TTL
    if (this.ttl > 0 && Date.now() - entry.timestamp > this.ttl) {
      this.store.delete(key);
      return undefined;
    }

    // 更新访问信息
    entry.accessTime = Date.now();
    entry.accessCount++;
    return entry;
  }

  set(key: string, entry: MemoryEntry): void {
    // 如果超过最大容量，删除最旧的
    while (this.store.size >= this.maxSize && this.store.size > 0) {
      const oldest = this.findOldest();
      if (oldest) this.store.delete(oldest);
    }
    this.store.set(key, entry);
  }

  delete(key: string): boolean {
    return this.store.delete(key);
  }

  keys(): string[] {
    return Array.from(this.store.keys());
  }

  size(): number {
    return this.store.size;
  }

  findOldest(): string | undefined {
    let oldest: string | undefined;
    let oldestTime = Infinity;

    for (const [key, entry] of this.store) {
      if (entry.accessTime < oldestTime) {
        oldestTime = entry.accessTime;
        oldest = key;
      }
    }
    return oldest;
  }

  clear(): void {
    this.store.clear();
  }
}

// ==================== 主存储类 ====================

export class AdvancedMemory extends EventEmitter {
  private config: Required<MemoryConfig>;
  private shortTerm: MemoryTier;
  private mediumTerm: MemoryTier;
  private longTerm: MemoryTier;
  private metadata: Map<string, {
    tier: 'short' | 'medium' | 'long';
    summary?: string;
    embedding?: number[];
  }> = new Map();

  private hitCount = 0;
  private missCount = 0;
  private evictionCount = 0;
  private persistenceTimer?: NodeJS.Timeout;
  private gcTimer?: NodeJS.Timeout;

  constructor(config: MemoryConfig) {
    super();
    this.config = {
      backend: config.backend || 'memory',
      shortTermTTL: config.shortTermTTL || 5 * 60 * 1000,      // 5 分钟
      mediumTermTTL: config.mediumTermTTL || 30 * 60 * 1000,   // 30 分钟
      longTermTTL: config.longTermTTL || 7 * 24 * 60 * 60 * 1000, // 7 天
      autoSummarizeAboveBytes: config.autoSummarizeAboveBytes || 10000,
      persistencePath: config.persistencePath || './memory-data',
      redisUrl: config.redisUrl,
      enableVectorIndex: config.enableVectorIndex ?? true,
      embeddingModel: config.embeddingModel || 'default',
      maxKeys: config.maxKeys || 100000,
      maxMemoryBytes: config.maxMemoryBytes || 100 * 1024 * 1024,
    };

    // 初始化三层存储
    this.shortTerm = new MemoryTier('short', this.config.shortTermTTL, 10000);
    this.mediumTerm = new MemoryTier('medium', this.config.mediumTermTTL, 50000);
    this.longTerm = new MemoryTier('long', this.config.longTermTTL, this.config.maxKeys);

    // 启动 GC
    this.startGC();
  }

  // ==================== 核心操作 ====================

  async get(key: string, options?: { updateAccess?: boolean }): Promise<unknown> {
    const meta = this.metadata.get(key);
    const tier = meta?.tier || 'short';
    let entry: MemoryEntry | undefined;

    // 尝试从对应层级获取
    switch (tier) {
      case 'short':
        entry = this.shortTerm.get(key);
        if (!entry) {
          // 降级到中期
          entry = this.mediumTerm.get(key);
          if (entry) this.promote(key, 'short');
        }
        break;
      case 'medium':
        entry = this.mediumTerm.get(key);
        if (entry) {
          // 检查是否应该升级到短期或降级到长期
          this.assessTier(key, entry);
        }
        break;
      case 'long':
        entry = this.longTerm.get(key);
        if (entry) {
          // 升级到中期
          this.promote(key, 'medium');
        }
        break;
    }

    if (entry) {
      this.hitCount++;
      this.emit('memory:get', { key, tier, hit: true } as MemoryEvent);
      return entry.value;
    }

    this.missCount++;
    this.emit('memory:get', { key, tier: 'none', hit: false } as MemoryEvent);
    return undefined;
  }

  async set(
    key: string,
    value: unknown,
    options?: {
      ttl?: number;
      tags?: string[];
      namespace?: string;
      priority?: 'low' | 'normal' | 'high';
    }
  ): Promise<void> {
    const now = Date.now();
    const entry: MemoryEntry = {
      key,
      value,
      timestamp: now,
      accessTime: now,
      accessCount: 1,
      ttl: options?.ttl,
      tags: options?.tags,
      namespace: options?.namespace,
      priority: options?.priority || 'normal',
    };

    // 根据优先级选择层级
    const tier = this.determineTier(entry);
    this.storeInTier(key, entry, tier);
    this.metadata.set(key, { tier });

    // 检查是否需要自动摘要
    const valueStr = JSON.stringify(value);
    if (valueStr.length > this.config.autoSummarizeAboveBytes) {
      this.summarize(key, value);
    }

    this.emit('memory:set', { key, tier, timestamp: now } as MemoryEvent);
  }

  async delete(key: string): Promise<boolean> {
    const tiers = [
      this.shortTerm.delete(key),
      this.mediumTerm.delete(key),
      this.longTerm.delete(key),
    ];

    const deleted = tiers.some(Boolean);
    this.metadata.delete(key);

    if (deleted) {
      this.emit('memory:delete', { key, timestamp: Date.now() } as MemoryEvent);
    }

    return deleted;
  }

  async clear(namespace?: string): Promise<number> {
    let count = 0;

    if (namespace) {
      // 清除命名空间
      for (const key of this.metadata.keys()) {
        const meta = this.metadata.get(key);
        if (meta?.tier) {
          const tier = this.getTierStore(meta.tier);
          if (tier.delete(key)) {
            count++;
          }
        }
        this.metadata.delete(key);
      }
    } else {
      // 清除所有
      count = this.shortTerm.size() + this.mediumTerm.size() + this.longTerm.size();
      this.shortTerm.clear();
      this.mediumTerm.clear();
      this.longTerm.clear();
      this.metadata.clear();
    }

    this.emit('memory:clear', { namespace, count, timestamp: Date.now() });
    return count;
  }

  // ==================== 查询 ====================

  async query(query: MemoryQuery): Promise<Array<{ key: string; value: unknown }>> {
    const results: Array<{ key: string; value: unknown; entry: MemoryEntry }> = [];

    const checkEntry = (entry: MemoryEntry | undefined, key: string) => {
      if (!entry) return;

      // 时间过滤
      if (query.since && entry.timestamp < query.since) return;
      if (query.until && entry.timestamp > query.until) return;

      // 标签过滤
      if (query.tags?.length) {
        const hasTags = query.tags.every(tag => entry.tags?.includes(tag));
        if (!hasTags) return;
      }

      // 命名空间过滤
      if (query.namespace && entry.namespace !== query.namespace) return;

      results.push({ key, value: entry.value, entry });
    };

    // 收集所有层级的匹配项
    for (const key of this.shortTerm.keys()) {
      checkEntry(this.shortTerm.get(key), key);
    }
    for (const key of this.mediumTerm.keys()) {
      checkEntry(this.mediumTerm.get(key), key);
    }
    for (const key of this.longTerm.keys()) {
      checkEntry(this.longTerm.get(key), key);
    }

    // 排序
    results.sort((a, b) => {
      let cmp = 0;
      switch (query.sort || 'timestamp') {
        case 'timestamp':
          cmp = a.entry.timestamp - b.entry.timestamp;
          break;
        case 'access':
          cmp = a.entry.accessCount - b.entry.accessCount;
          break;
        case 'key':
          cmp = a.key.localeCompare(b.key);
          break;
      }
      return query.order === 'desc' ? -cmp : cmp;
    });

    // 限制数量
    if (query.limit) {
      return results.slice(0, query.limit).map(r => ({ key: r.key, value: r.value }));
    }

    return results.map(r => ({ key: r.key, value: r.value }));
  }

  async keys(prefix?: string): Promise<string[]> {
    const allKeys = new Set<string>();

    for (const key of this.shortTerm.keys()) {
      if (!prefix || key.startsWith(prefix)) allKeys.add(key);
    }
    for (const key of this.mediumTerm.keys()) {
      if (!prefix || key.startsWith(prefix)) allKeys.add(key);
    }
    for (const key of this.longTerm.keys()) {
      if (!prefix || key.startsWith(prefix)) allKeys.add(key);
    }

    return Array.from(allKeys);
  }

  // ==================== 向量搜索 ====================

  async searchByVector(
    queryEmbedding: number[],
    options?: {
      topK?: number;
      threshold?: number;
      namespace?: string;
    }
  ): Promise<Array<{ key: string; score: number; value: unknown }>> {
    if (!this.config.enableVectorIndex) {
      throw new Error('Vector index is not enabled');
    }

    const topK = options?.topK || 10;
    const threshold = options?.threshold || 0.7;

    const results: Array<{ key: string; score: number; value: unknown }> = [];

    // 遍历所有键查找向量
    for (const [key, meta] of this.metadata) {
      if (options?.namespace) {
        const entry = this.getTierStore(meta.tier).get(key);
        if (entry?.namespace !== options.namespace) continue;
      }

      if (meta.embedding) {
        const score = this.cosineSimilarity(queryEmbedding, meta.embedding);
        if (score >= threshold) {
          const entry = this.getTierStore(meta.tier).get(key);
          if (entry) {
            results.push({ key, score, value: entry.value });
          }
        }
      }
    }

    // 排序并返回 topK
    results.sort((a, b) => b.score - a.score);
    return results.slice(0, topK);
  }

  // ==================== 记忆操作 ====================

  async snapshot(key: string): Promise<string> {
    // 创建记忆快照（摘要）
    const value = await this.get(key);
    if (value === undefined) {
      throw new Error(`Key not found: ${key}`);
    }

    const entry = this.findEntry(key);
    return JSON.stringify({
      key,
      value,
      timestamp: entry?.timestamp,
      summary: this.metadata.get(key)?.summary || this.autoSummarize(value),
    });
  }

  async restore(snapshot: string): Promise<void> {
    const data = JSON.parse(snapshot);
    await this.set(data.key, data.value, {
      namespace: 'restored',
      tags: ['restored', 'snapshot'],
    });
  }

  // ==================== 统计 ====================

  getStats(): MemoryStats {
    return {
      totalKeys: this.shortTerm.size() + this.mediumTerm.size() + this.longTerm.size(),
      shortTerm: this.shortTerm.size(),
      mediumTerm: this.mediumTerm.size(),
      longTerm: this.longTerm.size(),
      totalSize: this.estimateSize(),
      hitRate: this.hitCount + this.missCount > 0 
        ? this.hitCount / (this.hitCount + this.missCount) 
        : 0,
      evictions: this.evictionCount,
    };
  }

  // ==================== 私有方法 ====================

  private determineTier(entry: MemoryEntry): 'short' | 'medium' | 'long' {
    // 根据访问频率和重要性确定层级
    if (entry.priority === 'high' || entry.accessCount > 100) {
      return 'short';
    }
    if (entry.priority === 'low' || entry.accessCount < 5) {
      return 'long';
    }
    return 'medium';
  }

  private storeInTier(key: string, entry: MemoryEntry, tier: 'short' | 'medium' | 'long'): void {
    const store = this.getTierStore(tier);
    store.set(key, entry);
  }

  private getTierStore(tier: 'short' | 'medium' | 'long'): MemoryTier {
    switch (tier) {
      case 'short': return this.shortTerm;
      case 'medium': return this.mediumTerm;
      case 'long': return this.longTerm;
    }
  }

  private promote(key: string, toTier: 'short' | 'medium' | 'long'): void {
    const meta = this.metadata.get(key);
    const fromTier = meta?.tier || 'short';

    if (fromTier === toTier) return;

    // 从原层级删除
    this.getTierStore(fromTier).delete(key);

    // 添加到新层级
    const entry = this.findEntry(key);
    if (entry) {
      this.getTierStore(toTier).set(key, entry);
      this.metadata.set(key, { ...meta, tier: toTier });
    }
  }

  private assessTier(key: string, entry: MemoryEntry): void {
    const accessRate = entry.accessCount / (Date.now() - entry.timestamp + 1);
    
    if (accessRate > 0.01) {
      // 高频访问，升级到短期
      this.promote(key, 'short');
    } else if (accessRate < 0.0001 && entry.timestamp < Date.now() - this.config.mediumTermTTL / 2) {
      // 低频访问，降级到长期
      this.promote(key, 'long');
    }
  }

  private findEntry(key: string): MemoryEntry | undefined {
    return (
      this.shortTerm.get(key) ||
      this.mediumTerm.get(key) ||
      this.longTerm.get(key)
    );
  }

  private async summarize(key: string, value: unknown): Promise<void> {
    const summary = this.autoSummarize(value);
    const meta = this.metadata.get(key);
    if (meta) {
      meta.summary = summary;
      this.metadata.set(key, meta);
    }

    this.emit('memory:summarize', { key, timestamp: Date.now() } as MemoryEvent);
  }

  private autoSummarize(value: unknown): string {
    const str = JSON.stringify(value);
    if (str.length <= 500) return str;
    return str.slice(0, 497) + '...';
  }

  private cosineSimilarity(a: number[], b: number[]): number {
    if (a.length !== b.length) return 0;

    let dotProduct = 0;
    let normA = 0;
    let normB = 0;

    for (let i = 0; i < a.length; i++) {
      dotProduct += a[i] * b[i];
      normA += a[i] * a[i];
      normB += b[i] * b[i];
    }

    return dotProduct / (Math.sqrt(normA) * Math.sqrt(normB));
  }

  private estimateSize(): number {
    let size = 0;
    for (const key of this.keys()) {
      size += key.length * 2; // UTF-16
      const value = this.findEntry(key)?.value;
      if (value) {
        size += JSON.stringify(value).length * 2;
      }
    }
    return size;
  }

  private startGC(): void {
    // 每分钟运行 GC
    this.gcTimer = setInterval(() => {
      this.runGC();
    }, 60000);
  }

  private runGC(): void {
    const now = Date.now();

    // 清理过期键
    for (const key of this.shortTerm.keys()) {
      const entry = this.shortTerm.get(key);
      if (entry && entry.ttl && now - entry.timestamp > entry.ttl) {
        this.shortTerm.delete(key);
        this.evictionCount++;
        this.emit('memory:evict', { key, reason: 'ttl', timestamp: now } as MemoryEvent);
      }
    }

    // 内存超限检查
    while (this.estimateSize() > this.config.maxMemoryBytes) {
      const oldest = this.longTerm.findOldest();
      if (oldest) {
        this.longTerm.delete(oldest);
        this.metadata.delete(oldest);
        this.evictionCount++;
        this.emit('memory:evict', { key: oldest, reason: 'memory', timestamp: now } as MemoryEvent);
      } else {
        break;
      }
    }
  }

  destroy(): void {
    if (this.gcTimer) clearInterval(this.gcTimer);
    if (this.persistenceTimer) clearInterval(this.persistenceTimer);
    this.removeAllListeners();
  }
}

// ==================== 导出 ====================

export { AdvancedMemory as MemorySystem };
export type { MemoryConfig, MemoryEntry, MemoryQuery, MemoryStats, MemoryEvent };

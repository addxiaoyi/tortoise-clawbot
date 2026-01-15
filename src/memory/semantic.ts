/**
 * Semantic Memory System
 * Long-term memory with semantic search and intelligent forgetting
 */

import { randomUUID } from 'node:crypto';

// ============================================
// Memory Entry
// ============================================

export interface MemoryEntry {
  id: string;
  type: MemoryType;
  content: string;
  embedding?: number[];         // Semantic embedding vector
  importance: number;           // 0-1, higher = more important
  accessCount: number;
  lastAccessedAt: number;
  createdAt: number;
  expiresAt?: number;           // TTL
  tags: string[];
  metadata?: Record<string, unknown>;
  sessionKey?: string;          // For session-scoped memory
}

export type MemoryType = 
  | 'fact'           // Factual information
  | 'preference'     // User preferences
  | 'context'        // Conversation context
  | 'knowledge'       // Learned knowledge
  | 'task'           // Task-related memory
  | 'reflection';    // Self-reflections

// ============================================
// Forgetting Configuration
// ============================================

export interface ForgettingConfig {
  enabled: boolean;
  decayRate: number;           // Base decay rate (0-1)
  importanceThreshold: number;  // Below this, entry is a candidate for deletion
  maxEntries: number;          // Maximum entries to keep
  decayIntervalMs: number;    // How often to run decay
  preserveRecentDays: number;  // Never delete entries accessed within this many days
  preserveTypes: MemoryType[]; // Never delete these types
}

// ============================================
// Semantic Search
// ============================================

export interface SearchOptions {
  query: string;
  queryEmbedding?: number[];
  limit?: number;
  threshold?: number;          // Similarity threshold (0-1)
  types?: MemoryType[];
  tags?: string[];
  sessionKey?: string;
  includeExpired?: boolean;
}

export interface SearchResult {
  entry: MemoryEntry;
  similarity: number;          // 0-1, higher = more similar
  score: number;              // Combined relevance score
}

// ============================================
// Memory Store Interface
// ============================================

export interface MemoryStore {
  get(id: string): Promise<MemoryEntry | null>;
  set(entry: MemoryEntry): Promise<void>;
  delete(id: string): Promise<void>;
  list(sessionKey?: string): Promise<MemoryEntry[]>;
  search(options: SearchOptions): Promise<SearchResult[]>;
  count(sessionKey?: string): Promise<number>;
}

// ============================================
// In-Memory Store (for development)
// ============================================

export class InMemoryMemoryStore implements MemoryStore {
  private entries = new Map<string, MemoryEntry>();

  async get(id: string): Promise<MemoryEntry | null> {
    const entry = this.entries.get(id);
    if (!entry) return null;
    
    // Check expiration
    if (entry.expiresAt && entry.expiresAt < Date.now()) {
      await this.delete(id);
      return null;
    }
    
    return entry;
  }

  async set(entry: MemoryEntry): Promise<void> {
    this.entries.set(entry.id, entry);
  }

  async delete(id: string): Promise<void> {
    this.entries.delete(id);
  }

  async list(sessionKey?: string): Promise<MemoryEntry[]> {
    const now = Date.now();
    return Array.from(this.entries.values()).filter(e => {
      if (sessionKey && e.sessionKey !== sessionKey) return false;
      if (e.expiresAt && e.expiresAt < now) {
        this.delete(e.id); // Cleanup
        return false;
      }
      return true;
    });
  }

  async search(options: SearchOptions): Promise<SearchResult[]> {
    const all = await this.list(options.sessionKey);
    
    const results: SearchResult[] = [];
    
    for (const entry of all) {
      // Filter by types
      if (options.types?.length && !options.types.includes(entry.type)) {
        continue;
      }
      
      // Filter by tags
      if (options.tags?.length) {
        const hasTag = options.tags.some(t => entry.tags.includes(t));
        if (!hasTag) continue;
      }
      
      // Calculate similarity
      let similarity = 0;
      if (options.queryEmbedding && entry.embedding) {
        similarity = this.cosineSimilarity(options.queryEmbedding, entry.embedding);
      } else if (options.query) {
        // Fallback to simple text matching
        similarity = this.textSimilarity(options.query, entry.content);
      }
      
      // Apply threshold
      if (options.threshold && similarity < options.threshold) {
        continue;
      }
      
      // Calculate final score
      const score = similarity * (1 + entry.importance * 0.5);
      
      results.push({ entry, similarity, score });
    }
    
    // Sort by score and limit
    return results
      .sort((a, b) => b.score - a.score)
      .slice(0, options.limit || 10);
  }

  async count(sessionKey?: string): Promise<number> {
    return (await this.list(sessionKey)).length;
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
    
    return dotProduct / (Math.sqrt(normA) * Math.sqrt(normB) || 1);
  }

  private textSimilarity(query: string, text: string): number {
    const queryWords = new Set(query.toLowerCase().split(/\s+/));
    const textWords = new Set(text.toLowerCase().split(/\s+/));
    
    let matches = 0;
    for (const word of queryWords) {
      if (textWords.has(word) || text.toLowerCase().includes(word)) {
        matches++;
      }
    }
    
    return matches / queryWords.size;
  }
}

// ============================================
// Semantic Memory (Main Class)
// ============================================

export interface SemanticMemoryConfig {
  store: MemoryStore;
  forgetting?: Partial<ForgettingConfig>;
  embeddingProvider?: EmbeddingProvider;
}

export interface EmbeddingProvider {
  embed(text: string): Promise<number[]>;
}

export class SemanticMemory {
  private store: MemoryStore;
  private forgetting: ForgettingConfig;
  private embeddingProvider?: EmbeddingProvider;
  private decayTimer?: NodeJS.Timeout;

  constructor(config: SemanticMemoryConfig) {
    this.store = config.store;
    this.forgetting = {
      enabled: true,
      decayRate: 0.01,
      importanceThreshold: 0.1,
      maxEntries: 10000,
      decayIntervalMs: 3600000, // 1 hour
      preserveRecentDays: 7,
      preserveTypes: ['preference', 'fact'],
      ...config.forgetting,
    };
    this.embeddingProvider = config.embeddingProvider;

    // Start decay timer
    if (this.forgetting.enabled) {
      this.startDecayTimer();
    }
  }

  /**
   * Add a new memory entry
   */
  async remember(
    content: string,
    type: MemoryType,
    options?: {
      importance?: number;
      tags?: string[];
      metadata?: Record<string, unknown>;
      sessionKey?: string;
      ttlMs?: number;
      generateEmbedding?: boolean;
    }
  ): Promise<MemoryEntry> {
    const entry: MemoryEntry = {
      id: randomUUID(),
      type,
      content,
      importance: options?.importance ?? 0.5,
      accessCount: 0,
      lastAccessedAt: Date.now(),
      createdAt: Date.now(),
      tags: options?.tags || [],
      metadata: options?.metadata,
      sessionKey: options?.sessionKey,
      expiresAt: options?.ttlMs ? Date.now() + options.ttlMs : undefined,
    };

    // Generate embedding if provider available
    if (this.embeddingProvider && options?.generateEmbedding !== false) {
      entry.embedding = await this.embeddingProvider.embed(content);
    }

    await this.store.set(entry);
    
    return entry;
  }

  /**
   * Recall a memory by ID
   */
  async recall(id: string): Promise<MemoryEntry | null> {
    const entry = await this.store.get(id);
    
    if (entry) {
      // Update access metadata
      entry.accessCount++;
      entry.lastAccessedAt = Date.now();
      await this.store.set(entry);
    }
    
    return entry;
  }

  /**
   * Search memories semantically
   */
  async search(options: SearchOptions): Promise<SearchResult[]> {
    // Generate embedding for query if provider available
    if (this.embeddingProvider && !options.queryEmbedding) {
      options.queryEmbedding = await this.embeddingProvider.embed(options.query);
    }
    
    return this.store.search(options);
  }

  /**
   * Search by natural language query
   */
  async query(
    question: string,
    sessionKey?: string,
    limit = 5
  ): Promise<SearchResult[]> {
    return this.search({
      query: question,
      limit,
      sessionKey,
      threshold: 0.3,
    });
  }

  /**
   * Get all memories of a specific type
   */
  async getByType(type: MemoryType, sessionKey?: string): Promise<MemoryEntry[]> {
    const results = await this.store.search({
      query: '',
      types: [type],
      sessionKey,
      threshold: -1, // No similarity threshold
      limit: 1000,
    });
    
    return results.map(r => r.entry);
  }

  /**
   * Get recent memories
   */
  async getRecent(sessionKey?: string, limit = 10): Promise<MemoryEntry[]> {
    const all = await this.store.list(sessionKey);
    return all
      .sort((a, b) => b.lastAccessedAt - a.lastAccessedAt)
      .slice(0, limit);
  }

  /**
   * Update memory importance
   */
  async updateImportance(id: string, importance: number): Promise<void> {
    const entry = await this.store.get(id);
    if (entry) {
      entry.importance = Math.max(0, Math.min(1, importance));
      await this.store.set(entry);
    }
  }

  /**
   * Delete a memory
   */
  async forget(id: string): Promise<void> {
    await this.store.delete(id);
  }

  /**
   * Forget memories matching criteria
   */
  async forgetByTags(tags: string[]): Promise<number> {
    const all = await this.store.list();
    let count = 0;
    
    for (const entry of all) {
      if (tags.some(t => entry.tags.includes(t))) {
        await this.store.delete(entry.id);
        count++;
      }
    }
    
    return count;
  }

  /**
   * Clear session memories
   */
  async clearSession(sessionKey: string): Promise<number> {
    const all = await this.store.list(sessionKey);
    let count = 0;
    
    for (const entry of all) {
      await this.store.delete(entry.id);
      count++;
    }
    
    return count;
  }

  /**
   * Get memory statistics
   */
  async getStats(sessionKey?: string): Promise<{
    total: number;
    byType: Record<MemoryType, number>;
    avgImportance: number;
    oldest: number;
    newest: number;
  }> {
    const all = await this.store.list(sessionKey);
    
    const byType: Record<MemoryType, number> = {
      fact: 0,
      preference: 0,
      context: 0,
      knowledge: 0,
      task: 0,
      reflection: 0,
    };
    
    let totalImportance = 0;
    let oldest = Infinity;
    let newest = 0;
    
    for (const entry of all) {
      byType[entry.type]++;
      totalImportance += entry.importance;
      if (entry.createdAt < oldest) oldest = entry.createdAt;
      if (entry.createdAt > newest) newest = entry.createdAt;
    }
    
    return {
      total: all.length,
      byType,
      avgImportance: all.length > 0 ? totalImportance / all.length : 0,
      oldest: oldest === Infinity ? 0 : oldest,
      newest,
    };
  }

  // ============================================
  // Intelligent Forgetting
  // ============================================

  /**
   * Start the forgetting timer
   */
  private startDecayTimer(): void {
    if (this.decayTimer) {
      clearInterval(this.decayTimer);
    }
    
    this.decayTimer = setInterval(() => {
      this.runDecay();
    }, this.forgetting.decayIntervalMs);
  }

  /**
   * Run forgetting/decay process
   */
  async runDecay(): Promise<number> {
    if (!this.forgetting.enabled) return 0;
    
    const all = await this.store.list();
    const now = Date.now();
    const preserveRecentMs = this.forgetting.preserveRecentDays * 24 * 60 * 60 * 1000;
    
    // Sort by a "forgetability" score
    const candidates = all
      .filter(entry => {
        // Never forget preserved types
        if (this.forgetting.preserveTypes.includes(entry.type)) return false;
        
        // Never forget recently accessed
        if (now - entry.lastAccessedAt < preserveRecentMs) return false;
        
        // Already below threshold
        if (entry.importance < this.forgetting.importanceThreshold) return true;
        
        return false;
      })
      .sort((a, b) => {
        // Lower importance = more forgettable
        // Fewer accesses = more forgettable
        // Older = more forgettable
        const scoreA = a.importance * 0.5 + a.accessCount * 0.3 + (a.lastAccessedAt / now) * 0.2;
        const scoreB = b.importance * 0.5 + b.accessCount * 0.3 + (b.lastAccessedAt / now) * 0.2;
        return scoreA - scoreB;
      });
    
    // Remove oldest entries until under limit
    let removed = 0;
    const targetCount = this.forgetting.maxEntries;
    
    while (all.length - removed > targetCount && candidates.length > 0) {
      const toRemove = candidates.shift()!;
      await this.store.delete(toRemove.id);
      removed++;
    }
    
    // Also apply decay to remaining entries
    for (const entry of all) {
      if (removed > 0 && candidates.includes(entry)) continue;
      
      // Apply decay to importance
      const newImportance = entry.importance * (1 - this.forgetting.decayRate);
      if (newImportance < entry.importance) {
        entry.importance = newImportance;
        await this.store.set(entry);
      }
    }
    
    return removed;
  }

  /**
   * Boost memory importance after relevant use
   */
  async boostImportance(id: string, boost = 0.1): Promise<void> {
    const entry = await this.store.get(id);
    if (entry) {
      entry.importance = Math.min(1, entry.importance + boost);
      entry.accessCount++;
      entry.lastAccessedAt = Date.now();
      await this.store.set(entry);
    }
  }

  /**
   * Stop forgetting timer
   */
  destroy(): void {
    if (this.decayTimer) {
      clearInterval(this.decayTimer);
      this.decayTimer = undefined;
    }
  }
}

// ============================================
// Simple Embedding Provider (placeholder)
// ============================================

export class SimpleEmbeddingProvider implements EmbeddingProvider {
  private dimension: number;

  constructor(dimension = 1536) {
    this.dimension = dimension;
  }

  async embed(text: string): Promise<number[]> {
    // Generate a simple deterministic embedding based on text hash
    // In production, use OpenAI embeddings or similar
    const hash = this.hashString(text);
    const embedding = new Array(this.dimension);
    
    // Seed random with hash for reproducibility
    let seed = hash;
    for (let i = 0; i < this.dimension; i++) {
      seed = (seed * 1103515245 + 12345) & 0x7fffffff;
      embedding[i] = (seed % 1000) / 1000 - 0.5;
    }
    
    // Normalize
    const norm = Math.sqrt(embedding.reduce((sum, v) => sum + v * v, 0));
    return embedding.map(v => v / norm);
  }

  private hashString(str: string): number {
    let hash = 0;
    for (let i = 0; i < str.length; i++) {
      const char = str.charCodeAt(i);
      hash = ((hash << 5) - hash) + char;
      hash = hash & hash;
    }
    return Math.abs(hash);
  }
}

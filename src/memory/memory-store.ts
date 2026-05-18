/**
 * Memory Store - Neocortex-inspired memory system
 * 持久化记忆存储，支持语义搜索
 */

export interface Memory {
  id: string;
  type: 'episodic' | 'semantic' | 'procedural';
  content: string;
  embedding?: number[];
  metadata: Record<string, unknown>;
  importance: number;
  createdAt: Date;
  updatedAt: Date;
  accessCount: number;
  lastAccessed?: Date;
  tags: string[];
}

export interface MemoryQuery {
  text?: string;
  type?: Memory['type'];
  tags?: string[];
  minImportance?: number;
  limit?: number;
  offset?: number;
}

export class MemoryStore {
  private memories: Map<string, Memory> = new Map();
  private index: Map<string, Set<string>> = new Map(); // tag -> memory ids

  /**
   * 创建新记忆
   */
  async create(memory: Omit<Memory, 'id' | 'createdAt' | 'updatedAt' | 'accessCount'>): Promise<Memory> {
    const id = crypto.randomUUID();
    const now = new Date();
    
    const newMemory: Memory = {
      ...memory,
      id,
      createdAt: now,
      updatedAt: now,
      accessCount: 0,
    };

    this.memories.set(id, newMemory);
    
    // 更新索引
    for (const tag of newMemory.tags) {
      if (!this.index.has(tag)) {
        this.index.set(tag, new Set());
      }
      this.index.get(tag)!.add(id);
    }

    return newMemory;
  }

  /**
   * 获取记忆
   */
  async get(id: string): Promise<Memory | null> {
    const memory = this.memories.get(id);
    if (memory) {
      memory.accessCount++;
      memory.lastAccessed = new Date();
    }
    return memory || null;
  }

  /**
   * 查询记忆
   */
  async query(query: MemoryQuery): Promise<Memory[]> {
    let results = Array.from(this.memories.values());

    // 按类型过滤
    if (query.type) {
      results = results.filter(m => m.type === query.type);
    }

    // 按标签过滤
    if (query.tags && query.tags.length > 0) {
      results = results.filter(m => 
        query.tags!.some(tag => m.tags.includes(tag))
      );
    }

    // 按重要性过滤
    if (query.minImportance !== undefined) {
      results = results.filter(m => m.importance >= query.minImportance!);
    }

    // 文本搜索（简单实现）
    if (query.text) {
      const searchLower = query.text.toLowerCase();
      results = results.filter(m => 
        m.content.toLowerCase().includes(searchLower)
      );
    }

    // 排序：最近访问优先
    results.sort((a, b) => {
      const aTime = a.lastAccessed?.getTime() || a.createdAt.getTime();
      const bTime = b.lastAccessed?.getTime() || b.createdAt.getTime();
      return bTime - aTime;
    });

    // 分页
    const offset = query.offset || 0;
    const limit = query.limit || 20;
    return results.slice(offset, offset + limit);
  }

  /**
   * 更新记忆
   */
  async update(id: string, updates: Partial<Memory>): Promise<Memory | null> {
    const memory = this.memories.get(id);
    if (!memory) return null;

    const updated: Memory = {
      ...memory,
      ...updates,
      id: memory.id, // 防止ID被修改
      createdAt: memory.createdAt, // 防止创建时间被修改
      updatedAt: new Date(),
    };

    this.memories.set(id, updated);
    return updated;
  }

  /**
   * 删除记忆
   */
  async delete(id: string): Promise<boolean> {
    const memory = this.memories.get(id);
    if (!memory) return false;

    // 从索引中移除
    for (const tag of memory.tags) {
      this.index.get(tag)?.delete(id);
    }

    this.memories.delete(id);
    return true;
  }

  /**
   * 获取统计信息
   */
  getStats(): {
    total: number;
    byType: Record<Memory['type'], number>;
    avgImportance: number;
  } {
    const memories = Array.from(this.memories.values());
    const byType: Record<Memory['type'], number> = {
      episodic: 0,
      semantic: 0,
      procedural: 0,
    };

    for (const m of memories) {
      byType[m.type]++;
    }

    const avgImportance = memories.length > 0
      ? memories.reduce((sum, m) => sum + m.importance, 0) / memories.length
      : 0;

    return { total: memories.length, byType, avgImportance };
  }

  /**
   * 导出所有记忆
   */
  export(): Memory[] {
    return Array.from(this.memories.values());
  }

  /**
   * 导入记忆
   */
  async import(memories: Memory[]): Promise<number> {
    let count = 0;
    for (const memory of memories) {
      if (!this.memories.has(memory.id)) {
        this.memories.set(memory.id, memory);
        for (const tag of memory.tags) {
          if (!this.index.has(tag)) {
            this.index.set(tag, new Set());
          }
          this.index.get(tag)!.add(memory.id);
        }
        count++;
      }
    }
    return count;
  }
}

// 导出单例
export const memoryStore = new MemoryStore();

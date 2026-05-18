/**
 * Subconscious Loop - 自主学习循环
 * 模拟人类的潜意识处理过程
 */

import { memoryStore, Memory } from '../memory/memory-store';

export interface SubconsciousConfig {
  recallInterval: number;      // 回忆间隔(ms)
  coreMemorySize: number;      // 核心记忆数量
  learningEnabled: boolean;    // 是否启用学习
  consolidationInterval: number; // 整合间隔(ms)
}

const DEFAULT_CONFIG: SubconsciousConfig = {
  recallInterval: 30000,       // 30秒
  coreMemorySize: 5,
  learningEnabled: true,
  consolidationInterval: 300000, // 5分钟
};

/**
 * Subconscious Loop - 潜意识循环
 * 定期触发记忆召回和自主学习
 */
export class SubconsciousLoop {
  private config: SubconsciousConfig;
  private isRunning: boolean = false;
  private recallInterval: ReturnType<typeof setInterval> | null = null;
  private consolidationInterval: ReturnType<typeof setInterval> | null = null;
  private learnedPatterns: Set<string> = new Set();
  private callbacks: Set<(memory: Memory) => void> = new Set();

  constructor(config: Partial<SubconsciousConfig> = {}) {
    this.config = { ...DEFAULT_CONFIG, ...config };
  }

  /**
   * 启动潜意识循环
   */
  start(): void {
    if (this.isRunning) return;
    this.isRunning = true;

    // 启动随机记忆召回
    this.recallInterval = setInterval(() => {
      this.randomRecall();
    }, this.config.recallInterval);

    // 启动记忆整合
    if (this.config.learningEnabled) {
      this.consolidationInterval = setInterval(() => {
        this.consolidateMemories();
      }, this.config.consolidationInterval);
    }

    console.log('Subconscious loop started');
  }

  /**
   * 停止潜意识循环
   */
  stop(): void {
    this.isRunning = false;
    if (this.recallInterval) {
      clearInterval(this.recallInterval);
      this.recallInterval = null;
    }
    if (this.consolidationInterval) {
      clearInterval(this.consolidationInterval);
      this.consolidationInterval = null;
    }
    console.log('Subconscious loop stopped');
  }

  /**
   * 随机记忆召回
   * 模拟人类的随机记忆闪现
   */
  private async randomRecall(): Promise<void> {
    const memories = await memoryStore.query({
      limit: 20,
      minImportance: 0.3,
    });

    if (memories.length === 0) return;

    // 随机选择一个记忆（按重要性加权）
    const selected = this.weightedRandomSelect(memories);
    
    // 通知观察者
    for (const callback of this.callbacks) {
      callback(selected);
    }
  }

  /**
   * 加权随机选择
   */
  private weightedRandomSelect(memories: Memory[]): Memory {
    // 计算权重：重要性越高，被选中的概率越大
    const totalWeight = memories.reduce((sum, m) => sum + m.importance, 0);
    let random = Math.random() * totalWeight;

    for (const memory of memories) {
      random -= memory.importance;
      if (random <= 0) return memory;
    }

    return memories[memories.length - 1];
  }

  /**
   * 整合记忆
   * 合并相似的记忆，强化重要记忆
   */
  private async consolidateMemories(): Promise<void> {
    const memories = await memoryStore.query({ limit: 100 });
    
    if (memories.length < 2) return;

    // 找出相似的记忆
    for (let i = 0; i < memories.length; i++) {
      for (let j = i + 1; j < memories.length; j++) {
        const similarity = this.calculateSimilarity(memories[i], memories[j]);
        
        if (similarity > 0.7) {
          // 合并记忆
          await this.mergeMemories(memories[i], memories[j]);
        }
      }
    }

    // 强化频繁访问的记忆
    for (const memory of memories) {
      if (memory.accessCount > 5 && memory.importance < 0.9) {
        await memoryStore.update(memory.id, {
          importance: Math.min(1.0, memory.importance + 0.1),
        });
      }
    }
  }

  /**
   * 计算记忆相似度
   */
  private calculateSimilarity(a: Memory, b: Memory): number {
    // 文本相似度
    const textSimilarity = this.textSimilarity(a.content, b.content);
    
    // 标签重叠
    const tagOverlap = a.tags.filter(t => b.tags.includes(t)).length / 
                       Math.max(1, Math.min(a.tags.length, b.tags.length));
    
    // 时间接近度
    const timeDiff = Math.abs(a.createdAt.getTime() - b.createdAt.getTime());
    const timeSimilarity = Math.max(0, 1 - timeDiff / (24 * 60 * 60 * 1000)); // 24小时内的相似度

    return textSimilarity * 0.6 + tagOverlap * 0.3 + timeSimilarity * 0.1;
  }

  /**
   * 文本相似度（Jaccard）
   */
  private textSimilarity(text1: string, text2: string): number {
    const words1 = new Set(text1.toLowerCase().split(/\W+/).filter(w => w.length > 2));
    const words2 = new Set(text2.toLowerCase().split(/\W+/).filter(w => w.length > 2));
    
    const intersection = new Set([...words1].filter(x => words2.has(x)));
    const union = new Set([...words1, ...words2]);
    
    return union.size > 0 ? intersection.size / union.size : 0;
  }

  /**
   * 合并记忆
   */
  private async mergeMemories(a: Memory, b: Memory): Promise<void> {
    const [primary, secondary] = a.importance >= b.importance ? [a, b] : [b, a];

    // 更新主要记忆
    await memoryStore.update(primary.id, {
      importance: Math.max(primary.importance, secondary.importance),
      tags: [...new Set([...primary.tags, ...secondary.tags])],
    });

    // 删除次要记忆
    await memoryStore.delete(secondary.id);
  }

  /**
   * 注册召回回调
   */
  onRecall(callback: (memory: Memory) => void): () => void {
    this.callbacks.add(callback);
    return () => this.callbacks.delete(callback);
  }

  /**
   * 学习新模式
   */
  async learnPattern(pattern: string): Promise<void> {
    if (this.learnedPatterns.has(pattern)) return;
    
    this.learnedPatterns.add(pattern);
    
    // 创建程序记忆
    await memoryStore.create({
      type: 'procedural',
      content: `学习模式: ${pattern}`,
      metadata: { pattern },
      importance: 0.7,
      tags: ['pattern', 'learned'],
    });
  }

  /**
   * 获取学习状态
   */
  getStatus(): {
    isRunning: boolean;
    learnedPatterns: number;
    callbacks: number;
  } {
    return {
      isRunning: this.isRunning,
      learnedPatterns: this.learnedPatterns.size,
      callbacks: this.callbacks.size,
    };
  }
}

// 导出单例
export const subconsciousLoop = new SubconsciousLoop();

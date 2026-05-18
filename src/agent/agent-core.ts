/**
 * Agent Core - Subconscious-inspired autonomous agent
 * 自主学习代理核心
 */

import { memoryStore, Memory } from '../memory/memory-store';

export interface AgentConfig {
  name: string;
  model: string;
  temperature: number;
  maxTokens: number;
  systemPrompt?: string;
}

export interface AgentMessage {
  id: string;
  role: 'user' | 'assistant' | 'system';
  content: string;
  timestamp: Date;
}

export interface AgentContext {
  memories: Memory[];
  recentMessages: AgentMessage[];
  userPreferences: Record<string, unknown>;
}

export class AgentCore {
  private config: AgentConfig;
  private context: AgentContext;
  private messageHistory: AgentMessage[] = [];
  private isRunning: boolean = false;
  private learningLoop: NodeJS.Timeout | null = null;

  constructor(config: AgentConfig) {
    this.config = config;
    this.context = {
      memories: [],
      recentMessages: [],
      userPreferences: {},
    };
  }

  /**
   * 启动代理
   */
  async start(): Promise<void> {
    if (this.isRunning) return;
    this.isRunning = true;
    this.startLearningLoop();
    console.log(`Agent ${this.config.name} started`);
  }

  /**
   * 停止代理
   */
  async stop(): Promise<void> {
    this.isRunning = false;
    if (this.learningLoop) {
      clearInterval(this.learningLoop);
      this.learningLoop = null;
    }
    console.log(`Agent ${this.config.name} stopped`);
  }

  /**
   * 处理消息
   */
  async processMessage(content: string): Promise<string> {
    const message: AgentMessage = {
      id: crypto.randomUUID(),
      role: 'user',
      content,
      timestamp: new Date(),
    };

    this.messageHistory.push(message);
    this.context.recentMessages.push(message);

    // 保持最近消息在合理范围内
    if (this.context.recentMessages.length > 50) {
      this.context.recentMessages = this.context.recentMessages.slice(-50);
    }

    // 检索相关记忆
    await this.updateContext();

    // 生成响应
    const response = await this.generateResponse(content);

    const assistantMessage: AgentMessage = {
      id: crypto.randomUUID(),
      role: 'assistant',
      content: response,
      timestamp: new Date(),
    };

    this.messageHistory.push(assistantMessage);
    this.context.recentMessages.push(assistantMessage);

    // 学习新信息
    await this.learnFromInteraction(content, response);

    return response;
  }

  /**
   * 更新上下文
   */
  private async updateContext(): Promise<void> {
    // 获取最近的记忆
    const recentMemories = await memoryStore.query({
      limit: 10,
      minImportance: 0.5,
    });
    this.context.memories = recentMemories;
  }

  /**
   * 生成响应
   */
  private async generateResponse(input: string): Promise<string> {
    // 构建提示
    let prompt = this.config.systemPrompt || '';
    
    // 添加相关记忆
    if (this.context.memories.length > 0) {
      prompt += '\n\n相关记忆:\n';
      for (const memory of this.context.memories.slice(0, 3)) {
        prompt += `- ${memory.content}\n`;
      }
    }

    // 添加最近对话
    if (this.context.recentMessages.length > 0) {
      prompt += '\n\n最近对话:\n';
      for (const msg of this.context.recentMessages.slice(-6)) {
        prompt += `${msg.role}: ${msg.content}\n`;
      }
    }

    prompt += `\n用户: ${input}`;

    // 模拟AI响应（实际应调用AI API）
    return this.simulateAIResponse(input);
  }

  /**
   * 模拟AI响应（实际项目中替换为真实API调用）
   */
  private simulateAIResponse(input: string): string {
    // 这里应该调用真实的AI API
    return `[Agent ${this.config.name}] 已收到: "${input.substring(0, 50)}..."`;
  }

  /**
   * 从交互中学习
   */
  private async learnFromInteraction(input: string, response: string): Promise<void> {
    // 分析输入是否包含重要信息
    const importance = this.analyzeImportance(input);
    
    if (importance > 0.6) {
      await memoryStore.create({
        type: 'episodic',
        content: input,
        metadata: { response: response.substring(0, 100) },
        importance,
        tags: ['interaction'],
      });
    }
  }

  /**
   * 分析重要性
   */
  private analyzeImportance(text: string): number {
    // 简单的启发式分析
    let score = 0.3;
    
    // 问题标记
    if (/如何|怎么|为什么|是什么|帮助/.test(text)) score += 0.1;
    
    // 决策标记
    if (/决定|选择|应该|要|不要/.test(text)) score += 0.2;
    
    // 偏好标记
    if (/喜欢|讨厌|偏好|希望|想要/.test(text)) score += 0.3;
    
    return Math.min(score, 1.0);
  }

  /**
   * 启动学习循环
   */
  private startLearningLoop(): void {
    // 每5分钟执行一次学习
    this.learningLoop = setInterval(async () => {
      if (!this.isRunning) return;
      await this.consolidateMemories();
    }, 5 * 60 * 1000);
  }

  /**
   * 整合记忆
   */
  private async consolidateMemories(): Promise<void> {
    const memories = await memoryStore.query({ limit: 100 });
    
    // 识别可合并的相似记忆
    for (let i = 0; i < memories.length; i++) {
      for (let j = i + 1; j < memories.length; j++) {
        if (this.isSimilar(memories[i].content, memories[j].content)) {
          // 保留重要的，删除不重要的
          const toKeep = memories[i].importance >= memories[j].importance 
            ? memories[i] : memories[j];
          const toDelete = memories[i].importance >= memories[j].importance 
            ? memories[j] : memories[i];
          
          await memoryStore.update(toKeep.id, {
            importance: Math.max(toKeep.importance, toDelete.importance),
          });
          await memoryStore.delete(toDelete.id);
        }
      }
    }
  }

  /**
   * 检查相似度
   */
  private isSimilar(text1: string, text2: string): boolean {
    // 简单的词频比较
    const words1 = new Set(text1.toLowerCase().split(/\s+/));
    const words2 = new Set(text2.toLowerCase().split(/\s+/));
    
    let intersection = 0;
    for (const word of words1) {
      if (words2.has(word)) intersection++;
    }
    
    const union = words1.size + words2.size - intersection;
    return union > 0 && intersection / union > 0.5;
  }

  /**
   * 获取代理状态
   */
  getStatus(): {
    isRunning: boolean;
    messageCount: number;
    memoryCount: number;
    config: AgentConfig;
  } {
    return {
      isRunning: this.isRunning,
      messageCount: this.messageHistory.length,
      memoryCount: this.context.memories.length,
      config: this.config,
    };
  }
}

/**
 * Agent Runtime - 核心入口
 * 超越 OpenClaw/Hermes 的全功能 Agent Runtime
 */

import { AdvancedMemory } from './memory/advanced.js';
import { RBAC, TenantManager } from './security/rbac.js';
import { WebhookEventBus } from './events/webhook.js';
import { ObservabilityManager } from './observability/index.js';
import { RaftConsensus, DistributedLockManager, LoadBalancer } from './cluster/index.js';
import { ProviderRegistry } from '../providers/base.js';
import { ChannelRegistry } from '../channels/base.js';
import type { MemoryConfig } from './memory/advanced.js';
import type { ClusterConfig } from './cluster/index.js';
import crypto from 'node:crypto';

// ==================== 类型定义 ====================

export interface RuntimeConfig {
  name: string;
  version: string;
  port: number;
  host: string;
  memory: Partial<MemoryConfig>;
  cluster?: Partial<ClusterConfig>;
  observability: {
    enabled: boolean;
    logLevel: 'debug' | 'info' | 'warn' | 'error';
  };
  security: {
    rbacEnabled: boolean;
    multiTenantEnabled: boolean;
  };
}

export interface RuntimeStatus {
  status: 'starting' | 'running' | 'stopping' | 'stopped' | 'error';
  uptime: number;
  version: string;
  memory: {
    keys: number;
    hitRate: number;
  };
  cluster?: {
    leader: string | null;
    nodes: number;
    state: string;
  };
}

// ==================== Agent Runtime 核心 ====================

export class AgentRuntime {
  readonly id: string;
  readonly config: Required<RuntimeConfig>;

  private status: RuntimeStatus['status'] = 'stopped';
  private startedAt?: number;

  // 核心子系统
  readonly memory: AdvancedMemory;
  readonly rbac: RBAC;
  readonly tenantManager: TenantManager;
  readonly eventBus: WebhookEventBus;
  readonly observability: ObservabilityManager;
  readonly providers: ProviderRegistry;
  readonly channels: ChannelRegistry;
  
  // 分布式
  private cluster?: RaftConsensus;
  private locks?: DistributedLockManager;
  private loadBalancer?: LoadBalancer;

  constructor(config: RuntimeConfig) {
    this.id = crypto.randomUUID();
    this.config = {
      name: config.name || 'agent-runtime',
      version: config.version || '1.0.0',
      port: config.port || 3000,
      host: config.host || '127.0.0.1',
      memory: {
        backend: config.memory?.backend || 'memory',
        shortTermTTL: config.memory?.shortTermTTL || 5 * 60 * 1000,
        mediumTermTTL: config.memory?.mediumTermTTL || 30 * 60 * 1000,
        longTermTTL: config.memory?.longTermTTL || 7 * 24 * 60 * 60 * 1000,
        autoSummarizeAboveBytes: config.memory?.autoSummarizeAboveBytes || 10000,
        enableVectorIndex: config.memory?.enableVectorIndex ?? true,
        maxKeys: config.memory?.maxKeys || 100000,
        maxMemoryBytes: config.memory?.maxMemoryBytes || 100 * 1024 * 1024,
      },
      cluster: config.cluster,
      observability: {
        enabled: config.observability?.enabled ?? true,
        logLevel: config.observability?.logLevel || 'info',
      },
      security: {
        rbacEnabled: config.security?.rbacEnabled ?? true,
        multiTenantEnabled: config.security?.multiTenantEnabled ?? false,
      },
    };

    // 初始化子系统
    this.memory = new AdvancedMemory(this.config.memory);
    this.rbac = new RBAC();
    this.tenantManager = new TenantManager({
      maxUsers: 5,
      maxApiCalls: 1000,
      maxMemoryBytes: 100 * 1024 * 1024,
      maxSessions: 10,
      maxPlugins: 5,
      rateLimitPerMinute: 60,
    });
    this.eventBus = new WebhookEventBus();
    this.observability = new ObservabilityManager(this.config.name);
    this.providers = new ProviderRegistry();
    this.channels = new ChannelRegistry();
  }

  // ==================== 生命周期 ====================

  async start(): Promise<void> {
    if (this.status === 'running') {
      this.observability.logger.warn('Runtime already running');
      return;
    }

    this.status = 'starting';
    this.observability.logger.info(`Starting ${this.config.name} v${this.config.version}`);

    try {
      // 初始化集群
      if (this.config.cluster) {
        await this.startCluster();
      }

      // 启动观测性
      if (this.config.observability.enabled) {
        this.startObservability();
      }

      this.startedAt = Date.now();
      this.status = 'running';

      // 发布启动事件
      await this.eventBus.publish('runtime:start', {
        id: this.id,
        version: this.config.version,
        uptime: 0,
      });

      this.observability.logger.info('Runtime started successfully');
    } catch (error) {
      this.status = 'error';
      this.observability.logger.error('Failed to start runtime', error as Error);
      throw error;
    }
  }

  async stop(): Promise<void> {
    if (this.status !== 'running' && this.status !== 'error') {
      return;
    }

    this.status = 'stopping';
    this.observability.logger.info('Stopping runtime...');

    try {
      // 发布停止事件
      await this.eventBus.publish('runtime:stop', {
        id: this.id,
        uptime: this.getUptime(),
      });

      // 停止集群
      if (this.cluster) {
        this.cluster.destroy();
      }
      if (this.locks) {
        this.locks.destroy();
      }

      // 停止事件总线
      this.eventBus.destroy();

      // 停止记忆系统
      this.memory.destroy();

      this.status = 'stopped';
      this.observability.logger.info('Runtime stopped');
    } catch (error) {
      this.observability.logger.error('Error stopping runtime', error as Error);
      throw error;
    }
  }

  // ==================== 状态 ====================

  getStatus(): RuntimeStatus {
    return {
      status: this.status,
      uptime: this.getUptime(),
      version: this.config.version,
      memory: this.memory.getStats(),
      cluster: this.cluster ? {
        leader: this.cluster.getLeader()?.id || null,
        nodes: this.cluster.getAllNodes().length,
        state: this.cluster.getState().state,
      } : undefined,
    };
  }

  private getUptime(): number {
    return this.startedAt ? Date.now() - this.startedAt : 0;
  }

  // ==================== 集群 ====================

  private async startCluster(): Promise<void> {
    if (!this.config.cluster) return;

    this.cluster = new RaftConsensus({
      nodeName: this.config.name,
      host: this.config.host,
      port: this.config.port,
      seedNodes: this.config.cluster.seedNodes || [],
      heartbeatIntervalMs: this.config.cluster.heartbeatIntervalMs || 1500,
      heartbeatTimeoutMs: this.config.cluster.heartbeatTimeoutMs || 3000,
      electionTimeoutMs: this.config.cluster.electionTimeoutMs || 5000,
      replicationFactor: this.config.cluster.replicationFactor || 3,
    });

    this.locks = new DistributedLockManager(this.id);
    this.loadBalancer = new LoadBalancer();

    this.cluster.on('leader:elected', ({ nodeId }) => {
      this.observability.logger.info(`Leader elected: ${nodeId}`);
    });

    this.observability.logger.info('Cluster initialized');
  }

  // ==================== 观测性 ====================

  private startObservability(): void {
    // 记录启动指标
    this.observability.metrics.setGauge('runtime.uptime', 0);
    this.observability.metrics.setGauge('runtime.status', this.status === 'running' ? 1 : 0);

    // 定期更新指标
    setInterval(() => {
      const stats = this.memory.getStats();
      this.observability.metrics.setGauge('memory.keys', stats.totalKeys);
      this.observability.metrics.setGauge('memory.hit_rate', stats.hitRate);
      this.observability.metrics.setGauge('runtime.uptime', this.getUptime());
    }, 10000);
  }

  // ==================== 工具调用 ====================

  async invoke(
    skill: string,
    tool: string,
    args: Record<string, unknown>,
    options?: {
      timeoutMs?: number;
      userId?: string;
      sessionKey?: string;
    }
  ): Promise<unknown> {
    const op = this.observability.startOperation(`${skill}:${tool}`, {
      skill,
      tool,
      args: Object.keys(args),
    });

    try {
      // 权限检查
      if (this.config.security.rbacEnabled && options?.userId) {
        const hasPermission = await this.rbac.checkPermission(
          options.userId,
          undefined,
          'skill:invoke'
        );
        if (!hasPermission) {
          throw new Error('Permission denied');
        }
      }

      // 记录指标
      this.observability.metrics.incrementCounter('skill.invoke', 1, { skill, tool });

      // 调用技能（这里简化，实际需要从技能注册表获取）
      const result = { success: true, skill, tool, args };

      op.end();
      return result;
    } catch (error) {
      op.end(error as Error);
      this.observability.metrics.incrementCounter('skill.error', 1, { skill, tool });
      throw error;
    }
  }

  // ==================== 工具 ====================

  listTools(): Array<{ skill: string; tool: string; description: string }> {
    // 返回所有注册的工具
    return [
      // 来自 providers
      ...this.providers.getAll().map(p => ({
        skill: p.name,
        tool: 'complete',
        description: `调用 ${p.name} 模型`,
      })),
      // 来自 channels
      ...this.channels.getAll().map(c => ({
        skill: 'channel',
        tool: c.name,
        description: `通过 ${c.name} 发送消息`,
      })),
    ];
  }
}

// ==================== 导出 ====================

export { AgentRuntime };
export type { RuntimeConfig, RuntimeStatus };

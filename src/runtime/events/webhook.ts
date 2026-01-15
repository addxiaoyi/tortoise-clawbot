/**
 * Webhook Event System - 事件订阅 + Webhook + SSE
 * 超越 OpenClaw/Hermes 的事件能力
 */

import crypto from 'node:crypto';
import { EventEmitter } from 'node:events';

// ==================== 类型定义 ====================

export interface WebhookEvent {
  id: string;
  type: string;
  source: string;
  data: unknown;
  timestamp: number;
  metadata?: Record<string, unknown>;
}

export interface WebhookSubscription {
  id: string;
  name: string;
  url: string;
  events: string[];
  secret: string;
  active: boolean;
  headers?: Record<string, string>;
  retryPolicy: RetryPolicy;
  filters?: EventFilter;
  createdAt: number;
  lastTriggeredAt?: number;
  failureCount: number;
}

export interface EventFilter {
  tenantId?: string;
  userId?: string;
  channel?: string;
  tags?: string[];
  minPriority?: 'low' | 'normal' | 'high';
}

export interface RetryPolicy {
  maxRetries: number;
  initialDelayMs: number;
  maxDelayMs: number;
  backoffMultiplier: number;
}

export interface SSESubscription {
  id: string;
  clientId: string;
  events: string[];
  filters?: EventFilter;
  createdAt: number;
  lastPingAt: number;
}

export interface EventBusConfig {
  maxQueueSize: number;
  maxRetainedEvents: number;
  webhookTimeoutMs: number;
  sseHeartbeatMs: number;
}

// ==================== 事件类型常量 ====================

export const SYSTEM_EVENTS = {
  // Runtime 事件
  RUNTIME_START: 'runtime:start',
  RUNTIME_STOP: 'runtime:stop',
  RUNTIME_ERROR: 'runtime:error',

  // Memory 事件
  MEMORY_GET: 'memory:get',
  MEMORY_SET: 'memory:set',
  MEMORY_DELETE: 'memory:delete',
  MEMORY_CLEAR: 'memory:clear',

  // Skill 事件
  SKILL_INVOKE: 'skill:invoke',
  SKILL_SUCCESS: 'skill:success',
  SKILL_ERROR: 'skill:error',

  // Channel 事件
  CHANNEL_MESSAGE: 'channel:message',
  CHANNEL_CONNECT: 'channel:connect',
  CHANNEL_DISCONNECT: 'channel:disconnect',

  // Session 事件
  SESSION_CREATE: 'session:create',
  SESSION_EXPIRE: 'session:expire',

  // 安全事件
  AUTH_SUCCESS: 'auth:success',
  AUTH_FAILURE: 'auth:failure',
  PERMISSION_DENIED: 'auth:permission_denied',
} as const;

export const CHANNEL_EVENTS = {
  TELEGRAM_MESSAGE: 'channel:telegram:message',
  DISCORD_MESSAGE: 'channel:discord:message',
  SLACK_MESSAGE: 'channel:slack:message',
  WHATSAPP_MESSAGE: 'channel:whatsapp:message',
} as const;

// ==================== Webhook 事件系统 ====================

export class WebhookEventBus extends EventEmitter {
  private config: EventBusConfig;
  private webhooks = new Map<string, WebhookSubscription>();
  private sseConnections = new Map<string, SSESubscription>();
  private eventQueue: WebhookEvent[] = [];
  private processingTimer?: NodeJS.Timeout;
  private heartbeatTimer?: NodeJS.Timeout;
  private maxEventId = 0;

  constructor(config?: Partial<EventBusConfig>) {
    super();
    this.config = {
      maxQueueSize: config?.maxQueueSize || 10000,
      maxRetainedEvents: config?.maxRetainedEvents || 1000,
      webhookTimeoutMs: config?.webhookTimeoutMs || 30000,
      sseHeartbeatMs: config?.sseHeartbeatMs || 30000,
    };

    this.startProcessing();
    this.startHeartbeat();
  }

  // ==================== Webhook 管理 ====================

  async registerWebhook(data: {
    name: string;
    url: string;
    events: string[];
    secret?: string;
    headers?: Record<string, string>;
    retryPolicy?: Partial<RetryPolicy>;
    filters?: EventFilter;
  }): Promise<WebhookSubscription> {
    const webhook: WebhookSubscription = {
      id: crypto.randomUUID(),
      name: data.name,
      url: data.url,
      events: data.events,
      secret: data.secret || crypto.randomBytes(32).toString('hex'),
      active: true,
      headers: data.headers,
      retryPolicy: {
        maxRetries: data.retryPolicy?.maxRetries ?? 3,
        initialDelayMs: data.retryPolicy?.initialDelayMs ?? 1000,
        maxDelayMs: data.retryPolicy?.maxDelayMs ?? 60000,
        backoffMultiplier: data.retryPolicy?.backoffMultiplier ?? 2,
      },
      filters: data.filters,
      createdAt: Date.now(),
      failureCount: 0,
    };

    this.webhooks.set(webhook.id, webhook);
    this.emit('webhook:register', webhook);

    return webhook;
  }

  async updateWebhook(id: string, data: Partial<WebhookSubscription>): Promise<WebhookSubscription | undefined> {
    const webhook = this.webhooks.get(id);
    if (!webhook) return undefined;

    const updated = { ...webhook, ...data, id: webhook.id };
    this.webhooks.set(id, updated);
    this.emit('webhook:update', updated);

    return updated;
  }

  async deleteWebhook(id: string): Promise<boolean> {
    const deleted = this.webhooks.delete(id);
    if (deleted) {
      this.emit('webhook:delete', { id });
    }
    return deleted;
  }

  async listWebhooks(filters?: { active?: boolean; event?: string }): Promise<WebhookSubscription[]> {
    let webhooks = Array.from(this.webhooks.values());

    if (filters?.active !== undefined) {
      webhooks = webhooks.filter(w => w.active === filters.active);
    }
    if (filters?.event) {
      webhooks = webhooks.filter(w => w.events.includes(filters.event!) || w.events.includes('*'));
    }

    return webhooks;
  }

  // ==================== SSE 连接管理 ====================

  async createSSESubscription(
    clientId: string,
    events: string[],
    filters?: EventFilter
  ): Promise<SSESubscription> {
    const subscription: SSESubscription = {
      id: crypto.randomUUID(),
      clientId,
      events,
      filters,
      createdAt: Date.now(),
      lastPingAt: Date.now(),
    };

    this.sseConnections.set(subscription.id, subscription);
    this.emit('sse:connect', subscription);

    return subscription;
  }

  async closeSSESubscription(id: string): Promise<boolean> {
    const sub = this.sseConnections.get(id);
    if (!sub) return false;

    this.sseConnections.delete(id);
    this.emit('sse:disconnect', sub);

    return true;
  }

  async pingSSE(id: string): Promise<boolean> {
    const sub = this.sseConnections.get(id);
    if (!sub) return false;

    sub.lastPingAt = Date.now();
    return true;
  }

  listSSESubscriptions(): SSESubscription[] {
    return Array.from(this.sseConnections.values());
  }

  // ==================== 事件发布 ====================

  async publish(
    type: string,
    data: unknown,
    options?: {
      source?: string;
      metadata?: Record<string, unknown>;
      priority?: 'low' | 'normal' | 'high';
    }
  ): Promise<string> {
    const id = `${Date.now()}-${++this.maxEventId}`;

    const event: WebhookEvent = {
      id,
      type,
      source: options?.source || 'system',
      data,
      timestamp: Date.now(),
      metadata: {
        ...options?.metadata,
        priority: options?.priority || 'normal',
      },
    };

    // 保留事件
    this.eventQueue.push(event);
    if (this.eventQueue.length > this.config.maxRetainedEvents) {
      this.eventQueue = this.eventQueue.slice(-this.config.maxRetainedEvents);
    }

    // 异步处理
    this.scheduleProcess();

    // 通知 SSE 订阅者
    this.notifySSE(event);

    this.emit('event', event);
    this.emit(`event:${type}`, event);

    return id;
  }

  // ==================== 事件历史 ====================

  getEventHistory(options?: {
    since?: number;
    until?: number;
    type?: string;
    limit?: number;
  }): WebhookEvent[] {
    let events = [...this.eventQueue];

    if (options?.since) {
      events = events.filter(e => e.timestamp >= options.since!);
    }
    if (options?.until) {
      events = events.filter(e => e.timestamp <= options.until!);
    }
    if (options?.type) {
      events = events.filter(e => e.type === options.type);
    }

    events.sort((a, b) => b.timestamp - a.timestamp);

    if (options?.limit) {
      events = events.slice(0, options.limit);
    }

    return events;
  }

  // ==================== 私有方法 ====================

  private startProcessing(): void {
    // 批量处理 webhook
    this.processingTimer = setInterval(() => {
      this.processQueue();
    }, 1000);
  }

  private startHeartbeat(): void {
    // SSE 心跳
    this.heartbeatTimer = setInterval(() => {
      const now = Date.now();
      for (const [id, sub] of this.sseConnections) {
        // 超时断开（2 倍心跳间隔无 ping）
        if (now - sub.lastPingAt > this.config.sseHeartbeatMs * 2) {
          this.sseConnections.delete(id);
          this.emit('sse:timeout', sub);
        }
      }
    }, this.config.sseHeartbeatMs);
  }

  private scheduleProcess(): void {
    // 队列已满时立即处理
    if (this.eventQueue.length >= this.config.maxQueueSize) {
      this.processQueue();
    }
  }

  private async processQueue(): Promise<void> {
    if (this.eventQueue.length === 0) return;

    const event = this.eventQueue.shift()!;
    await this.deliverToWebhooks(event);
  }

  private async deliverToWebhooks(event: WebhookEvent): Promise<void> {
    const matchingWebhooks = this.getMatchingWebhooks(event);

    await Promise.allSettled(
      matchingWebhooks.map(webhook => this.deliverWebhook(webhook, event))
    );
  }

  private getMatchingWebhooks(event: WebhookEvent): WebhookSubscription[] {
    return Array.from(this.webhooks.values()).filter(webhook => {
      if (!webhook.active) return false;
      if (!this.matchesFilters(event, webhook.filters)) return false;
      return webhook.events.includes(event.type) || webhook.events.includes('*');
    });
  }

  private matchesFilters(event: WebhookEvent, filters?: EventFilter): boolean {
    if (!filters) return true;

    if (filters.tenantId && event.metadata?.tenantId !== filters.tenantId) return false;
    if (filters.userId && event.metadata?.userId !== filters.userId) return false;
    if (filters.channel && event.metadata?.channel !== filters.channel) return false;
    if (filters.tags?.length) {
      const eventTags = (event.metadata?.tags as string[]) || [];
      if (!filters.tags.some(t => eventTags.includes(t))) return false;
    }
    if (filters.minPriority) {
      const priorityOrder = { low: 0, normal: 1, high: 2 };
      const eventPriority = (event.metadata?.priority as string) || 'normal';
      if (priorityOrder[eventPriority as keyof typeof priorityOrder] < priorityOrder[filters.minPriority]) {
        return false;
      }
    }

    return true;
  }

  private async deliverWebhook(
    webhook: WebhookSubscription,
    event: WebhookEvent
  ): Promise<void> {
    const payload = JSON.stringify(event);
    const signature = this.signPayload(payload, webhook.secret);

    for (let attempt = 0; attempt <= webhook.retryPolicy.maxRetries; attempt++) {
      try {
        const controller = new AbortController();
        const timeout = setTimeout(() => controller.abort(), this.config.webhookTimeoutMs);

        const response = await fetch(webhook.url, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'X-Webhook-Event': event.type,
            'X-Webhook-Id': event.id,
            'X-Webhook-Signature': `sha256=${signature}`,
            'X-Webhook-Timestamp': String(event.timestamp),
            ...webhook.headers,
          },
          body: payload,
          signal: controller.signal,
        });

        clearTimeout(timeout);

        if (response.ok) {
          webhook.failureCount = 0;
          webhook.lastTriggeredAt = Date.now();
          this.emit('webhook:success', { webhook, event });
          return;
        }

        throw new Error(`HTTP ${response.status}`);
      } catch (error) {
        const isLastAttempt = attempt === webhook.retryPolicy.maxRetries;

        if (isLastAttempt) {
          webhook.failureCount++;
          this.emit('webhook:failed', { webhook, event, error, attempt });
        } else {
          // 指数退避
          const delay = Math.min(
            webhook.retryPolicy.initialDelayMs * Math.pow(webhook.retryPolicy.backoffMultiplier, attempt),
            webhook.retryPolicy.maxDelayMs
          );
          await new Promise(r => setTimeout(r, delay));
        }
      }
    }
  }

  private signPayload(payload: string, secret: string): string {
    const hmac = crypto.createHmac('sha256', secret);
    hmac.update(payload);
    return hmac.digest('hex');
  }

  private notifySSE(event: WebhookEvent): void {
    const matchingConnections = Array.from(this.sseConnections.values()).filter(sub => {
      if (!this.matchesFilters(event, sub.filters)) return false;
      return sub.events.includes(event.type) || sub.events.includes('*');
    });

    for (const sub of matchingConnections) {
      sub.lastPingAt = Date.now();
      this.emit('sse:event', { subscription: sub, event });
    }
  }

  destroy(): void {
    if (this.processingTimer) clearInterval(this.processingTimer);
    if (this.heartbeatTimer) clearInterval(this.heartbeatTimer);
    this.removeAllListeners();
  }
}

// ==================== SSE 辅助 ====================

export function createSSEHeaders(): Record<string, string> {
  return {
    'Content-Type': 'text/event-stream',
    'Cache-Control': 'no-cache',
    'Connection': 'keep-alive',
    'X-Accel-Buffering': 'no',
  };
}

export function formatSSEEvent(
  event: WebhookEvent,
  options?: { retryMs?: number; id?: string }
): string {
  const lines: string[] = [];

  if (options?.id) {
    lines.push(`id: ${options.id}`);
  }
  lines.push(`event: ${event.type}`);
  lines.push(`data: ${JSON.stringify(event)}`);

  if (options?.retryMs) {
    lines.push(`retry: ${options.retryMs}`);
  }

  lines.push('');
  return lines.join('\n');
}

export function formatSSEHeartbeat(): string {
  return ': heartbeat\n\n';
}

// ==================== 导出 ====================

export { WebhookEventBus };
export type {
  WebhookEvent, WebhookSubscription, EventFilter, RetryPolicy,
  SSESubscription, EventBusConfig
};

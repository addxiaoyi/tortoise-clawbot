/**
 * Slack Channel Adapter (OpenClaw Compatible)
 */

import { BaseChannelAdapter } from './base.js';
import type {
  ChannelCapability,
  OutboundMessage,
  PluginContext,
} from '../runtime/types.js';

export interface SlackConfig {
  botToken: string;
  signingSecret: string;
  appToken?: string;
  defaultChannel?: string;
}

export class SlackChannel extends BaseChannelAdapter {
  readonly name = 'slack';
  readonly capabilities: ChannelCapability[] = [
    'text',
    'markdown',
    'images',
    'files',
    'typing',
    'reactions',
    'reply',
    'threads',
  ];

  private config?: SlackConfig;
  private apiBase = 'https://slack.com/api';

  async onInit(ctx: PluginContext): Promise<void> {
    await super.onInit(ctx);
    this.config = ctx.getConfig<SlackConfig>();
  }

  async onStart(): Promise<void> {
    await super.onStart();
    if (!this.config?.botToken) {
      throw new Error('[slack] botToken is required');
    }
    // Verify token
    await this.apiCall('auth.test');
    this.ctx?.logger.info('[slack] Bot connected successfully');
  }

  async send(message: OutboundMessage): Promise<void> {
    this.validateMessage(message);
    const content = await this.formatForChannel(message.content);
    
    const payload: Record<string, unknown> = {
      channel: message.to,
      text: content,
    };

    if (message.options?.threadId) {
      payload.thread_ts = message.options.threadId;
    }

    await this.apiCall('chat.postMessage', payload);
  }

  async formatForChannel(content: string): Promise<string> {
    // Slack uses mrkdwn format
    return content
      .replace(/\*\*([^*]+)\*\*/g, '*$1*') // Bold
      .replace(/\*([^*]+)\*/g, '_$1_') // Italic
      .replace(/`([^`]+)`/g, '`$1`') // Code
      .replace(/```(\w+)?\n?([\s\S]*?)```/g, '```$1\n$2```'); // Code blocks
  }

  async handleUpdate(update: unknown): Promise<void> {
    const event = (update as { event?: unknown }).event;
    if (event) {
      this.ctx?.events.emit('channel:message', { channel: this.name, event });
    }
  }

  private async apiCall(method: string, payload: Record<string, unknown> = {}): Promise<unknown> {
    const response = await fetch(`${this.apiBase}/${method}`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${this.config?.botToken}`,
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(payload),
    });
    const data = await response.json() as { ok: boolean; error?: string };
    if (!data.ok) throw new Error(`Slack API error: ${data.error}`);
    return data;
  }
}

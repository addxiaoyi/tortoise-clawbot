/**
 * Discord Channel Adapter (OpenClaw Compatible)
 */

import { BaseChannelAdapter } from './base.js';
import type {
  ChannelMessage,
  ChannelCapability,
  OutboundMessage,
  PluginContext,
} from '../runtime/types.js';

export interface DiscordConfig {
  botToken: string;
  guildId?: string;
  allowedChannels?: string[];
  intents?: number;
}

export class DiscordChannel extends BaseChannelAdapter {
  readonly name = 'discord';
  readonly capabilities: ChannelCapability[] = [
    'text',
    'markdown',
    'images',
    'audio',
    'video',
    'files',
    'typing',
    'reactions',
    'reply',
    'threads',
  ];

  private config?: DiscordConfig;
  private apiBase = 'https://discord.com/api/v10';
  private socket?: WebSocket;
  private heartbeatInterval?: NodeJS.Timeout;
  private sessionId?: string;
  private sequence?: number;

  async onInit(ctx: PluginContext): Promise<void> {
    await super.onInit(ctx);
    this.config = ctx.getConfig<DiscordConfig>();
    
    if (!this.config?.botToken) {
      ctx.logger.warn('[discord] No bot token configured, channel disabled');
    }
  }

  async onStart(): Promise<void> {
    await super.onStart();
    
    if (!this.config?.botToken) {
      throw new Error('[discord] botToken is required');
    }

    // Verify bot token
    try {
      await this.apiCall('/users/@me');
      this.ctx?.logger.info('[discord] Bot connected successfully');
    } catch (err) {
      throw new Error(`[discord] Failed to connect: ${err}`);
    }
  }

  async onStop(): Promise<void> {
    if (this.socket) {
      this.socket.close();
      this.socket = undefined;
    }
    if (this.heartbeatInterval) {
      clearInterval(this.heartbeatInterval);
    }
    await super.onStop();
  }

  async send(message: OutboundMessage): Promise<void> {
    this.validateMessage(message);

    const content = await this.formatForChannel(message.content);
    
    const payload: Record<string, unknown> = {
      content,
    };

    if (message.options?.replyTo) {
      payload.message_reference = { message_id: message.options.replyTo };
    }

    await this.apiCall(`/channels/${message.to}/messages`, payload);
    this.ctx?.logger.debug(`[discord] Message sent to channel ${message.to}`);
  }

  async formatForChannel(content: string): Promise<string> {
    // Discord uses standard markdown-like formatting
    return content
      .replace(/```(\w+)?\n([\s\S]*?)```/g, '```$1\n$2```') // Code blocks
      .replace(/`([^`]+)`/g, '`$1`') // Inline code
      .replace(/\*\*([^*]+)\*\*/g, '**$1**') // Bold
      .replace(/\*([^*]+)\*/g, '*$1*') // Italic
      .replace(/__([^_]+)__/g, '__$1__') // Underline
      .replace(/~~([^~]+)~~/g, '~~$1~~'); // Strikethrough
  }

  async handleUpdate(update: unknown): Promise<void> {
    const discordPayload = update as {
      t?: string;
      d?: unknown;
      s?: number;
    };

    if (discordPayload.s) {
      this.sequence = discordPayload.s;
    }

    switch (discordPayload.t) {
      case 'MESSAGE_CREATE':
        this.handleMessageCreate(discordPayload.d);
        break;
      case 'MESSAGE_UPDATE':
        this.handleMessageUpdate(discordPayload.d);
        break;
      default:
        break;
    }
  }

  protected parseIncomingMessage(raw: unknown): ChannelMessage {
    const msg = raw as {
      id: string;
      author: { id: string; username?: string };
      channel_id: string;
      content: string;
      timestamp: string;
      guild_id?: string;
      thread?: unknown;
    };

    return {
      id: msg.id,
      channel: this.name,
      from: msg.author.id,
      content: msg.content,
      timestamp: new Date(msg.timestamp).getTime(),
      metadata: {
        channelId: msg.channel_id,
        guildId: msg.guild_id,
        userName: msg.author.username,
        isThread: !!msg.thread,
      },
    };
  }

  private handleMessageCreate(data: unknown): void {
    const msg = this.parseIncomingMessage(data);
    // Filter out bot messages if needed
    this.ctx?.events.emit('channel:message', { channel: this.name, message: msg });
  }

  private handleMessageUpdate(data: unknown): void {
    const msg = this.parseIncomingMessage(data);
    this.ctx?.events.emit('channel:message:edit', { channel: this.name, message: msg });
  }

  private async apiCall(endpoint: string, options: RequestInit = {}): Promise<unknown> {
    if (!this.config?.botToken) {
      throw new Error('Discord bot token not configured');
    }

    const response = await fetch(`${this.apiBase}${endpoint}`, {
      ...options,
      headers: {
        'Authorization': `Bot ${this.config.botToken}`,
        'Content-Type': 'application/json',
        ...options.headers,
      },
    });

    if (!response.ok) {
      const error = await response.text();
      throw new Error(`Discord API error: ${response.status} ${error}`);
    }

    if (response.status === 204) {
      return undefined;
    }

    return response.json();
  }

  /**
   * Send a DM to a user
   */
  async sendDM(userId: string, content: string): Promise<string> {
    // Create DM channel
    const dmChannel = await this.apiCall('/users/@me/channels', {
      method: 'POST',
      body: JSON.stringify({ recipient_id: userId }),
    }) as { id: string };

    // Send message
    await this.send({
      to: dmChannel.id,
      content,
      channel: this.name,
    });

    return dmChannel.id;
  }

  /**
   * Add reaction to a message
   */
  async addReaction(channelId: string, messageId: string, emoji: string): Promise<void> {
    const encodedEmoji = encodeURIComponent(emoji);
    await this.apiCall(
      `/channels/${channelId}/messages/${messageId}/reactions/${encodedEmoji}/@me`,
      { method: 'PUT' }
    );
  }
}

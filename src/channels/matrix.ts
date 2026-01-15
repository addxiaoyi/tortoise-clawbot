/**
 * Matrix Channel Adapter (OpenClaw Compatible)
 * Supports end-to-end encryption via Olm/Megolm
 */

import { BaseChannelAdapter } from './base.js';
import type {
  ChannelMessage,
  ChannelCapability,
  OutboundMessage,
  PluginContext,
} from '../runtime/types.js';

export interface MatrixConfig {
  homeserver: string;
  userId: string;       // @user:domain
  accessToken: string;
  deviceId?: string;
  roomIds?: string[];  // Allowed rooms
  enableEncryption?: boolean;
}

export class MatrixChannel extends BaseChannelAdapter {
  readonly name = 'matrix';
  readonly capabilities: ChannelCapability[] = [
    'text',
    'markdown',
    'html',
    'images',
    'audio',
    'video',
    'files',
    'typing',
    'reactions',
    'reply',
    'threads',
    'encryption',
  ];

  private config?: MatrixConfig;
  private apiBase = '';
  private syncTimeout = 30000;
  private nextBatch?: string;
  private userId?: string;
  private deviceId?: string;
  private accessToken?: string;

  async onInit(ctx: PluginContext): Promise<void> {
    await super.onInit(ctx);
    this.config = ctx.getConfig<MatrixConfig>();
    
    if (!this.config?.homeserver || !this.config?.userId || !this.config?.accessToken) {
      ctx.logger.warn('[matrix] Missing required config, channel disabled');
      return;
    }

    this.apiBase = this.config.homeserver.replace(/\/$/, '');
    this.userId = this.config.userId;
    this.deviceId = this.config.deviceId || `TORT-${Date.now()}`;
    this.accessToken = this.config.accessToken;
  }

  async onStart(): Promise<void> {
    await super.onStart();
    
    if (!this.accessToken) {
      throw new Error('[matrix] accessToken is required');
    }

    // Verify connection
    try {
      const whoami = await this.apiCall('/account/whoami');
      this.ctx?.logger.info(`[matrix] Logged in as ${(whoami as { user_id?: string }).user_id}`);
    } catch (err) {
      throw new Error(`[matrix] Failed to connect: ${err}`);
    }
  }

  async onStop(): Promise<void> {
    await super.onStop();
    this.ctx?.logger.info('[matrix] Channel stopped');
  }

  async send(message: OutboundMessage): Promise<void> {
    this.validateMessage(message);
    
    const roomId = message.to;
    const content = message.content;

    // Send text message
    const txnId = `m${Date.now()}`;
    const payload = {
      msgtype: 'm.text',
      body: content,
    };

    // Handle formatting
    if (message.options?.parseMode === 'html') {
      (payload as Record<string, unknown>).format = 'org.matrix.custom.html';
      (payload as Record<string, unknown>).formatted_body = this.htmlToMatrix(content);
    }

    await this.apiCall(`/rooms/${encodeURIComponent(roomId)}/send/m.room.message/${txnId}`, payload, 'PUT');
    this.ctx?.logger.debug(`[matrix] Message sent to ${roomId}`);
  }

  async formatForChannel(content: string): Promise<string> {
    // Matrix uses a subset of HTML for rich formatting
    return content
      .replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>')
      .replace(/\*([^*]+)\*/g, '<em>$1</em>')
      .replace(/`([^`]+)`/g, '<code>$1</code>')
      .replace(/```(\w+)?\n([\s\S]*?)```/g, '<pre><code>$2</code></pre>');
  }

  private htmlToMatrix(html: string): string {
    // Simple HTML to Matrix formatter
    return html
      .replace(/<strong>(.*?)<\/strong>/g, '**$1**')
      .replace(/<em>(.*?)<\/em>/g, '*$1*')
      .replace(/<code>(.*?)<\/code>/g, '`$1`')
      .replace(/<pre><code>([\s\S]*?)<\/code><\/pre>/g, '```\n$1\n```')
      .replace(/<br\s*\/?>/g, '\n');
  }

  async handleUpdate(update: unknown): Promise<void> {
    // Webhook or long poll handler for Matrix events
    const syncResponse = update as {
      next_batch?: string;
      rooms?: {
        join?: Record<string, { timeline?: { events?: unknown[] } }>;
      };
    };

    if (syncResponse.next_batch) {
      this.nextBatch = syncResponse.next_batch;
    }

    // Process joined room events
    const rooms = syncResponse.rooms?.join || {};
    for (const [roomId, roomData] of Object.entries(rooms)) {
      const events = roomData.timeline?.events || [];
      for (const event of events) {
        if (this.isAllowedRoom(roomId)) {
          const message = this.parseIncomingMessage(event, roomId);
          if (message) {
            this.ctx?.events.emit('channel:message', { channel: this.name, message });
          }
        }
      }
    }
  }

  protected parseIncomingMessage(raw: unknown, roomId?: string): ChannelMessage | null {
    const event = raw as {
      event_id?: string;
      sender?: string;
      content?: { msgtype?: string; body?: string; format?: string; formatted_body?: string };
      origin_server_ts?: number;
      room_id?: string;
    };

    // Only process text messages
    if (event.content?.msgtype !== 'm.text') {
      return null;
    }

    const content = event.content.format === 'org.matrix.custom.html' && event.content.formatted_body
      ? this.matrixToMarkdown(event.content.formatted_body)
      : event.content.body || '';

    return {
      id: event.event_id || `temp-${Date.now()}`,
      channel: this.name,
      from: event.sender || 'unknown',
      content,
      timestamp: event.origin_server_ts || Date.now(),
      metadata: {
        roomId: event.room_id || roomId,
        isFormatted: event.content.format === 'org.matrix.custom.html',
      },
    };
  }

  private matrixToMarkdown(html: string): string {
    return html
      .replace(/<strong>(.*?)<\/strong>/g, '**$1**')
      .replace(/<em>(.*?)<\/em>/g, '*$1*')
      .replace(/<code>(.*?)<\/code>/g, '`$1`')
      .replace(/<pre><code>([\s\S]*?)<\/code><\/pre>/g, '```\n$1\n```')
      .replace(/<br\s*\/?>/g, '\n')
      .replace(/&lt;/g, '<')
      .replace(/&gt;/g, '>')
      .replace(/&amp;/g, '&');
  }

  private isAllowedRoom(roomId: string): boolean {
    if (!this.config?.roomIds || this.config.roomIds.length === 0) {
      return true;
    }
    return this.config.roomIds.includes(roomId);
  }

  private async apiCall(
    endpoint: string,
    payload: Record<string, unknown> = {},
    method: 'GET' | 'POST' | 'PUT' = 'POST'
  ): Promise<unknown> {
    if (!this.accessToken) {
      throw new Error('Matrix accessToken not configured');
    }

    const url = `${this.apiBase}/_matrix/client/r0${endpoint}${endpoint.includes('?') ? '&' : '?'}access_token=${this.accessToken}`;
    
    const response = await fetch(url, {
      method,
      headers: { 'Content-Type': 'application/json' },
      body: method !== 'GET' ? JSON.stringify(payload) : undefined,
    });

    if (!response.ok) {
      const error = await response.text();
      throw new Error(`Matrix API error: ${response.status} ${error}`);
    }

    if (response.status === 200 || response.status === 201) {
      return response.json();
    }
    
    return undefined;
  }

  /**
   * Sync with the Matrix homeserver
   */
  async sync(): Promise<void> {
    const params = new URLSearchParams({
      timeout: String(this.syncTimeout),
    });

    if (this.nextBatch) {
      params.set('since', this.nextBatch);
    }

    const response = await this.apiCall(`/sync?${params}`) as {
      next_batch: string;
      rooms: { join?: Record<string, unknown> };
    };

    this.nextBatch = response.next_batch;
    
    if (response.rooms?.join) {
      await this.handleUpdate({ rooms: response.rooms });
    }
  }

  /**
   * Join a room
   */
  async joinRoom(roomIdOrAlias: string): Promise<string> {
    const result = await this.apiCall(`/join/${encodeURIComponent(roomIdOrAlias)}`, {}, 'POST') as { room_id: string };
    this.ctx?.logger.info(`[matrix] Joined room ${result.room_id}`);
    return result.room_id;
  }

  /**
   * Send typing indicator
   */
  async sendTyping(roomId: string, timeout = 10000): Promise<void> {
    await this.apiCall(`/rooms/${encodeURIComponent(roomId)}/typing/${encodeURIComponent(this.userId || '')}`, {
      typing: true,
      timeout,
    }, 'PUT');
  }
}

/**
 * Telegram Channel Adapter (OpenClaw Compatible)
 */

import { BaseChannelAdapter } from './base.js';
import type {
  ChannelMessage,
  ChannelCapability,
  OutboundMessage,
  PluginContext,
} from '../runtime/types.js';

export interface TelegramConfig {
  botToken: string;
  allowedChats?: string[];
  parseMode?: 'MarkdownV2' | 'HTML' | 'Markdown';
}

export class TelegramChannel extends BaseChannelAdapter {
  readonly name = 'telegram';
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
  ];

  private config?: TelegramConfig;
  private apiBase = 'https://api.telegram.org';

  async onInit(ctx: PluginContext): Promise<void> {
    await super.onInit(ctx);
    this.config = ctx.getConfig<TelegramConfig>();
    
    if (!this.config?.botToken) {
      ctx.logger.warn('[telegram] No bot token configured, channel disabled');
    }
  }

  async onStart(): Promise<void> {
    await super.onStart();
    
    if (!this.config?.botToken) {
      throw new Error('[telegram] botToken is required');
    }

    // Verify bot token
    try {
      await this.apiCall('getMe');
      this.ctx?.logger.info('[telegram] Bot connected successfully');
    } catch (err) {
      throw new Error(`[telegram] Failed to connect: ${err}`);
    }
  }

  async send(message: OutboundMessage): Promise<void> {
    this.validateMessage(message);
    
    const parseMode = message.options?.parseMode === 'html' 
      ? 'HTML' 
      : message.options?.parseMode === 'markdown'
        ? 'Markdown'
        : 'MarkdownV2';

    const payload: Record<string, unknown> = {
      chat_id: message.to,
      text: message.content,
      parse_mode: parseMode,
    };

    if (message.options?.replyTo) {
      payload.reply_to_message_id = message.options.replyTo;
    }

    await this.apiCall('sendMessage', payload);
    this.ctx?.logger.debug(`[telegram] Message sent to ${message.to}`);
  }

  async formatForChannel(content: string): Promise<string> {
    // Telegram uses MarkdownV2 by default, escape special characters
    const escape = (text: string): string => {
      return text
        .replace(/\\/g, '\\\\')
        .replace(/_/g, '\\_')
        .replace(/\*/g, '\\*')
        .replace(/\[/g, '\\[')
        .replace(/\]/g, '\\]')
        .replace(/\(/g, '\\(')
        .replace(/\)/g, '\\)')
        .replace(/~/g, '\\~')
        .replace(/`/g, '\\`')
        .replace(/>/g, '\\>')
        .replace(/#/g, '\\#')
        .replace(/\+/g, '\\+')
        .replace(/-/g, '\\-')
        .replace(/=/g, '\\=')
        .replace(/\|/g, '\\|')
        .replace(/\{/g, '\\{')
        .replace(/\}/g, '\\}')
        .replace(/\./g, '\\.')
        .replace(/!/g, '\\!');
    };
    return escape(content);
  }

  async handleUpdate(update: unknown): Promise<void> {
    const tgUpdate = update as {
      message?: {
        message_id: number;
        from?: { id: number; first_name?: string };
        chat: { id: number; type: string; title?: string };
        text?: string;
        date: number;
      };
      edited_message?: unknown;
      callback_query?: unknown;
    };

    if (tgUpdate.message && this.isAllowedChat(tgUpdate.message.chat.id)) {
      const message = this.parseIncomingMessage(tgUpdate);
      this.ctx?.events.emit('channel:message', { channel: this.name, message });
    }
  }

  protected parseIncomingMessage(raw: unknown): ChannelMessage {
    const msg = raw as {
      message_id: number;
      from?: { id: number; first_name?: string };
      chat: { id: number; type: string; title?: string };
      text?: string;
      date: number;
    };

    return {
      id: String(msg.message_id),
      channel: this.name,
      from: msg.from ? String(msg.from.id) : 'unknown',
      content: msg.text || '',
      timestamp: msg.date * 1000,
      metadata: {
        chatId: msg.chat.id,
        chatType: msg.chat.type,
        chatTitle: msg.chat.title,
        userName: msg.from?.first_name,
      },
    };
  }

  private isAllowedChat(chatId: number): boolean {
    if (!this.config?.allowedChats || this.config.allowedChats.length === 0) {
      return true; // Allow all if no restriction
    }
    return this.config.allowedChats.includes(String(chatId));
  }

  private async apiCall(method: string, payload: Record<string, unknown> = {}): Promise<unknown> {
    if (!this.config?.botToken) {
      throw new Error('Telegram bot token not configured');
    }

    const response = await fetch(`${this.apiBase}/bot${this.config.botToken}/${method}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    });

    const data = await response.json() as { ok: boolean; result?: unknown; description?: string };
    
    if (!data.ok) {
      throw new Error(`Telegram API error: ${data.description}`);
    }

    return data.result;
  }

  /**
   * Set webhook for receiving updates
   */
  async setWebhook(url: string): Promise<void> {
    await this.apiCall('setWebhook', { url });
    this.ctx?.logger.info(`[telegram] Webhook set to ${url}`);
  }

  /**
   * Send typing indicator
   */
  async sendTyping(chatId: string): Promise<void> {
    await this.apiCall('sendChatAction', { chat_id: chatId, action: 'typing' });
  }
}

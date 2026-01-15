/**
 * WhatsApp Channel Adapter (OpenClaw Compatible)
 */

import { BaseChannelAdapter } from './base.js';
import type {
  ChannelCapability,
  OutboundMessage,
  PluginContext,
} from '../runtime/types.js';

export interface WhatsAppConfig {
  phoneNumberId: string;
  accessToken: string;
  webhookVerifyToken?: string;
}

export class WhatsAppChannel extends BaseChannelAdapter {
  readonly name = 'whatsapp';
  readonly capabilities: ChannelCapability[] = [
    'text',
    'images',
    'audio',
    'video',
    'files',
  ];

  private config?: WhatsAppConfig;
  private apiBase = 'https://graph.facebook.com/v18.0';

  async onInit(ctx: PluginContext): Promise<void> {
    await super.onInit(ctx);
    this.config = ctx.getConfig<WhatsAppConfig>();
  }

  async onStart(): Promise<void> {
    await super.onStart();
    if (!this.config?.phoneNumberId || !this.config?.accessToken) {
      throw new Error('[whatsapp] phoneNumberId and accessToken are required');
    }
    this.ctx?.logger.info('[whatsapp] Channel ready');
  }

  async send(message: OutboundMessage): Promise<void> {
    this.validateMessage(message);
    
    const payload = {
      messaging_product: 'whatsapp',
      to: message.to.replace(/\D/g, ''), // Remove non-digits
      type: 'text',
      text: { body: message.content },
    };

    await this.apiCall('/messages', payload);
  }

  async formatForChannel(content: string): Promise<string> {
    // WhatsApp has limited formatting
    return content
      .replace(/\*\*([^*]+)\*\*/g, '*$1*') // Bold
      .replace(/__([^_]+)__/g, '_$1_'); // Italic
  }

  async handleUpdate(update: unknown): Promise<void> {
    const entry = (update as { entry?: Array<{ changes?: Array<{ value?: { messages?: unknown[] } }> }> }).entry?.[0];
    const messages = entry?.changes?.[0]?.value?.messages;
    if (messages?.length) {
      this.ctx?.events.emit('channel:message', { channel: this.name, messages });
    }
  }

  private async apiCall(endpoint: string, payload: Record<string, unknown>): Promise<unknown> {
    const response = await fetch(`${this.apiBase}/${this.config?.phoneNumberId}${endpoint}`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${this.config?.accessToken}`,
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(payload),
    });
    const data = await response.json() as { error?: { message: string } };
    if (data.error) throw new Error(`WhatsApp API error: ${data.error.message}`);
    return data;
  }
}

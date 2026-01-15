/**
 * Signal Channel Adapter (OpenClaw Compatible)
 * Privacy-focused messaging with end-to-end encryption
 */

import { BaseChannelAdapter } from './base.js';
import type {
  ChannelMessage,
  ChannelCapability,
  OutboundMessage,
  PluginContext,
} from '../runtime/types.js';

export interface SignalConfig {
  // Signal CLI path or API endpoint
  signalCliPath?: string;
  signalApiUrl?: string;  // For signal-web or signal-api
  phoneNumber: string;     // Bot phone number
  groups?: string[];      // Allowed group IDs
  contacts?: string[];     // Allowed contact numbers
}

export class SignalChannel extends BaseChannelAdapter {
  readonly name = 'signal';
  readonly capabilities: ChannelCapability[] = [
    'text',
    'images',
    'audio',
    'video',
    'files',
    'typing',  // Via typing indicators if supported
    'reactions',
    'reply',
  ];

  private config?: SignalConfig;
  private messageQueue: Array<{
    recipient: string;
    message: string;
    timestamp: number;
  }> = [];
  private queueTimer?: NodeJS.Timeout;

  async onInit(ctx: PluginContext): Promise<void> {
    await super.onInit(ctx);
    this.config = ctx.getConfig<SignalConfig>();
    
    if (!this.config?.phoneNumber) {
      ctx.logger.warn('[signal] Phone number not configured, channel disabled');
      return;
    }
  }

  async onStart(): Promise<void> {
    await super.onStart();
    
    // Verify Signal CLI/API connection
    await this.verifyConnection();
    
    // Start message processing queue
    this.startQueueProcessor();
    
    this.ctx?.logger.info('[signal] Signal channel started');
  }

  async onStop(): Promise<void> {
    if (this.queueTimer) {
      clearInterval(this.queueTimer);
    }
    await super.onStop();
  }

  private async verifyConnection(): Promise<void> {
    const hasSignalCli = !!this.config?.signalCliPath;
    const hasApi = !!this.config?.signalApiUrl;
    
    if (!hasSignalCli && !hasApi) {
      throw new Error('[signal] Either signalCliPath or signalApiUrl must be configured');
    }

    if (hasApi) {
      // Test API connection
      try {
        const response = await fetch(`${this.config.signalApiUrl}/v1/health`, {
          method: 'GET',
          headers: {
            'Authorization': `Bearer ${this.config?.phoneNumber}`,
          },
        });
        
        if (!response.ok) {
          throw new Error(`Signal API returned ${response.status}`);
        }
      } catch (err) {
        throw new Error(`[signal] Failed to connect to Signal API: ${err}`);
      }
    }

    this.ctx?.logger.info('[signal] Signal connection verified');
  }

  async send(message: OutboundMessage): Promise<void> {
    this.validateMessage(message);
    
    const recipient = message.to;
    const content = message.content;

    // Validate recipient
    if (!this.isAllowedRecipient(recipient)) {
      throw new Error(`[signal] Recipient ${recipient} is not in allowed list`);
    }

    // Queue the message
    this.messageQueue.push({
      recipient,
      message: content,
      timestamp: Date.now(),
    });

    this.ctx?.logger.debug(`[signal] Message queued for ${recipient}`);
  }

  private startQueueProcessor(): void {
    // Process queue every second
    this.queueTimer = setInterval(async () => {
      while (this.messageQueue.length > 0) {
        const item = this.messageQueue.shift();
        if (item) {
          try {
            await this.sendMessage(item.recipient, item.message);
          } catch (err) {
            this.ctx?.logger.error(`[signal] Failed to send message: ${err}`);
            // Re-queue with retry limit
          }
        }
      }
    }, 1000);
  }

  private async sendMessage(recipient: string, message: string): Promise<void> {
    if (this.config?.signalApiUrl) {
      await this.sendViaApi(recipient, message);
    } else if (this.config?.signalCliPath) {
      await this.sendViaCli(recipient, message);
    }
  }

  private async sendViaApi(recipient: string, message: string): Promise<void> {
    if (!this.config?.signalApiUrl) return;

    const response = await fetch(`${this.config.signalApiUrl}/v1/messages`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${this.config.phoneNumber}`,
      },
      body: JSON.stringify({
        recipient,
        message,
        timestamp: Date.now(),
      }),
    });

    if (!response.ok) {
      throw new Error(`Signal API error: ${response.status}`);
    }
  }

  private async sendViaCli(recipient: string, message: string): Promise<void> {
    if (!this.config?.signalCliPath) return;

    // Escape message for shell
    const escapedMessage = message.replace(/'/g, "'\\''");
    
    // Build Signal CLI command
    const isGroup = recipient.startsWith('group-');
    const recipientFlag = isGroup ? '-g' : '-n';
    
    const command = `${this.config.signalCliPath} send ${recipientFlag} ${recipient} -m '${escapedMessage}'`;

    const { exec } = await import('node:child_process');
    
    return new Promise((resolve, reject) => {
      exec(command, (error, stdout, stderr) => {
        if (error) {
          reject(new Error(`Signal CLI error: ${stderr || error.message}`));
          return;
        }
        resolve();
      });
    });
  }

  async formatForChannel(content: string): Promise<string> {
    // Signal uses standard text formatting
    // Bold: *text*
    // Italic: _text_
    // Code: `code`
    // Links: auto-detected
    return content;
  }

  async handleUpdate(update: unknown): Promise<void> {
    // Webhook handler for incoming messages
    const signalData = update as {
      envelope?: {
        source?: string;
        sourceNumber?: string;
        sourceUuid?: string;
        sourceName?: string;
        sourceDevice?: number;
        timestamp?: number;
        message?: {
          body?: string;
          attachments?: unknown[];
          groupInfo?: {
            groupId?: string;
          };
        };
      };
    };

    if (signalData.envelope) {
      const env = signalData.envelope;
      const sender = env.sourceNumber || env.source || env.sourceUuid || 'unknown';
      
      // Check if allowed
      if (!this.isAllowedRecipient(sender)) {
        this.ctx?.logger.debug(`[signal] Ignored message from ${sender}`);
        return;
      }

      const isGroup = !!env.message?.groupInfo;
      const recipient = isGroup 
        ? `group-${env.message.groupInfo?.groupId}`
        : sender;

      const message: ChannelMessage = {
        id: `signal-${env.timestamp || Date.now()}`,
        channel: this.name,
        from: sender,
        content: env.message?.body || '',
        timestamp: env.timestamp || Date.now(),
        metadata: {
          senderName: env.sourceName,
          sourceDevice: env.sourceDevice,
          isGroup,
          groupId: env.message?.groupInfo?.groupId,
          hasAttachments: (env.message?.attachments?.length || 0) > 0,
        },
      };

      this.ctx?.events.emit('channel:message', { channel: this.name, message });
    }
  }

  protected parseIncomingMessage(raw: unknown): ChannelMessage {
    return this.parseIncomingMessage(raw);
  }

  private isAllowedRecipient(recipient: string): boolean {
    // Check contacts whitelist
    if (this.config?.contacts?.length) {
      if (!this.config.contacts.some(c => 
        recipient.includes(c) || c.includes(recipient)
      )) {
        return false;
      }
    }

    // Check groups whitelist
    if (recipient.startsWith('group-')) {
      const groupId = recipient.substring(6);
      if (this.config?.groups?.length) {
        if (!this.config.groups.includes(groupId)) {
          return false;
        }
      }
    }

    return true;
  }

  /**
   * Send typing indicator
   */
  async sendTyping(recipient: string): Promise<void> {
    if (!this.config?.signalApiUrl) return;

    await fetch(`${this.config.signalApiUrl}/v1/typing`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${this.config.phoneNumber}`,
      },
      body: JSON.stringify({
        recipient,
        typing: true,
      }),
    });
  }

  /**
   * Leave a group
   */
  async leaveGroup(groupId: string): Promise<void> {
    if (this.config?.signalApiUrl) {
      await fetch(`${this.config.signalApiUrl}/v1/groups/${groupId}/leave`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${this.config.phoneNumber}`,
        },
      });
    }
  }

  /**
   * List groups the bot is part of
   */
  async listGroups(): Promise<Array<{ id: string; name: string; members: number }>> {
    if (this.config?.signalApiUrl) {
      const response = await fetch(`${this.config.signalApiUrl}/v1/groups`, {
        headers: {
          'Authorization': `Bearer ${this.config.phoneNumber}`,
        },
      });
      
      const groups = await response.json() as Array<{
        id: string;
        name: string;
        members: { number: string }[];
      }>;
      
      return groups.map(g => ({
        id: g.id,
        name: g.name,
        members: g.members.length,
      }));
    }
    
    return [];
  }
}

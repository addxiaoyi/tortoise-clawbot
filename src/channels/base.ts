/**
 * Channel System Base (OpenClaw Compatible)
 * Supports: Telegram, Discord, Slack, WhatsApp, Signal, iMessage
 */

import type {
  ChannelAdapter,
  ChannelCapability,
  ChannelMessage,
  OutboundMessage,
  PluginContext,
} from '../runtime/types.js';

/**
 * Base class for all channel adapters
 * Provides common functionality and lifecycle management
 */
export abstract class BaseChannelAdapter implements ChannelAdapter {
  abstract readonly name: string;
  abstract readonly capabilities: ChannelCapability[];
  
  protected ctx?: PluginContext;
  protected initialized = false;
  protected started = false;

  async onInit(ctx: PluginContext): Promise<void> {
    this.ctx = ctx;
    this.initialized = true;
    ctx.logger.info(`[${this.name}] Channel initialized`);
  }

  async onStart(): Promise<void> {
    if (!this.initialized) {
      throw new Error(`[${this.name}] Must call onInit before onStart`);
    }
    this.started = true;
    this.ctx?.logger.info(`[${this.name}] Channel started`);
  }

  async onStop(): Promise<void> {
    this.started = false;
    this.ctx?.logger.info(`[${this.name}] Channel stopped`);
  }

  async onConfigChange?(newConfig: Record<string, unknown>): Promise<void> {
    this.ctx?.logger.info(`[${this.name}] Config changed, restarting...`);
    await this.onStop();
    await this.onStart();
  }

  abstract send(message: OutboundMessage): Promise<void>;
  abstract formatForChannel(content: string): Promise<string>;

  /**
   * Check if channel supports a specific capability
   */
  hasCapability(cap: ChannelCapability): boolean {
    return this.capabilities.includes(cap);
  }

  /**
   * Validate message before sending
   */
  protected validateMessage(message: OutboundMessage): void {
    if (!message.to) {
      throw new Error('Message must have a recipient (to)');
    }
    if (!message.content) {
      throw new Error('Message must have content');
    }
  }

  /**
   * Parse incoming message from channel-specific format to unified format
   */
  protected abstract parseIncomingMessage(raw: unknown): ChannelMessage;

  /**
   * Send typing indicator
   */
  protected async sendTyping?(to: string): Promise<void>;

  /**
   * Handle incoming webhook/update
   */
  abstract handleUpdate(update: unknown): Promise<void>;
}

/**
 * Channel registry for managing all channel adapters
 */
export class ChannelRegistry {
  private channels = new Map<string, ChannelAdapter>();

  register(channel: ChannelAdapter): void {
    if (this.channels.has(channel.name)) {
      throw new Error(`Channel "${channel.name}" already registered`);
    }
    this.channels.set(channel.name, channel);
  }

  unregister(name: string): void {
    this.channels.delete(name);
  }

  get(name: string): ChannelAdapter | undefined {
    return this.channels.get(name);
  }

  getAll(): ChannelAdapter[] {
    return Array.from(this.channels.values());
  }

  getByCapability(cap: ChannelCapability): ChannelAdapter[] {
    return this.getAll().filter((ch) => ch.hasCapability(cap));
  }

  async startAll(): Promise<void> {
    await Promise.all(this.getAll().map((ch) => ch.onStart()));
  }

  async stopAll(): Promise<void> {
    await Promise.all(this.getAll().map((ch) => ch.onStop()));
  }
}

// ============================================
// Built-in Channel Implementations
// ============================================

export { TelegramChannel } from './telegram.js';
export { DiscordChannel } from './discord.js';
export { SlackChannel } from './slack.js';
export { WhatsAppChannel } from './whatsapp.js';
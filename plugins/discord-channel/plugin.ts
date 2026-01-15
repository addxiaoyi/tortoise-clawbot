/**
 * Discord Channel Plugin
 * 
 * Integrates Discord with Tortoise Agent Mesh
 */

import { ChannelPlugin, ChannelMessage, ChannelConfig } from '@tortoise/sdk';

export interface DiscordConfig extends ChannelConfig {
  token: string;
  guildId?: string;
  channels?: string[];
}

export class DiscordChannel implements ChannelPlugin {
  readonly manifest = {
    id: 'discord-channel',
    name: 'Discord Channel',
    version: '1.0.0',
    type: 'channel' as const,
  };
  
  private config?: DiscordConfig;
  private client?: any; // Discord.js client
  private messageHandlers: ((msg: ChannelMessage) => void)[] = [];
  
  async onLoad(): Promise<void> {
    console.log('Discord plugin loaded');
  }
  
  async onEnable(): Promise<void> {
    if (!this.config?.token) {
      throw new Error('Discord token not configured');
    }
    
    // Initialize Discord client
    // const { Client, GatewayIntentBits } = require('discord.js');
    // this.client = new Client({
    //   intents: [GatewayIntentBits.Guilds, GatewayIntentBits.GuildMessages]
    // });
    
    // this.client.on('messageCreate', (msg) => {
    //   if (msg.author.bot) return;
    //   this.handleMessage(msg);
    // });
    
    // await this.client.login(this.config.token);
    console.log('Discord plugin enabled');
  }
  
  async onDisable(): Promise<void> {
    if (this.client) {
      // await this.client.destroy();
      this.client = null;
    }
    console.log('Discord plugin disabled');
  }
  
  async configure(config: DiscordConfig): Promise<void> {
    this.config = config;
  }
  
  async sendMessage(channel: string, content: string): Promise<void> {
    // Find channel and send message
    // const ch = this.client.channels.cache.get(channel);
    // if (ch?.isTextBased()) {
    //   await ch.send(content);
    // }
    console.log(`Discord send to ${channel}: ${content}`);
  }
  
  async sendReply(channel: string, originalMessageId: string, content: string): Promise<void> {
    // Send reply to original message
    console.log(`Discord reply to ${originalMessageId}: ${content}`);
  }
  
  onMessage(handler: (msg: ChannelMessage) => void): void {
    this.messageHandlers.push(handler);
  }
  
  private handleMessage(msg: any): void {
    const message: ChannelMessage = {
      id: msg.id,
      channel: msg.channelId,
      sender: {
        id: msg.author.id,
        name: msg.author.username,
        avatar: msg.author.displayAvatarURL(),
      },
      content: msg.content,
      timestamp: msg.createdAt.toISOString(),
      attachments: msg.attachments.map((a: any) => ({
        id: a.id,
        name: a.name,
        url: a.url,
      })),
    };
    
    for (const handler of this.messageHandlers) {
      handler(message);
    }
  }
  
  async listChannels(): Promise<{ id: string; name: string }[]> {
    // Return list of available channels
    return [
      { id: '123456789', name: 'general' },
      { id: '987654321', name: 'ai-talk' },
    ];
  }
}

export default new DiscordChannel();

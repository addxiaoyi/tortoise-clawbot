
import { PluginContext, SkillPlugin, SkillDefinition } from '../types';
import { SlackService } from './service';
import { SlackConfig } from './types';

export class SlackPlugin implements SkillPlugin {
  private context?: PluginContext;
  private service?: SlackService;

  public async onInit(ctx: PluginContext): Promise<void> {
    this.context = ctx;
    ctx.logger.info('Slack Plugin initialized');
  }

  public async onStart(): Promise<void> {
    if (!this.context) throw new Error('SlackPlugin not initialized');

    const config = this.context.getConfig<SlackConfig>();
    if (!config || !config.token) {
      this.context.logger.warn('Slack token not configured');
      return;
    }

    this.service = new SlackService(config, this.context.logger);

    const isValid = await this.service.checkAuth();
    if (isValid) {
      this.context.logger.info('Slack auth status: valid');
    } else {
      this.context.logger.warn('Slack auth status: invalid or not logged in');
    }
  }

  public getService(): SlackService | undefined {
    return this.service;
  }

  public getSkillDefinition(): SkillDefinition {
    return {
      name: 'slack',
      version: '1.0.0',
      description: 'Slack integration for messaging',
      tools: [
        {
          name: 'send_message',
          description: 'Send a message to a Slack channel',
          parameters: [
            {
              name: 'channel',
              type: 'string',
              description: 'Channel ID or name',
              required: true
            },
            {
              name: 'text',
              type: 'string',
              description: 'Message text',
              required: true
            }
          ],
          execute: async (args: Record<string, any>, _context: PluginContext) => {
            if (!this.service) {
              throw new Error('Slack service not available');
            }
            return this.service.sendMessage({
              channel: args.channel,
              text: args.text
            });
          }
        }
      ]
    };
  }
}

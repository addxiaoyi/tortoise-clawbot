
import { PluginContext, SkillPlugin, SkillDefinition } from '../types';
import { NotionService } from './service';
import { NotionConfig } from './types';

export class NotionPlugin implements SkillPlugin {
  private context?: PluginContext;
  private service?: NotionService;

  public async onInit(ctx: PluginContext): Promise<void> {
    this.context = ctx;
    ctx.logger.info('Notion Plugin initialized');
  }

  public async onStart(): Promise<void> {
    if (!this.context) throw new Error('NotionPlugin not initialized');

    const config = this.context.getConfig<NotionConfig>();
    if (!config || !config.apiKey) {
      this.context.logger.warn('Notion API key not configured');
      return;
    }

    this.service = new NotionService(config, this.context.logger);

    const isValid = await this.service.checkAuth();
    if (isValid) {
      this.context.logger.info('Notion auth status: valid');
    } else {
      this.context.logger.warn('Notion auth status: invalid or not logged in');
    }
  }

  public getService(): NotionService | undefined {
    return this.service;
  }

  public getSkillDefinition(): SkillDefinition {
    return {
      name: 'notion',
      version: '1.0.0',
      description: 'Notion integration for managing pages and databases',
      tools: [
        {
          name: 'search',
          description: 'Search for pages or databases in Notion',
          parameters: [
            {
              name: 'query',
              type: 'string',
              description: 'Search query',
              required: true
            }
          ],
          execute: async (args: Record<string, any>, _context: PluginContext) => {
            if (!this.service) throw new Error('Notion service not available');
            return this.service.search(args.query);
          }
        },
        {
          name: 'get_page',
          description: 'Retrieve a Notion page by ID',
          parameters: [
            {
              name: 'pageId',
              type: 'string',
              description: 'The ID of the page to retrieve',
              required: true
            }
          ],
          execute: async (args: Record<string, any>, _context: PluginContext) => {
            if (!this.service) throw new Error('Notion service not available');
            return this.service.getPage(args.pageId);
          }
        }
      ]
    };
  }
}

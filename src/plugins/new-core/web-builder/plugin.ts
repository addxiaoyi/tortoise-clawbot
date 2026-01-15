import { PluginContext, SkillPlugin, SkillDefinition } from '../types';
import { WebBuilderConfig } from './types';
import { WebBuilderService } from './service';

export class WebBuilderPlugin implements SkillPlugin {
  private context?: PluginContext;
  private service?: WebBuilderService;
  private config?: WebBuilderConfig;

  public async onInit(ctx: PluginContext): Promise<void> {
    this.context = ctx;
    this.config = ctx.getConfig<WebBuilderConfig>();
    ctx.logger.info('WebBuilder Plugin initialized');
  }

  public async onStart(): Promise<void> {
    if (!this.context || !this.config) {
      throw new Error('WebBuilderPlugin not initialized');
    }

    this.service = new WebBuilderService(this.config, this.context.logger);
    await this.service.start();
  }

  public async onStop(): Promise<void> {
    if (this.service) {
      await this.service.stop();
    }
  }

  public getService(): WebBuilderService | undefined {
    return this.service;
  }

  public getSkillDefinition(): SkillDefinition {
    return {
      name: 'web-builder',
      version: '1.0.0',
      description: 'Web Builder integration for creating web components and pages',
      tools: [
        {
          name: 'build_react_component',
          description: 'Build a React component (writes to file for now)',
          parameters: [
            {
              name: 'filename',
              type: 'string',
              description: 'Filename (e.g. App.tsx)',
              required: true
            },
            {
              name: 'code',
              type: 'string',
              description: 'Component code',
              required: true
            }
          ],
          execute: async (args: Record<string, any>, _context: PluginContext) => {
            if (!this.service) throw new Error('WebBuilder service not available');
            const filePath = await this.service.buildReactComponent(args.filename, args.code);
            return { filePath };
          }
        },
        {
          name: 'generate_html',
          description: 'Generate an HTML page with Tailwind CSS',
          parameters: [
            {
              name: 'filename',
              type: 'string',
              description: 'Filename (e.g. index.html)',
              required: true
            },
            {
              name: 'title',
              type: 'string',
              description: 'Page title',
              required: true
            },
            {
              name: 'body',
              type: 'string',
              description: 'Body content (HTML)',
              required: true
            }
          ],
          execute: async (args: Record<string, any>, _context: PluginContext) => {
            if (!this.service) throw new Error('WebBuilder service not available');
            const filePath = await this.service.generateHtml(args.filename, args.title, args.body);
            return { filePath };
          }
        }
      ]
    };
  }
}


import { PluginContext, SkillPlugin, SkillDefinition } from '../types';
import { CanvasConfig } from './types';
import { CanvasService } from './service';

export class CanvasPlugin implements SkillPlugin {
  private context?: PluginContext;
  private service?: CanvasService;
  private config?: CanvasConfig;

  public async onInit(ctx: PluginContext): Promise<void> {
    this.context = ctx;
    this.config = ctx.getConfig<CanvasConfig>();
    ctx.logger.info('Canvas Plugin initialized');
  }

  public async onStart(): Promise<void> {
    if (!this.context || !this.config) {
      throw new Error('CanvasPlugin not initialized');
    }

    this.service = new CanvasService(this.config, this.context.logger);
    await this.service.start();
  }

  public async onStop(): Promise<void> {
    if (this.service) {
      await this.service.stop();
    }
  }

  public getService(): CanvasService | undefined {
    return this.service;
  }

  public getSkillDefinition(): SkillDefinition {
    return {
      name: 'canvas',
      version: '1.0.0',
      description: 'Canvas integration for serving and managing HTML content',
      tools: [
        {
          name: 'write_canvas_file',
          description: 'Create or update an HTML file in the canvas root',
          parameters: [
            {
              name: 'filename',
              type: 'string',
              description: 'Filename (e.g. index.html)',
              required: true
            },
            {
              name: 'content',
              type: 'string',
              description: 'HTML content',
              required: true
            }
          ],
          execute: async (args: Record<string, any>, _context: PluginContext) => {
            if (!this.service) throw new Error('Canvas service not available');
            const url = await this.service.writeFile(args.filename, args.content);
            return { url, filename: args.filename };
          }
        },
        {
          name: 'get_canvas_url',
          description: 'Get the URL for a canvas file',
          parameters: [
            {
              name: 'filename',
              type: 'string',
              description: 'Filename',
              required: true
            }
          ],
          execute: async (args: Record<string, any>, _context: PluginContext) => {
            if (!this.service) throw new Error('Canvas service not available');
            return { url: this.service.getUrl(args.filename) };
          }
        }
      ]
    };
  }
}

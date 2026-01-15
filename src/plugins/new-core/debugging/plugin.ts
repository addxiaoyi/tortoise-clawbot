import { PluginContext, SkillPlugin, SkillDefinition } from '../types';
import { DebuggingConfig } from './types';
import { DebuggingService } from './service';

export class DebuggingPlugin implements SkillPlugin {
  private context?: PluginContext;
  private service?: DebuggingService;
  private config?: DebuggingConfig;

  public async onInit(ctx: PluginContext): Promise<void> {
    this.context = ctx;
    this.config = ctx.getConfig<DebuggingConfig>();
    ctx.logger.info('Debugging Plugin initialized');
  }

  public async onStart(): Promise<void> {
    if (!this.context || !this.config) {
      throw new Error('DebuggingPlugin not initialized');
    }

    this.service = new DebuggingService(this.config, this.context.logger);
    await this.service.start();
  }

  public async onStop(): Promise<void> {
    if (this.service) {
      await this.service.stop();
    }
  }

  public getService(): DebuggingService | undefined {
    return this.service;
  }

  public getSkillDefinition(): SkillDefinition {
    return {
      name: 'debugging',
      version: '1.0.0',
      description: 'Debugging integration for inspecting and controlling execution',
      tools: [
        {
          name: 'set_breakpoint',
          description: 'Set a breakpoint at a specific file and line',
          parameters: [
            {
              name: 'file',
              type: 'string',
              description: 'File path',
              required: true
            },
            {
              name: 'line',
              type: 'number',
              description: 'Line number',
              required: true
            }
          ],
          execute: async (args: Record<string, any>, _context: PluginContext) => {
            if (!this.service) throw new Error('Debugging service not available');
            const result = await this.service.setBreakpoint(args.file, args.line);
            return { result };
          }
        },
        {
          name: 'inspect_variable',
          description: 'Inspect a variable value',
          parameters: [
            {
              name: 'name',
              type: 'string',
              description: 'Variable name',
              required: true
            }
          ],
          execute: async (args: Record<string, any>, _context: PluginContext) => {
            if (!this.service) throw new Error('Debugging service not available');
            const result = await this.service.inspectVariable(args.name);
            return { result };
          }
        }
      ]
    };
  }
}

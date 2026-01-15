import { PluginContext, SkillPlugin, SkillDefinition } from '../types';
import { TestingConfig } from './types';
import { TestingService } from './service';

export class TestingPlugin implements SkillPlugin {
  private context?: PluginContext;
  private service?: TestingService;
  private config?: TestingConfig;

  public async onInit(ctx: PluginContext): Promise<void> {
    this.context = ctx;
    this.config = ctx.getConfig<TestingConfig>();
    ctx.logger.info('Testing Plugin initialized');
  }

  public async onStart(): Promise<void> {
    if (!this.context || !this.config) {
      throw new Error('TestingPlugin not initialized');
    }

    this.service = new TestingService(this.config, this.context.logger);
    await this.service.start();
  }

  public async onStop(): Promise<void> {
    if (this.service) {
      await this.service.stop();
    }
  }

  public getService(): TestingService | undefined {
    return this.service;
  }

  public getSkillDefinition(): SkillDefinition {
    return {
      name: 'testing',
      version: '1.0.0',
      description: 'Testing integration for running and managing tests',
      tools: [
        {
          name: 'run_tests',
          description: 'Run tests matching a pattern',
          parameters: [
            {
              name: 'pattern',
              type: 'string',
              description: 'Test file pattern',
              required: true
            }
          ],
          execute: async (args: Record<string, any>, _context: PluginContext) => {
            if (!this.service) throw new Error('Testing service not available');
            const output = await this.service.runTests(args.pattern);
            return { output };
          }
        },
        {
          name: 'create_test',
          description: 'Create a new test file',
          parameters: [
            {
              name: 'filename',
              type: 'string',
              description: 'Filename (e.g. example.test.ts)',
              required: true
            },
            {
              name: 'content',
              type: 'string',
              description: 'Test content',
              required: true
            }
          ],
          execute: async (args: Record<string, any>, _context: PluginContext) => {
            if (!this.service) throw new Error('Testing service not available');
            const filePath = await this.service.createTestFile(args.filename, args.content);
            return { filePath };
          }
        }
      ]
    };
  }
}

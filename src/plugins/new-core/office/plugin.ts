import { PluginContext, SkillPlugin, SkillDefinition } from '../types';
import { OfficeConfig } from './types';
import { OfficeService } from './service';

export class OfficePlugin implements SkillPlugin {
  private context?: PluginContext;
  private service?: OfficeService;
  private config?: OfficeConfig;

  public async onInit(ctx: PluginContext): Promise<void> {
    this.context = ctx;
    this.config = ctx.getConfig<OfficeConfig>();
    ctx.logger.info('Office Plugin initialized');
  }

  public async onStart(): Promise<void> {
    if (!this.context || !this.config) {
      throw new Error('OfficePlugin not initialized');
    }

    this.service = new OfficeService(this.config, this.context.logger);
    await this.service.start();
  }

  public async onStop(): Promise<void> {
    if (this.service) {
      await this.service.stop();
    }
  }

  public getService(): OfficeService | undefined {
    return this.service;
  }

  public getSkillDefinition(): SkillDefinition {
    return {
      name: 'office',
      version: '1.0.0',
      description: 'Office integration for managing Excel, Word, and PowerPoint files',
      tools: [
        {
          name: 'create_excel',
          description: 'Create an Excel file with data',
          parameters: [
            {
              name: 'filename',
              type: 'string',
              description: 'Filename (e.g. data.xlsx)',
              required: true
            },
            {
              name: 'data',
              type: 'array',
              description: '2D array of data',
              required: true
            }
          ],
          execute: async (args: Record<string, any>, _context: PluginContext) => {
            if (!this.service) throw new Error('Office service not available');
            const filePath = await this.service.createExcel(args.filename, args.data);
            return { filePath };
          }
        },
        {
          name: 'read_excel',
          description: 'Read data from an Excel file',
          parameters: [
            {
              name: 'filename',
              type: 'string',
              description: 'Filename',
              required: true
            }
          ],
          execute: async (args: Record<string, any>, _context: PluginContext) => {
            if (!this.service) throw new Error('Office service not available');
            return this.service.readExcel(args.filename);
          }
        },
        {
          name: 'create_word',
          description: 'Create a Word document with text',
          parameters: [
            {
              name: 'filename',
              type: 'string',
              description: 'Filename (e.g. report.docx)',
              required: true
            },
            {
              name: 'text',
              type: 'string',
              description: 'Text content',
              required: true
            }
          ],
          execute: async (args: Record<string, any>, _context: PluginContext) => {
            if (!this.service) throw new Error('Office service not available');
            const filePath = await this.service.createWordDoc(args.filename, args.text);
            return { filePath };
          }
        }
      ]
    };
  }
}

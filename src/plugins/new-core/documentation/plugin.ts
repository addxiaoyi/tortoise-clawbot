import { PluginContext, SkillPlugin, SkillDefinition } from '../types';
import { DocumentationConfig } from './types';
import { DocumentationService } from './service';

export class DocumentationPlugin implements SkillPlugin {
  private context?: PluginContext;
  private service?: DocumentationService;
  private config?: DocumentationConfig;

  public async onInit(ctx: PluginContext): Promise<void> {
    this.context = ctx;
    this.config = ctx.getConfig<DocumentationConfig>();
    ctx.logger.info('Documentation Plugin initialized');
  }

  public async onStart(): Promise<void> {
    if (!this.context || !this.config) {
      throw new Error('DocumentationPlugin not initialized');
    }

    this.service = new DocumentationService(this.config, this.context.logger);
    await this.service.start();
  }

  public async onStop(): Promise<void> {
    if (this.service) {
      await this.service.stop();
    }
  }

  public getService(): DocumentationService | undefined {
    return this.service;
  }

  public getSkillDefinition(): SkillDefinition {
    return {
      name: 'documentation',
      version: '1.0.0',
      description: 'Documentation generator and manager',
      tools: [
        {
          name: 'generate_doc',
          description: 'Generate documentation',
          parameters: [
            {
              name: 'title',
              type: 'string',
              description: 'Document title',
              required: true
            },
            {
              name: 'content',
              type: 'string',
              description: 'Document content',
              required: true
            }
          ],
          execute: async (args: Record<string, any>, _context: PluginContext) => {
            if (!this.service) throw new Error('Documentation service not available');
            const filename = await this.service.generateDoc(args.title, args.content);
            return { filename };
          }
        },
        {
          name: 'list_docs',
          description: 'List all generated documents',
          parameters: [],
          execute: async (_args: Record<string, any>, _context: PluginContext) => {
            if (!this.service) throw new Error('Documentation service not available');
            const docs = await this.service.listDocs();
            return { docs };
          }
        },
        {
          name: 'get_doc_content',
          description: 'Get content of a document',
          parameters: [
            {
              name: 'filename',
              type: 'string',
              description: 'Filename of the document',
              required: true
            }
          ],
          execute: async (args: Record<string, any>, _context: PluginContext) => {
            if (!this.service) throw new Error('Documentation service not available');
            const content = await this.service.getDocContent(args.filename);
            if (!content) {
              throw new Error(`Document ${args.filename} not found`);
            }
            return { content };
          }
        }
      ]
    };
  }
}

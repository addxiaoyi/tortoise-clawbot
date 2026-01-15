import { PluginContext, SkillPlugin, SkillDefinition } from '../types';
import { GitWorkflowConfig } from './types';
import { GitWorkflowService } from './service';

export class GitWorkflowPlugin implements SkillPlugin {
  private context?: PluginContext;
  private service?: GitWorkflowService;
  private config?: GitWorkflowConfig;

  public async onInit(ctx: PluginContext): Promise<void> {
    this.context = ctx;
    this.config = ctx.getConfig<GitWorkflowConfig>();
    ctx.logger.info('GitWorkflow Plugin initialized');
  }

  public async onStart(): Promise<void> {
    if (!this.context || !this.config) {
      throw new Error('GitWorkflowPlugin not initialized');
    }

    this.service = new GitWorkflowService(this.config, this.context.logger);
    await this.service.start();
  }

  public async onStop(): Promise<void> {
    if (this.service) {
      await this.service.stop();
    }
  }

  public getService(): GitWorkflowService | undefined {
    return this.service;
  }

  public getSkillDefinition(): SkillDefinition {
    return {
      name: 'git-workflow',
      version: '1.0.0',
      description: 'Git Workflow integration for managing branches and commits',
      tools: [
        {
          name: 'create_branch',
          description: 'Create a new branch',
          parameters: [
            {
              name: 'name',
              type: 'string',
              description: 'Branch name',
              required: true
            }
          ],
          execute: async (args: Record<string, any>, _context: PluginContext) => {
            if (!this.service) throw new Error('GitWorkflow service not available');
            const result = await this.service.createBranch(args.name);
            return { result };
          }
        },
        {
          name: 'checkout_branch',
          description: 'Checkout an existing branch',
          parameters: [
            {
              name: 'name',
              type: 'string',
              description: 'Branch name',
              required: true
            }
          ],
          execute: async (args: Record<string, any>, _context: PluginContext) => {
            if (!this.service) throw new Error('GitWorkflow service not available');
            const result = await this.service.checkoutBranch(args.name);
            return { result };
          }
        },
        {
          name: 'commit_changes',
          description: 'Commit changes with a message',
          parameters: [
            {
              name: 'message',
              type: 'string',
              description: 'Commit message',
              required: true
            }
          ],
          execute: async (args: Record<string, any>, _context: PluginContext) => {
            if (!this.service) throw new Error('GitWorkflow service not available');
            const result = await this.service.commit(args.message);
            return { result };
          }
        }
      ]
    };
  }
}

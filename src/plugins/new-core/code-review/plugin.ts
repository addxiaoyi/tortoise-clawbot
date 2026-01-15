import { PluginContext, SkillPlugin, SkillDefinition } from '../types';
import { CodeReviewConfig } from './types';
import { CodeReviewService } from './service';

export class CodeReviewPlugin implements SkillPlugin {
  private context?: PluginContext;
  private service?: CodeReviewService;
  private config?: CodeReviewConfig;

  public async onInit(ctx: PluginContext): Promise<void> {
    this.context = ctx;
    this.config = ctx.getConfig<CodeReviewConfig>();
    ctx.logger.info('CodeReview Plugin initialized');
  }

  public async onStart(): Promise<void> {
    if (!this.context || !this.config) {
      throw new Error('CodeReviewPlugin not initialized');
    }

    this.service = new CodeReviewService(this.config, this.context.logger);
    await this.service.start();
  }

  public async onStop(): Promise<void> {
    if (this.service) {
      await this.service.stop();
    }
  }

  public getService(): CodeReviewService | undefined {
    return this.service;
  }

  public getSkillDefinition(): SkillDefinition {
    return {
      name: 'code-review',
      version: '1.0.0',
      description: 'Code Review integration for analyzing code and diffs',
      tools: [
        {
          name: 'review_code',
          description: 'Review a code snippet',
          parameters: [
            {
              name: 'code',
              type: 'string',
              description: 'Code content',
              required: true
            },
            {
              name: 'language',
              type: 'string',
              description: 'Programming language',
              required: true
            }
          ],
          execute: async (args: Record<string, any>, _context: PluginContext) => {
            if (!this.service) throw new Error('CodeReview service not available');
            const feedback = await this.service.reviewCode(args.code, args.language);
            return { feedback };
          }
        },
        {
          name: 'analyze_diff',
          description: 'Analyze a git diff',
          parameters: [
            {
              name: 'diff',
              type: 'string',
              description: 'Diff content',
              required: true
            }
          ],
          execute: async (args: Record<string, any>, _context: PluginContext) => {
            if (!this.service) throw new Error('CodeReview service not available');
            const analysis = await this.service.analyzeDiff(args.diff);
            return { analysis };
          }
        }
      ]
    };
  }
}

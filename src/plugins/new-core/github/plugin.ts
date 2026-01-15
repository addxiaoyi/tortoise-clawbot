
import { PluginContext, SkillPlugin, SkillDefinition } from '../types';
import { GitHubService } from './service';

export class GitHubPlugin implements SkillPlugin {
  private context?: PluginContext;
  private service: GitHubService;

  constructor() {
    this.service = new GitHubService();
  }

  public async onInit(ctx: PluginContext): Promise<void> {
    this.context = ctx;
    ctx.logger.info('GitHub Plugin initialized');
  }

  public async onStart(): Promise<void> {
    if (!this.context) {
      throw new Error('GitHubPlugin not initialized');
    }

    const isValid = await this.service.checkAuth();
    if (isValid) {
      this.context.logger.info('GitHub auth status: valid');
    } else {
      this.context.logger.warn('GitHub auth status: invalid or not logged in');
    }
  }

  public async onStop(): Promise<void> {
    // No cleanup needed yet
  }

  public getService(): GitHubService {
    return this.service;
  }

  public getSkillDefinition(): SkillDefinition {
    return {
      name: 'github',
      version: '1.0.0',
      description: 'GitHub integration for issue and repository management',
      tools: [
        {
          name: 'list_issues',
          description: 'List issues in a repository',
          parameters: [
            {
              name: 'repo',
              type: 'string',
              description: 'Repository name in format owner/repo',
              required: true
            }
          ],
          execute: async (args: Record<string, any>, _context: PluginContext) => {
            return this.service.listIssues(args.repo);
          }
        },
        {
          name: 'predict_labels',
          description: 'Predict labels for an issue based on title and body',
          parameters: [
            {
              name: 'title',
              type: 'string',
              description: 'Issue title',
              required: true
            },
            {
              name: 'body',
              type: 'string',
              description: 'Issue body',
              required: true
            }
          ],
          execute: async (args: Record<string, any>, _context: PluginContext) => {
            return this.service.predictLabels(args.title, args.body);
          }
        }
      ]
    };
  }
}

import { PluginContext, SkillPlugin, SkillDefinition } from '../types';
import { PlanningConfig } from './types';
import { PlanningService } from './service';

export class PlanningPlugin implements SkillPlugin {
  private context?: PluginContext;
  private service?: PlanningService;
  private config?: PlanningConfig;

  public async onInit(ctx: PluginContext): Promise<void> {
    this.context = ctx;
    this.config = ctx.getConfig<PlanningConfig>();
    ctx.logger.info('Planning Plugin initialized');
  }

  public async onStart(): Promise<void> {
    if (!this.context || !this.config) {
      throw new Error('PlanningPlugin not initialized');
    }

    this.service = new PlanningService(this.config, this.context.logger);
    await this.service.start();
  }

  public async onStop(): Promise<void> {
    if (this.service) {
      await this.service.stop();
    }
  }

  public getService(): PlanningService | undefined {
    return this.service;
  }

  public getSkillDefinition(): SkillDefinition {
    return {
      name: 'planning',
      version: '1.0.0',
      description: 'Planning integration for managing tasks and workflows',
      tools: [
        {
          name: 'create_task',
          description: 'Create a new task',
          parameters: [
            {
              name: 'title',
              type: 'string',
              description: 'Task title',
              required: true
            }
          ],
          execute: async (args: Record<string, any>, _context: PluginContext) => {
            if (!this.service) throw new Error('Planning service not available');
            const task = await this.service.createTask(args.title);
            return { task };
          }
        },
        {
          name: 'update_task_status',
          description: 'Update the status of a task',
          parameters: [
            {
              name: 'id',
              type: 'string',
              description: 'Task ID',
              required: true
            },
            {
              name: 'status',
              type: 'string',
              description: 'New status (pending, in_progress, completed)',
              required: true
            }
          ],
          execute: async (args: Record<string, any>, _context: PluginContext) => {
            if (!this.service) throw new Error('Planning service not available');
            const task = await this.service.updateTaskStatus(args.id, args.status);
            if (!task) {
              throw new Error(`Task with ID ${args.id} not found`);
            }
            return { task };
          }
        },
        {
          name: 'list_tasks',
          description: 'List all tasks',
          parameters: [],
          execute: async (_args: Record<string, any>, _context: PluginContext) => {
            if (!this.service) throw new Error('Planning service not available');
            const tasks = await this.service.getTasks();
            return { tasks };
          }
        }
      ]
    };
  }
}

import { PlanningConfig } from './types';
import { PluginLogger } from '../types';

interface Task {
  id: string;
  title: string;
  status: 'pending' | 'in_progress' | 'completed';
}

export class PlanningService {
  private config: PlanningConfig;
  private logger: PluginLogger;
  private tasks: Map<string, Task>;

  constructor(config: PlanningConfig, logger: PluginLogger) {
    this.config = config;
    this.logger = logger;
    this.tasks = new Map();
  }

  public async start(): Promise<void> {
    this.logger.info('Planning service started');
  }

  public async stop(): Promise<void> {
    this.logger.info('Planning service stopped');
  }

  public async createTask(title: string): Promise<Task> {
    const id = Math.random().toString(36).substring(7);
    const task: Task = {
      id,
      title,
      status: 'pending'
    };
    this.tasks.set(id, task);
    this.logger.info(`Created task: ${title} (${id})`);
    return task;
  }

  public async updateTaskStatus(id: string, status: 'pending' | 'in_progress' | 'completed'): Promise<Task | null> {
    const task = this.tasks.get(id);
    if (!task) {
      return null;
    }
    task.status = status;
    this.logger.info(`Updated task status: ${id} -> ${status}`);
    return task;
  }

  public async getTasks(): Promise<Task[]> {
    return Array.from(this.tasks.values());
  }
}

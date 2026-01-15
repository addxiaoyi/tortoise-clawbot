import { GitWorkflowConfig } from './types';
import { PluginLogger } from '../types';

export class GitWorkflowService {
  private config: GitWorkflowConfig;
  private logger: PluginLogger;
  private branches: string[];

  constructor(config: GitWorkflowConfig, logger: PluginLogger) {
    this.config = config;
    this.logger = logger;
    this.branches = ['main'];
  }

  public async start(): Promise<void> {
    this.logger.info('GitWorkflow service started');
  }

  public async stop(): Promise<void> {
    this.logger.info('GitWorkflow service stopped');
  }

  public async createBranch(name: string): Promise<string> {
    if (this.branches.includes(name)) {
      return `Branch ${name} already exists`;
    }
    this.branches.push(name);
    this.logger.info(`Created branch ${name}`);
    return `Branch ${name} created`;
  }

  public async checkoutBranch(name: string): Promise<string> {
    if (!this.branches.includes(name)) {
      return `Branch ${name} does not exist`;
    }
    this.logger.info(`Checked out branch ${name}`);
    return `Checked out branch ${name}`;
  }

  public async commit(message: string): Promise<string> {
    this.logger.info(`Committed with message: ${message}`);
    return `Committed changes: ${message}`;
  }
}

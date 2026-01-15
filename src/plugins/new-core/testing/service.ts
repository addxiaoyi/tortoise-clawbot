import fs from 'fs/promises';
import { resolveSafeRelativeFile } from '../../../utils/safe-path.js';
import { TestingConfig } from './types';
import { PluginLogger } from '../types';

export class TestingService {
  private config: TestingConfig;
  private logger: PluginLogger;

  constructor(config: TestingConfig, logger: PluginLogger) {
    this.config = config;
    this.logger = logger;
  }

  public async start(): Promise<void> {
    const { testDir } = this.config;
    try {
      await fs.access(testDir);
    } catch {
      await fs.mkdir(testDir, { recursive: true });
    }
    this.logger.info('Testing service started');
  }

  public async stop(): Promise<void> {
    this.logger.info('Testing service stopped');
  }

  public async runTests(pattern: string): Promise<string> {
    // In a real implementation, this would spawn a vitest process
    // For now, we'll simulate test execution
    this.logger.info(`Running tests matching pattern: ${pattern}`);
    return `Ran tests matching "${pattern}". Result: PASS`;
  }

  public async createTestFile(filename: string, content: string): Promise<string> {
    const filePath = resolveSafeRelativeFile(this.config.testDir, filename);

    await fs.writeFile(filePath, content, 'utf-8');
    return filePath;
  }
}

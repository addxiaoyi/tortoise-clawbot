import { DebuggingConfig } from './types';
import { PluginLogger } from '../types';

export class DebuggingService {
  private config: DebuggingConfig;
  private logger: PluginLogger;
  private breakpoints: Map<string, number[]>;

  constructor(config: DebuggingConfig, logger: PluginLogger) {
    this.config = config;
    this.logger = logger;
    this.breakpoints = new Map();
  }

  public async start(): Promise<void> {
    this.logger.info('Debugging service started');
  }

  public async stop(): Promise<void> {
    this.logger.info('Debugging service stopped');
  }

  public async setBreakpoint(file: string, line: number): Promise<string> {
    if (!this.breakpoints.has(file)) {
      this.breakpoints.set(file, []);
    }
    const lines = this.breakpoints.get(file);
    if (lines && !lines.includes(line)) {
      lines.push(line);
      this.logger.info(`Breakpoint set at ${file}:${line}`);
      return `Breakpoint set at ${file}:${line}`;
    }
    return `Breakpoint already exists at ${file}:${line}`;
  }

  public async removeBreakpoint(file: string, line: number): Promise<string> {
    if (this.breakpoints.has(file)) {
      const lines = this.breakpoints.get(file);
      if (lines) {
        const index = lines.indexOf(line);
        if (index > -1) {
          lines.splice(index, 1);
          this.logger.info(`Breakpoint removed at ${file}:${line}`);
          return `Breakpoint removed at ${file}:${line}`;
        }
      }
    }
    return `Breakpoint not found at ${file}:${line}`;
  }

  public async inspectVariable(name: string): Promise<string> {
    // Mock inspection
    this.logger.info(`Inspecting variable ${name}`);
    return `Variable ${name}: undefined (mocked)`;
  }
}

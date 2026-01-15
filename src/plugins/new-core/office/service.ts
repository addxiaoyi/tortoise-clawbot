import fs from 'fs/promises';
import { resolveSafeRelativeFile } from '../../../utils/safe-path.js';
import { OfficeConfig } from './types';
import { PluginLogger } from '../types';

export class OfficeService {
  private config: OfficeConfig;
  private logger: PluginLogger;

  constructor(config: OfficeConfig, logger: PluginLogger) {
    this.config = config;
    this.logger = logger;
  }

  public async start(): Promise<void> {
    const { outputDir } = this.config;
    try {
      await fs.access(outputDir);
    } catch {
      await fs.mkdir(outputDir, { recursive: true });
    }
    this.logger.info('Office service started');
  }

  public async stop(): Promise<void> {
    this.logger.info('Office service stopped');
  }

  public async createExcel(filename: string, data: any[][]): Promise<string> {
    const filePath = resolveSafeRelativeFile(this.config.outputDir, filename);
    // Simulation: Write CSV
    const csvContent = data.map(row => row.join(',')).join('\n');
    await fs.writeFile(filePath, csvContent, 'utf-8');
    
    return filePath;
  }

  public async readExcel(filename: string): Promise<any[][]> {
    const filePath = resolveSafeRelativeFile(this.config.outputDir, filename);
    const content = await fs.readFile(filePath, 'utf-8');
    // Simulation: Read CSV
    return content.split('\n').map(line => line.split(','));
  }

  public async createWordDoc(filename: string, text: string): Promise<string> {
    const filePath = resolveSafeRelativeFile(this.config.outputDir, filename);
    // Simulation: Write text file
    await fs.writeFile(filePath, text, 'utf-8');
    return filePath;
  }
}

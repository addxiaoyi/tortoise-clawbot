import fs from 'fs/promises';
import { escapeHtml } from '../../../utils/html-escape.js';
import { resolveSafeRelativeFile } from '../../../utils/safe-path.js';
import { WebBuilderConfig } from './types';
import { PluginLogger } from '../types';

export class WebBuilderService {
  private config: WebBuilderConfig;
  private logger: PluginLogger;

  constructor(config: WebBuilderConfig, logger: PluginLogger) {
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
    this.logger.info('WebBuilder service started');
  }

  public async stop(): Promise<void> {
    this.logger.info('WebBuilder service stopped');
  }

  public async buildReactComponent(filename: string, code: string): Promise<string> {
    // In a real implementation, this would use esbuild to bundle the component
    // For now, we'll just write the file to the output directory
    const filePath = resolveSafeRelativeFile(this.config.outputDir, filename);

    await fs.writeFile(filePath, code, 'utf-8');
    return filePath;
  }

  public async generateHtml(filename: string, title: string, body: string): Promise<string> {
    const filePath = resolveSafeRelativeFile(this.config.outputDir, filename);

    const safeTitle = escapeHtml(title);
    const safeBody = escapeHtml(body);

    const html = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>${safeTitle}</title>
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body>
    ${safeBody}
</body>
</html>`;

    await fs.writeFile(filePath, html, 'utf-8');
    return filePath;
  }
}

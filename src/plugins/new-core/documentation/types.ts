import { PluginConfig } from '../types';

export interface DocumentationConfig extends PluginConfig {
  outputDir?: string;
  format?: 'markdown' | 'html';
}

import { PluginConfig } from '../types';

export interface CanvasConfig extends PluginConfig {
  root: string;
  port: number;
  host?: string;
}


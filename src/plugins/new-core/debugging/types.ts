import { PluginConfig } from '../types';

export interface DebuggingConfig extends PluginConfig {
  attachTimeout?: number;
}

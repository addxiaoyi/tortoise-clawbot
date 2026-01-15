import { PluginConfig } from '../types';

export interface SecurityConfig extends PluginConfig {
  allowedTools?: string[];
  blockedTools?: string[];
}


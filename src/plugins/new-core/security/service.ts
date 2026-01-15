import { SecurityConfig } from './types';
import { PluginLogger } from '../types';

export class SecurityService {
  private config: SecurityConfig;
  private logger: PluginLogger;

  constructor(config: SecurityConfig, logger: PluginLogger) {
    this.config = config;
    this.logger = logger;
  }

  public async start(): Promise<void> {
    this.logger.info('Security service started');
  }

  public async stop(): Promise<void> {
    this.logger.info('Security service stopped');
  }

  public async checkToolAllowed(toolName: string): Promise<boolean> {
    const { allowedTools, blockedTools } = this.config;

    if (blockedTools && blockedTools.includes(toolName)) {
      this.logger.warn(`Tool ${toolName} is blocked by security policy`);
      return false;
    }

    if (allowedTools && allowedTools.length > 0) {
      const allowed = allowedTools.includes(toolName);
      if (!allowed) {
        this.logger.warn(`Tool ${toolName} is not in allowlist`);
      }
      return allowed;
    }

    return true;
  }

  /**
   * 启发式风险分级（易被绕过），仅作提示；真正约束应依赖工具白名单与调用侧权限。
   */
  public async analyzeInputRisk(input: string): Promise<'low' | 'medium' | 'high'> {
    const lowered = input.toLowerCase();
    if (
      lowered.includes('rm -rf') ||
      lowered.includes('drop table') ||
      /\|\s*(ba)?sh\b/.test(lowered) ||
      /\bcurl\b[^\n]{0,80}\|/.test(lowered) ||
      /\bwget\b[^\n]{0,80}\|/.test(lowered)
    ) {
      return 'high';
    }
    if (lowered.includes('password') || lowered.includes('secret')) {
      return 'medium';
    }
    return 'low';
  }
}


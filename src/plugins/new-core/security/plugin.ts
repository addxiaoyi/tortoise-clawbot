import { PluginContext, SkillPlugin, SkillDefinition } from '../types';
import { SecurityConfig } from './types';
import { SecurityService } from './service';

export class SecurityPlugin implements SkillPlugin {
  private context?: PluginContext;
  private service?: SecurityService;
  private config?: SecurityConfig;

  public async onInit(ctx: PluginContext): Promise<void> {
    this.context = ctx;
    this.config = ctx.getConfig<SecurityConfig>();
    ctx.logger.info('Security Plugin initialized');
  }

  public async onStart(): Promise<void> {
    if (!this.context || !this.config) {
      throw new Error('SecurityPlugin not initialized');
    }

    this.service = new SecurityService(this.config, this.context.logger);
    await this.service.start();
  }

  public async onStop(): Promise<void> {
    if (this.service) {
      await this.service.stop();
    }
  }

  public getService(): SecurityService | undefined {
    return this.service;
  }

  public getSkillDefinition(): SkillDefinition {
    return {
      name: 'security',
      version: '1.0.0',
      description: 'Security integration for tool allow/deny and risk analysis',
      tools: [
        {
          name: 'check_tool_allowed',
          description: 'Check if a tool is allowed by security policy',
          parameters: [
            {
              name: 'tool',
              type: 'string',
              description: 'Tool name',
              required: true
            }
          ],
          execute: async (args: Record<string, any>, _context: PluginContext) => {
            if (!this.service) throw new Error('Security service not available');
            const allowed = await this.service.checkToolAllowed(args.tool);
            return { allowed };
          }
        },
        {
          name: 'analyze_input_risk',
          description: 'Analyze risk level of a text input',
          parameters: [
            {
              name: 'input',
              type: 'string',
              description: 'User input to analyze',
              required: true
            }
          ],
          execute: async (args: Record<string, any>, _context: PluginContext) => {
            if (!this.service) throw new Error('Security service not available');
            const risk = await this.service.analyzeInputRisk(args.input);
            return { risk };
          }
        }
      ]
    };
  }
}


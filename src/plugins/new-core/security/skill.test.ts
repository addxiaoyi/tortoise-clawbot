import { describe, it, expect, vi, beforeEach } from 'vitest';
import { SecurityPlugin } from './plugin.js';
import { SkillTestHarness } from '../../../utils/skill-test-harness.js';

vi.mock('./service.js', () => {
  class SecurityService {
    start = vi.fn();
    stop = vi.fn();
    checkToolAllowed = vi.fn();
    analyzeInputRisk = vi.fn();
  }
  return { SecurityService };
});

describe('SecurityPlugin Skill', () => {
  let plugin: SecurityPlugin;
  let harness: SkillTestHarness;
  let mockServiceInstance: any;

  beforeEach(async () => {
    vi.clearAllMocks();

    plugin = new SecurityPlugin();
    harness = new SkillTestHarness(plugin);

    harness.context.getConfig = vi.fn().mockReturnValue({
      allowedTools: ['safe_tool'],
      blockedTools: ['dangerous_tool']
    });

    await harness.init();
    mockServiceInstance = (plugin as any).service;
    mockServiceInstance.checkToolAllowed.mockResolvedValue(true);
    mockServiceInstance.analyzeInputRisk.mockResolvedValue('low');
  });

  it('should expose skill definition', () => {
    const definition = plugin.getSkillDefinition();
    expect(definition.name).toBe('security');
    expect(definition.tools).toHaveLength(2);
  });

  it('should execute check_tool_allowed tool', async () => {
    const result = await harness.executeTool('check_tool_allowed', { tool: 'safe_tool' });
    expect(mockServiceInstance.checkToolAllowed).toHaveBeenCalledWith('safe_tool');
    expect(result).toEqual({ allowed: true });
  });

  it('should execute analyze_input_risk tool', async () => {
    mockServiceInstance.analyzeInputRisk.mockResolvedValue('high');
    const result = await harness.executeTool('analyze_input_risk', { input: 'rm -rf /' });
    expect(mockServiceInstance.analyzeInputRisk).toHaveBeenCalledWith('rm -rf /');
    expect(result).toEqual({ risk: 'high' });
  });
});

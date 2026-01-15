import { describe, it, expect, vi, beforeEach } from 'vitest';
import { DebuggingPlugin } from './plugin.js';
import { SkillTestHarness } from '../../../utils/skill-test-harness.js';

vi.mock('./service.js', () => {
  class DebuggingService {
    start = vi.fn();
    stop = vi.fn();
    setBreakpoint = vi.fn();
    inspectVariable = vi.fn();
  }
  return { DebuggingService };
});

describe('DebuggingPlugin Skill', () => {
  let plugin: DebuggingPlugin;
  let harness: SkillTestHarness;
  let mockServiceInstance: any;

  beforeEach(async () => {
    vi.clearAllMocks();
    
    plugin = new DebuggingPlugin();
    harness = new SkillTestHarness(plugin);
    
    // Mock config
    harness.context.getConfig = vi.fn().mockReturnValue({ attachTimeout: 5000 });
    
    await harness.init();
    mockServiceInstance = (plugin as any).service;
    mockServiceInstance.setBreakpoint.mockResolvedValue('Breakpoint set at test.ts:10');
    mockServiceInstance.inspectVariable.mockResolvedValue('Variable x: 42');
  });

  it('should expose skill definition', () => {
    const definition = plugin.getSkillDefinition();
    expect(definition.name).toBe('debugging');
    expect(definition.tools).toHaveLength(2);
  });

  it('should execute set_breakpoint tool', async () => {
    const result = await harness.executeTool('set_breakpoint', { file: 'test.ts', line: 10 });
    expect(mockServiceInstance.setBreakpoint).toHaveBeenCalledWith('test.ts', 10);
    expect(result).toEqual({ result: 'Breakpoint set at test.ts:10' });
  });

  it('should execute inspect_variable tool', async () => {
    const result = await harness.executeTool('inspect_variable', { name: 'x' });
    expect(mockServiceInstance.inspectVariable).toHaveBeenCalledWith('x');
    expect(result).toEqual({ result: 'Variable x: 42' });
  });
});

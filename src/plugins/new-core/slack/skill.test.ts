import { describe, it, expect, vi, beforeEach } from 'vitest';
import { SlackPlugin } from './plugin.js';
import { SlackService } from './service.js';
import { SkillTestHarness } from '../../../utils/skill-test-harness.js';

// Mock SlackService
vi.mock('./service.js');

describe('SlackPlugin Skill', () => {
  let plugin: SlackPlugin;
  let harness: SkillTestHarness;
  let mockService: any;

  beforeEach(async () => {
    vi.clearAllMocks();
    
    plugin = new SlackPlugin();
    harness = new SkillTestHarness(plugin);
    
    // Mock config
    harness.context.getConfig = vi.fn().mockReturnValue({ token: 'test-token' });
    
    // Initialize to create service instance
    await harness.init();
    
    // Get the mocked service instance (it's created inside onStart)
    mockService = (plugin as any).service;
    
    // Setup service mocks
    mockService.checkAuth.mockResolvedValue(true);
    mockService.sendMessage.mockResolvedValue({ ok: true, ts: '12345' });
  });

  it('should expose skill definition', () => {
    const definition = plugin.getSkillDefinition();
    expect(definition.name).toBe('slack');
    expect(definition.tools).toHaveLength(1);
    expect(definition.tools[0].name).toBe('send_message');
  });

  it('should execute send_message tool', async () => {
    const result = await harness.executeTool('send_message', { 
      channel: 'C12345', 
      text: 'Hello World' 
    });
    
    expect(mockService.sendMessage).toHaveBeenCalledWith({
      channel: 'C12345',
      text: 'Hello World'
    });
    expect(result).toEqual({ ok: true, ts: '12345' });
  });
});

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { NotionPlugin } from './plugin.js';
import { NotionService } from './service.js';
import { SkillTestHarness } from '../../../utils/skill-test-harness.js';

// Mock NotionService
vi.mock('./service.js');

describe('NotionPlugin Skill', () => {
  let plugin: NotionPlugin;
  let harness: SkillTestHarness;
  let mockService: any;

  beforeEach(async () => {
    vi.clearAllMocks();
    
    plugin = new NotionPlugin();
    harness = new SkillTestHarness(plugin);
    
    // Mock config
    harness.context.getConfig = vi.fn().mockReturnValue({ apiKey: 'test-key' });
    
    // Initialize
    await harness.init();
    
    // Get the mocked service instance
    mockService = (plugin as any).service;
    
    // Setup service mocks
    mockService.checkAuth.mockResolvedValue(true);
    mockService.search.mockResolvedValue({ results: [] });
    const samplePageId = '550e8400-e29b-41d4-a716-446655440000';
    mockService.getPage.mockResolvedValue({ id: samplePageId });
  });

  it('should expose skill definition', () => {
    const definition = plugin.getSkillDefinition();
    expect(definition.name).toBe('notion');
    expect(definition.tools).toHaveLength(2);
  });

  it('should execute search tool', async () => {
    const result = await harness.executeTool('search', { query: 'todo' });
    expect(mockService.search).toHaveBeenCalledWith('todo');
    expect(result).toEqual({ results: [] });
  });

  it('should execute get_page tool', async () => {
    const samplePageId = '550e8400-e29b-41d4-a716-446655440000';
    const result = await harness.executeTool('get_page', { pageId: samplePageId });
    expect(mockService.getPage).toHaveBeenCalledWith(samplePageId);
    expect(result).toEqual({ id: samplePageId });
  });
});

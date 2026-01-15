import { describe, it, expect, vi, beforeEach } from 'vitest';
import { GitHubPlugin } from './plugin.js';
import { GitHubService } from './service.js';
import { SkillTestHarness } from '../../../utils/skill-test-harness.js';

// Mock GitHubService
vi.mock('./service.js');

describe('GitHubPlugin Skill', () => {
  let plugin: GitHubPlugin;
  let harness: SkillTestHarness;
  let mockService: any;

  beforeEach(async () => {
    vi.clearAllMocks();
    
    plugin = new GitHubPlugin();
    harness = new SkillTestHarness(plugin);
    
    // Get the mocked service instance
    mockService = (plugin as any).service;
    
    // Setup service mocks
    mockService.checkAuth.mockResolvedValue(true);
    mockService.listIssues.mockResolvedValue([
      { number: 1, title: 'Test Issue' }
    ]);
    mockService.predictLabels.mockReturnValue(['bug']);
    
    await harness.init();
  });

  it('should expose skill definition', () => {
    const definition = plugin.getSkillDefinition();
    expect(definition.name).toBe('github');
    expect(definition.tools).toHaveLength(2);
  });

  it('should execute list_issues tool', async () => {
    const result = await harness.executeTool('list_issues', { repo: 'owner/repo' });
    expect(mockService.listIssues).toHaveBeenCalledWith('owner/repo');
    expect(result).toEqual([{ number: 1, title: 'Test Issue' }]);
  });

  it('should execute predict_labels tool', async () => {
    const result = await harness.executeTool('predict_labels', { 
      title: 'Crash on startup', 
      body: 'It fails with error' 
    });
    expect(mockService.predictLabels).toHaveBeenCalledWith('Crash on startup', 'It fails with error');
    expect(result).toEqual(['bug']);
  });
});

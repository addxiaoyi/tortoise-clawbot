import { describe, it, expect, vi, beforeEach } from 'vitest';
import { GitWorkflowPlugin } from './plugin.js';
import { SkillTestHarness } from '../../../utils/skill-test-harness.js';

vi.mock('./service.js', () => {
  class GitWorkflowService {
    start = vi.fn();
    stop = vi.fn();
    createBranch = vi.fn();
    checkoutBranch = vi.fn();
    commit = vi.fn();
  }
  return { GitWorkflowService };
});

describe('GitWorkflowPlugin Skill', () => {
  let plugin: GitWorkflowPlugin;
  let harness: SkillTestHarness;
  let mockServiceInstance: any;

  beforeEach(async () => {
    vi.clearAllMocks();
    
    plugin = new GitWorkflowPlugin();
    harness = new SkillTestHarness(plugin);
    
    // Mock config
    harness.context.getConfig = vi.fn().mockReturnValue({ repoPath: '.' });
    
    await harness.init();
    mockServiceInstance = (plugin as any).service;
    mockServiceInstance.createBranch.mockResolvedValue('Branch feature/test created');
    mockServiceInstance.checkoutBranch.mockResolvedValue('Checked out branch feature/test');
    mockServiceInstance.commit.mockResolvedValue('Committed changes: Test commit');
  });

  it('should expose skill definition', () => {
    const definition = plugin.getSkillDefinition();
    expect(definition.name).toBe('git-workflow');
    expect(definition.tools).toHaveLength(3);
  });

  it('should execute create_branch tool', async () => {
    const result = await harness.executeTool('create_branch', { name: 'feature/test' });
    expect(mockServiceInstance.createBranch).toHaveBeenCalledWith('feature/test');
    expect(result).toEqual({ result: 'Branch feature/test created' });
  });

  it('should execute checkout_branch tool', async () => {
    const result = await harness.executeTool('checkout_branch', { name: 'feature/test' });
    expect(mockServiceInstance.checkoutBranch).toHaveBeenCalledWith('feature/test');
    expect(result).toEqual({ result: 'Checked out branch feature/test' });
  });

  it('should execute commit_changes tool', async () => {
    const result = await harness.executeTool('commit_changes', { message: 'Test commit' });
    expect(mockServiceInstance.commit).toHaveBeenCalledWith('Test commit');
    expect(result).toEqual({ result: 'Committed changes: Test commit' });
  });
});


import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { GitHubPlugin } from './plugin';
import { PluginContext } from '../types';
import { GitHubService } from './service';

// Mock service
vi.mock('./service');

describe('GitHubPlugin', () => {
  let plugin: GitHubPlugin;
  let context: PluginContext;

  beforeEach(() => {
    context = {
      meta: { id: 'github', name: 'GitHub Plugin', version: '1.0.0' },
      logger: { debug: vi.fn(), info: vi.fn(), warn: vi.fn(), error: vi.fn() },
      storage: { getItem: vi.fn(), setItem: vi.fn(), removeItem: vi.fn(), clear: vi.fn() },
      events: { emit: vi.fn(), on: vi.fn(), off: vi.fn() },
      getConfig: <T extends import('../types').PluginConfig>(): T => ({}) as T
    };
    plugin = new GitHubPlugin();
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it('should initialize correctly', () => {
    plugin.onInit(context);
    expect(context.logger.info).toHaveBeenCalledWith('GitHub Plugin initialized');
  });

  it('should check auth on onStart', async () => {
    const mockCheckAuth = vi.fn().mockResolvedValue(true);
    vi.mocked(GitHubService.prototype.checkAuth).mockImplementation(mockCheckAuth);

    await plugin.onInit(context);
    await plugin.onStart();

    expect(mockCheckAuth).toHaveBeenCalled();
    expect(context.logger.info).toHaveBeenCalledWith('GitHub auth status: valid');
  });

  it('should warn if auth check fails on onStart', async () => {
    const mockCheckAuth = vi.fn().mockResolvedValue(false);
    vi.mocked(GitHubService.prototype.checkAuth).mockImplementation(mockCheckAuth);

    await plugin.onInit(context);
    await plugin.onStart();

    expect(context.logger.warn).toHaveBeenCalledWith('GitHub auth status: invalid or not logged in');
  });
});

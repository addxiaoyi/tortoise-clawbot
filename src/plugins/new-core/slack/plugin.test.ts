
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { SlackPlugin } from './plugin';
import { PluginContext } from '../types';
import { SlackService } from './service';

// Mock service
vi.mock('./service');

describe('SlackPlugin', () => {
  let plugin: SlackPlugin;
  let context: PluginContext;

  beforeEach(() => {
    context = {
      meta: { id: 'slack', name: 'Slack Plugin', version: '1.0.0' },
      logger: { debug: vi.fn(), info: vi.fn(), warn: vi.fn(), error: vi.fn() },
      storage: { getItem: vi.fn(), setItem: vi.fn(), removeItem: vi.fn(), clear: vi.fn() },
      events: { emit: vi.fn(), on: vi.fn(), off: vi.fn() },
      getConfig: <T extends import('../types').PluginConfig>(): T =>
        ({ token: 'mock-token' }) as unknown as T
    };
    plugin = new SlackPlugin();
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it('should initialize correctly', async () => {
    await plugin.onInit(context);
    expect(context.logger.info).toHaveBeenCalledWith('Slack Plugin initialized');
  });

  it('should check auth on onStart', async () => {
    const mockCheckAuth = vi.fn().mockResolvedValue(true);
    // @ts-ignore
    SlackService.mockImplementation(function() {
      return { checkAuth: mockCheckAuth };
    });

    await plugin.onInit(context);
    await plugin.onStart();

    expect(SlackService).toHaveBeenCalled();
    expect(mockCheckAuth).toHaveBeenCalled();
    expect(context.logger.info).toHaveBeenCalledWith('Slack auth status: valid');
  });

  it('should warn if config missing', async () => {
    // @ts-ignore
    context.getConfig = () => ({});
    await plugin.onInit(context);
    await plugin.onStart();
    expect(context.logger.warn).toHaveBeenCalledWith('Slack token not configured');
  });
});

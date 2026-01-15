
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { NotionPlugin } from './plugin';
import { PluginContext } from '../types';
import { NotionService } from './service';

// Mock service
vi.mock('./service');

describe('NotionPlugin', () => {
  let plugin: NotionPlugin;
  let context: PluginContext;

  beforeEach(() => {
    context = {
      meta: { id: 'notion', name: 'Notion Plugin', version: '1.0.0' },
      logger: { debug: vi.fn(), info: vi.fn(), warn: vi.fn(), error: vi.fn() },
      storage: { getItem: vi.fn(), setItem: vi.fn(), removeItem: vi.fn(), clear: vi.fn() },
      events: { emit: vi.fn(), on: vi.fn(), off: vi.fn() },
      getConfig: () => ({ apiKey: 'mock-key' } as any)
    } as unknown as PluginContext;
    plugin = new NotionPlugin();
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it('should initialize correctly', async () => {
    await plugin.onInit(context);
    expect(context.logger.info).toHaveBeenCalledWith('Notion Plugin initialized');
  });

  it('should check auth on onStart', async () => {
    const mockCheckAuth = vi.fn().mockResolvedValue(true);
    // @ts-ignore
    NotionService.mockImplementation(function() {
      return { checkAuth: mockCheckAuth };
    });

    await plugin.onInit(context);
    await plugin.onStart();

    expect(NotionService).toHaveBeenCalled();
    expect(mockCheckAuth).toHaveBeenCalled();
    expect(context.logger.info).toHaveBeenCalledWith('Notion auth status: valid');
  });

  it('should warn if config missing', async () => {
    // @ts-ignore
    context.getConfig = () => ({});
    await plugin.onInit(context);
    await plugin.onStart();
    expect(context.logger.warn).toHaveBeenCalledWith('Notion API key not configured');
  });
});

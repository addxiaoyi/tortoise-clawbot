
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { CanvasPlugin } from './plugin';
import { PluginContext } from '../types';
import { Container } from '../container';
import http from 'http';

// Mock dependencies
vi.mock('http');
vi.mock('fs/promises');

describe('CanvasPlugin', () => {
  let plugin: CanvasPlugin;
  let context: PluginContext;
  let container: Container;

  beforeEach(() => {
    container = new Container();
    context = {
      meta: { id: 'canvas', name: 'Canvas Plugin', version: '1.0.0' },
      logger: { debug: vi.fn(), info: vi.fn(), warn: vi.fn(), error: vi.fn() },
      storage: { getItem: vi.fn(), setItem: vi.fn(), removeItem: vi.fn(), clear: vi.fn() },
      events: { emit: vi.fn(), on: vi.fn(), off: vi.fn() } as any,
      getConfig: <T>() => ({ port: 18793, root: '/tmp/canvas' } as unknown as T)
    } as unknown as PluginContext;
    plugin = new CanvasPlugin();
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it('should initialize correctly', () => {
    plugin.onInit(context);
    expect(context.logger.info).toHaveBeenCalledWith('Canvas Plugin initialized');
  });

  it('should start http server on onStart', async () => {
    const mockServer = {
      listen: vi.fn((port, _host, cb) => cb && cb()), // (port, host?, callback)
      close: vi.fn(),
      on: vi.fn()
    };
    vi.mocked(http.createServer).mockReturnValue(mockServer as any);

    await plugin.onInit(context);
    await plugin.onStart();

    expect(http.createServer).toHaveBeenCalled();
    expect(mockServer.listen).toHaveBeenCalledWith(
      18793,
      '127.0.0.1',
      expect.any(Function),
    );
  });

  it('should stop http server on onStop', async () => {
    const mockServer = {
      listen: vi.fn((port, _host, cb) => cb && cb()),
      close: vi.fn((cb) => cb && cb()), // Ensure callback is called
      on: vi.fn()
    };
    vi.mocked(http.createServer).mockReturnValue(mockServer as any);

    await plugin.onInit(context);
    await plugin.onStart();
    await plugin.onStop();

    expect(mockServer.close).toHaveBeenCalled();
  });
});

import { describe, it, expect, vi } from 'vitest';
import { PluginRegistry, PluginDefinition } from './registry.js';
import { Container } from './container.js';
import { PluginContext } from './types.js';

describe('PluginRegistry', () => {
  const container = new Container();
  const mockContext = {
    logger: {
      info: vi.fn(),
      warn: vi.fn(),
      debug: vi.fn(),
      error: vi.fn()
    },
    // Mock for sandbox
    fs: {}
  } as unknown as PluginContext;

  it('should register and retrieve plugins', () => {
    const registry = new PluginRegistry(container);
    const plugin: PluginDefinition = {
      meta: { id: 'test', name: 'Test', version: '1.0' },
      lifecycle: { onInit: vi.fn() }
    };

    registry.register(plugin);
    expect(registry.get('test')).toBe(plugin);
    expect(() => registry.register(plugin)).toThrow('Plugin already registered');
  });

  it('should init and start plugins', async () => {
    const registry = new PluginRegistry(container);
    const lifecycle = {
      onInit: vi.fn(),
      onStart: vi.fn(),
      onStop: vi.fn()
    };
    
    registry.register({
      meta: { id: 'lifecycle-test', name: 'Test', version: '1.0' },
      lifecycle
    });

    await registry.initAll(mockContext);
    expect(lifecycle.onInit).toHaveBeenCalled();

    await registry.startAll();
    expect(lifecycle.onStart).toHaveBeenCalled();

    await registry.stopAll();
    expect(lifecycle.onStop).toHaveBeenCalled();
  });
});

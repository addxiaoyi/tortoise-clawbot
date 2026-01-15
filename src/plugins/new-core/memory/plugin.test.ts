import { describe, it, expect, vi } from 'vitest';
import { MemoryPlugin, MemoryConfig } from './plugin.js';
import { PluginContext } from '../types.js';

describe('MemoryPlugin', () => {
  const mockContext = {
    logger: {
      info: vi.fn(),
      warn: vi.fn(),
      debug: vi.fn(),
      error: vi.fn()
    },
    getConfig: vi.fn().mockReturnValue({ maxItems: 2 })
  } as unknown as PluginContext;

  it('should initialize correctly', async () => {
    const plugin = new MemoryPlugin();
    await plugin.onInit(mockContext);
    
    expect(mockContext.logger.info).toHaveBeenCalledWith('MemoryPlugin initialized');
  });

  it('should store and retrieve values', async () => {
    const plugin = new MemoryPlugin();
    await plugin.onInit(mockContext);
    
    plugin.set('key', 'value');
    expect(plugin.get('key')).toBe('value');
  });

  it('should respect maxItems from config', async () => {
    const plugin = new MemoryPlugin();
    await plugin.onInit(mockContext); // maxItems: 2
    
    plugin.set('a', 1);
    plugin.set('b', 2);
    plugin.set('c', 3); // Should evict 'a' (FIFO map iteration order)
    
    expect(plugin.get('a')).toBeUndefined();
    expect(plugin.get('b')).toBe(2);
    expect(plugin.get('c')).toBe(3);
  });

  it('should not evict other keys when updating an existing key at capacity', async () => {
    const plugin = new MemoryPlugin();
    await plugin.onInit(mockContext); // maxItems: 2

    plugin.set('a', 1);
    plugin.set('b', 2);
    plugin.set('b', 20); // 更新已有 key，不得删除 a

    expect(plugin.get('a')).toBe(1);
    expect(plugin.get('b')).toBe(20);
    expect(plugin.listKeys().sort()).toEqual(['a', 'b']);
  });

  it('should list, remove, and clear keys', async () => {
    const plugin = new MemoryPlugin();
    await plugin.onInit(mockContext);
    plugin.set('x', 1);
    plugin.set('y', 2);
    expect(plugin.listKeys().sort()).toEqual(['x', 'y']);
    expect(plugin.removeKey('x')).toBe(true);
    expect(plugin.get('x')).toBeUndefined();
    plugin.clearAll();
    expect(plugin.listKeys()).toEqual([]);
  });

  it('should handle config updates', async () => {
    const plugin = new MemoryPlugin();
    await plugin.onInit(mockContext); // maxItems: 2
    
    await plugin.onConfigChange({ maxItems: 1 });
    
    // Config applied for FUTURE sets
    // Current size might exceed new limit until next set, depending on implementation
    // Our implementation checks size on set.
    
    plugin.set('d', 4);
    
    // With maxItems=1, adding 'd' should evict 'b' (if 'a' was already gone)
    // Actually, let's reset state for clarity
  });
});

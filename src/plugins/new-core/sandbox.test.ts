import { describe, it, expect, vi } from 'vitest';
import { PluginSandbox } from './sandbox.js';
import { PluginContext } from './types.js';

describe('PluginSandbox', () => {
  const mockContext = {
    logger: {
      info: vi.fn(),
      warn: vi.fn(),
      debug: vi.fn(),
      error: vi.fn()
    },
    // Mock sensitive props
    fs: { readFileSync: vi.fn() },
    net: { createConnection: vi.fn() },
    // Normal prop
    getConfig: vi.fn()
  } as unknown as PluginContext;

  it('should allow allowed access', () => {
    const sandbox = PluginSandbox.create(mockContext, { allowFs: true, allowNet: true });
    
    // Access should not throw
    expect((sandbox as any).fs).toBeDefined();
    expect((sandbox as any).net).toBeDefined();
    expect(sandbox.logger).toBeDefined();
  });

  it('should deny denied access', () => {
    const sandbox = PluginSandbox.create(mockContext, { allowFs: false, allowNet: false });
    
    expect(() => (sandbox as any).fs).toThrow('FileSystem access denied by sandbox');
    expect(() => (sandbox as any).net).toThrow('Network access denied by sandbox');
  });

  it('should pass through safe properties', () => {
    const sandbox = PluginSandbox.create(mockContext, {});
    
    expect(sandbox.logger).toBe(mockContext.logger);
    sandbox.getConfig();
    expect(mockContext.getConfig).toHaveBeenCalled();
  });
});

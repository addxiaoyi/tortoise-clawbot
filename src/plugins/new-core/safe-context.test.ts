import { describe, it, expect, vi } from 'vitest';
import { SafePluginContext } from './safe-context.js';
import { PluginContext } from './types.js';

describe('SafePluginContext', () => {
  const mockEvents = {
    emit: vi.fn(),
    on: vi.fn().mockReturnValue(() => {}), // Returns unsubscribe
    off: vi.fn()
  };

  const mockBaseContext = {
    meta: { id: 'test' },
    logger: { info: vi.fn() },
    events: mockEvents,
    getConfig: vi.fn()
  } as unknown as PluginContext;

  it('should wrap event handlers and catch errors', () => {
    const errorBoundary = vi.fn();
    const safeCtx = new SafePluginContext(mockBaseContext, errorBoundary);

    const handler = () => { throw new Error('Boom'); };
    
    // Register handler
    safeCtx.events.on('test-event', handler);
    
    // Get the wrapped handler that was actually passed to baseContext.events.on
    // mockEvents.on.mock.calls is an array of calls, each call is an array of args
    // [[event, handler]]
    const wrappedHandler = (mockEvents.on as any).mock.calls[0][1];
    
    // Execute wrapped handler
    wrappedHandler({ payload: 1 });
    
    expect(errorBoundary).toHaveBeenCalledWith(expect.any(Error));
    expect(errorBoundary.mock.calls[0][0].message).toBe('Boom');
  });

  it('should delegate property access', () => {
    const safeCtx = new SafePluginContext(mockBaseContext, vi.fn());
    
    expect(safeCtx.meta).toBe(mockBaseContext.meta);
    expect(safeCtx.logger).toBe(mockBaseContext.logger);
    
    safeCtx.getConfig();
    expect(mockBaseContext.getConfig).toHaveBeenCalled();
  });
});

import { describe, it, expect, vi } from 'vitest';
import { PluginRegistry } from './registry.js';
import { Container } from './container.js';
import { MemoryPlugin } from './memory/plugin.js';
import { GatewayPlugin } from './gateway/plugin.js';
import { PluginContext, PluginEventBus } from './types.js';
import http from 'node:http';

describe('MCP Integration Test', () => {
  it('should run multiple plugins together and interact via events', async () => {
    const container = new Container();
    const registry = new PluginRegistry(container);
    
    // Simple event bus
    const handlers: Record<string, ((p: any) => void)[]> = {};
    const events: PluginEventBus = {
      emit: (event, payload) => {
        handlers[event]?.forEach(h => h(payload));
      },
      on: (event, handler) => {
        if (!handlers[event]) handlers[event] = [];
        handlers[event].push(handler);
        return () => {
          handlers[event] = handlers[event].filter(h => h !== handler);
        };
      },
      off: (event, handler) => {
        if (handlers[event]) {
          handlers[event] = handlers[event].filter(h => h !== handler);
        }
      }
    };

    const baseContext = {
      meta: { id: 'host', name: 'Host', version: '1.0' },
      logger: {
        info: vi.fn(),
        warn: vi.fn(),
        error: vi.fn(),
        debug: vi.fn()
      },
      events,
      storage: {
        getItem: vi.fn(),
        setItem: vi.fn(),
        removeItem: vi.fn(),
        clear: vi.fn()
      },
      // Corrected getConfig to return typed config
      getConfig: () => {
          // This mock returns a generic object that matches any T
          return { port: 3002, host: '127.0.0.1', maxItems: 10 } as any;
      }
    } as unknown as PluginContext;

    // Register Memory Plugin
    const memoryPlugin = new MemoryPlugin();
    registry.register({
      meta: { id: 'memory', name: 'Memory', version: '1.0' },
      lifecycle: memoryPlugin
    });

    // Register Gateway Plugin
    const gatewayPlugin = new GatewayPlugin();
    registry.register({
      meta: { id: 'gateway', name: 'Gateway', version: '1.0' },
      lifecycle: gatewayPlugin
    });

    // Link them: Memory plugin listens to gateway requests
    events.on('gateway:request', ({ req, res }) => {
      if (req.url === '/memory/set') {
        memoryPlugin.set('test-key', 'test-value');
        res.writeHead(200);
        res.end('Stored');
      } else if (req.url === '/memory/get') {
        const val = memoryPlugin.get('test-key');
        res.writeHead(200);
        res.end(val || 'not found');
      }
    });

    // Lifecycle
    await registry.initAll(baseContext);
    await registry.startAll();

    // Verify interaction
    const setRes = await new Promise<string>((resolve, reject) => {
      const req = http.get('http://127.0.0.1:3002/memory/set', (res) => {
        let data = '';
        res.on('data', c => data += c);
        res.on('end', () => resolve(data));
      });
      req.on('error', reject);
    });
    expect(setRes).toBe('Stored');

    const getRes = await new Promise<string>((resolve, reject) => {
      const req = http.get('http://127.0.0.1:3002/memory/get', (res) => {
        let data = '';
        res.on('data', c => data += c);
        res.on('end', () => resolve(data));
      });
      req.on('error', reject);
    });
    expect(getRes).toBe('test-value');

    // Cleanup
    await registry.stopAll();
  });
});

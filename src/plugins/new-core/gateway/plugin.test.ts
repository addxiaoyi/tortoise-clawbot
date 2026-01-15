import { describe, it, expect, vi, afterEach } from 'vitest';
import { GatewayPlugin } from './plugin.js';
import { PluginContext } from '../types.js';
import http from 'node:http';
import { reserveFreePort } from '../../../bridge/free-port.js';

describe('GatewayPlugin', () => {
  let plugin: GatewayPlugin;
  let port: number;
  let mockContext: PluginContext;

  afterEach(async () => {
    if (plugin) {
      await plugin.onStop();
    }
  });

  it('should initialize and start', async () => {
    port = await reserveFreePort();
    mockContext = {
      meta: { version: '1.0.0' },
      logger: {
        info: vi.fn(),
        warn: vi.fn(),
        debug: vi.fn(),
        error: vi.fn(),
      },
      events: {
        emit: vi.fn(),
      },
      getConfig: vi.fn().mockReturnValue({ port, host: '127.0.0.1' }),
    } as unknown as PluginContext;

    plugin = new GatewayPlugin();
    await plugin.onInit(mockContext);
    await plugin.onStart();

    expect(mockContext.logger.info).toHaveBeenCalledWith(
      expect.stringContaining(`listening on port ${port}`),
    );
  });

  it('should handle health check', async () => {
    port = await reserveFreePort();
    mockContext = {
      meta: { version: '1.0.0' },
      logger: {
        info: vi.fn(),
        warn: vi.fn(),
        debug: vi.fn(),
        error: vi.fn(),
      },
      events: {
        emit: vi.fn(),
      },
      getConfig: vi.fn().mockReturnValue({ port, host: '127.0.0.1' }),
    } as unknown as PluginContext;

    plugin = new GatewayPlugin();
    await plugin.onInit(mockContext);
    await plugin.onStart();

    const { body, headers } = await new Promise<{
      body: string;
      headers: http.IncomingHttpHeaders;
    }>((resolve) => {
      http.get(`http://127.0.0.1:${port}/health`, (res) => {
        let data = '';
        res.on('data', (chunk) => (data += chunk));
        res.on('end', () => resolve({ body: data, headers: res.headers }));
      });
    });

    const json = JSON.parse(body);
    expect(json.status).toBe('ok');
    expect(json.version).toBe('1.0.0');
    expect(headers['x-content-type-options']).toBe('nosniff');
    expect(headers['referrer-policy']).toBe('no-referrer');
    expect(headers['x-frame-options']).toBe('SAMEORIGIN');
  });

  it('should emit event for unknown routes', async () => {
    const localPort = await reserveFreePort();
    plugin = new GatewayPlugin();
    // Use a fresh mock context to avoid interference from other tests
    const mockContextLocal = {
      meta: { version: '1.0.0' },
      logger: {
        info: vi.fn(),
        warn: vi.fn(),
        debug: vi.fn(),
        error: vi.fn(),
      },
      events: {
        emit: vi.fn(),
      },
      getConfig: vi.fn().mockReturnValue({ port: localPort, host: '127.0.0.1' }),
    } as unknown as PluginContext;

    await plugin.onInit(mockContextLocal);
    await plugin.onStart();

    const eventPromise = new Promise<void>((resolve) => {
      (mockContextLocal.events.emit as any).mockImplementation((event: string, payload: any) => {
        if (event === 'gateway:request') {
          payload.res.end('Handled');
          resolve();
        }
      });
    });

    const req = http.get(`http://127.0.0.1:${localPort}/other`, (res) => {
       res.on('data', () => {});
    });
    
    req.on('error', () => {});

    await Promise.race([
      eventPromise,
      new Promise((_, reject) => setTimeout(() => reject(new Error('Event timeout')), 2000))
    ]);

    expect(mockContextLocal.events.emit).toHaveBeenCalledWith('gateway:request', expect.any(Object));
  });

  it('should respond to OPTIONS with CORS when cors enabled', async () => {
    const localPort = await reserveFreePort();
    const ctx = {
      meta: { version: '1.0.0' },
      logger: {
        info: vi.fn(),
        warn: vi.fn(),
        debug: vi.fn(),
        error: vi.fn(),
      },
      events: { emit: vi.fn() },
      getConfig: vi.fn().mockReturnValue({
        port: localPort,
        host: '127.0.0.1',
        cors: true,
      }),
    } as unknown as PluginContext;

    const gw = new GatewayPlugin();
    await gw.onInit(ctx);
    await gw.onStart();

    const { statusCode, headers } = await new Promise<{
      statusCode?: number;
      headers: http.IncomingHttpHeaders;
    }>((resolve, reject) => {
      const req = http.request(
        {
          hostname: '127.0.0.1',
          port: localPort,
          path: '/any',
          method: 'OPTIONS',
        },
        (res) => {
          resolve({ statusCode: res.statusCode, headers: res.headers });
        },
      );
      req.on('error', reject);
      req.end();
    });

    expect(statusCode).toBe(204);
    expect(headers['access-control-allow-origin']).toBe('*');
    expect(headers['x-content-type-options']).toBe('nosniff');
    expect(headers['referrer-policy']).toBe('no-referrer');
    expect(headers['x-frame-options']).toBe('SAMEORIGIN');

    await gw.onStop();
  });

  it('should reflect Origin only when corsAllowOrigins matches', async () => {
    const localPort = await reserveFreePort();
    const allowed = 'http://127.0.0.1:9999';
    const ctx = {
      meta: { version: '1.0.0' },
      logger: {
        info: vi.fn(),
        warn: vi.fn(),
        debug: vi.fn(),
        error: vi.fn(),
      },
      events: { emit: vi.fn() },
      getConfig: vi.fn().mockReturnValue({
        port: localPort,
        host: '127.0.0.1',
        cors: true,
        corsAllowOrigins: [allowed],
      }),
    } as unknown as PluginContext;

    const gw = new GatewayPlugin();
    await gw.onInit(ctx);
    await gw.onStart();

    const headersMatch = await new Promise<http.IncomingHttpHeaders>((resolve, reject) => {
      const req = http.request(
        {
          hostname: '127.0.0.1',
          port: localPort,
          path: '/any',
          method: 'OPTIONS',
          headers: { Origin: allowed },
        },
        (res) => {
          res.resume();
          resolve(res.headers);
        },
      );
      req.on('error', reject);
      req.end();
    });
    expect(headersMatch['access-control-allow-origin']).toBe(allowed);

    const headersNoMatch = await new Promise<http.IncomingHttpHeaders>((resolve, reject) => {
      const req = http.request(
        {
          hostname: '127.0.0.1',
          port: localPort,
          path: '/any',
          method: 'OPTIONS',
          headers: { Origin: 'https://evil.example' },
        },
        (res) => {
          res.resume();
          resolve(res.headers);
        },
      );
      req.on('error', reject);
      req.end();
    });
    expect(headersNoMatch['access-control-allow-origin']).toBeUndefined();

    await gw.onStop();
  });
});

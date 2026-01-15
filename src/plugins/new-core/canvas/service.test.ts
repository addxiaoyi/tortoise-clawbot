import { describe, it, expect, vi, afterEach } from 'vitest';
import fs from 'node:fs/promises';
import os from 'node:os';
import path from 'node:path';
import http from 'node:http';
import { CanvasService } from './service.js';
import { reserveFreePort } from '../../../bridge/free-port.js';
import type { PluginLogger } from '../types.js';

function makeLogger(): PluginLogger {
  return {
    debug: vi.fn(),
    info: vi.fn(),
    warn: vi.fn(),
    error: vi.fn(),
  };
}

describe('CanvasService HTTP', () => {
  let root: string | undefined;
  let service: CanvasService | undefined;

  afterEach(async () => {
    if (service) {
      await service.stop().catch(() => {});
      service = undefined;
    }
    if (root) {
      await fs.rm(root, { recursive: true, force: true });
      root = undefined;
    }
  });

  it('serves text files with security and Content-Type headers', async () => {
    root = await fs.mkdtemp(path.join(os.tmpdir(), 'canvas-http-'));
    await fs.writeFile(path.join(root, 'page.html'), '<html>ok</html>', 'utf-8');

    const port = await reserveFreePort();
    service = new CanvasService(
      { root, port, host: '127.0.0.1' },
      makeLogger(),
    );
    await service.start();

    const result = await new Promise<{
      statusCode?: number;
      headers: http.IncomingHttpHeaders;
      body: string;
    }>((resolve, reject) => {
      http
        .get(`http://127.0.0.1:${port}/page.html`, (res) => {
          const chunks: Buffer[] = [];
          res.on('data', (c) => chunks.push(c));
          res.on('end', () => {
            resolve({
              statusCode: res.statusCode,
              headers: res.headers,
              body: Buffer.concat(chunks).toString('utf-8'),
            });
          });
        })
        .on('error', reject);
    });

    expect(result.statusCode).toBe(200);
    expect(result.headers['content-type']).toMatch(/text\/html/);
    expect(result.headers['x-content-type-options']).toBe('nosniff');
    expect(result.headers['referrer-policy']).toBe('no-referrer');
    expect(result.headers['x-frame-options']).toBe('SAMEORIGIN');
    expect(result.body).toContain('ok');
  });

  it('serves binary without forcing utf-8 string corruption', async () => {
    root = await fs.mkdtemp(path.join(os.tmpdir(), 'canvas-bin-'));
    const png = Buffer.from([
      0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, 0x00, 0x00, 0x00, 0x0d,
    ]);
    await fs.writeFile(path.join(root, 'x.png'), png);

    const port = await reserveFreePort();
    service = new CanvasService(
      { root, port, host: '127.0.0.1' },
      makeLogger(),
    );
    await service.start();

    const buf = await new Promise<Buffer>((resolve, reject) => {
      http
        .get(`http://127.0.0.1:${port}/x.png`, (res) => {
          const chunks: Buffer[] = [];
          res.on('data', (c) => chunks.push(c));
          res.on('end', () => resolve(Buffer.concat(chunks)));
        })
        .on('error', reject);
    });

    expect(buf.equals(png)).toBe(true);
  });
});

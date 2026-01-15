import { describe, it, expect } from 'vitest';
import { applyHttpServerTimeouts } from './server-timeouts.js';

describe('applyHttpServerTimeouts', () => {
  it('sets requestTimeout and caps headersTimeout below requestTimeout', () => {
    const server = { requestTimeout: 0, headersTimeout: 0 };
    applyHttpServerTimeouts(server, 60_000, 30_000);
    expect(server.requestTimeout).toBe(60_000);
    expect(server.headersTimeout).toBe(30_000);
  });

  it('caps headers when header budget exceeds request budget', () => {
    const server = { requestTimeout: 0, headersTimeout: 0 };
    applyHttpServerTimeouts(server, 5000, 30_000);
    expect(server.requestTimeout).toBe(5000);
    expect(server.headersTimeout).toBe(4999);
  });

  it('disables request timeout when requestTimeoutMs is 0', () => {
    const server = { requestTimeout: 999, headersTimeout: 999 };
    applyHttpServerTimeouts(server, 0, 20_000);
    expect(server.requestTimeout).toBe(0);
    expect(server.headersTimeout).toBe(20_000);
  });

  it('clears headers timeout when headersTimeoutMs is 0', () => {
    const server = { requestTimeout: 0, headersTimeout: 0 };
    applyHttpServerTimeouts(server, 10_000, 0);
    expect(server.requestTimeout).toBe(10_000);
    expect(server.headersTimeout).toBe(0);
  });
});

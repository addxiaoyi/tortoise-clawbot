import type { Server } from 'node:http';

/**
 * 为 HTTP Server 设置请求与请求头超时（Node 要求 headersTimeout 小于 requestTimeout，当二者均 > 0 时）。
 */
export function applyHttpServerTimeouts(
  server: Pick<Server, 'requestTimeout' | 'headersTimeout'>,
  requestTimeoutMs: number,
  headersTimeoutMs: number,
): void {
  server.requestTimeout = requestTimeoutMs > 0 ? requestTimeoutMs : 0;
  if (headersTimeoutMs > 0) {
    const cap =
      requestTimeoutMs > 0 ? Math.max(1, requestTimeoutMs - 1) : headersTimeoutMs;
    server.headersTimeout = Math.min(headersTimeoutMs, cap);
  } else {
    server.headersTimeout = 0;
  }
}

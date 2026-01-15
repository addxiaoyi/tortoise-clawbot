
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { FETCH_NO_REDIRECT } from '../../../utils/fetch-safe.js';
import { SlackService } from './service';
import { SlackConfig } from './types';

// Mock fetch globally
const mockFetch = vi.fn();
global.fetch = mockFetch;

describe('SlackService', () => {
  let service: SlackService;
  const config: SlackConfig = { token: 'xoxb-token', defaultChannel: 'C123' };

  beforeEach(() => {
    service = new SlackService(config);
    mockFetch.mockReset();
  });

  afterEach(() => {
    vi.useRealTimers();
    vi.clearAllMocks();
  });

  it('should send message successfully', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => ({ ok: true, ts: '1234.5678' })
    });

    const result = await service.sendMessage({ channel: 'C123', text: 'Hello' });
    expect(result.ok).toBe(true);
    expect(result.ts).toBe('1234.5678');
    expect(mockFetch.mock.calls[0]?.[1]).toMatchObject(FETCH_NO_REDIRECT);
  });

  it('should throw on API error (after retries)', async () => {
    // Mock fetch to always return error response
    mockFetch.mockResolvedValue({
      ok: true,
      json: async () => ({ ok: false, error: 'invalid_auth' })
    });

    vi.useFakeTimers();
    const settled = expect(
      service.sendMessage({ channel: 'C123', text: 'Hello' }),
    ).rejects.toThrow('Slack API error: invalid_auth');
    await vi.runAllTimersAsync();
    await settled;

    expect(mockFetch.mock.calls.length).toBeGreaterThan(1);
    vi.useRealTimers();
  });

  it('should check auth successfully', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => ({ ok: true, user_id: 'U123' })
    });

    const isValid = await service.checkAuth();
    expect(isValid).toBe(true);
  });

  it('should return false on auth failure', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => ({ ok: false, error: 'invalid_auth' })
    });

    const isValid = await service.checkAuth();
    expect(isValid).toBe(false);
  });
});

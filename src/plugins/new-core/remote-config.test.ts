import { describe, it, expect, vi, afterEach } from 'vitest';
import { FETCH_NO_REDIRECT } from '../../utils/fetch-safe.js';
import { HttpRemoteConfigProvider } from './remote-config.js';

// Mock fetch
const mockFetch = vi.fn();
global.fetch = mockFetch;

describe('HttpRemoteConfigProvider', () => {
  let provider: HttpRemoteConfigProvider;

  afterEach(() => {
    if (provider) {
      provider.stopPolling();
    }
    vi.clearAllMocks();
  });

  it('should fetch config', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => ({ foo: 'bar' })
    });

    provider = new HttpRemoteConfigProvider({ url: 'https://example.com/config' });
    const config = await provider.fetch();

    expect(config).toEqual({ foo: 'bar' });
    expect(mockFetch).toHaveBeenCalledWith('https://example.com/config', {
      ...FETCH_NO_REDIRECT,
      headers: {},
    });
  });

  it('should handle fetch errors', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: false,
      status: 404,
      statusText: 'Not Found'
    });

    provider = new HttpRemoteConfigProvider({ url: 'https://example.com/config' });
    await expect(provider.fetch()).rejects.toThrow('Failed to fetch config: 404 Not Found');
  });

  it('should reject http URL when allowInsecureHttp is not set', () => {
    expect(
      () => new HttpRemoteConfigProvider({ url: 'http://example.com/config' }),
    ).toThrow(/HTTPS/);
  });

  it('should allow http URL when allowInsecureHttp is true', () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => ({ ok: true }),
    });
    provider = new HttpRemoteConfigProvider({
      url: 'http://127.0.0.1:9/config',
      allowInsecureHttp: true,
    });
    return expect(provider.fetch()).resolves.toEqual({ ok: true });
  });
});

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { FETCH_NO_REDIRECT } from '../../../utils/fetch-safe.js';
import { NotionService } from './service.js';

const mockFetch = vi.fn();
global.fetch = mockFetch;

describe('NotionService', () => {
  beforeEach(() => {
    mockFetch.mockReset();
  });

  it('calls search with no redirect follow', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => ({ object: 'list', results: [] }),
    });
    const service = new NotionService({
      apiKey: 'secret',
      version: '2022-06-28',
    });
    await service.search('hello');
    expect(mockFetch.mock.calls[0]?.[1]).toMatchObject({
      ...FETCH_NO_REDIRECT,
      method: 'POST',
    });
  });

  it('rejects invalid page id for getPage', async () => {
    const service = new NotionService({ apiKey: 'secret' });
    await expect(service.getPage('../etc/passwd')).rejects.toThrow(/Invalid/);
  });
});

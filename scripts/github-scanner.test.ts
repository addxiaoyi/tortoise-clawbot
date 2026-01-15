import { describe, it, expect, vi } from 'vitest';
import { FETCH_NO_REDIRECT } from '../src/utils/fetch-safe.js';
import { scanGitHub } from './github-scanner.js';
import fs from 'node:fs';

// Mock dependencies
vi.mock('node:fs');
vi.mock('global', () => ({
  fetch: vi.fn()
}));

describe('GitHub Scanner', () => {
  it('should construct queries correctly', async () => {
    // Mock FS
    const mockKeywords = {
      categories: {
        skill: {
          keywords: ['test-keyword'],
          tags: ['test']
        }
      },
      criteria: {
        minStars: 100,
        lastUpdatedDays: 30,
        languages: ['TypeScript']
      }
    };
    
    vi.spyOn(fs, 'existsSync').mockReturnValue(true);
    vi.spyOn(fs, 'readFileSync').mockReturnValue(JSON.stringify(mockKeywords));
    
    // Mock Fetch
    const mockFetch = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({ items: [] })
    });
    global.fetch = mockFetch;
    
    await scanGitHub({ limitPerCategory: 1, rateLimitDelayMs: 0 });

    expect(mockFetch).toHaveBeenCalled();
    const [url, init] = mockFetch.mock.calls[0] as [string, RequestInit];
    expect(url).toMatch(/api\.github\.com\/search\/repositories/);
    expect(url).toContain(encodeURIComponent('test-keyword'));
    expect(url).toMatch(/stars%3A%3E100/);
    expect(init).toMatchObject(FETCH_NO_REDIRECT);
    expect(init?.headers).toMatchObject({
      'User-Agent': 'OpenClaw-Optimization-Scanner',
    });
  });
});

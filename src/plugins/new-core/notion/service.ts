import { NotionConfig, NotionSearchResponse, NotionPage } from './types';
import { withBackoff } from '../../../utils/backoff';
import { FETCH_NO_REDIRECT } from '../../../utils/fetch-safe.js';
import { PluginLogger } from '../types';

/** Notion 页面 ID：32 位十六进制，可带标准连字符。 */
function assertValidNotionPageId(pageId: string): void {
  const raw = pageId.trim();
  if (!raw || raw.includes('\0') || /[/\\?#]/.test(raw)) {
    throw new Error('Invalid Notion page id');
  }
  const compact = raw.replace(/-/g, '');
  if (!/^[0-9a-f]{32}$/i.test(compact)) {
    throw new Error('Invalid Notion page id format');
  }
}

export class NotionService {
  private config: NotionConfig;
  private logger?: PluginLogger;
  private baseUrl = 'https://api.notion.com/v1';

  constructor(config: NotionConfig, logger?: PluginLogger) {
    this.config = config;
    this.logger = logger;
  }

  private get headers() {
    return {
      'Authorization': `Bearer ${this.config.apiKey}`,
      'Notion-Version': this.config.version || '2022-06-28',
      'Content-Type': 'application/json'
    };
  }

  public async search(query: string): Promise<NotionSearchResponse> {
    if (!this.config.apiKey) throw new Error('Notion API key missing');

    return withBackoff(async () => {
      const response = await fetch(`${this.baseUrl}/search`, {
        ...FETCH_NO_REDIRECT,
        method: 'POST',
        headers: this.headers,
        body: JSON.stringify({ query }),
      });

      if (!response.ok) {
        throw new Error(`Notion API error: ${response.status} ${response.statusText}`);
      }

      return await response.json() as NotionSearchResponse;
    });
  }

  public async getPage(pageId: string): Promise<NotionPage> {
    if (!this.config.apiKey) throw new Error('Notion API key missing');
    assertValidNotionPageId(pageId);

    return withBackoff(async () => {
      const response = await fetch(`${this.baseUrl}/pages/${pageId.trim()}`, {
        ...FETCH_NO_REDIRECT,
        method: 'GET',
        headers: this.headers,
      });

      if (!response.ok) {
        throw new Error(`Notion API error: ${response.status} ${response.statusText}`);
      }

      return await response.json() as NotionPage;
    });
  }

  public async checkAuth(): Promise<boolean> {
    if (!this.config.apiKey) return false;
    try {
      // Check auth by listing users (lightweight)
      const response = await fetch(`${this.baseUrl}/users/me`, {
        ...FETCH_NO_REDIRECT,
        method: 'GET',
        headers: this.headers,
      });
      return response.ok;
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : String(e);
      this.logger?.error(`Notion auth check failed: ${msg}`);
      return false;
    }
  }
}

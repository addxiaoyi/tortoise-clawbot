import { EventEmitter } from 'node:events';
import { FETCH_NO_REDIRECT } from '../../utils/fetch-safe.js';

export interface RemoteConfigProvider {
  /**
   * Fetch configuration from remote source.
   */
  fetch(): Promise<Record<string, any>>;
  
  /**
   * Subscribe to configuration changes.
   */
  on(event: 'change', listener: (config: Record<string, any>) => void): this;
  on(event: 'error', listener: (error: Error) => void): this;
}

export interface HttpConfigProviderOptions {
  url: string;
  headers?: Record<string, string>;
  pollInterval?: number;
  /**
   * 为 true 时允许 `http:` URL（仅建议在本地或隔离网络使用）。
   * 默认 false：要求 `https:`，降低中间人窃听/篡改配置的风险。
   */
  allowInsecureHttp?: boolean;
}

function assertRemoteConfigUrl(urlString: string, allowInsecureHttp: boolean): void {
  let u: URL;
  try {
    u = new URL(urlString);
  } catch {
    throw new Error(`Invalid remote config URL: ${urlString}`);
  }
  if (u.protocol === 'https:') {
    return;
  }
  if (u.protocol === 'http:' && allowInsecureHttp) {
    return;
  }
  throw new Error(
    'Remote config URL must use HTTPS, or set allowInsecureHttp: true for development-only HTTP endpoints',
  );
}

export class HttpRemoteConfigProvider extends EventEmitter implements RemoteConfigProvider {
  private url: string;
  private headers: Record<string, string>;
  private pollInterval: number;
  private timer: any = null; // Use any for NodeJS.Timeout vs window.Timer compatibility if needed
  private currentConfig: string = ''; // Store as string for easy comparison

  constructor(options: HttpConfigProviderOptions) {
    super();
    assertRemoteConfigUrl(options.url, options.allowInsecureHttp === true);
    this.url = options.url;
    this.headers = options.headers || {};
    this.pollInterval = options.pollInterval || 60000;
  }

  public async fetch(): Promise<Record<string, any>> {
    try {
      const response = await fetch(this.url, {
        ...FETCH_NO_REDIRECT,
        headers: this.headers,
      });

      if (!response.ok) {
        throw new Error(`Failed to fetch config: ${response.status} ${response.statusText}`);
      }

      const data = await response.json();
      // Ensure we return an object, not just any JSON
      if (typeof data === 'object' && data !== null && !Array.isArray(data)) {
        return data as Record<string, any>;
      }
      return { value: data };
    } catch (err) {
      const error = err instanceof Error ? err : new Error(String(err));
      this.emit('error', error);
      throw error;
    }
  }

  public startPolling() {
    if (this.timer) return;
    
    // Initial fetch
    this.check();
    
    this.timer = setInterval(() => {
      this.check();
    }, this.pollInterval);
  }

  public stopPolling(): void {
    if (this.timer) {
      clearInterval(this.timer);
      this.timer = null;
    }
  }

  private async check() {
    try {
      const config = await this.fetch();
      const configStr = JSON.stringify(config);
      
      if (configStr !== this.currentConfig) {
        this.currentConfig = configStr;
        this.emit('change', config);
      }
    } catch (error) {
      this.emit('error', error);
    }
  }
}

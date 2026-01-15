import { SlackConfig, SlackMessage, SlackResponse } from './types';
import { PluginLogger } from '../types';
import { withBackoff } from '../../../utils/backoff';
import { FETCH_NO_REDIRECT } from '../../../utils/fetch-safe.js';

export class SlackService {
  private config: SlackConfig;
  private readonly logger?: PluginLogger;

  constructor(config: SlackConfig, logger?: PluginLogger) {
    this.config = config;
    this.logger = logger;
  }

  public async sendMessage(message: SlackMessage): Promise<SlackResponse> {
    if (!this.config.token) {
      throw new Error('Slack token is missing');
    }

    return withBackoff(async () => {
      const response = await fetch('https://slack.com/api/chat.postMessage', {
        ...FETCH_NO_REDIRECT,
        method: 'POST',
        headers: {
          Authorization: `Bearer ${this.config.token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(message),
      });

      if (!response.ok) {
        throw new Error(`Slack API HTTP error: ${response.status} ${response.statusText}`);
      }

      const data = await response.json() as SlackResponse;
      if (!data.ok) {
        const errorCode = data.error || 'unknown_error';
        throw new Error(`Slack API error: ${errorCode}`);
      }
      
      return data;
    }, { maxRetries: 3, initialDelay: 500 });
  }

  public async checkAuth(): Promise<boolean> {
    if (!this.config.token) return false;

    try {
      const response = await fetch('https://slack.com/api/auth.test', {
        ...FETCH_NO_REDIRECT,
        method: 'POST',
        headers: {
          Authorization: `Bearer ${this.config.token}`,
          'Content-Type': 'application/json',
        },
      });
      const data = await response.json() as SlackResponse;
      return data.ok;
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : String(e);
      this.logger?.error(`Slack auth check failed: ${msg}`);
      return false;
    }
  }
}

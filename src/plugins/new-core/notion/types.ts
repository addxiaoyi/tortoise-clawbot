import { PluginConfig } from '../types';

export interface NotionConfig extends PluginConfig {
  apiKey: string;
  version?: string; // Default: 2025-09-03
}

export interface NotionPage {
  id: string;
  url: string;
  properties: Record<string, any>;
}

export interface NotionSearchResponse {
  results: NotionPage[];
  next_cursor: string | null;
  has_more: boolean;
}

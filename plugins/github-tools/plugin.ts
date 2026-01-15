/**
 * GitHub Tools Plugin
 * 
 * GitHub API integration for Tortoise
 */

import { ToolPlugin, ToolHandler } from '@tortoise/sdk';

interface GitHubConfig {
  token?: string;
  defaultOwner?: string;
}

interface Issue {
  id: number;
  number: number;
  title: string;
  state: string;
  body?: string;
  labels: string[];
  assignee?: string;
  created_at: string;
  updated_at: string;
}

interface SearchResult {
  total_count: number;
  items: {
    name: string;
    path: string;
    sha: string;
    url: string;
    repository: {
      full_name: string;
    };
  }[];
}

export class GitHubTools implements ToolPlugin {
  readonly manifest = {
    id: 'github-tools',
    name: 'GitHub Tools',
    version: '1.0.0',
    type: 'tool' as const,
  };
  
  private config?: GitHubConfig;
  private baseUrl = 'https://api.github.com';
  
  async onLoad(): Promise<void> {
    console.log('GitHub plugin loaded');
  }
  
  async onEnable(): Promise<void> {
    console.log('GitHub plugin enabled');
  }
  
  async onDisable(): Promise<void> {
    console.log('GitHub plugin disabled');
  }
  
  async configure(config: GitHubConfig): Promise<void> {
    this.config = config;
  }
  
  getTools(): ToolHandler[] {
    return [
      {
        name: 'github_list_issues',
        description: 'List issues in a repository',
        handler: this.listIssues.bind(this),
      },
      {
        name: 'github_create_issue',
        description: 'Create a new issue',
        handler: this.createIssue.bind(this),
      },
      {
        name: 'github_search_code',
        description: 'Search code across repositories',
        handler: this.searchCode.bind(this),
      },
    ];
  }
  
  private async listIssues(args: {
    owner: string;
    repo: string;
    state?: 'open' | 'closed' | 'all';
    limit?: number;
  }): Promise<string> {
    const { owner, repo, state = 'open', limit = 10 } = args;
    
    // In production, use fetch or axios
    // const response = await fetch(
    //   `${this.baseUrl}/repos/${owner}/${repo}/issues?state=${state}&per_page=${limit}`,
    //   { headers: this.getHeaders() }
    // );
    // const issues: Issue[] = await response.json();
    
    // Demo response
    const issues: Issue[] = [
      {
        id: 1,
        number: 42,
        title: 'Bug: Login fails with special characters',
        state: 'open',
        labels: ['bug', 'high-priority'],
        created_at: '2024-01-15T10:00:00Z',
        updated_at: '2024-01-16T14:30:00Z',
      },
      {
        id: 2,
        number: 41,
        title: 'Feature: Add dark mode support',
        state: 'open',
        labels: ['enhancement'],
        created_at: '2024-01-10T08:00:00Z',
        updated_at: '2024-01-12T16:00:00Z',
      },
    ];
    
    return JSON.stringify(issues.slice(0, limit), null, 2);
  }
  
  private async createIssue(args: {
    owner: string;
    repo: string;
    title: string;
    body?: string;
    labels?: string[];
  }): Promise<string> {
    const { owner, repo, title, body = '', labels = [] } = args;
    
    // In production, use fetch or axios
    // const response = await fetch(
    //   `${this.baseUrl}/repos/${owner}/${repo}/issues`,
    //   {
    //     method: 'POST',
    //     headers: this.getHeaders(),
    //     body: JSON.stringify({ title, body, labels }),
    //   }
    // );
    // const issue: Issue = await response.json();
    
    // Demo response
    const issue: Issue = {
      id: Date.now(),
      number: 99,
      title,
      state: 'open',
      body,
      labels,
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
    };
    
    return JSON.stringify(issue, null, 2);
  }
  
  private async searchCode(args: {
    q: string;
    per_page?: number;
  }): Promise<string> {
    const { q, per_page = 10 } = args;
    
    // In production, use fetch or axios
    // const response = await fetch(
    //   `${this.baseUrl}/search/code?q=${encodeURIComponent(q)}&per_page=${per_page}`,
    //   { headers: this.getHeaders() }
    // );
    // const results: SearchResult = await response.json();
    
    // Demo response
    const results: SearchResult = {
      total_count: 2,
      items: [
        {
          name: 'main.ts',
          path: 'src/main.ts',
          sha: 'abc123',
          url: 'https://github.com/example/repo/blob/main/src/main.ts',
          repository: { full_name: 'example/repo' },
        },
      ],
    };
    
    return JSON.stringify(results, null, 2);
  }
  
  private getHeaders(): Record<string, string> {
    const headers: Record<string, string> = {
      'Accept': 'application/vnd.github+json',
      'X-GitHub-Api-Version': '2022-11-28',
    };
    
    if (this.config?.token) {
      headers['Authorization'] = `Bearer ${this.config.token}`;
    }
    
    return headers;
  }
}

export default new GitHubTools();

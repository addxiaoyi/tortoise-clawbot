
import { execFile } from 'child_process';
import { LabelMatcher } from '../../../utils/label-matcher.js';

const REPO_PATTERN = /^[a-zA-Z0-9_.-]+\/[a-zA-Z0-9_.-]+$/;

function isValidRepo(repo: string): boolean {
  return REPO_PATTERN.test(repo);
}

export class GitHubService {
  private labelMatcher: LabelMatcher;

  constructor() {
    this.labelMatcher = new LabelMatcher();
  }

  public async checkAuth(): Promise<boolean> {
    return new Promise<boolean>((resolve) => {
      execFile('gh', ['auth', 'status'], (err: Error | null) => {
        resolve(!err);
      });
    });
  }

  public async listIssues(repo: string): Promise<any[]> {
    if (!isValidRepo(repo)) {
      throw new Error(`Invalid repository format: ${repo}`);
    }
    return new Promise<any[]>((resolve, reject) => {
      execFile('gh', ['issue', 'list', '--repo', repo, '--json', 'number,title,state,body,url'], (error: Error | null, stdout: string, _stderr: string) => {
        if (error) {
          reject(new Error(`Failed to list issues: ${error.message}`));
          return;
        }
        try {
          resolve(JSON.parse(stdout));
        } catch (e) {
          const message = e instanceof Error ? e.message : String(e);
          reject(new Error(`Failed to parse JSON: ${message}`));
        }
      });
    });
  }

  public predictLabels(title: string, body: string): string[] {
    const titleLabels = this.labelMatcher.matchLabels(title);
    const bodyLabels = this.labelMatcher.matchLabels(body);
    return Array.from(new Set([...titleLabels, ...bodyLabels]));
  }
}


import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { GitHubService } from './service';
import { execFile } from 'child_process';

vi.mock('child_process', () => ({
  execFile: vi.fn(),
}));

vi.mock('../../../utils/label-matcher.js', () => ({
  LabelMatcher: class {
    matchLabels(text: string) {
      if (text.includes('bug')) return ['bug'];
      if (text.includes('feature')) return ['enhancement'];
      return [];
    }
  }
}));

describe('GitHubService', () => {
  let service: GitHubService;

  beforeEach(() => {
    service = new GitHubService();
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it('should check auth status', async () => {
    vi.mocked(execFile).mockImplementation((cmd, args, cb: any) => {
      cb(null, 'Logged in to github.com', '');
      return {} as any;
    });

    const status = await service.checkAuth();
    expect(status).toBe(true);
    expect(execFile).toHaveBeenCalledWith('gh', ['auth', 'status'], expect.any(Function));
  });

  it('should return false if auth check fails', async () => {
    vi.mocked(execFile).mockImplementation((cmd, args, cb: any) => {
      cb(new Error('Not logged in'), '', 'error');
      return {} as any;
    });

    const status = await service.checkAuth();
    expect(status).toBe(false);
  });

  it('should list issues', async () => {
    const mockOutput = JSON.stringify([{ number: 1, title: 'Test Issue' }]);
    vi.mocked(execFile).mockImplementation((cmd, args, cb: any) => {
      cb(null, mockOutput, '');
      return {} as any;
    });

    const issues = await service.listIssues('owner/repo');
    expect(issues).toHaveLength(1);
    expect(issues[0].title).toBe('Test Issue');
    expect(execFile).toHaveBeenCalledWith('gh', ['issue', 'list', '--repo', 'owner/repo', '--json', 'number,title,state,body,url'], expect.any(Function));
  });

  it('should predict labels', () => {
    const labels = service.predictLabels('Fix a bug', 'This is a feature');
    expect(labels).toContain('bug');
    expect(labels).toContain('enhancement');
  });
});

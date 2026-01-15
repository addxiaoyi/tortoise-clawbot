import { describe, it, expect, vi } from 'vitest';
import { scoreRepos, GitHubRepo } from './repo-scorer.js';

describe('Repo Scorer', () => {
  it('should calculate scores correctly', () => {
    const mockRepo: GitHubRepo = {
      id: 1,
      name: 'test-repo',
      full_name: 'user/test-repo',
      html_url: 'https://github.com/user/test-repo',
      description: 'A test repo',
      stargazers_count: 5000, // 10 points
      forks_count: 500, // 5 points
      open_issues_count: 10,
      pushed_at: new Date().toISOString(), // 30 points (today)
      created_at: new Date(Date.now() - 365 * 24 * 60 * 60 * 1000).toISOString(), // 1 year ago -> 5 points
      size: 1000,
      language: 'TypeScript', // 15 points
      _category: 'skill',
      _scanned_at: new Date().toISOString()
    };

    const scored = scoreRepos([mockRepo]);
    
    expect(scored).toHaveLength(1);
    const result = scored[0];
    
    // Expected score:
    // Recency: 30 (today)
    // Community: (5000/10000)*20 + (500/2000)*20 = 10 + 5 = 15
    // Growth: (1 year / 1 year) * 5 = 5 (capped at 15)
    // Language: 15
    // Total: 30 + 15 + 5 + 15 = 65
    
    expect(result._score).toBeCloseTo(65, 0);
    expect(result._score_details.activity).toBe(30);
    expect(result._score_details.community).toBe(15);
    expect(result._score_details.growth).toBe(20); // 5 + 15 (lang bonus is added to growth in current implementation)
  });

  it('should handle old repos correctly', () => {
    const mockRepo: GitHubRepo = {
      id: 2,
      name: 'old-repo',
      full_name: 'user/old-repo',
      html_url: '',
      description: '',
      stargazers_count: 0,
      forks_count: 0,
      open_issues_count: 0,
      pushed_at: new Date(Date.now() - 100 * 24 * 60 * 60 * 1000).toISOString(), // > 90 days -> 0 points
      created_at: new Date().toISOString(),
      size: 0,
      language: 'C++', // 0 points
      _category: 'skill',
      _scanned_at: ''
    };

    const scored = scoreRepos([mockRepo]);
    expect(scored[0]._score).toBe(0);
  });
});

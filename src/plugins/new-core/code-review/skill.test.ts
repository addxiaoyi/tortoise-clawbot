import { describe, it, expect, vi, beforeEach } from 'vitest';
import { CodeReviewPlugin } from './plugin.js';
import { SkillTestHarness } from '../../../utils/skill-test-harness.js';

// Mock CodeReviewService
const mockServiceInstance = {
  start: vi.fn(),
  stop: vi.fn(),
  reviewCode: vi.fn(),
  analyzeDiff: vi.fn()
};

vi.mock('./service.js', () => {
  const CodeReviewService = vi.fn().mockImplementation(function () {
    return mockServiceInstance;
  });
  return { CodeReviewService };
});

describe('CodeReviewPlugin Skill', () => {
  let plugin: CodeReviewPlugin;
  let harness: SkillTestHarness;

  beforeEach(async () => {
    vi.clearAllMocks();
    
    plugin = new CodeReviewPlugin();
    harness = new SkillTestHarness(plugin);
    
    // Mock config
    harness.context.getConfig = vi.fn().mockReturnValue({ maxLines: 100 });
    
    // Initialize
    await harness.init();
    
    // Setup service mocks
    mockServiceInstance.reviewCode.mockResolvedValue('Code looks good!');
    mockServiceInstance.analyzeDiff.mockResolvedValue('Diff analysis: No major issues found.');
  });

  it('should expose skill definition', () => {
    const definition = plugin.getSkillDefinition();
    expect(definition.name).toBe('code-review');
    expect(definition.tools).toHaveLength(2);
  });

  it('should execute review_code tool', async () => {
    const result = await harness.executeTool('review_code', { code: 'console.log("hello")', language: 'javascript' });
    expect(mockServiceInstance.reviewCode).toHaveBeenCalledWith('console.log("hello")', 'javascript');
    expect(result).toEqual({ feedback: 'Code looks good!' });
  });

  it('should execute analyze_diff tool', async () => {
    const result = await harness.executeTool('analyze_diff', { diff: 'diff --git a/file.ts b/file.ts' });
    expect(mockServiceInstance.analyzeDiff).toHaveBeenCalledWith('diff --git a/file.ts b/file.ts');
    expect(result).toEqual({ analysis: 'Diff analysis: No major issues found.' });
  });
});

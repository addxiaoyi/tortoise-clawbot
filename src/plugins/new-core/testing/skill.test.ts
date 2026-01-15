import { describe, it, expect, vi, beforeEach } from 'vitest';
import { TestingPlugin } from './plugin.js';
import { SkillTestHarness } from '../../../utils/skill-test-harness.js';

// Mock TestingService
const mockServiceInstance = {
  start: vi.fn(),
  stop: vi.fn(),
  runTests: vi.fn(),
  createTestFile: vi.fn()
};

vi.mock('./service.js', () => {
  const TestingService = vi.fn().mockImplementation(function () {
    return mockServiceInstance;
  });
  return { TestingService };
});

describe('TestingPlugin Skill', () => {
  let plugin: TestingPlugin;
  let harness: SkillTestHarness;

  beforeEach(async () => {
    vi.clearAllMocks();
    
    plugin = new TestingPlugin();
    harness = new SkillTestHarness(plugin);
    
    // Mock config
    harness.context.getConfig = vi.fn().mockReturnValue({ testDir: '/tmp/testing' });
    
    // Initialize
    await harness.init();
    
    // Setup service mocks
    mockServiceInstance.runTests.mockResolvedValue('Ran tests matching ".*". Result: PASS');
    mockServiceInstance.createTestFile.mockResolvedValue('/tmp/testing/example.test.ts');
  });

  it('should expose skill definition', () => {
    const definition = plugin.getSkillDefinition();
    expect(definition.name).toBe('testing');
    expect(definition.tools).toHaveLength(2);
  });

  it('should execute run_tests tool', async () => {
    const result = await harness.executeTool('run_tests', { pattern: '.*' });
    expect(mockServiceInstance.runTests).toHaveBeenCalledWith('.*');
    expect(result).toEqual({ output: 'Ran tests matching ".*". Result: PASS' });
  });

  it('should execute create_test tool', async () => {
    const result = await harness.executeTool('create_test', { filename: 'example.test.ts', content: 'test code' });
    expect(mockServiceInstance.createTestFile).toHaveBeenCalledWith('example.test.ts', 'test code');
    expect(result).toEqual({ filePath: '/tmp/testing/example.test.ts' });
  });
});

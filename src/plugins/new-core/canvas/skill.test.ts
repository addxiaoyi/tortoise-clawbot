import { describe, it, expect, vi, beforeEach } from 'vitest';
import { CanvasPlugin } from './plugin.js';
import { SkillTestHarness } from '../../../utils/skill-test-harness.js';

// Mock CanvasService
const mockServiceInstance = {
  start: vi.fn(),
  stop: vi.fn(),
  writeFile: vi.fn(),
  getUrl: vi.fn()
};

vi.mock('./service.js', () => {
  const CanvasService = vi.fn().mockImplementation(function () {
    return mockServiceInstance;
  });
  return { CanvasService };
});

describe('CanvasPlugin Skill', () => {
  let plugin: CanvasPlugin;
  let harness: SkillTestHarness;

  beforeEach(async () => {
    vi.clearAllMocks();
    
    plugin = new CanvasPlugin();
    harness = new SkillTestHarness(plugin);
    
    // Mock config
    harness.context.getConfig = vi.fn().mockReturnValue({ port: 3000, root: '/tmp/canvas' });
    
    // Initialize
    await harness.init();
    
    // Setup service mocks
    mockServiceInstance.writeFile.mockResolvedValue('http://localhost:3000/test.html');
    mockServiceInstance.getUrl.mockReturnValue('http://localhost:3000/test.html');
  });

  it('should expose skill definition', () => {
    const definition = plugin.getSkillDefinition();
    expect(definition.name).toBe('canvas');
    expect(definition.tools).toHaveLength(2);
  });

  it('should execute write_canvas_file tool', async () => {
    const result = await harness.executeTool('write_canvas_file', { filename: 'test.html', content: '<h1>Hello</h1>' });
    expect(mockServiceInstance.writeFile).toHaveBeenCalledWith('test.html', '<h1>Hello</h1>');
    expect(result).toEqual({ url: 'http://localhost:3000/test.html', filename: 'test.html' });
  });

  it('should execute get_canvas_url tool', async () => {
    const result = await harness.executeTool('get_canvas_url', { filename: 'test.html' });
    expect(mockServiceInstance.getUrl).toHaveBeenCalledWith('test.html');
    expect(result).toEqual({ url: 'http://localhost:3000/test.html' });
  });
});

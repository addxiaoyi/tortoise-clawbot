import { describe, it, expect, vi, beforeEach } from 'vitest';
import { WebBuilderPlugin } from './plugin.js';
import { SkillTestHarness } from '../../../utils/skill-test-harness.js';

// Mock WebBuilderService
const mockServiceInstance = {
  start: vi.fn(),
  stop: vi.fn(),
  buildReactComponent: vi.fn(),
  generateHtml: vi.fn()
};

vi.mock('./service.js', () => {
  return {
    WebBuilderService: class {
      constructor() {
        return mockServiceInstance;
      }
    }
  };
});

describe('WebBuilderPlugin Skill', () => {
  let plugin: WebBuilderPlugin;
  let harness: SkillTestHarness;

  beforeEach(async () => {
    vi.clearAllMocks();
    
    plugin = new WebBuilderPlugin();
    harness = new SkillTestHarness(plugin);
    
    // Mock config
    harness.context.getConfig = vi.fn().mockReturnValue({ outputDir: '/tmp/web-builder' });
    
    // Initialize
    await harness.init();
    
    // Setup service mocks
    mockServiceInstance.buildReactComponent.mockResolvedValue('/tmp/web-builder/App.tsx');
    mockServiceInstance.generateHtml.mockResolvedValue('/tmp/web-builder/index.html');
  });

  it('should expose skill definition', () => {
    const definition = plugin.getSkillDefinition();
    expect(definition.name).toBe('web-builder');
    expect(definition.tools).toHaveLength(2);
  });

  it('should execute build_react_component tool', async () => {
    const result = await harness.executeTool('build_react_component', { filename: 'App.tsx', code: 'export default () => <div>Hello</div>' });
    expect(mockServiceInstance.buildReactComponent).toHaveBeenCalledWith('App.tsx', 'export default () => <div>Hello</div>');
    expect(result).toEqual({ filePath: '/tmp/web-builder/App.tsx' });
  });

  it('should execute generate_html tool', async () => {
    const result = await harness.executeTool('generate_html', { filename: 'index.html', title: 'Test Page', body: '<h1>Hello</h1>' });
    expect(mockServiceInstance.generateHtml).toHaveBeenCalledWith('index.html', 'Test Page', '<h1>Hello</h1>');
    expect(result).toEqual({ filePath: '/tmp/web-builder/index.html' });
  });
});

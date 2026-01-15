import { describe, it, expect, vi, beforeEach } from 'vitest';
import { DocumentationPlugin } from './plugin.js';
import { SkillTestHarness } from '../../../utils/skill-test-harness.js';

vi.mock('./service.js', () => {
  class DocumentationService {
    start = vi.fn();
    stop = vi.fn();
    generateDoc = vi.fn();
    listDocs = vi.fn();
    getDocContent = vi.fn();
  }
  return { DocumentationService };
});

describe('DocumentationPlugin Skill', () => {
  let plugin: DocumentationPlugin;
  let harness: SkillTestHarness;
  let mockServiceInstance: any;

  beforeEach(async () => {
    vi.clearAllMocks();
    
    plugin = new DocumentationPlugin();
    harness = new SkillTestHarness(plugin);
    
    // Mock config
    harness.context.getConfig = vi.fn().mockReturnValue({ format: 'markdown' });
    
    await harness.init();
    mockServiceInstance = (plugin as any).service;
    mockServiceInstance.generateDoc.mockResolvedValue('test-doc.markdown');
    mockServiceInstance.listDocs.mockResolvedValue(['test-doc.markdown']);
    mockServiceInstance.getDocContent.mockResolvedValue('# Test Doc\nContent');
  });

  it('should expose skill definition', () => {
    const definition = plugin.getSkillDefinition();
    expect(definition.name).toBe('documentation');
    expect(definition.tools).toHaveLength(3);
  });

  it('should execute generate_doc tool', async () => {
    const result = await harness.executeTool('generate_doc', { title: 'Test Doc', content: '# Test Doc\nContent' });
    expect(mockServiceInstance.generateDoc).toHaveBeenCalledWith('Test Doc', '# Test Doc\nContent');
    expect(result).toEqual({ filename: 'test-doc.markdown' });
  });

  it('should execute list_docs tool', async () => {
    const result = await harness.executeTool('list_docs', {});
    expect(mockServiceInstance.listDocs).toHaveBeenCalled();
    expect(result).toEqual({ docs: ['test-doc.markdown'] });
  });

  it('should execute get_doc_content tool', async () => {
    const result = await harness.executeTool('get_doc_content', { filename: 'test-doc.markdown' });
    expect(mockServiceInstance.getDocContent).toHaveBeenCalledWith('test-doc.markdown');
    expect(result).toEqual({ content: '# Test Doc\nContent' });
  });
});

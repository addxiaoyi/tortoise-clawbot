import { describe, it, expect, vi, beforeEach } from 'vitest';
import { OfficePlugin } from './plugin.js';
import { SkillTestHarness } from '../../../utils/skill-test-harness.js';

// Mock OfficeService
const mockServiceInstance = {
  start: vi.fn(),
  stop: vi.fn(),
  createExcel: vi.fn(),
  readExcel: vi.fn(),
  createWordDoc: vi.fn()
};

vi.mock('./service.js', () => {
  const OfficeService = vi.fn().mockImplementation(function () {
    return mockServiceInstance;
  });
  return { OfficeService };
});

describe('OfficePlugin Skill', () => {
  let plugin: OfficePlugin;
  let harness: SkillTestHarness;

  beforeEach(async () => {
    vi.clearAllMocks();
    
    plugin = new OfficePlugin();
    harness = new SkillTestHarness(plugin);
    
    // Mock config
    harness.context.getConfig = vi.fn().mockReturnValue({ outputDir: '/tmp/office' });
    
    // Initialize
    await harness.init();
    
    // Setup service mocks
    mockServiceInstance.createExcel.mockResolvedValue('/tmp/office/test.xlsx');
    mockServiceInstance.readExcel.mockResolvedValue([['a', 'b'], ['1', '2']]);
    mockServiceInstance.createWordDoc.mockResolvedValue('/tmp/office/test.docx');
  });

  it('should expose skill definition', () => {
    const definition = plugin.getSkillDefinition();
    expect(definition.name).toBe('office');
    expect(definition.tools).toHaveLength(3);
  });

  it('should execute create_excel tool', async () => {
    const result = await harness.executeTool('create_excel', { filename: 'test.xlsx', data: [['a', 'b'], ['1', '2']] });
    expect(mockServiceInstance.createExcel).toHaveBeenCalledWith('test.xlsx', [['a', 'b'], ['1', '2']]);
    expect(result).toEqual({ filePath: '/tmp/office/test.xlsx' });
  });

  it('should execute read_excel tool', async () => {
    const result = await harness.executeTool('read_excel', { filename: 'test.xlsx' });
    expect(mockServiceInstance.readExcel).toHaveBeenCalledWith('test.xlsx');
    expect(result).toEqual([['a', 'b'], ['1', '2']]);
  });
  
  it('should execute create_word tool', async () => {
    const result = await harness.executeTool('create_word', { filename: 'test.docx', text: 'Hello World' });
    expect(mockServiceInstance.createWordDoc).toHaveBeenCalledWith('test.docx', 'Hello World');
    expect(result).toEqual({ filePath: '/tmp/office/test.docx' });
  });
});

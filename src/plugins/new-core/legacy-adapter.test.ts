import { describe, it, expect, vi, beforeEach } from 'vitest';
import { LegacySkillAdapter } from './legacy-adapter.js';
import { PluginContext } from './types.js';
import { readdir, readFile, stat } from 'node:fs/promises';

// Mock fs
vi.mock('node:fs/promises', () => ({
  readdir: vi.fn(),
  readFile: vi.fn(),
  stat: vi.fn(),
}));

describe('LegacySkillAdapter', () => {
  let adapter: LegacySkillAdapter;
  let mockContext: PluginContext;

  beforeEach(() => {
    adapter = new LegacySkillAdapter('/mock/skills');
    mockContext = {
      logger: {
        info: vi.fn(),
        warn: vi.fn(),
        error: vi.fn(),
        debug: vi.fn(),
      },
      events: {
        emit: vi.fn(),
        on: vi.fn(),
        off: vi.fn(),
      },
      meta: { id: 'test', name: 'test', version: '1.0' },
      storage: { getItem: vi.fn(), setItem: vi.fn(), removeItem: vi.fn(), clear: vi.fn() },
      getConfig: vi.fn(),
    } as unknown as PluginContext;
    
    vi.clearAllMocks();
  });

  it('should load skills on init', async () => {
    // Mock fs behavior
    vi.mocked(stat).mockResolvedValue({ isFile: () => true } as any);
    vi.mocked(readdir).mockResolvedValue([
      { name: 'skill1', isDirectory: () => true } as any,
    ]);
    vi.mocked(readFile).mockResolvedValue('# Skill 1\n\nThis is a description.');

    await adapter.onInit(mockContext);

    expect(adapter.getSkills()).toHaveLength(1);
    expect(adapter.getSkills()[0].name).toBe('skill1');
    expect(adapter.getSkills()[0].description).toBe('This is a description.');
  });

  it('should handle missing directory', async () => {
    vi.mocked(stat).mockRejectedValue(new Error('ENOENT'));
    
    await adapter.onInit(mockContext);
    
    expect(mockContext.logger.warn).toHaveBeenCalled();
    expect(adapter.getSkills()).toHaveLength(0);
  });

  it('should emit loaded event on start', async () => {
    // Need to init first to set context
    vi.mocked(stat).mockResolvedValue({ isFile: () => true } as any);
    vi.mocked(readdir).mockResolvedValue([]);
    
    await adapter.onInit(mockContext);
    await adapter.onStart();
    expect(mockContext.events.emit).toHaveBeenCalledWith('skills:loaded', expect.any(Array));
  });
});

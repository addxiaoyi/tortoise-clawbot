import { vi } from 'vitest';
import { PluginContext, SkillPlugin, SkillTool } from '../plugins/new-core/types';

export class SkillTestHarness {
  public context: PluginContext;
  public plugin: SkillPlugin;

  constructor(plugin: SkillPlugin) {
    this.plugin = plugin;
    this.context = {
      meta: {
        id: 'test-skill',
        name: 'Test Skill',
        version: '1.0.0',
      },
      logger: {
        debug: vi.fn(),
        info: vi.fn(),
        warn: vi.fn(),
        error: vi.fn(),
      },
      storage: {
        getItem: vi.fn(),
        setItem: vi.fn(),
        removeItem: vi.fn(),
        clear: vi.fn(),
      },
      events: {
        emit: vi.fn(),
        on: vi.fn(),
        off: vi.fn(),
      },
      getConfig: vi.fn().mockReturnValue({}),
    };
  }

  async init() {
    await this.plugin.onInit(this.context);
    if (this.plugin.onStart) {
      await this.plugin.onStart();
    }
  }

  async executeTool(toolName: string, args: any) {
    const definition = this.plugin.getSkillDefinition();
    const tool = definition.tools.find((t: SkillTool) => t.name === toolName);
    if (!tool) {
      throw new Error(`Tool ${toolName} not found in skill ${definition.name}`);
    }
    return tool.execute(args, this.context);
  }
}

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { PlanningPlugin } from './plugin.js';
import { SkillTestHarness } from '../../../utils/skill-test-harness.js';

vi.mock('./service.js', () => {
  class PlanningService {
    start = vi.fn();
    stop = vi.fn();
    createTask = vi.fn();
    updateTaskStatus = vi.fn();
    getTasks = vi.fn();
  }
  return { PlanningService };
});

describe('PlanningPlugin Skill', () => {
  let plugin: PlanningPlugin;
  let harness: SkillTestHarness;
  let mockServiceInstance: any;

  beforeEach(async () => {
    vi.clearAllMocks();
    
    plugin = new PlanningPlugin();
    harness = new SkillTestHarness(plugin);
    
    // Mock config
    harness.context.getConfig = vi.fn().mockReturnValue({ maxTasks: 50 });
    
    await harness.init();
    mockServiceInstance = (plugin as any).service;
    mockServiceInstance.createTask.mockResolvedValue({ id: '123', title: 'Test Task', status: 'pending' });
    mockServiceInstance.updateTaskStatus.mockResolvedValue({ id: '123', title: 'Test Task', status: 'completed' });
    mockServiceInstance.getTasks.mockResolvedValue([{ id: '123', title: 'Test Task', status: 'completed' }]);
  });

  it('should expose skill definition', () => {
    const definition = plugin.getSkillDefinition();
    expect(definition.name).toBe('planning');
    expect(definition.tools).toHaveLength(3);
  });

  it('should execute create_task tool', async () => {
    const result = await harness.executeTool('create_task', { title: 'Test Task' });
    expect(mockServiceInstance.createTask).toHaveBeenCalledWith('Test Task');
    expect(result).toEqual({ task: { id: '123', title: 'Test Task', status: 'pending' } });
  });

  it('should execute update_task_status tool', async () => {
    const result = await harness.executeTool('update_task_status', { id: '123', status: 'completed' });
    expect(mockServiceInstance.updateTaskStatus).toHaveBeenCalledWith('123', 'completed');
    expect(result).toEqual({ task: { id: '123', title: 'Test Task', status: 'completed' } });
  });

  it('should execute list_tasks tool', async () => {
    const result = await harness.executeTool('list_tasks', {});
    expect(mockServiceInstance.getTasks).toHaveBeenCalled();
    expect(result).toEqual({ tasks: [{ id: '123', title: 'Test Task', status: 'completed' }] });
  });
});

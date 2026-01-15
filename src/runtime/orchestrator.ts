/**
 * Multi-Agent Orchestrator
 * Coordinates multiple AI agents for complex tasks
 */

import type { PluginContext } from '../plugins/new-core/types.js';

// ============================================
// Agent Definition
// ============================================

export interface Agent {
  id: string;
  name: string;
  description: string;
  role: AgentRole;
  skills: string[];           // Skill IDs this agent can use
  capabilities: string[];     // Capability IDs
  maxConcurrentTasks: number;
  priority: number;           // Higher = more priority
  model?: string;            // Override default model
}

export type AgentRole = 
  | 'coordinator'    // Orchestrates other agents
  | 'researcher'    // Gathers information
  | 'coder'          // Writes code
  | 'reviewer'       // Reviews and critiques
  | 'executor'       // Executes tasks
  | 'specialist';    // Domain expert

// ============================================
// Task Definition
// ============================================

export interface Task {
  id: string;
  type: string;
  description: string;
  input: unknown;
  priority?: number;
  deadline?: number;
  dependencies?: string[];    // Task IDs that must complete first
  requiredCapabilities?: string[];
  preferredAgentId?: string;
  metadata?: Record<string, unknown>;
}

export interface TaskResult {
  taskId: string;
  success: boolean;
  output?: unknown;
  error?: string;
  durationMs: number;
  agentId?: string;
}

// ============================================
// Orchestration Strategy
// ============================================

export type OrchestrationStrategy = 
  | 'sequential'    // Tasks in order
  | 'parallel'      // Tasks concurrently
  | 'pipeline'       // Output of one feeds into next
  | 'hierarchical';  // Coordinator delegates

// ============================================
// Orchestrator
// ============================================

export interface OrchestratorConfig {
  maxConcurrentTasks: number;
  taskTimeoutMs: number;
  enablePriority: boolean;
  enableFallback: boolean;     // Fallback to other agents on failure
  maxRetries: number;
  retryDelayMs: number;
}

export class AgentOrchestrator {
  private agents: Map<string, Agent> = new Map();
  private tasks: Map<string, Task> = new Map();
  private results: Map<string, TaskResult> = new Map();
  private runningTasks: Set<string> = new Set();
  private taskQueue: string[] = [];
  
  private config: OrchestratorConfig;
  private ctx?: PluginContext;

  constructor(config?: Partial<OrchestratorConfig>) {
    this.config = {
      maxConcurrentTasks: 5,
      taskTimeoutMs: 60000,
      enablePriority: true,
      enableFallback: true,
      maxRetries: 2,
      retryDelayMs: 1000,
      ...config,
    };
  }

  setContext(ctx: PluginContext): void {
    this.ctx = ctx;
  }

  // ============================================
  // Agent Management
  // ============================================

  /**
   * Register an agent
   */
  registerAgent(agent: Agent): void {
    this.agents.set(agent.id, agent);
    this.ctx?.logger.info(`[orchestrator] Registered agent: ${agent.name} (${agent.id})`);
  }

  /**
   * Unregister an agent
   */
  unregisterAgent(agentId: string): void {
    this.agents.delete(agentId);
    this.ctx?.logger.info(`[orchestrator] Unregistered agent: ${agentId}`);
  }

  /**
   * Get all registered agents
   */
  getAgents(): Agent[] {
    return Array.from(this.agents.values());
  }

  /**
   * Get agents by role
   */
  getAgentsByRole(role: AgentRole): Agent[] {
    return this.getAgents().filter(a => a.role === role);
  }

  /**
   * Find best agent for a task
   */
  findBestAgent(task: Task): Agent | null {
    const candidates = this.getAgents().filter(agent => {
      // Check capabilities
      if (task.requiredCapabilities?.length) {
        const hasCapabilities = task.requiredCapabilities.every(
          cap => agent.capabilities.includes(cap)
        );
        if (!hasCapabilities) return false;
      }

      // Check availability
      const activeTasks = this.getAgentActiveTaskCount(agent.id);
      if (activeTasks >= agent.maxConcurrentTasks) return false;

      return true;
    });

    if (candidates.length === 0) return null;

    // Sort by priority and preference
    return candidates.sort((a, b) => {
      // Prefer requested agent
      if (task.preferredAgentId === a.id) return -1;
      if (task.preferredAgentId === b.id) return 1;

      // Higher priority wins
      if (a.priority !== b.priority) {
        return this.config.enablePriority 
          ? b.priority - a.priority 
          : 0;
      }

      // Less busy agent wins
      const aTasks = this.getAgentActiveTaskCount(a.id);
      const bTasks = this.getAgentActiveTaskCount(b.id);
      return aTasks - bTasks;
    })[0];
  }

  private getAgentActiveTaskCount(agentId: string): number {
    let count = 0;
    for (const taskId of this.runningTasks) {
      const task = this.tasks.get(taskId);
      if (task && this.results.get(taskId)?.agentId === agentId) {
        count++;
      }
    }
    return count;
  }

  // ============================================
  // Task Management
  // ============================================

  /**
   * Submit a task for execution
   */
  async submitTask(task: Task): Promise<string> {
    // Check dependencies
    if (task.dependencies?.length) {
      for (const depId of task.dependencies) {
        const depResult = this.results.get(depId);
        if (!depResult?.success) {
          throw new Error(`Task ${task.id} blocked by failed dependency: ${depId}`);
        }
      }
    }

    this.tasks.set(task.id, task);
    this.taskQueue.push(task.id);
    
    this.ctx?.logger.info(`[orchestrator] Task submitted: ${task.id} (${task.type})`);
    
    // Process immediately if possible
    this.processQueue();
    
    return task.id;
  }

  /**
   * Submit multiple tasks
   */
  async submitTasks(tasks: Task[]): Promise<string[]> {
    return Promise.all(tasks.map(t => this.submitTask(t)));
  }

  /**
   * Get task result
   */
  getResult(taskId: string): TaskResult | null {
    return this.results.get(taskId) || null;
  }

  /**
   * Wait for task completion
   */
  async waitForTask(taskId: string, timeoutMs?: number): Promise<TaskResult> {
    const timeout = timeoutMs || this.config.taskTimeoutMs;
    const start = Date.now();
    
    while (Date.now() - start < timeout) {
      const result = this.results.get(taskId);
      if (result) return result;
      
      await new Promise(resolve => setTimeout(resolve, 100));
    }
    
    throw new Error(`Task ${taskId} timed out after ${timeout}ms`);
  }

  // ============================================
  // Task Processing
  // ============================================

  private processQueue(): void {
    while (this.runningTasks.size < this.config.maxConcurrentTasks) {
      const taskId = this.taskQueue.shift();
      if (!taskId) break;
      
      // Skip if already running
      if (this.runningTasks.has(taskId)) continue;
      
      // Check dependencies again
      const task = this.tasks.get(taskId);
      if (!task) continue;
      
      this.executeTask(task);
    }
  }

  private async executeTask(task: Task): Promise<void> {
    this.runningTasks.add(task.id);
    
    const agent = this.findBestAgent(task);
    
    if (!agent) {
      this.results.set(task.id, {
        taskId: task.id,
        success: false,
        error: 'No available agent found',
        durationMs: 0,
      });
      this.runningTasks.delete(task.id);
      this.processQueue();
      return;
    }

    const startTime = Date.now();
    
    this.ctx?.logger.info(
      `[orchestrator] Executing task ${task.id} with agent ${agent.name}`
    );

    try {
      // Execute with timeout
      const result = await Promise.race([
        this.executeOnAgent(agent, task),
        new Promise<never>((_, reject) => 
          setTimeout(() => reject(new Error('Task timeout')), this.config.taskTimeoutMs)
        ),
      ]);

      this.results.set(task.id, {
        taskId: task.id,
        success: true,
        output: result,
        durationMs: Date.now() - startTime,
        agentId: agent.id,
      });
      
      this.ctx?.logger.info(
        `[orchestrator] Task ${task.id} completed in ${Date.now() - startTime}ms`
      );
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : String(error);
      
      this.ctx?.logger.error(
        `[orchestrator] Task ${task.id} failed: ${errorMessage}`
      );

      // Retry logic
      const retryCount = this.getRetryCount(task.id);
      if (retryCount < this.config.maxRetries && this.config.enableFallback) {
        this.ctx?.logger.info(
          `[orchestrator] Retrying task ${task.id} (attempt ${retryCount + 1})`
        );
        
        await new Promise(resolve => setTimeout(resolve, this.config.retryDelayMs));
        this.taskQueue.unshift(task.id);
        this.incrementRetryCount(task.id);
      } else {
        this.results.set(task.id, {
          taskId: task.id,
          success: false,
          error: errorMessage,
          durationMs: Date.now() - startTime,
          agentId: agent.id,
        });
      }
    }

    this.runningTasks.delete(task.id);
    this.processQueue();
  }

  private retryCounts = new Map<string, number>();
  
  private getRetryCount(taskId: string): number {
    return this.retryCounts.get(taskId) || 0;
  }
  
  private incrementRetryCount(taskId: string): void {
    const current = this.getRetryCount(taskId);
    this.retryCounts.set(taskId, current + 1);
  }

  /**
   * Execute task on agent (implement actual execution logic)
   */
  private async executeOnAgent(agent: Agent, task: Task): Promise<unknown> {
    // This is where you'd integrate with your actual agent runtime
    // For now, we emit an event for external processing
    
    this.ctx?.events.emit('orchestrator:task:execute', {
      agentId: agent.id,
      taskId: task.id,
      taskType: task.type,
      input: task.input,
    });

    // Wait for completion event
    return new Promise((resolve, reject) => {
      const timeout = setTimeout(() => {
        cleanup();
        reject(new Error('Task execution timeout'));
      }, this.config.taskTimeoutMs);

      const handler = (payload: unknown) => {
        const result = payload as { taskId: string; success: boolean; output?: unknown; error?: string };
        if (result.taskId === task.id) {
          cleanup();
          if (result.success) {
            resolve(result.output);
          } else {
            reject(new Error(result.error || 'Task failed'));
          }
        }
      };

      const cleanup = () => {
        clearTimeout(timeout);
        this.ctx?.events.off('orchestrator:task:result', handler);
      };

      this.ctx?.events.on('orchestrator:task:result', handler);
    });
  }

  // ============================================
  // High-Level Workflows
  // ============================================

  /**
   * Execute tasks in parallel
   */
  async executeParallel(tasks: Task[]): Promise<TaskResult[]> {
    await this.submitTasks(tasks);
    return Promise.all(tasks.map(t => this.waitForTask(t.id)));
  }

  /**
   * Execute tasks in pipeline (output feeds into next)
   */
  async executePipeline(tasks: Task[]): Promise<TaskResult[]> {
    const results: TaskResult[] = [];
    
    for (const task of tasks) {
      // Wait for dependencies
      if (task.dependencies?.length) {
        for (const depId of task.dependencies) {
          await this.waitForTask(depId);
        }
        
        // Pass dependency results as input
        const depResults = task.dependencies.map(depId => 
          this.results.get(depId)?.output
        );
        task.input = { ...task.input as object, pipelineInput: depResults };
      }
      
      await this.submitTask(task);
      const result = await this.waitForTask(task.id);
      results.push(result);
      
      if (!result.success) {
        break; // Stop pipeline on failure
      }
    }
    
    return results;
  }

  /**
   * Create a coordinator-based workflow
   */
  async executeHierarchical(
    coordinatorTask: Task,
    subtasks: Task[]
  ): Promise<TaskResult> {
    // Submit subtasks with coordinator task as dependency
    const subtasksWithDeps = subtasks.map(st => ({
      ...st,
      dependencies: [...(st.dependencies || []), coordinatorTask.id],
    }));

    await this.submitTasks(subtasksWithDeps);
    
    // Submit coordinator task
    await this.submitTask(coordinatorTask);
    
    // Wait for coordinator
    const result = await this.waitForTask(coordinatorTask.id);
    
    // Wait for all subtasks
    await Promise.all(subtasks.map(st => this.waitForTask(st.id)));
    
    return result;
  }

  // ============================================
  // Status & Metrics
  // ============================================

  getStatus(): {
    totalAgents: number;
    activeTasks: number;
    queuedTasks: number;
    completedTasks: number;
    failedTasks: number;
    avgDurationMs: number;
  } {
    let completed = 0;
    let failed = 0;
    let totalDuration = 0;
    
    for (const result of this.results.values()) {
      if (result.success) {
        completed++;
        totalDuration += result.durationMs;
      } else {
        failed++;
      }
    }

    return {
      totalAgents: this.agents.size,
      activeTasks: this.runningTasks.size,
      queuedTasks: this.taskQueue.length,
      completedTasks: completed,
      failedTasks: failed,
      avgDurationMs: completed > 0 ? totalDuration / completed : 0,
    };
  }

  /**
   * Get active agents with their current tasks
   */
  getActiveAgents(): Array<{ agent: Agent; tasks: string[] }> {
    const active: Array<{ agent: Agent; tasks: string[] }> = [];
    
    for (const agent of this.agents.values()) {
      const tasks: string[] = [];
      
      for (const taskId of this.runningTasks) {
        const result = this.results.get(taskId);
        if (result?.agentId === agent.id) {
          tasks.push(taskId);
        }
      }
      
      if (tasks.length > 0) {
        active.push({ agent, tasks });
      }
    }
    
    return active;
  }
}

// ============================================
// Predefined Agent Templates
// ============================================

export const AgentTemplates = {
  coordinator: (id: string, name: string): Agent => ({
    id,
    name,
    description: 'Coordinates multiple agents for complex tasks',
    role: 'coordinator',
    skills: ['planning', 'delegation', 'monitoring'],
    capabilities: ['planning', 'delegation', 'monitoring'],
    maxConcurrentTasks: 3,
    priority: 100,
  }),

  researcher: (id: string, name: string): Agent => ({
    id,
    name,
    description: 'Gathers and analyzes information',
    role: 'researcher',
    skills: ['web-search', 'analysis', 'summarization'],
    capabilities: ['search', 'analysis', 'writing'],
    maxConcurrentTasks: 5,
    priority: 50,
  }),

  coder: (id: string, name: string): Agent => ({
    id,
    name,
    description: 'Writes and refactors code',
    role: 'coder',
    skills: ['code-generation', 'code-review', 'debugging'],
    capabilities: ['coding', 'refactoring', 'testing'],
    maxConcurrentTasks: 3,
    priority: 60,
  }),

  reviewer: (id: string, name: string): Agent => ({
    id,
    name,
    description: 'Reviews and critiques work',
    role: 'reviewer',
    skills: ['code-review', 'quality', 'security'],
    capabilities: ['review', 'analysis', 'security'],
    maxConcurrentTasks: 4,
    priority: 40,
  }),

  executor: (id: string, name: string): Agent => ({
    id,
    name,
    description: 'Executes commands and scripts',
    role: 'executor',
    skills: ['bash', 'git', 'deployment'],
    capabilities: ['execution', 'deployment', 'automation'],
    maxConcurrentTasks: 2,
    priority: 30,
  }),
};

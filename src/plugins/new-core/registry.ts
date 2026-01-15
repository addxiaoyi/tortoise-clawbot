import { PluginContext, PluginLifecycle, PluginMetadata } from './types.js';
import { Container } from './container.js';
import { PluginSandbox, SandboxOptions } from './sandbox.js';

export interface PluginDefinition {
  meta: PluginMetadata;
  lifecycle: PluginLifecycle;
  permissions?: SandboxOptions;
}

/**
 * Registry manages plugin lifecycle and dependencies.
 */
export class PluginRegistry {
  private plugins = new Map<string, PluginDefinition>();
  private container: Container;

  constructor(container: Container) {
    this.container = container;
  }

  /**
   * Register a plugin.
   */
  public register(plugin: PluginDefinition): void {
    if (this.plugins.has(plugin.meta.id)) {
      throw new Error(`Plugin already registered: ${plugin.meta.id}`);
    }
    this.plugins.set(plugin.meta.id, plugin);
  }

  /**
   * Initialize all registered plugins.
   */
  public async initAll(baseContext: PluginContext): Promise<void> {
    for (const plugin of this.plugins.values()) {
      const sandbox = PluginSandbox.create(baseContext, plugin.permissions || {});
      // In a real implementation, we would create a scoped context here
      // with a unique logger and storage for each plugin.
      await plugin.lifecycle.onInit(sandbox);
    }
  }

  /**
   * Start all registered plugins.
   */
  public async startAll(): Promise<void> {
    for (const plugin of this.plugins.values()) {
      if (plugin.lifecycle.onStart) {
        await plugin.lifecycle.onStart();
      }
    }
  }

  /**
   * Stop all registered plugins.
   */
  public async stopAll(): Promise<void> {
    for (const plugin of this.plugins.values()) {
      if (plugin.lifecycle.onStop) {
        await plugin.lifecycle.onStop();
      }
    }
  }

  /**
   * Get a registered plugin definition.
   */
  public get(id: string): PluginDefinition | undefined {
    return this.plugins.get(id);
  }
}

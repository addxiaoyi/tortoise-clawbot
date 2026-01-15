/**
 * Core types for the new Micro-Core Plugin (MCP) architecture.
 */

export interface PluginMetadata {
  id: string;
  name: string;
  version: string;
  description?: string;
  author?: string;
  license?: string;
}

export interface PluginConfig {
  [key: string]: unknown;
}

export interface PluginLogger {
  debug(message: string, ...args: any[]): void;
  info(message: string, ...args: any[]): void;
  warn(message: string, ...args: any[]): void;
  error(message: string, error?: Error, ...args: any[]): void;
}

export interface PluginStorage {
  getItem<T>(key: string): Promise<T | null>;
  setItem<T>(key: string, value: T): Promise<void>;
  removeItem(key: string): Promise<void>;
  clear(): Promise<void>;
}

export interface PluginEventBus {
  emit(event: string, payload: any): void;
  on(event: string, handler: (payload: any) => void): () => void;
  off(event: string, handler: (payload: any) => void): void;
}

export interface PluginContext {
  readonly meta: PluginMetadata;
  readonly logger: PluginLogger;
  readonly storage: PluginStorage;
  readonly events: PluginEventBus;
  
  /**
   * Get typed configuration for this plugin.
   */
  getConfig<T extends PluginConfig>(): T;
}

/**
 * Interface that every plugin must implement.
 */
export interface PluginLifecycle {
  /**
   * Called when the plugin is first loaded.
   * Use this for setting up static resources or validating configuration.
   */
  onInit(ctx: PluginContext): Promise<void> | void;

  /**
   * Called when the system starts or the plugin is enabled.
   * Connect to external services, start listeners, etc.
   */
  onStart?(): Promise<void> | void;

  /**
   * Called when the system stops or the plugin is disabled.
   * Cleanup resources, close connections, etc.
   */
  onStop?(): Promise<void> | void;

  /**
   * Called when the plugin configuration changes at runtime.
   * Implement this for hot-reloading support.
   */
  onConfigChange?(newConfig: PluginConfig): Promise<void> | void;
}

/**
 * Helper type for plugin definition.
 */
export type PluginFactory = () => PluginLifecycle;

/**
 * Skill Tool Parameter Definition
 */
export interface SkillToolParameter {
  name: string;
  type: string;
  description: string;
  required?: boolean;
}

/**
 * Skill Tool Definition
 */
export interface SkillTool {
  name: string;
  description: string;
  parameters: SkillToolParameter[];
  execute(args: Record<string, any>, context: PluginContext): Promise<any>;
}

/**
 * Skill Definition
 */
export interface SkillDefinition {
  name: string;
  version: string;
  description: string;
  tools: SkillTool[];
}

/**
 * Interface for Skill Plugins
 */
export interface SkillPlugin extends PluginLifecycle {
  getSkillDefinition(): SkillDefinition;
}

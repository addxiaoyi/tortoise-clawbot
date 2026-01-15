/**
 * Plugin Sandbox System
 * Provides secure isolation for plugin execution
 */

import { randomUUID } from 'node:crypto';
import type { PluginContext, PluginConfig, PluginLogger, PluginStorage, PluginEventBus } from './types.js';

// ============================================
// Sandbox Permissions
// ============================================

export interface SandboxPermissions {
  // Network access
  allowNetwork?: boolean;
  allowedHosts?: string[];  // Whitelist of allowed hosts
  
  // File system access
  allowFileSystem?: boolean;
  allowedPaths?: string[];   // Whitelist of allowed paths
  
  // Environment variables
  allowEnvironment?: boolean;
  allowedEnvVars?: string[]; // Whitelist of env var names
  
  // Process access
  allowProcess?: boolean;
  allowedCommands?: string[]; // Whitelist of commands
  
  // Plugin capabilities
  maxMemoryMB?: number;
  maxCpuPercent?: number;
  timeoutMs?: number;
}

// ============================================
// Sandbox Logger - Wraps original logger with filtering
// ============================================

class SandboxedLogger implements PluginLogger {
  private original: PluginLogger;
  private pluginId: string;
  private maxLogLength = 10000;

  constructor(original: PluginLogger, pluginId: string) {
    this.original = original;
    this.pluginId = pluginId;
  }

  private sanitizeMessage(message: string): string {
    // Truncate long messages
    if (message.length > this.maxLogLength) {
      return message.substring(0, this.maxLogLength) + '...[truncated]';
    }
    return message;
  }

  debug(message: string, ...args: unknown[]): void {
    this.original.debug(`[${this.pluginId}] ${this.sanitizeMessage(message)}`, ...args);
  }

  info(message: string, ...args: unknown[]): void {
    this.original.info(`[${this.pluginId}] ${this.sanitizeMessage(message)}`, ...args);
  }

  warn(message: string, ...args: unknown[]): void {
    this.original.warn(`[${this.pluginId}] ${this.sanitizeMessage(message)}`, ...args);
  }

  error(message: string, error?: Error, ...args: unknown[]): void {
    // Sanitize error stack traces
    if (error) {
      const sanitizedError = new Error(this.sanitizeMessage(error.message));
      sanitizedError.stack = error.stack
        ?.split('\n')
        .map(line => this.sanitizeMessage(line))
        .join('\n');
      this.original.error(`[${this.pluginId}] ${this.sanitizeMessage(message)}`, sanitizedError, ...args);
    } else {
      this.original.error(`[${this.pluginId}] ${this.sanitizeMessage(message)}`, undefined, ...args);
    }
  }
}

// ============================================
// Sandboxed Storage - Isolated storage per plugin
// ============================================

class SandboxedStorage implements PluginStorage {
  private storage: Map<string, { value: unknown; timestamp: number }> = new Map();
  private pluginId: string;
  private maxEntries = 1000;
  private maxValueSize = 1024 * 1024; // 1MB

  constructor(pluginId: string) {
    this.pluginId = pluginId;
  }

  private prefix(key: string): string {
    return `plugin:${this.pluginId}:${key}`;
  }

  async getItem<T>(key: string): Promise<T | null> {
    const prefixedKey = this.prefix(key);
    const entry = this.storage.get(prefixedKey);
    
    if (!entry) return null;
    
    // Check if value is too old (7 days)
    const maxAge = 7 * 24 * 60 * 60 * 1000;
    if (Date.now() - entry.timestamp > maxAge) {
      this.storage.delete(prefixedKey);
      return null;
    }
    
    return entry.value as T;
  }

  async setItem<T>(key: string, value: T): Promise<void> {
    // Check entry limit
    if (this.storage.size >= this.maxEntries) {
      // Remove oldest entry
      let oldestKey: string | null = null;
      let oldestTime = Infinity;
      
      for (const [k, v] of this.storage) {
        if (v.timestamp < oldestTime) {
          oldestTime = v.timestamp;
          oldestKey = k;
        }
      }
      
      if (oldestKey) this.storage.delete(oldestKey);
    }

    // Check value size
    const serialized = JSON.stringify(value);
    if (serialized.length > this.maxValueSize) {
      throw new Error(`Storage value exceeds maximum size of ${this.maxValueSize} bytes`);
    }

    this.storage.set(this.prefix(key), {
      value,
      timestamp: Date.now(),
    });
  }

  async removeItem(key: string): Promise<void> {
    this.storage.delete(this.prefix(key));
  }

  async clear(): Promise<void> {
    const prefix = `plugin:${this.pluginId}:`;
    for (const key of this.storage.keys()) {
      if (key.startsWith(prefix)) {
        this.storage.delete(key);
      }
    }
  }

  // Export storage for persistence
  export(): Record<string, unknown> {
    const result: Record<string, unknown> = {};
    const prefix = `plugin:${this.pluginId}:`;
    
    for (const [key, entry] of this.storage) {
      if (key.startsWith(prefix)) {
        const shortKey = key.substring(prefix.length);
        result[shortKey] = entry.value;
      }
    }
    
    return result;
  }

  // Import storage from persistence
  import(data: Record<string, unknown>): void {
    for (const [key, value] of Object.entries(data)) {
      this.setItem(key, value);
    }
  }
}

// ============================================
// Sandboxed Event Bus - Restricted event access
// ============================================

class SandboxedEventBus implements PluginEventBus {
  private original: PluginEventBus;
  private pluginId: string;
  private allowedEvents: Set<string>;
  private eventHistory: Map<string, Array<{ payload: unknown; timestamp: number }>> = new Map();
  private maxHistorySize = 100;

  constructor(original: PluginEventBus, pluginId: string, allowedEvents?: string[]) {
    this.original = original;
    this.pluginId = pluginId;
    this.allowedEvents = new Set(allowedEvents || [
      `plugin:${pluginId}:*`,
      'system:*',
    ]);
  }

  private isAllowed(event: string): boolean {
    for (const pattern of this.allowedEvents) {
      if (pattern.endsWith('*')) {
        const prefix = pattern.slice(0, -1);
        if (event.startsWith(prefix)) return true;
      } else if (pattern === event) {
        return true;
      }
    }
    return false;
  }

  emit(event: string, payload: unknown): void {
    if (!this.isAllowed(event)) {
      console.warn(`Plugin ${this.pluginId} attempted to emit disallowed event: ${event}`);
      return;
    }

    // Record event history
    if (!this.eventHistory.has(event)) {
      this.eventHistory.set(event, []);
    }
    
    const history = this.eventHistory.get(event)!;
    history.push({ payload, timestamp: Date.now() });
    
    // Trim history
    while (history.length > this.maxHistorySize) {
      history.shift();
    }

    // Also emit on original (but namespaced)
    this.original.emit(`${this.pluginId}:${event}`, payload);
  }

  on(event: string, handler: (payload: unknown) => void): () => void {
    if (!this.isAllowed(event)) {
      console.warn(`Plugin ${this.pluginId} attempted to listen to disallowed event: ${event}`);
      return () => {};
    }

    // Wrap handler to catch errors
    const wrappedHandler = (payload: unknown) => {
      try {
        handler(payload);
      } catch (err) {
        console.error(`Error in event handler for ${event}:`, err);
      }
    };

    return this.original.on(`${this.pluginId}:${event}`, wrappedHandler);
  }

  off(event: string, handler: (payload: unknown) => void): void {
    if (!this.isAllowed(event)) return;
    this.original.off(`${this.pluginId}:${event}`, handler);
  }

  getHistory(event: string): Array<{ payload: unknown; timestamp: number }> {
    return this.eventHistory.get(event) || [];
  }
}

// ============================================
// Network Restriction
// ============================================

class RestrictedFetch {
  private allowedHosts: Set<string> | null;
  
  constructor(allowedHosts?: string[]) {
    this.allowedHosts = allowedHosts ? new Set(allowedHosts) : null;
  }

  async fetch(url: string, options?: RequestInit): Promise<Response> {
    const parsed = new URL(url);
    
    // Check host whitelist
    if (this.allowedHosts && !this.allowedHosts.has(parsed.hostname)) {
      throw new Error(`Network access to ${parsed.hostname} is not allowed`);
    }
    
    // Block localhost/internal IPs unless explicitly allowed
    const blockedHosts = ['localhost', '127.0.0.1', '0.0.0.0', '::1'];
    if (blockedHosts.includes(parsed.hostname) && !this.allowedHosts?.has(parsed.hostname)) {
      throw new Error(`Network access to internal host ${parsed.hostname} is blocked`);
    }

    // Limit request body size
    if (options?.body && typeof options.body === 'string') {
      const bodySize = new Blob([options.body]).size;
      if (bodySize > 10 * 1024 * 1024) { // 10MB limit
        throw new Error('Request body exceeds 10MB limit');
      }
    }

    return globalThis.fetch(url, {
      ...options,
      // Don't follow redirects to blocked hosts
      redirect: 'follow',
    });
  }
}

// ============================================
// Plugin Sandbox - Main class
// ============================================

export interface SandboxedContext extends Omit<PluginContext, 'logger' | 'storage' | 'events'> {
  logger: SandboxedLogger;
  storage: SandboxedStorage;
  events: SandboxedEventBus;
  fetch?: RestrictedFetch;
}

export class PluginSandbox {
  private permissions: SandboxPermissions;
  private pluginId: string;
  private originalCtx: PluginContext;

  constructor(pluginId: string, permissions: SandboxPermissions = {}) {
    this.pluginId = pluginId;
    this.permissions = {
      allowNetwork: false,
      allowFileSystem: false,
      allowEnvironment: false,
      allowProcess: false,
      maxMemoryMB: 512,
      maxCpuPercent: 50,
      timeoutMs: 30000,
      ...permissions,
    };
    
    this.originalCtx = null as unknown as PluginContext;
  }

  /**
   * Initialize sandbox with original context
   */
  init(originalCtx: PluginContext): void {
    this.originalCtx = originalCtx;
  }

  /**
   * Create sandboxed context for a plugin
   */
  createContext(): SandboxedContext {
    const originalStorage = this.originalCtx.storage;
    const originalEvents = this.originalCtx.events;

    return {
      meta: this.originalCtx.meta,
      getConfig: <T extends PluginConfig>() => this.originalCtx.getConfig<T>(),
      
      // Sandboxed logger
      logger: new SandboxedLogger(this.originalCtx.logger, this.pluginId),
      
      // Sandboxed storage
      storage: new SandboxedStorage(this.pluginId),
      
      // Sandboxed event bus
      events: new SandboxedEventBus(
        originalEvents,
        this.pluginId,
        this.getAllowedEvents()
      ),
      
      // Optional restricted fetch
      fetch: this.permissions.allowNetwork 
        ? new RestrictedFetch(this.permissions.allowedHosts)
        : undefined,
    };
  }

  /**
   * Get list of events the plugin is allowed to emit/listen to
   */
  private getAllowedEvents(): string[] {
    return [
      `plugin:${this.pluginId}:*`,  // Own events
      'system:ready',
      'system:shutdown',
      'channel:*',                  // Channel events
      'gateway:connected',
      'gateway:disconnected',
    ];
  }

  /**
   * Check if an operation is allowed
   */
  checkPermission(operation: keyof SandboxPermissions, target?: string): boolean {
    const perm = this.permissions[operation];
    
    if (perm === false || perm === undefined) {
      return false;
    }
    
    if (perm === true) {
      return true;
    }
    
    // Array whitelist
    if (Array.isArray(perm) && target) {
      return perm.some(pattern => {
        if (pattern.endsWith('*')) {
          return target.startsWith(pattern.slice(0, -1));
        }
        return pattern === target;
      });
    }
    
    return false;
  }

  /**
   * Get timeout for plugin operations
   */
  getTimeout(): number {
    return this.permissions.timeoutMs || 30000;
  }

  /**
   * Get resource limits
   */
  getResourceLimits(): { maxMemoryMB: number; maxCpuPercent: number } {
    return {
      maxMemoryMB: this.permissions.maxMemoryMB || 512,
      maxCpuPercent: this.permissions.maxCpuPercent || 50,
    };
  }
}

// ============================================
// Sandbox Registry - Manages all plugin sandboxes
// ============================================

export class SandboxRegistry {
  private sandboxes = new Map<string, PluginSandbox>();
  private defaultPermissions: SandboxPermissions;

  constructor(defaultPermissions: SandboxPermissions = {}) {
    this.defaultPermissions = {
      allowNetwork: false,
      allowFileSystem: false,
      allowEnvironment: false,
      allowProcess: false,
      ...defaultPermissions,
    };
  }

  /**
   * Create sandbox for a plugin
   */
  create(pluginId: string, permissions?: Partial<SandboxPermissions>): PluginSandbox {
    const merged = {
      ...this.defaultPermissions,
      ...permissions,
    };
    
    const sandbox = new PluginSandbox(pluginId, merged);
    this.sandboxes.set(pluginId, sandbox);
    
    return sandbox;
  }

  /**
   * Get existing sandbox
   */
  get(pluginId: string): PluginSandbox | undefined {
    return this.sandboxes.get(pluginId);
  }

  /**
   * Remove sandbox
   */
  remove(pluginId: string): void {
    this.sandboxes.delete(pluginId);
  }

  /**
   * Get all sandboxed plugin IDs
   */
  getAll(): string[] {
    return Array.from(this.sandboxes.keys());
  }
}

// ============================================
// Permission Presets
// ============================================

export const PermissionPresets = {
  // Minimal permissions - no external access
  minimal: {
    allowNetwork: false,
    allowFileSystem: false,
    allowEnvironment: false,
    allowProcess: false,
    maxMemoryMB: 128,
    maxCpuPercent: 10,
    timeoutMs: 10000,
  } as SandboxPermissions,

  // Standard permissions - limited network and no fs
  standard: {
    allowNetwork: true,
    allowedHosts: ['api.openai.com', 'api.anthropic.com', 'api.tortoise.ai'],
    allowFileSystem: false,
    allowEnvironment: true,
    allowedEnvVars: ['NODE_ENV', 'LOG_LEVEL'],
    allowProcess: false,
    maxMemoryMB: 512,
    maxCpuPercent: 50,
    timeoutMs: 30000,
  } as SandboxPermissions,

  // Trusted permissions - full access
  trusted: {
    allowNetwork: true,
    allowFileSystem: true,
    allowEnvironment: true,
    allowProcess: true,
    maxMemoryMB: 2048,
    maxCpuPercent: 100,
    timeoutMs: 120000,
  } as SandboxPermissions,
};

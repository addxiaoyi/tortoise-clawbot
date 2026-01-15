/**
 * Simple Dependency Injection Container for MCP.
 */

export class Container {
  private readonly services = new Map<string, any>();
  private factories = new Map<string, () => any>();

  /**
   * Register a singleton service instance.
   */
  public register<T>(id: string, instance: T): void {
    this.services.set(id, instance);
  }

  /**
   * Register a factory function that creates the service on demand.
   * By default, it's treated as a singleton (cached after first create).
   */
  public registerFactory<T>(id: string, factory: () => T): void {
    this.factories.set(id, factory);
  }

  /**
   * Resolve a service by ID.
   * Throws if not found.
   */
  public resolve<T>(id: string): T {
    if (this.services.has(id)) {
      return this.services.get(id);
    }

    if (this.factories.has(id)) {
      const factory = this.factories.get(id)!;
      const instance = factory();
      // Cache it (Singleton behavior)
      this.services.set(id, instance);
      return instance;
    }

    throw new Error(`Service not registered: ${id}`);
  }

  /**
   * Check if a service is registered.
   */
  public has(id: string): boolean {
    return this.services.has(id) || this.factories.has(id);
  }

  /**
   * Clear all services.
   */
  public clear(): void {
    this.services.clear();
    this.factories.clear();
  }
}

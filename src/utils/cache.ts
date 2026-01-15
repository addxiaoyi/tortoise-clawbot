/**
 * Simple LRU (Least Recently Used) Cache implementation.
 */

export interface CacheOptions {
  /** Maximum number of items in the cache */
  maxSize: number;
  /** Default time-to-live in milliseconds */
  ttl?: number;
}

interface CacheEntry<T> {
  value: T;
  expiresAt?: number;
}

export class CacheWrapper<T> {
  private readonly cache: Map<string, CacheEntry<T>>;
  private readonly maxSize: number;
  private readonly defaultTtl?: number;

  constructor(options: CacheOptions) {
    this.cache = new Map();
    this.maxSize = options.maxSize;
    this.defaultTtl = options.ttl;
  }

  public get(key: string): T | undefined {
    const entry = this.cache.get(key);
    
    if (!entry) {
      return undefined;
    }

    // Check expiration
    if (entry.expiresAt && Date.now() > entry.expiresAt) {
      this.cache.delete(key);
      return undefined;
    }

    // Refresh LRU order: delete and re-set (Map preserves insertion order)
    this.cache.delete(key);
    this.cache.set(key, entry);

    return entry.value;
  }

  public set(key: string, value: T, ttl?: number): void {
    // 1. If exists, remove to update position (LRU)
    if (this.cache.has(key)) {
      this.cache.delete(key);
    } 
    // 2. If new and full, evict oldest (first item in Map)
    else if (this.cache.size >= this.maxSize) {
      const firstKey = this.cache.keys().next().value;
      if (firstKey !== undefined) {
        this.cache.delete(firstKey);
      }
    }

    const effectiveTtl = ttl ?? this.defaultTtl;
    const expiresAt = effectiveTtl ? Date.now() + effectiveTtl : undefined;

    this.cache.set(key, { value, expiresAt });
  }

  public has(key: string): boolean {
    return this.get(key) !== undefined;
  }

  public delete(key: string): void {
    this.cache.delete(key);
  }

  public clear(): void {
    this.cache.clear();
  }
  
  public size(): number {
    return this.cache.size;
  }
}

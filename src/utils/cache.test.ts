import { describe, it, expect, vi } from 'vitest';
import { CacheWrapper } from './cache.js';

describe('CacheWrapper', () => {
  it('should store and retrieve values', () => {
    const cache = new CacheWrapper<string>({ maxSize: 10 });
    cache.set('key', 'value');
    expect(cache.get('key')).toBe('value');
  });

  it('should return undefined for missing keys', () => {
    const cache = new CacheWrapper<string>({ maxSize: 10 });
    expect(cache.get('missing')).toBeUndefined();
  });

  it('should respect TTL', () => {
    vi.useFakeTimers();
    const cache = new CacheWrapper<string>({ maxSize: 10 });
    
    cache.set('key', 'value', 100);
    expect(cache.get('key')).toBe('value');
    
    vi.advanceTimersByTime(101);
    expect(cache.get('key')).toBeUndefined();
    
    vi.useRealTimers();
  });

  it('should evict least recently used items', () => {
    const cache = new CacheWrapper<string>({ maxSize: 2 });
    
    cache.set('a', '1');
    cache.set('b', '2');
    
    // Access 'a' to make it most recently used
    cache.get('a');
    
    // Add 'c', should evict 'b' (since 'a' was just used)
    cache.set('c', '3');
    
    expect(cache.get('a')).toBe('1');
    expect(cache.get('b')).toBeUndefined();
    expect(cache.get('c')).toBe('3');
  });

  it('should update LRU order on set', () => {
    const cache = new CacheWrapper<string>({ maxSize: 2 });
    
    cache.set('a', '1');
    cache.set('b', '2');
    
    // Update 'a', making it most recent
    cache.set('a', '1-updated');
    
    // Add 'c', should evict 'b'
    cache.set('c', '3');
    
    expect(cache.get('a')).toBe('1-updated');
    expect(cache.get('b')).toBeUndefined();
    expect(cache.get('c')).toBe('3');
  });
});

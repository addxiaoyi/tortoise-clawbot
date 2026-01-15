import { describe, it, expect } from 'vitest';
import { deepMerge } from './deep-merge.js';

describe('deepMerge', () => {
  it('should merge nested objects', () => {
    const target = {
      a: 1,
      b: { c: 2 }
    };
    const source = {
      b: { d: 3 },
      e: 4
    };
    
    const result = deepMerge(target, source as any);
    
    expect(result).toEqual({
      a: 1,
      b: { c: 2, d: 3 },
      e: 4
    });
  });

  it('should overwrite primitive values', () => {
    const target = { a: 1 };
    const source = { a: 2 };
    
    expect(deepMerge(target, source)).toEqual({ a: 2 });
  });

  it('should handle arrays (replace, not merge)', () => {
    const target = { a: [1, 2] };
    const source = { a: [3, 4] };
    
    expect(deepMerge(target, source)).toEqual({ a: [3, 4] });
  });
  
  it('should handle multiple sources', () => {
      const t = { a: 1 };
      const s1 = { b: 2 };
      const s2 = { c: 3 };
      
      expect(deepMerge(t, s1 as any, s2 as any)).toEqual({ a: 1, b: 2, c: 3 });
  });

  it('should not merge prototype-pollution keys', () => {
    const target: Record<string, unknown> = { a: 1 };
    const source: Record<string, unknown> = { b: 2 };
    Object.defineProperty(source, '__proto__', {
      enumerable: true,
      configurable: true,
      value: { polluted: true },
    });
    Object.defineProperty(source, 'constructor', {
      enumerable: true,
      configurable: true,
      value: { evil: true },
    });
    deepMerge(target, source as any);
    expect(target.b).toBe(2);
    expect(Object.prototype.hasOwnProperty.call(target, '__proto__')).toBe(false);
    expect((target as { polluted?: boolean }).polluted).toBeUndefined();
  });
});

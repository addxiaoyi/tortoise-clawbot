/**
 * Deep merge implementation for objects.
 * Useful for merging configuration objects.
 */

function isObject(item: unknown): item is Record<string, unknown> {
  return (item !== null && typeof item === 'object' && !Array.isArray(item));
}

/** 禁止通过合并键污染原型链 */
const UNSAFE_MERGE_KEYS = new Set(['__proto__', 'constructor', 'prototype']);

export function deepMerge<T extends Record<string, unknown>>(target: T, ...sources: Partial<T>[]): T {
  if (!sources.length) return target;
  const source = sources.shift();
  if (!source) return target;

  if (isObject(target) && isObject(source)) {
    for (const key of Object.keys(source)) {
      if (UNSAFE_MERGE_KEYS.has(key)) {
        continue;
      }
      if (isObject(source[key])) {
        if (!target[key]) {
          Object.assign(target, { [key]: {} });
        }
        deepMerge(target[key] as Record<string, unknown>, source[key] as Record<string, unknown>);
      } else {
        Object.assign(target, { [key]: source[key] });
      }
    }
  }

  return deepMerge(target, ...sources);
}

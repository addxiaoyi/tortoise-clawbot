import { describe, it, expect } from 'vitest';
import { normalizePath } from './path-normalizer.js';
import path from 'node:path';

describe('PathNormalizer', () => {
  it('should normalize windows paths', () => {
    const winPath = String.raw`C:\Users\test\file.txt`;
    const expected = 'C:/Users/test/file.txt';
    expect(normalizePath(winPath)).toBe(expected);
  });

  it('should keep posix paths', () => {
    const posixPath = '/usr/local/bin';
    expect(normalizePath(posixPath)).toBe(posixPath);
  });

  it('should handle mixed paths', () => {
    const mixed = 'src/utils\\test.ts';
    expect(normalizePath(mixed)).toBe('src/utils/test.ts');
  });
});

import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { EnvLoader } from './env-loader.js';

describe('EnvLoader', () => {
  const originalEnv = process.env;

  beforeEach(() => {
    process.env = { ...originalEnv };
  });

  afterEach(() => {
    process.env = originalEnv;
  });

  it('should get string value', () => {
    process.env.TEST_KEY = 'test';
    expect(EnvLoader.getString('TEST_KEY')).toBe('test');
  });

  it('should throw for missing required string', () => {
    expect(() => EnvLoader.getString('MISSING_KEY')).toThrow();
  });

  it('should get optional string with default', () => {
    expect(EnvLoader.getStringOptional('MISSING_KEY', 'default')).toBe('default');
  });

  it('should get number value', () => {
    process.env.TEST_NUM = '123';
    expect(EnvLoader.getNumber('TEST_NUM')).toBe(123);
  });

  it('should throw for invalid number', () => {
    process.env.TEST_INVALID = 'abc';
    expect(() => EnvLoader.getNumber('TEST_INVALID')).toThrow();
  });

  it('should get boolean value', () => {
    process.env.TEST_TRUE = 'true';
    process.env.TEST_1 = '1';
    process.env.TEST_FALSE = 'false';
    
    expect(EnvLoader.getBoolean('TEST_TRUE')).toBe(true);
    expect(EnvLoader.getBoolean('TEST_1')).toBe(true);
    expect(EnvLoader.getBoolean('TEST_FALSE')).toBe(false);
    expect(EnvLoader.getBoolean('MISSING_BOOL', true)).toBe(true); // Default true
  });
});

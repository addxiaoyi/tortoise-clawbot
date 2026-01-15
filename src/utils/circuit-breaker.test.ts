import { describe, it, expect, vi } from 'vitest';
import { CircuitBreaker, CircuitState } from './circuit-breaker.js';

describe('CircuitBreaker', () => {
  it('should initially be closed', () => {
    const breaker = new CircuitBreaker({
      failureThreshold: 3,
      resetTimeout: 1000
    });
    expect(breaker.getState()).toBe(CircuitState.CLOSED);
  });

  it('should open after threshold failures', async () => {
    const breaker = new CircuitBreaker({
      failureThreshold: 2,
      resetTimeout: 1000
    });
    
    // 1st failure
    try {
      await breaker.execute(async () => { throw new Error('fail'); });
    } catch {}
    expect(breaker.getState()).toBe(CircuitState.CLOSED);

    // 2nd failure
    try {
      await breaker.execute(async () => { throw new Error('fail'); });
    } catch {}
    expect(breaker.getState()).toBe(CircuitState.OPEN);
  });

  it('should reject requests when open', async () => {
    const breaker = new CircuitBreaker({
      failureThreshold: 1,
      resetTimeout: 1000
    });
    
    // Fail once
    try {
      await breaker.execute(async () => { throw new Error('fail'); });
    } catch {}
    
    expect(breaker.getState()).toBe(CircuitState.OPEN);
    
    // Should throw specific error
    await expect(breaker.execute(async () => 1)).rejects.toThrow('Circuit Breaker is OPEN');
  });

  it('should transition to half-open after timeout', async () => {
    vi.useFakeTimers();
    
    const breaker = new CircuitBreaker({
      failureThreshold: 1,
      resetTimeout: 1000
    });
    
    // Fail once
    try {
      await breaker.execute(async () => { throw new Error('fail'); });
    } catch {}
    
    expect(breaker.getState()).toBe(CircuitState.OPEN);
    
    // Advance time
    vi.advanceTimersByTime(1100);
    
    // Execute successful call
    const result = await breaker.execute(async () => 'success');
    expect(result).toBe('success');
    expect(breaker.getState()).toBe(CircuitState.CLOSED);
    
    vi.useRealTimers();
  });
});

import { describe, it, expect, vi } from 'vitest';
import { RateLimiter } from './rate-limiter.js';

describe('RateLimiter', () => {
  it('should initialize with full capacity', () => {
    const limiter = new RateLimiter({
      capacity: 10,
      refillRate: 1,
      interval: 100
    });
    expect(limiter.getTokens()).toBe(10);
  });

  it('should consume tokens', () => {
    const limiter = new RateLimiter({
      capacity: 10,
      refillRate: 1,
      interval: 1000
    });
    
    expect(limiter.tryConsume(5)).toBe(true);
    expect(limiter.getTokens()).toBe(5);
  });

  it('should fail if not enough tokens', () => {
    const limiter = new RateLimiter({
      capacity: 5,
      refillRate: 1,
      interval: 1000
    });
    
    expect(limiter.tryConsume(6)).toBe(false);
  });

  it('should refill over time', async () => {
    const limiter = new RateLimiter({
      capacity: 5,
      refillRate: 5,
      interval: 100 // Refill 5 tokens every 100ms
    });
    
    // Consume all
    expect(limiter.tryConsume(5)).toBe(true);
    expect(limiter.getTokens()).toBe(0);
    
    // Wait for refill
    await new Promise(resolve => setTimeout(resolve, 150));
    
    // Should have refilled
    expect(limiter.getTokens()).toBe(5);
  });
});

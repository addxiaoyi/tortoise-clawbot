import { describe, it, expect, vi } from 'vitest';
import { Backoff } from './backoff.js';

describe('Backoff', () => {
  it('should increase delay exponentially', () => {
    const backoff = new Backoff({
      initialDelay: 100,
      factor: 2,
      jitter: 0,
      maxRetries: 5
    });

    expect(backoff.next()).toBe(100);
    expect(backoff.next()).toBe(200);
    expect(backoff.next()).toBe(400);
    expect(backoff.next()).toBe(800);
    expect(backoff.next()).toBe(1600);
    expect(backoff.next()).toBe(-1); // Max retries exceeded
  });

  it('should respect max delay', () => {
    const backoff = new Backoff({
      initialDelay: 100,
      factor: 10,
      maxDelay: 500,
      jitter: 0
    });

    expect(backoff.next()).toBe(100);
    expect(backoff.next()).toBe(500); // 1000 capped at 500
  });

  it('should apply jitter', () => {
    const backoff = new Backoff({
      initialDelay: 100,
      factor: 1,
      jitter: 0.1,
      maxRetries: 10
    });

    const delays: number[] = [];
    for (let i = 0; i < 10; i++) {
      delays.push(backoff.next());
    }
    
    // Check if delays vary slightly around 100
    // Range: 90 - 110
    const allInRange = delays.every(d => d >= 90 && d <= 110);
    const someVariation = new Set(delays).size > 1;

    expect(allInRange).toBe(true);
    expect(someVariation).toBe(true);
  });
});

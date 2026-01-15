/**
 * Token Bucket Rate Limiter.
 * Useful for controlling the rate of API requests or other operations.
 */

export interface RateLimiterOptions {
  capacity: number;
  refillRate: number;
  interval: number;
  minWaitMs?: number;
}

export class RateLimiter {
  private tokens: number;
  private lastRefill: number;
  private readonly capacity: number;
  private readonly refillRate: number;
  private readonly interval: number;
  private readonly minWaitMs: number;

  constructor(options: RateLimiterOptions) {
    this.capacity = options.capacity;
    this.refillRate = options.refillRate;
    this.interval = options.interval;
    this.minWaitMs = options.minWaitMs ?? 10;
    this.tokens = options.capacity;
    this.lastRefill = Date.now();
  }

  private refill() {
    const now = Date.now();
    const elapsed = now - this.lastRefill;
    
    if (elapsed >= this.interval) {
      const tokensToAdd = Math.floor(elapsed / this.interval) * this.refillRate;
      this.tokens = Math.min(this.capacity, this.tokens + tokensToAdd);
      this.lastRefill = now;
    }
  }

  /**
   * Attempt to consume tokens.
   * @param count Number of tokens to consume (default: 1)
   * @returns true if tokens were consumed, false otherwise
   */
  public tryConsume(count: number = 1): boolean {
    this.refill();
    
    if (this.tokens >= count) {
      this.tokens -= count;
      return true;
    }
    
    return false;
  }

  /**
   * Wait until enough tokens are available.
   * @param count Number of tokens to consume (default: 1)
   * @param maxWaitMs Maximum time to wait in milliseconds (default: 30000)
   */
  public async consume(count: number = 1, maxWaitMs: number = 30000): Promise<void> {
    const startTime = Date.now();
    while (!this.tryConsume(count)) {
      const now = Date.now();
      if (now - startTime >= maxWaitMs) {
        throw new Error(`Rate limiter wait timeout after ${maxWaitMs}ms`);
      }
      const timeToRefill = this.interval - (now - this.lastRefill);
      await new Promise(resolve => setTimeout(resolve, Math.max(this.minWaitMs, timeToRefill)));
    }
  }
  
  public getTokens(): number {
    this.refill();
    return this.tokens;
  }
}

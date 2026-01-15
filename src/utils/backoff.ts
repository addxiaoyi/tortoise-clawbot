/**
 * Exponential Backoff implementation.
 * Adapted from common open-source patterns (e.g., in Google Cloud client libraries, AWS SDKs).
 */

export interface BackoffOptions {
  /** Initial delay in milliseconds */
  initialDelay?: number;
  /** Maximum delay in milliseconds */
  maxDelay?: number;
  /** Multiplier for the delay */
  factor?: number;
  /** Jitter factor (0-1) to randomize delay */
  jitter?: number;
  /** Maximum number of retries */
  maxRetries?: number;
}

const DEFAULT_OPTIONS: Required<BackoffOptions> = {
  initialDelay: 100,
  maxDelay: 10000,
  factor: 2,
  jitter: 0.1,
  maxRetries: 10
};

export class Backoff {
  private readonly options: Required<BackoffOptions>;
  private attempt: number = 0;

  constructor(options: BackoffOptions = {}) {
    this.options = { ...DEFAULT_OPTIONS, ...options };
  }

  /**
   * Calculate the next delay in milliseconds.
   * Returns -1 if max retries exceeded.
   */
  public next(): number {
    if (this.attempt >= this.options.maxRetries) {
      return -1;
    }

    const { initialDelay, maxDelay, factor, jitter } = this.options;
    
    // Calculate exponential delay
    let delay = initialDelay * Math.pow(factor, this.attempt);
    
    // Apply jitter: delay = delay * (1 + (random() * 2 - 1) * jitter)
    // This gives a range of [delay * (1-jitter), delay * (1+jitter)]
    if (jitter > 0) {
      const jitterMultiplier = 1 + (Math.random() * 2 - 1) * jitter;
      delay = delay * jitterMultiplier;
    }

    // Cap at maxDelay
    delay = Math.min(delay, maxDelay);
    
    this.attempt++;
    
    return Math.round(delay);
  }

  public reset(): void {
    this.attempt = 0;
  }
  
  public getAttempt(): number {
    return this.attempt;
  }
}

/**
 * Helper to run a function with exponential backoff.
 */
export async function withBackoff<T>(
  fn: () => Promise<T>,
  options: BackoffOptions = {}
): Promise<T> {
  const backoff = new Backoff(options);
  let lastError: any;

  while (true) {
    try {
      return await fn();
    } catch (error: any) {
      lastError = error;
      
      // Check if we should retry based on error type if needed (not implemented here for simplicity)
      // For now, retry on all errors
      
      const delay = backoff.next();
      if (delay === -1) {
        throw lastError;
      }
      
      await new Promise(resolve => setTimeout(resolve, delay));
    }
  }
}

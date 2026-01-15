/**
 * Feature Flag / Rollout Strategy implementation.
 */

export interface RolloutContext {
  userId?: string;
  sessionId?: string;
  attributes?: Record<string, any>;
}

export interface RolloutStrategy {
  isEnabled(featureKey: string, context: RolloutContext): boolean;
}

export class PercentageRollout implements RolloutStrategy {
  private percentage: number;

  constructor(percentage: number) {
    if (percentage < 0 || percentage > 100) {
      throw new Error('Percentage must be between 0 and 100');
    }
    this.percentage = percentage;
  }

  public isEnabled(featureKey: string, context: RolloutContext): boolean {
    const seed = featureKey + (context.userId || context.sessionId || Math.random().toString());
    const hash = this.hashString(seed);
    return (hash % 100) < this.percentage;
  }

  private hashString(str: string): number {
    let hash = 0;
    for (let i = 0; i < str.length; i++) {
      const codePoint = str.codePointAt(i) ?? 0;
      hash = ((hash << 5) - hash) + codePoint;
      hash = hash & hash; // Convert to 32bit integer
    }
    return Math.abs(hash);
  }
}

export class UserListRollout implements RolloutStrategy {
  private readonly allowedUsers: Set<string>;

  constructor(allowedUsers: string[]) {
    this.allowedUsers = new Set(allowedUsers);
  }

  public isEnabled(featureKey: string, context: RolloutContext): boolean {
    if (!context.userId) return false;
    return this.allowedUsers.has(context.userId);
  }
}

export class CompositeRollout implements RolloutStrategy {
  private strategies: RolloutStrategy[];

  constructor(strategies: RolloutStrategy[]) {
    this.strategies = strategies;
  }

  public isEnabled(featureKey: string, context: RolloutContext): boolean {
    // OR logic: enabled if ANY strategy allows it
    return this.strategies.some(s => s.isEnabled(featureKey, context));
  }
}


import { describe, it, expect, beforeEach } from 'vitest';
import { ABTestManager, Experiment } from './ab-test';

describe('ABTestManager', () => {
  let manager: ABTestManager;

  beforeEach(() => {
    manager = new ABTestManager();
  });

  it('should register an experiment', () => {
    const experiment: Experiment = {
      id: 'test-exp',
      variants: [
        { name: 'control', weight: 50 },
        { name: 'treatment', weight: 50 }
      ]
    };
    manager.registerExperiment(experiment);
    expect(manager.getExperiment('test-exp')).toEqual(experiment);
  });

  it('should assign variant deterministically', () => {
    const experiment: Experiment = {
      id: 'test-exp',
      variants: [
        { name: 'control', weight: 50 },
        { name: 'treatment', weight: 50 }
      ]
    };
    manager.registerExperiment(experiment);

    // Same user should get same variant
    const variant1 = manager.assignVariant('test-exp', 'user1');
    const variant2 = manager.assignVariant('test-exp', 'user1');
    expect(variant1).toBe(variant2);

    // Different users might get different variants (probabilistic but deterministic per user)
    // We can't guarantee different variants without knowing the hash function details,
    // but we can verify consistency.
    const variant3 = manager.assignVariant('test-exp', 'user2');
    expect(manager.assignVariant('test-exp', 'user2')).toBe(variant3);
  });

  it('should return null for unknown experiment', () => {
    expect(manager.assignVariant('unknown', 'user1')).toBeNull();
  });

  it('should handle weighted distribution', () => {
    const experiment: Experiment = {
      id: 'weighted-exp',
      variants: [
        { name: 'A', weight: 100 },
        { name: 'B', weight: 0 }
      ]
    };
    manager.registerExperiment(experiment);

    expect(manager.assignVariant('weighted-exp', 'user1')).toBe('A');
    expect(manager.assignVariant('weighted-exp', 'user2')).toBe('A');
  });
});

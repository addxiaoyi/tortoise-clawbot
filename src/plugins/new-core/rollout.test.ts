import { describe, it, expect } from 'vitest';
import { PercentageRollout, UserListRollout, CompositeRollout } from './rollout.js';

describe('Rollout Strategies', () => {
  describe('PercentageRollout', () => {
    it('should respect 0%', () => {
      const rollout = new PercentageRollout(0);
      expect(rollout.isEnabled('feature', { userId: 'user1' })).toBe(false);
      expect(rollout.isEnabled('feature', { userId: 'user2' })).toBe(false);
    });

    it('should respect 100%', () => {
      const rollout = new PercentageRollout(100);
      expect(rollout.isEnabled('feature', { userId: 'user1' })).toBe(true);
      expect(rollout.isEnabled('feature', { userId: 'user2' })).toBe(true);
    });

    it('should be deterministic for same user', () => {
      const rollout = new PercentageRollout(50);
      const result1 = rollout.isEnabled('feature', { userId: 'user1' });
      const result2 = rollout.isEnabled('feature', { userId: 'user1' });
      expect(result1).toBe(result2);
    });
  });

  describe('UserListRollout', () => {
    it('should allow specific users', () => {
      const rollout = new UserListRollout(['alice', 'bob']);
      expect(rollout.isEnabled('feature', { userId: 'alice' })).toBe(true);
      expect(rollout.isEnabled('feature', { userId: 'charlie' })).toBe(false);
    });
  });

  describe('CompositeRollout', () => {
    it('should use OR logic', () => {
      const s1 = new UserListRollout(['alice']);
      const s2 = new PercentageRollout(0); // Disabled
      const composite = new CompositeRollout([s1, s2]);

      expect(composite.isEnabled('feature', { userId: 'alice' })).toBe(true);
      expect(composite.isEnabled('feature', { userId: 'bob' })).toBe(false);
    });
  });
});

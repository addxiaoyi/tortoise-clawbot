import { describe, it, expect } from 'vitest';
import { LabelMatcher } from './label-matcher.js';

describe('LabelMatcher', () => {
  it('should match bug label', () => {
    const matcher = new LabelMatcher();
    const labels = matcher.matchLabels('This is a critical crash in production');
    expect(labels).toContain('bug');
  });

  it('should match enhancement label', () => {
    const matcher = new LabelMatcher();
    const labels = matcher.matchLabels('Please add a new feature');
    expect(labels).toContain('enhancement');
  });

  it('should match multiple labels', () => {
    const matcher = new LabelMatcher();
    const labels = matcher.matchLabels('Fix the crash and add documentation');
    expect(labels).toContain('bug');
    expect(labels).toContain('documentation');
  });

  it('should return empty array for no match', () => {
    const matcher = new LabelMatcher();
    const labels = matcher.matchLabels('Just some random text');
    expect(labels).toEqual([]);
  });

  it('should handle custom config path (failure fallback)', () => {
    const matcher = new LabelMatcher('non-existent.json');
    expect(matcher.getRules()).toBeDefined();
    expect(matcher.matchLabels('bug')).toContain('bug');
  });
});

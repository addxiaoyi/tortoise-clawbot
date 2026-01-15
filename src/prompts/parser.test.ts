
import { describe, it, expect } from 'vitest';
import { StructureParser } from './parser';

describe('StructureParser', () => {
  const parser = new StructureParser();

  describe('parseJSON', () => {
    it('should parse simple JSON string', () => {
      const input = '{"key": "value"}';
      expect(parser.parseJSON(input)).toEqual({ key: 'value' });
    });

    it('should parse JSON from markdown code block', () => {
      const input = 'Here is the result:\n```json\n{\n  "key": "value"\n}\n```';
      expect(parser.parseJSON(input)).toEqual({ key: 'value' });
    });

    it('should parse JSON from markdown code block without language identifier', () => {
      const input = '```\n{\n  "key": "value"\n}\n```';
      expect(parser.parseJSON(input)).toEqual({ key: 'value' });
    });

    it('should throw error for invalid JSON', () => {
      const input = 'invalid json';
      expect(() => parser.parseJSON(input)).toThrow();
    });
  });

  describe('extractTag', () => {
    it('should extract content within tags', () => {
      const input = '<result>content</result>';
      expect(parser.extractTag(input, 'result')).toBe('content');
    });

    it('should return null if tag not found', () => {
      const input = 'no tags here';
      expect(parser.extractTag(input, 'result')).toBeNull();
    });

    it('should handle multiline content', () => {
      const input = '<result>\nline1\nline2\n</result>';
      expect(parser.extractTag(input, 'result')).toBe('\nline1\nline2\n');
    });
  });
});

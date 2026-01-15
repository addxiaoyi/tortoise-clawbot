import { describe, it, expect } from 'vitest';
import { contentTypeForFile, shouldServeFileAsUtf8Text } from './mime.js';

describe('canvas mime', () => {
  it('maps common extensions', () => {
    expect(contentTypeForFile('/a/b/index.html')).toContain('text/html');
    expect(contentTypeForFile('x.JSON')).toContain('application/json');
    expect(contentTypeForFile('p.png')).toBe('image/png');
    expect(contentTypeForFile('u.unknown')).toBe('application/octet-stream');
  });

  it('classifies text vs binary', () => {
    expect(shouldServeFileAsUtf8Text('/r/file.html')).toBe(true);
    expect(shouldServeFileAsUtf8Text('/r/x.js')).toBe(true);
    expect(shouldServeFileAsUtf8Text('/r/p.png')).toBe(false);
  });
});

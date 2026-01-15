import { describe, it, expect } from 'vitest';
import { PromptRegistry } from './registry.js';
import { join } from 'node:path';

describe('PromptRegistry', () => {
  // We use the actual templates dir for testing since we created the file
  const registry = new PromptRegistry();

  it('should load system prompt v1', () => {
    const tpl = registry.load('system', 'v1');
    expect(tpl).toBeDefined();
    // Verify it's the correct template by rendering something
    const output = tpl.render({ skillsPrompt: 'test' });
    expect(output).toContain('## Skills (mandatory)');
  });

  it('should load latest system prompt', () => {
    const tpl = registry.load('system', 'latest');
    expect(tpl).toBeDefined();
  });

  it('should throw for unknown prompt', () => {
    expect(() => registry.load('unknown')).toThrow();
  });
});

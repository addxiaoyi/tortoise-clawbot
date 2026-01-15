import { describe, it, expect } from 'vitest';
import { SystemPromptBuilder } from './system-prompt-builder.js';

describe('SystemPromptBuilder', () => {
  it('should build full prompt', () => {
    const builder = new SystemPromptBuilder();
    const prompt = builder.build({
      mode: 'full',
      skillsPrompt: '  Some skills  ',
      readToolName: 'read',
      availableTools: new Set(['memory_search', 'memory_get']),
      citationsMode: 'auto',
      ownerLine: 'Authorized: Alice',
      userTimezone: 'UTC',
      currentTime: '12:00'
    });

    expect(prompt).toContain('## Skills (mandatory)');
    expect(prompt).toContain('Some skills');
    expect(prompt).toContain('## Memory Recall');
    expect(prompt).toContain('Citations: include Source');
    expect(prompt).toContain('## Authorized Senders');
    expect(prompt).toContain('Authorized: Alice');
    expect(prompt).toContain('## Time');
    expect(prompt).toContain('12:00 (UTC)');
  });

  it('should handle minimal mode', () => {
    const builder = new SystemPromptBuilder();
    const prompt = builder.build({
      mode: 'minimal',
      skillsPrompt: 'Skills',
      availableTools: new Set(['memory_search']),
      ownerLine: 'Alice'
    });

    expect(prompt).toContain('## Skills (mandatory)');
    expect(prompt).not.toContain('## Memory Recall');
    expect(prompt).not.toContain('## Authorized Senders');
  });

  it('should handle disabled memory', () => {
    const builder = new SystemPromptBuilder();
    const prompt = builder.build({
      mode: 'full',
      skillsPrompt: 'Skills',
      availableTools: new Set(['other_tool']) // No memory tools
    });

    expect(prompt).not.toContain('## Memory Recall');
  });
});

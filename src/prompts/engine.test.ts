import { describe, it, expect } from 'vitest';
import { PromptTemplate } from './engine.js';

describe('PromptTemplate', () => {
  it('should replace variables', () => {
    const tpl = new PromptTemplate('Hello {{name}}!');
    expect(tpl.render({ name: 'World' })).toBe('Hello World!');
  });

  it('should handle nested variables', () => {
    const tpl = new PromptTemplate('Hello {{user.name}}!');
    expect(tpl.render({ user: { name: 'Alice' } })).toBe('Hello Alice!');
  });

  it('should handle conditionals', () => {
    const tpl = new PromptTemplate('{{#if show}}Shown{{/if}}{{#if hide}}Hidden{{/if}}');
    expect(tpl.render({ show: true, hide: false })).toBe('Shown');
  });

  it('should handle loops', () => {
    const tpl = new PromptTemplate('Items: {{#each items}}- {{this}}\n{{/each}}');
    expect(tpl.render({ items: ['A', 'B'] })).toBe('Items: - A\n- B\n');
  });

  it('should handle loops with objects', () => {
    const tpl = new PromptTemplate('Users: {{#each users}}{{name}} ({{role}})\n{{/each}}');
    expect(tpl.render({ users: [{ name: 'Alice', role: 'Admin' }, { name: 'Bob', role: 'User' }] }))
      .toBe('Users: Alice (Admin)\nBob (User)\n');
  });
});

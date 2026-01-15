
import { describe, it, expect, beforeEach } from 'vitest';
import { ContextManager, Message } from './context-manager';

describe('ContextManager', () => {
  let manager: ContextManager;

  beforeEach(() => {
    manager = new ContextManager(100);
  });

  it('should add messages correctly', () => {
    const msg: Message = { role: 'user', content: 'hello' };
    manager.addMessage(msg);
    expect(manager.getMessages()).toEqual([msg]);
  });

  it('should estimate token count (heuristic)', () => {
    const content = '1234'; 
    expect(manager.estimateTokens(content)).toBe(1);
  });

  it('should truncate oldest messages when exceeding limit', () => {
    // 200 chars = 50 tokens
    manager.addMessage({ role: 'user', content: 'a'.repeat(200) }); 
    manager.addMessage({ role: 'assistant', content: 'b'.repeat(200) }); 
    
    expect(manager.getMessages().length).toBe(2);

    manager.addMessage({ role: 'user', content: 'c'.repeat(200) }); 

    // Total 150 > 100. Should remove first (a).
    // Remaining: b (50), c (50) -> 100 <= 100.
    const messages = manager.getMessages();
    expect(messages.length).toBe(2);
    expect(messages[0].content).toContain('b');
    expect(messages[1].content).toContain('c');
  });

  it('should preserve system messages', () => {
    manager.addMessage({ role: 'system', content: 'sys' }); // 1 token
    manager.addMessage({ role: 'user', content: 'a'.repeat(200) }); // 50
    manager.addMessage({ role: 'assistant', content: 'b'.repeat(200) }); // 50
    
    // Total 101 > 100. Remove 'a'.
    // Remaining: sys (1), b (50) -> 51 <= 100.
    
    const messages = manager.getMessages();
    expect(messages.length).toBe(2);
    expect(messages[0].role).toBe('system');
    expect(messages[1].content).toContain('b');
  });
});

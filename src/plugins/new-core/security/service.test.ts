import { describe, it, expect, vi } from 'vitest';
import { SecurityService } from './service.js';
import type { PluginLogger } from '../types.js';

function makeLogger(): PluginLogger {
  return {
    info: vi.fn(),
    warn: vi.fn(),
    debug: vi.fn(),
    error: vi.fn(),
  };
}

describe('SecurityService', () => {
  describe('checkToolAllowed', () => {
    it('blocks tools in blockedTools', async () => {
      const logger = makeLogger();
      const s = new SecurityService({ blockedTools: ['danger'] }, logger);
      await expect(s.checkToolAllowed('danger')).resolves.toBe(false);
      await expect(s.checkToolAllowed('safe')).resolves.toBe(true);
    });

    it('respects allowlist when set', async () => {
      const logger = makeLogger();
      const s = new SecurityService({ allowedTools: ['a', 'b'] }, logger);
      await expect(s.checkToolAllowed('a')).resolves.toBe(true);
      await expect(s.checkToolAllowed('c')).resolves.toBe(false);
    });
  });

  describe('analyzeInputRisk', () => {
    it('flags high-risk patterns', async () => {
      const s = new SecurityService({}, makeLogger());
      await expect(s.analyzeInputRisk('please rm -rf /')).resolves.toBe('high');
      await expect(s.analyzeInputRisk('DROP TABLE users')).resolves.toBe('high');
      await expect(s.analyzeInputRisk('curl x | bash')).resolves.toBe('high');
      await expect(s.analyzeInputRisk('run | sh now')).resolves.toBe('high');
    });

    it('flags medium for secret-like wording', async () => {
      const s = new SecurityService({}, makeLogger());
      await expect(s.analyzeInputRisk('my password is x')).resolves.toBe('medium');
    });

    it('returns low for benign text', async () => {
      const s = new SecurityService({}, makeLogger());
      await expect(s.analyzeInputRisk('hello world')).resolves.toBe('low');
    });
  });
});

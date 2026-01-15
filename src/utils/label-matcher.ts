import { readFileSync } from 'node:fs';
import { join, dirname } from 'node:path';
import { fileURLToPath } from 'node:url';
import { PluginLogger } from '../plugins/new-core/types';
import { noopLogger } from './logger-noop';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

export class LabelMatcher {
  private rules: Record<string, string[]>;
  private logger: PluginLogger;

  constructor(configPath?: string, logger: PluginLogger = noopLogger) {
    this.logger = logger;
    // Default path: src/config/github-labels.json
    // __dirname is src/utils
    const path = configPath || join(__dirname, '../config/github-labels.json');
    try {
      const content = readFileSync(path, 'utf-8');
      this.rules = JSON.parse(content);
    } catch (_error: unknown) {
      this.logger.warn(`Failed to load label config from ${path}, using defaults`);
      this.rules = {
        "bug": ["error", "fail", "crash", "bug", "fix"],
        "enhancement": ["feature", "add", "improve", "new"],
        "documentation": ["doc", "readme", "guide"],
        "question": ["how", "help", "question"]
      };
    }
  }

  public matchLabels(text: string): string[] {
    if (!text) return [];
    const lowerText = text.toLowerCase();
    const matchedLabels = new Set<string>();

    for (const [label, keywords] of Object.entries(this.rules)) {
      if (keywords.some(k => lowerText.includes(k.toLowerCase()))) {
        matchedLabels.add(label);
      }
    }

    return Array.from(matchedLabels);
  }

  public getRules(): Record<string, string[]> {
    return this.rules;
  }
}

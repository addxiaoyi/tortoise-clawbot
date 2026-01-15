import { readFileSync } from 'node:fs';
import { join, dirname } from 'node:path';
import { fileURLToPath } from 'node:url';
import { PromptTemplate } from './engine.js';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

const FILE_PATTERN = /^[a-zA-Z0-9_.-]+$/;

function isValidFilename(name: string, version: string): boolean {
  return FILE_PATTERN.test(name) && FILE_PATTERN.test(version);
}

export class PromptRegistry {
  private templatesDir: string;
  private cache = new Map<string, PromptTemplate>();

  constructor(templatesDir?: string) {
    this.templatesDir = templatesDir || join(__dirname, 'templates');
  }

  public load(name: string, version: string = 'latest'): PromptTemplate {
    if (!isValidFilename(name, version)) {
      throw new Error(`Invalid prompt name or version: ${name}/${version}`);
    }

    if (version === 'latest') {
      if (name === 'system') version = 'v1';
    }

    const cacheKey = `${name}:${version}`;
    if (this.cache.has(cacheKey)) {
      return this.cache.get(cacheKey)!;
    }

    const filename = `${name}.${version}.prompt`;
    const path = join(this.templatesDir, filename);
    
    try {
        const content = readFileSync(path, 'utf-8');
        const template = new PromptTemplate(content);
        this.cache.set(cacheKey, template);
        return template;
    } catch (error: unknown) {
      const message = error instanceof Error ? error.message : String(error);
      throw new Error(`Failed to load prompt template ${name} version ${version}: ${message}`);
    }
  }
}

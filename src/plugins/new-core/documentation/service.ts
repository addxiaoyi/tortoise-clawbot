import { DocumentationConfig } from './types';
import { PluginLogger } from '../types';

/** 文档存储键：仅允许安全文件名，防止路径与键注入。 */
const DOC_LOOKUP_KEY_RE = /^[a-z0-9][a-z0-9._-]{0,254}$/i;

function slugifyDocTitle(title: string): string {
  const base = title
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '')
    .slice(0, 200);
  return base || 'doc';
}

function safeDocExtension(ext: string | undefined): string {
  const raw = (ext ?? 'markdown').toLowerCase().replace(/[^a-z0-9]/g, '');
  return raw || 'markdown';
}

function assertSafeDocLookupKey(filename: string): void {
  const f = filename.trim();
  if (!f || f.includes('..') || f.includes('/') || f.includes('\\') || !DOC_LOOKUP_KEY_RE.test(f)) {
    throw new Error('Invalid document filename');
  }
}

export class DocumentationService {
  private config: DocumentationConfig;
  private logger: PluginLogger;
  private docs: Map<string, string>;

  constructor(config: DocumentationConfig, logger: PluginLogger) {
    this.config = config;
    this.logger = logger;
    this.docs = new Map();
  }

  public async start(): Promise<void> {
    this.logger.info('Documentation service started');
  }

  public async stop(): Promise<void> {
    this.logger.info('Documentation service stopped');
  }

  public async generateDoc(title: string, content: string): Promise<string> {
    const filename = `${slugifyDocTitle(title)}.${safeDocExtension(this.config.format)}`;
    this.docs.set(filename, content);
    this.logger.info(`Generated documentation: ${filename}`);
    return filename;
  }

  public async listDocs(): Promise<string[]> {
    return Array.from(this.docs.keys());
  }

  public async getDocContent(filename: string): Promise<string | undefined> {
    assertSafeDocLookupKey(filename);
    return this.docs.get(filename.trim());
  }
}

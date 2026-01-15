import http from 'http';
import fs from 'fs/promises';
import path from 'path';
import { resolveSafeRelativeFile, resolveUrlPathnameToSafePath } from '../../../utils/safe-path.js';
import { contentTypeForFile, shouldServeFileAsUtf8Text } from './mime.js';
import { PluginLogger } from '../types';

interface CanvasServiceConfig {
  root: string;
  port: number;
  host?: string;
}

export class CanvasService {
  private server?: http.Server;
  private config: CanvasServiceConfig;
  private logger: PluginLogger;

  constructor(config: CanvasServiceConfig, logger: PluginLogger) {
    this.config = config;
    this.logger = logger;
  }

  public async start(): Promise<void> {
    const { port, root } = this.config;
    const host = this.config.host ?? '127.0.0.1';
    
    // Ensure root directory exists
    try {
      await fs.access(root);
    } catch {
      await fs.mkdir(root, { recursive: true });
    }

    const server = http.createServer(async (req, res) => {
      try {
        let pathname: string;
        try {
          pathname = new URL(req.url || '/', `http://localhost:${port}`).pathname;
        } catch {
          res.statusCode = 400;
          res.end('Bad Request');
          return;
        }
        let filePath: string;
        try {
          filePath = resolveUrlPathnameToSafePath(root, pathname);
        } catch (e: unknown) {
          const message = e instanceof Error ? e.message : String(e);
          this.logger.warn(`Path traversal or invalid path: ${pathname} (${message})`);
          res.statusCode = 403;
          res.end('Forbidden');
          return;
        }

        let stat;
        try {
            stat = await fs.stat(filePath);
        } catch (e: unknown) {
            const message = e instanceof Error ? e.message : String(e);
            this.logger.debug(`File not found: ${filePath}, error: ${message}`);
        }

        if (stat?.isDirectory()) {
            filePath = path.join(filePath, 'index.html');
            try {
                stat = await fs.stat(filePath);
            } catch (e: unknown) {
                const message = e instanceof Error ? e.message : String(e);
                this.logger.debug(`index.html not found in: ${filePath}, error: ${message}`);
                stat = undefined;
            }
        }

        if (!stat) {
            res.statusCode = 404;
            res.end('Not Found');
            return;
        }

        const raw = await fs.readFile(filePath);
        const ctype = contentTypeForFile(filePath);
        res.setHeader('Content-Type', ctype);
        res.setHeader('X-Content-Type-Options', 'nosniff');
        res.setHeader('Referrer-Policy', 'no-referrer');
        res.setHeader('X-Frame-Options', 'SAMEORIGIN');
        res.statusCode = 200;
        if (shouldServeFileAsUtf8Text(filePath)) {
          res.end(raw.toString('utf-8'));
        } else {
          res.end(raw);
        }
      } catch (err: unknown) {
        const message = err instanceof Error ? err.message : String(err);
        this.logger.error(`Error serving file: ${message}`);
        res.statusCode = 500;
        res.end('Internal Server Error');
      }
    });

    this.server = server;

    return new Promise((resolve: () => void, reject: (arg0: any) => void) => {
      this.server?.listen(port, host, () => {
        this.logger.info(`Canvas server listening on ${host}:${port}`);
        resolve();
      });
      this.server?.on('error', (err: Error | undefined) => {
        this.logger.error('Canvas server error', err);
        reject(err);
      });
    });
  }

  public async stop(): Promise<void> {
    if (this.server) {
      return new Promise((resolve: () => void, reject: (arg0: any) => void) => {
        this.server?.close((err: Error | undefined) => {
          if (err) {
            this.logger.error('Error closing canvas server', err);
            reject(err);
          } else {
            this.logger.info('Canvas server stopped');
            resolve();
          }
        });
      });
    }
  }

  public async writeFile(filename: string, content: string): Promise<string> {
    const filePath = resolveSafeRelativeFile(this.config.root, filename);

    await fs.writeFile(filePath, content, 'utf-8');
    return this.getUrl(filename);
  }

  public getUrl(filename: string): string {
    resolveSafeRelativeFile(this.config.root, filename);
    const host = this.config.host ?? '127.0.0.1';
    const parts = filename
      .trim()
      .split(/[/\\]+/)
      .filter((p) => p.length > 0)
      .map((p) => encodeURIComponent(p));
    return `http://${host}:${this.config.port}/${parts.join('/')}`;
  }

  public async listFiles(): Promise<string[]> {
    try {
        const files = await fs.readdir(this.config.root);
        return files;
    } catch {
        return [];
    }
  }
}

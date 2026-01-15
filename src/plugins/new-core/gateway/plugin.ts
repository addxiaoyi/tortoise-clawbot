import { PluginContext, PluginLifecycle, PluginConfig } from '../types.js';
import {
  IncomingMessage,
  ServerResponse,
  createServer,
  Server,
} from 'node:http';
import { applyHttpServerTimeouts } from './server-timeouts.js';
import { createHermesRuntimeApi } from '../../../hermes/runtime/workspace.js';
import {
  execHermesInvokeSkill,
  execHermesListSkills,
  execHermesMemory,
} from '../../../hermes/runtime/tools.js';

export interface GatewayConfig extends PluginConfig {
  port?: number;
  host?: string;
  /** 启用 CORS 响应头 */
  cors?: boolean;
  /**
   * 允许的浏览器 Origin 列表（精确匹配 `Origin` 请求头）。
   * - 未设置或为空数组时：若 `cors` 为 true，仍使用 `Access-Control-Allow-Origin: *`（兼容旧行为）。
   * - 非空时：仅当请求带 `Origin` 且命中列表时回显该 Origin，否则不返回 ACAO（浏览器跨域将失败）。
   */
  corsAllowOrigins?: string[];
  /**
   * 单条 HTTP 请求（含 body）在服务器侧的最长耗时（毫秒），超时断开。默认 60000。
   * 设为 0 表示不覆盖（使用 Node 默认，通常较长）。
   */
  requestTimeoutMs?: number;
  /**
   * 等待客户端发完完整请求头的最长时间（毫秒）。须小于生效后的 requestTimeout。默认 30000。
   * 设为 0 表示不覆盖（使用 Node 默认）。
   */
  headersTimeoutMs?: number;
}

export class GatewayPlugin implements PluginLifecycle {
  private ctx!: PluginContext;
  private server: Server | null = null;
  private port: number = 3000;
  /** 默认仅本机回环，避免无意暴露到局域网；生产需对外时可配置 host: 0.0.0.0 */
  private host: string = '127.0.0.1';
  private corsEnabled: boolean = false;
  private corsAllowOrigins: string[] = [];
  private requestTimeoutMs: number = 60_000;
  private headersTimeoutMs: number = 30_000;

  public async onInit(ctx: PluginContext): Promise<void> {
    this.ctx = ctx;
    ctx.logger.info('GatewayPlugin initialized');
    const config = ctx.getConfig<GatewayConfig>();
    this.applyConfig(config);
  }

  public async onStart(): Promise<void> {
    this.ctx.logger.info(`GatewayPlugin starting on ${this.host}:${this.port}`);
    
    this.server = createServer((req: IncomingMessage, res: ServerResponse) => {
      this.handleRequest(req, res);
    });

    return new Promise((resolve, reject) => {
      const server = this.server;
      if (!server) return reject(new Error('Server not created'));

      applyHttpServerTimeouts(server, this.requestTimeoutMs, this.headersTimeoutMs);

      server.listen(this.port, this.host, () => {
        this.ctx.logger.info(
          `GatewayPlugin listening on port ${this.port} (requestTimeoutMs=${server.requestTimeout}, headersTimeoutMs=${server.headersTimeout})`,
        );
        resolve();
      });
      
      server.on('error', (err) => {
        this.ctx.logger.error('GatewayPlugin server error', err);
        reject(err);
      });
    });
  }

  public async onStop(): Promise<void> {
    this.ctx.logger.info('GatewayPlugin stopping');
    return new Promise((resolve, reject) => {
      if (this.server) {
        this.server.close((err) => {
          if (err) return reject(err);
          this.server = null;
          resolve();
        });
      } else {
        resolve();
      }
    });
  }

  public async onConfigChange(newConfig: GatewayConfig): Promise<void> {
    this.ctx.logger.info('GatewayPlugin config updated, restarting...');
    this.applyConfig(newConfig);
    
    await this.onStop();
    await this.onStart();
  }

  private applyConfig(config: GatewayConfig) {
    if (config.port !== undefined) this.port = config.port;
    if (config.host !== undefined) this.host = config.host;
    if (config.cors !== undefined) this.corsEnabled = config.cors;
    if (config.corsAllowOrigins !== undefined) {
      this.corsAllowOrigins = [...config.corsAllowOrigins];
    }
    if (config.requestTimeoutMs !== undefined) {
      this.requestTimeoutMs = config.requestTimeoutMs;
    }
    if (config.headersTimeoutMs !== undefined) {
      this.headersTimeoutMs = config.headersTimeoutMs;
    }
  }

  private applySecurityHeaders(res: ServerResponse) {
    res.setHeader('X-Content-Type-Options', 'nosniff');
    res.setHeader('Referrer-Policy', 'no-referrer');
    res.setHeader('X-Frame-Options', 'SAMEORIGIN');
  }

  private applyCorsHeaders(req: IncomingMessage, res: ServerResponse) {
    if (!this.corsEnabled) return;
    const allowList = this.corsAllowOrigins;
    if (allowList.length > 0) {
      const origin = req.headers.origin;
      if (typeof origin === 'string' && allowList.includes(origin)) {
        res.setHeader('Access-Control-Allow-Origin', origin);
        res.setHeader('Vary', 'Origin');
      }
    } else {
      res.setHeader('Access-Control-Allow-Origin', '*');
    }
    res.setHeader(
      'Access-Control-Allow-Methods',
      'GET, HEAD, POST, PUT, DELETE, OPTIONS',
    );
    res.setHeader(
      'Access-Control-Allow-Headers',
      'Content-Type, Authorization',
    );
    res.setHeader('Access-Control-Max-Age', '86400');
  }

  private handleRequest(req: IncomingMessage, res: ServerResponse) {
    void this.handleRequestAsync(req, res).catch((err) => {
      this.ctx.logger.error('GatewayPlugin request error', err instanceof Error ? err : undefined);
      if (!res.writableEnded) {
        this.writeJson(res, 500, {
          error: 'INTERNAL_ERROR',
          message: err instanceof Error ? err.message : String(err),
        });
      }
    });
  }

  private async handleRequestAsync(req: IncomingMessage, res: ServerResponse) {
    this.applySecurityHeaders(res);
    this.applyCorsHeaders(req, res);

    if (req.method === 'OPTIONS') {
      res.writeHead(204);
      res.end();
      return;
    }

    if (req.method === 'GET' && req.url === '/health') {
      this.writeJson(res, 200, {
        status: 'ok',
        runtime: 'hermes-agent',
        version: this.ctx.meta.version,
      });
      return;
    }

    if (req.method === 'GET' && req.url === '/tools') {
      this.writeJson(res, 200, execHermesListSkills());
      return;
    }

    if (req.method === 'POST' && req.url === '/invoke') {
      const body = await this.readJsonBody(req);
      const api = this.createRuntimeApi();
      const result = await execHermesInvokeSkill(api, {
        skill: typeof body.skill === 'string' ? body.skill : '',
        tool: typeof body.tool === 'string' ? body.tool : '',
        args:
          body.args !== null && typeof body.args === 'object' && !Array.isArray(body.args)
            ? body.args as Record<string, unknown>
            : {},
        timeoutMs: typeof body.timeoutMs === 'number' ? body.timeoutMs : undefined,
      });
      this.writeJson(res, 200, result);
      return;
    }

    if (req.method === 'POST' && req.url === '/memory') {
      const body = await this.readJsonBody(req);
      const api = this.createRuntimeApi();
      const result = await execHermesMemory(api, {
        action: body.action as 'get' | 'set' | 'list' | 'delete' | 'clear',
        key: typeof body.key === 'string' ? body.key : undefined,
        value: body.value,
      });
      this.writeJson(res, 200, result);
      return;
    }

    this.ctx.events.emit('gateway:request', { req, res });

    if (!res.writableEnded) {
      res.writeHead(404);
      res.end('Not Found');
    }
  }

  private createRuntimeApi() {
    return createHermesRuntimeApi({
      config: this.ctx.getConfig<Record<string, unknown>>(),
      logger: this.ctx.logger,
    });
  }

  private readJsonBody(req: IncomingMessage): Promise<Record<string, unknown>> {
    return new Promise((resolve, reject) => {
      let data = '';
      req.on('data', (chunk) => {
        data += chunk;
        if (Buffer.byteLength(data, 'utf8') > 1_048_576) {
          reject(new Error('request body too large'));
          req.destroy();
        }
      });
      req.on('end', () => {
        if (!data.trim()) {
          resolve({});
          return;
        }
        try {
          const parsed = JSON.parse(data);
          if (parsed !== null && typeof parsed === 'object' && !Array.isArray(parsed)) {
            resolve(parsed as Record<string, unknown>);
          } else {
            reject(new Error('JSON body must be an object'));
          }
        } catch {
          reject(new Error('invalid JSON body'));
        }
      });
      req.on('error', reject);
    });
  }

  private writeJson(res: ServerResponse, status: number, payload: unknown) {
    res.writeHead(status, { 'Content-Type': 'application/json; charset=utf-8' });
    res.end(JSON.stringify(payload));
  }
}

// Tortoise TypeScript SDK - Core Client

export interface ClientOptions {
  apiKey?: string;
  baseUrl: string;
  timeout?: number;
  headers?: Record<string, string>;
}

export interface Session {
  id: string;
  userId?: string;
  status: 'active' | 'idle' | 'archived';
  createdAt: string;
  lastActiveAt: string;
  messageCount: number;
  config: SessionConfig;
}

export interface SessionConfig {
  model: string;
  temperature: number;
  maxTokens: number;
  systemPrompt?: string;
}

export interface Message {
  id: string;
  role: 'user' | 'assistant' | 'system';
  content: string;
  createdAt: string;
  attachments?: Attachment[];
}

export interface Attachment {
  type: 'image' | 'audio' | 'video' | 'file';
  url: string;
  mimeType?: string;
  size?: number;
}

export interface ToolDefinition {
  name: string;
  description: string;
  parameters: Record<string, any>;
}

export interface ToolCall {
  id: string;
  name: string;
  arguments: Record<string, any>;
}

export interface ToolResult {
  toolName: string;
  success: boolean;
  result: any;
  executionTimeMs: number;
}

export interface StreamEvent {
  type: 'message_start' | 'content_chunk' | 'tool_call' | 'tool_result' | 'message_end' | 'error';
  data: any;
}

export interface Memory {
  id: string;
  content: string;
  type: 'fact' | 'preference' | 'experience';
  tags: string[];
  importance: number;
  createdAt: string;
}

export interface Plugin {
  id: string;
  name: string;
  version: string;
  description: string;
  enabled: boolean;
}

export class TortoiseError extends Error {
  code: string;
  statusCode: number;
  requestId?: string;

  constructor(message: string, code: string = 'UNKNOWN', statusCode: number = 500) {
    super(message);
    this.name = 'TortoiseError';
    this.code = code;
    this.statusCode = statusCode;
  }
}

export class AuthError extends TortoiseError {
  constructor(message: string) {
    super(message, 'AUTH_ERROR', 401);
    this.name = 'AuthError';
  }
}

export class RateLimitError extends TortoiseError {
  retryAfter?: number;

  constructor(message: string, retryAfter?: number) {
    super(message, 'RATE_LIMITED', 429);
    this.name = 'RateLimitError';
    this.retryAfter = retryAfter;
  }
}

export class TortoiseClient {
  private apiKey: string;
  private baseUrl: string;
  private headers: Record<string, string>;
  private connected: boolean = false;
  private eventHandlers: Map<string, Set<Function>> = new Map();

  constructor(options: ClientOptions) {
    this.apiKey = options.apiKey || '';
    this.baseUrl = options.baseUrl;
    this.headers = {
      'Content-Type': 'application/json',
      ...options.headers,
    };
    
    if (this.apiKey) {
      this.headers['Authorization'] = `Bearer ${this.apiKey}`;
      this.headers['X-API-Key'] = this.apiKey;
    }
  }

  // Connect to the server
  async connect(): Promise<void> {
    try {
      const response = await fetch(`${this.baseUrl}/health`);
      if (!response.ok) {
        throw new TortoiseError('Failed to connect to server', 'CONNECTION_FAILED', response.status);
      }
      this.connected = true;
      this.emit('connected');
    } catch (error) {
      this.connected = false;
      throw error;
    }
  }

  // Disconnect from the server
  async disconnect(): Promise<void> {
    this.connected = false;
    this.emit('disconnected');
  }

  // Check if connected
  isConnected(): boolean {
    return this.connected;
  }

  // Event handling
  on(event: string, handler: Function): void {
    if (!this.eventHandlers.has(event)) {
      this.eventHandlers.set(event, new Set());
    }
    this.eventHandlers.get(event)!.add(handler);
  }

  off(event: string, handler: Function): void {
    this.eventHandlers.get(event)?.delete(handler);
  }

  private emit(event: string, data?: any): void {
    this.eventHandlers.get(event)?.forEach(handler => handler(data));
  }

  // Session management
  sessions = {
    async create(options: { userId?: string; config?: Partial<SessionConfig> } = {}): Promise<Session> {
      const response = await this.request('POST', '/sessions', {
        userId: options.userId,
        config: options.config,
      });
      return response;
    },

    async get(sessionId: string): Promise<Session> {
      return this.request('GET', `/sessions/${sessionId}`);
    },

    async list(options: { userId?: string; status?: string; limit?: number; cursor?: string } = {}): Promise<{ sessions: Session[]; nextCursor?: string }> {
      const params = new URLSearchParams();
      if (options.userId) params.set('userId', options.userId);
      if (options.status) params.set('status', options.status);
      if (options.limit) params.set('limit', String(options.limit));
      if (options.cursor) params.set('cursor', options.cursor);
      
      const query = params.toString();
      return this.request('GET', `/sessions${query ? `?${query}` : ''}`);
    },

    async update(sessionId: string, updates: Partial<SessionConfig>): Promise<Session> {
      return this.request('PUT', `/sessions/${sessionId}`, updates);
    },

    async delete(sessionId: string): Promise<void> {
      await this.request('DELETE', `/sessions/${sessionId}`);
    },
  };

  // Message management
  messages = {
    async send(sessionId: string, options: {
      content: string;
      type?: string;
      attachments?: Attachment[];
      tools?: ToolDefinition[];
      stream?: boolean;
    }): Promise<Message | AsyncIterableStream<StreamEvent>> {
      if (options.stream) {
        return this.streamMessages(sessionId, options);
      }
      return this.request('POST', `/sessions/${sessionId}/messages`, options);
    },

    async *streamMessages(sessionId: string, options: any): AsyncIterableStream<StreamEvent> {
      const response = await fetch(`${this.baseUrl}/sessions/${sessionId}/messages`, {
        method: 'POST',
        headers: this.headers,
        body: JSON.stringify({ ...options, stream: true }),
      });

      if (!response.ok) {
        throw new TortoiseError('Failed to send message', 'SEND_FAILED', response.status);
      }

      const reader = response.body?.getReader();
      if (!reader) throw new TortoiseError('No response body');

      const decoder = new TextDecoder();
      let buffer = '';

      try {
        while (true) {
          const { done, value } = await reader.read();
          if (done) break;

          buffer += decoder.decode(value, { stream: true });
          const lines = buffer.split('\n');
          buffer = lines.pop() || '';

          for (const line of lines) {
            if (line.startsWith('event:')) {
              const eventType = line.slice(6).trim();
              yield { type: eventType as any, data: null };
            } else if (line.startsWith('data:')) {
              const data = JSON.parse(line.slice(5));
              yield { type: 'content_chunk', data };
            }
          }
        }
      } finally {
        reader.releaseLock();
      }
    },

    async list(sessionId: string, options: { before?: string; limit?: number } = {}): Promise<{ messages: Message[]; hasMore: boolean }> {
      const params = new URLSearchParams();
      if (options.before) params.set('before', options.before);
      if (options.limit) params.set('limit', String(options.limit));
      
      const query = params.toString();
      return this.request('GET', `/sessions/${sessionId}/messages${query ? `?${query}` : ''}`);
    },
  };

  // Tool management
  tools = {
    async list(): Promise<{ tools: ToolDefinition[] }> {
      return this.request('GET', '/tools');
    },

    async invoke(toolName: string, options: { sessionId?: string; arguments: Record<string, any> }): Promise<ToolResult> {
      return this.request('POST', `/tools/${toolName}/invoke`, options);
    },
  };

  // Memory management
  memory = {
    async store(options: {
      sessionId?: string;
      content: string;
      type?: string;
      tags?: string[];
      importance?: number;
    }): Promise<{ id: string; success: boolean }> {
      return this.request('POST', '/memory', options);
    },

    async search(options: { query?: string; sessionId?: string; limit?: number } = {}): Promise<{ results: Memory[] }> {
      const params = new URLSearchParams();
      if (options.query) params.set('query', options.query);
      if (options.sessionId) params.set('sessionId', options.sessionId);
      if (options.limit) params.set('limit', String(options.limit));
      
      const query = params.toString();
      return this.request('GET', `/memory/search${query ? `?${query}` : ''}`);
    },
  };

  // Plugin management
  plugins = {
    async list(): Promise<{ plugins: Plugin[] }> {
      return this.request('GET', '/plugins');
    },

    async install(options: { source: string; pluginId: string }): Promise<Plugin> {
      return this.request('POST', '/plugins/install', options);
    },

    async uninstall(pluginId: string): Promise<void> {
      await this.request('DELETE', `/plugins/${pluginId}`);
    },
  };

  // WebSocket streaming
  async createWebSocket(): Promise<TortoiseWebSocket> {
    const wsUrl = this.baseUrl.replace('http', 'ws') + '/ws';
    return new TortoiseWebSocket(wsUrl, this.apiKey);
  }

  // Private request helper
  private async request(method: string, path: string, body?: any): Promise<any> {
    const url = `${this.baseUrl}${path}`;
    
    const response = await fetch(url, {
      method,
      headers: this.headers,
      body: body ? JSON.stringify(body) : undefined,
    });

    if (!response.ok) {
      if (response.status === 401) {
        throw new AuthError('Authentication failed');
      }
      if (response.status === 429) {
        const retryAfter = response.headers.get('Retry-After');
        throw new RateLimitError('Rate limited', retryAfter ? parseInt(retryAfter) : undefined);
      }

      const error = await response.json().catch(() => ({}));
      throw new TortoiseError(
        error.message || 'Request failed',
        error.code || 'REQUEST_FAILED',
        response.status
      );
    }

    return response.json();
  }
}

// WebSocket client
export class TortoiseWebSocket {
  private ws: WebSocket | null = null;
  private url: string;
  private apiKey: string;
  private eventHandlers: Map<string, Set<Function>> = new Map();

  constructor(url: string, apiKey: string) {
    this.url = url;
    this.apiKey = apiKey;
  }

  connect(): Promise<void> {
    return new Promise((resolve, reject) => {
      this.ws = new WebSocket(this.url);
      
      this.ws.onopen = () => {
        // Send handshake
        this.ws?.send(JSON.stringify({
          type: 'handshake',
          apiKey: this.apiKey,
        }));
        this.emit('connected');
        resolve();
      };

      this.ws.onerror = (error) => {
        this.emit('error', error);
        reject(error);
      };

      this.ws.onmessage = (event) => {
        const data = JSON.parse(event.data);
        this.emit(data.type, data);
      };

      this.ws.onclose = () => {
        this.emit('disconnected');
      };
    });
  }

  disconnect(): void {
    this.ws?.close();
    this.ws = null;
  }

  send(message: any): void {
    this.ws?.send(JSON.stringify(message));
  }

  on(event: string, handler: Function): void {
    if (!this.eventHandlers.has(event)) {
      this.eventHandlers.set(event, new Set());
    }
    this.eventHandlers.get(event)!.add(handler);
  }

  off(event: string, handler: Function): void {
    this.eventHandlers.get(event)?.delete(handler);
  }

  private emit(event: string, data?: any): void {
    this.eventHandlers.get(event)?.forEach(handler => handler(data));
    this.eventHandlers.get('*')?.forEach(handler => handler({ type: event, data }));
  }
}

// Async iterable stream type
export interface AsyncIterableStream<T> {
  [Symbol.asyncIterator](): AsyncIterator<T>;
}

export default TortoiseClient;

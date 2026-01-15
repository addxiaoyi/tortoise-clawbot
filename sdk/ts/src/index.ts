/**
 * Tortoise SDK - TypeScript/JavaScript SDK
 * 
 * 高性能 AI Agent TypeScript SDK
 */

export class TortoiseClient {
  private baseUrl: string;
  private apiKey?: string;
  private sessionId?: string;

  constructor(baseUrl: string = 'http://localhost:18792', apiKey?: string) {
    this.baseUrl = baseUrl;
    this.apiKey = apiKey;
  }

  private async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<T> {
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
      ...(options.headers as Record<string, string>),
    };

    if (this.apiKey) {
      headers['Authorization'] = `Bearer ${this.apiKey}`;
    }

    const response = await fetch(`${this.baseUrl}${endpoint}`, {
      ...options,
      headers,
    });

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    return response.json();
  }

  // Session Management
  async createSession(userId: string, metadata?: Record<string, string>) {
    return this.request<Session>('/sessions', {
      method: 'POST',
      body: JSON.stringify({ user_id: userId, metadata }),
    });
  }

  async getSession(sessionId: string) {
    return this.request<Session>(`/sessions/${sessionId}`);
  }

  async deleteSession(sessionId: string) {
    return this.request<void>(`/sessions/${sessionId}`, { method: 'DELETE' });
  }

  async listSessions(userId?: string) {
    return this.request<SessionsResponse>(
      `/sessions${userId ? `?user_id=${userId}` : ''}`
    );
  }

  // Messages
  async sendMessage(
    sessionId: string,
    content: string,
    options?: {
      type?: MessageType;
      format?: ContentFormat;
      stream?: boolean;
    }
  ) {
    const session = await this.getSession(sessionId);
    this.sessionId = session.id;

    return this.request<SendMessageResponse>('/messages', {
      method: 'POST',
      body: JSON.stringify({
        session_id: sessionId,
        content,
        type: options?.type || 'text',
        format: options?.format || 'plain',
        stream: options?.stream || false,
      }),
    });
  }

  async *sendMessageStream(
    sessionId: string,
    content: string
  ): AsyncGenerator<MessageChunk> {
    const response = await fetch(`${this.baseUrl}/messages/stream`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        ...(this.apiKey ? { Authorization: `Bearer ${this.apiKey}` } : {}),
      },
      body: JSON.stringify({ session_id: sessionId, content, stream: true }),
    });

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    const reader = response.body?.getReader();
    if (!reader) throw new Error('No response body');

    const decoder = new TextDecoder();
    let buffer = '';

    while (true) {
      const { done, value } = await reader.read();
      if (done) break;

      buffer += decoder.decode(value, { stream: true });
      const lines = buffer.split('\n');
      buffer = lines.pop() || '';

      for (const line of lines) {
        if (line.trim()) {
          yield JSON.parse(line);
        }
      }
    }
  }

  async getMessages(sessionId: string, limit = 50, offset = 0) {
    return this.request<MessagesResponse>(
      `/sessions/${sessionId}/messages?limit=${limit}&offset=${offset}`
    );
  }

  // Tools
  async listTools() {
    return this.request<ToolDefinition[]>('/tools');
  }

  async executeTool(
    pluginId: string,
    toolName: string,
    arguments: Record<string, unknown>
  ) {
    return this.request<ToolExecutionResult>('/tools/execute', {
      method: 'POST',
      body: JSON.stringify({
        plugin_id: pluginId,
        tool_name: toolName,
        arguments: JSON.stringify(arguments),
      }),
    });
  }

  // Memory
  async saveMemory(
    type: MemoryType,
    content: string,
    importance = 0.5,
    metadata?: Record<string, string>
  ) {
    return this.request<{ memory_id: string }>('/memory', {
      method: 'POST',
      body: JSON.stringify({
        type,
        content,
        importance,
        metadata,
      }),
    });
  }

  async queryMemory(
    query: string,
    type?: MemoryType,
    limit = 10,
    similarityThreshold = 0.7
  ) {
    return this.request<MemoryQueryResult>('/memory/query', {
      method: 'POST',
      body: JSON.stringify({
        query,
        type,
        limit,
        similarity_threshold: similarityThreshold,
      }),
    });
  }

  async deleteMemory(memoryId: string) {
    return this.request<void>(`/memory/${memoryId}`, { method: 'DELETE' });
  }

  // Plugins
  async listPlugins() {
    return this.request<Plugin[]>('/plugins');
  }

  async installPlugin(source: string, config?: Record<string, string>) {
    return this.request<Plugin>('/plugins/install', {
      method: 'POST',
      body: JSON.stringify({ source, config }),
    });
  }

  async uninstallPlugin(pluginId: string, force = false) {
    return this.request<void>(`/plugins/${pluginId}?force=${force}`, {
      method: 'DELETE',
    });
  }

  // Channels
  async listChannels() {
    return this.request<Channel[]>('/channels');
  }

  async connectChannel(
    type: ChannelType,
    credentials: Record<string, string>,
    config?: Record<string, string>
  ) {
    return this.request<Channel>('/channels/connect', {
      method: 'POST',
      body: JSON.stringify({ type, credentials, config }),
    });
  }

  async disconnectChannel(channelId: string) {
    return this.request<void>(`/channels/${channelId}`, { method: 'DELETE' });
  }

  // Configuration
  async getConfig() {
    return this.request<Config>('/config');
  }

  async updateConfig(config: Partial<Config>) {
    return this.request<void>('/config', {
      method: 'PUT',
      body: JSON.stringify(config),
    });
  }

  // Health
  async healthCheck() {
    return this.request<HealthCheckResponse>('/health');
  }

  // Metrics
  async getMetrics() {
    return this.request<MetricsResponse>('/metrics');
  }

  // Event Subscription
  subscribe(
    events: EventType[],
    sessionId?: string
  ): WebSocket {
    const wsUrl = this.baseUrl.replace('http', 'ws');
    const params = new URLSearchParams({
      events: events.join(','),
      ...(sessionId ? { session_id: sessionId } : {}),
    });
    const ws = new WebSocket(`${wsUrl}/subscribe?${params}`);

    ws.onmessage = (event) => {
      const data = JSON.parse(event.data);
      // Handle events
    };

    return ws;
  }
}

// Type Definitions
export interface Session {
  id: string;
  user_id: string;
  state: SessionState;
  created_at: number;
  updated_at: number;
  metadata?: Record<string, string>;
}

export enum SessionState {
  Active = 'active',
  Paused = 'paused',
  Closed = 'closed',
}

export enum MessageType {
  Text = 'text',
  Image = 'image',
  Audio = 'audio',
  Video = 'video',
  File = 'file',
  Location = 'location',
  Contact = 'contact',
}

export enum ContentFormat {
  Plain = 'plain',
  Markdown = 'markdown',
  Html = 'html',
  Json = 'json',
}

export interface Message {
  id: string;
  session_id: string;
  role: string;
  content: string;
  format: ContentFormat;
  type: MessageType;
  timestamp: number;
  metadata?: Record<string, string>;
}

export interface SendMessageResponse {
  message_id: string;
  session_id: string;
  streaming: boolean;
}

export interface MessageChunk {
  message_id: string;
  content: string;
  is_final: boolean;
  delta: string;
}

export interface MessagesResponse {
  messages: Message[];
  total: number;
}

export interface ToolDefinition {
  name: string;
  description: string;
  parameters: ToolParameter[];
  require_confirmation: boolean;
  category: string;
}

export interface ToolParameter {
  name: string;
  type: string;
  description: string;
  required: boolean;
  default?: string;
}

export interface ToolExecutionResult {
  success: boolean;
  result: string;
  error?: string;
  duration_ms: number;
}

export enum MemoryType {
  Working = 'working',
  Semantic = 'semantic',
  Episodic = 'episodic',
}

export interface Memory {
  id: string;
  type: MemoryType;
  content: string;
  importance: number;
  created_at: number;
  accessed_at: number;
  metadata?: Record<string, string>;
}

export interface MemoryQueryResult {
  memories: Memory[];
  scores: number[];
}

export interface Plugin {
  id: string;
  name: string;
  version: string;
  description: string;
  state: PluginState;
  tools: ToolDefinition[];
  installed_at: number;
}

export enum PluginState {
  Installed = 'installed',
  Loaded = 'loaded',
  Running = 'running',
  Disabled = 'disabled',
  Error = 'error',
}

export enum ChannelType {
  Telegram = 'telegram',
  Discord = 'discord',
  Slack = 'slack',
  WhatsApp = 'whatsapp',
  Web = 'web',
  WeChat = 'wechat',
  LINE = 'line',
  Signal = 'signal',
  SMS = 'sms',
}

export interface Channel {
  id: string;
  type: ChannelType;
  name: string;
  state: ChannelState;
  created_at: number;
}

export enum ChannelState {
  Disconnected = 'disconnected',
  Connecting = 'connecting',
  Connected = 'connected',
  Error = 'error',
}

export interface Config {
  version: string;
  gateway: GatewayConfig;
  models: ModelConfig[];
  channels: ChannelConfig[];
}

export interface GatewayConfig {
  bind_address: string;
  port: number;
  tls_enabled: boolean;
  max_connections: number;
}

export interface ModelConfig {
  provider: string;
  model: string;
  api_key?: string;
  base_url?: string;
  temperature: number;
  max_tokens: number;
}

export interface ChannelConfig {
  id: string;
  type: ChannelType;
  config: Record<string, string>;
}

export interface HealthCheckResponse {
  healthy: boolean;
  version: string;
  uptime_seconds: number;
  components: HealthCheckComponent[];
}

export interface HealthCheckComponent {
  name: string;
  healthy: boolean;
  message: string;
}

export interface MetricsResponse {
  timestamp: number;
  system: SystemMetrics;
  sessions: SessionMetrics[];
  channels: ChannelMetrics[];
}

export interface SystemMetrics {
  cpu_percent: number;
  memory_used_bytes: number;
  memory_total_bytes: number;
  active_connections: number;
  goroutines: number;
}

export interface SessionMetrics {
  session_id: string;
  message_count: number;
  total_tokens: number;
  last_activity: number;
}

export interface ChannelMetrics {
  channel_id: string;
  type: ChannelType;
  message_count: number;
  last_message: number;
}

export enum EventType {
  Message = 'message',
  Session = 'session',
  ToolCall = 'tool_call',
  Error = 'error',
  Metrics = 'metrics',
  Channel = 'channel',
}

export interface Event {
  type: EventType;
  id: string;
  timestamp: number;
  payload: unknown;
}

// Default export
export default TortoiseClient;

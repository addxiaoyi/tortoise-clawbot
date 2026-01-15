/**
 * Core types for Agent Runtime
 * Supports both Hermes Core and OpenClaw Compatible modes
 */

// ============================================
// Plugin System Types
// ============================================

export interface PluginMetadata {
  id: string;
  name: string;
  version: string;
  description?: string;
  author?: string;
  license?: string;
}

export interface PluginConfig {
  [key: string]: unknown;
}

export interface PluginLogger {
  debug(message: string, ...args: unknown[]): void;
  info(message: string, ...args: unknown[]): void;
  warn(message: string, ...args: unknown[]): void;
  error(message: string, error?: Error, ...args: unknown[]): void;
}

export interface PluginStorage {
  getItem<T>(key: string): Promise<T | null>;
  setItem<T>(key: string, value: T): Promise<void>;
  removeItem(key: string): Promise<void>;
  clear(): Promise<void>;
}

export interface PluginEventBus {
  emit(event: string, payload: unknown): void;
  on(event: string, handler: (payload: unknown) => void): () => void;
  off(event: string, handler: (payload: unknown) => void): void;
}

export interface PluginContext {
  readonly meta: PluginMetadata;
  readonly logger: PluginLogger;
  readonly storage: PluginStorage;
  readonly events: PluginEventBus;
  readonly getConfig: <T extends PluginConfig>() => T;
}

export interface PluginLifecycle {
  onInit(ctx: PluginContext): Promise<void> | void;
  onStart?(): Promise<void> | void;
  onStop?(): Promise<void> | void;
  onConfigChange?(newConfig: PluginConfig): Promise<void> | void;
}

export type PluginFactory = () => PluginLifecycle;

// ============================================
// Skill System Types
// ============================================

export interface SkillToolParameter {
  name: string;
  type: string;
  description: string;
  required?: boolean;
  default?: unknown;
}

export interface SkillTool {
  name: string;
  description: string;
  parameters: SkillToolParameter[];
  execute(args: Record<string, unknown>, context: PluginContext): Promise<unknown>;
}

export interface SkillDefinition {
  name: string;
  version: string;
  description: string;
  prerequisites?: string;
  tools: SkillTool[];
}

export interface SkillPlugin extends PluginLifecycle {
  getSkillDefinition(): SkillDefinition;
}

// ============================================
// Channel System Types (OpenClaw Compatible)
// ============================================

export type ChannelCapability =
  | 'text'
  | 'markdown'
  | 'html'
  | 'images'
  | 'audio'
  | 'video'
  | 'files'
  | 'typing'
  | 'read-receipts'
  | 'reactions'
  | 'threads'
  | 'reply';

export interface ChannelMessage {
  id: string;
  channel: string;
  from: string;
  content: string;
  timestamp: number;
  metadata?: Record<string, unknown>;
}

export interface OutboundMessage {
  to: string;
  content: string;
  channel: string;
  options?: {
    replyTo?: string;
    threadId?: string;
    parseMode?: 'markdown' | 'html' | 'plain';
    typing?: boolean;
  };
}

export interface ChannelAdapter extends PluginLifecycle {
  readonly name: string;
  readonly capabilities: ChannelCapability[];
  send(message: OutboundMessage): Promise<void>;
  formatForChannel(content: string): Promise<string>;
}

// ============================================
// Provider System Types (OpenClaw Compatible)
// ============================================

export interface CompletionOptions {
  model: string;
  messages: Array<{
    role: 'system' | 'user' | 'assistant' | 'developer';
    content: string;
  }>;
  temperature?: number;
  maxTokens?: number;
  topP?: number;
  stop?: string[];
  stream?: boolean;
  tools?: Array<{
    name: string;
    description?: string;
    inputSchema: Record<string, unknown>;
  }>;
  toolChoice?: 'auto' | 'none' | { type: 'function'; function: { name: string } };
}

export interface CompletionResult {
  content: string;
  usage?: {
    promptTokens: number;
    completionTokens: number;
    totalTokens: number;
  };
  model: string;
  stopReason?: string;
  toolCalls?: Array<{
    name: string;
    arguments: string;
  }>;
}

export interface StreamChunk {
  type: 'content' | 'tool-call' | 'done' | 'error';
  content?: string;
  toolCall?: {
    name: string;
    arguments: string;
  };
  usage?: CompletionResult['usage'];
  stopReason?: string;
}

export interface EmbedOptions {
  model: string;
  input: string | string[];
}

export interface EmbedResult {
  embeddings: number[][];
  model: string;
}

export interface ModelProvider extends PluginLifecycle {
  readonly name: string;
  readonly defaultModel: string;
  readonly supportedModels: string[];
  complete(options: CompletionOptions): Promise<CompletionResult>;
  completeStream(options: CompletionOptions): AsyncGenerator<StreamChunk>;
  embed(options: EmbedOptions): Promise<EmbedResult>;
}

// ============================================
// Memory System Types
// ============================================

export interface MemoryConfig {
  prefix: string;
  maxValueBytes: number;
  sessionScoped: boolean;
  requireSessionKeyForRead: boolean;
  requireSessionKeyForWrite: boolean;
  persistence?: {
    type: 'file' | 'redis' | 'sqlite';
    path?: string;
    url?: string;
  };
}

export interface MemoryEntry {
  key: string;
  value: unknown;
  timestamp: number;
  sessionKey?: string;
}

export interface MemoryAction {
  action: 'get' | 'set' | 'list' | 'delete' | 'clear';
  key?: string;
  value?: unknown;
}

// ============================================
// Session System Types
// ============================================

export interface Session {
  id: string;
  key: string;
  createdAt: number;
  lastActiveAt: number;
  metadata?: Record<string, unknown>;
}

export interface SessionContext {
  session: Session;
  memory: MemoryAction;
  user?: {
    id: string;
    name?: string;
  };
}

// ============================================
// Gateway Types
// ============================================

export interface GatewayConfig {
  port: number;
  host: string;
  auth?: {
    token?: string;
    sessionSecret?: string;
  };
  cors?: {
    enabled: boolean;
    origins?: string[];
  };
  rateLimit?: {
    windowMs: number;
    maxRequests: number;
  };
  requestTimeoutMs?: number;
  headersTimeoutMs?: number;
}

export interface InvokeOptions {
  signal?: AbortSignal;
  timeoutMs?: number;
  sessionKey?: string;
}

export interface InvokeRequest {
  skill: string;
  tool: string;
  args?: Record<string, unknown>;
  timeoutMs?: number;
}

export interface InvokeResponse {
  skill: string;
  tool: string;
  result: unknown;
  durationMs: number;
}

// ============================================
// Runtime API (Hermes Compatible)
// ============================================

export interface RuntimeApi {
  resolvePath(input: string): string;
  logger: PluginLogger;
  config?: Record<string, unknown>;
  registerTool?(tool: unknown, opts?: { optional?: boolean }): void;
}

// ============================================
// MCP Types
// ============================================

export const MCP_TOOL_NAMES = [
  'tohelp_ping',
  'tohelp_list_skills',
  'tohelp_resolve_workspace_path',
  'tohelp_invoke_skill',
  'tohelp_memory',
  'tohelp_gateway_health_probe',
  'tohelp_session_create',
  'tohelp_session_get',
  'tohelp_channel_send',
] as const;

export type McpToolName = (typeof MCP_TOOL_NAMES)[number];

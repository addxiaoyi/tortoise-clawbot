/**
 * TypeScript Types
 */

export type AgentState = 'init' | 'created' | 'running' | 'paused' | 'stopped' | 'error';

export type MemoryType = 'episodic' | 'semantic' | 'procedural';

export type TaskPriority = 'low' | 'normal' | 'high' | 'critical';

export type NodeStatus = 'online' | 'offline' | 'busy';

export interface Task {
  id: string;
  description: string;
  priority: TaskPriority;
  requirements?: TaskRequirement[];
}

export interface TaskRequirement {
  capability: string;
  minLevel: number;
}

export interface MeshMessage {
  id: string;
  from: string;
  to: string;
  type: 'request' | 'response' | 'event' | 'delegate' | 'collaborate';
  payload: unknown;
  timestamp: string;
}

export interface PluginManifest {
  id: string;
  name: string;
  version: string;
  type: 'channel' | 'tool' | 'skill' | 'memory' | 'llm' | 'custom';
  capabilities: {
    tools?: ToolDefinition[];
    resources?: ResourceDefinition[];
    channels?: ChannelDefinition[];
  };
  permissions: Permission[];
}

export interface ToolDefinition {
  name: string;
  description: string;
  inputSchema: Record<string, unknown>;
}

export interface ResourceDefinition {
  uriPattern: string;
  mimeTypes: string[];
}

export interface ChannelDefinition {
  name: string;
  configSchema: Record<string, unknown>;
}

export interface Permission {
  name: string;
  description?: string;
  required: boolean;
}

/**
 * Agent Module
 * 
 * Agent management and interaction
 */

import { AxiosInstance } from 'axios';
import { AgentState } from './types';

export interface AgentConfig {
  name: string;
  modelProvider: string;
  model: string;
  skills?: string[];
}

export interface AgentInfo {
  id: string;
  name?: string;
  state: AgentState | string;
  model?: string;
}

export class Agent {
  constructor(private http: AxiosInstance) {}
  
  /**
   * Create a new agent
   */
  async create(config: AgentConfig): Promise<string> {
    const response = await this.http.post<{ id: string }>('/api/v1/agents', {
      name: config.name,
      model_provider: config.modelProvider,
      model: config.model,
      skills: config.skills || [],
    });
    return response.data.id;
  }
  
  /**
   * List all agents
   */
  async list(): Promise<AgentInfo[]> {
    const response = await this.http.get<{ agents: AgentInfo[] }>('/api/v1/agents');
    return response.data.agents;
  }
  
  /**
   * Get agent details
   */
  async get(id: string): Promise<AgentInfo> {
    const response = await this.http.get<AgentInfo>(`/api/v1/agents/${id}`);
    return response.data;
  }
  
  /**
   * Start an agent
   */
  async start(id: string): Promise<void> {
    await this.http.post(`/api/v1/agents/${id}/start`);
  }
  
  /**
   * Stop an agent
   */
  async stop(id: string): Promise<void> {
    await this.http.post(`/api/v1/agents/${id}/stop`);
  }
  
  /**
   * Delete an agent
   */
  async delete(id: string): Promise<void> {
    await this.http.delete(`/api/v1/agents/${id}`);
  }
  
  /**
   * Send a message to an agent
   */
  async sendMessage(id: string, message: string): Promise<string> {
    const response = await this.http.post<{ response: string }>(`/api/v1/agents/${id}/message`, {
      message,
    });
    return response.data.response;
  }
}

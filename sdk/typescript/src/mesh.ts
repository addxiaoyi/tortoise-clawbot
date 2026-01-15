/**
 * Mesh Module
 * 
 * Multi-agent mesh networking
 */

import { AxiosInstance } from 'axios';

export interface MeshNode {
  id: string;
  name: string;
  address: string;
  capabilities: string[];
  status: 'online' | 'offline' | 'busy';
}

export interface DelegateRequest {
  nodeId: string;
  task: string;
  priority?: 'low' | 'normal' | 'high' | 'critical';
}

export class Mesh {
  constructor(private http: AxiosInstance) {}
  
  /**
   * List all connected nodes
   */
  async listNodes(): Promise<MeshNode[]> {
    const response = await this.http.get<{ nodes: MeshNode[] }>('/api/v1/mesh/nodes');
    return response.data.nodes;
  }
  
  /**
   * Connect to a new node
   */
  async connect(address: string): Promise<void> {
    await this.http.post('/api/v1/mesh/connect', { address });
  }
  
  /**
   * Disconnect from a node
   */
  async disconnect(nodeId: string): Promise<void> {
    await this.http.post('/api/v1/mesh/disconnect', { node_id: nodeId });
  }
  
  /**
   * Delegate a task to another node
   */
  async delegate(request: DelegateRequest): Promise<void> {
    await this.http.post('/api/v1/mesh/delegate', {
      node_id: request.nodeId,
      task: request.task,
      priority: request.priority || 'normal',
    });
  }
  
  /**
   * Collaborate on a task with multiple nodes
   */
  async collaborate(nodeIds: string[], task: string): Promise<void> {
    await this.http.post('/api/v1/mesh/collaborate', {
      node_ids: nodeIds,
      task,
    });
  }
  
  /**
   * Get node status
   */
  async getNodeStatus(nodeId: string): Promise<{ status: string; cpuUsage: number; memoryUsage: number }> {
    const response = await this.http.get(`/api/v1/mesh/nodes/${nodeId}/status`);
    return response.data;
  }
}

/**
 * Gateway Cluster System
 * Distributed deployment with load balancing and failover
 */

import { randomUUID } from 'node:crypto';

// ============================================
// Cluster Node Types
// ============================================

export interface ClusterNode {
  id: string;
  host: string;
  port: number;
  region?: string;
  roles: NodeRole[];
  status: NodeStatus;
  health: NodeHealth;
  capabilities: string[];
  load: number;
  connections: number;
  version: string;
  startedAt: number;
  lastHeartbeat: number;
}

export type NodeRole = 'primary' | 'replica' | 'worker' | 'api' | 'gateway';

export type NodeStatus = 'online' | 'offline' | 'draining' | 'maintenance';

export interface NodeHealth {
  score: number;           // 0-100
  checks: HealthCheck[];
  lastCheck: number;
}

export interface HealthCheck {
  name: string;
  status: 'pass' | 'fail' | 'warn';
  latencyMs?: number;
  message?: string;
}

// ============================================
// Cluster Configuration
// ============================================

export interface ClusterConfig {
  nodeId: string;
  host: string;
  port: number;
  region?: string;
  roles: NodeRole[];
  
  // Discovery
  discoveryMode: 'static' | 'dns' | 'consul' | 'k8s';
  seedNodes?: string[];
  discoveryUrl?: string;
  
  // Clustering
  heartbeatInterval: number;    // ms
  heartbeatTimeout: number;    // ms
  reconnectDelay: number;       // ms
  
  // Load Balancing
  loadBalanceStrategy: 'round-robin' | 'least-connections' | 'weighted' | 'latency';
  
  // Replication
  replicationFactor: number;
  consensusTimeout: number;    // ms
  
  // TLS
  tls?: {
    enabled: boolean;
    certPath?: string;
    keyPath?: string;
    caPath?: string;
  };
}

// ============================================
// Cluster Manager
// ============================================

export class ClusterManager {
  private config: ClusterConfig;
  private nodes: Map<string, ClusterNode> = new Map();
  private localNode?: ClusterNode;
  private isLeader = false;
  private leaderId?: string;
  private eventHandlers: Map<string, Set<(data: unknown) => void>> = new Map();
  
  // WebSocket connections to other nodes
  private connections: Map<string, WebSocket> = new Map();
  
  private heartbeatTimer?: NodeJS.Timeout;
  private healthCheckTimer?: NodeJS.Timeout;
  private reconnectTimers: Map<string, NodeJS.Timeout> = new Map();

  constructor(config: ClusterConfig) {
    this.config = config;
    this.localNode = this.createLocalNode();
  }

  // ==============================
  // Lifecycle
  // ==============================

  async start(): Promise<void> {
    // Join cluster
    await this.joinCluster();
    
    // Start heartbeat
    this.startHeartbeat();
    
    // Start health checks
    this.startHealthChecks();
    
    // Sync state with peers
    await this.syncState();
    
    this.emit('cluster:start', { nodeId: this.config.nodeId });
  }

  async stop(): Promise<void> {
    // Stop timers
    this.heartbeatTimer && clearInterval(this.heartbeatTimer);
    this.healthCheckTimer && clearInterval(this.healthCheckTimer);
    
    // Close connections
    for (const [nodeId, conn] of this.connections) {
      conn.close(1000, 'Node shutting down');
    }
    
    // Leave cluster gracefully
    await this.leaveCluster();
    
    this.emit('cluster:stop', { nodeId: this.config.nodeId });
  }

  // ==============================
  // Node Management
  // ==============================

  private createLocalNode(): ClusterNode {
    return {
      id: this.config.nodeId,
      host: this.config.host,
      port: this.config.port,
      region: this.config.region,
      roles: this.config.roles,
      status: 'online',
      health: {
        score: 100,
        checks: [],
        lastCheck: Date.now(),
      },
      capabilities: ['gateway', 'api', 'websocket', 'channels'],
      load: 0,
      connections: 0,
      version: '1.0.0',
      startedAt: Date.now(),
      lastHeartbeat: Date.now(),
    };
  }

  getLocalNode(): ClusterNode | undefined {
    return this.localNode;
  }

  getNode(nodeId: string): ClusterNode | undefined {
    return this.nodes.get(nodeId);
  }

  getAllNodes(): ClusterNode[] {
    return Array.from(this.nodes.values());
  }

  getOnlineNodes(): ClusterNode[] {
    return this.getAllNodes().filter(n => n.status === 'online');
  }

  getNodesByRole(role: NodeRole): ClusterNode[] {
    return this.getOnlineNodes().filter(n => n.roles.includes(role));
  }

  // ==============================
  // Leader Election (Raft-like)
  // ==============================

  async requestElection(): Promise<boolean> {
    if (this.isLeader) return true;

    const term = Date.now();
    const votes: string[] = [this.config.nodeId];
    
    // Request votes from other nodes
    for (const [nodeId, conn] of this.connections) {
      try {
        const voted = await this.sendMessage<{ vote: boolean }>(
          nodeId,
          { type: 'vote-request', term, candidateId: this.config.nodeId }
        );
        
        if (voted?.vote) {
          votes.push(nodeId);
        }
      } catch {
        // Node unreachable, skip
      }
    }

    // Check if we have majority
    const majority = Math.floor(this.nodes.size / 2) + 1;
    const elected = votes.length >= majority;

    if (elected) {
      this.isLeader = true;
      this.leaderId = this.config.nodeId;
      this.emit('leader:elected', { nodeId: this.config.nodeId, term });
    }

    return elected;
  }

  isLeaderNode(): boolean {
    return this.isLeader;
  }

  getLeader(): ClusterNode | undefined {
    if (this.isLeader) return this.localNode;
    if (!this.leaderId) return undefined;
    return this.nodes.get(this.leaderId);
  }

  // ==============================
  // Load Balancing
  // ==============================

  selectNode(capability?: string): ClusterNode | undefined {
    const nodes = this.getOnlineNodes();
    
    if (nodes.length === 0) return undefined;
    
    if (nodes.length === 1) return nodes[0];

    switch (this.config.loadBalanceStrategy) {
      case 'round-robin':
        return this.roundRobinSelect(nodes);
        
      case 'least-connections':
        return this.leastConnectionsSelect(nodes);
        
      case 'weighted':
        return this.weightedSelect(nodes, capability);
        
      case 'latency':
        return this.latencySelect(nodes);
        
      default:
        return nodes[0];
    }
  }

  private roundRobinSelect(nodes: ClusterNode[]): ClusterNode {
    // Simple round-robin based on node ID hash
    const index = Date.now() % nodes.length;
    return nodes[index];
  }

  private leastConnectionsSelect(nodes: ClusterNode[]): ClusterNode {
    return nodes.reduce((min, node) => 
      node.connections < min.connections ? node : min
    );
  }

  private weightedSelect(nodes: ClusterNode[], capability?: string): ClusterNode {
    // Weight by inverse of load and capability match
    const scored = nodes.map(node => {
      let weight = 100 - node.load;
      
      if (capability && node.capabilities.includes(capability)) {
        weight *= 2;
      }
      
      return { node, weight };
    });

    const total = scored.reduce((sum, s) => sum + s.weight, 0);
    let random = Math.random() * total;
    
    for (const { node, weight } of scored) {
      random -= weight;
      if (random <= 0) return node;
    }

    return nodes[0];
  }

  private latencySelect(nodes: ClusterNode[]): ClusterNode {
    // Select node with best health score
    return nodes.reduce((best, node) =>
      node.health.score > best.health.score ? node : best
    );
  }

  // ==============================
  // Replication
  // ==============================

  async replicate(data: unknown, key: string): Promise<boolean> {
    if (!this.isLeader) {
      throw new Error('Can only replicate from leader');
    }

    const replicas = this.getNodesByRole('replica');
    const required = Math.min(this.config.replicationFactor, replicas.length + 1);
    
    const responses: Promise<boolean>[] = [];

    // Send to all replicas in parallel
    for (const replica of replicas) {
      responses.push(
        this.sendMessage<{ success: boolean }>(
          replica.id,
          { type: 'replicate', key, data }
        ).then(r => r?.success ?? false).catch(() => false)
      );
    }

    const results = await Promise.all(responses);
    const successCount = results.filter(Boolean).length;

    return successCount + 1 >= required; // +1 for leader
  }

  // ==============================
  // Service Discovery
  // ==============================

  private async joinCluster(): Promise<void> {
    switch (this.config.discoveryMode) {
      case 'static':
        await this.discoverStatic();
        break;
      case 'dns':
        await this.discoverDns();
        break;
      case 'consul':
        await this.discoverConsul();
        break;
      case 'k8s':
        await this.discoverK8s();
        break;
    }
  }

  private async discoverStatic(): Promise<void> {
    if (!this.config.seedNodes) return;

    for (const seed of this.config.seedNodes) {
      try {
        await this.connectToNode(seed);
      } catch {
        // Seed node unavailable
      }
    }
  }

  private async discoverDns(): Promise<void> {
    if (!this.config.discoveryUrl) return;

    // Resolve DNS SRV records
    // Implementation depends on DNS provider
  }

  private async discoverConsul(): Promise<void> {
    if (!this.config.discoveryUrl) return;

    // Query Consul for Tortoise nodes
    // curl ${CONSUL_URL}/v1/catalog/service/tortoise
  }

  private async discoverK8s(): Promise<void> {
    // Use Kubernetes API for service discovery
    // Requires service account permissions
  }

  private async connectToNode(address: string): Promise<void> {
    const [host, port] = address.split(':');
    const url = `${this.config.tls?.enabled ? 'wss' : 'ws'}://${host}:${port}/cluster`;
    
    const ws = new WebSocket(url);
    
    ws.on('open', () => {
      this.connections.set(address, ws);
      this.sendHandshake(ws, address);
    });

    ws.on('message', (data) => this.handleMessage(data, address));
    ws.on('close', () => this.handleDisconnect(address));
    ws.on('error', () => this.handleDisconnect(address));
  }

  private sendHandshake(ws: WebSocket, nodeId: string): void {
    ws.send(JSON.stringify({
      type: 'handshake',
      node: this.localNode,
    }));
  }

  private async leaveCluster(): Promise<void> {
    // Notify other nodes we're leaving
    for (const [nodeId, conn] of this.connections) {
      try {
        conn.send(JSON.stringify({ type: 'leave', nodeId: this.config.nodeId }));
      } catch {
        // Ignore
      }
    }
    
    this.connections.clear();
    this.nodes.clear();
  }

  // ==============================
  // Heartbeat & Health
  // ==============================

  private startHeartbeat(): void {
    this.heartbeatTimer = setInterval(() => {
      this.sendHeartbeat();
    }, this.config.heartbeatInterval);
  }

  private sendHeartbeat(): void {
    if (!this.localNode) return;

    const heartbeat = {
      type: 'heartbeat',
      nodeId: this.config.nodeId,
      status: this.localNode.status,
      load: this.localNode.load,
      connections: this.localNode.connections,
      timestamp: Date.now(),
    };

    for (const [nodeId, conn] of this.connections) {
      try {
        conn.send(JSON.stringify(heartbeat));
      } catch {
        // Connection dead, will be cleaned up
      }
    }

    // Update local heartbeat time
    this.localNode.lastHeartbeat = Date.now();
  }

  private startHealthChecks(): void {
    this.healthCheckTimer = setInterval(() => {
      this.performHealthCheck();
    }, 5000);
  }

  private performHealthCheck(): void {
    if (!this.localNode) return;

    const checks: HealthCheck[] = [];
    let totalScore = 100;

    // Check memory usage
    const memUsage = process.memoryUsage();
    const memPercent = (memUsage.heapUsed / memUsage.heapTotal) * 100;
    
    if (memPercent > 90) {
      checks.push({ name: 'memory', status: 'fail', message: `Memory at ${memPercent.toFixed(1)}%` });
      totalScore -= 50;
    } else if (memPercent > 75) {
      checks.push({ name: 'memory', status: 'warn', message: `Memory at ${memPercent.toFixed(1)}%` });
      totalScore -= 20;
    } else {
      checks.push({ name: 'memory', status: 'pass' });
    }

    // Check connections health
    const deadConnections = Array.from(this.connections.entries())
      .filter(([_, conn]) => conn.readyState !== WebSocket.OPEN);

    if (deadConnections.length > 0) {
      checks.push({ 
        name: 'connections', 
        status: 'warn', 
        message: `${deadConnections.length} dead connections` 
      });
      totalScore -= 10 * deadConnections.length;
    }

    this.localNode.health = {
      score: Math.max(0, totalScore),
      checks,
      lastCheck: Date.now(),
    };

    // Update load metric
    this.localNode.load = Math.min(100, memPercent);
  }

  // ==============================
  // Message Handling
  // ==============================

  private async handleMessage(data: unknown, fromNode: string): Promise<void> {
    const msg = JSON.parse(data as string);

    switch (msg.type) {
      case 'handshake':
        this.handleHandshake(msg.node);
        break;
        
      case 'heartbeat':
        this.handleHeartbeat(msg);
        break;
        
      case 'leave':
        this.handleLeave(msg.nodeId);
        break;
        
      case 'vote-request':
        this.handleVoteRequest(msg);
        break;
        
      case 'replicate':
        // Handle replication
        break;
        
      case 'state-sync':
        this.handleStateSync(msg);
        break;
    }
  }

  private handleHandshake(node: ClusterNode): void {
    this.nodes.set(node.id, node);
    this.emit('node:joined', node);
  }

  private handleHeartbeat(msg: { nodeId: string; status: NodeStatus; load: number }): void {
    const node = this.nodes.get(msg.nodeId);
    if (node) {
      node.status = msg.status;
      node.load = msg.load;
      node.lastHeartbeat = Date.now();
    }
  }

  private handleLeave(nodeId: string): void {
    this.nodes.delete(nodeId);
    const conn = this.connections.get(nodeId);
    conn?.close();
    this.connections.delete(nodeId);
    this.emit('node:left', { nodeId });
  }

  private handleVoteRequest(msg: { term: number; candidateId: string }): void {
    // Simple vote: vote yes if candidate has higher score
    const candidate = this.nodes.get(msg.candidateId);
    const localScore = this.localNode?.health.score ?? 0;
    
    if (candidate && candidate.health.score > localScore) {
      this.sendMessage(this.config.nodeId, { type: 'vote', vote: true });
    }
  }

  private handleStateSync(msg: { state: Map<string, unknown> }): void {
    // Sync cluster state from leader
    this.emit('state:sync', msg.state);
  }

  private handleDisconnect(nodeId: string): void {
    const timer = this.reconnectTimers.get(nodeId);
    if (timer) {
      clearTimeout(timer);
    }

    this.connections.delete(nodeId);
    const node = this.nodes.get(nodeId);
    
    if (node) {
      node.status = 'offline';
      this.emit('node:disconnected', node);
    }

    // Schedule reconnect
    this.reconnectTimers.set(
      nodeId,
      setTimeout(() => this.reconnect(nodeId), this.config.reconnectDelay)
    );
  }

  private async reconnect(nodeId: string): Promise<void> {
    const node = this.nodes.get(nodeId);
    if (!node) return;

    try {
      await this.connectToNode(`${node.host}:${node.port}`);
    } catch {
      // Will retry on next heartbeat
    }
  }

  // ==============================
  // Communication
  // ==============================

  private async sendMessage<T>(nodeId: string, message: unknown): Promise<T | undefined> {
    const conn = this.connections.get(nodeId);
    if (!conn || conn.readyState !== WebSocket.OPEN) {
      throw new Error(`Node ${nodeId} not connected`);
    }

    return new Promise((resolve, reject) => {
      const timeout = setTimeout(() => {
        reject(new Error('Request timeout'));
      }, this.config.consensusTimeout);

      const handler = (data: unknown) => {
        clearTimeout(timeout);
        conn.removeEventListener('message', handler);
        resolve(JSON.parse(data as string) as T);
      };

      conn.addEventListener('message', handler);
      conn.send(JSON.stringify(message));
    });
  }

  // ==============================
  // State Sync
  // ==============================

  private async syncState(): Promise<void> {
    if (this.isLeader) {
      // Send state to new nodes
      const state = this.getClusterState();
      for (const [nodeId, conn] of this.connections) {
        conn.send(JSON.stringify({ type: 'state-sync', state }));
      }
    } else {
      // Request state from leader
      const leader = this.getLeader();
      if (leader) {
        await this.sendMessage(leader.id, { type: 'state-request' });
      }
    }
  }

  getClusterState(): Map<string, unknown> {
    return new Map([
      ['nodes', Array.from(this.nodes.values())],
      ['leader', this.leaderId],
      ['term', Date.now()],
    ]);
  }

  getClusterStats(): {
    totalNodes: number;
    onlineNodes: number;
    leaderNode: string | null;
    averageLoad: number;
    totalConnections: number;
  } {
    const nodes = this.getAllNodes();
    const online = nodes.filter(n => n.status === 'online');
    
    return {
      totalNodes: nodes.length,
      onlineNodes: online.length,
      leaderNode: this.leaderId || null,
      averageLoad: online.length > 0 
        ? online.reduce((sum, n) => sum + n.load, 0) / online.length 
        : 0,
      totalConnections: online.reduce((sum, n) => sum + n.connections, 0),
    };
  }

  // ==============================
  // Event Emitter
  // ==============================

  on(event: string, handler: (data: unknown) => void): () => void {
    if (!this.eventHandlers.has(event)) {
      this.eventHandlers.set(event, new Set());
    }
    this.eventHandlers.get(event)!.add(handler);
    
    return () => this.eventHandlers.get(event)?.delete(handler);
  }

  private emit(event: string, data: unknown): void {
    this.eventHandlers.get(event)?.forEach(handler => {
      try {
        handler(data);
      } catch {
        // Handler error, continue
      }
    });
  }
}

// ============================================
// Factory
// ============================================

export function createCluster(config: Partial<ClusterConfig> & { nodeId: string; host: string; port: number }): ClusterManager {
  return new ClusterManager({
    heartbeatInterval: 5000,
    heartbeatTimeout: 15000,
    reconnectDelay: 3000,
    loadBalanceStrategy: 'least-connections',
    replicationFactor: 2,
    consensusTimeout: 5000,
    discoveryMode: 'static',
    roles: ['gateway', 'api'],
    ...config,
  });
}

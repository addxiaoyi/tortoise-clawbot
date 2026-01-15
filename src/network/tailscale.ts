/**
 * Tailscale Network Integration
 * Zero-trust networking for device discovery and secure connections
 */

export interface TailscaleConfig {
  apiKey: string;
  tailnet: string;           // e.g., "example.com" or "*"
  filterSelf?: boolean;       // Exclude self from device list
}

export interface TailscaleDevice {
  id: string;
  name: string;              // hostname.tailnet
  hostname: string;
  ip: string;                // Tailscale IP (100.x.x.x)
  ipv6?: string;
  mac?: string;
  manufacturer?: string;
  os: string;
  online: boolean;
  lastSeen?: number;          // Unix timestamp
  lastSeenAgo?: string;       // Human-readable
  keyExpiry?: number;
  authorized: boolean;
  isEphemeral: boolean;
  isBlocked: boolean;
  advertiseRoutes?: string[];
  advertisedRoutes?: string[];
  tags?: string[];
  tailnetLockError?: string;
  updatedAt?: number;
}

export interface TailscaleACLPolicy {
  acls: ACLRule[];
  autoApprovers?: AutoApprover[];
  derpMap?: DERPMap;
  groups?: Record<string, string[]>;
  noise?: Record<string, unknown>;
  ssh?: SSHSettings;
  tagOwners?: Record<string, string[]>;
  tests?: ACLTest[];
}

export interface ACLRule {
  action: 'accept' | 'drop';
  src: string[];             // IPs, tags, groups, or wildcard "*"
  dst: string[];             // Same format
  proto?: string;            // tcp, udp, icmp, or "*"
  ports?: string | string[]; // e.g., "80,443" or ["80", "443"]
}

export interface DERPMap {
  regions: Record<string, DERPRegion>;
}

export interface DERPRegion {
  regionID: number;
  regionCode: string;
  regionName: string;
  nodes: DERPNode[];
}

export interface DERPNode {
  name: string;
  regionID: number;
  hostName: string;
  ipv4?: string;
  ipv6?: string;
  port?: number;
  stunPort?: number;
  stunOnly?: boolean;
  relay?: string;
}

export interface SSHSettings {
  allow?: SSHAllowRule[];
  deny?: SSHDenyRule[];
}

export interface SSHAllowRule {
  principals: SSHPrincipal[];
  sshUsers: Record<string, string>; // Map from AD provider to SSH user format
}

export interface SSHPrincipal {
  any?: boolean;
  user?: string;
  group?: string;
  tag?: string;
}

export interface AutoApprover {
  exitNode?: string[];
  route?: string[];
  subnet?: AutoApproverSubnet[];
  device?: string[];
}

export interface AutoApproverSubnet {
  action: 'accept' | 'suggest';
  src?: string[];
  dst?: string[];
}

export interface ACLTest {
  // ACL test configuration
}

export class TailscaleClient {
  private apiKey: string;
  private baseUrl = 'https://api.tailscale.com/api/v2';
  
  constructor(config: TailscaleConfig) {
    this.apiKey = config.apiKey;
  }

  /**
   * Make authenticated API request
   */
  private async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<T> {
    const response = await fetch(`${this.baseUrl}${endpoint}`, {
      ...options,
      headers: {
        'Authorization': `Bearer ${this.apiKey}`,
        'Content-Type': 'application/json',
        ...options.headers,
      },
    });

    if (!response.ok) {
      const error = await response.text();
      throw new Error(`Tailscale API error: ${response.status} ${error}`);
    }

    if (response.status === 204) {
      return undefined as T;
    }

    return response.json() as T;
  }

  /**
   * List all devices in the tailnet
   */
  async listDevices(filter?: {
    online?: boolean;
    name?: string;
    tag?: string;
  }): Promise<TailscaleDevice[]> {
    const params = new URLSearchParams();
    if (filter?.online !== undefined) params.set('online', String(filter.online));
    if (filter?.name) params.set('name', filter.name);
    if (filter?.tag) params.set('tag', filter.tag);

    const queryString = params.toString();
    const endpoint = `/devices${queryString ? '?' + queryString : ''}`;
    
    return this.request<TailscaleDevice[]>(endpoint);
  }

  /**
   * Get a specific device by hostname or ID
   */
  async getDevice(device: string): Promise<TailscaleDevice> {
    return this.request<TailscaleDevice>(`/device/${encodeURIComponent(device)}`);
  }

  /**
   * Delete/revoke a device key
   */
  async deleteDevice(device: string): Promise<void> {
    await this.request(`/device/${encodeURIComponent(device)}`, {
      method: 'DELETE',
    });
  }

  /**
   * Authorize a device
   */
  async authorizeDevice(device: string): Promise<void> {
    await this.request(`/device/${encodeURIComponent(device)}/authorized`, {
      method: 'POST',
      body: JSON.stringify({ authorized: true }),
    });
  }

  /**
   * Get network routes advertised by a device
   */
  async getDeviceRoutes(device: string): Promise<{
    advertised: string[];
    enabled: string[];
    advertising: boolean;
  }> {
    return this.request(`/device/${encodeURIComponent(device)}/routes`);
  }

  /**
   * Set enabled routes for a device
   */
  async setDeviceRoutes(device: string, routes: string[]): Promise<void> {
    await this.request(`/device/${encodeURIComponent(device)}/routes`, {
      method: 'POST',
      body: JSON.stringify({ enabledRoutes: routes }),
    });
  }

  /**
   * Get current ACL policy
   */
  async getACLPolicy(): Promise<TailscaleACLPolicy> {
    return this.request<TailscaleACLPolicy>('/acl');
  }

  /**
   * Replace ACL policy
   */
  async setACLPolicy(policy: TailscaleACLPolicy): Promise<{ resetDate: string }> {
    return this.request<{ resetDate: string }>('/acl', {
      method: 'POST',
      body: JSON.stringify(policy),
    });
  }

  /**
   * Validate ACL policy without applying
   */
  async validateACLPolicy(policy: TailscaleACLPolicy): Promise<{
    valid: boolean;
    errors?: string[];
  }> {
    return this.request('/acl/validate', {
      method: 'POST',
      body: JSON.stringify(policy),
    });
  }

  /**
   * Get DNS configuration
   */
  async getDNS(): Promise<{
    nameservers: string[];
    dnsSuffix: string;
    magicDNSSuffix: string;
    requireExitNode?: boolean;
    searchDomains: string[];
  }> {
    return this.request('/dns/nameservers');
  }

  /**
   * Set DNS configuration
   */
  async setDNS(config: {
    nameservers: string[];
    searchDomains?: string[];
    allowGroup?: string[];
  }): Promise<void> {
    await this.request('/dns/nameservers', {
      method: 'POST',
      body: JSON.stringify(config),
    });
  }

  /**
   * Get filter lists configuration
   */
  async getFilterLists(): Promise<{
    asn: string[];
    country: string[];
  }> {
    return this.request('/filterlists');
  }

  /**
   * Create logout key
   */
  async createLogoutKey(): Promise<{ key: string; expiry: number }> {
    return this.request('/device/<mydevice>/keys', {
      method: 'POST',
    });
  }

  /**
   * Get tailnet lock status
   */
  async getTailnetLockStatus(): Promise<{
    enabled: boolean;
    nodes: Array<{
      nodeKey: string;
      locked: boolean;
    }>;
    publicKey?: string;
  }> {
    return this.request('/tailnet-lock');
  }
}

// ============================================
// Device Discovery Service
// ============================================

export interface DiscoveredService {
  deviceId: string;
  deviceName: string;
  serviceType: 'gateway' | 'api' | 'mcp' | 'storage' | 'other';
  host: string;
  port: number;
  protocol: 'http' | 'https' | 'ws' | 'wss' | 'tcp';
  metadata?: Record<string, string>;
  discoveredAt: number;
}

export class DeviceDiscoveryService {
  private tailscale: TailscaleClient;
  private services: Map<string, DiscoveredService> = new Map();
  private onServiceDiscovered?: (service: DiscoveredService) => void;
  private onServiceLost?: (serviceId: string) => void;

  constructor(tailscale: TailscaleClient) {
    this.tailscale = tailscale;
  }

  /**
   * Set callbacks for service events
   */
  onDiscovery(callback: (service: DiscoveredService) => void): void {
    this.onServiceDiscovered = callback;
  }

  onLost(callback: (serviceId: string) => void): void {
    this.onServiceLost = callback;
  }

  /**
   * Discover all Tortoise/Gateway services on the tailnet
   */
  async discoverServices(): Promise<DiscoveredService[]> {
    const devices = await this.tailscale.listDevices({ online: true });
    const services: DiscoveredService[] = [];

    for (const device of devices) {
      // Try common service ports
      const commonPorts = [
        { port: 8080, service: 'gateway' as const, protocol: 'http' as const },
        { port: 8443, service: 'gateway' as const, protocol: 'https' as const },
        { port: 18789, service: 'gateway' as const, protocol: 'http' as const },
        { port: 3000, service: 'api' as const, protocol: 'http' as const },
        { port: 3001, service: 'api' as const, protocol: 'https' as const },
        { port: 8081, service: 'mcp' as const, protocol: 'http' as const },
        { port: 5432, service: 'storage' as const, protocol: 'tcp' as const },
        { port: 6379, service: 'storage' as const, protocol: 'tcp' as const },
      ];

      for (const { port, service, protocol } of commonPorts) {
        const serviceId = `${device.id}:${port}`;
        
        // Check if service is reachable
        const isReachable = await this.checkService(device.ip, port);
        
        if (isReachable) {
          const discovered: DiscoveredService = {
            deviceId: device.id,
            deviceName: device.name,
            serviceType: service,
            host: device.ip,
            port,
            protocol,
            metadata: {
              deviceHostname: device.hostname,
              deviceOS: device.os,
              deviceOnline: String(device.online),
            },
            discoveredAt: Date.now(),
          };

          services.push(discovered);
          this.services.set(serviceId, discovered);
          this.onServiceDiscovered?.(discovered);
        } else if (this.services.has(serviceId)) {
          // Service was previously discovered but is now unreachable
          this.services.delete(serviceId);
          this.onServiceLost?.(serviceId);
        }
      }
    }

    return services;
  }

  /**
   * Check if a service is reachable at the given IP and port
   */
  private async checkService(ip: string, port: number): Promise<boolean> {
    try {
      const controller = new AbortController();
      const timeoutId = setTimeout(() => controller.abort(), 2000);

      const response = await fetch(`http://${ip}:${port}/health`, {
        method: 'GET',
        signal: controller.signal,
      });

      clearTimeout(timeoutId);
      return response.ok;
    } catch {
      return false;
    }
  }

  /**
   * Get all discovered services
   */
  getServices(): DiscoveredService[] {
    return Array.from(this.services.values());
  }

  /**
   * Get services by type
   */
  getServicesByType(type: DiscoveredService['serviceType']): DiscoveredService[] {
    return this.getServices().filter(s => s.serviceType === type);
  }

  /**
   * Get primary gateway service
   */
  getPrimaryGateway(): DiscoveredService | undefined {
    const gateways = this.getServicesByType('gateway');
    return gateways.sort((a, b) => {
      // Prefer https
      if (a.protocol === 'https' && b.protocol !== 'https') return -1;
      if (b.protocol === 'https' && a.protocol !== 'https') return 1;
      return 0;
    })[0];
  }
}

// ============================================
// Tailscale Plugin for Tortoise
// ============================================

import type { PluginContext } from '../plugins/new-core/types.js';

export type { PluginContext } from '../plugins/new-core/types.js';

export interface TailscalePluginConfig {
  apiKey: string;
  tailnet: string;
  autoDiscover?: boolean;
  discoverInterval?: number; // ms
  whitelistTags?: string[]; // Only connect to devices with these tags
}

export class TailscalePlugin implements PluginLifecycle {
  private config?: TailscalePluginConfig;
  private client?: TailscaleClient;
  private discovery?: DeviceDiscoveryService;
  private discoveryTimer?: NodeJS.Timeout;

  async onInit(ctx: PluginContext): Promise<void> {
    this.config = ctx.getConfig<TailscalePluginConfig>();
    
    if (!this.config?.apiKey) {
      ctx.logger.warn('[tailscale] No API key configured, plugin disabled');
      return;
    }

    this.client = new TailscaleClient({
      apiKey: this.config.apiKey,
      tailnet: this.config.tailnet || '*',
    });

    ctx.logger.info('[tailscale] Tailscale plugin initialized');
  }

  async onStart(): Promise<void> {
    if (!this.client) {
      throw new Error('[tailscale] Not initialized');
    }

    // Verify API key
    try {
      await this.client.listDevices();
      this.ctx?.logger.info('[tailscale] Connected to Tailscale API');
    } catch (err) {
      throw new Error(`[tailscale] Failed to connect: ${err}`);
    }

    // Setup device discovery
    if (this.config?.autoDiscover) {
      this.startDiscovery();
    }
  }

  async onStop(): Promise<void> {
    if (this.discoveryTimer) {
      clearInterval(this.discoveryTimer);
    }
    this.ctx?.logger.info('[tailscale] Plugin stopped');
  }

  private ctx?: PluginContext;

  /**
   * Start automatic device/service discovery
   */
  private startDiscovery(): void {
    if (!this.client) return;

    this.discovery = new DeviceDiscoveryService(this.client);
    
    this.discovery.onDiscovery((service) => {
      this.ctx?.logger.info(
        `[tailscale] Discovered ${service.serviceType} at ${service.host}:${service.port}`
      );
      this.ctx?.events.emit('tailscale:service:discovered', service);
    });

    this.discovery.onLost((serviceId) => {
      this.ctx?.logger.debug(`[tailscale] Service lost: ${serviceId}`);
      this.ctx?.events.emit('tailscale:service:lost', { serviceId });
    });

    // Initial discovery
    this.discovery.discoverServices();

    // Periodic discovery
    const interval = this.config?.discoverInterval || 60000;
    this.discoveryTimer = setInterval(() => {
      this.discovery?.discoverServices();
    }, interval);
  }

  /**
   * Get list of online devices
   */
  async getDevices(): Promise<TailscaleDevice[]> {
    return this.client?.listDevices({ online: true }) || [];
  }

  /**
   * Get discovered services
   */
  getServices(): DiscoveredService[] {
    return this.discovery?.getServices() || [];
  }

  /**
   * Get primary gateway
   */
  getPrimaryGateway(): DiscoveredService | undefined {
    return this.discovery?.getPrimaryGateway();
  }
}

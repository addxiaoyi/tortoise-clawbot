/**
 * Security System - RBAC + 多租户 + 审计
 * 超越 OpenClaw/Hermes 的安全能力
 */

import crypto from 'node:crypto';

// ==================== 类型定义 ====================

// 权限
export type Permission =
  | 'admin:*'
  | 'runtime:read' | 'runtime:write' | 'runtime:execute'
  | 'memory:read' | 'memory:write' | 'memory:delete' | 'memory:admin'
  | 'skill:invoke' | 'skill:register' | 'skill:admin'
  | 'channel:read' | 'channel:write' | 'channel:admin'
  | 'provider:read' | 'provider:write' | 'provider:admin'
  | 'user:read' | 'user:write' | 'user:admin'
  | 'config:read' | 'config:write'
  | 'audit:read' | 'audit:admin';

// 角色
export type Role = 'owner' | 'admin' | 'operator' | 'developer' | 'user' | 'guest';

// 权限定义
export interface RolePermissions {
  role: Role;
  permissions: Permission[];
  extends?: Role[];
}

// 租户
export interface Tenant {
  id: string;
  name: string;
  plan: 'free' | 'pro' | 'enterprise';
  quota: TenantQuota;
  metadata?: Record<string, unknown>;
  createdAt: number;
  updatedAt: number;
}

export interface TenantQuota {
  maxUsers: number;
  maxApiCalls: number;
  maxMemoryBytes: number;
  maxSessions: number;
  maxPlugins: number;
  rateLimitPerMinute: number;
}

// 用户
export interface User {
  id: string;
  tenantId: string;
  username: string;
  email: string;
  passwordHash?: string;
  roles: Role[];
  permissions: Permission[];
  metadata?: Record<string, unknown>;
  createdAt: number;
  updatedAt: number;
  lastActiveAt?: number;
}

// API Key
export interface ApiKey {
  id: string;
  tenantId: string;
  userId?: string;
  name: string;
  keyHash: string;
  prefix: string;
  permissions: Permission[];
  expiresAt?: number;
  lastUsedAt?: number;
  createdAt: number;
}

// 审计日志
export interface AuditLog {
  id: string;
  tenantId: string;
  userId?: string;
  apiKeyId?: string;
  action: string;
  resource: string;
  resourceId?: string;
  permission: Permission;
  result: 'success' | 'denied' | 'error';
  ip?: string;
  userAgent?: string;
  metadata?: Record<string, unknown>;
  timestamp: number;
}

// ==================== 角色权限配置 ====================

const ROLE_PERMISSIONS: Record<Role, RolePermissions> = {
  owner: {
    role: 'owner',
    permissions: ['admin:*'],
  },
  admin: {
    role: 'admin',
    permissions: [
      'admin:*',
      'runtime:read', 'runtime:write',
      'memory:admin',
      'skill:admin',
      'channel:admin',
      'provider:admin',
      'user:read', 'user:write', 'user:admin',
      'config:read', 'config:write',
      'audit:read', 'audit:admin',
    ],
  },
  operator: {
    role: 'operator',
    permissions: [
      'runtime:read', 'runtime:write', 'runtime:execute',
      'memory:read', 'memory:write',
      'skill:invoke', 'skill:register',
      'channel:read', 'channel:write',
      'provider:read',
      'audit:read',
    ],
  },
  developer: {
    role: 'developer',
    permissions: [
      'runtime:read', 'runtime:execute',
      'memory:read', 'memory:write',
      'skill:invoke',
      'channel:read',
      'provider:read',
    ],
  },
  user: {
    role: 'user',
    permissions: [
      'runtime:read', 'runtime:execute',
      'memory:read', 'memory:write',
      'skill:invoke',
      'channel:read', 'channel:write',
    ],
  },
  guest: {
    role: 'guest',
    permissions: [
      'runtime:read',
      'memory:read',
      'channel:read',
    ],
  },
};

// ==================== RBAC 核心类 ====================

export class RBAC {
  private users = new Map<string, User>();
  private apiKeys = new Map<string, ApiKey>();
  private auditLogs: AuditLog[] = [];
  private readonly maxAuditLogs = 10000;

  constructor(private config?: { maxAuditLogs?: number }) {
    if (config?.maxAuditLogs) {
      this.maxAuditLogs = config.maxAuditLogs;
    }
  }

  // ==================== 用户管理 ====================

  async createUser(data: {
    tenantId: string;
    username: string;
    email: string;
    password?: string;
    roles?: Role[];
    metadata?: Record<string, unknown>;
  }): Promise<User> {
    const user: User = {
      id: crypto.randomUUID(),
      tenantId: data.tenantId,
      username: data.username,
      email: data.email,
      passwordHash: data.password ? await this.hashPassword(data.password) : undefined,
      roles: data.roles || ['user'],
      permissions: this.resolvePermissions(data.roles || ['user']),
      metadata: data.metadata,
      createdAt: Date.now(),
      updatedAt: Date.now(),
    };

    this.users.set(user.id, user);
    this.log(user.tenantId, user.id, 'user:create', 'user', user.id, 'memory:write', 'success');
    
    return user;
  }

  async getUser(id: string): Promise<User | undefined> {
    return this.users.get(id);
  }

  async getUserByEmail(email: string): Promise<User | undefined> {
    for (const user of this.users.values()) {
      if (user.email === email) return user;
    }
    return undefined;
  }

  async updateUser(id: string, data: Partial<User>): Promise<User | undefined> {
    const user = this.users.get(id);
    if (!user) return undefined;

    const updated = {
      ...user,
      ...data,
      id: user.id, // 防止修改 ID
      tenantId: user.tenantId, // 防止修改租户
      updatedAt: Date.now(),
    };

    if (data.roles) {
      updated.roles = data.roles;
      updated.permissions = this.resolvePermissions(data.roles);
    }

    this.users.set(id, updated);
    return updated;
  }

  async deleteUser(id: string): Promise<boolean> {
    const user = this.users.get(id);
    if (!user) return false;

    this.users.delete(id);
    this.log(user.tenantId, user.id, 'user:delete', 'user', id, 'user:admin', 'success');
    return true;
  }

  async listUsers(tenantId: string): Promise<User[]> {
    return Array.from(this.users.values()).filter(u => u.tenantId === tenantId);
  }

  // ==================== API Key 管理 ====================

  async createApiKey(data: {
    tenantId: string;
    userId?: string;
    name: string;
    permissions: Permission[];
    expiresAt?: number;
  }): Promise<{ key: string; apiKey: ApiKey }> {
    const key = `${data.name.toLowerCase().replace(/\s+/g, '_')}_${crypto.randomBytes(24).toString('base64url')}`;
    const keyHash = await this.hashPassword(key);
    const prefix = key.slice(0, 12);

    const apiKey: ApiKey = {
      id: crypto.randomUUID(),
      tenantId: data.tenantId,
      userId: data.userId,
      name: data.name,
      keyHash,
      prefix,
      permissions: data.permissions,
      expiresAt: data.expiresAt,
      createdAt: Date.now(),
    };

    this.apiKeys.set(keyHash, apiKey);
    this.log(data.tenantId, data.userId, 'apikey:create', 'apikey', apiKey.id, 'user:admin', 'success');

    return { key, apiKey };
  }

  async validateApiKey(key: string): Promise<{ valid: boolean; apiKey?: ApiKey; user?: User }> {
    // 查找匹配的 key
    for (const apiKey of this.apiKeys.values()) {
      if (await this.comparePassword(key, apiKey.keyHash)) {
        // 检查过期
        if (apiKey.expiresAt && apiKey.expiresAt < Date.now()) {
          return { valid: false };
        }

        // 更新最后使用时间
        apiKey.lastUsedAt = Date.now();

        // 获取关联用户
        const user = apiKey.userId ? this.users.get(apiKey.userId) : undefined;

        return { valid: true, apiKey, user };
      }
    }

    return { valid: false };
  }

  async revokeApiKey(id: string): Promise<boolean> {
    for (const [hash, key] of this.apiKeys) {
      if (key.id === id) {
        this.apiKeys.delete(hash);
        this.log(key.tenantId, key.userId, 'apikey:revoke', 'apikey', id, 'user:admin', 'success');
        return true;
      }
    }
    return false;
  }

  // ==================== 权限检查 ====================

  async checkPermission(
    userId: string | undefined,
    apiKeyId: string | undefined,
    permission: Permission,
    resourceId?: string
  ): Promise<boolean> {
    let permissions: Permission[] = [];

    if (apiKeyId) {
      const key = this.findApiKeyById(apiKeyId);
      if (key) permissions = key.permissions;
    } else if (userId) {
      const user = this.users.get(userId);
      if (user) permissions = user.permissions;
    }

    return this.hasPermission(permissions, permission);
  }

  hasPermission(userPermissions: Permission[], required: Permission): boolean {
    // 管理员权限
    if (userPermissions.includes('admin:*')) return true;

    // 精确匹配
    if (userPermissions.includes(required)) return true;

    // 通配符检查
    const [resource, action] = required.split(':');
    if (userPermissions.includes(`${resource}:*` as Permission)) return true;
    if (userPermissions.includes(`*:${action}` as Permission)) return true;

    return false;
  }

  // ==================== 审计日志 ====================

  log(
    tenantId: string,
    userId: string | undefined,
    action: string,
    resource: string,
    resourceId: string | undefined,
    permission: Permission,
    result: 'success' | 'denied' | 'error',
    metadata?: Record<string, unknown>
  ): void {
    const entry: AuditLog = {
      id: crypto.randomUUID(),
      tenantId,
      userId,
      action,
      resource,
      resourceId,
      permission,
      result,
      metadata,
      timestamp: Date.now(),
    };

    this.auditLogs.push(entry);

    // 限制日志数量
    if (this.auditLogs.length > this.maxAuditLogs) {
      this.auditLogs = this.auditLogs.slice(-this.maxAuditLogs);
    }
  }

  async getAuditLogs(options: {
    tenantId: string;
    userId?: string;
    resource?: string;
    since?: number;
    until?: number;
    limit?: number;
  }): Promise<AuditLog[]> {
    let logs = this.auditLogs.filter(l => l.tenantId === options.tenantId);

    if (options.userId) {
      logs = logs.filter(l => l.userId === options.userId);
    }
    if (options.resource) {
      logs = logs.filter(l => l.resource === options.resource);
    }
    if (options.since) {
      logs = logs.filter(l => l.timestamp >= options.since!);
    }
    if (options.until) {
      logs = logs.filter(l => l.timestamp <= options.until!);
    }

    logs.sort((a, b) => b.timestamp - a.timestamp);

    if (options.limit) {
      return logs.slice(0, options.limit);
    }

    return logs;
  }

  // ==================== 辅助方法 ====================

  private resolvePermissions(roles: Role[]): Permission[] {
    const permissions = new Set<Permission>();

    const resolve = (role: Role) => {
      const config = ROLE_PERMISSIONS[role];
      if (!config) return;

      for (const perm of config.permissions) {
        permissions.add(perm);
      }

      if (config.extends) {
        for (const ext of config.extends) {
          resolve(ext);
        }
      }
    };

    for (const role of roles) {
      resolve(role);
    }

    return Array.from(permissions);
  }

  private async hashPassword(password: string): Promise<string> {
    const salt = crypto.randomBytes(16).toString('hex');
    const hash = crypto.pbkdf2Sync(password, salt, 100000, 64, 'sha512').toString('hex');
    return `${salt}:${hash}`;
  }

  private async comparePassword(password: string, stored: string): Promise<boolean> {
    const [salt, hash] = stored.split(':');
    const inputHash = crypto.pbkdf2Sync(password, salt, 100000, 64, 'sha512').toString('hex');
    return crypto.timingSafeEqual(Buffer.from(hash), Buffer.from(inputHash));
  }

  private findApiKeyById(id: string): ApiKey | undefined {
    for (const key of this.apiKeys.values()) {
      if (key.id === id) return key;
    }
    return undefined;
  }
}

// ==================== 多租户管理器 ====================

export class TenantManager {
  private tenants = new Map<string, Tenant>();
  private quotas = new Map<string, { used: number; resetAt: number }>();

  constructor(private defaultQuota: TenantQuota) {}

  async createTenant(data: {
    name: string;
    plan?: 'free' | 'pro' | 'enterprise';
    metadata?: Record<string, unknown>;
  }): Promise<Tenant> {
    const quota = this.getQuotaForPlan(data.plan || 'free');

    const tenant: Tenant = {
      id: crypto.randomUUID(),
      name: data.name,
      plan: data.plan || 'free',
      quota,
      metadata: data.metadata,
      createdAt: Date.now(),
      updatedAt: Date.now(),
    };

    this.tenants.set(tenant.id, tenant);
    this.quotas.set(tenant.id, { used: 0, resetAt: Date.now() + 24 * 60 * 60 * 1000 });

    return tenant;
  }

  async getTenant(id: string): Promise<Tenant | undefined> {
    return this.tenants.get(id);
  }

  async updateTenant(id: string, data: Partial<Tenant>): Promise<Tenant | undefined> {
    const tenant = this.tenants.get(id);
    if (!tenant) return undefined;

    const updated = {
      ...tenant,
      ...data,
      id: tenant.id,
      updatedAt: Date.now(),
    };

    this.tenants.set(id, updated);
    return updated;
  }

  async deleteTenant(id: string): Promise<boolean> {
    const deleted = this.tenants.delete(id);
    if (deleted) {
      this.quotas.delete(id);
    }
    return deleted;
  }

  async checkQuota(tenantId: string, cost: number = 1): Promise<boolean> {
    const quota = this.quotas.get(tenantId);
    const tenant = this.tenants.get(tenantId);
    if (!quota || !tenant) return false;

    // 重置配额
    if (Date.now() > quota.resetAt) {
      quota.used = 0;
      quota.resetAt = Date.now() + 24 * 60 * 60 * 1000;
    }

    return quota.used + cost <= tenant.quota.maxApiCalls;
  }

  async useQuota(tenantId: string, cost: number = 1): Promise<boolean> {
    const canUse = await this.checkQuota(tenantId, cost);
    if (canUse) {
      const quota = this.quotas.get(tenantId);
      if (quota) {
        quota.used += cost;
      }
    }
    return canUse;
  }

  getQuotaUsage(tenantId: string): { used: number; limit: number; percent: number } {
    const quota = this.quotas.get(tenantId);
    const tenant = this.tenants.get(tenantId);

    if (!quota || !tenant) {
      return { used: 0, limit: 0, percent: 0 };
    }

    const percent = tenant.quota.maxApiCalls > 0
      ? (quota.used / tenant.quota.maxApiCalls) * 100
      : 0;

    return { used: quota.used, limit: tenant.quota.maxApiCalls, percent };
  }

  private getQuotaForPlan(plan: 'free' | 'pro' | 'enterprise'): TenantQuota {
    const baseQuota: TenantQuota = {
      maxUsers: 5,
      maxApiCalls: 1000,
      maxMemoryBytes: 100 * 1024 * 1024,
      maxSessions: 10,
      maxPlugins: 5,
      rateLimitPerMinute: 60,
    };

    switch (plan) {
      case 'free':
        return baseQuota;
      case 'pro':
        return {
          ...baseQuota,
          maxUsers: 50,
          maxApiCalls: 50000,
          maxMemoryBytes: 1024 * 1024 * 1024,
          maxSessions: 100,
          maxPlugins: 50,
          rateLimitPerMinute: 300,
        };
      case 'enterprise':
        return {
          ...baseQuota,
          maxUsers: Infinity,
          maxApiCalls: Infinity,
          maxMemoryBytes: Infinity,
          maxSessions: Infinity,
          maxPlugins: Infinity,
          rateLimitPerMinute: Infinity,
        };
    }
  }
}

// ==================== 导出 ====================

export { RBAC, TenantManager };
export type { 
  Permission, Role, RolePermissions,
  User, ApiKey, AuditLog,
  Tenant, TenantQuota 
};

/**
 * OAuth Provider - OAuth 2.0 认证提供者
 */

export interface OAuthConfig {
  clientId: string;
  clientSecret?: string;
  redirectUri: string;
  scope: string[];
  authUrl: string;
  tokenUrl: string;
}

export interface OAuthToken {
  accessToken: string;
  refreshToken?: string;
  expiresAt: Date;
  scope: string[];
}

export interface OAuthUser {
  id: string;
  name: string;
  email?: string;
  avatar?: string;
  raw: Record<string, unknown>;
}

/**
 * OAuth 提供者基类
 */
export abstract class OAuthProvider {
  protected config: OAuthConfig;
  protected token: OAuthToken | null = null;
  protected state: string = '';

  constructor(config: OAuthConfig) {
    this.config = config;
  }

  /**
   * 获取授权 URL
   */
  abstract getAuthUrl(): string;

  /**
   * 处理回调
   */
  abstract handleCallback(code: string): Promise<OAuthToken>;

  /**
   * 获取用户信息
   */
  abstract getUserInfo(): Promise<OAuthUser>;

  /**
   * 刷新令牌
   */
  abstract refreshToken(): Promise<OAuthToken>;

  /**
   * 检查是否已授权
   */
  isAuthenticated(): boolean {
    if (!this.token) return false;
    return this.token.expiresAt > new Date();
  }

  /**
   * 生成状态参数
   */
  protected generateState(): string {
    this.state = Math.random().toString(36).substring(2);
    return this.state;
  }

  /**
   * 验证状态
   */
  protected verifyState(state: string): boolean {
    return state === this.state;
  }

  /**
   * 获取访问令牌
   */
  getAccessToken(): string | null {
    return this.token?.accessToken || null;
  }

  /**
   * 登出
   */
  logout(): void {
    this.token = null;
    this.state = '';
  }
}

// Google OAuth
export class GoogleOAuth extends OAuthProvider {
  constructor(redirectUri: string, clientId: string, clientSecret?: string) {
    super({
      clientId,
      clientSecret,
      redirectUri,
      scope: ['openid', 'email', 'profile'],
      authUrl: 'https://accounts.google.com/o/oauth2/v2/auth',
      tokenUrl: 'https://oauth2.googleapis.com/token',
    });
  }

  getAuthUrl(): string {
    const params = new URLSearchParams({
      client_id: this.config.clientId,
      redirect_uri: this.config.redirectUri,
      response_type: 'code',
      scope: this.config.scope.join(' '),
      state: this.generateState(),
      access_type: 'offline',
      prompt: 'consent',
    });
    return `${this.config.authUrl}?${params.toString()}`;
  }

  async handleCallback(code: string): Promise<OAuthToken> {
    const response = await fetch(this.config.tokenUrl, {
      method: 'POST',
      headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
      body: new URLSearchParams({
        code,
        client_id: this.config.clientId,
        client_secret: this.config.clientSecret || '',
        redirect_uri: this.config.redirectUri,
        grant_type: 'authorization_code',
      }),
    });

    const data = await response.json();
    
    this.token = {
      accessToken: data.access_token,
      refreshToken: data.refresh_token,
      expiresAt: new Date(Date.now() + data.expires_in * 1000),
      scope: data.scope?.split(' ') || this.config.scope,
    };

    return this.token;
  }

  async getUserInfo(): Promise<OAuthUser> {
    const response = await fetch('https://www.googleapis.com/oauth2/v2/userinfo', {
      headers: { Authorization: `Bearer ${this.token?.accessToken}` },
    });

    const data = await response.json();
    
    return {
      id: data.id,
      name: data.name,
      email: data.email,
      avatar: data.picture,
      raw: data,
    };
  }

  async refreshToken(): Promise<OAuthToken> {
    if (!this.token?.refreshToken) {
      throw new Error('No refresh token available');
    }

    const response = await fetch(this.config.tokenUrl, {
      method: 'POST',
      headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
      body: new URLSearchParams({
        refresh_token: this.token.refreshToken,
        client_id: this.config.clientId,
        client_secret: this.config.clientSecret || '',
        grant_type: 'refresh_token',
      }),
    });

    const data = await response.json();
    
    this.token = {
      accessToken: data.access_token,
      refreshToken: this.token.refreshToken,
      expiresAt: new Date(Date.now() + data.expires_in * 1000),
      scope: data.scope?.split(' ') || this.config.scope,
    };

    return this.token;
  }
}

// GitHub OAuth
export class GitHubOAuth extends OAuthProvider {
  constructor(redirectUri: string, clientId: string, clientSecret?: string) {
    super({
      clientId,
      clientSecret,
      redirectUri,
      scope: ['read:user', 'user:email'],
      authUrl: 'https://github.com/login/oauth/authorize',
      tokenUrl: 'https://github.com/login/oauth/access_token',
    });
  }

  getAuthUrl(): string {
    const params = new URLSearchParams({
      client_id: this.config.clientId,
      redirect_uri: this.config.redirectUri,
      scope: this.config.scope.join(' '),
      state: this.generateState(),
    });
    return `${this.config.authUrl}?${params.toString()}`;
  }

  async handleCallback(code: string): Promise<OAuthToken> {
    const response = await fetch(this.config.tokenUrl, {
      method: 'POST',
      headers: { Accept: 'application/json' },
      body: new URLSearchParams({
        code,
        client_id: this.config.clientId,
        client_secret: this.config.clientSecret || '',
        redirect_uri: this.config.redirectUri,
      }),
    });

    const data = await response.json();
    
    this.token = {
      accessToken: data.access_token,
      expiresAt: new Date(Date.now() + 3600 * 1000), // GitHub tokens don't expire
      scope: data.scope?.split(',') || this.config.scope,
    };

    return this.token;
  }

  async getUserInfo(): Promise<OAuthUser> {
    const response = await fetch('https://api.github.com/user', {
      headers: { Authorization: `token ${this.token?.accessToken}` },
    });

    const data = await response.json();
    
    return {
      id: data.id.toString(),
      name: data.name || data.login,
      email: data.email,
      avatar: data.avatar_url,
      raw: data,
    };
  }

  async refreshToken(): Promise<OAuthToken> {
    // GitHub tokens don't need refresh
    return this.token!;
  }
}

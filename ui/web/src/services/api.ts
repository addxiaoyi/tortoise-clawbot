import axios, { AxiosInstance, AxiosError } from 'axios'

// API 基础地址
const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:18792'

// 创建 axios 实例
const apiClient: AxiosInstance = axios.create({
  baseURL: `${API_BASE_URL}/api/v1`,
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// 请求拦截器 - 添加认证
apiClient.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('auth_token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// 响应拦截器 - 处理错误
apiClient.interceptors.response.use(
  (response) => response,
  (error: AxiosError) => {
    if (error.response?.status === 401) {
      // 未授权，清除 token 并跳转登录
      localStorage.removeItem('auth_token')
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

// ==================== 类型定义 ====================

// AI 提供商
export interface AIProvider {
  id: string
  name: string
  enabled: boolean
  api_key?: string
  model: string
  base_url: string
}

// 渠道配置
export interface ChannelConfig {
  enabled: boolean
  botToken?: string
  guildId?: string
  allowedChats?: string
  signingSecret?: string
  webhookUrl?: string
}

// 渠道状态
export interface ChannelStatus {
  type: string
  name: string
  connected: boolean
  messageCount: number
  lastMessage?: string
}

// 记忆
export interface Memory {
  id: string
  content: string
  tags: string[]
  createdAt: string
}

// 会话
export interface Session {
  id: string
  userId: string
  model: string
  createdAt: string
  messageCount: number
}

// 消息
export interface Message {
  id: string
  role: 'user' | 'assistant' | 'system'
  content: string
  timestamp: string
}

// 统计
export interface Stats {
  ai: {
    requestsTotal: number
    requestsSuccess: number
    requestsFailed: number
    avgLatencyMs: number
    tokensUsed: number
    costUsd: number
  }
  channels: {
    messagesReceived: number
    messagesSent: number
    messagesDropped: number
    channelsActive: number
  }
}

// API Key
export interface APIKey {
  id: string
  key: string
  name: string
  created_at: string
}

// 完整配置
export interface FullConfig {
  server: {
    port: number
    host: string
    log_level: string
  }
  ai: {
    providers: AIProvider[]
    default_model: string
    routing: string
  }
  channels: {
    telegram: ChannelConfig
    discord: ChannelConfig
    slack: ChannelConfig
    whatsapp: ChannelConfig
    teams: ChannelConfig
  }
  discovery: {
    enabled: boolean
    mdns: boolean
    upnp: boolean
    ssdp: boolean
    port: number
    advertise_interval: number
  }
  database: {
    type: string
    sqlite: { path: string }
    redis: { url: string; password: string; db: number }
  }
  security: {
    api_key_required: boolean
    api_keys: APIKey[]
    jwt_secret: string
    rate_limit: { enabled: boolean; requests_per_minute: number }
    cors: { enabled: boolean; allowed_origins: string[] }
  }
  advanced: {
    max_sessions: number
    session_timeout: number
    message_buffer_size: number
    worker_pool_size: number
    memory_pool_size: number
    log_level: string
    enable_metrics: boolean
    enable_tracing: boolean
  }
}

// ==================== API 函数 ====================

export const api = {
  // 健康检查
  health: async (): Promise<{ status: string; version: string }> => {
    const response = await apiClient.get('/health')
    return response.data
  },

  // 会话管理
  sessions: {
    list: async (): Promise<Session[]> => {
      const response = await apiClient.get('/sessions')
      return response.data
    },
    create: async (userId: string, model?: string): Promise<{ id: string }> => {
      const response = await apiClient.post('/sessions', { user_id: userId, model })
      return response.data
    },
    get: async (id: string): Promise<Session> => {
      const response = await apiClient.get(`/sessions/${id}`)
      return response.data
    },
    delete: async (id: string): Promise<void> => {
      await apiClient.delete(`/sessions/${id}`)
    },
    messages: {
      list: async (sessionId: string): Promise<Message[]> => {
        const response = await apiClient.get(`/sessions/${sessionId}/messages`)
        return response.data
      },
      send: async (
        sessionId: string,
        content: string
      ): Promise<{ user_message: string; ai_response: string; usage: any }> => {
        const response = await apiClient.post(`/sessions/${sessionId}/messages`, {
          content,
        })
        return response.data
      },
    },
  },

  // 记忆管理
  memories: {
    list: async (query?: string): Promise<Memory[]> => {
      const params = query ? { q: query } : {}
      const response = await apiClient.get('/memories', { params })
      return response.data
    },
    add: async (content: string, tags?: string[]): Promise<{ id: string }> => {
      const response = await apiClient.post('/memories', { content, tags })
      return response.data
    },
    delete: async (id: string): Promise<void> => {
      await apiClient.delete(`/memories/${id}`)
    },
  },

  // AI 提供商
  ai: {
    providers: async (): Promise<AIProvider[]> => {
      const response = await apiClient.get('/ai/providers')
      return response.data
    },
    chat: async (
      model: string,
      messages: { role: string; content: string }[],
      temperature?: number,
      maxTokens?: number
    ): Promise<any> => {
      const response = await apiClient.post('/ai/chat', {
        model,
        messages,
        temperature,
        max_tokens: maxTokens,
      })
      return response.data
    },
  },

  // 渠道管理
  channels: {
    list: async (): Promise<ChannelStatus[]> => {
      const response = await apiClient.get('/channels')
      return response.data
    },
    connect: async (type: string): Promise<void> => {
      await apiClient.post(`/channels/${type}/connect`)
    },
    disconnect: async (type: string): Promise<void> => {
      await apiClient.post(`/channels/${type}/disconnect`)
    },
    status: async (type: string): Promise<ChannelStatus> => {
      const response = await apiClient.get(`/channels/${type}/status`)
      return response.data
    },
  },

  // 配置管理
  config: {
    get: async (): Promise<FullConfig> => {
      const response = await apiClient.get('/config')
      return response.data
    },
    update: async (updates: Partial<FullConfig>): Promise<FullConfig> => {
      const response = await apiClient.patch('/config', updates)
      return response.data
    },
  },

  // 统计
  stats: async (): Promise<Stats> => {
    const response = await apiClient.get('/stats')
    return response.data
  },

  // API Keys
  apiKeys: {
    list: async (): Promise<APIKey[]> => {
      const response = await apiClient.get('/api-keys')
      return response.data
    },
    create: async (name: string): Promise<APIKey> => {
      const response = await apiClient.post('/api-keys', { name })
      return response.data
    },
    delete: async (id: string): Promise<void> => {
      await apiClient.delete(`/api-keys/${id}`)
    },
  },
}

// WebSocket 服务
export class WebSocketService {
  private ws: WebSocket | null = null
  private reconnectAttempts = 0
  private maxReconnectAttempts = 5
  private reconnectDelay = 1000
  private listeners: Map<string, Set<(data: any) => void>> = new Map()

  connect(url?: string) {
    const wsUrl = url || `${API_BASE_URL.replace('http', 'ws')}/ws`
    
    this.ws = new WebSocket(wsUrl)

    this.ws.onopen = () => {
      console.log('WebSocket connected')
      this.reconnectAttempts = 0
      this.emit('connected', null)
    }

    this.ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data)
        this.emit(data.type, data)
        this.emit('message', data)
      } catch (e) {
        console.error('Failed to parse WebSocket message:', e)
      }
    }

    this.ws.onclose = () => {
      console.log('WebSocket disconnected')
      this.emit('disconnected', null)
      this.attemptReconnect()
    }

    this.ws.onerror = (error) => {
      console.error('WebSocket error:', error)
      this.emit('error', error)
    }
  }

  disconnect() {
    if (this.ws) {
      this.ws.close()
      this.ws = null
    }
  }

  send(data: any) {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(data))
    }
  }

  chat(
    messages: { role: string; content: string }[],
    model?: string,
    temperature?: number
  ) {
    this.send({
      type: 'chat',
      id: crypto.randomUUID(),
      model: model || 'gpt-4',
      messages,
      temperature: temperature || 0.7,
    })
  }

  ping() {
    this.send({ type: 'ping' })
  }

  on(event: string, callback: (data: any) => void) {
    if (!this.listeners.has(event)) {
      this.listeners.set(event, new Set())
    }
    this.listeners.get(event)!.add(callback)
  }

  off(event: string, callback: (data: any) => void) {
    this.listeners.get(event)?.delete(callback)
  }

  private emit(event: string, data: any) {
    this.listeners.get(event)?.forEach((callback) => callback(data))
  }

  private attemptReconnect() {
    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      this.reconnectAttempts++
      setTimeout(() => {
        console.log(`Attempting to reconnect (${this.reconnectAttempts}/${this.maxReconnectAttempts})`)
        this.connect()
      }, this.reconnectDelay * this.reconnectAttempts)
    }
  }
}

// 导出单例
export const wsService = new WebSocketService()

export default api

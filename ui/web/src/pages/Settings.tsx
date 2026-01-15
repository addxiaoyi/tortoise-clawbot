import { useState, useEffect } from 'react'
import { useAppStore } from '../store/appStore'
import { api } from '../services/api'
import clsx from 'clsx'
import {
  Settings as SettingsIcon,
  Globe,
  Palette,
  Bot,
  MessageSquare,
  Puzzle,
  Network,
  Key,
  Database,
  Terminal,
  Save,
  RefreshCw,
  Plus,
  Trash2,
  Eye,
  EyeOff,
  Check,
  X,
  AlertCircle,
  CheckCircle2,
} from 'lucide-react'

// 配置标签页
const tabs = [
  { id: 'ai', name: 'AI 配置', icon: Bot },
  { id: 'channels', name: '消息渠道', icon: MessageSquare },
  { id: 'plugins', name: '插件管理', icon: Puzzle },
  { id: 'discovery', name: '设备发现', icon: Network },
  { id: 'security', name: '安全设置', icon: Key },
  { id: 'database', name: '数据库', icon: Database },
  { id: 'advanced', name: '高级设置', icon: Terminal },
]

// 类型定义
interface AIProvider {
  id: string
  name: string
  enabled: boolean
  api_key?: string
  model: string
  base_url: string
}

interface ChannelConfig {
  enabled: boolean
  botToken?: string
  guildId?: string
  allowedChats?: string
  signingSecret?: string
  webhookUrl?: string
  phoneNumber?: string
  apiUrl?: string
  apiToken?: string
  appId?: string
  appPassword?: string
  tenantId?: string
}

interface DiscoveryConfig {
  enabled: boolean
  mdns: boolean
  upnp: boolean
  ssdp: boolean
  port: number
  advertiseInterval: number
}

interface DBConfig {
  type: string
  sqlite: { path: string }
  redis: { url: string; password: string; db: number }
}

interface SecurityConfig {
  apiKeyRequired: boolean
  apiKeys: Array<{ id: string; key: string; name: string; created_at: string }>
  jwtSecret: string
  rateLimit: { enabled: boolean; requestsPerMinute: number }
  cors: { enabled: boolean; allowedOrigins: string[] }
}

interface AdvancedConfig {
  maxSessions: number
  sessionTimeout: number
  messageBufferSize: number
  workerPoolSize: number
  memoryPoolSize: number
  logLevel: string
  enableMetrics: boolean
  enableTracing: boolean
}

interface FullConfig {
  ai: { providers: AIProvider[]; routing: string; default_model: string }
  channels: {
    telegram: ChannelConfig
    discord: ChannelConfig
    slack: ChannelConfig
    whatsapp: ChannelConfig
    teams: ChannelConfig
  }
  discovery: DiscoveryConfig
  database: DBConfig
  security: SecurityConfig
  advanced: AdvancedConfig
}

export default function Settings() {
  const { theme, setTheme } = useAppStore()
  const [activeTab, setActiveTab] = useState('ai')
  const [isSaving, setIsSaving] = useState(false)
  const [showPasswords, setShowPasswords] = useState<Record<string, boolean>>({})
  const [loadError, setLoadError] = useState<string | null>(null)
  const [saveStatus, setSaveStatus] = useState<'idle' | 'success' | 'error'>('idle')

  // AI 配置状态
  const [aiConfig, setAiConfig] = useState({
    providers: [
      {
        id: 'openai',
        name: 'OpenAI',
        enabled: true,
        api_key: '',
        model: 'gpt-4',
        base_url: 'https://api.openai.com/v1',
      },
      {
        id: 'anthropic',
        name: 'Anthropic',
        enabled: false,
        api_key: '',
        model: 'claude-3-sonnet-20240229',
        base_url: 'https://api.anthropic.com',
      },
      {
        id: 'ollama',
        name: 'Ollama (本地)',
        enabled: false,
        api_key: '',
        model: 'llama2',
        base_url: 'http://localhost:11434',
      },
    ],
    routing: 'latency',
    defaultModel: 'gpt-4',
  })

  // 渠道配置状态
  const [channelConfig, setChannelConfig] = useState({
    telegram: {
      enabled: false,
      botToken: '',
      allowedChats: '',
    },
    discord: {
      enabled: false,
      botToken: '',
      guildId: '',
    },
    slack: {
      enabled: false,
      botToken: '',
      signingSecret: '',
      webhookUrl: '',
    },
    whatsapp: {
      enabled: false,
      phoneNumber: '',
      apiUrl: '',
      apiToken: '',
    },
    teams: {
      enabled: false,
      appId: '',
      appPassword: '',
      tenantId: '',
    },
  })

  // 设备发现配置
  const [discoveryConfig, setDiscoveryConfig] = useState({
    enabled: true,
    mdns: true,
    upnp: true,
    ssdp: false,
    port: 18792,
    advertiseInterval: 30,
  })

  // 数据库配置
  const [dbConfig, setDbConfig] = useState({
    type: 'memory',
    sqlite: { path: './data/tortoise.db' },
    redis: { url: 'redis://localhost:6379', password: '', db: 0 },
  })

  // 安全配置
  const [securityConfig, setSecurityConfig] = useState({
    apiKeyRequired: false,
    apiKeys: [] as Array<{ id: string; key: string; name: string; created_at: string }>,
    jwtSecret: '',
    rateLimit: { enabled: true, requestsPerMinute: 60 },
    cors: { enabled: true, allowedOrigins: ['*'] as string[] },
  })

  // 高级配置
  const [advancedConfig, setAdvancedConfig] = useState({
    maxSessions: 100000,
    sessionTimeout: 86400,
    messageBufferSize: 10000,
    workerPoolSize: 10000,
    memoryPoolSize: 1024,
    logLevel: 'info',
    enableMetrics: true,
    enableTracing: false,
  })

  // 加载配置
  useEffect(() => {
    loadConfig()
  }, [])

  const loadConfig = async () => {
    try {
      setLoadError(null)
      const config: FullConfig = await api.getConfig()
      
      // 转换后端格式到前端格式
      setAiConfig({
        providers: config.ai.providers.map(p => ({
          id: p.id,
          name: p.name,
          enabled: p.enabled,
          api_key: p.api_key || '',
          model: p.model,
          base_url: p.base_url,
        })),
        routing: config.ai.routing,
        defaultModel: config.ai.default_model,
      })

      setChannelConfig({
        telegram: {
          enabled: config.channels.telegram.enabled,
          botToken: config.channels.telegram.botToken || '',
          allowedChats: config.channels.telegram.allowedChats || '',
        },
        discord: {
          enabled: config.channels.discord.enabled,
          botToken: config.channels.discord.botToken || '',
          guildId: config.channels.discord.guildId || '',
        },
        slack: {
          enabled: config.channels.slack.enabled,
          botToken: config.channels.slack.botToken || '',
          signingSecret: config.channels.slack.signingSecret || '',
          webhookUrl: config.channels.slack.webhookUrl || '',
        },
        whatsapp: {
          enabled: config.channels.whatsapp.enabled,
          phoneNumber: config.channels.whatsapp.phoneNumber || '',
          apiUrl: config.channels.whatsapp.apiUrl || '',
          apiToken: config.channels.whatsapp.apiToken || '',
        },
        teams: {
          enabled: config.channels.teams.enabled,
          appId: config.channels.teams.appId || '',
          appPassword: config.channels.teams.appPassword || '',
          tenantId: config.channels.teams.tenantId || '',
        },
      })

      setDiscoveryConfig({
        enabled: config.discovery.enabled,
        mdns: config.discovery.mdns,
        upnp: config.discovery.upnp,
        ssdp: config.discovery.ssdp,
        port: config.discovery.port,
        advertiseInterval: config.discovery.advertiseInterval,
      })

      setDbConfig({
        type: config.database.type,
        sqlite: config.database.sqlite,
        redis: config.database.redis,
      })

      setSecurityConfig({
        apiKeyRequired: config.security.apiKeyRequired,
        apiKeys: config.security.apiKeys || [],
        jwtSecret: config.security.jwtSecret || '',
        rateLimit: config.security.rateLimit,
        cors: config.security.cors,
      })

      setAdvancedConfig({
        maxSessions: config.advanced.maxSessions,
        sessionTimeout: config.advanced.sessionTimeout,
        messageBufferSize: config.advanced.messageBufferSize,
        workerPoolSize: config.advanced.workerPoolSize,
        memoryPoolSize: config.advanced.memoryPoolSize,
        logLevel: config.advanced.logLevel,
        enableMetrics: config.advanced.enableMetrics,
        enableTracing: config.advanced.enableTracing,
      })
    } catch (error) {
      console.error('加载配置失败:', error)
      setLoadError('加载配置失败，请检查后端服务是否运行')
    }
  }

  const handleSave = async () => {
    setIsSaving(true)
    setSaveStatus('idle')
    
    try {
      // 转换前端格式到后端格式
      const config: Partial<FullConfig> = {
        ai: {
          providers: aiConfig.providers.map(p => ({
            ...p,
            api_key: p.api_key,
          })),
          routing: aiConfig.routing,
          default_model: aiConfig.defaultModel,
        },
        channels: {
          telegram: {
            enabled: channelConfig.telegram.enabled,
            bot_token: channelConfig.telegram.botToken,
            allowed_chats: channelConfig.telegram.allowedChats,
          },
          discord: {
            enabled: channelConfig.discord.enabled,
            bot_token: channelConfig.discord.botToken,
            guild_id: channelConfig.discord.guildId,
          },
          slack: {
            enabled: channelConfig.slack.enabled,
            bot_token: channelConfig.slack.botToken,
            signing_secret: channelConfig.slack.signingSecret,
            webhook_url: channelConfig.slack.webhookUrl,
          },
          whatsapp: {
            enabled: channelConfig.whatsapp.enabled,
            phone_number: channelConfig.whatsapp.phoneNumber,
            api_url: channelConfig.whatsapp.apiUrl,
            api_token: channelConfig.whatsapp.apiToken,
          },
          teams: {
            enabled: channelConfig.teams.enabled,
            app_id: channelConfig.teams.appId,
            app_password: channelConfig.teams.appPassword,
            tenant_id: channelConfig.teams.tenantId,
          },
        },
        discovery: {
          enabled: discoveryConfig.enabled,
          mdns: discoveryConfig.mdns,
          upnp: discoveryConfig.upnp,
          ssdp: discoveryConfig.ssdp,
          port: discoveryConfig.port,
          advertise_interval: discoveryConfig.advertiseInterval,
        },
        database: {
          type: dbConfig.type,
          sqlite: dbConfig.sqlite,
          redis: dbConfig.redis,
        },
        security: {
          apiKeyRequired: securityConfig.apiKeyRequired,
          apiKeys: securityConfig.apiKeys,
          jwtSecret: securityConfig.jwtSecret,
          rateLimit: securityConfig.rateLimit,
          cors: securityConfig.cors,
        },
        advanced: advancedConfig,
      }

      await api.updateConfig(config as Record<string, unknown>)
      setSaveStatus('success')
      setTimeout(() => setSaveStatus('idle'), 3000)
    } catch (error) {
      console.error('保存配置失败:', error)
      setSaveStatus('error')
    } finally {
      setIsSaving(false)
    }
  }

  const togglePasswordVisibility = (key: string) => {
    setShowPasswords((prev) => ({ ...prev, [key]: !prev[key] }))
  }

  return (
    <div className="flex h-full gap-6">
      {/* Tabs */}
      <div className="w-56 flex-shrink-0">
        <div className="card p-2">
          {tabs.map((tab) => {
            const Icon = tab.icon
            return (
              <button
                key={tab.id}
                onClick={() => setActiveTab(tab.id)}
                className={clsx(
                  'w-full flex items-center gap-3 px-4 py-3 rounded-lg text-left transition-colors',
                  activeTab === tab.id
                    ? 'bg-tortoise-500/20 text-tortoise-400'
                    : 'text-gray-400 hover:bg-dark-200 hover:text-white'
                )}
              >
                <Icon className="w-5 h-5" />
                <span>{tab.name}</span>
              </button>
            )
          })}
        </div>
      </div>

      {/* Content */}
      <div className="flex-1 overflow-y-auto">
        <div className="card p-6 space-y-6">
          
          {/* 错误提示 */}
          {loadError && (
            <div className="flex items-center gap-2 p-4 bg-red-500/20 border border-red-500/50 rounded-lg text-red-400">
              <AlertCircle className="w-5 h-5" />
              <span>{loadError}</span>
              <button 
                onClick={loadConfig}
                className="ml-auto text-sm hover:underline"
              >
                重试
              </button>
            </div>
          )}

          {/* AI 配置 */}
          {activeTab === 'ai' && (
            <div className="space-y-6">
              <div className="flex items-center justify-between">
                <h2 className="text-xl font-semibold text-white">AI 配置</h2>
                <span className="badge badge-success">核心功能</span>
              </div>

              {/* AI 提供商 */}
              <div className="space-y-4">
                <h3 className="text-lg font-medium text-white">AI 提供商</h3>
                {aiConfig.providers.map((provider) => (
                  <div key={provider.id} className="card p-4 bg-dark-100">
                    <div className="flex items-center justify-between mb-4">
                      <div className="flex items-center gap-3">
                        <button
                          onClick={() => {
                            const updated = aiConfig.providers.map((p) =>
                              p.id === provider.id ? { ...p, enabled: !p.enabled } : p
                            )
                            setAiConfig({ ...aiConfig, providers: updated })
                          }}
                          className={clsx(
                            'w-12 h-6 rounded-full transition-colors relative',
                            provider.enabled ? 'bg-tortoise-500' : 'bg-dark-300'
                          )}
                        >
                          <span
                            className={clsx(
                              'absolute top-1 w-4 h-4 bg-white rounded-full transition-transform',
                              provider.enabled ? 'left-7' : 'left-1'
                            )}
                          />
                        </button>
                        <span className="text-white font-medium">{provider.name}</span>
                      </div>
                    </div>

                    {provider.enabled && (
                      <div className="space-y-4">
                        {/* API Key */}
                        <div>
                          <label className="block text-sm text-gray-400 mb-2">API Key</label>
                          <div className="relative">
                            <input
                              type={showPasswords[provider.id] ? 'text' : 'password'}
                              value={provider.api_key || ''}
                              onChange={(e) => {
                                const updated = aiConfig.providers.map((p) =>
                                  p.id === provider.id ? { ...p, api_key: e.target.value } : p
                                )
                                setAiConfig({ ...aiConfig, providers: updated })
                              }}
                              placeholder="sk-..."
                              className="w-full px-4 py-2 pr-10 bg-dark-200 border border-dark-100 rounded-lg text-white"
                            />
                            <button
                              onClick={() => togglePasswordVisibility(provider.id)}
                              className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 hover:text-white"
                            >
                              {showPasswords[provider.id] ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
                            </button>
                          </div>
                        </div>

                        {/* 模型选择 */}
                        <div>
                          <label className="block text-sm text-gray-400 mb-2">模型</label>
                          <select
                            value={provider.model}
                            onChange={(e) => {
                              const updated = aiConfig.providers.map((p) =>
                                p.id === provider.id ? { ...p, model: e.target.value } : p
                              )
                              setAiConfig({ ...aiConfig, providers: updated })
                            }}
                            className="w-full px-4 py-2 bg-dark-200 border border-dark-100 rounded-lg text-white"
                          >
                            {provider.id === 'openai' && (
                              <>
                                <option value="gpt-4">GPT-4</option>
                                <option value="gpt-4-turbo">GPT-4 Turbo</option>
                                <option value="gpt-3.5-turbo">GPT-3.5 Turbo</option>
                              </>
                            )}
                            {provider.id === 'anthropic' && (
                              <>
                                <option value="claude-3-opus-20240229">Claude 3 Opus</option>
                                <option value="claude-3-sonnet-20240229">Claude 3 Sonnet</option>
                                <option value="claude-3-haiku-20240307">Claude 3 Haiku</option>
                              </>
                            )}
                            {provider.id === 'ollama' && (
                              <>
                                <option value="llama2">Llama 2</option>
                                <option value="mistral">Mistral</option>
                                <option value="codellama">Code Llama</option>
                              </>
                            )}
                          </select>
                        </div>

                        {/* API 地址 */}
                        <div>
                          <label className="block text-sm text-gray-400 mb-2">API 地址</label>
                          <input
                            type="text"
                            value={provider.base_url}
                            onChange={(e) => {
                              const updated = aiConfig.providers.map((p) =>
                                p.id === provider.id ? { ...p, base_url: e.target.value } : p
                              )
                              setAiConfig({ ...aiConfig, providers: updated })
                            }}
                            placeholder="https://api.openai.com/v1"
                            className="w-full px-4 py-2 bg-dark-200 border border-dark-100 rounded-lg text-white"
                          />
                        </div>
                      </div>
                    )}
                  </div>
                ))}
              </div>

              {/* 路由策略 */}
              <div>
                <h3 className="text-lg font-medium text-white mb-4">路由策略</h3>
                <div className="grid grid-cols-3 gap-4">
                  {[
                    { id: 'latency', name: '延迟最优', desc: '选择响应最快的模型' },
                    { id: 'load', name: '负载均衡', desc: '分散请求到多个模型' },
                    { id: 'cost', name: '成本最优', desc: '优先使用便宜的模型' },
                  ].map((strategy) => (
                    <button
                      key={strategy.id}
                      onClick={() => setAiConfig({ ...aiConfig, routing: strategy.id })}
                      className={clsx(
                        'p-4 rounded-xl border transition-colors text-left',
                        aiConfig.routing === strategy.id
                          ? 'border-tortoise-500 bg-tortoise-500/10'
                          : 'border-dark-100 hover:border-dark-200'
                      )}
                    >
                      <p className="text-white font-medium">{strategy.name}</p>
                      <p className="text-sm text-gray-500">{strategy.desc}</p>
                    </button>
                  ))}
                </div>
              </div>
            </div>
          )}

          {/* 消息渠道配置 */}
          {activeTab === 'channels' && (
            <div className="space-y-6">
              <h2 className="text-xl font-semibold text-white">消息渠道</h2>

              {/* Telegram */}
              <div className="card p-4 bg-dark-100">
                <div className="flex items-center justify-between mb-4">
                  <div className="flex items-center gap-3">
                    <div className="w-10 h-10 bg-blue-500/20 rounded-lg flex items-center justify-center">
                      <MessageSquare className="w-5 h-5 text-blue-400" />
                    </div>
                    <div>
                      <p className="text-white font-medium">Telegram</p>
                      <p className="text-sm text-gray-500">即时消息</p>
                    </div>
                  </div>
                  <button
                    onClick={() => setChannelConfig({
                      ...channelConfig,
                      telegram: { ...channelConfig.telegram, enabled: !channelConfig.telegram.enabled }
                    })}
                    className={clsx(
                      'w-12 h-6 rounded-full transition-colors relative',
                      channelConfig.telegram.enabled ? 'bg-tortoise-500' : 'bg-dark-300'
                    )}
                  >
                    <span
                      className={clsx(
                        'absolute top-1 w-4 h-4 bg-white rounded-full transition-transform',
                        channelConfig.telegram.enabled ? 'left-7' : 'left-1'
                      )}
                    />
                  </button>
                </div>
                {channelConfig.telegram.enabled && (
                  <div className="space-y-4">
                    <div>
                      <label className="block text-sm text-gray-400 mb-2">Bot Token</label>
                      <input
                        type="password"
                        value={channelConfig.telegram.botToken}
                        onChange={(e) => setChannelConfig({
                          ...channelConfig,
                          telegram: { ...channelConfig.telegram, botToken: e.target.value }
                        })}
                        placeholder="123456789:ABCdefGHIjklMNOpqrsTUVwxyz"
                        className="w-full px-4 py-2 bg-dark-200 border border-dark-100 rounded-lg text-white"
                      />
                    </div>
                    <div>
                      <label className="block text-sm text-gray-400 mb-2">允许的 Chat ID (逗号分隔)</label>
                      <input
                        type="text"
                        value={channelConfig.telegram.allowedChats}
                        onChange={(e) => setChannelConfig({
                          ...channelConfig,
                          telegram: { ...channelConfig.telegram, allowedChats: e.target.value }
                        })}
                        placeholder="123456789, 987654321"
                        className="w-full px-4 py-2 bg-dark-200 border border-dark-100 rounded-lg text-white"
                      />
                    </div>
                  </div>
                )}
              </div>

              {/* Discord */}
              <div className="card p-4 bg-dark-100">
                <div className="flex items-center justify-between mb-4">
                  <div className="flex items-center gap-3">
                    <div className="w-10 h-10 bg-indigo-500/20 rounded-lg flex items-center justify-center">
                      <MessageSquare className="w-5 h-5 text-indigo-400" />
                    </div>
                    <div>
                      <p className="text-white font-medium">Discord</p>
                      <p className="text-sm text-gray-500">社区/服务器</p>
                    </div>
                  </div>
                  <button
                    onClick={() => setChannelConfig({
                      ...channelConfig,
                      discord: { ...channelConfig.discord, enabled: !channelConfig.discord.enabled }
                    })}
                    className={clsx(
                      'w-12 h-6 rounded-full transition-colors relative',
                      channelConfig.discord.enabled ? 'bg-tortoise-500' : 'bg-dark-300'
                    )}
                  >
                    <span
                      className={clsx(
                        'absolute top-1 w-4 h-4 bg-white rounded-full transition-transform',
                        channelConfig.discord.enabled ? 'left-7' : 'left-1'
                      )}
                    />
                  </button>
                </div>
                {channelConfig.discord.enabled && (
                  <div className="space-y-4">
                    <div>
                      <label className="block text-sm text-gray-400 mb-2">Bot Token</label>
                      <input
                        type="password"
                        value={channelConfig.discord.botToken}
                        onChange={(e) => setChannelConfig({
                          ...channelConfig,
                          discord: { ...channelConfig.discord, botToken: e.target.value }
                        })}
                        placeholder="MTEwMTExMTExMTEx.exampl3.tOkEnHeRe"
                        className="w-full px-4 py-2 bg-dark-200 border border-dark-100 rounded-lg text-white"
                      />
                    </div>
                    <div>
                      <label className="block text-sm text-gray-400 mb-2">Guild ID (服务器 ID)</label>
                      <input
                        type="text"
                        value={channelConfig.discord.guildId}
                        onChange={(e) => setChannelConfig({
                          ...channelConfig,
                          discord: { ...channelConfig.discord, guildId: e.target.value }
                        })}
                        placeholder="123456789012345678"
                        className="w-full px-4 py-2 bg-dark-200 border border-dark-100 rounded-lg text-white"
                      />
                    </div>
                  </div>
                )}
              </div>

              {/* Slack */}
              <div className="card p-4 bg-dark-100">
                <div className="flex items-center justify-between mb-4">
                  <div className="flex items-center gap-3">
                    <div className="w-10 h-10 bg-green-500/20 rounded-lg flex items-center justify-center">
                      <MessageSquare className="w-5 h-5 text-green-400" />
                    </div>
                    <div>
                      <p className="text-white font-medium">Slack</p>
                      <p className="text-sm text-gray-500">团队协作</p>
                    </div>
                  </div>
                  <button
                    onClick={() => setChannelConfig({
                      ...channelConfig,
                      slack: { ...channelConfig.slack, enabled: !channelConfig.slack.enabled }
                    })}
                    className={clsx(
                      'w-12 h-6 rounded-full transition-colors relative',
                      channelConfig.slack.enabled ? 'bg-tortoise-500' : 'bg-dark-300'
                    )}
                  >
                    <span
                      className={clsx(
                        'absolute top-1 w-4 h-4 bg-white rounded-full transition-transform',
                        channelConfig.slack.enabled ? 'left-7' : 'left-1'
                      )}
                    />
                  </button>
                </div>
                {channelConfig.slack.enabled && (
                  <div className="space-y-4">
                    <div>
                      <label className="block text-sm text-gray-400 mb-2">Bot Token</label>
                      <input
                        type="password"
                        value={channelConfig.slack.botToken}
                        onChange={(e) => setChannelConfig({
                          ...channelConfig,
                          slack: { ...channelConfig.slack, botToken: e.target.value }
                        })}
                        placeholder="xoxb-..."
                        className="w-full px-4 py-2 bg-dark-200 border border-dark-100 rounded-lg text-white"
                      />
                    </div>
                    <div>
                      <label className="block text-sm text-gray-400 mb-2">Signing Secret</label>
                      <input
                        type="password"
                        value={channelConfig.slack.signingSecret}
                        onChange={(e) => setChannelConfig({
                          ...channelConfig,
                          slack: { ...channelConfig.slack, signingSecret: e.target.value }
                        })}
                        placeholder="..."
                        className="w-full px-4 py-2 bg-dark-200 border border-dark-100 rounded-lg text-white"
                      />
                    </div>
                  </div>
                )}
              </div>

              {/* WhatsApp */}
              <div className="card p-4 bg-dark-100">
                <div className="flex items-center justify-between mb-4">
                  <div className="flex items-center gap-3">
                    <div className="w-10 h-10 bg-green-600/20 rounded-lg flex items-center justify-center">
                      <MessageSquare className="w-5 h-5 text-green-500" />
                    </div>
                    <div>
                      <p className="text-white font-medium">WhatsApp</p>
                      <p className="text-sm text-gray-500">企业微信/WhatsApp Business</p>
                    </div>
                  </div>
                  <button
                    onClick={() => setChannelConfig({
                      ...channelConfig,
                      whatsapp: { ...channelConfig.whatsapp, enabled: !channelConfig.whatsapp.enabled }
                    })}
                    className={clsx(
                      'w-12 h-6 rounded-full transition-colors relative',
                      channelConfig.whatsapp.enabled ? 'bg-tortoise-500' : 'bg-dark-300'
                    )}
                  >
                    <span
                      className={clsx(
                        'absolute top-1 w-4 h-4 bg-white rounded-full transition-transform',
                        channelConfig.whatsapp.enabled ? 'left-7' : 'left-1'
                      )}
                    />
                  </button>
                </div>
              </div>

              {/* Teams */}
              <div className="card p-4 bg-dark-100">
                <div className="flex items-center justify-between mb-4">
                  <div className="flex items-center gap-3">
                    <div className="w-10 h-10 bg-purple-500/20 rounded-lg flex items-center justify-center">
                      <MessageSquare className="w-5 h-5 text-purple-400" />
                    </div>
                    <div>
                      <p className="text-white font-medium">Microsoft Teams</p>
                      <p className="text-sm text-gray-500">企业协作</p>
                    </div>
                  </div>
                  <button
                    onClick={() => setChannelConfig({
                      ...channelConfig,
                      teams: { ...channelConfig.teams, enabled: !channelConfig.teams.enabled }
                    })}
                    className={clsx(
                      'w-12 h-6 rounded-full transition-colors relative',
                      channelConfig.teams.enabled ? 'bg-tortoise-500' : 'bg-dark-300'
                    )}
                  >
                    <span
                      className={clsx(
                        'absolute top-1 w-4 h-4 bg-white rounded-full transition-transform',
                        channelConfig.teams.enabled ? 'left-7' : 'left-1'
                      )}
                    />
                  </button>
                </div>
              </div>
            </div>
          )}

          {/* 设备发现配置 */}
          {activeTab === 'discovery' && (
            <div className="space-y-6">
              <h2 className="text-xl font-semibold text-white">设备发现</h2>

              <div className="card p-4 bg-dark-100">
                <div className="flex items-center justify-between mb-4">
                  <div>
                    <p className="text-white font-medium">启用设备发现</p>
                    <p className="text-sm text-gray-500">允许局域网内其他设备发现本服务</p>
                  </div>
                  <button
                    onClick={() => setDiscoveryConfig({ ...discoveryConfig, enabled: !discoveryConfig.enabled })}
                    className={clsx(
                      'w-12 h-6 rounded-full transition-colors relative',
                      discoveryConfig.enabled ? 'bg-tortoise-500' : 'bg-dark-300'
                    )}
                  >
                    <span
                      className={clsx(
                        'absolute top-1 w-4 h-4 bg-white rounded-full transition-transform',
                        discoveryConfig.enabled ? 'left-7' : 'left-1'
                      )}
                    />
                  </button>
                </div>
              </div>

              {discoveryConfig.enabled && (
                <div className="space-y-4">
                  {[
                    { id: 'mdns', name: 'mDNS/Bonjour', desc: '苹果设备发现协议' },
                    { id: 'upnp', name: 'UPnP', desc: '通用即插即用协议' },
                    { id: 'ssdp', name: 'SSDP', desc: '简单服务发现协议' },
                  ].map((protocol) => (
                    <div key={protocol.id} className="card p-4 bg-dark-100">
                      <div className="flex items-center justify-between">
                        <div>
                          <p className="text-white font-medium">{protocol.name}</p>
                          <p className="text-sm text-gray-500">{protocol.desc}</p>
                        </div>
                        <button
                          onClick={() => setDiscoveryConfig({
                            ...discoveryConfig,
                            [protocol.id]: !discoveryConfig[protocol.id as 'mdns' | 'upnp' | 'ssdp']
                          })}
                          className={clsx(
                            'w-12 h-6 rounded-full transition-colors relative',
                            discoveryConfig[protocol.id as 'mdns' | 'upnp' | 'ssdp'] ? 'bg-tortoise-500' : 'bg-dark-300'
                          )}
                        >
                          <span
                            className={clsx(
                              'absolute top-1 w-4 h-4 bg-white rounded-full transition-transform',
                              discoveryConfig[protocol.id as 'mdns' | 'upnp' | 'ssdp'] ? 'left-7' : 'left-1'
                            )}
                          />
                        </button>
                      </div>
                    </div>
                  ))}

                  <div>
                    <label className="block text-sm text-gray-400 mb-2">服务端口</label>
                    <input
                      type="number"
                      value={discoveryConfig.port}
                      onChange={(e) => setDiscoveryConfig({ ...discoveryConfig, port: parseInt(e.target.value) })}
                      className="w-full px-4 py-2 bg-dark-200 border border-dark-100 rounded-lg text-white"
                    />
                  </div>
                </div>
              )}
            </div>
          )}

          {/* 安全设置 */}
          {activeTab === 'security' && (
            <div className="space-y-6">
              <h2 className="text-xl font-semibold text-white">安全设置</h2>

              <div className="space-y-4">
                <div className="card p-4 bg-dark-100">
                  <div className="flex items-center justify-between">
                    <div>
                      <p className="text-white font-medium">API Key 认证</p>
                      <p className="text-sm text-gray-500">要求请求携带有效的 API Key</p>
                    </div>
                    <button
                      onClick={() => setSecurityConfig({ ...securityConfig, apiKeyRequired: !securityConfig.apiKeyRequired })}
                      className={clsx(
                        'w-12 h-6 rounded-full transition-colors relative',
                        securityConfig.apiKeyRequired ? 'bg-tortoise-500' : 'bg-dark-300'
                      )}
                    >
                      <span
                        className={clsx(
                          'absolute top-1 w-4 h-4 bg-white rounded-full transition-transform',
                          securityConfig.apiKeyRequired ? 'left-7' : 'left-1'
                        )}
                      />
                    </button>
                  </div>
                </div>

                {/* API Keys 管理 */}
                {securityConfig.apiKeyRequired && (
                  <div className="card p-4 bg-dark-100">
                    <div className="flex items-center justify-between mb-4">
                      <h3 className="text-white font-medium">API Keys</h3>
                      <button className="btn btn-primary text-sm">
                        <Plus className="w-4 h-4" />
                        添加
                      </button>
                    </div>
                    <div className="space-y-2">
                      {securityConfig.apiKeys.length === 0 ? (
                        <p className="text-gray-500 text-sm">暂无 API Keys</p>
                      ) : (
                        securityConfig.apiKeys.map((key) => (
                          <div key={key.id} className="flex items-center justify-between p-3 bg-dark-200 rounded-lg">
                            <div>
                              <p className="text-white">{key.key}</p>
                              <p className="text-xs text-gray-500">创建于 {key.created_at}</p>
                            </div>
                            <button className="text-red-400 hover:text-red-300">
                              <Trash2 className="w-4 h-4" />
                            </button>
                          </div>
                        ))
                      )}
                    </div>
                  </div>
                )}

                {/* JWT Secret */}
                <div>
                  <label className="block text-sm text-gray-400 mb-2">JWT Secret</label>
                  <input
                    type="password"
                    value={securityConfig.jwtSecret}
                    onChange={(e) => setSecurityConfig({ ...securityConfig, jwtSecret: e.target.value })}
                    placeholder="至少 32 个字符"
                    className="w-full px-4 py-2 bg-dark-200 border border-dark-100 rounded-lg text-white"
                  />
                </div>

                {/* 限流 */}
                <div className="card p-4 bg-dark-100">
                  <div className="flex items-center justify-between mb-4">
                    <div>
                      <p className="text-white font-medium">请求限流</p>
                      <p className="text-sm text-gray-500">防止滥用</p>
                    </div>
                    <button
                      onClick={() => setSecurityConfig({
                        ...securityConfig,
                        rateLimit: { ...securityConfig.rateLimit, enabled: !securityConfig.rateLimit.enabled }
                      })}
                      className={clsx(
                        'w-12 h-6 rounded-full transition-colors relative',
                        securityConfig.rateLimit.enabled ? 'bg-tortoise-500' : 'bg-dark-300'
                      )}
                    >
                      <span
                        className={clsx(
                          'absolute top-1 w-4 h-4 bg-white rounded-full transition-transform',
                          securityConfig.rateLimit.enabled ? 'left-7' : 'left-1'
                        )}
                      />
                    </button>
                  </div>
                  {securityConfig.rateLimit.enabled && (
                    <div>
                      <label className="block text-sm text-gray-400 mb-2">每分钟请求数</label>
                      <input
                        type="number"
                        value={securityConfig.rateLimit.requestsPerMinute}
                        onChange={(e) => setSecurityConfig({
                          ...securityConfig,
                          rateLimit: { ...securityConfig.rateLimit, requestsPerMinute: parseInt(e.target.value) }
                        })}
                        className="w-full px-4 py-2 bg-dark-200 border border-dark-100 rounded-lg text-white"
                      />
                    </div>
                  )}
                </div>
              </div>
            </div>
          )}

          {/* 数据库配置 */}
          {activeTab === 'database' && (
            <div className="space-y-6">
              <h2 className="text-xl font-semibold text-white">数据库</h2>

              <div className="grid grid-cols-3 gap-4">
                {[
                  { id: 'memory', name: '内存存储', desc: '仅开发使用，重启丢失', icon: Terminal },
                  { id: 'sqlite', name: 'SQLite', desc: '轻量级文件数据库', icon: Database },
                  { id: 'redis', name: 'Redis', desc: '高性能内存数据库', icon: Globe },
                ].map((db) => (
                  <button
                    key={db.id}
                    onClick={() => setDbConfig({ ...dbConfig, type: db.id })}
                    className={clsx(
                      'p-4 rounded-xl border transition-colors text-left',
                      dbConfig.type === db.id
                        ? 'border-tortoise-500 bg-tortoise-500/10'
                        : 'border-dark-100 hover:border-dark-200'
                    )}
                  >
                    <db.icon className={clsx(
                      'w-8 h-8 mb-2',
                      dbConfig.type === db.id ? 'text-tortoise-400' : 'text-gray-400'
                    )} />
                    <p className="text-white font-medium">{db.name}</p>
                    <p className="text-sm text-gray-500">{db.desc}</p>
                  </button>
                ))}
              </div>

              {dbConfig.type === 'sqlite' && (
                <div>
                  <label className="block text-sm text-gray-400 mb-2">数据库路径</label>
                  <input
                    type="text"
                    value={dbConfig.sqlite.path}
                    onChange={(e) => setDbConfig({
                      ...dbConfig,
                      sqlite: { ...dbConfig.sqlite, path: e.target.value }
                    })}
                    className="w-full px-4 py-2 bg-dark-200 border border-dark-100 rounded-lg text-white"
                  />
                </div>
              )}

              {dbConfig.type === 'redis' && (
                <div className="space-y-4">
                  <div>
                    <label className="block text-sm text-gray-400 mb-2">Redis URL</label>
                    <input
                      type="text"
                      value={dbConfig.redis.url}
                      onChange={(e) => setDbConfig({
                        ...dbConfig,
                        redis: { ...dbConfig.redis, url: e.target.value }
                      })}
                      placeholder="redis://localhost:6379"
                      className="w-full px-4 py-2 bg-dark-200 border border-dark-100 rounded-lg text-white"
                    />
                  </div>
                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <label className="block text-sm text-gray-400 mb-2">密码 (可选)</label>
                      <input
                        type="password"
                        value={dbConfig.redis.password}
                        onChange={(e) => setDbConfig({
                          ...dbConfig,
                          redis: { ...dbConfig.redis, password: e.target.value }
                        })}
                        className="w-full px-4 py-2 bg-dark-200 border border-dark-100 rounded-lg text-white"
                      />
                    </div>
                    <div>
                      <label className="block text-sm text-gray-400 mb-2">数据库编号</label>
                      <input
                        type="number"
                        value={dbConfig.redis.db}
                        onChange={(e) => setDbConfig({
                          ...dbConfig,
                          redis: { ...dbConfig.redis, db: parseInt(e.target.value) }
                        })}
                        className="w-full px-4 py-2 bg-dark-200 border border-dark-100 rounded-lg text-white"
                      />
                    </div>
                  </div>
                </div>
              )}
            </div>
          )}

          {/* 高级设置 */}
          {activeTab === 'advanced' && (
            <div className="space-y-6">
              <h2 className="text-xl font-semibold text-white">高级设置</h2>

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm text-gray-400 mb-2">最大会话数</label>
                  <input
                    type="number"
                    value={advancedConfig.maxSessions}
                    onChange={(e) => setAdvancedConfig({ ...advancedConfig, maxSessions: parseInt(e.target.value) })}
                    className="w-full px-4 py-2 bg-dark-200 border border-dark-100 rounded-lg text-white"
                  />
                </div>
                <div>
                  <label className="block text-sm text-gray-400 mb-2">会话超时 (秒)</label>
                  <input
                    type="number"
                    value={advancedConfig.sessionTimeout}
                    onChange={(e) => setAdvancedConfig({ ...advancedConfig, sessionTimeout: parseInt(e.target.value) })}
                    className="w-full px-4 py-2 bg-dark-200 border border-dark-100 rounded-lg text-white"
                  />
                </div>
                <div>
                  <label className="block text-sm text-gray-400 mb-2">消息缓冲区大小</label>
                  <input
                    type="number"
                    value={advancedConfig.messageBufferSize}
                    onChange={(e) => setAdvancedConfig({ ...advancedConfig, messageBufferSize: parseInt(e.target.value) })}
                    className="w-full px-4 py-2 bg-dark-200 border border-dark-100 rounded-lg text-white"
                  />
                </div>
                <div>
                  <label className="block text-sm text-gray-400 mb-2"> Worker Pool 大小</label>
                  <input
                    type="number"
                    value={advancedConfig.workerPoolSize}
                    onChange={(e) => setAdvancedConfig({ ...advancedConfig, workerPoolSize: parseInt(e.target.value) })}
                    className="w-full px-4 py-2 bg-dark-200 border border-dark-100 rounded-lg text-white"
                  />
                </div>
              </div>

              <div className="space-y-4">
                <div className="flex items-center justify-between py-3">
                  <div>
                    <p className="text-white">日志级别</p>
                    <select
                      value={advancedConfig.logLevel}
                      onChange={(e) => setAdvancedConfig({ ...advancedConfig, logLevel: e.target.value })}
                      className="mt-2 px-4 py-2 bg-dark-200 border border-dark-100 rounded-lg text-white"
                    >
                      <option value="debug">Debug</option>
                      <option value="info">Info</option>
                      <option value="warn">Warning</option>
                      <option value="error">Error</option>
                    </select>
                  </div>
                </div>

                <div className="flex items-center justify-between py-3">
                  <div>
                    <p className="text-white">启用指标收集</p>
                    <p className="text-sm text-gray-500">收集性能指标用于监控</p>
                  </div>
                  <button
                    onClick={() => setAdvancedConfig({ ...advancedConfig, enableMetrics: !advancedConfig.enableMetrics })}
                    className={clsx(
                      'w-12 h-6 rounded-full transition-colors relative',
                      advancedConfig.enableMetrics ? 'bg-tortoise-500' : 'bg-dark-300'
                    )}
                  >
                    <span
                      className={clsx(
                        'absolute top-1 w-4 h-4 bg-white rounded-full transition-transform',
                        advancedConfig.enableMetrics ? 'left-7' : 'left-1'
                      )}
                    />
                  </button>
                </div>

                <div className="flex items-center justify-between py-3">
                  <div>
                    <p className="text-white">启用链路追踪</p>
                    <p className="text-sm text-gray-500">OpenTelemetry 分布式追踪</p>
                  </div>
                  <button
                    onClick={() => setAdvancedConfig({ ...advancedConfig, enableTracing: !advancedConfig.enableTracing })}
                    className={clsx(
                      'w-12 h-6 rounded-full transition-colors relative',
                      advancedConfig.enableTracing ? 'bg-tortoise-500' : 'bg-dark-300'
                    )}
                  >
                    <span
                      className={clsx(
                        'absolute top-1 w-4 h-4 bg-white rounded-full transition-transform',
                        advancedConfig.enableTracing ? 'left-7' : 'left-1'
                      )}
                    />
                  </button>
                </div>
              </div>
            </div>
          )}

          {/* Save Button */}
          <div className="pt-6 border-t border-dark-100 flex justify-between items-center">
            <button className="text-gray-400 hover:text-white">
              重置为默认
            </button>
            
            {/* 状态提示 */}
            {saveStatus === 'success' && (
              <div className="flex items-center gap-2 text-green-400">
                <CheckCircle2 className="w-5 h-5" />
                <span>保存成功</span>
              </div>
            )}
            {saveStatus === 'error' && (
              <div className="flex items-center gap-2 text-red-400">
                <AlertCircle className="w-5 h-5" />
                <span>保存失败</span>
              </div>
            )}
            
            <button
              onClick={handleSave}
              disabled={isSaving}
              className={clsx(
                'btn btn-primary',
                isSaving && 'opacity-70 cursor-not-allowed'
              )}
            >
              {isSaving ? (
                <>
                  <RefreshCw className="w-5 h-5 animate-spin" />
                  保存中...
                </>
              ) : (
                <>
                  <Save className="w-5 h-5" />
                  保存所有设置
                </>
              )}
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}

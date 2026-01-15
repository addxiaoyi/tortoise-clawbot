import { useState, useEffect } from 'react'
import { useAppStore } from '../store/appStore'
import { api } from '../services/api'
import clsx from 'clsx'
import {
  MessageSquare,
  Brain,
  Clock,
  TrendingUp,
  Activity,
  Server,
  Bot,
  Zap,
  Users,
  MessageCircle,
  Database,
  Shield,
  CheckCircle2,
  XCircle,
  AlertCircle,
  RefreshCw,
} from 'lucide-react'

interface Stats {
  sessions: number
  messages: number
  memories: number
  plugins: number
  channels: number
}

interface AIStats {
  providers: string[]
  totalRequests: number
  totalTokens: number
  averageLatency: number
}

export default function Dashboard() {
  const { isConnected } = useAppStore()
  const [stats, setStats] = useState<Stats>({
    sessions: 0,
    messages: 0,
    memories: 0,
    plugins: 0,
    channels: 0,
  })
  const [aiStats, setAIStats] = useState<AIStats | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  // 获取统计数据
  const fetchStats = async () => {
    try {
      setIsLoading(true)
      setError(null)

      const [statsData, aiStatsData] = await Promise.all([
        api.getStats(),
        api.getAIStats().catch(() => null),
      ])

      setStats(statsData)
      setAIStats(aiStatsData)
    } catch (err) {
      console.error('Failed to fetch stats:', err)
      setError('无法获取统计数据')
    } finally {
      setIsLoading(false)
    }
  }

  useEffect(() => {
    if (isConnected) {
      fetchStats()
    }
  }, [isConnected])

  // 刷新数据
  const handleRefresh = () => {
    fetchStats()
  }

  // 统计卡片组件
  const StatCard = ({
    title,
    value,
    icon: Icon,
    color,
    subtitle,
  }: {
    title: string
    value: number | string
    icon: React.ElementType
    color: string
    subtitle?: string
  }) => (
    <div className="card p-6">
      <div className="flex items-start justify-between">
        <div>
          <p className="text-gray-400 text-sm mb-1">{title}</p>
          <p className="text-3xl font-bold text-white">{value}</p>
          {subtitle && (
            <p className="text-gray-500 text-xs mt-1">{subtitle}</p>
          )}
        </div>
        <div className={clsx('w-12 h-12 rounded-xl flex items-center justify-center', color)}>
          <Icon className="w-6 h-6 text-white" />
        </div>
      </div>
    </div>
  )

  return (
    <div className="space-y-6">
      {/* 头部 */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white">控制台</h1>
          <p className="text-gray-400 text-sm mt-1">
            实时监控 Tortoise AI 框架运行状态
          </p>
        </div>
        <button
          onClick={handleRefresh}
          disabled={isLoading}
          className="btn btn-secondary flex items-center gap-2"
        >
          <RefreshCw className={clsx('w-4 h-4', isLoading && 'animate-spin')} />
          刷新
        </button>
      </div>

      {/* 连接状态 */}
      {!isConnected && (
        <div className="card p-4 bg-yellow-500/10 border-yellow-500/30">
          <div className="flex items-center gap-3">
            <AlertCircle className="w-5 h-5 text-yellow-400" />
            <span className="text-yellow-400">后端服务未连接，请确保服务器正在运行</span>
          </div>
        </div>
      )}

      {/* 错误提示 */}
      {error && (
        <div className="card p-4 bg-red-500/10 border-red-500/30">
          <div className="flex items-center gap-3">
            <XCircle className="w-5 h-5 text-red-400" />
            <span className="text-red-400">{error}</span>
          </div>
        </div>
      )}

      {/* 统计卡片 */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <StatCard
          title="活跃会话"
          value={stats.sessions}
          icon={MessageSquare}
          color="bg-blue-500"
          subtitle="对话会话数"
        />
        <StatCard
          title="消息总数"
          value={stats.messages}
          icon={MessageCircle}
          color="bg-purple-500"
          subtitle="已处理消息"
        />
        <StatCard
          title="记忆存储"
          value={stats.memories}
          icon={Brain}
          color="bg-green-500"
          subtitle="长期记忆"
        />
        <StatCard
          title="插件数量"
          value={stats.plugins}
          icon={Zap}
          color="bg-orange-500"
          subtitle="已安装"
        />
      </div>

      {/* AI 状态和渠道状态 */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* AI 引擎状态 */}
        <div className="card p-6">
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-lg font-semibold text-white flex items-center gap-2">
              <Bot className="w-5 h-5" />
              AI 引擎状态
            </h2>
            <span className={clsx(
              'px-2 py-1 rounded text-xs',
              aiStats?.providers?.length ? 'bg-green-500/20 text-green-400' : 'bg-yellow-500/20 text-yellow-400'
            )}>
              {aiStats?.providers?.length ? '已配置' : '未配置'}
            </span>
          </div>

          {aiStats && (
            <div className="space-y-4">
              {/* 提供商列表 */}
              <div>
                <p className="text-sm text-gray-400 mb-2">已配置模型</p>
                <div className="flex flex-wrap gap-2">
                  {aiStats.providers?.map((provider) => (
                    <span
                      key={provider}
                      className="px-3 py-1 bg-blue-500/20 text-blue-400 rounded-lg text-sm"
                    >
                      {provider}
                    </span>
                  )) || (
                    <span className="text-gray-500 text-sm">无</span>
                  )}
                </div>
              </div>

              {/* 统计指标 */}
              <div className="grid grid-cols-3 gap-4">
                <div className="text-center p-3 bg-dark-100 rounded-lg">
                  <p className="text-2xl font-bold text-white">{aiStats.totalRequests || 0}</p>
                  <p className="text-xs text-gray-500">请求数</p>
                </div>
                <div className="text-center p-3 bg-dark-100 rounded-lg">
                  <p className="text-2xl font-bold text-white">
                    {aiStats.totalTokens ? Math.round(aiStats.totalTokens / 1000) + 'K' : '0'}
                  </p>
                  <p className="text-xs text-gray-500">Token</p>
                </div>
                <div className="text-center p-3 bg-dark-100 rounded-lg">
                  <p className="text-2xl font-bold text-white">
                    {aiStats.averageLatency ? aiStats.averageLatency.toFixed(0) + 'ms' : '0ms'}
                  </p>
                  <p className="text-xs text-gray-500">平均延迟</p>
                </div>
              </div>
            </div>
          )}

          {!aiStats && (
            <div className="text-center py-8 text-gray-500">
              <Activity className="w-12 h-12 mx-auto mb-3 opacity-50" />
              <p>暂无 AI 统计数据</p>
              <p className="text-sm mt-1">配置 API Key 后开始使用</p>
            </div>
          )}
        </div>

        {/* 渠道状态 */}
        <div className="card p-6">
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-lg font-semibold text-white flex items-center gap-2">
              <Server className="w-5 h-5" />
              消息渠道
            </h2>
            <span className="text-sm text-gray-400">{stats.channels} 个渠道</span>
          </div>

          <div className="space-y-3">
            {/* Telegram */}
            <div className="flex items-center justify-between p-3 bg-dark-100 rounded-lg">
              <div className="flex items-center gap-3">
                <div className="w-10 h-10 bg-blue-500/20 rounded-lg flex items-center justify-center">
                  <MessageSquare className="w-5 h-5 text-blue-400" />
                </div>
                <div>
                  <p className="text-white font-medium">Telegram</p>
                  <p className="text-xs text-gray-500">即时消息</p>
                </div>
              </div>
              <CheckCircle2 className="w-5 h-5 text-gray-500" />
            </div>

            {/* Discord */}
            <div className="flex items-center justify-between p-3 bg-dark-100 rounded-lg">
              <div className="flex items-center gap-3">
                <div className="w-10 h-10 bg-indigo-500/20 rounded-lg flex items-center justify-center">
                  <MessageSquare className="w-5 h-5 text-indigo-400" />
                </div>
                <div>
                  <p className="text-white font-medium">Discord</p>
                  <p className="text-xs text-gray-500">社区/服务器</p>
                </div>
              </div>
              <CheckCircle2 className="w-5 h-5 text-gray-500" />
            </div>

            {/* Slack */}
            <div className="flex items-center justify-between p-3 bg-dark-100 rounded-lg">
              <div className="flex items-center gap-3">
                <div className="w-10 h-10 bg-green-500/20 rounded-lg flex items-center justify-center">
                  <MessageSquare className="w-5 h-5 text-green-400" />
                </div>
                <div>
                  <p className="text-white font-medium">Slack</p>
                  <p className="text-xs text-gray-500">团队协作</p>
                </div>
              </div>
              <XCircle className="w-5 h-5 text-gray-500" />
            </div>

            {/* Web */}
            <div className="flex items-center justify-between p-3 bg-dark-100 rounded-lg">
              <div className="flex items-center gap-3">
                <div className="w-10 h-10 bg-purple-500/20 rounded-lg flex items-center justify-center">
                  <Server className="w-5 h-5 text-purple-400" />
                </div>
                <div>
                  <p className="text-white font-medium">Web UI</p>
                  <p className="text-xs text-gray-500">网页面板</p>
                </div>
              </div>
              <CheckCircle2 className="w-5 h-5 text-green-400" />
            </div>
          </div>
        </div>
      </div>

      {/* 系统状态 */}
      <div className="card p-6">
        <h2 className="text-lg font-semibold text-white flex items-center gap-2 mb-4">
          <Shield className="w-5 h-5" />
          系统状态
        </h2>

        <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
          {/* 数据库 */}
          <div className="p-4 bg-dark-100 rounded-lg">
            <div className="flex items-center gap-2 mb-2">
              <Database className="w-4 h-4 text-gray-400" />
              <span className="text-sm text-gray-400">数据库</span>
            </div>
            <div className="flex items-center gap-2">
              <CheckCircle2 className="w-4 h-4 text-green-400" />
              <span className="text-white text-sm">Redis</span>
            </div>
          </div>

          {/* 认证 */}
          <div className="p-4 bg-dark-100 rounded-lg">
            <div className="flex items-center gap-2 mb-2">
              <Shield className="w-4 h-4 text-gray-400" />
              <span className="text-sm text-gray-400">认证方式</span>
            </div>
            <div className="flex items-center gap-2">
              <CheckCircle2 className="w-4 h-4 text-green-400" />
              <span className="text-white text-sm">JWT + API Key</span>
            </div>
          </div>

          {/* 监控 */}
          <div className="p-4 bg-dark-100 rounded-lg">
            <div className="flex items-center gap-2 mb-2">
              <Activity className="w-4 h-4 text-gray-400" />
              <span className="text-sm text-gray-400">监控状态</span>
            </div>
            <div className="flex items-center gap-2">
              <CheckCircle2 className="w-4 h-4 text-green-400" />
              <span className="text-white text-sm">已启用</span>
            </div>
          </div>

          {/* 服务发现 */}
          <div className="p-4 bg-dark-100 rounded-lg">
            <div className="flex items-center gap-2 mb-2">
              <Users className="w-4 h-4 text-gray-400" />
              <span className="text-sm text-gray-400">服务发现</span>
            </div>
            <div className="flex items-center gap-2">
              <CheckCircle2 className="w-4 h-4 text-green-400" />
              <span className="text-white text-sm">mDNS/SSDP</span>
            </div>
          </div>
        </div>
      </div>

      {/* 快速操作 */}
      <div className="card p-6">
        <h2 className="text-lg font-semibold text-white flex items-center gap-2 mb-4">
          <Zap className="w-5 h-5" />
          快速操作
        </h2>

        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          <button className="p-4 bg-dark-100 hover:bg-dark-200 rounded-lg text-center transition-colors">
            <MessageSquare className="w-6 h-6 mx-auto mb-2 text-blue-400" />
            <span className="text-sm text-white">新建会话</span>
          </button>
          <button className="p-4 bg-dark-100 hover:bg-dark-200 rounded-lg text-center transition-colors">
            <Brain className="w-6 h-6 mx-auto mb-2 text-green-400" />
            <span className="text-sm text-white">添加记忆</span>
          </button>
          <button className="p-4 bg-dark-100 hover:bg-dark-200 rounded-lg text-center transition-colors">
            <Zap className="w-6 h-6 mx-auto mb-2 text-orange-400" />
            <span className="text-sm text-white">管理插件</span>
          </button>
          <button className="p-4 bg-dark-100 hover:bg-dark-200 rounded-lg text-center transition-colors">
            <Server className="w-6 h-6 mx-auto mb-2 text-purple-400" />
            <span className="text-sm text-white">配置渠道</span>
          </button>
        </div>
      </div>
    </div>
  )
}

import { useState, useEffect } from 'react'
import { useAppStore } from '../store/appStore'
import clsx from 'clsx'
import {
  Puzzle,
  Search,
  Plus,
  Trash2,
  Power,
  ExternalLink,
  Download,
  X,
  Check,
  AlertCircle,
  Loader2,
} from 'lucide-react'

export default function Plugins() {
  const { plugins, loadPlugins, togglePlugin } = useAppStore()
  const [searchQuery, setSearchQuery] = useState('')
  const [showInstallModal, setShowInstallModal] = useState(false)
  const [installUrl, setInstallUrl] = useState('')
  const [isInstalling, setIsInstalling] = useState(false)
  const [filter, setFilter] = useState<'all' | 'enabled' | 'disabled'>('all')

  useEffect(() => {
    loadPlugins()
  }, [loadPlugins])

  const filteredPlugins = plugins.filter((plugin) => {
    if (filter === 'enabled' && !plugin.enabled) return false
    if (filter === 'disabled' && plugin.enabled) return false
    if (searchQuery && !plugin.name.toLowerCase().includes(searchQuery.toLowerCase())) return false
    return true
  })

  const getStatusBadge = (plugin: typeof plugins[0]) => {
    if (plugin.status === 'error') {
      return <span className="badge badge-error">错误</span>
    }
    if (!plugin.enabled) {
      return <span className="badge badge-warning">已禁用</span>
    }
    return <span className="badge badge-success">运行中</span>
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white">插件管理</h1>
          <p className="text-gray-400 mt-1">扩展 Tortoise 的功能</p>
        </div>
        <button
          onClick={() => setShowInstallModal(true)}
          className="btn btn-primary"
        >
          <Download className="w-5 h-5" />
          安装插件
        </button>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-3 gap-4">
        <div className="card p-4">
          <p className="text-gray-400 text-sm">已安装</p>
          <p className="text-2xl font-bold text-white">{plugins.length}</p>
        </div>
        <div className="card p-4">
          <p className="text-gray-400 text-sm">已启用</p>
          <p className="text-2xl font-bold text-green-400">
            {plugins.filter((p) => p.enabled).length}
          </p>
        </div>
        <div className="card p-4">
          <p className="text-gray-400 text-sm">可用工具</p>
          <p className="text-2xl font-bold text-blue-400">
            {plugins.reduce((acc, p) => acc + p.tools.length, 0)}
          </p>
        </div>
      </div>

      {/* Filters */}
      <div className="flex items-center gap-4">
        {/* Search */}
        <div className="relative flex-1">
          <Search className="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-500" />
          <input
            type="text"
            placeholder="搜索插件..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="w-full pl-12 pr-4 py-3 bg-dark-200 border border-dark-100 rounded-xl text-white placeholder-gray-500"
          />
        </div>

        {/* Filter Tabs */}
        <div className="flex bg-dark-200 rounded-lg p-1">
          {(['all', 'enabled', 'disabled'] as const).map((f) => (
            <button
              key={f}
              onClick={() => setFilter(f)}
              className={clsx(
                'px-4 py-2 rounded-md text-sm font-medium transition-colors',
                filter === f
                  ? 'bg-tortoise-500 text-white'
                  : 'text-gray-400 hover:text-white'
              )}
            >
              {f === 'all' ? '全部' : f === 'enabled' ? '已启用' : '已禁用'}
            </button>
          ))}
        </div>
      </div>

      {/* Plugins Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        {filteredPlugins.length === 0 ? (
          <div className="col-span-2 card p-12 text-center">
            <Puzzle className="w-12 h-12 text-gray-600 mx-auto mb-4" />
            <p className="text-gray-400">暂无插件</p>
            <p className="text-sm text-gray-500 mt-1">安装插件来扩展功能</p>
          </div>
        ) : (
          filteredPlugins.map((plugin) => (
            <div key={plugin.id} className="card p-5">
              <div className="flex items-start justify-between mb-4">
                <div className="flex items-center gap-3">
                  <div className={clsx(
                    'w-12 h-12 rounded-xl flex items-center justify-center',
                    plugin.enabled ? 'bg-tortoise-500/20' : 'bg-dark-100'
                  )}>
                    <Puzzle className={clsx(
                      'w-6 h-6',
                      plugin.enabled ? 'text-tortoise-400' : 'text-gray-500'
                    )} />
                  </div>
                  <div>
                    <h3 className="text-white font-semibold">{plugin.name}</h3>
                    <p className="text-sm text-gray-500">v{plugin.version}</p>
                  </div>
                </div>
                {getStatusBadge(plugin)}
              </div>

              <p className="text-gray-400 text-sm mb-4 line-clamp-2">
                {plugin.description}
              </p>

              {/* Tools */}
              {plugin.tools.length > 0 && (
                <div className="mb-4">
                  <p className="text-xs text-gray-500 mb-2">
                    {plugin.tools.length} 个工具
                  </p>
                  <div className="flex flex-wrap gap-1">
                    {plugin.tools.slice(0, 3).map((tool) => (
                      <span
                        key={tool.name}
                        className="px-2 py-0.5 bg-dark-100 rounded text-xs text-gray-400"
                      >
                        {tool.name}
                      </span>
                    ))}
                    {plugin.tools.length > 3 && (
                      <span className="px-2 py-0.5 text-xs text-gray-500">
                        +{plugin.tools.length - 3} 更多
                      </span>
                    )}
                  </div>
                </div>
              )}

              <div className="flex items-center justify-between pt-4 border-t border-dark-100">
                <span className="text-xs text-gray-500">
                  by {plugin.author}
                </span>
                <div className="flex items-center gap-2">
                  <button
                    onClick={() => togglePlugin(plugin.id)}
                    className={clsx(
                      'p-2 rounded-lg transition-colors',
                      plugin.enabled
                        ? 'text-red-400 hover:bg-red-500/10'
                        : 'text-green-400 hover:bg-green-500/10'
                    )}
                  >
                    <Power className="w-4 h-4" />
                  </button>
                  <button className="p-2 text-gray-400 hover:text-red-400 hover:bg-red-500/10 rounded-lg transition-colors">
                    <Trash2 className="w-4 h-4" />
                  </button>
                </div>
              </div>
            </div>
          ))
        )}
      </div>

      {/* Install Modal */}
      {showInstallModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-dark-200 rounded-xl p-6 w-[500px] border border-dark-100">
            <div className="flex items-center justify-between mb-6">
              <h3 className="text-lg font-semibold text-white">安装插件</h3>
              <button
                onClick={() => setShowInstallModal(false)}
                className="p-1 text-gray-400 hover:text-white"
              >
                <X className="w-5 h-5" />
              </button>
            </div>

            <div className="mb-6">
              <label className="block text-sm text-gray-400 mb-2">
                插件 URL 或 NPM 包名
              </label>
              <input
                type="text"
                value={installUrl}
                onChange={(e) => setInstallUrl(e.target.value)}
                placeholder="例如: @tortoise/plugin-example 或 https://..."
                className="w-full px-4 py-3 bg-dark-300 border border-dark-100 rounded-lg text-white placeholder-gray-500"
              />
            </div>

            <div className="bg-dark-300 rounded-lg p-4 mb-6">
              <p className="text-sm text-gray-400 mb-2">安装来源:</p>
              <ul className="space-y-1 text-sm text-gray-300">
                <li>• NPM 包: @tortoise/plugin-*</li>
                <li>• GitHub: https://github.com/... </li>
                <li>• 本地: ./path/to/plugin</li>
              </ul>
            </div>

            <div className="flex gap-3">
              <button
                onClick={() => setShowInstallModal(false)}
                className="flex-1 px-4 py-2 bg-dark-300 text-gray-300 rounded-lg hover:bg-dark-100 transition-colors"
              >
                取消
              </button>
              <button
                onClick={async () => {
                  if (!installUrl.trim()) return
                  setIsInstalling(true)
                  // Simulate install
                  await new Promise((r) => setTimeout(r, 1500))
                  setIsInstalling(false)
                  setShowInstallModal(false)
                  setInstallUrl('')
                }}
                disabled={!installUrl.trim() || isInstalling}
                className={clsx(
                  'flex-1 px-4 py-2 rounded-lg transition-colors flex items-center justify-center gap-2',
                  installUrl.trim() && !isInstalling
                    ? 'bg-tortoise-500 text-white hover:bg-tortoise-600'
                    : 'bg-dark-300 text-gray-500 cursor-not-allowed'
                )}
              >
                {isInstalling ? (
                  <>
                    <Loader2 className="w-4 h-4 animate-spin" />
                    安装中...
                  </>
                ) : (
                  <>
                    <Download className="w-4 h-4" />
                    安装
                  </>
                )}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

import { useAppStore } from '../store/appStore'
import { format } from 'date-fns'
import { zhCN } from 'date-fns/locale'
import {
  Wifi,
  WifiOff,
  RefreshCw,
  Bell,
  User,
  Search,
} from 'lucide-react'
import { useState } from 'react'

export default function Header() {
  const { isConnected, isConnecting, serverUrl, sessions, currentSession } = useAppStore()
  const [searchQuery, setSearchQuery] = useState('')

  return (
    <header className="h-16 bg-dark-300 border-b border-dark-100 flex items-center justify-between px-6">
      {/* Left - Search */}
      <div className="flex items-center gap-4">
        <div className="relative">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-500" />
          <input
            type="text"
            placeholder="搜索..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="w-64 pl-10 pr-4 py-2 bg-dark-200 border border-dark-100 rounded-lg text-sm text-gray-200 placeholder-gray-500 focus:border-tortoise-500 transition-colors"
          />
        </div>
      </div>

      {/* Center - Current Session */}
      {currentSession && (
        <div className="flex items-center gap-2 text-sm text-gray-400">
          <span>当前对话:</span>
          <span className="text-white font-medium">{currentSession.name}</span>
        </div>
      )}

      {/* Right - Status & Actions */}
      <div className="flex items-center gap-4">
        {/* Connection Status */}
        <div className="flex items-center gap-2">
          {isConnecting ? (
            <RefreshCw className="w-4 h-4 text-yellow-500 animate-spin" />
          ) : isConnected ? (
            <Wifi className="w-4 h-4 text-green-500" />
          ) : (
            <WifiOff className="w-4 h-4 text-red-500" />
          )}
          <span className="text-sm text-gray-400">
            {isConnecting ? '连接中...' : isConnected ? '已连接' : '未连接'}
          </span>
        </div>

        {/* Server URL */}
        <div className="text-xs text-gray-500 font-mono">
          {serverUrl}
        </div>

        {/* Time */}
        <div className="text-sm text-gray-500">
          {format(new Date(), 'HH:mm', { locale: zhCN })}
        </div>

        {/* Notifications */}
        <button className="relative p-2 text-gray-400 hover:text-white transition-colors">
          <Bell className="w-5 h-5" />
          <span className="absolute top-1 right-1 w-2 h-2 bg-red-500 rounded-full" />
        </button>

        {/* User */}
        <button className="flex items-center gap-2 p-2 text-gray-400 hover:text-white transition-colors">
          <div className="w-8 h-8 bg-tortoise-500/20 rounded-full flex items-center justify-center">
            <User className="w-4 h-4 text-tortoise-400" />
          </div>
        </button>
      </div>
    </header>
  )
}

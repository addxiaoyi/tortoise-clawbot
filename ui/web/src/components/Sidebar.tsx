import { NavLink, useLocation } from 'react-router-dom'
import { useAppStore } from '../store/appStore'
import clsx from 'clsx'
import {
  LayoutDashboard,
  MessageSquare,
  Brain,
  Puzzle,
  Settings,
  ChevronLeft,
  ChevronRight,
  Bot,
} from 'lucide-react'

const navItems = [
  { path: '/dashboard', icon: LayoutDashboard, label: '仪表盘' },
  { path: '/chat', icon: MessageSquare, label: '对话' },
  { path: '/memory', icon: Brain, label: '记忆' },
  { path: '/plugins', icon: Puzzle, label: '插件' },
  { path: '/settings', icon: Settings, label: '设置' },
]

export default function Sidebar() {
  const { sidebarCollapsed, toggleSidebar, isConnected } = useAppStore()
  const location = useLocation()

  return (
    <aside
      className={clsx(
        'fixed left-0 top-0 h-screen bg-dark-300 border-r border-dark-100 transition-all duration-300 z-50 flex flex-col',
        sidebarCollapsed ? 'w-16' : 'w-64'
      )}
    >
      {/* Logo */}
      <div className="h-16 flex items-center px-4 border-b border-dark-100">
        <div className="flex items-center gap-3">
          <div className="relative">
            <Bot className="w-8 h-8 text-tortoise-500" />
            {isConnected && (
              <span className="absolute -top-1 -right-1 w-3 h-3 bg-green-500 rounded-full border-2 border-dark-300" />
            )}
          </div>
          {!sidebarCollapsed && (
            <div>
              <h1 className="text-lg font-bold gradient-text">Tortoise</h1>
              <p className="text-xs text-gray-500">AI Agent Framework</p>
            </div>
          )}
        </div>
      </div>

      {/* Navigation */}
      <nav className="flex-1 py-4">
        {navItems.map((item) => {
          const isActive = location.pathname.startsWith(item.path)
          const Icon = item.icon
          
          return (
            <NavLink
              key={item.path}
              to={item.path}
              className={clsx(
                'flex items-center gap-3 px-4 py-3 mx-2 rounded-lg transition-all duration-200',
                isActive
                  ? 'bg-tortoise-500/20 text-tortoise-400'
                  : 'text-gray-400 hover:bg-dark-200 hover:text-white'
              )}
            >
              <Icon className="w-5 h-5 flex-shrink-0" />
              {!sidebarCollapsed && (
                <span className="font-medium">{item.label}</span>
              )}
            </NavLink>
          )
        })}
      </nav>

      {/* Collapse Toggle */}
      <div className="p-4 border-t border-dark-100">
        <button
          onClick={toggleSidebar}
          className="w-full flex items-center justify-center gap-2 py-2 text-gray-400 hover:text-white transition-colors"
        >
          {sidebarCollapsed ? (
            <ChevronRight className="w-5 h-5" />
          ) : (
            <>
              <ChevronLeft className="w-5 h-5" />
              <span className="text-sm">收起</span>
            </>
          )}
        </button>
      </div>
    </aside>
  )
}

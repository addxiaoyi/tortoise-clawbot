import { useState, useEffect } from 'react'
import { Link, useLocation } from 'react-router-dom'
import {
  Menu,
  X,
  Home,
  MessageSquare,
  Settings,
  Database,
  ChevronRight,
} from 'lucide-react'
import clsx from 'clsx'

interface MobileLayoutProps {
  children: React.ReactNode
}

export default function MobileLayout({ children }: MobileLayoutProps) {
  const [isMenuOpen, setIsMenuOpen] = useState(false)
  const location = useLocation()

  // 移动端检测
  const isMobile = typeof window !== 'undefined' && window.innerWidth < 768

  useEffect(() => {
    // 移动端下自动关闭菜单
    if (isMobile && isMenuOpen) {
      setIsMenuOpen(false)
    }
  }, [location.pathname])

  const navItems = [
    { path: '/dashboard', icon: Home, label: '首页' },
    { path: '/chat', icon: MessageSquare, label: '对话' },
    { path: '/memory', icon: Database, label: '记忆' },
    { path: '/settings', icon: Settings, label: '设置' },
  ]

  return (
    <div className="min-h-screen bg-dark-300 flex flex-col">
      {/* 顶部导航 */}
      <header className="sticky top-0 z-50 bg-dark-100 border-b border-dark-200">
        <div className="flex items-center justify-between h-14 px-4">
          <button
            onClick={() => setIsMenuOpen(!isMenuOpen)}
            className="p-2 -ml-2 hover:bg-dark-200 rounded-lg"
          >
            {isMenuOpen ? (
              <X className="w-6 h-6" />
            ) : (
              <Menu className="w-6 h-6" />
            )}
          </button>

          <h1 className="text-lg font-semibold text-white">Tortoise</h1>

          <div className="w-10" /> {/* 占位 */}
        </div>
      </header>

      {/* 侧边菜单 */}
      <div
        className={clsx(
          'fixed inset-0 z-40 bg-black/50 transition-opacity lg:hidden',
          isMenuOpen ? 'opacity-100' : 'opacity-0 pointer-events-none'
        )}
        onClick={() => setIsMenuOpen(false)}
      />

      <nav
        className={clsx(
          'fixed top-14 left-0 bottom-0 w-64 z-50 bg-dark-100 border-r border-dark-200 transition-transform lg:hidden',
          isMenuOpen ? 'translate-x-0' : '-translate-x-full'
        )}
      >
        <div className="p-4 space-y-1">
          {navItems.map(item => {
            const Icon = item.icon
            const isActive = location.pathname === item.path

            return (
              <Link
                key={item.path}
                to={item.path}
                className={clsx(
                  'flex items-center gap-3 px-4 py-3 rounded-lg transition-colors',
                  isActive
                    ? 'bg-tortoise-500/20 text-tortoise-400'
                    : 'text-gray-400 hover:bg-dark-200 hover:text-white'
                )}
              >
                <Icon className="w-5 h-5" />
                <span className="font-medium">{item.label}</span>
                {isActive && <ChevronRight className="w-4 h-4 ml-auto" />}
              </Link>
            )
          })}
        </div>
      </nav>

      {/* 主内容 */}
      <main className="flex-1 overflow-hidden">
        {children}
      </main>

      {/* 移动端底部导航 */}
      {isMobile && (
        <nav className="fixed bottom-0 left-0 right-0 z-40 bg-dark-100 border-t border-dark-200 safe-area-inset-bottom">
          <div className="flex items-center justify-around h-16">
            {navItems.map(item => {
              const Icon = item.icon
              const isActive = location.pathname === item.path

              return (
                <Link
                  key={item.path}
                  to={item.path}
                  className={clsx(
                    'flex flex-col items-center gap-1 px-4 py-2 rounded-lg transition-colors',
                    isActive
                      ? 'text-tortoise-400'
                      : 'text-gray-500'
                  )}
                >
                  <Icon className="w-5 h-5" />
                  <span className="text-xs">{item.label}</span>
                </Link>
              )
            })}
          </div>
        </nav>
      )}
    </div>
  )
}

// Hook: 检测移动端
export function useIsMobile(breakpoint = 768) {
  const [isMobile, setIsMobile] = useState(false)

  useEffect(() => {
    const checkMobile = () => {
      setIsMobile(window.innerWidth < breakpoint)
    }

    checkMobile()
    window.addEventListener('resize', checkMobile)
    return () => window.removeEventListener('resize', checkMobile)
  }, [breakpoint])

  return isMobile
}

// Hook: 检测触摸设备
export function useIsTouchDevice() {
  const [isTouch, setIsTouch] = useState(false)

  useEffect(() => {
    setIsTouch('ontouchstart' in window || navigator.maxTouchPoints > 0)
  }, [])

  return isTouch
}

// 移动端优化的按钮
interface MobileButtonProps {
  children: React.ReactNode
  onClick?: () => void
  className?: string
  disabled?: boolean
  variant?: 'primary' | 'secondary' | 'ghost'
}

export function MobileButton({
  children,
  onClick,
  className,
  disabled,
  variant = 'primary',
}: MobileButtonProps) {
  const variants = {
    primary: 'bg-tortoise-500 hover:bg-tortoise-600 text-white',
    secondary: 'bg-dark-200 hover:bg-dark-300 text-white',
    ghost: 'hover:bg-dark-200 text-gray-400 hover:text-white',
  }

  return (
    <button
      onClick={onClick}
      disabled={disabled}
      className={clsx(
        'min-h-[48px] px-4 py-2 rounded-lg font-medium transition-colors active:scale-95',
        variants[variant],
        disabled && 'opacity-50 cursor-not-allowed',
        className
      )}
    >
      {children}
    </button>
  )
}

// 移动端优化的输入框
interface MobileInputProps {
  value: string
  onChange: (value: string) => void
  placeholder?: string
  type?: 'text' | 'password' | 'email'
  className?: string
  autoFocus?: boolean
}

export function MobileInput({
  value,
  onChange,
  placeholder,
  type = 'text',
  className,
  autoFocus,
}: MobileInputProps) {
  return (
    <input
      type={type}
      value={value}
      onChange={e => onChange(e.target.value)}
      placeholder={placeholder}
      autoFocus={autoFocus}
      className={clsx(
        'w-full min-h-[48px] px-4 py-3 bg-dark-200 border border-dark-100 rounded-lg text-white placeholder-gray-500',
        'focus:outline-none focus:border-tortoise-500 focus:ring-1 focus:ring-tortoise-500',
        'text-base', // 防止 iOS 缩放
        className
      )}
    />
  )
}

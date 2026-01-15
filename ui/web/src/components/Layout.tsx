import { Outlet } from 'react-router-dom'
import Sidebar from './Sidebar'
import Header from './Header'
import { useAppStore } from '../store/appStore'
import clsx from 'clsx'

export default function Layout() {
  const { sidebarCollapsed } = useAppStore()

  return (
    <div className="flex h-screen bg-dark-500">
      {/* Sidebar */}
      <Sidebar />
      
      {/* Main Content */}
      <div className={clsx(
        "flex-1 flex flex-col transition-all duration-300",
        sidebarCollapsed ? "ml-16" : "ml-64"
      )}>
        {/* Header */}
        <Header />
        
        {/* Page Content */}
        <main className="flex-1 overflow-auto p-6">
          <Outlet />
        </main>
      </div>
    </div>
  )
}

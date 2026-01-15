import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { useEffect } from 'react'
import Layout from './components/Layout'
import Dashboard from './pages/Dashboard'
import Chat from './pages/Chat'
import Memory from './pages/Memory'
import Plugins from './pages/Plugins'
import Settings from './pages/Settings'
import { useAppStore } from './store/appStore'

function App() {
  const { initializeApp, isConnected } = useAppStore()

  useEffect(() => {
    initializeApp()
  }, [initializeApp])

  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Layout />}>
          <Route index element={<Navigate to="/dashboard" replace />} />
          <Route path="dashboard" element={<Dashboard />} />
          <Route path="chat" element={<Chat />} />
          <Route path="chat/:sessionId" element={<Chat />} />
          <Route path="memory" element={<Memory />} />
          <Route path="plugins" element={<Plugins />} />
          <Route path="settings" element={<Settings />} />
        </Route>
      </Routes>
    </BrowserRouter>
  )
}

export default App

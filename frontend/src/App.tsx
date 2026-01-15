import { Routes, Route, Navigate } from 'react-router-dom'
import Layout from '@/components/Layout'
import Dashboard from '@/pages/Dashboard'
import Accounts from '@/pages/Accounts'
import Exams from '@/pages/Exams'
import Scheduler from '@/pages/Scheduler'
import Questions from '@/pages/Questions'
import Settings from '@/pages/Settings'
import Login from '@/pages/Login'

function App() {
  const token = localStorage.getItem('token')

  return (
    <Routes>
      <Route path="/login" element={<Login />} />
      <Route
        path="/*"
        element={
          token ? (
            <Layout>
              <Routes>
                <Route path="/" element={<Dashboard />} />
                <Route path="/accounts" element={<Accounts />} />
                <Route path="/exams" element={<Exams />} />
                <Route path="/scheduler" element={<Scheduler />} />
                <Route path="/questions" element={<Questions />} />
                <Route path="/settings" element={<Settings />} />
              </Routes>
            </Layout>
          ) : (
            <Navigate to="/login" replace />
          )
        }
      />
    </Routes>
  )
}

export default App

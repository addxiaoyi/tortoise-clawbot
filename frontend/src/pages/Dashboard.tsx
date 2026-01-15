import { useQuery } from '@tanstack/react-query'
import { 
  TrendingUp, 
  Users, 
  FileQuestion, 
  CheckCircle, 
  XCircle, 
  Clock,
  Activity
} from 'lucide-react'
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, PieChart, Pie, Cell } from 'recharts'
import api from '@/lib/api'
import type { Stats } from '@/types'

const mockStats: Stats = {
  totalExams: 156,
  completedExams: 142,
  failedExams: 14,
  successRate: 91.0,
  avgScore: 87.5,
  avgDuration: 25,
  platformStats: [
    { platform: 'chaoxing', count: 45, successRate: 93.3 },
    { platform: 'icve', count: 38, successRate: 89.5 },
    { platform: 'yuketang', count: 32, successRate: 96.9 },
    { platform: 'mooc', count: 25, successRate: 88.0 },
    { platform: 'generic', count: 16, successRate: 87.5 },
  ],
  dailyStats: [
    { date: '05-01', completed: 12, failed: 1, avgScore: 85 },
    { date: '05-02', completed: 15, failed: 2, avgScore: 88 },
    { date: '05-03', completed: 18, failed: 1, avgScore: 90 },
    { date: '05-04', completed: 14, failed: 3, avgScore: 86 },
    { date: '05-05', completed: 20, failed: 2, avgScore: 89 },
    { date: '05-06', completed: 22, failed: 1, avgScore: 91 },
    { date: '05-07', completed: 18, failed: 2, avgScore: 88 },
  ],
}

const COLORS = ['#0ea5e9', '#22c55e', '#f59e0b', '#8b5cf6', '#ec4899']

export default function Dashboard() {
  const stats = mockStats

  const statCards = [
    { label: '总考试数', value: stats.totalExams, icon: FileQuestion, color: 'text-primary-500' },
    { label: '已完成', value: stats.completedExams, icon: CheckCircle, color: 'text-tortoise-500' },
    { label: '失败', value: stats.failedExams, icon: XCircle, color: 'text-red-500' },
    { label: '平均分', value: `${stats.avgScore}%`, icon: TrendingUp, color: 'text-amber-500' },
  ]

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-800 dark:text-white">仪表盘</h1>
          <p className="text-gray-500 dark:text-gray-400 mt-1">系统运行概览</p>
        </div>
        <div className="flex items-center gap-2 px-3 py-1.5 bg-tortoise-50 dark:bg-tortoise-900/30 rounded-full">
          <Activity size={16} className="text-tortoise-500" />
          <span className="text-sm font-medium text-tortoise-700 dark:text-tortoise-300">系统正常</span>
        </div>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        {statCards.map((stat, i) => {
          const Icon = stat.icon
          return (
            <div key={i} className="bg-white dark:bg-gray-800 rounded-xl p-6 shadow-sm border border-gray-100 dark:border-gray-700">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm text-gray-500 dark:text-gray-400">{stat.label}</p>
                  <p className="text-3xl font-bold text-gray-800 dark:text-white mt-1">{stat.value}</p>
                </div>
                <div className={`p-3 rounded-lg bg-gray-50 dark:bg-gray-700 ${stat.color}`}>
                  <Icon size={24} />
                </div>
              </div>
            </div>
          )
        })}
      </div>

      {/* Charts */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Line Chart */}
        <div className="lg:col-span-2 bg-white dark:bg-gray-800 rounded-xl p-6 shadow-sm border border-gray-100 dark:border-gray-700">
          <h3 className="text-lg font-semibold text-gray-800 dark:text-white mb-4">考试趋势</h3>
          <div className="h-64">
            <ResponsiveContainer width="100%" height="100%">
              <LineChart data={stats.dailyStats}>
                <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
                <XAxis dataKey="date" stroke="#9ca3af" fontSize={12} />
                <YAxis stroke="#9ca3af" fontSize={12} />
                <Tooltip 
                  contentStyle={{ 
                    backgroundColor: '#1f2937', 
                    border: 'none', 
                    borderRadius: '8px',
                    color: '#fff'
                  }}
                />
                <Line 
                  type="monotone" 
                  dataKey="completed" 
                  stroke="#22c55e" 
                  strokeWidth={2}
                  dot={{ fill: '#22c55e', strokeWidth: 2 }}
                  name="完成"
                />
                <Line 
                  type="monotone" 
                  dataKey="failed" 
                  stroke="#ef4444" 
                  strokeWidth={2}
                  dot={{ fill: '#ef4444', strokeWidth: 2 }}
                  name="失败"
                />
              </LineChart>
            </ResponsiveContainer>
          </div>
        </div>

        {/* Pie Chart */}
        <div className="bg-white dark:bg-gray-800 rounded-xl p-6 shadow-sm border border-gray-100 dark:border-gray-700">
          <h3 className="text-lg font-semibold text-gray-800 dark:text-white mb-4">平台分布</h3>
          <div className="h-64">
            <ResponsiveContainer width="100%" height="100%">
              <PieChart>
                <Pie
                  data={stats.platformStats}
                  cx="50%"
                  cy="50%"
                  innerRadius={50}
                  outerRadius={80}
                  paddingAngle={5}
                  dataKey="count"
                  nameKey="platform"
                  label={({ platform }) => {
                    const names: Record<string, string> = {
                      chaoxing: '学习通',
                      icve: '智慧职教',
                      yuketang: '雨课堂',
                      mooc: 'MOOC',
                      generic: '通用',
                    }
                    return names[platform] || platform
                  }}
                >
                  {stats.platformStats.map((_, index) => (
                    <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                  ))}
                </Pie>
                <Tooltip />
              </PieChart>
            </ResponsiveContainer>
          </div>
        </div>
      </div>

      {/* Recent Activity */}
      <div className="bg-white dark:bg-gray-800 rounded-xl p-6 shadow-sm border border-gray-100 dark:border-gray-700">
        <h3 className="text-lg font-semibold text-gray-800 dark:text-white mb-4">最近活动</h3>
        <div className="space-y-4">
          {[
            { time: '10:32', action: '考试完成', detail: '学习通 - 高等数学期末考试', score: 92, status: 'success' },
            { time: '09:45', action: '考试开始', detail: '智慧职教 - 计算机网络', status: 'running' },
            { time: '09:12', action: '账号登录', detail: '雨课堂账号: student@edu.cn', status: 'success' },
            { time: '08:30', action: '定时任务', detail: '每日自动考试已执行', status: 'success' },
          ].map((item, i) => (
            <div key={i} className="flex items-center gap-4 p-3 rounded-lg bg-gray-50 dark:bg-gray-700/50">
              <div className={`w-2 h-2 rounded-full ${item.status === 'success' ? 'bg-tortoise-500' : item.status === 'running' ? 'bg-primary-500 animate-pulse' : 'bg-red-500'}`} />
              <div className="flex-1">
                <p className="text-sm font-medium text-gray-800 dark:text-white">{item.action}</p>
                <p className="text-xs text-gray-500 dark:text-gray-400">{item.detail}</p>
              </div>
              <div className="text-right">
                <p className="text-xs text-gray-400 dark:text-gray-500">{item.time}</p>
                {item.score && <p className="text-sm font-medium text-tortoise-600 dark:text-tortoise-400">{item.score}分</p>}
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}

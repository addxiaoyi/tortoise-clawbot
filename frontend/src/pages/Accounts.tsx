import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Plus, Search, Edit2, Trash2, RefreshCw, MoreVertical, CheckCircle, XCircle } from 'lucide-react'
import { cn } from '@/lib/utils'
import { formatDate } from '@/lib/utils'
import type { Account, Platform, PlatformNames } from '@/types'
import api from '@/lib/api'

const mockAccounts: Account[] = [
  { id: '1', username: 'student001@edu.cn', password: '******', platform: 'chaoxing', status: 'active', examCount: 45, successRate: 92.5, createdAt: '2024-01-15' },
  { id: '2', username: 'student002@edu.cn', password: '******', platform: 'icve', status: 'active', examCount: 32, successRate: 88.2, createdAt: '2024-02-20' },
  { id: '3', username: 'student003@edu.cn', password: '******', platform: 'yuketang', status: 'active', examCount: 28, successRate: 95.7, createdAt: '2024-03-10' },
  { id: '4', username: 'student004@edu.cn', password: '******', platform: 'mooc', status: 'inactive', examCount: 15, successRate: 80.0, createdAt: '2024-04-05' },
  { id: '5', username: 'student005@edu.cn', password: '******', platform: 'zhihuishu', status: 'error', examCount: 8, successRate: 75.0, createdAt: '2024-04-18' },
]

const platformNames: Record<Platform, string> = {
  chaoxing: '学习通',
  icve: '智慧职教',
  yuketang: '雨课堂',
  edx: '学堂在线',
  mooc: '中国大学MOOC',
  zhihuishu: '智慧树',
  tencentschool: '腾讯课堂',
  dingtalk: '钉钉云课堂',
  generic: '通用平台',
}

export default function Accounts() {
  const [search, setSearch] = useState('')
  const [platformFilter, setPlatformFilter] = useState<Platform | 'all'>('all')
  const [showAddModal, setShowAddModal] = useState(false)

  const filteredAccounts = mockAccounts.filter(account => {
    const matchesSearch = account.username.toLowerCase().includes(search.toLowerCase())
    const matchesPlatform = platformFilter === 'all' || account.platform === platformFilter
    return matchesSearch && matchesPlatform
  })

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-gray-800 dark:text-white">账号管理</h1>
          <p className="text-gray-500 dark:text-gray-400 mt-1">管理考试平台账号</p>
        </div>
        <button
          onClick={() => setShowAddModal(true)}
          className="px-4 py-2 bg-primary-500 hover:bg-primary-600 text-white rounded-lg font-medium flex items-center gap-2 transition-colors shadow-lg shadow-primary-500/30"
        >
          <Plus size={20} />
          添加账号
        </button>
      </div>

      {/* Filters */}
      <div className="flex flex-col sm:flex-row gap-4">
        <div className="relative flex-1">
          <Search size={20} className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
          <input
            type="text"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="搜索账号..."
            className="w-full pl-10 pr-4 py-2.5 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent transition-colors text-gray-800 dark:text-white"
          />
        </div>
        <select
          value={platformFilter}
          onChange={(e) => setPlatformFilter(e.target.value as Platform | 'all')}
          className="px-4 py-2.5 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent transition-colors text-gray-800 dark:text-white"
        >
          <option value="all">全部平台</option>
          {Object.entries(platformNames).map(([value, label]) => (
            <option key={value} value={value}>{label}</option>
          ))}
        </select>
      </div>

      {/* Table */}
      <div className="bg-white dark:bg-gray-800 rounded-xl shadow-sm border border-gray-100 dark:border-gray-700 overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead className="bg-gray-50 dark:bg-gray-700/50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">账号</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">平台</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">状态</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">考试数</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">成功率</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">添加时间</th>
                <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">操作</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100 dark:divide-gray-700">
              {filteredAccounts.map((account) => (
                <tr key={account.id} className="hover:bg-gray-50 dark:hover:bg-gray-700/50 transition-colors">
                  <td className="px-6 py-4 whitespace-nowrap">
                    <p className="text-sm font-medium text-gray-800 dark:text-white">{account.username}</p>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <span className="px-2.5 py-1 bg-primary-50 dark:bg-primary-900/30 text-primary-700 dark:text-primary-300 text-xs font-medium rounded-full">
                      {platformNames[account.platform]}
                    </span>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <span className={cn(
                      "px-2.5 py-1 text-xs font-medium rounded-full",
                      account.status === 'active' && "bg-tortoise-50 dark:bg-tortoise-900/30 text-tortoise-700 dark:text-tortoise-300",
                      account.status === 'inactive' && "bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-400",
                      account.status === 'error' && "bg-red-50 dark:bg-red-900/30 text-red-700 dark:text-red-300"
                    )}>
                      {account.status === 'active' ? '正常' : account.status === 'inactive' ? '停用' : '错误'}
                    </span>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-600 dark:text-gray-400">
                    {account.examCount}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="flex items-center gap-2">
                      <div className="w-16 h-1.5 bg-gray-200 dark:bg-gray-700 rounded-full overflow-hidden">
                        <div 
                          className="h-full bg-tortoise-500 rounded-full"
                          style={{ width: `${account.successRate}%` }}
                        />
                      </div>
                      <span className="text-sm text-gray-600 dark:text-gray-400">{account.successRate}%</span>
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
                    {account.createdAt}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-right">
                    <div className="flex items-center justify-end gap-2">
                      <button className="p-1.5 text-gray-400 hover:text-primary-500 transition-colors" title="编辑">
                        <Edit2 size={16} />
                      </button>
                      <button className="p-1.5 text-gray-400 hover:text-tortoise-500 transition-colors" title="测试登录">
                        <RefreshCw size={16} />
                      </button>
                      <button className="p-1.5 text-gray-400 hover:text-red-500 transition-colors" title="删除">
                        <Trash2 size={16} />
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  )
}

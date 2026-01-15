import { useState, useEffect } from 'react'
import { api } from '../services/api'
import clsx from 'clsx'
import {
  Brain,
  Plus,
  Search,
  Trash2,
  Edit3,
  Tag,
  Filter,
  SortAsc,
  Clock,
  Star,
  Sparkles,
  X,
  AlertCircle,
  CheckCircle2,
  Loader2,
} from 'lucide-react'

interface Memory {
  id: string
  type: 'working' | 'semantic' | 'episodic'
  content: string
  importance: number
  tags: string[]
  createdAt: string
  updatedAt: string
}

type MemoryFilter = 'all' | 'working' | 'semantic' | 'episodic'
type SortBy = 'date' | 'importance' | 'name'

export default function Memory() {
  const [memories, setMemories] = useState<Memory[]>([])
  const [filteredMemories, setFilteredMemories] = useState<Memory[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  
  // 过滤器状态
  const [filter, setFilter] = useState<MemoryFilter>('all')
  const [sortBy, setSortBy] = useState<SortBy>('date')
  const [searchQuery, setSearchQuery] = useState('')
  
  // 新建记忆
  const [showNewMemory, setShowNewMemory] = useState(false)
  const [newMemory, setNewMemory] = useState({
    type: 'semantic' as Memory['type'],
    content: '',
    importance: 5,
    tags: '',
  })
  const [isCreating, setIsCreating] = useState(false)
  
  // 编辑记忆
  const [editingMemory, setEditingMemory] = useState<Memory | null>(null)
  
  // 加载记忆
  useEffect(() => {
    loadMemories()
  }, [])

  // 过滤和排序
  useEffect(() => {
    let result = [...memories]
    
    // 过滤
    if (filter !== 'all') {
      result = result.filter(m => m.type === filter)
    }
    
    // 搜索
    if (searchQuery) {
      const query = searchQuery.toLowerCase()
      result = result.filter(m => 
        m.content.toLowerCase().includes(query) ||
        m.tags.some(tag => tag.toLowerCase().includes(query))
      )
    }
    
    // 排序
    result.sort((a, b) => {
      switch (sortBy) {
        case 'importance':
          return b.importance - a.importance
        case 'name':
          return a.content.localeCompare(b.content)
        case 'date':
        default:
          return new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime()
      }
    })
    
    setFilteredMemories(result)
  }, [memories, filter, sortBy, searchQuery])

  // 加载记忆
  const loadMemories = async () => {
    try {
      setIsLoading(true)
      setError(null)
      const data = await api.getMemories()
      setMemories(data.map((m: any) => ({
        ...m,
        createdAt: new Date(m.createdAt).toISOString(),
        updatedAt: new Date(m.updatedAt).toISOString(),
      })))
    } catch (err) {
      console.error('Failed to load memories:', err)
      setError('加载记忆失败')
    } finally {
      setIsLoading(false)
    }
  }

  // 创建记忆
  const handleCreate = async () => {
    if (!newMemory.content.trim()) return
    
    try {
      setIsCreating(true)
      const tags = newMemory.tags
        .split(',')
        .map(t => t.trim())
        .filter(t => t.length > 0)
      
      await api.addMemory(
        newMemory.type,
        newMemory.content,
        newMemory.importance
      )
      
      // 重新加载
      await loadMemories()
      
      // 重置表单
      setNewMemory({
        type: 'semantic',
        content: '',
        importance: 5,
        tags: '',
      })
      setShowNewMemory(false)
    } catch (err) {
      console.error('Failed to create memory:', err)
      setError('创建记忆失败')
    } finally {
      setIsCreating(false)
    }
  }

  // 删除记忆
  const handleDelete = async (id: string) => {
    try {
      await api.deleteMemory(id)
      setMemories(prev => prev.filter(m => m.id !== id))
    } catch (err) {
      console.error('Failed to delete memory:', err)
      setError('删除记忆失败')
    }
  }

  // 获取记忆类型标签
  const getTypeBadge = (type_: Memory['type']) => {
    const config = {
      working: { color: 'bg-yellow-500/20 text-yellow-400', label: '工作' },
      semantic: { color: 'bg-blue-500/20 text-blue-400', label: '语义' },
      episodic: { color: 'bg-purple-500/20 text-purple-400', label: '情景' },
    }
    return config[type_]
  }

  // 获取重要性星级
  const getImportanceStars = (importance: number) => {
    return Array.from({ length: 5 }, (_, i) => (
      <Star
        key={i}
        className={clsx(
          'w-3 h-3',
          i < importance ? 'text-yellow-400 fill-yellow-400' : 'text-gray-600'
        )}
      />
    ))
  }

  return (
    <div className="h-full flex flex-col">
      {/* 头部 */}
      <div className="px-6 py-4 border-b border-dark-100">
        <div className="flex items-center justify-between mb-4">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 bg-tortoise-500/20 rounded-lg flex items-center justify-center">
              <Brain className="w-5 h-5 text-tortoise-400" />
            </div>
            <div>
              <h1 className="text-xl font-bold text-white">记忆管理</h1>
              <p className="text-sm text-gray-400">{memories.length} 条记忆</p>
            </div>
          </div>
          
          <button
            onClick={() => setShowNewMemory(true)}
            className="btn btn-primary flex items-center gap-2"
          >
            <Plus className="w-4 h-4" />
            添加记忆
          </button>
        </div>
        
        {/* 搜索和过滤器 */}
        <div className="flex items-center gap-4">
          {/* 搜索 */}
          <div className="relative flex-1 max-w-md">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-500" />
            <input
              type="text"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              placeholder="搜索记忆..."
              className="w-full pl-10 pr-4 py-2 bg-dark-100 border border-dark-100 rounded-lg text-white placeholder-gray-500 focus:outline-none focus:border-tortoise-500"
            />
          </div>
          
          {/* 类型过滤 */}
          <div className="flex items-center gap-1 bg-dark-100 rounded-lg p-1">
            {(['all', 'working', 'semantic', 'episodic'] as MemoryFilter[]).map((f) => (
              <button
                key={f}
                onClick={() => setFilter(f)}
                className={clsx(
                  'px-3 py-1.5 rounded-md text-sm transition-colors',
                  filter === f
                    ? 'bg-tortoise-500/20 text-tortoise-400'
                    : 'text-gray-400 hover:text-white'
                )}
              >
                {f === 'all' ? '全部' : f === 'working' ? '工作' : f === 'semantic' ? '语义' : '情景'}
              </button>
            ))}
          </div>
          
          {/* 排序 */}
          <select
            value={sortBy}
            onChange={(e) => setSortBy(e.target.value as SortBy)}
            className="px-3 py-2 bg-dark-100 border border-dark-100 rounded-lg text-white text-sm focus:outline-none focus:border-tortoise-500"
          >
            <option value="date">按时间</option>
            <option value="importance">按重要性</option>
            <option value="name">按名称</option>
          </select>
        </div>
      </div>
      
      {/* 内容区 */}
      <div className="flex-1 overflow-y-auto p-6">
        {/* 错误提示 */}
        {error && (
          <div className="mb-4 p-4 bg-red-500/10 border border-red-500/30 rounded-lg flex items-center gap-3 text-red-400">
            <AlertCircle className="w-5 h-5" />
            <span>{error}</span>
            <button
              onClick={() => setError(null)}
              className="ml-auto hover:text-red-300"
            >
              <X className="w-4 h-4" />
            </button>
          </div>
        )}
        
        {/* 加载状态 */}
        {isLoading ? (
          <div className="flex items-center justify-center h-64">
            <Loader2 className="w-8 h-8 text-tortoise-400 animate-spin" />
          </div>
        ) : filteredMemories.length === 0 ? (
          /* 空状态 */
          <div className="flex flex-col items-center justify-center h-64 text-gray-500">
            <Brain className="w-16 h-16 mb-4 opacity-50" />
            <p className="text-lg mb-2">
              {searchQuery || filter !== 'all' ? '没有找到匹配的记录' : '还没有记忆'}
            </p>
            <p className="text-sm mb-4">
              {searchQuery || filter !== 'all' ? '尝试调整搜索条件' : '点击上方按钮添加第一条记忆'}
            </p>
            {!searchQuery && filter === 'all' && (
              <button
                onClick={() => setShowNewMemory(true)}
                className="btn btn-primary"
              >
                <Plus className="w-4 h-4 mr-2" />
                添加记忆
              </button>
            )}
          </div>
        ) : (
          /* 记忆列表 */
          <div className="space-y-4">
            {filteredMemories.map((memory) => {
              const typeBadge = getTypeBadge(memory.type)
              
              return (
                <div
                  key={memory.id}
                  className="card p-4 hover:border-dark-100 transition-colors group"
                >
                  <div className="flex items-start justify-between mb-3">
                    <div className="flex items-center gap-2">
                      {/* 类型标签 */}
                      <span className={clsx(
                        'px-2 py-0.5 rounded text-xs font-medium',
                        typeBadge.color
                      )}>
                        {typeBadge.label}
                      </span>
                      
                      {/* 重要性 */}
                      <div className="flex items-center gap-0.5">
                        {getImportanceStars(memory.importance)}
                      </div>
                    </div>
                    
                    {/* 操作按钮 */}
                    <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                      <button
                        onClick={() => setEditingMemory(memory)}
                        className="p-1.5 rounded hover:bg-dark-100 text-gray-400 hover:text-white transition-colors"
                      >
                        <Edit3 className="w-4 h-4" />
                      </button>
                      <button
                        onClick={() => handleDelete(memory.id)}
                        className="p-1.5 rounded hover:bg-red-500/20 text-gray-400 hover:text-red-400 transition-colors"
                      >
                        <Trash2 className="w-4 h-4" />
                      </button>
                    </div>
                  </div>
                  
                  {/* 内容 */}
                  <p className="text-white mb-3 whitespace-pre-wrap">{memory.content}</p>
                  
                  {/* 标签和时间 */}
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-2">
                      {memory.tags.map((tag) => (
                        <span
                          key={tag}
                          className="px-2 py-0.5 bg-dark-200 rounded text-xs text-gray-400 flex items-center gap-1"
                        >
                          <Tag className="w-3 h-3" />
                          {tag}
                        </span>
                      ))}
                    </div>
                    <span className="text-xs text-gray-500 flex items-center gap-1">
                      <Clock className="w-3 h-3" />
                      {new Date(memory.createdAt).toLocaleDateString('zh-CN')}
                    </span>
                  </div>
                </div>
              )
            })}
          </div>
        )}
      </div>
      
      {/* 新建记忆弹窗 */}
      {showNewMemory && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="card w-full max-w-lg p-6">
            <div className="flex items-center justify-between mb-6">
              <h2 className="text-lg font-semibold text-white flex items-center gap-2">
                <Sparkles className="w-5 h-5 text-tortoise-400" />
                添加新记忆
              </h2>
              <button
                onClick={() => setShowNewMemory(false)}
                className="p-1 hover:bg-dark-100 rounded"
              >
                <X className="w-5 h-5 text-gray-400" />
              </button>
            </div>
            
            <div className="space-y-4">
              {/* 类型 */}
              <div>
                <label className="block text-sm text-gray-400 mb-2">记忆类型</label>
                <div className="flex gap-2">
                  {[
                    { value: 'working', label: '工作记忆', color: 'bg-yellow-500' },
                    { value: 'semantic', label: '语义记忆', color: 'bg-blue-500' },
                    { value: 'episodic', label: '情景记忆', color: 'bg-purple-500' },
                  ].map((type) => (
                    <button
                      key={type.value}
                      onClick={() => setNewMemory({ ...newMemory, type: type.value as Memory['type'] })}
                      className={clsx(
                        'flex-1 py-2 rounded-lg text-sm font-medium transition-colors',
                        newMemory.type === type.value
                          ? `${type.color} text-white`
                          : 'bg-dark-100 text-gray-400 hover:text-white'
                      )}
                    >
                      {type.label}
                    </button>
                  ))}
                </div>
              </div>
              
              {/* 内容 */}
              <div>
                <label className="block text-sm text-gray-400 mb-2">记忆内容</label>
                <textarea
                  value={newMemory.content}
                  onChange={(e) => setNewMemory({ ...newMemory, content: e.target.value })}
                  placeholder="输入要记住的内容..."
                  rows={4}
                  className="w-full px-4 py-3 bg-dark-100 border border-dark-100 rounded-lg text-white placeholder-gray-500 resize-none focus:outline-none focus:border-tortoise-500"
                />
              </div>
              
              {/* 重要性 */}
              <div>
                <label className="block text-sm text-gray-400 mb-2">重要性</label>
                <div className="flex items-center gap-2">
                  {Array.from({ length: 5 }, (_, i) => (
                    <button
                      key={i}
                      onClick={() => setNewMemory({ ...newMemory, importance: i + 1 })}
                      className="p-1"
                    >
                      <Star
                        className={clsx(
                          'w-6 h-6 transition-colors',
                          i < newMemory.importance
                            ? 'text-yellow-400 fill-yellow-400'
                            : 'text-gray-600 hover:text-gray-500'
                        )}
                      />
                    </button>
                  ))}
                  <span className="ml-2 text-sm text-gray-400">
                    {newMemory.importance > 3 ? '重要' : newMemory.importance > 1 ? '一般' : '低'}
                  </span>
                </div>
              </div>
              
              {/* 标签 */}
              <div>
                <label className="block text-sm text-gray-400 mb-2">标签 (逗号分隔)</label>
                <input
                  type="text"
                  value={newMemory.tags}
                  onChange={(e) => setNewMemory({ ...newMemory, tags: e.target.value })}
                  placeholder="工作, 项目, 重要"
                  className="w-full px-4 py-2 bg-dark-100 border border-dark-100 rounded-lg text-white placeholder-gray-500 focus:outline-none focus:border-tortoise-500"
                />
              </div>
            </div>
            
            {/* 操作按钮 */}
            <div className="flex justify-end gap-3 mt-6">
              <button
                onClick={() => setShowNewMemory(false)}
                className="btn btn-secondary"
              >
                取消
              </button>
              <button
                onClick={handleCreate}
                disabled={!newMemory.content.trim() || isCreating}
                className={clsx(
                  'btn btn-primary',
                  (!newMemory.content.trim() || isCreating) && 'opacity-70 cursor-not-allowed'
                )}
              >
                {isCreating ? (
                  <>
                    <Loader2 className="w-4 h-4 animate-spin" />
                    保存中...
                  </>
                ) : (
                  <>
                    <CheckCircle2 className="w-4 h-4" />
                    保存
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

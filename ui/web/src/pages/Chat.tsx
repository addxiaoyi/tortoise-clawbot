import { useState, useEffect, useRef, useCallback } from 'react'
import { useAppStore } from '../store/appStore'
import { api } from '../services/api'
import clsx from 'clsx'
import {
  Send,
  Plus,
  MoreVertical,
  Trash2,
  Edit3,
  Copy,
  Check,
  Loader2,
  MessageSquare,
  Bot,
  User,
  Settings,
  X,
  Search,
  Sparkles,
} from 'lucide-react'

interface Message {
  id: string
  role: 'user' | 'assistant' | 'system'
  content: string
  timestamp: Date
  attachments?: Array<{
    id: string
    type: 'image' | 'file'
    name: string
    url: string
  }>
  streaming?: boolean
  error?: boolean
}

export default function Chat() {
  const { 
    sessions, 
    currentSession, 
    messages, 
    loadSessions, 
    createSession, 
    selectSession, 
    deleteSession,
    loadMessages,
    sendMessage,
  } = useAppStore()

  const [input, setInput] = useState('')
  const [isStreaming, setIsStreaming] = useState(false)
  const [showNewChat, setShowNewChat] = useState(false)
  const [newChatName, setNewChatName] = useState('')
  const [contextMenu, setContextMenu] = useState<{ x: number; y: number; sessionId: string } | null>(null)
  const [copiedId, setCopiedId] = useState<string | null>(null)
  const [isLoading, setIsLoading] = useState(false)

  const messagesEndRef = useRef<HTMLDivElement>(null)
  const textareaRef = useRef<HTMLTextAreaElement>(null)
  const contextMenuRef = useRef<HTMLDivElement>(null)

  // 加载会话
  useEffect(() => {
    loadSessions()
  }, [])

  // 自动滚动到底部
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages])

  // 点击外部关闭菜单
  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (contextMenuRef.current && !contextMenuRef.current.contains(e.target as Node)) {
        setContextMenu(null)
      }
    }
    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [])

  // 创建新会话
  const handleCreateChat = async () => {
    if (!newChatName.trim()) return
    
    try {
      const session = await createSession(newChatName.trim(), 'user')
      await selectSession(session.id)
      setNewChatName('')
      setShowNewChat(false)
    } catch (error) {
      console.error('创建会话失败:', error)
    }
  }

  // 发送消息
  const handleSend = async () => {
    if (!input.trim() || !currentSession || isStreaming) return

    const messageContent = input.trim()
    setInput('')
    setIsStreaming(true)
    setIsLoading(true)

    // 添加用户消息
    const userMsg: Message = {
      id: crypto.randomUUID(),
      role: 'user',
      content: messageContent,
      timestamp: new Date(),
    }

    // 添加占位 AI 消息
    const aiMsg: Message = {
      id: crypto.randomUUID(),
      role: 'assistant',
      content: '',
      timestamp: new Date(),
      streaming: true,
    }

    // 更新消息列表
    await sendMessage(messageContent)

    try {
      // 调用 API
      const response = await api.sendMessage(currentSession.id, messageContent)
      
      // 更新 AI 消息
      setIsStreaming(false)
      setIsLoading(false)
    } catch (error) {
      console.error('发送消息失败:', error)
      setIsStreaming(false)
      setIsLoading(false)
    }
  }

  // 处理回车发送
  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      handleSend()
    }
  }

  // 复制消息
  const handleCopy = async (content: string, id: string) => {
    await navigator.clipboard.writeText(content)
    setCopiedId(id)
    setTimeout(() => setCopiedId(null), 2000)
  }

  // 删除会话
  const handleDeleteSession = async (id: string) => {
    await deleteSession(id)
    setContextMenu(null)
  }

  // 调整文本框高度
  const adjustTextareaHeight = () => {
    const textarea = textareaRef.current
    if (textarea) {
      textarea.style.height = 'auto'
      textarea.style.height = Math.min(textarea.scrollHeight, 200) + 'px'
    }
  }

  return (
    <div className="flex h-full">
      {/* 侧边栏 */}
      <div className={clsx(
        'w-72 border-r border-dark-100 flex flex-col transition-all duration-300',
        'bg-dark-50'
      )}>
        {/* 新建聊天按钮 */}
        <div className="p-4">
          <button
            onClick={() => setShowNewChat(true)}
            className="w-full btn btn-primary flex items-center justify-center gap-2"
          >
            <Plus className="w-4 h-4" />
            新建对话
          </button>
        </div>

        {/* 新建聊天弹窗 */}
        {showNewChat && (
          <div className="px-4 pb-4">
            <div className="card p-4 bg-dark-100">
              <input
                type="text"
                value={newChatName}
                onChange={(e) => setNewChatName(e.target.value)}
                onKeyDown={(e) => e.key === 'Enter' && handleCreateChat()}
                placeholder="对话标题..."
                className="w-full px-3 py-2 bg-dark-200 border border-dark-100 rounded-lg text-white mb-3"
                autoFocus
              />
              <div className="flex gap-2">
                <button
                  onClick={handleCreateChat}
                  className="flex-1 btn btn-primary text-sm"
                >
                  创建
                </button>
                <button
                  onClick={() => {
                    setShowNewChat(false)
                    setNewChatName('')
                  }}
                  className="btn btn-secondary text-sm"
                >
                  取消
                </button>
              </div>
            </div>
          </div>
        )}

        {/* 会话列表 */}
        <div className="flex-1 overflow-y-auto px-2">
          {sessions.length === 0 ? (
            <div className="text-center py-8 text-gray-500">
              <MessageSquare className="w-12 h-12 mx-auto mb-3 opacity-50" />
              <p className="text-sm">暂无会话</p>
              <p className="text-xs mt-1">点击上方按钮开始新对话</p>
            </div>
          ) : (
            <div className="space-y-1 py-2">
              {sessions.map((session) => (
                <div
                  key={session.id}
                  className={clsx(
                    'group relative px-3 py-2 rounded-lg cursor-pointer transition-colors',
                    currentSession?.id === session.id
                      ? 'bg-tortoise-500/20 text-tortoise-400'
                      : 'hover:bg-dark-100 text-gray-400'
                  )}
                  onClick={() => selectSession(session.id)}
                >
                  <div className="flex items-center gap-2">
                    <MessageSquare className="w-4 h-4 flex-shrink-0" />
                    <span className="truncate text-sm">{session.name}</span>
                  </div>
                  
                  {/* 上下文菜单按钮 */}
                  <button
                    onClick={(e) => {
                      e.stopPropagation()
                      setContextMenu({
                        x: e.clientX,
                        y: e.clientY,
                        sessionId: session.id,
                      })
                    }}
                    className="absolute right-2 top-1/2 -translate-y-1/2 p-1 rounded opacity-0 group-hover:opacity-100 hover:bg-dark-200 transition-opacity"
                  >
                    <MoreVertical className="w-4 h-4" />
                  </button>
                </div>
              ))}
            </div>
          )}
        </div>

        {/* 设置入口 */}
        <div className="p-4 border-t border-dark-100">
          <button className="w-full flex items-center gap-2 px-3 py-2 rounded-lg text-gray-400 hover:bg-dark-100 hover:text-white transition-colors">
            <Settings className="w-4 h-4" />
            <span className="text-sm">设置</span>
          </button>
        </div>
      </div>

      {/* 主聊天区域 */}
      <div className="flex-1 flex flex-col bg-dark-50">
        {currentSession ? (
          <>
            {/* 聊天头部 */}
            <div className="px-6 py-4 border-b border-dark-100 bg-dark-50">
              <div className="flex items-center justify-between">
                <div>
                  <h2 className="text-white font-medium">{currentSession.name}</h2>
                  <p className="text-xs text-gray-500 mt-1">
                    {messages.length} 条消息
                  </p>
                </div>
                <div className="flex items-center gap-2">
                  <button className="p-2 rounded-lg hover:bg-dark-100 text-gray-400 hover:text-white transition-colors">
                    <Search className="w-4 h-4" />
                  </button>
                </div>
              </div>
            </div>

            {/* 消息列表 */}
            <div className="flex-1 overflow-y-auto px-6 py-4 space-y-4">
              {messages.length === 0 ? (
                <div className="h-full flex items-center justify-center">
                  <div className="text-center">
                    <div className="w-16 h-16 bg-tortoise-500/20 rounded-2xl flex items-center justify-center mx-auto mb-4">
                      <Sparkles className="w-8 h-8 text-tortoise-400" />
                    </div>
                    <h3 className="text-xl font-semibold text-white mb-2">
                      开始新对话
                    </h3>
                    <p className="text-gray-400 text-sm max-w-md">
                      发送消息与 AI 开始对话，或使用快捷命令获取帮助
                    </p>
                  </div>
                </div>
              ) : (
                messages.map((msg) => (
                  <div
                    key={msg.id}
                    className={clsx(
                      'flex gap-3',
                      msg.role === 'user' && 'flex-row-reverse'
                    )}
                  >
                    {/* 头像 */}
                    <div className={clsx(
                      'w-8 h-8 rounded-lg flex items-center justify-center flex-shrink-0',
                      msg.role === 'user'
                        ? 'bg-blue-500'
                        : msg.role === 'system'
                        ? 'bg-orange-500'
                        : 'bg-tortoise-500'
                    )}>
                      {msg.role === 'user' ? (
                        <User className="w-4 h-4 text-white" />
                      ) : (
                        <Bot className="w-4 h-4 text-white" />
                      )}
                    </div>

                    {/* 消息内容 */}
                    <div className={clsx(
                      'max-w-[70%] group',
                      msg.role === 'user' && 'items-end'
                    )}>
                      <div className={clsx(
                        'rounded-2xl px-4 py-3',
                        msg.role === 'user'
                          ? 'bg-blue-500 text-white rounded-tr-sm'
                          : msg.role === 'system'
                          ? 'bg-orange-500/20 text-orange-400'
                          : 'bg-dark-100 text-white rounded-tl-sm'
                      )}>
                        {msg.error ? (
                          <p className="text-red-400 text-sm">{msg.content}</p>
                        ) : (
                          <p className="whitespace-pre-wrap">{msg.content}</p>
                        )}
                      </div>

                      {/* 消息操作 */}
                      <div className={clsx(
                        'flex items-center gap-1 mt-1 opacity-0 group-hover:opacity-100 transition-opacity',
                        msg.role === 'user' && 'justify-end'
                      )}>
                        <span className="text-xs text-gray-500">
                          {new Date(msg.timestamp).toLocaleTimeString()}
                        </span>
                        <button
                          onClick={() => handleCopy(msg.content, msg.id)}
                          className="p-1 rounded hover:bg-dark-100 text-gray-400 hover:text-white transition-colors"
                        >
                          {copiedId === msg.id ? (
                            <Check className="w-3 h-3" />
                          ) : (
                            <Copy className="w-3 h-3" />
                          )}
                        </button>
                      </div>
                    </div>
                  </div>
                ))
              )}

              {/* 加载指示器 */}
              {isLoading && (
                <div className="flex gap-3">
                  <div className="w-8 h-8 rounded-lg bg-tortoise-500 flex items-center justify-center">
                    <Bot className="w-4 h-4 text-white" />
                  </div>
                  <div className="bg-dark-100 rounded-2xl rounded-tl-sm px-4 py-3">
                    <div className="flex items-center gap-2 text-gray-400">
                      <Loader2 className="w-4 h-4 animate-spin" />
                      <span className="text-sm">正在思考...</span>
                    </div>
                  </div>
                </div>
              )}

              <div ref={messagesEndRef} />
            </div>

            {/* 输入区域 */}
            <div className="px-6 py-4 border-t border-dark-100 bg-dark-50">
              <div className="flex items-end gap-3">
                <div className="flex-1 relative">
                  <textarea
                    ref={textareaRef}
                    value={input}
                    onChange={(e) => {
                      setInput(e.target.value)
                      adjustTextareaHeight()
                    }}
                    onKeyDown={handleKeyDown}
                    placeholder="输入消息... (Shift+Enter 换行)"
                    className="w-full px-4 py-3 bg-dark-100 border border-dark-100 rounded-xl text-white placeholder-gray-500 resize-none focus:outline-none focus:border-tortoise-500 transition-colors"
                    rows={1}
                    disabled={isStreaming}
                  />
                </div>
                <button
                  onClick={handleSend}
                  disabled={!input.trim() || isStreaming}
                  className={clsx(
                    'p-3 rounded-xl transition-colors',
                    input.trim() && !isStreaming
                      ? 'bg-tortoise-500 hover:bg-tortoise-600 text-white'
                      : 'bg-dark-200 text-gray-500 cursor-not-allowed'
                  )}
                >
                  <Send className="w-5 h-5" />
                </button>
              </div>
              
              <p className="text-xs text-gray-500 mt-2 text-center">
                AI 可能会产生不准确的信息，请仔细核对重要内容
              </p>
            </div>
          </>
        ) : (
          <div className="flex-1 flex items-center justify-center">
            <div className="text-center">
              <div className="w-20 h-20 bg-tortoise-500/20 rounded-2xl flex items-center justify-center mx-auto mb-4">
                <MessageSquare className="w-10 h-10 text-tortoise-400" />
              </div>
              <h2 className="text-xl font-semibold text-white mb-2">
                Tortoise AI
              </h2>
              <p className="text-gray-400 mb-4">
                选择或创建一个会话开始对话
              </p>
              <button
                onClick={() => setShowNewChat(true)}
                className="btn btn-primary"
              >
                <Plus className="w-4 h-4 mr-2" />
                新建对话
              </button>
            </div>
          </div>
        )}
      </div>

      {/* 上下文菜单 */}
      {contextMenu && (
        <div
          ref={contextMenuRef}
          className="fixed z-50 w-48 card bg-dark-100 border border-dark-100 shadow-xl py-1"
          style={{ left: contextMenu.x, top: contextMenu.y }}
        >
          <button
            onClick={() => {
              setNewChatName(sessions.find(s => s.id === contextMenu.sessionId)?.name || '')
              handleDeleteSession(contextMenu.sessionId)
            }}
            className="w-full px-4 py-2 text-left text-sm text-red-400 hover:bg-dark-200 flex items-center gap-2"
          >
            <Trash2 className="w-4 h-4" />
            删除对话
          </button>
        </div>
      )}
    </div>
  )
}

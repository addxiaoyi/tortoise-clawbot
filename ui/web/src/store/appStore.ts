import { create } from 'zustand'
import { api } from '../services/api'

export interface Message {
  id: string
  sessionId: string
  role: 'user' | 'assistant' | 'system'
  content: string
  timestamp: Date
  attachments?: Attachment[]
  metadata?: Record<string, unknown>
}

export interface Attachment {
  id: string
  type: 'image' | 'file' | 'audio'
  name: string
  url: string
  size?: number
}

export interface Session {
  id: string
  name: string
  userId: string
  createdAt: Date
  updatedAt: Date
  messageCount: number
  lastMessage?: string
}

export interface Memory {
  id: string
  type: 'working' | 'semantic' | 'episodic'
  content: string
  importance: number
  createdAt: Date
  updatedAt: Date
  tags?: string[]
}

export interface Plugin {
  id: string
  name: string
  version: string
  description: string
  author: string
  enabled: boolean
  tools: Tool[]
  status: 'active' | 'inactive' | 'error'
}

export interface Tool {
  name: string
  description: string
  parameters: Parameter[]
}

export interface Parameter {
  name: string
  type: 'string' | 'number' | 'boolean' | 'object' | 'array'
  required: boolean
  description?: string
}

interface AppState {
  // Connection
  isConnected: boolean
  isConnecting: boolean
  serverUrl: string
  
  // Sessions
  sessions: Session[]
  currentSession: Session | null
  messages: Message[]
  
  // Memory
  memories: Memory[]
  memoryType: 'all' | 'working' | 'semantic' | 'episodic'
  
  // Plugins
  plugins: Plugin[]
  
  // UI State
  sidebarCollapsed: boolean
  theme: 'dark' | 'light'
  
  // Actions
  initializeApp: () => Promise<void>
  connect: (serverUrl: string) => Promise<void>
  disconnect: () => void
  
  // Session Actions
  loadSessions: () => Promise<void>
  createSession: (name: string, userId: string) => Promise<Session>
  selectSession: (sessionId: string) => Promise<void>
  deleteSession: (sessionId: string) => Promise<void>
  
  // Message Actions
  loadMessages: (sessionId: string) => Promise<void>
  sendMessage: (content: string, attachments?: Attachment[]) => Promise<void>
  
  // Memory Actions
  loadMemories: (type?: string) => Promise<void>
  addMemory: (type: string, content: string, importance: number) => Promise<void>
  deleteMemory: (memoryId: string) => Promise<void>
  searchMemories: (query: string) => Promise<Memory[]>
  
  // Plugin Actions
  loadPlugins: () => Promise<void>
  togglePlugin: (pluginId: string) => Promise<void>
  
  // UI Actions
  toggleSidebar: () => void
  setTheme: (theme: 'dark' | 'light') => void
}

export const useAppStore = create<AppState>((set, get) => ({
  // Initial State
  isConnected: false,
  isConnecting: false,
  serverUrl: 'http://localhost:18792',
  
  sessions: [],
  currentSession: null,
  messages: [],
  
  memories: [],
  memoryType: 'all',
  
  plugins: [],
  
  sidebarCollapsed: false,
  theme: 'dark',
  
  // Initialize App
  initializeApp: async () => {
    try {
      await get().connect(get().serverUrl)
      await get().loadSessions()
      await get().loadPlugins()
    } catch (error) {
      console.error('Failed to initialize app:', error)
    }
  },
  
  // Connection
  connect: async (serverUrl: string) => {
    set({ isConnecting: true, serverUrl })
    try {
      await api.healthCheck(serverUrl)
      set({ isConnected: true, isConnecting: false })
    } catch (error) {
      set({ isConnected: false, isConnecting: false })
      throw error
    }
  },
  
  disconnect: () => {
    set({ isConnected: false })
  },
  
  // Session Actions
  loadSessions: async () => {
    try {
      const sessions = await api.getSessions()
      set({ sessions })
    } catch (error) {
      console.error('Failed to load sessions:', error)
    }
  },
  
  createSession: async (name: string, userId: string) => {
    const session = await api.createSession(name, userId)
    set((state) => ({ sessions: [session, ...state.sessions] }))
    return session
  },
  
  selectSession: async (sessionId: string) => {
    const session = get().sessions.find((s) => s.id === sessionId)
    if (session) {
      set({ currentSession: session })
      await get().loadMessages(sessionId)
    }
  },
  
  deleteSession: async (sessionId: string) => {
    await api.deleteSession(sessionId)
    set((state) => ({
      sessions: state.sessions.filter((s) => s.id !== sessionId),
      currentSession: state.currentSession?.id === sessionId ? null : state.currentSession,
    }))
  },
  
  // Message Actions
  loadMessages: async (sessionId: string) => {
    try {
      const messages = await api.getMessages(sessionId)
      set({ messages })
    } catch (error) {
      console.error('Failed to load messages:', error)
    }
  },
  
  sendMessage: async (content: string, attachments?: Attachment[]) => {
    const { currentSession } = get()
    if (!currentSession) return
    
    // Add user message
    const userMessage: Message = {
      id: crypto.randomUUID(),
      sessionId: currentSession.id,
      role: 'user',
      content,
      timestamp: new Date(),
      attachments,
    }
    
    set((state) => ({
      messages: [...state.messages, userMessage],
    }))
    
    try {
      // Send to API
      const response = await api.sendMessage(currentSession.id, content)
      
      // Add assistant message
      const assistantMessage: Message = {
        id: response.messageId || crypto.randomUUID(),
        sessionId: currentSession.id,
        role: 'assistant',
        content: response.content || response.text || '...',
        timestamp: new Date(),
      }
      
      set((state) => ({
        messages: [...state.messages, assistantMessage],
      }))
    } catch (error) {
      console.error('Failed to send message:', error)
    }
  },
  
  // Memory Actions
  loadMemories: async (type?: string) => {
    try {
      const memories = await api.getMemories(type)
      set({ memories, memoryType: (type || 'all') as any })
    } catch (error) {
      console.error('Failed to load memories:', error)
    }
  },
  
  addMemory: async (type: string, content: string, importance: number) => {
    await api.addMemory(type, content, importance)
    await get().loadMemories(get().memoryType)
  },
  
  deleteMemory: async (memoryId: string) => {
    await api.deleteMemory(memoryId)
    set((state) => ({
      memories: state.memories.filter((m) => m.id !== memoryId),
    }))
  },
  
  searchMemories: async (query: string) => {
    return await api.searchMemories(query)
  },
  
  // Plugin Actions
  loadPlugins: async () => {
    try {
      const plugins = await api.getPlugins()
      set({ plugins })
    } catch (error) {
      console.error('Failed to load plugins:', error)
    }
  },
  
  togglePlugin: async (pluginId: string) => {
    const plugin = get().plugins.find((p) => p.id === pluginId)
    if (plugin) {
      await api.togglePlugin(pluginId, !plugin.enabled)
      set((state) => ({
        plugins: state.plugins.map((p) =>
          p.id === pluginId ? { ...p, enabled: !p.enabled } : p
        ),
      }))
    }
  },
  
  // UI Actions
  toggleSidebar: () => {
    set((state) => ({ sidebarCollapsed: !state.sidebarCollapsed }))
  },
  
  setTheme: (theme: 'dark' | 'light') => {
    set({ theme })
    document.documentElement.classList.toggle('light', theme === 'light')
  },
}))

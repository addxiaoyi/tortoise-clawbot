import { useEffect, useRef, useState, useCallback } from 'react'

interface WSMessage {
  type: string
  data: any
  timestamp: number
}

interface StreamChunk {
  content: string
  done: boolean
  tokens?: number
}

interface Message {
  id: string
  role: 'user' | 'assistant' | 'system'
  content: string
  createdAt: string
}

interface UseChatOptions {
  sessionId: string
  baseURL?: string
  onError?: (error: string) => void
  onConnected?: () => void
  onDisconnected?: () => void
}

export function useChat({ sessionId, baseURL = 'ws://localhost:18792', onError, onConnected, onDisconnected }: UseChatOptions) {
  const wsRef = useRef<WebSocket | null>(null)
  const [isConnected, setIsConnected] = useState(false)
  const [isStreaming, setIsStreaming] = useState(false)
  const [streamingContent, setStreamingContent] = useState('')
  const [error, setError] = useState<string | null>(null)

  const connect = useCallback(() => {
    if (wsRef.current?.readyState === WebSocket.OPEN) return

    const ws = new WebSocket(`${baseURL}/ws`)
    wsRef.current = ws

    ws.onopen = () => {
      console.log('WebSocket connected')
      setIsConnected(true)
      onConnected?.()
    }

    ws.onclose = () => {
      console.log('WebSocket disconnected')
      setIsConnected(false)
      onDisconnected?.()
      // 自动重连
      setTimeout(connect, 3000)
    }

    ws.onerror = (e) => {
      console.error('WebSocket error:', e)
      setError('连接失败')
    }

    ws.onmessage = (event) => {
      try {
        const msg: WSMessage = JSON.parse(event.data)
        handleMessage(msg)
      } catch (e) {
        console.error('Failed to parse message:', e)
      }
    }
  }, [baseURL, onConnected, onDisconnected])

  const handleMessage = useCallback((msg: WSMessage) => {
    switch (msg.type) {
      case 'pong':
        // 心跳响应
        break

      case 'error':
        setError(msg.data.message)
        onError?.(msg.data.message)
        setIsStreaming(false)
        break

      case 'stream':
        setStreamingContent(prev => prev + msg.data.content)
        break

      case 'user_message':
      case 'assistant_message':
        setIsStreaming(false)
        setStreamingContent('')
        break
    }
  }, [onError])

  const sendMessage = useCallback((content: string, model?: string) => {
    if (!wsRef.current || wsRef.current.readyState !== WebSocket.OPEN) {
      setError('未连接')
      return
    }

    setIsStreaming(true)
    setStreamingContent('')
    setError(null)

    wsRef.current.send(JSON.stringify({
      type: 'chat',
      session: sessionId,
      content,
      model,
    }))
  }, [sessionId])

  const sendPing = useCallback(() => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify({ type: 'ping' }))
    }
  }, [])

  useEffect(() => {
    connect()
    return () => {
      wsRef.current?.close()
    }
  }, [connect])

  // 定期发送心跳
  useEffect(() => {
    const interval = setInterval(sendPing, 25000)
    return () => clearInterval(interval)
  }, [sendPing])

  return {
    isConnected,
    isStreaming,
    streamingContent,
    error,
    sendMessage,
    reconnect: connect,
  }
}

// 非 hook 版本的 WebSocket 服务
export class ChatService {
  private ws: WebSocket | null = null
  private url: string
  private sessionId: string
  private onStream?: (chunk: StreamChunk) => void
  private onComplete?: (content: string) => void
  private onError?: (error: string) => void
  private onConnect?: () => void
  private onDisconnect?: () => void

  constructor(url: string = 'ws://localhost:18792', sessionId: string = '') {
    this.url = url
    this.sessionId = sessionId
  }

  setSession(sessionId: string) {
    this.sessionId = sessionId
  }

  connect(options: {
    onStream?: (chunk: StreamChunk) => void
    onComplete?: (content: string) => void
    onError?: (error: string) => void
    onConnect?: () => void
    onDisconnect?: () => void
  }) {
    this.onStream = options.onStream
    this.onComplete = options.onComplete
    this.onError = options.onError
    this.onConnect = options.onConnect
    this.onDisconnect = options.onDisconnect

    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.close()
    }

    this.ws = new WebSocket(`${this.url}/ws`)

    this.ws.onopen = () => {
      console.log('ChatService connected')
      this.onConnect?.()
    }

    this.ws.onclose = () => {
      console.log('ChatService disconnected')
      this.onDisconnect?.()
    }

    this.ws.onerror = (e) => {
      console.error('ChatService error:', e)
      this.onError?.('连接失败')
    }

    this.ws.onmessage = (event) => {
      try {
        const msg: WSMessage = JSON.parse(event.data)
        this.handleMessage(msg)
      } catch (e) {
        console.error('Failed to parse:', e)
      }
    }
  }

  private handleMessage(msg: WSMessage) {
    switch (msg.type) {
      case 'stream':
        this.onStream?.({ content: msg.data.content, done: false })
        break

      case 'assistant_message':
        this.onComplete?.(msg.data.message.content)
        break

      case 'error':
        this.onError?.(msg.data.message)
        break
    }
  }

  send(content: string, model?: string) {
    if (this.ws?.readyState !== WebSocket.OPEN) {
      this.onError?.('未连接')
      return
    }

    this.ws.send(JSON.stringify({
      type: 'chat',
      session: this.sessionId,
      content,
      model,
    }))
  }

  disconnect() {
    this.ws?.close()
    this.ws = null
  }

  isConnected() {
    return this.ws?.readyState === WebSocket.OPEN
  }
}

export type { WSMessage, StreamChunk, Message }

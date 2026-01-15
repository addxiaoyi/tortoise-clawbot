package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许所有来源
	},
}

// WSClient WebSocket 客户端
type WSClient struct {
	ID       string
	Conn     *websocket.Conn
	Send     chan []byte
	SendLock sync.Mutex
}

// WSManager WebSocket 管理器
type WSManager struct {
	clients    map[string]*WSClient
	broadcast  chan []byte
	register   chan *WSClient
	unregister chan *WSClient
	mu         sync.RWMutex
}

func NewWSManager() *WSManager {
	return &WSManager{
		clients:    make(map[string]*WSClient),
		broadcast:  make(chan []byte),
		register:   make(chan *WSClient),
		unregister: make(chan *WSClient),
	}
}

func (m *WSManager) Run() {
	for {
		select {
		case client := <-m.register:
			m.mu.Lock()
			m.clients[client.ID] = client
			m.mu.Unlock()
			log.Printf("WebSocket client connected: %s", client.ID)

		case client := <-m.unregister:
			m.mu.Lock()
			if _, ok := m.clients[client.ID]; ok {
				delete(m.clients, client.ID)
				close(client.Send)
			}
			m.mu.Unlock()
			log.Printf("WebSocket client disconnected: %s", client.ID)

		case message := <-m.broadcast:
			m.mu.RLock()
			for _, client := range m.clients {
				client.SendLock.Lock()
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(m.clients, client.ID)
				}
				client.SendLock.Unlock()
			}
			m.mu.RUnlock()
		}
	}
}

// SendTo 发送消息给指定客户端
func (m *WSManager) SendTo(clientID string, msgType string, data interface{}) error {
	m.mu.RLock()
	client, ok := m.clients[clientID]
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("client not found: %s", clientID)
	}

	msg := WSMessage{
		Type:      msgType,
		Data:      data,
		Timestamp: time.Now().Unix(),
	}

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	client.SendLock.Lock()
	defer client.SendLock.Unlock()

	select {
	case client.Send <- msgBytes:
		return nil
	default:
		return fmt.Errorf("send channel full")
	}
}

// Broadcast 广播消息
func (m *WSManager) Broadcast(msgType string, data interface{}) {
	msg := WSMessage{
		Type:      msgType,
		Data:      data,
		Timestamp: time.Now().Unix(),
	}

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Failed to marshal broadcast message: %v", err)
		return
	}

	m.broadcast <- msgBytes
}

// ClientCount 获取客户端数量
func (m *WSManager) ClientCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.clients)
}

// WSMessage WebSocket 消息结构
type WSMessage struct {
	Type      string      `json:"type"`
	Data      interface{} `json:"data"`
	Timestamp int64       `json:"timestamp"`
}

// handleWebSocket 处理 WebSocket 连接
func (s *Server) handleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	clientID := fmt.Sprintf("%d", time.Now().UnixNano())
	client := &WSClient{
		ID:   clientID,
		Conn: conn,
		Send: make(chan []byte, 256),
	}

	s.wsManager.Register(client)

	go client.WritePump()
	go client.ReadPump(s)
}

// Register 注册客户端
func (m *WSManager) Register(client *WSClient) {
	m.register <- client
}

// Unregister 注销客户端
func (m *WSManager) Unregister(client *WSClient) {
	m.unregister <- client
}

// WritePump 处理发送
func (c *WSClient) WritePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.SendLock.Lock()
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				c.SendLock.Unlock()
				return
			}

			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				c.SendLock.Unlock()
				return
			}
			c.SendLock.Unlock()

		case <-ticker.C:
			c.SendLock.Lock()
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				c.SendLock.Unlock()
				return
			}
			c.SendLock.Unlock()
		}
	}
}

// ReadPump 处理接收
func (c *WSClient) ReadPump(s *Server) {
	defer func() {
		s.wsManager.Unregister(c)
		c.Conn.Close()
	}()

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// 解析消息
		var req WSRequest
		if err := json.Unmarshal(message, &req); err != nil {
			log.Printf("Failed to parse WebSocket message: %v", err)
			continue
		}

		// 处理请求
		s.handleWSRequest(c, &req)
	}
}

// WSRequest WebSocket 请求
type WSRequest struct {
	Type    string          `json:"type"`
	Session string          `json:"session,omitempty"`
	Content string          `json:"content,omitempty"`
	Model   string          `json:"model,omitempty"`
}

// handleWSRequest 处理 WebSocket 请求
func (s *Server) handleWSRequest(client *WSClient, req *WSRequest) {
	switch req.Type {
	case "ping":
		client.SendLock.Lock()
		client.Send <- []byte(`{"type":"pong","timestamp":` + fmt.Sprintf("%d", time.Now().Unix()) + `}`)
		client.SendLock.Unlock()

	case "chat":
		s.handleWSChat(client, req)

	case "session_create":
		s.handleWSCreateSession(client, req)

	default:
		s.wsManager.SendTo(client.ID, "error", map[string]string{
			"message": fmt.Sprintf("Unknown message type: %s", req.Type),
		})
	}
}

// handleWSChat 处理聊天请求
func (s *Server) handleWSChat(client *WSClient, req *WSRequest) {
	if req.Session == "" || req.Content == "" {
		s.wsManager.SendTo(client.ID, "error", map[string]string{
			"message": "session and content are required",
		})
		return
	}

	// 检查会话是否存在
	if _, ok := s.sessionStore.GetSession(req.Session); !ok {
		s.wsManager.SendTo(client.ID, "error", map[string]string{
			"message": "session not found",
		})
		return
	}

	// 保存用户消息
	userMsg := s.messageStore.CreateMessage(req.Session, "user", req.Content)
	s.sessionStore.IncrementMessageCount(req.Session, truncateString(req.Content, 50))

	// 发送用户消息确认
	s.wsManager.SendTo(client.ID, "user_message", userMsg)

	// 获取历史消息
	historyMessages := s.messageStore.GetMessages(req.Session, 20)

	// 构建 AI 请求
	var streamErr error
	fullContent := ""

	if s.aiEngine != nil {
		chatReq := &ai.ChatRequest{
			Model:       req.Model,
			Temperature: 0.7,
			MaxTokens:   4096,
		}

		for _, msg := range historyMessages {
			role := "user"
			if msg.Role == "assistant" {
				role = "assistant"
			}
			chatReq.Messages = append(chatReq.Messages, ai.Message{
				Role:    role,
				Content: msg.Content,
			})
		}

		// 流式调用 AI
		streamErr = s.aiEngine.ChatWithStreaming(nil, chatReq, func(chunk *ai.StreamingChunk) {
			fullContent += chunk.Content

			// 发送流式片段
			s.wsManager.SendTo(client.ID, "stream", map[string]interface{}{
				"content": chunk.Content,
				"done":    chunk.Done,
				"tokens":  chunk.TotalTokens,
			})
		})
	} else {
		// 模拟流式响应
		responses := []string{
			"我理解你的问题。",
			"让我来分析一下...",
			"\n\n首先，我们需要考虑基本原理。",
			"\n然后逐步深入。",
		}
		for _, resp := range responses {
			fullContent += resp
			s.wsManager.SendTo(client.ID, "stream", map[string]interface{}{
				"content": resp,
				"done":    false,
			})
			time.Sleep(200 * time.Millisecond)
		}
		s.wsManager.SendTo(client.ID, "stream", map[string]interface{}{
			"content": "",
			"done":    true,
			"tokens":  100,
		})
	}

	if streamErr != nil {
		s.wsManager.SendTo(client.ID, "error", map[string]string{
			"message": fmt.Sprintf("AI request failed: %v", streamErr),
		})
		return
	}

	// 保存 AI 响应
	assistantMsg := s.messageStore.CreateMessage(req.Session, "assistant", fullContent)
	s.sessionStore.IncrementMessageCount(req.Session, truncateString(fullContent, 50))

	// 发送完成消息
	s.wsManager.SendTo(client.ID, "assistant_message", map[string]interface{}{
		"message": assistantMsg,
		"model":   "gpt-4",
		"tokens":  len(fullContent) / 4, // 估算
	})
}

// handleWSCreateSession 创建会话
func (s *Server) handleWSCreateSession(client *WSClient, req *WSRequest) {
	var name string
	if data, ok := req.Content.(string); ok {
		name = data
	} else {
		name = fmt.Sprintf("Chat %d", time.Now().Unix())
	}

	session := s.sessionStore.CreateSession(name, "websocket")
	s.wsManager.SendTo(client.ID, "session_created", session)
}

// InitWSManager 初始化 WebSocket 管理器
func (s *Server) InitWSManager() {
	s.wsManager = NewWSManager()
	go s.wsManager.Run()
}

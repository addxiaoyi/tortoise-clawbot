package websocket

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

// Config WebSocket配置
type Config struct {
	MaxConnections int
	ReadBuffer   int
	WriteBuffer  int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// Server WebSocket服务器
type Server struct {
	config Config

	// 连接管理
	connections map[string]*Conn
	upgrader    websocket.Upgrader

	// 消息处理
	handlers map[string]MessageHandler

	// 统计
	stats Stats

	// 控制
	ctx    context.Context
	cancel context.CancelFunc

	mu sync.RWMutex
}

// Conn 连接
type Conn struct {
	ID        string
	UserID   string
	Socket   *websocket.Conn
	RemoteAddr string
	Protocol   string
	ConnectedAt time.Time
	LastPong  time.Time
	mu        sync.RWMutex
}

// Message 消息
type Message struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
	ID     string          `json:"id,omitempty"`
}

// MessageHandler 消息处理器
type Handler func(conn *Conn, msg *Message)

// Stats WebSocket统计
type Stats struct {
	ConnectionsTotal  atomic.Int64
	ConnectionsActive atomic.Int64
	MessagesReceived  atomic.Int64
	MessagesSent     atomic.Int64
	BytesReceived    atomic.Int64
	BytesSent       atomic.Int64
	PingsSent        atomic.Int64
	PongsReceived    atomic.Int64
	Errors           atomic.Int64
}

// NewServer 创建WebSocket服务器
func NewServer(cfg Config) *Server {
	if cfg.MaxConnections == 0 {
		cfg.MaxConnections = 100000
	}
	if cfg.ReadBuffer == 0 {
		cfg.ReadBuffer = 4096
	}
	if cfg.WriteBuffer == 0 {
		cfg.WriteBuffer = 4096
	}
	if cfg.ReadTimeout == 0 {
		cfg.ReadTimeout = 60 * time.Second
	}
	if cfg.WriteTimeout == 0 {
		cfg.WriteTimeout = 60 * time.Second
	}

	ctx, cancel := context.WithCancel(context.Background())

	s := &Server{
		config:      cfg,
		connections: make(map[string]*Conn),
		handlers:   make(map[string]MessageHandler),
		ctx:        ctx,
		cancel:     cancel,
	}

	s.upgrader = websocket.Upgrader{
		ReadBufferSize:  cfg.ReadBuffer,
		WriteBufferSize: cfg.WriteBuffer,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	// 注册默认处理器
	s.registerDefaultHandlers()

	// 启动ping/pong
	go s.runPingPong()

	return s
}

// registerDefaultHandlers 注册默认处理器
func (s *Server) registerDefaultHandlers() {
	s.RegisterHandler("message", s.handleMessage)
	s.RegisterHandler("ping", s.handlePing)
	s.RegisterHandler("subscribe", s.handleSubscribe)
	s.RegisterHandler("unsubscribe", s.handleUnsubscribe)
}

// RegisterHandler 注册消息处理器
func (s *Server) RegisterHandler(msgType string, handler MessageHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handlers[msgType] = handler
}

// HandleWebSocket 处理WebSocket连接
func (s *Server) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[WebSocket] 升级失败: %v", err)
		return
	}

	// 检查连接数
	s.mu.Lock()
	if len(s.connections) >= s.config.MaxConnections {
		s.mu.Unlock()
		conn.Close()
		return
	}

	// 创建连接
	wsConn := &Conn{
		ID:         generateID(),
		Socket:    conn,
		RemoteAddr: conn.RemoteAddr().String(),
		ConnectedAt: time.Now(),
	}

	s.connections[wsConn.ID] = wsConn
	s.stats.ConnectionsTotal.Add(1)
	s.stats.ConnectionsActive.Add(1)
	s.mu.Unlock()

	log.Printf("[WebSocket] 新连接: %s", wsConn.ID)

	// 启动读协程
	go s.readLoop(wsConn)
}

// readLoop 读取循环
func (s *Server) readLoop(conn *Conn) {
	defer func() {
		s.disconnect(conn)
	}()

	conn.Socket.SetReadDeadline(time.Now().Add(s.config.ReadTimeout))
	conn.Socket.SetPongHandler(func(string) error {
		conn.mu.Lock()
		conn.LastPong = time.Now()
		conn.mu.Unlock()
		s.stats.PongsReceived.Add(1)
		return nil
	})

	for {
		_, data, err := conn.Socket.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				s.stats.Errors.Add(1)
			}
			return
		}

		s.stats.BytesReceived.Add(int64(len(data)))
		s.stats.MessagesReceived.Add(1)

		// 解析消息
		var msg Message
		if err := json.Unmarshal(data, &msg); err != nil {
			s.stats.Errors.Add(1)
			continue
		}

		// 处理消息
		s.mu.RLock()
		handler, ok := s.handlers[msg.Type]
		s.mu.RUnlock()

		if ok {
			handler(conn, &msg)
		}
	}
}

// writeLoop 写入循环
func (s *Server) writeLoop(conn *Conn, msgs <-chan []byte) {
	defer conn.Socket.Close()

	for {
		select {
		case msg, ok := <-msgs:
			if !ok {
				return
			}

			conn.Socket.SetWriteDeadline(time.Now().Add(s.config.WriteTimeout))
			if err := conn.Socket.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}

			s.stats.BytesSent.Add(int64(len(msg)))
			s.stats.MessagesSent.Add(1)
		}
	}
}

// Send 发送消息
func (s *Server) Send(connID string, msg *Message) error {
	s.mu.RLock()
	conn, ok := s.connections[connID]
	s.mu.RUnlock()

	if !ok {
		return ErrConnectionNotFound
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	conn.Socket.SetWriteDeadline(time.Now().Add(s.config.WriteTimeout))
	return conn.Socket.WriteMessage(websocket.TextMessage, data)
}

// Broadcast 广播
func (s *Server) Broadcast(msg *Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, conn := range s.connections {
		go func(c *Conn) {
			c.Socket.SetWriteDeadline(time.Now().Add(s.config.WriteTimeout))
			c.Socket.WriteMessage(websocket.TextMessage, data)
		}(conn)
	}
}

// disconnect 断开连接
func (s *Server) disconnect(conn *Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.connections[conn.ID]; ok {
		conn.Socket.Close()
		delete(s.connections, conn.ID)
		s.stats.ConnectionsActive.Add(-1)
		log.Printf("[WebSocket] 断开连接: %s", conn.ID)
	}
}

// runPingPong 运行ping/pong
func (s *Server) runPingPong() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.mu.RLock()
			for _, conn := range s.connections {
				conn.Socket.SetWriteDeadline(time.Now().Add(10 * time.Second))
				conn.Socket.WriteMessage(websocket.PingMessage, nil)
				s.stats.PingsSent.Add(1)
			}
			s.mu.RUnlock()
		}
	}
}

// Stats 获取统计
func (s *Server) Stats() Stats {
	return Stats{
		ConnectionsTotal:  s.stats.ConnectionsTotal,
		ConnectionsActive: s.stats.ConnectionsActive,
		MessagesReceived: s.stats.MessagesReceived,
		MessagesSent:     s.stats.MessagesSent,
		BytesReceived:    s.stats.BytesReceived,
		BytesSent:       s.stats.BytesSent,
		PingsSent:        s.stats.PingsSent,
		PongsReceived:    s.stats.PongsReceived,
		Errors:           s.stats.Errors,
	}
}

// Stop 停止服务器
func (s *Server) Stop() {
	s.cancel()

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, conn := range s.connections {
		conn.Socket.Close()
	}
}

// ============ 默认处理器 ============

func (s *Server) handleMessage(conn *Conn, msg *Message) {
	// 广播消息
	s.Broadcast(&Message{
		Type:    "message",
		Payload: msg.Payload,
	})
}

func (s *Server) handlePing(conn *Conn, msg *Message) {
	s.Send(conn.ID, &Message{Type: "pong"})
}

func (s *Server) handleSubscribe(conn *Conn, msg *Message) {
	var payload struct {
		Channel string `json:"channel"`
	}
	json.Unmarshal(msg.Payload, &payload)

	s.Send(conn.ID, &Message{
		Type:    "subscribed",
		Payload: msg.Payload,
	})
}

func (s *Server) handleUnsubscribe(conn *Conn, msg *Message) {
	var payload struct {
		Channel string `json:"channel"`
	}
	json.Unmarshal(msg.Payload, &payload)

	s.Send(conn.ID, &Message{
		Type:    "unsubscribed",
		Payload: msg.Payload,
	})
}

// Errors
var (
	ErrConnectionNotFound = &WSError{Code: "CONNECTION_NOT_FOUND", Message: "连接未找到"}
)

// WSError WebSocket错误
type WSError struct {
	Code    string
	Message string
}

func (e *WSError) Error() string {
	return e.Code + ": " + e.Message
}

// generateID 生成ID
func generateID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

// randomString 生成随机字符串
func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}

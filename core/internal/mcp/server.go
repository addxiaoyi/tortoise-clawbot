package mcp

import (
	"context"
	"encoding/json"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

// ProtocolVersion MCP协议版本
const ProtocolVersion = "1.0"

// MessageType 消息类型
type MessageType string

const (
	TypeRequest    MessageType = "request"
	TypeResponse   MessageType = "response"
	TypeNotification MessageType = "notification"
	TypeError      MessageType = "error"
)

// JSONRPCRequest JSON-RPC请求
type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
	ID     interface{}      `json:"id,omitempty"`
}

// JSONRPCResponse JSON-RPC响应
type JSONRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	Result  interface{}     `json:"result,omitempty"`
	Error  *JSONRPCError   `json:"error,omitempty"`
	ID     interface{}      `json:"id,omitempty"`
}

// JSONRPCError JSON-RPC错误
type JSONRPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Config MCP服务器配置
type Config struct {
	Protocol string
	Timeout  time.Duration
}

// Server MCP协议服务器
type Server struct {
	config Config

	// 方法注册
	methods map[string]Handler

	// 会话
	sessions map[string]*MCPSession

	// 连接
	connections map[string]*Connection

	// 统计
	stats Stats

	// 控制
	ctx    context.Context
	cancel context.CancelFunc

	mu sync.RWMutex
}

// Handler 方法处理器
type Handler func(ctx context.Context, params json.RawMessage) (interface{}, error)

// MCPSession MCP会话
type MCPSession struct {
	ID        string
	UserID   string
	CreatedAt time.Time
	Metadata  map[string]interface{}
}

// Connection MCP连接
type Connection struct {
	ID        string
	SessionID string
	Protocol  string
	CreatedAt time.Time
	LastPing time.Time
	mu       sync.RWMutex
}

// Stats MCP统计
type Stats struct {
	RequestsReceived  atomic.Int64
	RequestsSent     atomic.Int64
	NotificationsSent atomic.Int64
	Errors           atomic.Int64
	ActiveSessions   atomic.Int64
	ActiveConnections atomic.Int64
}

// NewServer 创建MCP服务器
func NewServer(cfg Config) *Server {
	if cfg.Protocol == "" {
		cfg.Protocol = ProtocolVersion
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}

	ctx, cancel := context.WithCancel(context.Background())

	s := &Server{
		config:     cfg,
		methods:   make(map[string]Handler),
		sessions:  make(map[string]*MCPSession),
		connections: make(map[string]*Connection),
		ctx:       ctx,
		cancel:    cancel,
	}

	// 注册内置方法
	s.registerBuiltinMethods()

	return s
}

// registerBuiltinMethods 注册内置方法
func (s *Server) registerBuiltinMethods() {
	// 初始化方法
	s.RegisterMethod("initialize", s.handleInitialize)
	s.RegisterMethod("initialize/complete", s.handleInitializeComplete)

	// 工具方法
	s.RegisterMethod("tools/list", s.handleToolsList)
	s.RegisterMethod("tools/call", s.handleToolsCall)

	// 资源方法
	s.RegisterMethod("resources/list", s.handleResourcesList)
	s.RegisterMethod("resources/read", s.handleResourcesRead)
	s.RegisterMethod("resources/subscribe", s.handleResourcesSubscribe)

	// 提示方法
	s.RegisterMethod("prompts/list", s.handlePromptsList)
	s.RegisterMethod("prompts/get", s.handlePromptsGet)

	// Ping
	s.RegisterMethod("ping", s.handlePing)
}

// RegisterMethod 注册方法
func (s *Server) RegisterMethod(name string, handler Handler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.methods[name] = handler
}

// HandleRequest 处理请求
func (s *Server) HandleRequest(connID string, req *JSONRPCRequest) (*JSONRPCResponse, error) {
	s.stats.RequestsReceived.Add(1)

	s.mu.RLock()
	handler, ok := s.methods[req.Method]
	s.mu.RUnlock()

	if !ok {
		return &JSONRPCResponse{
			JSONRPC: "2.0",
			Error: &JSONRPCError{
				Code:    -32601,
				Message: "Method not found: " + req.Method,
			},
			ID: req.ID,
		}, nil
	}

	// 执行处理
	ctx, cancel := context.WithTimeout(s.ctx, s.config.Timeout)
	defer cancel()

	result, err := handler(ctx, req.Params)
	if err != nil {
		s.stats.Errors.Add(1)
		return &JSONRPCResponse{
			JSONRPC: "2.0",
			Error: &JSONRPCError{
				Code:    -32603,
				Message: err.Error(),
			},
			ID: req.ID,
		}, nil
	}

	s.stats.RequestsSent.Add(1)
	return &JSONRPCResponse{
		JSONRPC: "2.0",
		Result:  result,
		ID:     req.ID,
	}, nil
}

// HandleNotification 处理通知
func (s *Server) HandleNotification(connID string, req *JSONRPCRequest) error {
	s.mu.RLock()
	handler, ok := s.methods[req.Method]
	s.mu.RUnlock()

	if !ok {
		return nil
	}

	ctx := s.ctx
	handler(ctx, req.Params)
	return nil
}

// CreateSession 创建会话
func (s *Server) CreateSession(userID string) string {
	session := &MCPSession{
		ID:        uuid.New().String(),
		UserID:   userID,
		CreatedAt: time.Now(),
		Metadata:  make(map[string]interface{}),
	}

	s.mu.Lock()
	s.sessions[session.ID] = session
	s.stats.ActiveSessions.Add(1)
	s.mu.Unlock()

	return session.ID
}

// DeleteSession 删除会话
func (s *Server) DeleteSession(sessionID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if session, ok := s.sessions[sessionID]; ok {
		delete(s.sessions, sessionID)
		s.stats.ActiveSessions.Add(-1)

		// 清理关联连接
		for connID, conn := range s.connections {
			if conn.SessionID == session.ID {
				delete(s.connections, connID)
			}
		}
	}
}

// RegisterConnection 注册连接
func (s *Server) RegisterConnection(sessionID, protocol string) string {
	conn := &Connection{
		ID:        uuid.New().String(),
		SessionID: sessionID,
		Protocol:  protocol,
		CreatedAt: time.Now(),
		LastPing: time.Now(),
	}

	s.mu.Lock()
	s.connections[conn.ID] = conn
	s.stats.ActiveConnections.Add(1)
	s.mu.Unlock()

	return conn.ID
}

// UnregisterConnection 注销连接
func (s *Server) UnregisterConnection(connID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.connections[connID]; ok {
		delete(s.connections, connID)
		s.stats.ActiveConnections.Add(-1)
	}
}

// UpdatePing 更新ping
func (s *Server) UpdatePing(connID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if conn, ok := s.connections[connID]; ok {
		conn.LastPing = time.Now()
	}
}

// Stats 获取统计
func (s *Server) Stats() Stats {
	return Stats{
		RequestsReceived:  s.stats.RequestsReceived,
		RequestsSent:     s.stats.RequestsSent,
		NotificationsSent: s.stats.NotificationsSent,
		Errors:           s.stats.Errors,
		ActiveSessions:   s.stats.ActiveSessions,
		ActiveConnections: s.stats.ActiveConnections,
	}
}

// ============ 内置方法处理器 ============

func (s *Server) handleInitialize(ctx context.Context, params json.RawMessage) (interface{}, error) {
	return map[string]interface{}{
		"protocolVersion": s.config.Protocol,
		"serverInfo": map[string]string{
			"name":    "tortoise-mcp",
			"version": "1.0.0",
		},
		"capabilities": map[string]interface{}{
			"tools":      true,
			"resources":  true,
			"prompts":    true,
		},
	}, nil
}

func (s *Server) handleInitializeComplete(ctx context.Context, params json.RawMessage) (interface{}, error) {
	return map[string]bool{"initialized": true}, nil
}

func (s *Server) handleToolsList(ctx context.Context, params json.RawMessage) (interface{}, error) {
	return map[string]interface{}{
		"tools": []interface{}{},
	}, nil
}

func (s *Server) handleToolsCall(ctx context.Context, params json.RawMessage) (interface{}, error) {
	return map[string]interface{}{
		"content": []map[string]interface{}{
			{"type": "text", "text": "Tool called successfully"},
		},
	}, nil
}

func (s *Server) handleResourcesList(ctx context.Context, params json.RawMessage) (interface{}, error) {
	return map[string]interface{}{
		"resources": []interface{}{},
	}, nil
}

func (s *Server) handleResourcesRead(ctx context.Context, params json.RawMessage) (interface{}, error) {
	return map[string]interface{}{
		"contents": []map[string]interface{}{
			{"type": "text", "text": "Resource content"},
		},
	}, nil
}

func (s *Server) handleResourcesSubscribe(ctx context.Context, params json.RawMessage) (interface{}, error) {
	return map[string]bool{"subscribed": true}, nil
}

func (s *Server) handlePromptsList(ctx context.Context, params json.RawMessage) (interface{}, error) {
	return map[string]interface{}{
		"prompts": []interface{}{},
	}, nil
}

func (s *Server) handlePromptsGet(ctx context.Context, params json.RawMessage) (interface{}, error) {
	return map[string]interface{}{
		"messages": []interface{}{},
	}, nil
}

func (s *Server) handlePing(ctx context.Context, params json.RawMessage) (interface{}, error) {
	return map[string]string{"status": "pong"}, nil
}

// Stop 停止服务器
func (s *Server) Stop() {
	s.cancel()
}

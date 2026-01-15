package gateway

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
)

// Config Gateway配置
type Config struct {
	Port         int
	TLSEnabled   bool
	MaxConns    int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// Server Gateway服务器
type Server struct {
	config Config
	router *gin.Engine
	server *http.Server

	// 连接管理
	connections map[string]*Connection

	// 统计
	stats Stats

	// 控制
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	mu sync.RWMutex
}

// Connection 连接
type Connection struct {
	ID        string
	RemoteAddr string
	ConnectedAt time.Time
	LastActive time.Time
	Protocol   string
}

// Stats 网关统计
type Stats struct {
	ConnectionsActive   atomic.Int64
	ConnectionsTotal   atomic.Int64
	MessagesReceived  atomic.Int64
	MessagesSent     atomic.Int64
	BytesReceived    atomic.Int64
	BytesSent       atomic.Int64
	Errors          atomic.Int64
}

// NewServer 创建Gateway服务器
func NewServer(cfg Config) *Server {
	if cfg.Port == 0 {
		cfg.Port = 18792
	}
	if cfg.MaxConns == 0 {
		cfg.MaxConns = 100000
	}
	if cfg.ReadTimeout == 0 {
		cfg.ReadTimeout = 30 * time.Second
	}
	if cfg.WriteTimeout == 0 {
		cfg.WriteTimeout = 30 * time.Second
	}

	ctx, cancel := context.WithCancel(context.Background())

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())

	s := &Server{
		config:      cfg,
		router:      router,
		connections: make(map[string]*Connection),
		ctx:         ctx,
		cancel:      cancel,
	}

	s.setupRoutes()
	return s
}

// setupRoutes 设置路由
func (s *Server) setupRoutes() {
	// 健康检查
	s.router.GET("/health", s.handleHealth)
	s.router.GET("/api/v1/health", s.handleHealth)

	// API路由组
	v1 := s.router.Group("/api/v1")
	{
		// 会话管理
		sessions := v1.Group("/sessions")
		{
			sessions.GET("", s.handleListSessions)
			sessions.POST("", s.handleCreateSession)
			sessions.GET("/:id", s.handleGetSession)
			sessions.DELETE("/:id", s.handleDeleteSession)
			sessions.GET("/:id/messages", s.handleGetMessages)
			sessions.POST("/:id/messages", s.handleSendMessage)
		}

		// 记忆管理
		memories := v1.Group("/memories")
		{
			memories.GET("", s.handleListMemories)
			memories.POST("", s.handleCreateMemory)
			memories.DELETE("/:id", s.handleDeleteMemory)
			memories.GET("/search", s.handleSearchMemories)
		}

		// 插件管理
		plugins := v1.Group("/plugins")
		{
			plugins.GET("", s.handleListPlugins)
			plugins.POST("", s.handleInstallPlugin)
			plugins.GET("/:id", s.handleGetPlugin)
			plugins.PATCH("/:id", s.handleUpdatePlugin)
			plugins.DELETE("/:id", s.handleUninstallPlugin)
		}

		// 工具管理
		tools := v1.Group("/tools")
		{
			tools.GET("", s.handleListTools)
			tools.POST("/execute", s.handleExecuteTool)
		}

		// 统计
		v1.GET("/stats", s.handleStats)

		// WebSocket
		v1.GET("/ws", s.handleWebSocket)
	}
}

// Router 获取路由
func (s *Server) Router() *gin.Engine {
	return s.router
}

// Start 启动服务器
func (s *Server) Start(addr string) error {
	s.server = &http.Server{
		Addr:         addr,
		Handler:      s.router,
		ReadTimeout:  s.config.ReadTimeout,
		WriteTimeout: s.config.WriteTimeout,
	}

	log.Printf("[Gateway] 启动中: %s", addr)
	return s.server.ListenAndServe()
}

// Stop 停止服务器
func (s *Server) Stop() error {
	s.cancel()
	return s.server.Shutdown(context.Background())
}

// ============ 处理器 ============

func (s *Server) handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"version":   "1.0.0",
	})
}

func (s *Server) handleListSessions(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"sessions": []interface{}{},
		"total":   0,
	})
}

func (s *Server) handleCreateSession(c *gin.Context) {
	var req struct {
		Name   string `json:"name" binding:"required"`
		UserID string `json:"user_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":        "sess_" + randomID(),
		"name":      req.Name,
		"userId":    req.UserID,
		"createdAt": time.Now(),
	})
}

func (s *Server) handleGetSession(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"id":        c.Param("id"),
		"name":      "Session",
		"createdAt": time.Now(),
	})
}

func (s *Server) handleDeleteSession(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func (s *Server) handleGetMessages(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"messages": []interface{}{},
		"total":   0,
	})
}

func (s *Server) handleSendMessage(c *gin.Context) {
	var req struct {
		Content string `json:"content" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	s.stats.MessagesReceived.Add(1)

	c.JSON(http.StatusOK, gin.H{
		"id":      "msg_" + randomID(),
		"content": req.Content,
		"role":    "assistant",
	})
}

func (s *Server) handleListMemories(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"memories": []interface{}{},
		"total":   0,
	})
}

func (s *Server) handleCreateMemory(c *gin.Context) {
	var req struct {
		Type       string  `json:"type" binding:"required"`
		Content    string  `json:"content" binding:"required"`
		Importance float64 `json:"importance"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":         "mem_" + randomID(),
		"type":       req.Type,
		"content":    req.Content,
		"importance": req.Importance,
		"createdAt":  time.Now(),
	})
}

func (s *Server) handleDeleteMemory(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func (s *Server) handleSearchMemories(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"memories": []interface{}{},
		"total":   0,
	})
}

func (s *Server) handleListPlugins(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"plugins": []interface{}{},
		"total":   0,
	})
}

func (s *Server) handleInstallPlugin(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "installed"})
}

func (s *Server) handleGetPlugin(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{})
}

func (s *Server) handleUpdatePlugin(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{})
}

func (s *Server) handleUninstallPlugin(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "uninstalled"})
}

func (s *Server) handleListTools(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"tools": []interface{}{},
		"total": 0,
	})
}

func (s *Server) handleExecuteTool(c *gin.Context) {
	var req struct {
		Tool string                 `json:"tool" binding:"required"`
		Args map[string]interface{} `json:"args"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"tool":    req.Tool,
		"result":  "executed",
	})
}

func (s *Server) handleStats(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"connections":   s.stats.ConnectionsTotal.Load(),
		"messages":     s.stats.MessagesReceived.Load(),
		"bytesReceived": s.stats.BytesReceived.Load(),
	})
}

func (s *Server) handleWebSocket(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "websocket endpoint"})
}

// Stats 获取统计
func (s *Server) Stats() Stats {
	return Stats{
		ConnectionsActive: s.stats.ConnectionsActive,
		ConnectionsTotal: s.stats.ConnectionsTotal,
		MessagesReceived: s.stats.MessagesReceived,
		MessagesSent:    s.stats.MessagesSent,
		BytesReceived:   s.stats.BytesReceived,
		BytesSent:      s.stats.BytesSent,
		Errors:         s.stats.Errors,
	}
}

// randomID 生成随机ID
func randomID() string {
	b := make([]byte, 8)
	for i := range b {
		b[i] = byte(time.Now().UnixNano() % 62)
		if b[i] < 10 {
			b[i] += '0'
		} else if b[i] < 36 {
			b[i] += 'a' - 10
		} else {
			b[i] += 'A' - 36
		}
	}
	return string(b)
}
